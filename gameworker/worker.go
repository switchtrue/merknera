package gameworker

import (
	"fmt"
	"log"

	"github.com/mleonard87/merknera/games"
	"github.com/mleonard87/merknera/repository"
	"github.com/mleonard87/merknera/rpchelper"
)

type GameMoveWorker struct {
	Id                  int
	GameMoveRequestWork chan GameMoveRequest
	WorkerQueue         chan chan GameMoveRequest
	QuitChan            chan bool
}

// NewWorker creates, and returns a new GameMoveWorker object. Its only argument
// is a channel that the worker can add itself to whenever it has done its work.
func NewGameMoveWorker(id int, gameMoveQueue chan chan GameMoveRequest) GameMoveWorker {
	worker := GameMoveWorker{
		Id:                  id,
		GameMoveRequestWork: make(chan GameMoveRequest),
		WorkerQueue:         gameMoveQueue,
		QuitChan:            make(chan bool),
	}

	return worker
}

type PingArgs struct{}

type PingResult struct {
	Ping string `json:"ping"`
}

// Start begins the worker by starting a goroutine, this is, an infinite
// for-select loop.
func (gmw GameMoveWorker) Start() {
	go func() {
		for {
			// Add ourselves into the worker queue
			gmw.WorkerQueue <- gmw.GameMoveRequestWork

			select {
			case work := <-gmw.GameMoveRequestWork:
				bot := work.GameMove.GameBot.Bot
				game := work.GameMove.GameBot.Game
				gameType := game.GameType

				fmt.Printf("Bot: %s Endpoint: %s For: %s\n", bot.Name, bot.RPCEndpoint, gameType.Name)

				// If the Bot is marked as ERROR or OFFLINE then don't process this move.
				if bot.Status != repository.BOT_STATUS_ONLINE {
					return
				}

				// Ping the bot to ensure its still online.
				success := bot.Ping()
				if success == false {
					err := bot.MarkOffline()
					if err != nil {
						log.Fatal(err)
					}
					return
				}

				gameManager, err := games.GetGameManager(gameType)
				if err != nil {
					log.Fatal(err)
				}

				err = game.MarkInProgress()
				if err != nil {
					log.Fatal(err)
				}

				method := gameManager.GetNextMoveRPCMethodName()

				params, err := gameManager.GetNextMoveRPCParams(work.GameMove)
				if err != nil {
					log.Fatal(err)
				}
				reply := gameManager.GetNextMoveRPCResult(work.GameMove)

				var rsr rpchelper.RPCServerResponse
				rsr.Result = reply
				err = rpchelper.Call(bot.RPCEndpoint, method, params, &rsr)
				if err != nil {
					fmt.Println("Call failed")
					log.Fatal(err)
				}

				if res, ok := rsr.Result.(map[string]interface{}); ok {
					gs, winner, err := gameManager.ProcessMove(work.GameMove, res)
					if err != nil {
						log.Fatal(err)
					}

					err = game.SetGameState(gs)
					if err != nil {
						log.Fatal(err)
					}

					if !winner {
						nextBot, err := gameManager.GetGameBotForNextMove(work.GameMove)
						if err != nil {
							log.Fatal(err)
						}
						nextMove, err := repository.CreateGameMove(nextBot)
						if err != nil {
							log.Fatal(err)
						}
						QueueGameMove(nextMove)
					} else {
						game.MarkComplete()
						p, err := game.Players()
						if err != nil {
							log.Fatal(err)
						}

						// Send each player a Complete notification.
						cm := gameManager.GetCompleteRPCMethodName()
						for _, p := range p {
							cp, err := gameManager.GetCompleteRPCParams(p)
							if err != nil {
								log.Fatal(err)
							}
							err = rpchelper.Notify(p.Bot.RPCEndpoint, cm, cp)
							if err != nil {
								log.Fatal(err)
							}
						}
					}
				}

				err = work.GameMove.MarkComplete()
				if err != nil {
					log.Fatal(err)
				}

			case <-gmw.QuitChan:
				// We have been asked to stop.
				fmt.Printf("worker%d stopping\n", gmw.Id)
				return
			}
		}
	}()
}

// Stop tells the worker to stop listening for work requests.
//
// Note that the worker will only stop *after* it has finished its work.
func (gmw GameMoveWorker) Stop() {
	go func() {
		gmw.QuitChan <- true
	}()
}
