package games

import (
	"log"

	"fmt"

	"github.com/mleonard87/merknera/repository"
)

type TicTacToeGameManager struct{}

func (ttt TicTacToeGameManager) GenerateGames(bot repository.Bot) {
	gameType, err := repository.GetGameTypeByMnemonic(GAME_TYPE_TICTACTOE)
	if err != nil {
		log.Fatal(err)
	}

	botList, err := repository.ListBotsForGameType(gameType)
	if err != nil {
		log.Fatal(err)
	}

	for _, b := range botList {
		// If its not the same bot as we are invoking this game for then create the game.
		if b.Id != bot.Id {
			fmt.Println("Creating game...")
			// Create a game for these two bots with the initial bot as player 1
			err := registerPlayers(gameType, &b, &bot)
			if err != nil {
				log.Fatal(err)
			}
			// Create a game for these two bots with the initial bot as player 2
			err = registerPlayers(gameType, &bot, &b)
			if err != nil {
				log.Fatal(err)
			}
		}
	}

	fmt.Println(botList)
}

func registerPlayers(gameType repository.GameType, playerOne *repository.Bot, playerTwo *repository.Bot) error {
	game, err := repository.CreateGame(gameType)
	if err != nil {
		return err
	}
	_, err = repository.CreateGameBot(game, *playerOne, 1)
	if err != nil {
		return err
	}
	repository.CreateGameBot(game, *playerTwo, 2)
	if err != nil {
		return err
	}

	return nil
}
