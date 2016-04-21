package repository

import (
	"database/sql"
	"log"
)

type Game struct {
	Id       int
	GameType GameType
	Status   string
	//Players  []GameBot
}

func (g *Game) GetNextMoveId(db *sql.DB) (int, error) {
	log.Println("GetNextMoveId")
	var nextMoveId int
	err := db.QueryRow(`
	SELECT
	  id
	FROM move
	WHERE id = $1
	AND status = 'NOT STARTED'
	`, g.Id).Scan(&nextMoveId)
	if err != nil {
		return -1, err
	}

	return nextMoveId, nil
}

func (g *Game) Players(db *sql.DB) ([]GameBot, error) {
	rows, err := db.Query(`
	SELECT
	  bot_id
	FROM game_bot
	WHERE game_id = $1
	ORDER BY play_sequence
	`, g.Id)

	var gameBotList []GameBot
	for rows.Next() {
		var gameBot GameBot
		var gameBotId int
		rows.Scan(&gameBotId)
		gameBot, err = GetGameBotById(db, gameBotId)
		if err != nil {
			return []GameBot{}, err
		}
		gameBotList = append(gameBotList, gameBot)
	}

	return gameBotList, nil
}

func CreateGame(db *sql.DB, gameType GameType) (Game, error) {
	log.Println("CreateGame")
	var gameId int
	err := db.QueryRow(`
	INSERT INTO game (
	  game_type_id
	) VALUES (
	  $1
	) RETURNING id
	`, gameType.Id).Scan(&gameId)
	if err != nil {
		log.Println("got herer")
		return Game{}, err
	}

	game, err := GetGameById(db, gameId)
	if err != nil {
		return Game{}, err
	}
	return game, nil
}

func GetGameById(db *sql.DB, id int) (Game, error) {
	log.Println("GetGameById")
	var game Game
	var gameTypeId int
	err := db.QueryRow(`
	SELECT
	  g.id
	, g.status
	, g.game_type_id
	FROM game g
	WHERE g.id = $1
	`, id).Scan(&game.Id, &game.Status, &gameTypeId)
	if err != nil {
		return Game{}, err
	}
	game.GameType, _ = GetGameTypeById(db, gameTypeId)

	return game, nil
}
