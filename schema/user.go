package schema

import (
	"fmt"

	"github.com/graphql-go/graphql"
	"github.com/graphql-go/relay"
	"github.com/mleonard87/merknera/repository"
)

var userTokenType *graphql.Object
var newUserTokenType *graphql.Object
var userType *graphql.Object

var userConnectionDefinition *relay.GraphQLConnectionDefinitions

func UserConnectionDefinition() *relay.GraphQLConnectionDefinitions {
	if userConnectionDefinition == nil {
		userConnectionDefinition = relay.ConnectionDefinitions(relay.ConnectionConfig{
			Name:     "User",
			NodeType: UserType(),
		})
	}

	return userConnectionDefinition
}

func UserTokenType() *graphql.Object {
	if userTokenType == nil {
		userTokenType = graphql.NewObject(
			graphql.ObjectConfig{
				Name:        "UserTokenType",
				Description: "A user token that does not include the token for security purposes.",
				Fields: graphql.Fields{
					"id": relay.GlobalIDField("UserToken", nil),
					"userTokenId": &graphql.Field{
						Type:        graphql.Int,
						Description: "The unique ID of the user token.",
						Resolve: func(p graphql.ResolveParams) (interface{}, error) {
							if t, ok := p.Source.(repository.UserToken); ok {
								return t.Id, nil
							}
							fmt.Println("error getting token ID.")
							return nil, nil
						},
					},
					"description": &graphql.Field{
						Type:        graphql.String,
						Description: "A description of the usage of this token.",
						Resolve: func(p graphql.ResolveParams) (interface{}, error) {
							if t, ok := p.Source.(repository.UserToken); ok {
								return t.Description, nil
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

	return userTokenType
}

func NewUserTokenType() *graphql.Object {
	if newUserTokenType == nil {
		newUserTokenType = graphql.NewObject(
			graphql.ObjectConfig{
				Name:        "NewUserTokenType",
				Description: "A user token that includes the token. This is only used when a token is first created.",
				Fields: graphql.Fields{
					"id": relay.GlobalIDField("NewUserToken", nil),
					"userTokenId": &graphql.Field{
						Type:        graphql.Int,
						Description: "The unique ID of the user token.",
						Resolve: func(p graphql.ResolveParams) (interface{}, error) {
							if t, ok := p.Source.(repository.UserToken); ok {
								return t.Id, nil
							}
							fmt.Println("error getting token ID.")
							return nil, nil
						},
					},
					"description": &graphql.Field{
						Type:        graphql.String,
						Description: "A description of the usage of this token.",
						Resolve: func(p graphql.ResolveParams) (interface{}, error) {
							if t, ok := p.Source.(repository.UserToken); ok {
								return t.Description, nil
							}
							return nil, nil
						},
					},
					"token": &graphql.Field{
						Type:        graphql.String,
						Description: "The user token itself.",
						Resolve: func(p graphql.ResolveParams) (interface{}, error) {
							if t, ok := p.Source.(repository.UserToken); ok {
								return t.Token, nil
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

	return newUserTokenType
}

func UserType() *graphql.Object {
	if userType == nil {
		userType = graphql.NewObject(
			graphql.ObjectConfig{
				Name:        "UserType",
				Description: "A user registered with Merknera. Users may create and register bots.",
				Fields: graphql.Fields{
					"id": relay.GlobalIDField("User", nil),
					"userId": &graphql.Field{
						Type:        graphql.Int,
						Description: "The unique ID of the user.",
						Resolve: func(p graphql.ResolveParams) (interface{}, error) {
							if u, ok := p.Source.(repository.User); ok {
								return u.Id, nil
							}
							return nil, nil
						},
					},
					"name": &graphql.Field{
						Type:        graphql.String,
						Description: "The name of this user.",
						Resolve: func(p graphql.ResolveParams) (interface{}, error) {
							if u, ok := p.Source.(repository.User); ok {
								return u.Name, nil
							}
							return nil, nil
						},
					},
					"imageUrl": &graphql.Field{
						Type:        graphql.String,
						Description: "The URL of the users Google+ profile image.",
						Resolve: func(p graphql.ResolveParams) (interface{}, error) {
							if u, ok := p.Source.(repository.User); ok {
								imageUrl, err := u.ImageUrl.Value()
								if err != nil {
									return nil, nil
								}
								return imageUrl, nil
							}
							return nil, nil
						},
					},
					"email": &graphql.Field{
						Type:        graphql.String,
						Description: "The email address of the currently logged in user, otherwise null.",
						Resolve: func(p graphql.ResolveParams) (interface{}, error) {
							userId, isOK := p.Context.Value("userId").(float64)
							if isOK {
								user, err := repository.GetUserById(int(userId))
								if err != nil {
									return nil, err
								}

								return user.Email, nil
							}

							return nil, nil
						},
					},
					"tokenList": &graphql.Field{
						Type:        graphql.NewList(UserTokenType()),
						Description: "The list of tokens for the currently logged in user, otherwise null.",
						Resolve: func(p graphql.ResolveParams) (interface{}, error) {
							userId, isOK := p.Context.Value("userId").(float64)
							if isOK {
								user, err := repository.GetUserById(int(userId))
								if err != nil {
									return nil, err
								}

								return user.Tokens()
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

	return userType
}
