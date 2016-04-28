package graphql

import (
	"net/http"

	"github.com/graphql-go/handler"
	"golang.org/x/net/context"
)

type CORSHandler struct {
	graphQLGoHandler *handler.Handler
}

func (c CORSHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	// allow cross domain AJAX requests
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Access-Control-Allow-Headers", "content-type")

	c.graphQLGoHandler.ContextHandler(context.Background(), w, r)
}

func NewCORSHandler(p *handler.Config) *CORSHandler {
	if p == nil {
		p = handler.NewConfig()
	}
	if p.Schema == nil {
		panic("undefined GraphQL schema")
	}

	return &CORSHandler{
		graphQLGoHandler: &handler.Handler{
			Schema: p.Schema,
		},
	}
}
