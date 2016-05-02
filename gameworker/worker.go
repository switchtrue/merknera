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

				// Lock the current game move - its possible that the same move can end up in the work
				// queue more than once if bots keep registering on the server is bounced. Lock the
				// game move so it will only be processed one at a time.
				GetGameMoveLock(work.GameMove)
				defer ReleaseGameMoveLock(work.GameMove)

				// If for some reason the game move got added to the work queue twice and has already
				// been processed just return.
				if work.GameMove.Status != repository.GAMEMOVE_STATUS_AWAITING {
					continue
				}

				gameBot, err := work.GameMove.GameBot()
				if err != nil {
					log.Printf("[wkr%d] Error retrieving GameBot:\n%v\n", gmw.Id, err)
					continue
				}

				bot, err := gameBot.Bot()
				if err != nil {
					log.Printf("[wkr%d] Error retrieving GameBot Bot:\n%v\n", gmw.Id, err)
					continue
				}

				game, err := gameBot.Game()
				if err != nil {
					log.Printf("[wkr%d] Error retrieving GameBot Game:\n%v\n", gmw.Id, err)
					continue
				}

				gameType, err := game.GameType()
				if err != nil {
					log.Printf("[wkr%d] Error retrieving GameType:\n%v\n", gmw.Id, err)
					continue
				}

				// If the Bot is marked as ERROR or OFFLINE then don't process this move.
				if bot.Status != repository.BOT_STATUS_ONLINE {
					continue
				}

				// Ping the bot to ensure its still online.
				beforeStatus := bot.Status
				success, err := bot.Ping()
				if err != nil {
					log.Printf("[wkr%d] Error Pinging Bot (bot id: %d):\n%v\n", gmw.Id, err, bot.Id)
					continue
				}
				if success == false {
					err = bot.MarkOffline()
					if err != nil {
						log.Printf("[wkr%d] Error marking bot offline (bot id: %d):\n%v\n", gmw.Id, err, bot.Id)
					}
					continue
				}

				// If the bots status has changed and its now online then re-queue any awaiting moves
				// for this bot.
				if beforeStatus != bot.Status && bot.Status == repository.BOT_STATUS_ONLINE {
					awaitingMoves, err := bot.ListAwaitingMoves()
					if err != nil {
						log.Printf("[wkr%d] Error getting awaiting moves for bot (bot id: %d):\n%v\n", gmw.Id, err, bot.Id)
						continue
					}
					for _, gm := range awaitingMoves {
						QueueGameMove(gm)
					}
				}

				gameManager, err := games.GetGameManager(gameType)
				if err != nil {
					log.Printf("[wkr%d] Error obtaining GameManager for game (gameType: %s):\n%v\n", gmw.Id, err, gameType)
					continue
				}

				err = game.MarkInProgress()
				if err != nil {
					log.Printf("[wkr%d] Error marking game in progress (game id: %d):\n%v\n", gmw.Id, err, game.Id)
					continue
				}

				method := gameManager.GetNextMoveRPCMethodName()

				params, err := gameManager.GetNextMoveRPCParams(work.GameMove)
				if err != nil {
					log.Printf("[wkr%d] Error obtaining next move RPC params (game move id: %d):\n%v\n", gmw.Id, err, work.GameMove.Id)
					continue
				}

				reply := gameManager.GetNextMoveRPCResult(work.GameMove)

				var rsr rpchelper.RPCServerResponse
				rsr.Result = reply
				log.Printf("[wkr%d] Calling %s for %s (move id: %d)\n", gmw.Id, method, bot.Name, work.GameMove.Id)
				err = rpchelper.Call(bot.RPCEndpoint, method, params, &rsr)
				log.Printf("[wkr%d] Call %s complete for %s (move id: %d)\n", gmw.Id, method, bot.Name, work.GameMove.Id)
				if err != nil {
					sendError(gameManager, work.GameMove, err)
					err = bot.MarkError()
					if err != nil {
						log.Printf("[wkr%d] Error marking a bot as error status (bot id: %d):\n%v\n", gmw.Id, err, bot.Id)
					}
					continue
				}

				if res, ok := rsr.Result.(map[string]interface{}); ok {
					gs, gameResult, err := gameManager.ProcessMove(work.GameMove, res)
					if err != nil {
						sendError(gameManager, work.GameMove, err)
						err = bot.MarkError()
						if err != nil {
							log.Printf("[wkr%d] Error marking a bot as error status after process move (bot id: %d):\n%v\n", gmw.Id, err, bot.Id)
						}
						continue
					}

					if gameResult == games.GAME_RESULT_UNDECIDED {
						nextBot, err := gameManager.GetGameBotForNextMove(work.GameMove)
						if err != nil {
							log.Printf("[wkr%d] Error obtaining game bot for next move (game move id: %d):\n%v\n", gmw.Id, err, work.GameMove.Id)
							continue
						}
						nextMove, err := repository.CreateGameMove(nextBot, gs)
						if err != nil {
							log.Printf("[wkr%d] Error creating next game move (current game move id: %d, next game bot id: %d):\n%v\n", gmw.Id, err, work.GameMove.Id, nextBot.Id)
							continue
						}
						QueueGameMove(nextMove)
					} else {
						if gameResult == games.GAME_RESULT_WIN {
							err = work.GameMove.MarkAsWin()
							if err != nil {
								log.Printf("[wkr%d] Error marking game move as win (game move id: %d):\n%v\n", gmw.Id, err, work.GameMove.Id)
							}
						}

						err = game.MarkComplete()
						if err != nil {
							log.Printf("[wkr%d] Error marking game as complete (game id: %d):\n%v\n", gmw.Id, err, game.Id)
						}
						players, err := game.Players()
						if err != nil {
							log.Printf("[wkr%d] Error obtaining a player list for the game (game id: %d):\n%v\n", gmw.Id, err, game.Id)
							continue
						}

						// Send each player a Complete notification.
						cm := gameManager.GetCompleteRPCMethodName()
						for _, p := range players {
							cp, err := gameManager.GetCompleteRPCParams(p, gameResult)
							if err != nil {
								log.Printf("[wkr%d] Error obtaining complete RPC params for player (game bot id: %d):\n%v\n", gmw.Id, err, p.Id)
								continue
							}

							pb, err := p.Bot()
							if err != nil {
								log.Printf("[wkr%d] Error obtaining bot for player (game bot id: %d):\n%v\n", gmw.Id, err, p.Id)
								continue
							}

							err = rpchelper.Notify(pb.RPCEndpoint, cm, cp)
							if err != nil {
								log.Printf("[wkr%d] Error notifying player for complete game (bot id: %d):\n%v\n", gmw.Id, err, pb.Id)
								continue
							}
						}
					}

					err = work.GameMove.SetGameState(gs)
					if err != nil {
						log.Printf("[wkr%d] Error setting game state (game move id: %d):\n%v\n", gmw.Id, err, work.GameMove.Id)
						continue
					}
				}

				err = work.GameMove.MarkComplete()
				if err != nil {
					log.Printf("[wkr%d] Error marking game move as complete (game move id: %d):\n%v\n", gmw.Id, err, work.GameMove.Id)
					continue
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

func sendError(gm games.GameManager, gameMove repository.GameMove, originalErr error) {
	em := gm.GetErrorRPCMethodName()
	ep := gm.GetErrorRPCParams(gameMove, originalErr.Error())

	fmt.Printf("")

	gb, err := gameMove.GameBot()
	if err != nil {
		log.Println("Error in sendError (game move id %d):1:\n%s\n", gameMove.Id, err)
	}

	bot, err := gb.Bot()
	if err != nil {
		log.Println("Error in sendError (game move id %d):2:\n%s\n", gameMove.Id, err)
	}

	err = rpchelper.Notify(bot.RPCEndpoint, em, ep)
	if err != nil {
		log.Println("Error in sendError (game move id %d):3:\n%s\n", gameMove.Id, err)
	}
}

// Stop tells the worker to stop listening for work requests.
// Note that the worker will only stop *after* it has finished its work.
func (gmw GameMoveWorker) Stop() {
	go func() {
		gmw.QuitChan <- true
	}()
}
