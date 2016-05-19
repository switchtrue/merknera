package schema

import (
	"errors"
	"fmt"
	"log"
	"strconv"

	"golang.org/x/net/context"

	"github.com/graphql-go/graphql"
	"github.com/graphql-go/relay"
	"github.com/mleonard87/merknera/repository"
)

var nodeDefinitions *relay.NodeDefinitions

func init() {
	/**
	 * We get the node interface and field from the relay library.
	 *
	 * The first method is the way we resolve an ID to its object. The second is the
	 * way we resolve an object that implements node to its type.
	 */
	nodeDefinitions = relay.NewNodeDefinitions(relay.NodeDefinitionsConfig{
		IDFetcher: func(id string, info graphql.ResolveInfo, ctx context.Context) (interface{}, error) {
			// resolve id from global id
			resolvedID := relay.FromGlobalID(id)

			// based on id and its type, return the object
			switch resolvedID.Type {
			case "Bot":
				i, _ := strconv.Atoi(resolvedID.ID)
				return repository.GetBotById(i)
			case "Game":
				i, _ := strconv.Atoi(resolvedID.ID)
				return repository.GetGameById(i)
			case "GameBot":
				i, _ := strconv.Atoi(resolvedID.ID)
				return repository.GetGameBotById(i)
			case "GameMove":
				i, _ := strconv.Atoi(resolvedID.ID)
				return repository.GetGameMoveById(i)
			case "GameTypeType":
				i, _ := strconv.Atoi(resolvedID.ID)
				return repository.GetGameTypeById(i)
			case "User":
				i, _ := strconv.Atoi(resolvedID.ID)
				return repository.GetUserById(i)
			default:
				return nil, errors.New("Unknown node type")
			}
		},
		TypeResolve: func(value interface{}, info graphql.ResolveInfo) *graphql.Object {
			// based on the type of the value, return GraphQLObjectType
			switch value.(type) {
			case repository.Bot:
				return BotType()
			case repository.Game:
				return GameType()
			case repository.GameBot:
				return GameBotType()
			case repository.GameMove:
				return GameMoveType()
			case repository.GameType:
				return GameTypeType()
			case repository.User:
				return UserType()
			default:
				return UserType()
			}
		},
	})
}

