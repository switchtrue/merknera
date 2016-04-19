package repository

type GameType struct {
	Id       int
	Mnemonic string
	Name     string
}

func CreateGameType(mnemonic string, name string) (GameType, error) {
	db := GetDatabaseConnection()
	defer db.Close()

	var gameTypeId int
	err := db.QueryRow(`
	INSERT INTO game_type (
	  mnemonic
	, name
	) VALUES (
	  $1
	, $2
	) RETURNING id
	`, mnemonic, name).Scan(&gameTypeId)
	if err != nil {
		return GameType{}, err
	}

	gameType, err := GetGameTypeById(gameTypeId)
	if err != nil {
		return GameType{}, err
	}
	return gameType, nil
}

func GetGameTypeByMnemonic(mnemonic string) (GameType, error) {
	db := GetDatabaseConnection()
	defer db.Close()

	var gameType GameType
	err := db.QueryRow(`
	SELECT
	  id
	, mnemonic
	, name
	FROM game_type
	WHERE mnemonic = $1
	`, mnemonic).Scan(&gameType.Id, &gameType.Mnemonic, &gameType.Name)
	if err != nil {
		return GameType{}, err
	}

	return gameType, nil
}

func GetGameTypeById(id int) (GameType, error) {
	db := GetDatabaseConnection()
	defer db.Close()

	var gameType GameType
	err := db.QueryRow(`
	SELECT
	  id
	, mnemonic
	, name
	FROM game_type
	WHERE id = $1
	`, id).Scan(&gameType.Id, &gameType.Mnemonic, &gameType.Name)
	if err != nil {
		return GameType{}, err
	}

	return gameType, nil
}
