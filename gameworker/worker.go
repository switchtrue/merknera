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
				log.Printf("[wkr%d] Working (move id: %d)\n", gmw.Id, work.GameMove.Id)

				// If for some reason the game move got added to the work queue twice and has already
				// been processed just return.
				if work.GameMove.Status != repository.GAMEMOVE_STATUS_AWAITING {
					continue
				}

				gameBot, err := work.GameMove.GameBot()
				if err != nil {
					log.Fatal(err)
				}

				bot, err := gameBot.Bot()
				if err != nil {
					log.Fatal(err)
				}

				game, err := gameBot.Game()
				if err != nil {
					log.Fatal(err)
				}

				gameType, err := game.GameType()
				if err != nil {
					log.Fatal(err)
				}

				// If the Bot is marked as ERROR or OFFLINE then don't process this move.
				if bot.Status != repository.BOT_STATUS_ONLINE {
					continue
				}

				// Ping the bot to ensure its still online.
				beforeStatus := bot.Status
				success, err := bot.Ping()
				if err != nil {
					log.Fatal(err)
				}
				if success == false {
					bot.MarkOffline()
					continue
				}

				// If the bots status has changed and its now online then re-queue any awaiting moves
				// for this bot.
				if beforeStatus != bot.Status && bot.Status == repository.BOT_STATUS_ONLINE {
					awaitingMoves, err := repository.GetAwaitingMovesForBot(bot)
					if err != nil {
						log.Fatal(err)
					}
					for _, gm := range awaitingMoves {
						QueueGameMove(gm)
					}
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
				log.Printf("[wkr%d] Calling %s for %s (move id: %d)\n", gmw.Id, method, bot.Name, work.GameMove.Id)
				err = rpchelper.Call(bot.RPCEndpoint, method, params, &rsr)
				log.Printf("[wkr%d] Call %s complete for %s (move id: %d)\n", gmw.Id, method, bot.Name, work.GameMove.Id)
				if err != nil {
					fmt.Println("Call failed")
					log.Fatal(err)
				}

				if res, ok := rsr.Result.(map[string]interface{}); ok {
					gs, winner, err := gameManager.ProcessMove(work.GameMove, res)
					if err != nil {
						log.Fatal(err)
					}

					if !winner {
						nextBot, err := gameManager.GetGameBotForNextMove(work.GameMove)
						if err != nil {
							log.Fatal(err)
						}
						nextMove, err := repository.CreateGameMove(nextBot, gs)
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

							pb, err := p.Bot()
							if err != nil {
								log.Fatal(err)
							}

							err = rpchelper.Notify(pb.RPCEndpoint, cm, cp)
							if err != nil {
								log.Fatal(err)
							}
						}
					}

					err = work.GameMove.SetGameState(gs)
					if err != nil {
						log.Fatal(err)
					}
				}

				err = work.GameMove.MarkComplete()
				if err != nil {
					log.Fatal(err)
				}

				continue

			case <-gmw.QuitChan:
				// We have been asked to stop.
				fmt.Printf("worker%d stopping\n", gmw.Id)
				return
			}
		}
	}()
}

// Stop tells the worker to stop listening for work requests.
// Note that the worker will only stop *after* it has finished its work.
func (gmw GameMoveWorker) Stop() {
	go func() {
		gmw.QuitChan <- true
	}()
}
