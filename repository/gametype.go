package repository

import "log"

type GameType struct {
	Id       int
	Mnemonic string
	Name     string
}

func CreateGameType(mnemonic string, name string) (GameType, error) {
	log.Print("CreateGameType")
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
		return GameType{}, err
	}

	gameType, err := GetGameTypeById(gameTypeId)
	if err != nil {
		return GameType{}, err
	}
	return gameType, nil
}

func GetGameTypeByMnemonic(mnemonic string) (GameType, error) {
	log.Print("GetGameTypeByMnemonic")
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
		return GameType{}, err
	}

	return gameType, nil
}

func GetGameTypeById(id int) (GameType, error) {
	log.Printf("GetGameTypeById: %d", id)
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
		return GameType{}, err
	}

	return gameType, nil
}
