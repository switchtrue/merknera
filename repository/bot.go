package repository

import (
	"fmt"
	"log"
	"math"
	"time"

	"database/sql"

	"strings"

	"github.com/mleonard87/merknera/rpchelper"
)

type BotStatus string

const (
	BOT_STATUS_ONLINE     BotStatus = "ONLINE"
	BOT_STATUS_OFFLINE    BotStatus = "OFFLINE"
	BOT_STATUS_ERROR      BotStatus = "ERROR"
	BOT_STATUS_SUPERSEDED BotStatus = "SUPERSEDED"
)

type Bot struct {
	Id                     int `json:"id"`
	Name                   string
	Version                string
	gameTypeId             int
	gameType               GameType
	userId                 int
	user                   User
	RPCEndpoint            string
	ProgrammingLanguage    string
	Website                string
	Description            string
	Status                 BotStatus
	LastOnlineDateTime     time.Time
	gamesWonCountLoaded    bool
	gamesWonCount          int
	gamesDrawnCountLoaded  bool
	gamesDrawnCount        int
	gamesPlayedCountLoaded bool
	gamesPlayedCount       int
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
	b.Logf("Ping [BEGIN]: Endpoint %s", b.RPCEndpoint)
	err := rpchelper.Ping(b.RPCEndpoint)
	if err != nil {
		err2 := b.MarkOffline()
		if err2 != nil {
			log.Printf("An error occurred in bot.Ping():1:\n%s\n", err2)
			return false, err2
		}
		// This is actually fine. If we can't reach the bot it gets marked
		// as offline and all is good.
		b.Logf("Ping [END]: Offline: ", err.Error())
		return false, nil
	}

	err = b.MarkOnline()
	if err != nil {
		log.Printf("An error occurred in bot.Ping():3:\n%s\n", err)
		return false, err
	}

	b.Log("Ping [END]: Online")
	return true, nil
}

func (b *Bot) setStatus(status BotStatus) error {
	db := GetDB()

	var err error
	if status == BOT_STATUS_ONLINE || status == BOT_STATUS_ERROR {
		_, err = db.Exec(`
		UPDATE bot
		SET
		  status = $1
		, last_online_datetime = $2
		WHERE id = $3
		AND status != $4
		`, string(status), time.Now(), b.Id, string(BOT_STATUS_SUPERSEDED))
	} else {
		_, err = db.Exec(`
		UPDATE bot
		SET status = $1
		WHERE id = $2
		AND status != $3
		`, string(status), b.Id, string(BOT_STATUS_SUPERSEDED))
	}

	if err != nil {
		log.Printf("An error occurred in bot.setStatus():\n%s\n", err)
		return err
	}

	return nil
}

func (b *Bot) MarkOffline() error {
	err := b.setStatus(BOT_STATUS_OFFLINE)
	if err != nil {
		return err
	}
	b.Status = BOT_STATUS_OFFLINE
	return nil
}

func (b *Bot) MarkOnline() error {
	err := b.setStatus(BOT_STATUS_ONLINE)
	if err != nil {
		return err
	}
	b.Status = BOT_STATUS_ONLINE
	return nil
}

