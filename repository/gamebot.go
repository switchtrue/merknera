package repository

import (
	"database/sql"
	"log"
)

type GameBot struct {
	Id       int
	Game     Game
	Bot      Bot
	Sequence int
}

func CreateGameBot(db *sql.DB, game Game, bot Bot, sequence int) (GameBot, error) {
	log.Println("CreateGameBot")
	var gameBotId int
	err := db.QueryRow(`
	INSERT INTO game_bot (
	  game_id
	, bot_id
	, play_sequence
	) VALUES (
	  $1
	, $2
	, $3
	) RETURNING id
	`, game.Id, bot.Id, sequence).Scan(&gameBotId)
	if err != nil {
		return GameBot{}, err
	}

	gameBot, err := GetGameBotById(db, gameBotId)
	if err != nil {
		return GameBot{}, err
	}
	return gameBot, nil
}

func GetGameBotById(db *sql.DB, id int) (GameBot, error) {
	log.Println("GetGameBotById")
	var gameBot GameBot
	var gameId int
	var botId int
	err := db.QueryRow(`
	SELECT
	  gb.id
	, gb.play_sequence
	, gb.game_id
	, gb.bot_id
	FROM game_bot gb
	WHERE gb.id = $1
	`, id).Scan(&gameBot.Id, &gameBot.Sequence, &gameId, &botId)
	if err != nil {
		return GameBot{}, err
	}
	gameBot.Game, _ = GetGameById(db, gameId)
	gameBot.Bot, _ = GetBotById(db, botId)

	return gameBot, nil
}
