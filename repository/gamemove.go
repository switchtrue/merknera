package repository

import "log"

type GameMove struct {
	Id        int
	gameBotId int
	gameBot   GameBot
	Status    GameMoveStatus
}

type GameMoveStatus string

const (
	GAMEMOVE_STATUS_AWAITING GameMoveStatus = "AWAITING"
	GAMEMOVE_STATUS_COMPLETE GameMoveStatus = "COMPLETE"
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
	`, string(GAMEMOVE_STATUS_COMPLETE), gm.Id)
	if err != nil {
		log.Printf("An error occurred in gamemove.MarkComplete():\n%s\n", err)
		return err
	}

	return nil
}

func CreateGameMove(gameBot GameBot) (GameMove, error) {
	var gameMoveId int
	db := GetDB()
	err := db.QueryRow(`
	INSERT INTO move (
	  game_bot_id
	) VALUES (
	  $1
	) RETURNING id
	`, gameBot.Id).Scan(&gameMoveId)
	if err != nil {
		log.Printf("An error occurred in gamemove.CreateGameMove():1:\n%s\n", err)
		return GameMove{}, err
	}

	gameMove, err := GetGameMoveById(gameMoveId)
	if err != nil {
		log.Printf("An error occurred in gamemove.CreateGameMove():2:\n%s\n", err)
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
	FROM move
	WHERE id = $1
	`, id).Scan(&gameMove.Id, &gameMove.gameBotId, &status)
	if err != nil {
		log.Printf("An error occurred in gamemove.GetGameMoveById():\n%s\n", err)
		return GameMove{}, err
	}

	gameMove.Status = GameMoveStatus(status)

	return gameMove, nil
}
