package repository

import (
	"fmt"
	"log"
)

type GameMove struct {
	Id      int
	GameBot GameBot
	Status  GameMoveStatus
}

type GameMoveStatus string

const (
	GAMEMOVE_STATUS_AWAITING GameMoveStatus = "AWAITING"
	GAMEMOVE_STATUS_COMPLETE GameMoveStatus = "COMPLETE"
)

func (gm *GameMove) MarkComplete() error {
	fmt.Println("MarkComplete")
	db := GetDB()
	_, err := db.Exec(`
	UPDATE move
	SET status = $1
	WHERE id = $2
	`, string(GAMEMOVE_STATUS_COMPLETE), gm.Id)
	if err != nil {
		return err
	}

	return nil
}

func CreateGameMove(gameBot GameBot) (GameMove, error) {
	log.Println("CreateGameMove")
	var gameMoveId int
	db := GetDB()
	err := db.QueryRow(`
	INSERT INTO move (
	  game_bot_id
	) VALUES (
	  $1
	) RETURNING id
	`, gameBot.Id).Scan(&gameMoveId)
	if err != nil {
		return GameMove{}, err
	}

	gameMove, err := GetGameMoveById(gameMoveId)
	if err != nil {
		return GameMove{}, err
	}
	return gameMove, nil
}

func GetGameMoveById(id int) (GameMove, error) {
	log.Println("GetGameMoveById")
	var gameMove GameMove
	var gameBotId int
	var status string
	db := GetDB()
	err := db.QueryRow(`
	SELECT
	  id
	, game_bot_id
	, status
	FROM move
	WHERE id = $1
	`, id).Scan(&gameMove.Id, &gameBotId, &status)
	if err != nil {
		return GameMove{}, err
	}
	gameMove.Status = GameMoveStatus(status)

	gameBot, err := GetGameBotById(gameBotId)
	if err != nil {
		return GameMove{}, err
	}

	gameMove.GameBot = gameBot

	return gameMove, nil
}
