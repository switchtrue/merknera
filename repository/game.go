package repository

import (
	"errors"
	"log"
)

type GameStatus string

type Game struct {
	Id         int
	gameTypeId int
	gameType   GameType
	Status     GameStatus
}

const (
	GAME_STATUS_NOT_STARTED GameStatus = "NOT STARTED"
	GAME_STATUS_IN_PROGRESS GameStatus = "IN PROGRESS"
	GAME_STATUS_COMPLETE    GameStatus = "COMPLETE"
)

func (g *Game) GameType() (GameType, error) {
	if g.gameType == (GameType{}) {
		gt, err := GetGameTypeById(g.gameTypeId)
		if err != nil {
			log.Printf("An error occurred in game.GameType():\n%s\n", err)
			return GameType{}, err
		}
		g.gameType = gt
	}

	return g.gameType, nil
}

func (g *Game) GetNextMoveId() (int, error) {
	db := GetDB()
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
		log.Printf("An error occurred in game.GetNextMoveId():\n%s\n", err)
		return -1, err
	}

	return nextMoveId, nil
}

func (g *Game) NextGameMove() (GameMove, error) {
	moveId, err := g.GetNextMoveId()
	if err != nil {
		log.Printf("An error occurred in game.NextGameMove():1:\n%s\n", err)
		return GameMove{}, err
	}

	gameMove, err := GetGameMoveById(moveId)
	if err != nil {
		log.Printf("An error occurred in game.NextGameMove():2:\n%s\n", err)
		return gameMove, err
	}

	return gameMove, nil
}

func (g *Game) GetWinningMoveId() (int, error) {
	db := GetDB()
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
		log.Printf("An error occurred in game.GetWinningMoveId():\n%s\n", err)
		return -1, err
	}

	return nextMoveId, nil
}

func (g *Game) WinningMove() (GameMove, error) {
	if g.Status != GAME_STATUS_COMPLETE {
		return GameMove{}, errors.New("This game is not yet complete. You should not call WinningMove() on an uncomplete game.")
	}
	moveId, err := g.GetWinningMoveId()
	if err != nil {
		log.Printf("An error occurred in game.GetWinningMoveId():1:\n%s\n", err)
		return GameMove{}, err
	}

	gameMove, err := GetGameMoveById(moveId)
	if err != nil {
		log.Printf("An error occurred in game.GetWinningMoveId():2:\n%s\n", err)
		return gameMove, err
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
			log.Printf("An error occurred in game.Players():\n%s\n", err)
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
	  m.game_state
	FROM game_bot gb
	JOIN move m
	  ON gb.id = m.game_bot_id
	WHERE gb.game_id = $1
	ORDER BY m.created_datetime DESC
	LIMIT 1
	`, g.Id).Scan(&gs)
	if err != nil {
		log.Printf("An error occurred in game.GameState():\n%s\n", err)
		return "", err
	}

	return gs, nil
}

func (g *Game) setStatus(status GameStatus) error {
	db := GetDB()
	_, err := db.Exec(`
	UPDATE game
	SET status = $1
	WHERE id = $2
	`, string(status), g.Id)
	if err != nil {
		log.Printf("An error occurred in game.setStatus():\n%s\n", err)
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

func (g *Game) Moves() ([]GameMove, error) {
	db := GetDB()
	rows, err := db.Query(`
	SELECT
	  m.id
	, m.game_bot_id
	, m.status
	FROM game_bot gb
	JOIN move m
	  ON gb.id = m.game_bot_id
	WHERE gb.game_id = $1
	ORDER BY m.created_datetime
	`, g.Id)
	if err != nil {
		log.Printf("An error occurred in game.Moves():1:\n%s\n", err)
		return []GameMove{}, err
	}

	var gameMoves []GameMove
	for rows.Next() {
		var gm GameMove
		var status string
		err := rows.Scan(&gm.Id, &gm.gameBotId, &status)
		if err != nil {
			log.Printf("An error occurred in game.Moves():2:\n%s\n", err)
			return gameMoves, err
		}
		gm.Status = GameMoveStatus(status)
		gameMoves = append(gameMoves, gm)
	}

	return gameMoves, nil
}

func CreateGame(gameType GameType) (Game, error) {
	var gameId int

	db := GetDB()
	err := db.QueryRow(`
	INSERT INTO game (
	  game_type_id
	) VALUES (
	  $1
	) RETURNING id
	`, gameType.Id).Scan(&gameId)
	if err != nil {
		log.Printf("An error occurred in game.CreateGame():2:\n%s\n", err)
		return Game{}, err
	}

	game, err := GetGameById(gameId)
	if err != nil {
		log.Printf("An error occurred in game.CreateGame():3:\n%s\n", err)
		return game, err
	}
	return game, nil
}

func GetGameById(id int) (Game, error) {
	var game Game
	var status string
	db := GetDB()
	err := db.QueryRow(`
	SELECT
	  g.id
	, g.status
	, g.game_type_id
	FROM game g
	WHERE g.id = $1
	`, id).Scan(&game.Id, &status, &game.gameTypeId)
	if err != nil {
		log.Printf("An error occurred in game.GetGameById():\n%s\n", err)
		return Game{}, err
	}
	game.Status = GameStatus(status)

	return game, nil
}

func ListGames() ([]Game, error) {
	db := GetDB()
	rows, err := db.Query(`
	SELECT
	  g.id
	, g.game_type_id
	, g.status
	FROM game g
	`)
	if err != nil {
		log.Printf("An error occurred in game.ListGames():1:\n%s\n", err)
		return []Game{}, err
	}

	var gameList []Game
	for rows.Next() {
		var game Game
		var status string
		err := rows.Scan(&game.Id, &game.gameTypeId, &status)
		if err != nil {
			log.Printf("An error occurred in game.ListGames():2:\n%s\n", err)
			return gameList, err
		}
		game.Status = GameStatus(status)
		gameList = append(gameList, game)
	}

	return gameList, nil
}
