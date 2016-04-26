package games

import (
	"errors"
	"log"

	"reflect"

	"github.com/mleonard87/merknera/repository"
)

type GameManager interface {
	GenerateGames(bot repository.Bot) []repository.Game
	Mnemonic() string
	Name() string
	GetNextMoveRPCMethodName() string
	GetNextMoveRPCParams(gameMove repository.GameMove) (interface{}, error)
	GetNextMoveRPCResult(gameMove repository.GameMove) interface{}

	GetCompleteRPCMethodName() string
	GetCompleteRPCParams(gb repository.GameBot) (interface{}, error)

	GetErrorRPCMethodName() string
	GetErrorRPCParams(gm repository.GameMove, errorMessage string) interface{}

	ProcessMove(gameMove repository.GameMove, result map[string]interface{}) (interface{}, bool, error)
	GetGameBotForNextMove(currentMove repository.GameMove) (repository.GameBot, error)
}

type GameManagerMeta struct {
	GameManager           GameManager
	nextMoveRPCParamsType reflect.Type
	nextMoveRPCResultType reflect.Type
	gameStateType         reflect.Type
}

var RegisteredGameManagers []GameManagerMeta

func RegisterGameManager(gm GameManager) error {
	_, err := repository.GetGameTypeByMnemonic(gm.Mnemonic())
	if err != nil {
		_, err := repository.CreateGameType(gm.Mnemonic(), gm.Name())
		if err != nil {
			return err
		}
	}

	gmm := GameManagerMeta{}
	gmm.GameManager = gm

	RegisteredGameManagers = append(RegisteredGameManagers, gmm)

	return nil
}

func GetGameManager(gameType repository.GameType) (GameManager, error) {
	log.Println("GetGameManager")
	for _, gm := range RegisteredGameManagers {
		if gm.GameManager.Mnemonic() == gameType.Mnemonic {
			return gm.GameManager, nil
		}
	}

	return nil, errors.New("Unknown game type.")
}
