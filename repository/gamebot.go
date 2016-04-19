package repository

type GameBot struct {
	Id       int
	Game     Game
	Bot      Bot
	Sequence int
}

func CreateGameBot(game Game, bot Bot, sequence int) (GameBot, error) {
	db := GetDatabaseConnection()
	defer db.Close()

	var gameBotId int
	err := db.QueryRow(`
	INSERT INTO game_bot (
	  game_id
	, bot_id
	, sequence
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
	db := GetDatabaseConnection()
	defer db.Close()

	var gameBot GameBot
	var gameId int
	var botId int
	err := db.QueryRow(`
	SELECT
	  gb.id
	, gb.sequence
	, gb.game_id
	, gb.bot_id
	FROM game_bot gb
	WHERE gb.id = $1
	`, id).Scan(&gameBot.Id, &gameBot.Sequence, &gameId, &botId)
	if err != nil {
		return GameBot{}, err
	}
	gameBot.Game, _ = GetGameById(gameId)
	gameBot.Bot, _ = GetBotById(botId)

	return gameBot, nil
}
