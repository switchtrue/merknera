package services

import (
	"net/http"

	"github.com/mleonard87/merknera/games"
	"github.com/mleonard87/merknera/repository"
)

type RegistrationArgs struct {
	BotName    string
	BotVersion string
	Game       string
	Token      string
	RpcUrl     string
}

type RegistrationReply struct {
	Message string
}

type RegistrationService struct{}

func (h *RegistrationService) Register(r *http.Request, args *RegistrationArgs, reply *RegistrationReply) error {
	gameType, err := repository.GetGameTypeByMnemonic(args.Game)
	if err != nil {
		return err
	}

	user, err := repository.GetUserByToken(args.Token)
	if err != nil {
		return err
	}

	bot, err := repository.RegisterBot(args.BotName, args.BotVersion, gameType, user, args.RpcUrl)
	if err != nil {
		return err
	}

	gameManager, err := games.GetGameManager(gameType)
	if err != nil {
		return err
	}
	gameManager.GenerateGames(bot)

	reply.Message = "Hello, " + bot.Name + ", enjoy " + bot.GameType.Name + "!"

	return nil
}