func (b *Bot) MarkError() error {
	err := b.setStatus(BOT_STATUS_ERROR)
	if err != nil {
		return err
	}
	b.Status = BOT_STATUS_ERROR
	return nil
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

func (b *Bot) GamesPlayed() ([]Game, error) {
	db := GetDB()
	rows, err := db.Query(`
	SELECT
	  g.id
	, g.game_type_id
	, g.status
	FROM game_bot gb
	JOIN game g
	  ON gb.game_id = g.id
	 AND g.status != $1
	WHERE bot_id = $2
	ORDER BY g.status
	`, string(GAME_STATUS_SUPERSEDED), b.Id)
	if err != nil {
		log.Printf("An error occurred in bot.GamesPlayedCount():\n%s\n", err)
		return []Game{}, err
	}

	var gameList []Game
	for rows.Next() {
		var game Game
		var status string
		err := rows.Scan(&game.Id, &game.gameTypeId, &status)
		if err != nil {
			log.Printf("An error occurred in bot.ListBotsForGameType():\n%s\n", err)
			return gameList, err
		}
		game.Status = GameStatus(status)
		gameList = append(gameList, game)
	}

	return gameList, nil
}

func (b *Bot) GamesPlayedCount() (int, error) {
	if b.gamesPlayedCountLoaded {
		return b.gamesPlayedCount, nil
	}
	var count int
	db := GetDB()
	err := db.QueryRow(`
	SELECT COUNT(*)
	FROM game_bot gb
	JOIN game g
	  ON gb.game_id = g.id
	 AND g.status = $1
	WHERE bot_id = $2
	`, string(GAME_STATUS_COMPLETE), b.Id).Scan(&count)
	if err != nil {
		if err == sql.ErrNoRows {
			return 0, nil
		}
		log.Printf("An error occurred in bot.GamesPlayedCount():\n%s\n", err)
		return 0, err
	}

	b.gamesPlayedCountLoaded = true
	b.gamesPlayedCount = count

	return count, nil
}

func (b *Bot) GamesWonCount() (int, error) {
	if b.gamesWonCountLoaded {
		return b.gamesWonCount, nil
	}
	var count int
	db := GetDB()
	err := db.QueryRow(`
	SELECT COUNT(*)
	FROM game_bot gb
	JOIN move m
	  ON gb.id = m.game_bot_id
	 AND m.winner = true
	WHERE gb.bot_id = $1
	`, b.Id).Scan(&count)
	if err != nil {
		if err == sql.ErrNoRows {
			return 0, nil
		}
		log.Printf("An error occurred in bot.GamesWonCount():\n%s\n", err)
		return 0, err
	}

	b.gamesWonCountLoaded = true
	b.gamesWonCount = count

	return count, nil
}

func (b *Bot) GamesDrawnCount() (int, error) {
	if b.gamesDrawnCountLoaded {
		return b.gamesDrawnCount, nil
	}
	var count int
	db := GetDB()
	err := db.QueryRow(`
	SELECT COUNT(t.*) games_drawn_count
	FROM (
	  SELECT
	    g.id
	  FROM game_bot gb
	  JOIN game g
	    ON gb.game_id = g.id
	    AND g.status = $1
	  JOIN game_bot gb2
	    ON g.id = gb2.game_id
	  JOIN move m
	    ON gb2.id = m.game_bot_id
	  WHERE gb.bot_id = $2
	  GROUP BY g.id
	  HAVING
	    MAX(
	      CASE m.winner
		WHEN TRUE THEN 1
		ELSE 0
	      END
	    ) = 0
	) t
	`, string(GAME_STATUS_COMPLETE), b.Id).Scan(&count)
	if err != nil {
		if err == sql.ErrNoRows {
			return 0, nil
		}
		log.Printf("An error occurred in bot.GamesDrawnCount():\n%s\n", err)
		return 0, err
	}

	b.gamesDrawnCountLoaded = true
	b.gamesDrawnCount = count

	return count, nil
}

func (b *Bot) Update(rpcEndpoint string, programmingLanguage string, website string, description string) error {
	db := GetDB()
	_, err := db.Exec(`
	UPDATE bot
	SET
	  rpc_endpoint = $1
	, programming_language = $2
	, website = $3
	, description = $4
	WHERE id = $5
	AND status != $6
	`, rpcEndpoint, programmingLanguage, website, strings.Trim(description, " "), b.Id, string(BOT_STATUS_SUPERSEDED))
	if err != nil {
		log.Printf("An error occurred in bot.Update():1:\n%s\n", err)
		return err
	}

	b.RPCEndpoint = rpcEndpoint
	b.ProgrammingLanguage = programmingLanguage
	b.Website = website
	b.Description = description

	return nil
}

func (b *Bot) ListAwaitingMoves() ([]GameMove, error) {
	db := GetDB()
	rows, err := db.Query(`
	SELECT
	  m.id
	, m.game_bot_id
	, m.status
	, m.winner
	FROM game_bot gb
	JOIN move m
	ON gb.id = m.game_bot_id
	AND m.status = $1
	WHERE gb.bot_id = $2
	`, string(GAMEMOVE_STATUS_AWAITING), b.Id)
	if err != nil {
		return []GameMove{}, err
	}

	var gameMoveList []GameMove
	for rows.Next() {
		var gameMove GameMove
		var status string
		err := rows.Scan(&gameMove.Id, &gameMove.gameBotId, &status, &gameMove.Winner)
		if err != nil {
			log.Printf("An error occurred in bot.ListBotsForGameType():\n%s\n", err)
			return gameMoveList, err
		}
		gameMove.Status = GameMoveStatus(status)
		gameMoveList = append(gameMoveList, gameMove)
	}

	return gameMoveList, nil
}

func (b *Bot) CurrentScore() (float64, error) {
	played, err := b.GamesPlayedCount()
	if err != nil {
		played = 0
	}

	if played == 0 {
		return 0, err
	}

	won, err := b.GamesWonCount()
	if err != nil {
		won = 0
	}

	drawn, err := b.GamesDrawnCount()
	if err != nil {
		drawn = 0
	}

	score := ((float64(won) + float64(drawn)) / float64(played)) * 100.0

	shift := math.Pow(10, float64(2))
	rounded := math.Floor((score*shift)+.5) / shift

	return rounded, nil
}

func RegisterBot(name string, version string, gameType GameType, user User, rpcEndpoint string, programmingLanguage string, website string, description string) (Bot, error) {
	var botId int
	db := GetDB()

	tx, err := db.Begin()
	if err != nil {
		log.Printf("An error occurred in bot.RegisterBot():1:\n%s\n", err)
		return Bot{}, err
	}

	_, err = tx.Exec(`
	UPDATE bot
	SET status = $1
	WHERE name = $2
	`, string(BOT_STATUS_SUPERSEDED), name)
	if err != nil {
		log.Printf("An error occurred in bot.RegisterBot():2:\n%s\n", err)
		tx.Rollback()
		return Bot{}, err
	}

	_, err = tx.Exec(`
	UPDATE game
	SET status = $1
	WHERE id IN (
	  SELECT gb.game_id
	  FROM bot b
	  JOIN game_bot gb
	    ON b.id = gb.bot_id
	  WHERE b.name = $2
	)
	AND status != $3
	`, string(GAME_STATUS_SUPERSEDED), strings.Trim(name, " "), string(GAME_STATUS_COMPLETE))
	if err != nil {
		log.Printf("An error occurred in bot.RegisterBot():3:\n%s\n", err)
		tx.Rollback()
		return Bot{}, err
	}

	_, err = tx.Exec(`
	UPDATE move
	SET status = $1
	WHERE game_bot_id IN (
	  SELECT gb2.id
	  FROM bot b
	  JOIN game_bot gb1
	    ON b.id = gb1.bot_id
	  JOIN game_bot gb2
	    ON gb1.game_id = gb2.game_id
	  WHERE b.name = $2
	)
	AND status != $3
	`, string(GAMEMOVE_STATUS_SUPERSEDED), strings.Trim(name, " "), string(GAMEMOVE_STATUS_COMPLETE))
	if err != nil {
		log.Printf("An error occurred in bot.RegisterBot():4:\n%s\n", err)
		tx.Rollback()
		return Bot{}, err
	}

	err = tx.QueryRow(`
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
	`, strings.Trim(name, " "), strings.Trim(version, " "), gameType.Id, user.Id, rpcEndpoint, programmingLanguage, website, strings.Trim(description, " "), string(BOT_STATUS_ONLINE)).Scan(&botId)
	if err != nil {
		log.Printf("An error occurred in bot.RegisterBot():5:\n%s\n", err)
		tx.Rollback()
		return Bot{}, err
	}

	tx.Commit()

	bot, err := GetBotById(botId)
	if err != nil {
		log.Printf("An error occurred in bot.RegisterBot():6:\n%s\n", err)
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
	, last_online_datetime
	FROM bot
	WHERE id = $1
	`, id).Scan(&bot.Id, &bot.Name, &bot.Version, &bot.gameTypeId, &bot.userId, &bot.RPCEndpoint, &bot.ProgrammingLanguage, &bot.Website, &bot.Description, &status, &bot.LastOnlineDateTime)
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
	, last_online_datetime
	FROM bot
	WHERE name = $1
	AND status != $2
	`, name, string(BOT_STATUS_SUPERSEDED)).Scan(&bot.Id, &bot.Name, &bot.Version, &bot.gameTypeId, &bot.userId, &bot.RPCEndpoint, &bot.ProgrammingLanguage, &bot.Website, &bot.Description, &status, &bot.LastOnlineDateTime)
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
	, b.last_online_datetime
	FROM bot b
	WHERE b.game_type_id = $1
	AND b.status != $2
	ORDER BY b.name, b.version
	`, gameType.Id, string(BOT_STATUS_SUPERSEDED))
	if err != nil {
		return []Bot{}, err
	}

	var botList []Bot
	for rows.Next() {
		var bot Bot
		var status string
		err := rows.Scan(&bot.Id, &bot.Name, &bot.Version, &bot.gameTypeId, &bot.userId, &bot.RPCEndpoint, &bot.ProgrammingLanguage, &bot.Website, &bot.Description, &status, &bot.LastOnlineDateTime)
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
	, b.last_online_datetime
	FROM bot b
	WHERE status != $1
	ORDER BY b.name, b.version
	`, string(BOT_STATUS_SUPERSEDED))
	if err != nil {
		log.Printf("An error occurred in bot.ListBots():1:\n%s\n", err)
		return []Bot{}, err
	}

	var botList []Bot
	for rows.Next() {
		var bot Bot
		var status string
		err := rows.Scan(&bot.Id, &bot.Name, &bot.Version, &bot.gameTypeId, &bot.userId, &bot.RPCEndpoint, &bot.ProgrammingLanguage, &bot.Website, &bot.Description, &status, &bot.LastOnlineDateTime)
		if err != nil {
			log.Printf("An error occurred in bot.ListBots():2:\n%s\n", err)
			return botList, err
		}
		bot.Status = BotStatus(status)
		botList = append(botList, bot)
	}

	return botList, nil
}
