package repository

import "log"

type GameBot struct {
	Id           int
	gameId       int
	game         Game
	botId        int
	bot          Bot
	PlaySequence int
}

func (gb *GameBot) Game() (Game, error) {
	if gb.game == (Game{}) {
		g, err := GetGameById(gb.gameId)
		if err != nil {
			log.Printf("An error occurred in gamebot.Game():\n%s\n", err)
			return Game{}, err
		}
		gb.game = g
	}

	return gb.game, nil
}

func (gb *GameBot) Bot() (Bot, error) {
	if gb.bot == (Bot{}) {
		b, err := GetBotById(gb.botId)
		if err != nil {
			log.Printf("An error occurred in gamebot.Bot():\n%s\n", err)
			return Bot{}, err
		}
		gb.bot = b
	}

	return gb.bot, nil
}

func CreateGameBot(game Game, bot Bot, sequence int) (GameBot, error) {
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
		log.Printf("An error occurred in gamebot.CreateGameBot():1:\n%s\n", err)
		return GameBot{}, err
	}

	gameBot, err := GetGameBotById(gameBotId)
	if err != nil {
		log.Printf("An error occurred in gamebot.CreateGameBot():2:\n%s\n", err)
		return GameBot{}, err
	}
	return gameBot, nil
}

func GetGameBotById(id int) (GameBot, error) {
	var gameBot GameBot
	db := GetDB()
	err := db.QueryRow(`
	SELECT
	  gb.id
	, gb.play_sequence
	, gb.game_id
	, gb.bot_id
	FROM game_bot gb
	WHERE gb.id = $1
	`, id).Scan(&gameBot.Id, &gameBot.PlaySequence, &gameBot.gameId, &gameBot.botId)
	if err != nil {
		log.Printf("An error occurred in gamebot.GetGameBotById():\n%s\n", err)
		return GameBot{}, err
	}

	return gameBot, nil
}
