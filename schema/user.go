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
		},
	},
)
