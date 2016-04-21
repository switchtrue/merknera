package games

import (
	"database/sql"
	"errors"
	"log"

	"github.com/mleonard87/merknera/repository"
)

type GameManager interface {
	GenerateGames(db *sql.DB, bot repository.Bot)
	Mnemonic() string
	Name() string
}

var RegisteredGameManagers []GameManager

func RegisterGameManager(gameManager GameManager) error {
	db := repository.NewTransaction()
	_, err := repository.GetGameTypeByMnemonic(db, gameManager.Mnemonic())
	if err != nil {
		_, err := repository.CreateGameType(db, gameManager.Mnemonic(), gameManager.Name())
		if err != nil {
			//db.Rollback()
			return err
		}
	}
	//db.Commit()

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
