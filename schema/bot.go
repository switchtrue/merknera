package schema

import (
	"github.com/graphql-go/graphql"
	"github.com/mleonard87/merknera/repository"
)

//type Bot struct {
//	Id                  int
//	Name                string
//	Version             string
//	gameTypeId          int
//	gameType            GameType
//	userId              int
//	user                User
//	RPCEndpoint         string
//	ProgrammingLanguage string
//	Website             string
//	Status              BotStatus
//}

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
				Description: "The website for the bot.",
				Resolve: func(p graphql.ResolveParams) (interface{}, error) {
					if bot, ok := p.Source.(repository.Bot); ok {
						return bot.Website, nil
					}
					return nil, nil
				},
			},
			"status": &graphql.Field{
				Type:        graphql.String,
				Description: "The current status of the bot.",
				Resolve: func(p graphql.ResolveParams) (interface{}, error) {
					if bot, ok := p.Source.(repository.Bot); ok {
						return bot.Status, nil
					}
					return nil, nil
				},
			},
		},
	},
)
