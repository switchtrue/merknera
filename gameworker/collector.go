package gameworker

import "github.com/mleonard87/merknera/repository"

var GameMoveQueue = make(chan GameMoveRequest, 100)

func QueueGameMove(rpcMethod string, move repository.GameMove) {
	gmr := *NewGameMoveRequest(rpcMethod, move)
	GameMoveQueue <- gmr
}
