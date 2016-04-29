package schema

import (
	"log"

	"github.com/graphql-go/graphql"
	"github.com/mleonard87/merknera/repository"
)

func MerkneraSchema() *graphql.Schema {
	fields := graphql.Fields{
		"botList": &graphql.Field{
			Type:        graphql.NewList(BotType),
			Description: "List of bots",
			Resolve: func(p graphql.ResolveParams) (interface{}, error) {
				return repository.ListBots()
			},
		},
		"gameList": &graphql.Field{
			Type:        graphql.NewList(GameType),
			Description: "List of bots",
			Args: graphql.FieldConfigArgument{
				"botId": &graphql.ArgumentConfig{
					Type:        graphql.Int,
					Description: "If a Bot ID is provided a list of games will be returned for the specifie bot.",
				},
			},
			Resolve: func(p graphql.ResolveParams) (interface{}, error) {
				botId, isOK := p.Args["botId"].(int)
				if isOK {
					b, err := repository.GetBotById(botId)
					if err != nil {
						return nil, err
					}
					return b.GamesPlayed()
				}
				return repository.ListGames()
			},
		},
		"game": &graphql.Field{
			Type:        GameType,
			Description: "Information about a specific game.",
			Args: graphql.FieldConfigArgument{
				"id": &graphql.ArgumentConfig{
					Type:        graphql.Int,
					Description: "The ID of the game you want information for.",
				},
			},
			Resolve: func(p graphql.ResolveParams) (interface{}, error) {
				idQuery, isOK := p.Args["id"].(int)
				if isOK {
					return repository.GetGameById(idQuery)
				}
				return nil, nil
			},
		},
		"bot": &graphql.Field{
			Type:        BotType,
			Description: "Information about a specific bot.",
			Args: graphql.FieldConfigArgument{
				"id": &graphql.ArgumentConfig{
					Type:        graphql.Int,
					Description: "The ID of the bot you want information for.",
				},
			},
			Resolve: func(p graphql.ResolveParams) (interface{}, error) {
				idQuery, isOK := p.Args["id"].(int)
				if isOK {
					return repository.GetBotById(idQuery)
				}
				return nil, nil
			},
		},
	}
	rootQuery := graphql.ObjectConfig{Name: "RootQuery", Fields: fields}
	schemaConfig := graphql.SchemaConfig{Query: graphql.NewObject(rootQuery)}
	schema, err := graphql.NewSchema(schemaConfig)
	if err != nil {
		log.Fatalf("failed to create new schema, error: %v", err)
	}

	return &schema
}
