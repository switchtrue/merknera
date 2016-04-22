package games

import (
	"database/sql"
	"log"

	"github.com/mleonard87/merknera/repository"
)

const (
	TICTACTOE_MNEMONIC             = "TICTACTOE"
	TICTACTOE_NAME                 = "Tic-Tac-Toe"
	TICTACTOE_RPC_METHOD_NEXT_MOVE = "TicTacToe.NextMove"
)

func init() {
	repository.InitializeDatabasePool()
	err := RegisterGameManager(new(TicTacToeGameManager))
	if err != nil {
		log.Fatal(err)
	}
}

type TicTacToeGameManager struct{}

func (tgm TicTacToeGameManager) GenerateGames(db *sql.DB, bot repository.Bot) []repository.Game {
	log.Println("GenerateGames")
	gameType, err := repository.GetGameTypeByMnemonic(db, TICTACTOE_MNEMONIC)
	if err != nil {
		log.Fatal(err)
	}

	botList, err := repository.ListBotsForGameType(db, gameType)
	if err != nil {
		log.Fatal(err)
	}

	var gameList []repository.Game
	for _, b := range botList {
		// If its not the same bot as we are invoking this game for then create the game.
		if b.Id != bot.Id {
			// Create a game for these two bots with the initial bot as player 1
			game1, err := createGameWithPlayers(db, gameType, &b, &bot)
			if err != nil {
				log.Fatal(err)
			}
			// Create a game for these two bots with the initial bot as player 2
			game2, err := createGameWithPlayers(db, gameType, &bot, &b)
			if err != nil {
				log.Fatal(err)
			}
			gameList = append(gameList, game1, game2)
		}
	}

	return gameList
}

func (tgm TicTacToeGameManager) Mnemonic() string {
	return TICTACTOE_MNEMONIC
}

func (tgm TicTacToeGameManager) Name() string {
	return TICTACTOE_NAME
}

func (tgm TicTacToeGameManager) GetNextMoveRPCMethodName() string {
	return TICTACTOE_RPC_METHOD_NEXT_MOVE
}

func (tgm TicTacToeGameManager) GetNextMoveRPCParams(gameMove repository.GameMove) interface{} {
	return nil
}

func createGameWithPlayers(db *sql.DB, gameType repository.GameType, playerOne *repository.Bot, playerTwo *repository.Bot) (repository.Game, error) {
	game, err := repository.CreateGame(db, gameType)
	if err != nil {
		return game, err
	}

	_, err = repository.CreateGameBot(db, game, *playerOne, 1)
	if err != nil {
		return game, err
	}

	_, err = repository.CreateGameBot(db, game, *playerTwo, 2)
	if err != nil {
		return game, err
	}

	err = createFirstGameMove(db, game)
	if err != nil {
		return game, err
	}

	return game, nil
}

func createFirstGameMove(db *sql.DB, game repository.Game) error {
	log.Println("beginGame")

	players, err := game.Players(db)
	if err != nil {
		return err
	}
	firstPlayer := players[0]
	_, err = repository.CreateGameMove(db, firstPlayer)
	if err != nil {
		return err
	}

	return nil
}

type nextMoveParams struct {
	GameId    int      `json:"gameid"`
	Mark      string   `json:"mark"`
	GameState []string `json:"gamestate"`
}
