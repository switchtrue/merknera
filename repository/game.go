package repository

type Game struct {
	Id       int
	GameType GameType
	Status   string
}

func CreateGame(gameType GameType) (Game, error) {
	db := GetDatabaseConnection()
	defer db.Close()

	var gameId int
	err := db.QueryRow(`
	INSERT INTO game (
	  game_type_id
	) VALUES (
	  $1
	) RETURNING id
	`, gameType.Id).Scan(&gameId)
	if err != nil {
		return Game{}, err
	}

	game, err := GetGameById(gameId)
	if err != nil {
		return Game{}, err
	}
	return game, nil
}

func GetGameById(id int) (Game, error) {
	db := GetDatabaseConnection()
	defer db.Close()

	var game Game
	var gameTypeId int
	err := db.QueryRow(`
	SELECT
	  g.id
	, g.status
	, g.game_type_id
	FROM game g
	WHERE g.id = $1
	`, id).Scan(&game.Id, &game.Status, &gameTypeId)
	if err != nil {
		return Game{}, err
	}
	game.GameType, _ = GetGameTypeById(gameTypeId)

	return game, nil
}
