package games

import (
	"errors"
	"fmt"

	"github.com/mleonard87/merknera/repository"
)

type GameManager interface {
	GenerateGames(bot repository.Bot)
	Mnemonic() string
	Name() string
}

var RegisteredGameManagers []GameManager

func RegisterGameManager(gameManager GameManager) error {
	gameType, err := repository.GetGameTypeByMnemonic(gameManager.Mnemonic())
	if err != nil {
		_, err := repository.CreateGameType(gameManager.Mnemonic(), gameManager.Name())
		if err != nil {
			return err
		}
	}

	RegisteredGameManagers = append(RegisteredGameManagers, gameManager)

	fmt.Println("Registered game type:")
	fmt.Println(gameType)

	return nil
}

func GetGameManager(gameType repository.GameType) (GameManager, error) {

	for _, gm := range RegisteredGameManagers {
		if gm.Mnemonic() == gameType.Mnemonic {
			return gm, nil
		}
	}

	return nil, errors.New("Unknown game type.")
}
