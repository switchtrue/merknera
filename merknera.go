package main

import (
	"fmt"
	"log"
	"net/http"

	"github.com/graphql-go/handler"
	"github.com/mleonard87/merknera/gameworker"
	"github.com/mleonard87/merknera/graphql"
	"github.com/mleonard87/merknera/repository"
	"github.com/mleonard87/merknera/schema"
	"github.com/mleonard87/merknera/services"
	"github.com/mleonard87/rpc"
	"github.com/mleonard87/rpc/json"
)

func registerRPCHandler() {
	s := rpc.NewServer()
	s.RegisterCodec(json.NewCodec(), "application/json")
	s.RegisterService(new(services.RegistrationService), "")
	http.Handle("/rpc", s)
}

func registerGraphQLHandler() {
	schema := schema.MerkneraSchema()

	h := graphql.NewCORSHandler(&handler.Config{
		Schema: schema,
		Pretty: true,
	})

	http.Handle("/graphql", h)
}

func registerStaticFileServerHandler() {
	fs := http.FileServer(http.Dir("static"))
	http.Handle("/", fs)
}

func registerAboutHandler() {
	http.HandleFunc("/about", func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprint(w, "<h1>Welcome to Merknera</h1>")
	})
}

func verifyBotsAndQueueMoves() {
	botList, err := repository.ListBots()
	if err != nil {
		log.Fatal(err)
	}
	for _, b := range botList {
		b.Ping()
	}

	// Find all moves that are currently awaiting play and queue them.
	awaitingMoves, err := repository.GetAwaitingMoves()
	if err != nil {
		log.Fatal(err)
	}
	for _, gm := range awaitingMoves {
		log.Printf("Re-queuing: %d\n", gm.Id)
		gameworker.QueueGameMove(gm)
	}
}

func main() {
	registerRPCHandler()
	registerGraphQLHandler()
	registerStaticFileServerHandler()
	registerAboutHandler()

	gameworker.StartGameMoveDispatcher(4)

	verifyBotsAndQueueMoves()

	fmt.Println("Merknera is now listening on localhost:8080")
	http.ListenAndServe(":8080", nil)
}