func MerkneraSchema() *graphql.Schema {
	rootQuery := graphql.NewObject(graphql.ObjectConfig{
		Name: "RootQuery",
		Fields: graphql.Fields{
			"bot": &graphql.Field{
				Type:        BotType(),
				Description: "Information about a specific bot.",
				Args: graphql.FieldConfigArgument{
					"id": &graphql.ArgumentConfig{
						Type:        graphql.ID,
						Description: "The ID of the bot you want information for.",
					},
					"botId": &graphql.ArgumentConfig{
						Type:        graphql.Int,
						Description: "The ID of the bot you want information for.",
					},
				},
				Resolve: func(p graphql.ResolveParams) (interface{}, error) {
					idStr, isOK := p.Args["id"].(string)
					if isOK {
						resolvedID := relay.FromGlobalID(idStr)
						i, _ := strconv.Atoi(resolvedID.ID)

						return repository.GetBotById(i)
					}

					idInt, isOK := p.Args["botId"].(int)
					if isOK {
						return repository.GetBotById(idInt)
					}
					return nil, nil
				},
			},
			"bots": &graphql.Field{
				Type: BotConnectionDefinition().ConnectionType,
				Args: graphql.FieldConfigArgument{
					"before": &graphql.ArgumentConfig{
						Type: graphql.String,
					},
					"after": &graphql.ArgumentConfig{
						Type: graphql.String,
					},
					"first": &graphql.ArgumentConfig{
						Type: graphql.Int,
					},
					"last": &graphql.ArgumentConfig{
						Type: graphql.Int,
					},
					"userId": &graphql.ArgumentConfig{
						Type:        graphql.Int,
						Description: "If a User ID is provided a list of bots will be returned that are owned by the specified user.",
					},
				},
				Resolve: func(p graphql.ResolveParams) (interface{}, error) {
					args := relay.NewConnectionArguments(p.Args)

					var bots []repository.Bot
					userId, isOK := p.Args["userId"].(int)
					if isOK {
						user, err := repository.GetUserById(int(userId))
						if err != nil {
							return nil, err
						}
						bots, err = user.ListBots()
						if err != nil {
							return nil, err
						}
					} else {
						bots, _ = repository.ListBots()
					}

					botsArray := []interface{}{}
					for _, b := range bots {
						botsArray = append(botsArray, b)
					}

					return relay.ConnectionFromArray(botsArray, args), nil
				},
			},
			"games": &graphql.Field{
				Type: GameConnectionDefinition().ConnectionType,
				Args: graphql.FieldConfigArgument{
					"before": &graphql.ArgumentConfig{
						Type: graphql.String,
					},
					"after": &graphql.ArgumentConfig{
						Type: graphql.String,
					},
					"first": &graphql.ArgumentConfig{
						Type: graphql.Int,
					},
					"last": &graphql.ArgumentConfig{
						Type: graphql.Int,
					},
					"botId": &graphql.ArgumentConfig{
						Type:        graphql.Int,
						Description: "If a Bot ID is provided a list of games will be returned for the specifie bot.",
					},
				},
				Resolve: func(p graphql.ResolveParams) (interface{}, error) {
					args := relay.NewConnectionArguments(p.Args)

					var games []repository.Game
					botId, isOK := p.Args["botId"].(int)
					if isOK {
						b, err := repository.GetBotById(botId)
						if err != nil {
							return nil, err
						}
						games, err = b.GamesPlayed()
						if err != nil {
							return nil, err
						}
					} else {
						games, _ = repository.ListGames()
					}

					gamesArray := []interface{}{}
					for _, g := range games {
						gamesArray = append(gamesArray, g)
					}

					return relay.ConnectionFromArray(gamesArray, args), nil
				},
			},
			"game": &graphql.Field{
				Type:        GameType(),
				Description: "Information about a specific game.",
				Args: graphql.FieldConfigArgument{
					"id": &graphql.ArgumentConfig{
						Type:        graphql.ID,
						Description: "The ID of the game you want information for.",
					},
					"gameId": &graphql.ArgumentConfig{
						Type:        graphql.Int,
						Description: "The ID of the game you want information for.",
					},
				},
				Resolve: func(p graphql.ResolveParams) (interface{}, error) {
					idStr, isOK := p.Args["id"].(string)
					if isOK {
						resolvedID := relay.FromGlobalID(idStr)
						i, _ := strconv.Atoi(resolvedID.ID)

						return repository.GetGameById(i)
					}

					idInt, isOK := p.Args["gameId"].(int)
					if isOK {
						return repository.GetGameById(idInt)
					}
					return nil, nil
				},
			},
			"users": &graphql.Field{
				Type: UserConnectionDefinition().ConnectionType,
				Args: relay.ConnectionArgs,
				Resolve: func(p graphql.ResolveParams) (interface{}, error) {
					args := relay.NewConnectionArguments(p.Args)

					users, _ := repository.ListUsers()
					usersArray := []interface{}{}
					for _, u := range users {
						usersArray = append(usersArray, u)
					}

					return relay.ConnectionFromArray(usersArray, args), nil
				},
			},
			"user": &graphql.Field{
				Type:        UserType(),
				Description: "Information about a specific user.",
				Args: graphql.FieldConfigArgument{
					"id": &graphql.ArgumentConfig{
						Type:        graphql.ID,
						Description: "The ID of the user you want information for.",
					},
					"userId": &graphql.ArgumentConfig{
						Type:        graphql.Int,
						Description: "The ID of the user you want information for.",
					},
				},
				Resolve: func(p graphql.ResolveParams) (interface{}, error) {
					idStr, isOK := p.Args["id"].(string)
					if isOK {
						resolvedID := relay.FromGlobalID(idStr)
						i, _ := strconv.Atoi(resolvedID.ID)

						return repository.GetUserById(i)
					}

					idInt, isOK := p.Args["userId"].(int)
					if isOK {
						return repository.GetUserById(idInt)
					}
					return nil, nil
				},
			},
			"currentUser": &graphql.Field{
				Type:        UserType(),
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
			"node": nodeDefinitions.NodeField,
		},
	})

	rootMutation := graphql.NewObject(graphql.ObjectConfig{
		Name: "RootMutation",
		Fields: graphql.Fields{
			"generateToken": &graphql.Field{
				Type:        NewUserTokenType(),
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
			"deleteBot": &graphql.Field{
				Type:        graphql.Int,
				Description: "Permanently delete a bot with the given id and all its prevous versions.",
				Args: graphql.FieldConfigArgument{
					"botId": &graphql.ArgumentConfig{
						Type:        graphql.NewNonNull(graphql.Int),
						Description: "The id bot to be deleted.",
					},
				},
				Resolve: func(p graphql.ResolveParams) (interface{}, error) {
					userId, isOK := p.Context.Value("userId").(float64)
					if isOK {
						user, err := repository.GetUserById(int(userId))
						if err != nil {
							return nil, err
						}

						botId, isOK := p.Args["botId"].(int)
						if isOK {
							bot, err := repository.GetBotById(botId)
							if err != nil {
								return nil, err
							}

							botUser, err := bot.User()
							if err != nil {
								return nil, err
							}

							if botUser.Id == user.Id {
								allBots, err := repository.ListBotsByName(bot.Name)
								if err != nil {
									return nil, err
								}

								for _, b := range allBots {
									err = b.Delete()
									if err != nil {
										return nil, err
									}
									return bot.Id, nil
								}
							} else {
								return nil, nil
							}
						}

						return nil, nil
					}

					return nil, nil
				},
			},
		},
	})

	schema, err := graphql.NewSchema(graphql.SchemaConfig{
		Query:    rootQuery,
		Mutation: rootMutation,
	})
	if err != nil {
		log.Fatalf("failed to create new schema, error: %v", err)
	}

	return &schema
}
