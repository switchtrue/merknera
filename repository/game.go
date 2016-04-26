package repository

import (
	"encoding/json"
	"errors"
	"fmt"
	"log"
)

type GameStatus string

type Game struct {
	Id       int
	GameType GameType
	Status   GameStatus
}

const (
	GAME_STATUS_NOT_STARTED GameStatus = "NOT STARTED"
	GAME_STATUS_IN_PROGRESS GameStatus = "IN PROGRESS"
	GAME_STATUS_COMPLETE    GameStatus = "COMPLETE"
)

func (g *Game) GetNextMoveId() (int, error) {
	db := GetDB()
	log.Println("GetNextMoveId")
	var nextMoveId int
	err := db.QueryRow(`
	SELECT m.id
	FROM game g
	JOIN game_bot gb
	  ON g.id = gb.game_id
	JOIN move m
	  ON gb.id = m.game_bot_id
	WHERE g.id = $1
	AND m.status = $2
	`, g.Id, string(GAMEMOVE_STATUS_AWAITING)).Scan(&nextMoveId)
	if err != nil {
		return -1, errors.New("GetNextMoveId: " + err.Error())
	}

	return nextMoveId, nil
}

func (g *Game) NextGameMove() (GameMove, error) {
	fmt.Println("NextGameMove")
	moveId, err := g.GetNextMoveId()
	if err != nil {
		return GameMove{}, errors.New("NextGameMove: " + err.Error())
	}

	gameMove, err := GetGameMoveById(moveId)
	if err != nil {
		return gameMove, errors.New("NextGameMove: " + err.Error())
	}

	return gameMove, nil
}

func (g *Game) GetWinningMoveId() (int, error) {
	db := GetDB()
	log.Println("GetWinningMoveId")
	var nextMoveId int
	err := db.QueryRow(`
	SELECT
	  m.id
	FROM game g
	JOIN game_bot gb
	  ON g.id = gb.game_id
	JOIN move m
	  ON gb.id = m.game_bot_id
	WHERE g.id = $1
	ORDER BY m.created_datetime DESC
	LIMIT 1
	`, g.Id).Scan(&nextMoveId)
	if err != nil {
		return -1, errors.New("GetWinningMoveId: " + err.Error())
	}

	return nextMoveId, nil
}

func (g *Game) WinningMove() (GameMove, error) {
	fmt.Println("WinningMove")
	moveId, err := g.GetWinningMoveId()
	if err != nil {
		return GameMove{}, errors.New("WinningMove: " + err.Error())
	}

	gameMove, err := GetGameMoveById(moveId)
	if err != nil {
		return gameMove, errors.New("WinningMove: " + err.Error())
	}

	return gameMove, nil
}

func (g *Game) Players() ([]GameBot, error) {
	db := GetDB()
	rows, err := db.Query(`
	SELECT
	  id
	FROM game_bot
	WHERE game_id = $1
	ORDER BY play_sequence
	`, g.Id)

	var gameBotList []GameBot
	for rows.Next() {
		var gameBot GameBot
		var gameBotId int
		rows.Scan(&gameBotId)
		gameBot, err = GetGameBotById(gameBotId)
		if err != nil {
			return []GameBot{}, err
		}
		gameBotList = append(gameBotList, gameBot)
	}

	return gameBotList, nil
}

func (g *Game) GameState() (string, error) {
	db := GetDB()
	var gs string
	err := db.QueryRow(`
	SELECT
	  state
	FROM game
	WHERE id = $1
	`, g.Id).Scan(&gs)
	if err != nil {
		return "", err
	}

	return gs, nil
}

func (g *Game) SetGameState(gs interface{}) error {
	gsB, err := json.Marshal(gs)
	if err != nil {
		return err
	}

	db := GetDB()
	_, err = db.Exec(`
	UPDATE game
	SET state = $1
	WHERE id = $2
	`, string(gsB), g.Id)
	if err != nil {
		return err
	}

	return nil
}

func (g *Game) setStatus(status GameStatus) error {
	db := GetDB()
	_, err := db.Exec(`
	UPDATE game
	SET status = $1
	WHERE id = $2
	`, string(status), g.Id)
	if err != nil {
		return err
	}

	return nil
}

func (g *Game) MarkInProgress() error {
	return g.setStatus(GAME_STATUS_IN_PROGRESS)
}

func (g *Game) MarkComplete() error {
	return g.setStatus(GAME_STATUS_COMPLETE)
}

func CreateGame(gameType GameType, initialGameState interface{}) (Game, error) {
	log.Println("CreateGame")
	var gameId int

	igsB, err := json.Marshal(initialGameState)
	if err != nil {
		return Game{}, err
	}

	db := GetDB()
	err = db.QueryRow(`
	INSERT INTO game (
	  game_type_id
	, state
	) VALUES (
	  $1
	, $2
	) RETURNING id
	`, gameType.Id, string(igsB)).Scan(&gameId)
	if err != nil {
		return Game{}, err
	}

	game, err := GetGameById(gameId)
	if err != nil {
		return game, err
	}
	return game, nil
}

func GetGameById(id int) (Game, error) {
	log.Println("GetGameById")
	var game Game
	var gameTypeId int
	var status string
	db := GetDB()
	err := db.QueryRow(`
	SELECT
	  g.id
	, g.status
	, g.game_type_id
	FROM game g
	WHERE g.id = $1
	`, id).Scan(&game.Id, &status, &gameTypeId)
	if err != nil {
		return Game{}, err
	}
	game.Status = GameStatus(status)
	game.GameType, _ = GetGameTypeById(gameTypeId)

	return game, nil
}
