package schema

import (
	"github.com/graphql-go/graphql"
	"github.com/graphql-go/relay"
	"github.com/mleonard87/merknera/repository"
)

var gameMoveType *graphql.Object

func GameMoveType() *graphql.Object {
	if gameMoveType == nil {
		gameMoveType = graphql.NewObject(
			graphql.ObjectConfig{
				Name:        "GameMove",
				Description: "A move by a bot in a game.",
				Fields: graphql.Fields{
					"id": relay.GlobalIDField("GameMove", nil),
					"gameMoveId": &graphql.Field{
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
						Type:        GameBotType(),
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
								return gm.GameStateString()
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
					"startDateTime": &graphql.Field{
						Type:        graphql.String,
						Description: "The date and time that this move was started. This may not be accurate if endDateTime is null.",
						Resolve: func(p graphql.ResolveParams) (interface{}, error) {
							if gm, ok := p.Source.(repository.GameMove); ok {
								t := gm.StartDateTime
								if t.Valid {
									return t.Time.UTC().Format("2006-01-02T15:04:05Z"), nil
								}
								return nil, nil
							}
							return nil, nil
						},
					},
					"endDateTime": &graphql.Field{
						Type:        graphql.String,
						Description: "The date and time that this move was completed.",
						Resolve: func(p graphql.ResolveParams) (interface{}, error) {
							if gm, ok := p.Source.(repository.GameMove); ok {
								t := gm.EndDateTime
								if t.Valid {
									return t.Time.UTC().Format("2006-01-02T15:04:05Z"), nil
								}
								return nil, nil
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

	return gameMoveType
}
