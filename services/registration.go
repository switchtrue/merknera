package services

import (
	"log"
	"net/http"

	"errors"

	"database/sql"

	"fmt"

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
	Description         string `json:"description"`
}

type RegistrationReply struct {
	Message string `json:"message"`
}

type RegistrationService struct{}

func (h *RegistrationService) Register(r *http.Request, args *RegistrationArgs, reply *RegistrationReply) error {
	log.Printf("Registering %s (%s)\n", args.BotName, args.BotVersion)
	gameType, err := repository.GetGameTypeByMnemonic(args.Game)
	if err != nil {
		return err
	}

	user, err := repository.GetUserByToken(args.Token)
	if err != nil {
		return err
	}

	// Check to see if a bot with this name already exists.
	bot, err := repository.GetBotByName(args.BotName)
	if err != nil && err != sql.ErrNoRows {
		em := "An error occurred whilst registering your bot."
		log.Printf("%s\n%s\n", em, err)
		return errors.New(em)
	}

	// Get a list of all bots, l
	allBots, err := repository.ListBots()
	if err != nil {
		em := "An error occurred whilst registering your bot."
		log.Printf("%s\n%s\n", em, err)
		return errors.New(em)
	}

	// The zero-value for an int is 0 so if no bot was found then this will be 0.
	if bot.Id > 0 {
		botUser, err := bot.User()
		if err != nil {
			em := "An error occurred whilst registering your bot."
			log.Printf("%s\n%s\n", em, err)
			return errors.New(em)
		}

		// If there are bots and the bot we've retrieved is not owned by the current user then it has been taken by
		// someone else.
		if len(allBots) > 0 && botUser.Id != user.Id {
			em := fmt.Sprintf("The bot name \"%s\" has already been used by another user. All bot names must be unique, please use another name.", args.BotName)
			return errors.New(em)
		}

		exists, err := bot.DoesVersionExist(args.BotVersion)
		if err != nil {
			em := "An error occurred whilst registering your bot."
			log.Printf("%s\n%s\n", em, err)
			return errors.New(em)
		}

		if !exists {

			bot, err = repository.RegisterBot(args.BotName, args.BotVersion, gameType, user, args.RPCEndpoint, args.ProgrammingLanguage, args.Website, args.Description)
			if err != nil {
				em := "An error occurred whilst registering your bot."
				log.Printf("%s\n%s\n", em, err)
				return errors.New(em)
			}

			gt, err := bot.GameType()
			if err != nil {
				em := "An error occurred whilst registering your bot."
				log.Printf("%s\n%s\n", em, err)
				return errors.New(em)
			}

			responseMessage := fmt.Sprintf("A new version of your bot has been registered as %s (version: %s), good luck with %s!", bot.Name, bot.Version, gt.Name)
			reply.Message = responseMessage
		} else {
			responseMessage := fmt.Sprintf(`Hello, %s. The version \"%s\" of your bot is already registered. RPC Endpoint, Programming Language, Website and Description have been updated. No new games will
		be scheduled but your bot will marked as online and if there are any outstanding games they will be
		continued.`, bot.Name, bot.Version)
			reply.Message = responseMessage

			err = bot.Update(args.RPCEndpoint, args.ProgrammingLanguage, args.Website, args.Description)
			if err != nil {
				log.Fatal(err)
			}

			// Mark the bot as online again.
			err = bot.MarkOnline()
			if err != nil {
				log.Fatal(err)
			}

			return nil
		}
	} else {
		bot, err = repository.RegisterBot(args.BotName, args.BotVersion, gameType, user, args.RPCEndpoint, args.ProgrammingLanguage, args.Website, args.Description)
		if err != nil {
			em := "An error occurred whilst registering your bot."
			log.Printf("%s\n%s\n", em, err)
			return errors.New(em)
		}

		gt, err := bot.GameType()
		if err != nil {
			em := "An error occurred whilst registering your bot."
			log.Printf("%s\n%s\n", em, err)
			return errors.New(em)
		}

		responseMessage := fmt.Sprintf("Hello, %s (version: %s), good luck with %s!", bot.Name, bot.Version, gt.Name)
		reply.Message = responseMessage
	}

	gameManager, err := games.GetGameManager(gameType)
	if err != nil {
		em := "An error occurred whilst registering your bot."
		log.Printf("%s\n%s\n", em, err)
		return errors.New(em)
	}

	games := gameManager.GenerateGames(bot)
	for _, g := range games {
		gameMove, err := g.NextGameMove()
		if err != nil {
			em := "An error occurred whilst generating games for your bot."
			log.Printf("%s\n%s\n", em, err)
			return errors.New(em)
		}
		gameworker.QueueGameMove(gameMove)
	}

	return nil
}
