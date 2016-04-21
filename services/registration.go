package services

import (
	"log"
	"net/http"

	"github.com/mleonard87/merknera/games"
	"github.com/mleonard87/merknera/repository"
)

type RegistrationArgs struct {
	BotName             string `json:"botname"`
	BotVersion          string `json:"botversion"`
	Game                string `json:"game"`
	Token               string
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
	db := repository.NewTransaction()

	gameType, err := repository.GetGameTypeByMnemonic(db, args.Game)
	if err != nil {
		//db.Rollback()
		return err
	}

	user, err := repository.GetUserByToken(db, args.Token)
	if err != nil {
		//db.Rollback()
		return err
	}

	bot, err := repository.RegisterBot(db, args.BotName, args.BotVersion, gameType, user, args.RPCEndpoint, args.ProgrammingLanguage, args.Website)
	if err != nil {
		//db.Rollback()
		return err
	}

	gameManager, err := games.GetGameManager(gameType)
	if err != nil {
		//db.Rollback()
		return err
	}

	gameManager.GenerateGames(db, bot)

	//db.Commit()

	reply.Message = "Hello, " + bot.Name + ", enjoy " + bot.GameType.Name + "!"

	return nil
}
