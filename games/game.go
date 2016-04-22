package games

import (
	"database/sql"
	"errors"
	"log"

	"github.com/mleonard87/merknera/repository"
)

type GameManager interface {
	GenerateGames(db *sql.DB, bot repository.Bot) []repository.Game
	Mnemonic() string
	Name() string
	GetNextMoveRPCMethodName() string
	GetNextMoveRPCParams(gameMove repository.GameMove) interface{}
}

var RegisteredGameManagers []GameManager

func RegisterGameManager(gameManager GameManager) error {
	db := repository.NewTransaction()
	_, err := repository.GetGameTypeByMnemonic(db, gameManager.Mnemonic())
	if err != nil {
		_, err := repository.CreateGameType(db, gameManager.Mnemonic(), gameManager.Name())
		if err != nil {
			return err
		}
	}

	RegisteredGameManagers = append(RegisteredGameManagers, gameManager)

	return nil
}

func GetGameManager(gameType repository.GameType) (GameManager, error) {
	log.Println("GetGameManager")
	for _, gm := range RegisteredGameManagers {
		if gm.Mnemonic() == gameType.Mnemonic {
			return gm, nil
		}
	}

	return nil, errors.New("Unknown game type.")
}
