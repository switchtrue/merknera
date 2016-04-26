package repository

import (
	"log"

	"github.com/mleonard87/merknera/rpchelper"
)

type BotStatus string

const (
	BOT_STATUS_ONLINE  BotStatus = "ONLINE"
	BOT_STATUS_OFFLINE BotStatus = "OFFLINE"
	BOT_STATUS_ERROR   BotStatus = "ERROR"
)

type Bot struct {
	Id                  string
	Name                string
	Version             string
	GameType            GameType
	User                User
	RPCEndpoint         string
	ProgrammingLanguage string
	Website             string
	Status              BotStatus
}

// Ping will make an RPC call to the Status.Ping method. If this does not return
// then mark te bot as offline and will not participate in any further games until
// it is found to be online again.
func (b *Bot) Ping() bool {
	err := rpchelper.Ping(b.RPCEndpoint)
	if err != nil {
		// If we can't ping the bot, assume its offline and return.
		b.MarkOffline()
		return false
	}

	b.MarkOnline()
	return true
}

func (b *Bot) setStatus(status BotStatus) error {
	db := GetDB()
	err := db.QueryRow(`
	UPDATE bot
	SET status = $1
	WHERE id = $2
	`, string(status), b.Id).Scan()
	if err != nil {
		return err
	}

	return nil
}

func (b *Bot) MarkOffline() error {
	return b.setStatus(BOT_STATUS_OFFLINE)
}

func (b *Bot) MarkOnline() error {
	return b.setStatus(BOT_STATUS_ONLINE)
}

func (b *Bot) MarkError() error {
	return b.setStatus(BOT_STATUS_ERROR)
}

func RegisterBot(name string, version string, gameType GameType, user User, rpcEndpoint string, programmingLanguage string, website string) (Bot, error) {
	log.Println("RegisterBot")
	var botId int
	db := GetDB()
	err := db.QueryRow(`
	INSERT INTO bot (
	  name
	, version
	, game_type_id
	, user_id
	, rpc_endpoint
	, programming_language
	, website
	, status
	) VALUES (
	  $1
	, $2
	, $3
	, $4
	, $5
	, $6
	, $7
	, $8
	) RETURNING id
	`, name, version, gameType.Id, user.Id, rpcEndpoint, programmingLanguage, website, string(BOT_STATUS_ONLINE)).Scan(&botId)
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
	log.Println("GetBotById")
	var bot Bot
	var gameTypeId int
	var userId int
	var status string
	db := GetDB()
	err := db.QueryRow(`
	SELECT
	  id
	, name
	, version
	, game_type_id
	, user_id
	, rpc_endpoint
	, programming_language
	, website
	, status
	FROM bot
	WHERE id = $1
	`, id).Scan(&bot.Id, &bot.Name, &bot.Version, &gameTypeId, &userId, &bot.RPCEndpoint, &bot.ProgrammingLanguage, &bot.Website, &status)
	if err != nil {
		return Bot{}, err
	}
	bot.Status = BotStatus(status)

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
	log.Println("ListBotsForGameType")
	db := GetDB()
	rows, err := db.Query(`
	SELECT
	  b.id
	, b.name
	, b.version
	, b.game_type_id
	, b.user_id
	, b.rpc_endpoint
	, b.programming_language
	, b.website
	, b.status
	FROM bot b
	WHERE b.game_type_id = $1
	`, gameType.Id)
	if err != nil {
		return []Bot{}, err
	}

	var botList []Bot
	for rows.Next() {
		var bot Bot
		var status string
		err := rows.Scan(&bot.Id, &bot.Name, &bot.Version, &bot.GameType.Id, &bot.User.Id, &bot.RPCEndpoint, &bot.ProgrammingLanguage, &bot.Website, &status)
		if err != nil {
			return botList, err
		}
		bot.Status = BotStatus(status)
		botList = append(botList, bot)
	}

	for _, b := range botList {
		b.GameType, err = GetGameTypeById(b.GameType.Id)
		if err != nil {
			return []Bot{}, err
		}
		b.User, err = GetUserById(b.User.Id)
		if err != nil {
			return []Bot{}, err
		}
	}

	return botList, nil
}

func ListBots() ([]Bot, error) {
	log.Println("ListBots")
	db := GetDB()
	rows, err := db.Query(`
	SELECT
	  b.id
	, b.name
	, b.version
	, b.game_type_id
	, b.user_id
	, b.rpc_endpoint
	, b.programming_language
	, b.website
	, b.status
	FROM bot b
	`)
	if err != nil {
		return []Bot{}, err
	}

	var botList []Bot
	for rows.Next() {
		var bot Bot
		var status string
		err := rows.Scan(&bot.Id, &bot.Name, &bot.Version, &bot.GameType.Id, &bot.User.Id, &bot.RPCEndpoint, &bot.ProgrammingLanguage, &bot.Website, &status)
		if err != nil {
			return botList, err
		}
		bot.Status = BotStatus(status)
		botList = append(botList, bot)
	}

	for _, b := range botList {
		b.GameType, err = GetGameTypeById(b.GameType.Id)
		if err != nil {
			return []Bot{}, err
		}
		b.User, err = GetUserById(b.User.Id)
		if err != nil {
			return []Bot{}, err
		}
	}

	return botList, nil
}
