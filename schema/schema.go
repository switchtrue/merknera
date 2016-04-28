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
			Resolve: func(p graphql.ResolveParams) (interface{}, error) {
				return repository.ListGames()
			},
		},
		"game": &graphql.Field{
			Type:        GameType,
			Description: "List of bots",
			Args: graphql.FieldConfigArgument{
				"id": &graphql.ArgumentConfig{
					Type: graphql.Int,
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
	}
	rootQuery := graphql.ObjectConfig{Name: "RootQuery", Fields: fields}
	schemaConfig := graphql.SchemaConfig{Query: graphql.NewObject(rootQuery)}
	schema, err := graphql.NewSchema(schemaConfig)
	if err != nil {
		log.Fatalf("failed to create new schema, error: %v", err)
	}

	return &schema
}
