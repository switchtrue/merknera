package repository

import (
	"database/sql"
	"errors"
	"fmt"
	"log"
)

type GameType struct {
	Id       int
	Mnemonic string
	Name     string
}

func CreateGameType(mnemonic string, name string) (GameType, error) {
	var gameTypeId int
	db := GetDB()
	err := db.QueryRow(`
	INSERT INTO game_type (
	  mnemonic
	, name
	) VALUES (
	  $1
	, $2
	) RETURNING id
	`, mnemonic, name).Scan(&gameTypeId)
	if err != nil {
		log.Printf("An error occurred in gametype.CreateGameType():1:\n%s\n", err)
		return GameType{}, err
	}

	gameType, err := GetGameTypeById(gameTypeId)
	if err != nil {
		log.Printf("An error occurred in gametype.CreateGameType():2:\n%s\n", err)
		return GameType{}, err
	}
	return gameType, nil
}

func GetGameTypeByMnemonic(mnemonic string) (GameType, error) {
	var gameType GameType
	db := GetDB()
	err := db.QueryRow(`
	SELECT
	  id
	, mnemonic
	, name
	FROM game_type
	WHERE mnemonic = $1
	`, mnemonic).Scan(&gameType.Id, &gameType.Mnemonic, &gameType.Name)
	if err != nil {
		if err == sql.ErrNoRows {
			em := fmt.Sprintf("Game \"%s\" is not known", mnemonic)
			return GameType{}, errors.New(em)
		}
		log.Printf("An error occurred in gametype.GetGameTypeByMnemonic():\n%s\n", err)
		return GameType{}, err
	}

	return gameType, nil
}

func GetGameTypeById(id int) (GameType, error) {
	var gameType GameType
	db := GetDB()
	err := db.QueryRow(`
	SELECT
	  gt.id
	, gt.mnemonic
	, gt.name
	FROM game_type gt
	WHERE gt.id = $1
	`, id).Scan(&gameType.Id, &gameType.Mnemonic, &gameType.Name)
	if err != nil {
		log.Printf("An error occurred in gametype.GetGameTypeById():\n%s\n", err)
		return GameType{}, err
	}

	return gameType, nil
}
