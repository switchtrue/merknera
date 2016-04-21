package repository

import (
	"database/sql"
	"fmt"
	"log"
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
	FROM bot
	WHERE id = $1
	`, id).Scan(&bot.Id, &bot.Name, &bot.Version, &gameTypeId, &userId, &bot.RPCEndpoint, &bot.ProgrammingLanguage, &bot.Website)
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
	FROM bot b
	WHERE b.game_type_id = $1
	`, gameType.Id)
	if err != nil {
		return []Bot{}, err
	}

	var botList []Bot
	for rows.Next() {
		var bot Bot
		err := rows.Scan(&bot.Id, &bot.Name, &bot.Version, &bot.GameType.Id, &bot.User.Id, &bot.RPCEndpoint, &bot.ProgrammingLanguage, &bot.Website)
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

	fmt.Println("***")
	fmt.Println(botList)
	fmt.Println("---")

	return botList, nil
}
