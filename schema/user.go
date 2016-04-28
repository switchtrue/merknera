package schema

import (
	"github.com/graphql-go/graphql"
	"github.com/mleonard87/merknera/repository"
)

var UserType = graphql.NewObject(
	graphql.ObjectConfig{
		Name:        "UserType",
		Description: "A user registered with Merknera. Users may create and register bots.",
		Fields: graphql.Fields{
			"id": &graphql.Field{
				Type:        graphql.Int,
				Description: "The unique ID of the user.",
				Resolve: func(p graphql.ResolveParams) (interface{}, error) {
					if u, ok := p.Source.(repository.User); ok {
						return u.Id, nil
					}
					return nil, nil
				},
			},
			"username": &graphql.Field{
				Type:        graphql.String,
				Description: "The username of this user.",
				Resolve: func(p graphql.ResolveParams) (interface{}, error) {
					if u, ok := p.Source.(repository.User); ok {
						return u.Username, nil
					}
					return nil, nil
				},
			},
		},
	},
)
