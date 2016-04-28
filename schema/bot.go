package schema

import (
	"github.com/graphql-go/graphql"
	"github.com/mleonard87/merknera/repository"
)

var BotType = graphql.NewObject(
	graphql.ObjectConfig{
		Name:        "Bot",
		Description: "A bot that plays a game.",
		Fields: graphql.Fields{
			"id": &graphql.Field{
				Type:        graphql.Int,
				Description: "The unique ID of the bot.",
				Resolve: func(p graphql.ResolveParams) (interface{}, error) {
					if bot, ok := p.Source.(repository.Bot); ok {
						return bot.Id, nil
					}
					return nil, nil
				},
			},
			"name": &graphql.Field{
				Type:        graphql.String,
				Description: "The name of the bot.",
				Resolve: func(p graphql.ResolveParams) (interface{}, error) {
					if bot, ok := p.Source.(repository.Bot); ok {
						return bot.Name, nil
					}
					return nil, nil
				},
			},
			"version": &graphql.Field{
				Type:        graphql.String,
				Description: "The version of the bot.",
				Resolve: func(p graphql.ResolveParams) (interface{}, error) {
					if bot, ok := p.Source.(repository.Bot); ok {
						return bot.Version, nil
					}
					return nil, nil
				},
			},
			"gameType": &graphql.Field{
				Type:        GameTypeType,
				Description: "Data about the game type that this bot is registered for.",
				Resolve: func(p graphql.ResolveParams) (interface{}, error) {
					if bot, ok := p.Source.(repository.Bot); ok {
						return bot.GameType()
					}
					return nil, nil
				},
			},
			"user": &graphql.Field{
				Type:        UserType,
				Description: "Data about the user that registered and owns this bot.",
				Resolve: func(p graphql.ResolveParams) (interface{}, error) {
					if bot, ok := p.Source.(repository.Bot); ok {
						return bot.User()
					}
					return nil, nil
				},
			},
			"rpcEndpoint": &graphql.Field{
				Type:        graphql.String,
				Description: "The RPC endpoint that will be called when this bot is required to make a move or to notify the bot of something.",
				Resolve: func(p graphql.ResolveParams) (interface{}, error) {
					if bot, ok := p.Source.(repository.Bot); ok {
						return bot.RPCEndpoint, nil
					}
					return nil, nil
				},
			},
			"programmingLanguage": &graphql.Field{
				Type:        graphql.String,
				Description: "The programming language used to write the bot.",
				Resolve: func(p graphql.ResolveParams) (interface{}, error) {
					if bot, ok := p.Source.(repository.Bot); ok {
						return bot.ProgrammingLanguage, nil
					}
					return nil, nil
				},
			},
			"website": &graphql.Field{
				Type:        graphql.String,
				Description: "An optional website for the bot.",
				Resolve: func(p graphql.ResolveParams) (interface{}, error) {
					if bot, ok := p.Source.(repository.Bot); ok {
						return bot.Website, nil
					}
					return nil, nil
				},
			},
			"description": &graphql.Field{
				Type:        graphql.String,
				Description: "An optional description about how the bot is implemented.",
				Resolve: func(p graphql.ResolveParams) (interface{}, error) {
					if bot, ok := p.Source.(repository.Bot); ok {
						return bot.Description, nil
					}
					return nil, nil
				},
			},
			"status": &graphql.Field{
				Type:        graphql.String,
				Description: "The current status of the bot.",
				Resolve: func(p graphql.ResolveParams) (interface{}, error) {
					if bot, ok := p.Source.(repository.Bot); ok {
						return string(bot.Status), nil
					}
					return nil, nil
				},
			},
			"gamesPlayed": &graphql.Field{
				Type:        graphql.Int,
				Description: "The number of games this bot has played.",
				Resolve: func(p graphql.ResolveParams) (interface{}, error) {
					if bot, ok := p.Source.(repository.Bot); ok {
						return bot.GamesPlayedCount()
					}
					return nil, nil
				},
			},
			"gamesWon": &graphql.Field{
				Type:        graphql.Int,
				Description: "The number of games this bot has won.",
				Resolve: func(p graphql.ResolveParams) (interface{}, error) {
					if bot, ok := p.Source.(repository.Bot); ok {
						return bot.GamesWonCount()
					}
					return nil, nil
				},
			},
		},
	},
)
