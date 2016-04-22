package gameworker

import (
	"fmt"
	"log"

	"github.com/mleonard87/merknera/games"
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
				gameType := work.GameMove.GameBot.Game.GameType
				fmt.Printf("Bot: %s Endpoint: %s For: %s\n", bot.Name, bot.RPCEndpoint, gameType.Name)

				// Ping the bot to ensure its still online.
				success := bot.Ping()
				if success == false {
					return
				}

				gameManager, err := games.GetGameManager(gameType)
				if err != nil {
					log.Fatal(err)
				}
				err = work.GameMove.MarkStarted()
				if err != nil {
					log.Fatal(err)
				}

				method := gameManager.GetNextMoveRPCMethodName()
				params := gameManager.GetNextMoveRPCParams(work.GameMove)

				rpchelper.Call(bot.RPCEndpoint, method, params)

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
