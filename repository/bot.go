package repository

type Bot struct {
	Id       string
	Name     string
	Version  string
	GameType GameType
	User     User
	RpcUrl   string
}

func RegisterBot(name string, version string, gameType GameType, user User, rpcUrl string) (Bot, error) {
	db := GetDatabaseConnection()
	defer db.Close()

	var botId int
	err := db.QueryRow(`
	INSERT INTO bot (
	  name
	, version
	, game_type_id
	, user_id
	, rpc_url
	) VALUES (
	  $1
	, $2
	, $3
	, $4
	, $5
	) RETURNING id
	`, name, version, gameType.Id, user.Id, rpcUrl).Scan(&botId)
	if err != nil {
		return Bot{}, err
	}

	bot, err := GetBotById(botId)
	if err != nil {
		return Bot{}, err
	}
	return bot, nil
}

func GetBotById(id int) (Bot, error) {
	db := GetDatabaseConnection()
	defer db.Close()

	var bot Bot
	var gameTypeId int
	var userId int
	err := db.QueryRow(`
	SELECT
	  id
	, name
	, version
	, game_type_id
	, user_id
	, rpc_url
	FROM bot
	WHERE id = $1
	`, id).Scan(&bot.Id, &bot.Name, &bot.Version, &gameTypeId, &userId, &bot.RpcUrl)
	if err != nil {
		return Bot{}, err
	}

	user, err := GetUserById(userId)
	if err != nil {
		return Bot{}, err
	}
	bot.User = user

	gameType, err := GetGameTypeById(gameTypeId)
	if err != nil {
		return Bot{}, err
	}

	bot.GameType = gameType

	return bot, nil
}

func ListBotsForGameType(gameType GameType) ([]Bot, error) {
	db := GetDatabaseConnection()
	defer db.Close()

	rows, err := db.Query(`
	SELECT
	  b.id
	, b.name
	, b.version
	, b.game_type_id
	, b.user_id
	FROM bot b
	WHERE b.game_type_id = $1
	`, gameType.Id)
	if err != nil {
		return []Bot{}, err
	}

	var botList []Bot
	for rows.Next() {
		var bot Bot
		var gameTypeId int
		var userId int
		err := rows.Scan(&bot.Id, &bot.Name, &bot.Version, &gameTypeId, &userId)
		if err != nil {
			return botList, err
		}
		bot.GameType, _ = GetGameTypeById(gameTypeId)
		bot.User, _ = GetUserById(userId)
		botList = append(botList, bot)
	}

	return botList, nil
}
