package schema

import (
	"github.com/graphql-go/graphql"
	"github.com/mleonard87/merknera/repository"
)

var GameBotType = graphql.NewObject(
	graphql.ObjectConfig{
		Name:        "GameBot",
		Description: "An intersection of a bot and a game. That is, a bot playing a specific game in a specific play position (i.e. first player).",
		Fields: graphql.Fields{
			"id": &graphql.Field{
				Type:        graphql.Int,
				Description: "The unique ID of the game bot.",
				Resolve: func(p graphql.ResolveParams) (interface{}, error) {
					if gb, ok := p.Source.(repository.GameBot); ok {
						return gb.Id, nil
					}
					return nil, nil
				},
			},
			//"game": &graphql.Field{
			//	Type:        GameType,
			//	Description: "The game that this game bot was playing.",
			//	Resolve: func(p graphql.ResolveParams) (interface{}, error) {
			//		if gb, ok := p.Source.(repository.GameBot); ok {
			//			return gb.Game()
			//		}
			//		return nil, nil
			//	},
			//},
			"bot": &graphql.Field{
				Type:        BotType,
				Description: "The bot playing the game.",
				Resolve: func(p graphql.ResolveParams) (interface{}, error) {
					if gb, ok := p.Source.(repository.GameBot); ok {
						return gb.Bot()
					}
					return nil, nil
				},
			},
			"playSequence": &graphql.Field{
				Type:        graphql.Int,
				Description: "The order of play in which this bot played the game (e.g. 1 = player 1)",
				Resolve: func(p graphql.ResolveParams) (interface{}, error) {
					if gb, ok := p.Source.(repository.GameBot); ok {
						return gb.PlaySequence, nil
					}
					return nil, nil
				},
			},
		},
	},
)
