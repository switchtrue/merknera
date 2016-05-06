package schema

import (
	"fmt"
	"log"

	"github.com/graphql-go/graphql"
	"github.com/mleonard87/merknera/repository"
)

var rootQuery = graphql.NewObject(graphql.ObjectConfig{
	Name: "RootQuery",
	Fields: graphql.Fields{
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
		"userList": &graphql.Field{
			Type:        graphql.NewList(UserType),
			Description: "List of users",
			Resolve: func(p graphql.ResolveParams) (interface{}, error) {
				return repository.ListUsers()
			},
		},
		"currentUser": &graphql.Field{
			Type:        UserType,
			Description: "Information about the currently logged in user.",
			Resolve: func(p graphql.ResolveParams) (interface{}, error) {
				userId, isOK := p.Context.Value("userId").(float64)
				if isOK {
					return repository.GetUserById(int(userId))
				} else {
					fmt.Println("not ok!")
				}
				return nil, nil
			},
		},
	},
})

var rootMutation = graphql.NewObject(graphql.ObjectConfig{
	Name: "RootMutation",
	Fields: graphql.Fields{
		"generateToken": &graphql.Field{
			Type:        NewUserTokenType,
			Description: "Generate a new token for the currently logged in user.",
			Args: graphql.FieldConfigArgument{
				"description": &graphql.ArgumentConfig{
					Type:        graphql.NewNonNull(graphql.String),
					Description: "A description to identfy the usage of this token.",
				},
			},
			Resolve: func(p graphql.ResolveParams) (interface{}, error) {
				userId, isOK := p.Context.Value("userId").(float64)
				if isOK {
					user, err := repository.GetUserById(int(userId))
					if err != nil {
						return nil, err
					}

					description, isOK := p.Args["description"].(string)
					if isOK {
						return user.CreateToken(description)
					}

					return nil, nil
				}

				return nil, nil
			},
		},
		"revokeToken": &graphql.Field{
			Type:        graphql.Int,
			Description: "Revoke an existing token belonging to the currently logged in user.",
			Args: graphql.FieldConfigArgument{
				"id": &graphql.ArgumentConfig{
					Type:        graphql.NewNonNull(graphql.Int),
					Description: "The id of the token we wish to revoke.",
				},
			},
			Resolve: func(p graphql.ResolveParams) (interface{}, error) {
				userId, isOK := p.Context.Value("userId").(float64)
				if isOK {
					user, err := repository.GetUserById(int(userId))
					if err != nil {
						return nil, err
					}

					tokenId, isOK := p.Args["id"].(int)
					if isOK {
						user.RevokeToken(tokenId)
						return tokenId, nil
					}

					return nil, nil
				}

				return nil, nil
			},
		},
	},
})

func MerkneraSchema() *graphql.Schema {
	schema, err := graphql.NewSchema(graphql.SchemaConfig{
		Query:    rootQuery,
		Mutation: rootMutation,
	})
	if err != nil {
		log.Fatalf("failed to create new schema, error: %v", err)
	}

	return &schema
}
