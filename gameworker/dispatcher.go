package gameworker

import "fmt"

var WorkerQueue chan chan GameMoveRequest

func StartGameMoveDispatcher(numworkers int) {
	fmt.Println("Starting dispatcher")
	// first, initialize the channel we are going to put the workers work channels into.
	WorkerQueue = make(chan chan GameMoveRequest, numworkers)

	// Now, create all our workers.
	for i := 0; i < numworkers; i++ {
		fmt.Println("Starting worker", i+1)
		worker := NewGameMoveWorker(i+1, WorkerQueue)
		worker.Start()
	}

	go func() {
		for {
			select {
			case work := <-GameMoveQueue:
				go func() {
					worker := <-WorkerQueue
					worker <- work
				}()
			}
		}
	}()
}
