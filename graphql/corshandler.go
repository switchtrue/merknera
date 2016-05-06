package graphql

import (
	"net/http"

	"os"

	"github.com/graphql-go/handler"
	"github.com/mleonard87/merknera/security"
	"golang.org/x/net/context"
)

type CORSHandler struct {
	graphQLGoHandler *handler.Handler
}

func (c CORSHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	// allow cross domain AJAX requests
	w.Header().Set("Access-Control-Allow-Origin", os.Getenv("MERKNERA_ALLOW_ORIGIN"))
	w.Header().Set("Access-Control-Allow-Headers", "content-type,authorization")
	w.Header().Set("Access-Control-Allow-Credentials", "true")

	if r.Method != http.MethodOptions {
		jwtCookie, _ := r.Cookie(security.JWT_COOKIE_NAME)
		gqlContext := context.Background()

		if jwtCookie != nil {
			tokenStr := jwtCookie.Value

			// Validate the JWT
			token, _ := security.ValidateToken(tokenStr)

			// Add the userId to the context for GraphQL.
			gqlContext = context.WithValue(context.Background(), "userId", token.Claims["userId"])
		}

		c.graphQLGoHandler.ContextHandler(gqlContext, w, r)
	}
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
