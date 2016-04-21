package gameworker

import "github.com/mleonard87/merknera/repository"

var GameMoveQueue = make(chan GameMoveRequest, 100)

func QueueGameMove(move repository.GameMove) {
	gmrequest := GameMoveRequest{GameMove: move}

	GameMoveQueue <- gmrequest
}
