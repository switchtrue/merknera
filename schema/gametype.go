package schema

import (
	"github.com/graphql-go/graphql"
	"github.com/graphql-go/relay"
	"github.com/mleonard87/merknera/repository"
)

var gameTypeType *graphql.Object

func GameTypeType() *graphql.Object {
	if gameTypeType == nil {
		gameTypeType = graphql.NewObject(
			graphql.ObjectConfig{
				Name:        "GameTypeType",
				Description: "A game type registered with Merknera (e.g. Tic-Tac-Toe).",
				Fields: graphql.Fields{
					"id": relay.GlobalIDField("GameType", nil),
					"gameTypeId": &graphql.Field{
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
				Interfaces: []*graphql.Interface{
					nodeDefinitions.NodeInterface,
				},
			},
		)
	}

	return gameTypeType
}
