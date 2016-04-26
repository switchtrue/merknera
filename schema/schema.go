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
	}
	rootQuery := graphql.ObjectConfig{Name: "RootQuery", Fields: fields}
	schemaConfig := graphql.SchemaConfig{Query: graphql.NewObject(rootQuery)}
	schema, err := graphql.NewSchema(schemaConfig)
	if err != nil {
		log.Fatalf("failed to create new schema, error: %v", err)
	}

	return &schema
}
