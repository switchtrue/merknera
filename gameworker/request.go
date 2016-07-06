package gameworker

import (
	"strings"

	"github.com/mleonard87/merknera/repository"
)

type GameMoveRequest struct {
	RPCMethod     string
	RPCNamespace  string
	RPCMethodName string
	GameMove      repository.GameMove
}

func NewGameMoveRequest(rpcMethod string, gm repository.GameMove) *GameMoveRequest {
	rpcMethodParts := strings.Split(rpcMethod, ".")

	return &GameMoveRequest{
		RPCMethod:     rpcMethod,
		RPCNamespace:  rpcMethodParts[0],
		RPCMethodName: rpcMethodParts[1],
		GameMove:      gm,
	}
}
