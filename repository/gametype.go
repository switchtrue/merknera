package repository

type GameType struct {
	Id       int
	Mnemonic string
	Name     string
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
