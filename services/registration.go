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

			bot.Logf("Registered %s (version: %s)", bot.Name, bot.Version)
			responseMessage := fmt.Sprintf("A new version of your bot has been registered as %s (version: %s), good luck with %s!", bot.Name, bot.Version, gt.Name)
			reply.Message = responseMessage
		} else {
			bot.Logf("Re-registered %s (version: %s)", bot.Name, bot.Version)
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

			awaitingMoves, err := bot.ListAwaitingMoves()
			if err != nil {
				log.Fatal(err)
			}

			for _, am := range awaitingMoves {
				// TODO: Put a proper RPCMethod name in below
				gameworker.QueueGameMove("", am)
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

		bot.Logf("Registered %s (version: %s)", bot.Name, bot.Version)
		responseMessage := fmt.Sprintf("Hello, %s (version: %s), good luck with %s!", bot.Name, bot.Version, gt.Name)
		reply.Message = responseMessage
	}

	gameManagerConfig, err := games.GetGameManagerConfigByMnemonic(gameType.Mnemonic)
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

	otherBots, err := repository.ListBotsForGameTypeExcludingBot(gt, bot)
	if err != nil {
		em := "An error occurred whilst obtaining the list of other bots to play against."
		log.Printf("%s\n%s\n", em, err)
		return errors.New(em)
	}

	games := gameManagerConfig.GameProvider.GetGamesForBot(bot, otherBots)

	for _, players := range games {
		game, err := repository.CreateGame(gt)
		if err != nil {
			em := "An error occurred whilst creating a game."
			log.Printf("%s\n%s\n", em, err)
			return errors.New(em)
		}

		for i, player := range players {
			_, err = repository.CreateGameBot(game, *player, i)
			if err != nil {
				em := "An error occurred whilst creating game player."
				log.Printf("%s\n%s\n", em, err)
				return errors.New(em)
			}
		}

		gm, err := game.NextGameMove()
		if err != nil {
			em := "An error occurred whilst generating games for your bot."
			log.Printf("%s\n%s\n", em, err)
			return errors.New(em)
		}

		rpcMethod, initialPlayer, initialState, err := gameManagerConfig.GameProvider.Begin(game)
		if err != nil {
			em := "An error occurred whilst beginning the game."
			log.Printf("%s\n%s\n", em, err)
			return errors.New(em)
		}

		_, err = repository.CreateGameMove(initialPlayer, initialState)
		if err != nil {
			em := "An error occurred whilst creating the first move."
			log.Printf("%s\n%s\n", em, err)
			return errors.New(em)
		}

		gameworker.QueueGameMove(rpcMethod, gm)
	}

	return nil
}
