package schema

import (
	"github.com/graphql-go/graphql"
	"github.com/mleonard87/merknera/repository"
)

var GameMoveType = graphql.NewObject(
	graphql.ObjectConfig{
		Name:        "GameMove",
		Description: "A move by a bot in a game.",
		Fields: graphql.Fields{
			"id": &graphql.Field{
				Type:        graphql.Int,
				Description: "The unique ID of the game move.",
				Resolve: func(p graphql.ResolveParams) (interface{}, error) {
					if gm, ok := p.Source.(repository.GameMove); ok {
						return gm.Id, nil
					}
					return nil, nil
				},
			},
			"gameBot": &graphql.Field{
				Type:        GameBotType,
				Description: "The game bot that played this move.",
				Resolve: func(p graphql.ResolveParams) (interface{}, error) {
					if gm, ok := p.Source.(repository.GameMove); ok {
						return gm.GameBot()
					}
					return nil, nil
				},
			},
			"gameState": &graphql.Field{
				Type:        graphql.String,
				Description: "The current state of the game at the this this move was played. If the move is COMPLETE then this is the state of the game after the move was played, otherwise it is the same of the game before the move is played.",
				Resolve: func(p graphql.ResolveParams) (interface{}, error) {
					if gm, ok := p.Source.(repository.GameMove); ok {
						return gm.GameState()
					}
					return nil, nil
				},
			},
			"status": &graphql.Field{
				Type:        graphql.String,
				Description: "The user-friendly name of this game type.",
				Resolve: func(p graphql.ResolveParams) (interface{}, error) {
					if gm, ok := p.Source.(repository.GameMove); ok {
						return string(gm.Status), nil
					}
					return nil, nil
				},
			},
			"winner": &graphql.Field{
				Type:        graphql.Boolean,
				Description: "True if this move was the winning game move.",
				Resolve: func(p graphql.ResolveParams) (interface{}, error) {
					if gm, ok := p.Source.(repository.GameMove); ok {
						return gm.Winner, nil
					}
					return nil, nil
				},
			},
		},
	},
)
