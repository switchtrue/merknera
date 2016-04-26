package services

import (
	"log"
	"net/http"

	"github.com/mleonard87/merknera/games"
	"github.com/mleonard87/merknera/gameworker"
	"github.com/mleonard87/merknera/repository"
)

type RegistrationArgs struct {
	BotName             string `json:"botname"`
	BotVersion          string `json:"botversion"`
	Game                string `json:"game"`
	Token               string `json:"token"`
	RPCEndpoint         string `json:"rpcendpoint"`
	ProgrammingLanguage string `json:"programminglanguage"`
	Website             string `json:"website"`
}

type RegistrationReply struct {
	Message string `json:"message"`
}

type RegistrationService struct{}

func (h *RegistrationService) Register(r *http.Request, args *RegistrationArgs, reply *RegistrationReply) error {
	log.Print("Register")

	gameType, err := repository.GetGameTypeByMnemonic(args.Game)
	if err != nil {
		return err
	}

	user, err := repository.GetUserByToken(args.Token)
	if err != nil {
		return err
	}

	bot, err := repository.RegisterBot(args.BotName, args.BotVersion, gameType, user, args.RPCEndpoint, args.ProgrammingLanguage, args.Website)
	if err != nil {
		return err
	}

	gameManager, err := games.GetGameManager(gameType)
	if err != nil {
		return err
	}

	games := gameManager.GenerateGames(bot)
	for _, g := range games {
		gameMove, err := g.NextGameMove()
		if err != nil {
			return err
		}
		gameworker.QueueGameMove(gameMove)
	}

	reply.Message = "Hello, " + bot.Name + ", enjoy " + bot.GameType.Name + "!"

	return nil
}
