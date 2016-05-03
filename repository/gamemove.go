package repository

import (
	"encoding/json"
	"log"
)

type GameMove struct {
	Id        int
	gameBotId int
	gameBot   GameBot
	Status    GameMoveStatus
	Winner    bool
}

type GameMoveStatus string

const (
	GAMEMOVE_STATUS_AWAITING   GameMoveStatus = "AWAITING"
	GAMEMOVE_STATUS_COMPLETE   GameMoveStatus = "COMPLETE"
	GAMEMOVE_STATUS_SUPERSEDED GameMoveStatus = "SUPERSEDED"
)

func (gm *GameMove) GameBot() (GameBot, error) {
	if gm.gameBot == (GameBot{}) {
		gb, err := GetGameBotById(gm.gameBotId)
		if err != nil {
			log.Printf("An error occurred in gamemove.GameBot():\n%s\n", err)
			return GameBot{}, err
		}
		gm.gameBot = gb
	}

	return gm.gameBot, nil
}

func (gm *GameMove) MarkComplete() error {
	db := GetDB()
	_, err := db.Exec(`
	UPDATE move
	SET status = $1
	WHERE id = $2
	AND status != $3
	`, string(GAMEMOVE_STATUS_COMPLETE), gm.Id, string(GAME_STATUS_SUPERSEDED))
	if err != nil {
		log.Printf("An error occurred in gamemove.MarkComplete():\n%s\n", err)
		return err
	}

	return nil
}

func (gm *GameMove) MarkSuperseded() error {
	db := GetDB()
	_, err := db.Exec(`
	UPDATE move
	SET status = $1
	WHERE id = $2
	AND status != $3
	`, string(GAMEMOVE_STATUS_COMPLETE), gm.Id, string(GAME_STATUS_SUPERSEDED))
	if err != nil {
		log.Printf("An error occurred in gamemove.MarkComplete():\n%s\n", err)
		return err
	}

	return nil
}

func (gm *GameMove) MarkAsWin() error {
	db := GetDB()
	_, err := db.Exec(`
	UPDATE move
	SET winner = true
	WHERE id = $1
	`, gm.Id)
	if err != nil {
		log.Printf("An error occurred in gamemove.MarkAsWin():\n%s\n", err)
		return err
	}

	gm.Winner = true

	return nil
}

func (gm *GameMove) SetGameState(gs interface{}) error {
	gsB, err := json.Marshal(gs)
	if err != nil {
		log.Printf("An error occurred in gamemove.SetGameState():1:\n%s\n", err)
		return err
	}

	db := GetDB()
	_, err = db.Exec(`
	UPDATE move
	SET game_state = $1
	WHERE id = $2
	`, string(gsB), gm.Id)
	if err != nil {
		log.Printf("An error occurred in gamemove.SetGameState():2:\n%s\n", err)
		return err
	}

	return nil
}

func (gm *GameMove) GameState() (string, error) {
	db := GetDB()
	var gs string
	err := db.QueryRow(`
	SELECT
	  m.game_state
	FROM move m
	WHERE m.id = $1
	`, gm.Id).Scan(&gs)
	if err != nil {
		log.Printf("An error occurred in gamemove.GameState():\n%s\n", err)
		return "", err
	}

	return gs, nil
}

func CreateGameMove(gameBot GameBot, currentGameState interface{}) (GameMove, error) {
	gsB, err := json.Marshal(currentGameState)
	if err != nil {
		log.Printf("An error occurred in gamemove.CreateGameMove():1:\n%s\n", err)
		return GameMove{}, err
	}

	var gameMoveId int
	db := GetDB()
	err = db.QueryRow(`
	INSERT INTO move (
	  game_bot_id
	, game_state
	) VALUES (
	  $1
	, $2
	) RETURNING id
	`, gameBot.Id, string(gsB)).Scan(&gameMoveId)
	if err != nil {
		log.Printf("An error occurred in gamemove.CreateGameMove():2:\n%s\n", err)
		return GameMove{}, err
	}

	gameMove, err := GetGameMoveById(gameMoveId)
	if err != nil {
		log.Printf("An error occurred in gamemove.CreateGameMove():3:\n%s\n", err)
		return GameMove{}, err
	}
	return gameMove, nil
}

func GetGameMoveById(id int) (GameMove, error) {
	var gameMove GameMove
	var status string
	db := GetDB()
	err := db.QueryRow(`
	SELECT
	  id
	, game_bot_id
	, status
	, winner
	FROM move
	WHERE id = $1
	`, id).Scan(&gameMove.Id, &gameMove.gameBotId, &status, &gameMove.Winner)
	if err != nil {
		log.Printf("An error occurred in gamemove.GetGameMoveById():\n%s\n", err)
		return GameMove{}, err
	}

	gameMove.Status = GameMoveStatus(status)

	return gameMove, nil
}

func ListAwaitingMoves() ([]GameMove, error) {
	db := GetDB()
	rows, err := db.Query(`
	SELECT id
	FROM move
	WHERE status = $1
	`, string(GAMEMOVE_STATUS_AWAITING))
	if err != nil {
		log.Printf("An error occurred in gamemove.GetAwaitingMoves():1:\n%s\n", err)
		return []GameMove{}, err
	}

	var gameMoves []GameMove
	for rows.Next() {
		var gameMoveId int
		rows.Scan(&gameMoveId)
		gameMove, err := GetGameMoveById(gameMoveId)
		if err != nil {
			log.Printf("An error occurred in gamemove.GetAwaitingMoves():2:\n%s\n", err)
			return []GameMove{}, err
		}
		gameMoves = append(gameMoves, gameMove)
	}

	return gameMoves, nil
}
