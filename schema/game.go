package schema

import (
	"github.com/graphql-go/graphql"
	"github.com/mleonard87/merknera/repository"
)

var GameType = graphql.NewObject(
	graphql.ObjectConfig{
		Name:        "Game",
		Description: "A game played between one or more bots depending on the game type.",
		Fields: graphql.Fields{
			"id": &graphql.Field{
				Type:        graphql.Int,
				Description: "The unique ID of the game.",
				Resolve: func(p graphql.ResolveParams) (interface{}, error) {
					if g, ok := p.Source.(repository.Game); ok {
						return g.Id, nil
					}
					return nil, nil
				},
			},
			"gameType": &graphql.Field{
				Type:        GameTypeType,
				Description: "The mnemonic used to represent this game type.",
				Resolve: func(p graphql.ResolveParams) (interface{}, error) {
					if g, ok := p.Source.(repository.Game); ok {
						return g.GameType()
					}
					return nil, nil
				},
			},
			"players": &graphql.Field{
				Type:        graphql.NewList(GameBotType),
				Description: "The bots playing this game against each other.",
				Resolve: func(p graphql.ResolveParams) (interface{}, error) {
					if g, ok := p.Source.(repository.Game); ok {
						return g.Players()
					}
					return nil, nil
				},
			},
			"status": &graphql.Field{
				Type:        graphql.String,
				Description: "The user-friendly name of this game type.",
				Resolve: func(p graphql.ResolveParams) (interface{}, error) {
					if g, ok := p.Source.(repository.Game); ok {
						return string(g.Status), nil
					}
					return nil, nil
				},
			},
			"moves": &graphql.Field{
				Type:        graphql.NewList(GameMoveType),
				Description: "The moves played for this game, order by time ascending.",
				Resolve: func(p graphql.ResolveParams) (interface{}, error) {
					if g, ok := p.Source.(repository.Game); ok {
						return g.Moves()
					}
					return nil, nil
				},
			},
		},
	},
)
