package schema

import (
	"github.com/graphql-go/graphql"
	"github.com/graphql-go/relay"
	"github.com/mleonard87/merknera/repository"
)

var gameType *graphql.Object
var gameConnectionDefinition *relay.GraphQLConnectionDefinitions

func GameConnectionDefinition() *relay.GraphQLConnectionDefinitions {
	if gameConnectionDefinition == nil {
		gameConnectionDefinition = relay.ConnectionDefinitions(relay.ConnectionConfig{
			Name:     "Game",
			NodeType: GameType(),
			ConnectionFields: graphql.Fields{
				"totalCount": &graphql.Field{
					Type:        graphql.Int,
					Description: "The total number of games.",
					Resolve: func(p graphql.ResolveParams) (interface{}, error) {
						games, err := repository.ListGames()
						if err != nil {
							return nil, err
						}
						return len(games), nil
					},
				},
			},
		})
	}

	return gameConnectionDefinition
}

func GameType() *graphql.Object {
	if gameType == nil {
		gameType = graphql.NewObject(
			graphql.ObjectConfig{
				Name:        "Game",
				Description: "A game played between one or more bots depending on the game type.",
				Fields: graphql.Fields{
					"id": relay.GlobalIDField("Game", nil),
					"gameId": &graphql.Field{
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
						Type:        GameTypeType(),
						Description: "The mnemonic used to represent this game type.",
						Resolve: func(p graphql.ResolveParams) (interface{}, error) {
							if g, ok := p.Source.(repository.Game); ok {
								return g.GameType()
							}
							return nil, nil
						},
					},
					"players": &graphql.Field{
						Type:        graphql.NewList(GameBotType()),
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
						Type:        graphql.NewList(GameMoveType()),
						Description: "The moves played for this game, order by time ascending.",
						Resolve: func(p graphql.ResolveParams) (interface{}, error) {
							if g, ok := p.Source.(repository.Game); ok {
								return g.Moves()
							}
							return nil, nil
						},
					},
					"winningMove": &graphql.Field{
						Type:        GameMoveType(),
						Description: "The winning move of this game if the game is complete.",
						Resolve: func(p graphql.ResolveParams) (interface{}, error) {
							if g, ok := p.Source.(repository.Game); ok {
								wm, err := g.WinningMove()
								if err != nil {
									return nil, nil
								}
								if wm.Id == 0 {
									return nil, nil
								}
								return wm, nil
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

	return gameType
}
