package games

import (
	"errors"

	"github.com/mleonard87/merknera/repository"
)

const (
	GAME_TYPE_TICTACTOE = "TICTACTOE"
)

type GameManager interface {
	GenerateGames(bot repository.Bot)
}

func GetGameManager(gameType repository.GameType) (GameManager, error) {
	switch gameType.Mnemonic {
	case GAME_TYPE_TICTACTOE:
		return TicTacToeGameManager{}, nil
	default:
		return nil, errors.New("Unknown game type: " + gameType.Mnemonic)
	}
}
