package repository

import "log"

type GameBot struct {
	Id           int
	Game         Game
	Bot          Bot
	PlaySequence int
}

func CreateGameBot(game Game, bot Bot, sequence int) (GameBot, error) {
	log.Println("CreateGameBot")
	var gameBotId int
	db := GetDB()
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

	gameBot, err := GetGameBotById(gameBotId)
	if err != nil {
		return GameBot{}, err
	}
	return gameBot, nil
}

func GetGameBotById(id int) (GameBot, error) {
	log.Println("GetGameBotById")
	var gameBot GameBot
	var gameId int
	var botId int
	db := GetDB()
	err := db.QueryRow(`
	SELECT
	  gb.id
	, gb.play_sequence
	, gb.game_id
	, gb.bot_id
	FROM game_bot gb
	WHERE gb.id = $1
	`, id).Scan(&gameBot.Id, &gameBot.PlaySequence, &gameId, &botId)
	if err != nil {
		return GameBot{}, err
	}
	gameBot.Game, _ = GetGameById(gameId)
	gameBot.Bot, _ = GetBotById(botId)

	return gameBot, nil
}
