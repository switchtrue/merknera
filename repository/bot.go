package repository

import (
	"database/sql"
	"log"

	"github.com/mleonard87/merknera/rpchelper"
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
	Status              string
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

func (b *Bot) MarkOffline() error {
	db := GetDB()
	err := db.QueryRow(`
	UPDATE bot
	SET status = 'OFFLINE'
	WHERE id = $1
	`, b.Id).Scan()
	if err != nil {
		return err
	}

	return nil
}

func (b *Bot) MarkOnline() error {
	db := GetDB()
	err := db.QueryRow(`
	UPDATE bot
	SET status = 'ONLINE'
	WHERE id = $1
	`, b.Id).Scan()
	if err != nil {
		return err
	}

	return nil
}

func RegisterBot(db *sql.DB, name string, version string, gameType GameType, user User, rpcEndpoint string, programmingLanguage string, website string) (Bot, error) {
	log.Println("RegisterBot")
	var botId int
	err := db.QueryRow(`
	INSERT INTO bot (
	  name
	, version
	, game_type_id
	, user_id
	, rpc_endpoint
	, programming_language
	, website
	) VALUES (
	  $1
	, $2
	, $3
	, $4
	, $5
	, $6
	, $7
	) RETURNING id
	`, name, version, gameType.Id, user.Id, rpcEndpoint, programmingLanguage, website).Scan(&botId)
	if err != nil {
		return Bot{}, err
	}

	bot, err := GetBotById(db, botId)
	if err != nil {
		return Bot{}, err
	}
	return bot, nil
}

func GetBotById(db *sql.DB, id int) (Bot, error) {
	log.Println("GetBotById")
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
	, rpc_endpoint
	, programming_language
	, website
	, status
	FROM bot
	WHERE id = $1
	`, id).Scan(&bot.Id, &bot.Name, &bot.Version, &gameTypeId, &userId, &bot.RPCEndpoint, &bot.ProgrammingLanguage, &bot.Website, &bot.Status)
	if err != nil {
		return Bot{}, err
	}

	user, err := GetUserById(db, userId)
	if err != nil {
		return Bot{}, err
	}
	bot.User = user

	gameType, err := GetGameTypeById(db, gameTypeId)
	if err != nil {
		return Bot{}, err
	}

	bot.GameType = gameType

	return bot, nil
}

func ListBotsForGameType(db *sql.DB, gameType GameType) ([]Bot, error) {
	log.Println("ListBotsForGameType")
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
		err := rows.Scan(&bot.Id, &bot.Name, &bot.Version, &bot.GameType.Id, &bot.User.Id, &bot.RPCEndpoint, &bot.ProgrammingLanguage, &bot.Website, &bot.Status)
		if err != nil {
			return botList, err
		}
		botList = append(botList, bot)
	}

	for _, b := range botList {
		b.GameType, err = GetGameTypeById(db, b.GameType.Id)
		if err != nil {
			return []Bot{}, err
		}
		b.User, err = GetUserById(db, b.User.Id)
		if err != nil {
			return []Bot{}, err
		}
	}

	return botList, nil
}

func ListBots(db *sql.DB) ([]Bot, error) {
	log.Println("ListBots")
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
		err := rows.Scan(&bot.Id, &bot.Name, &bot.Version, &bot.GameType.Id, &bot.User.Id, &bot.RPCEndpoint, &bot.ProgrammingLanguage, &bot.Website, &bot.Status)
		if err != nil {
			return botList, err
		}
		botList = append(botList, bot)
	}

	for _, b := range botList {
		b.GameType, err = GetGameTypeById(db, b.GameType.Id)
		if err != nil {
			return []Bot{}, err
		}
		b.User, err = GetUserById(db, b.User.Id)
		if err != nil {
			return []Bot{}, err
		}
	}

	return botList, nil
}
