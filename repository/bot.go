package repository

import (
	"log"

	"database/sql"

	"github.com/mleonard87/merknera/rpchelper"
)

type BotStatus string

const (
	BOT_STATUS_ONLINE  BotStatus = "ONLINE"
	BOT_STATUS_OFFLINE BotStatus = "OFFLINE"
	BOT_STATUS_ERROR   BotStatus = "ERROR"
)

type Bot struct {
	Id                  int
	Name                string
	Version             string
	gameTypeId          int
	gameType            GameType
	userId              int
	user                User
	RPCEndpoint         string
	ProgrammingLanguage string
	Website             string
	Description         string
	Status              BotStatus
}

func (b *Bot) GameType() (GameType, error) {
	if b.gameType == (GameType{}) {
		gt, err := GetGameTypeById(b.gameTypeId)
		if err != nil {
			log.Printf("An error occurred in bot.GameType():\n%s\n", err)
			return GameType{}, err
		}
		b.gameType = gt
	}

	return b.gameType, nil
}

func (b *Bot) User() (User, error) {
	if b.user == (User{}) {
		u, err := GetUserById(b.userId)
		if err != nil {
			log.Printf("An error occurred in bot.User():\n%s\n", err)
			return User{}, err
		}
		b.user = u
	}

	return b.user, nil
}

// Ping will make an RPC call to the Status.Ping method. If this does not return
// then mark te bot as offline and will not participate in any further games until
// it is found to be online again.
func (b *Bot) Ping() (bool, error) {
	err := rpchelper.Ping(b.RPCEndpoint)
	if err != nil {
		err2 := b.MarkOffline()
		if err2 != nil {
			log.Printf("An error occurred in bot.Ping():1:\n%s\n", err2)
			return false, err2
		}
		log.Printf("An error occurred in bot.Ping():2:\n%s\n", err)
		return false, err
	}

	err = b.MarkOnline()
	if err != nil {
		log.Printf("An error occurred in bot.Ping():3:\n%s\n", err)
		return false, err
	}

	return true, nil
}

func (b *Bot) setStatus(status BotStatus) error {
	db := GetDB()
	_, err := db.Exec(`
	UPDATE bot
	SET status = $1
	WHERE id = $2
	`, string(status), b.Id)
	if err != nil {
		log.Printf("An error occurred in bot.setStatus():\n%s\n", err)
		return err
	}

	return nil
}

func (b *Bot) MarkOffline() error {
	b.Status = BOT_STATUS_OFFLINE
	return b.setStatus(BOT_STATUS_OFFLINE)
}

func (b *Bot) MarkOnline() error {
	b.Status = BOT_STATUS_ONLINE
	return b.setStatus(BOT_STATUS_ONLINE)
}

func (b *Bot) MarkError() error {
	b.Status = BOT_STATUS_ERROR
	return b.setStatus(BOT_STATUS_ERROR)
}

func (b *Bot) DoesVersionExist(version string) (bool, error) {
	var botId int
	db := GetDB()
	err := db.QueryRow(`
	SELECT
	  id
	FROM bot
	WHERE name = $1
	AND version = $2
	`, b.Name, version).Scan(&botId)
	if err != nil {
		if err == sql.ErrNoRows {
			return false, nil
		}
		log.Printf("An error occurred in bot.DoesVersionExist():\n%s\n", err)
		return true, err
	}

	return true, nil
}

func RegisterBot(name string, version string, gameType GameType, user User, rpcEndpoint string, programmingLanguage string, website string, description string) (Bot, error) {
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
	, description
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
	, $9
	) RETURNING id
	`, name, version, gameType.Id, user.Id, rpcEndpoint, programmingLanguage, website, description, string(BOT_STATUS_ONLINE)).Scan(&botId)
	if err != nil {
		log.Printf("An error occurred in bot.RegisterBot():1:\n%s\n", err)
		return Bot{}, err
	}

	bot, err := GetBotById(botId)
	if err != nil {
		log.Printf("An error occurred in bot.RegisterBot():2:\n%s\n", err)
		return Bot{}, err
	}
	return bot, nil
}

func GetBotById(id int) (Bot, error) {
	var bot Bot
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
	, description
	, status
	FROM bot
	WHERE id = $1
	`, id).Scan(&bot.Id, &bot.Name, &bot.Version, &bot.gameTypeId, &bot.userId, &bot.RPCEndpoint, &bot.ProgrammingLanguage, &bot.Website, &bot.Description, &status)
	if err != nil {
		log.Printf("An error occurred in bot.GetBotById():\n%s\n", err)
		return Bot{}, err
	}
	bot.Status = BotStatus(status)

	return bot, nil
}

func GetBotByName(name string) (Bot, error) {
	var bot Bot
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
	, description
	, status
	FROM bot
	WHERE name = $1
	`, name).Scan(&bot.Id, &bot.Name, &bot.Version, &bot.gameTypeId, &bot.userId, &bot.RPCEndpoint, &bot.ProgrammingLanguage, &bot.Website, &bot.Description, &status)
	if err != nil {
		if err != sql.ErrNoRows {
			log.Printf("An error occurred in bot.GetBotByName():\n%s\n", err)
		}
		return Bot{}, err
	}
	bot.Status = BotStatus(status)

	return bot, nil
}

func ListBotsForGameType(gameType GameType) ([]Bot, error) {
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
	, b.description
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
		err := rows.Scan(&bot.Id, &bot.Name, &bot.Version, &bot.gameTypeId, &bot.userId, &bot.RPCEndpoint, &bot.ProgrammingLanguage, &bot.Website, &bot.Description, &status)
		if err != nil {
			log.Printf("An error occurred in bot.ListBotsForGameType():\n%s\n", err)
			return botList, err
		}
		bot.Status = BotStatus(status)
		botList = append(botList, bot)
	}

	return botList, nil
}

func ListBots() ([]Bot, error) {
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
	, b.description
	, b.status
	FROM bot b
	`)
	if err != nil {
		log.Printf("An error occurred in bot.ListBots():1:\n%s\n", err)
		return []Bot{}, err
	}

	var botList []Bot
	for rows.Next() {
		var bot Bot
		var status string
		err := rows.Scan(&bot.Id, &bot.Name, &bot.Version, &bot.gameTypeId, &bot.userId, &bot.RPCEndpoint, &bot.ProgrammingLanguage, &bot.Website, &bot.Description, &status)
		if err != nil {
			log.Printf("An error occurred in bot.ListBots():2:\n%s\n", err)
			return botList, err
		}
		bot.Status = BotStatus(status)
		botList = append(botList, bot)
	}

	return botList, nil
}
