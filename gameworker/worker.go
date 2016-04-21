package gameworker

import (
	"fmt"
	"time"
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

// Start begins the worker by starting a goroutine, this is, an infinite
// for-select loop.
func (gmw GameMoveWorker) Start() {
	go func() {
		for {
			// Add ourselves into the worker queue
			gmw.WorkerQueue <- gmw.GameMoveRequestWork

			select {
			case work := <-gmw.GameMoveRequestWork:
				// Receive a work request.
				fmt.Printf("worker%d: Received work request, delaying for %f seconds\n", gmw.Id, time.Second)

				time.Sleep(time.Second)
				fmt.Printf("worker%d: Hello, %s!\n", gmw.Id, work.GameMove.GameBot.Bot.Name)
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
