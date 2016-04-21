package games

import (
	"database/sql"
	"log"

	"fmt"

	"github.com/mleonard87/merknera/gameworker"
	"github.com/mleonard87/merknera/repository"
)

const (
	TICTACTOE_MNEMONIC = "TICTACTOE"
	TICTACTOE_NAME     = "Tic-Tac-Toe"
)

func init() {
	repository.InitializeDatabasePool()
	err := RegisterGameManager(new(TicTacToeGameManager))
	if err != nil {
		log.Fatal(err)
	}
}

type TicTacToeGameManager struct{}

func (tgm TicTacToeGameManager) GenerateGames(db *sql.DB, bot repository.Bot) {
	log.Println("GenerateGames")
	gameType, err := repository.GetGameTypeByMnemonic(db, TICTACTOE_MNEMONIC)
	if err != nil {
		log.Fatal(err)
	}

	botList, err := repository.ListBotsForGameType(db, gameType)
	if err != nil {
		log.Fatal(err)
	}

	for _, b := range botList {
		// If its not the same bot as we are invoking this game for then create the game.
		if b.Id != bot.Id {
			// Create a game for these two bots with the initial bot as player 1
			err := registerPlayers(db, gameType, &b, &bot)
			if err != nil {
				log.Fatal(err)
			}
			// Create a game for these two bots with the initial bot as player 2
			err = registerPlayers(db, gameType, &bot, &b)
			if err != nil {
				log.Fatal(err)
			}
		}
	}
	fmt.Println("rerererer")
	fmt.Println(botList)
	fmt.Println("rerererer")
}

func (tgm TicTacToeGameManager) Mnemonic() string {
	return TICTACTOE_MNEMONIC
}

func (tgm TicTacToeGameManager) Name() string {
	return TICTACTOE_NAME
}

func registerPlayers(db *sql.DB, gameType repository.GameType, playerOne *repository.Bot, playerTwo *repository.Bot) error {
	game, err := repository.CreateGame(db, gameType)
	if err != nil {
		return err
	}

	_, err = repository.CreateGameBot(db, game, *playerOne, 1)
	if err != nil {
		return err
	}

	_, err = repository.CreateGameBot(db, game, *playerTwo, 2)
	if err != nil {
		return err
	}

	err = beginGame(db, game)
	if err != nil {
		return err
	}

	return nil
}

func beginGame(db *sql.DB, game repository.Game) error {
	log.Println("beginGame")

	players, err := game.Players(db)
	if err != nil {
		return err
	}
	firstPlayer := players[0]
	gameMove, err := repository.CreateGameMove(db, firstPlayer)
	if err != nil {
		return err
	}

	gameworker.QueueGameMove(gameMove)

	return nil
}
