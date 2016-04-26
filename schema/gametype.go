package schema

import (
	"github.com/graphql-go/graphql"
	"github.com/mleonard87/merknera/repository"
)

var GameTypeType = graphql.NewObject(
	graphql.ObjectConfig{
		Name:        "GameType",
		Description: "A game type registered with Merknera (e.g. Tic-Tac-Toe).",
		Fields: graphql.Fields{
			"id": &graphql.Field{
				Type:        graphql.Int,
				Description: "The unique ID of the game type.",
				Resolve: func(p graphql.ResolveParams) (interface{}, error) {
					if gt, ok := p.Source.(repository.GameType); ok {
						return gt.Id, nil
					}
					return nil, nil
				},
			},
			"mnemonic": &graphql.Field{
				Type:        graphql.String,
				Description: "The mnemonic used to represent this game type.",
				Resolve: func(p graphql.ResolveParams) (interface{}, error) {
					if gt, ok := p.Source.(repository.GameType); ok {
						return gt.Mnemonic, nil
					}
					return nil, nil
				},
			},
			"name": &graphql.Field{
				Type:        graphql.String,
				Description: "The user-friendly name of this game type.",
				Resolve: func(p graphql.ResolveParams) (interface{}, error) {
					if gt, ok := p.Source.(repository.GameType); ok {
						return gt.Name, nil
					}
					return nil, nil
				},
			},
		},
	},
)
