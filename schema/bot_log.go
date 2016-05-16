package schema

import (
	"github.com/graphql-go/graphql"
	"github.com/mleonard87/merknera/repository"
)

var botLogType *graphql.Object

func BotLogType() *graphql.Object {
	if botLogType == nil {
		botLogType = graphql.NewObject(
			graphql.ObjectConfig{
				Name:        "BotLog",
				Description: "A log (message) for a bot.",
				Fields: graphql.Fields{
					//"id": relay.GlobalIDField("BotLog", nil),
					"botLogId": &graphql.Field{
						Type:        graphql.Int,
						Description: "The unique ID of the bot.",
						Resolve: func(p graphql.ResolveParams) (interface{}, error) {
							if botLog, ok := p.Source.(repository.BotLog); ok {
								return botLog.Id, nil
							}
							return nil, nil
						},
					},
					"message": &graphql.Field{
						Type:        graphql.String,
						Description: "The log message itself.",
						Resolve: func(p graphql.ResolveParams) (interface{}, error) {
							if botLog, ok := p.Source.(repository.BotLog); ok {
								return botLog.Message, nil
							}
							return nil, nil
						},
					},
					"createdDatetime": &graphql.Field{
						Type:        graphql.String,
						Description: "The last known date/time that this bot was online.",
						Resolve: func(p graphql.ResolveParams) (interface{}, error) {
							if botLog, ok := p.Source.(repository.BotLog); ok {
								t := botLog.CreatedDateTime
								return t.UTC().Format("2006-01-02T15:04:05Z"), nil
							}
							return nil, nil
						},
					},
				},
				//Interfaces: []*graphql.Interface{
				//	nodeDefinitions.NodeInterface,
				//},
			},
		)
	}

	return botLogType
}
