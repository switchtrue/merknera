package repository

import (
	"database/sql"
	"log"
)

type GameMove struct {
	Id      int
	GameBot GameBot
}

func (gm *GameMove) MarkStarted() error {
	db := GetDB()
	err := db.QueryRow(`
	UPDATE game_move
	SET status = 'STARTED'
	WHERE id = $1
	`, gm.Id).Scan()
	if err != nil {
		return err
	}

	return nil
}

func (gm *GameMove) MarkComplete() error {
	db := GetDB()
	err := db.QueryRow(`
	UPDATE game_move
	SET status = 'COMPLETE'
	WHERE id = $1
	`, gm.Id).Scan()
	if err != nil {
		return err
	}

	return nil
}

func CreateGameMove(db *sql.DB, gameBot GameBot) (GameMove, error) {
	log.Println("CreateGameMove")
	var gameMoveId int
	err := db.QueryRow(`
	INSERT INTO move (
	  game_bot_id
	) VALUES (
	  $1
	) RETURNING id
	`, gameBot.Id).Scan(&gameMoveId)
	if err != nil {
		log.Println("got herer")
		return GameMove{}, err
	}

	gameMove, err := GetGameMoveById(db, gameMoveId)
	if err != nil {
		return GameMove{}, err
	}
	return gameMove, nil
}

func GetGameMoveById(db *sql.DB, id int) (GameMove, error) {
	log.Println("GetGameMoveById")
	var gameMove GameMove
	var gameBotId int
	err := db.QueryRow(`
	SELECT
	  id
	, game_bot_id
	FROM move
	WHERE id = $1
	`, id).Scan(&gameMove.Id, &gameBotId)
	if err != nil {
		return GameMove{}, err
	}

	gameBot, err := GetGameBotById(db, gameBotId)
	if err != nil {
		return GameMove{}, err
	}

	gameMove.GameBot = gameBot

	return gameMove, nil
}
