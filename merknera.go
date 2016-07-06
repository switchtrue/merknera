package main

import (
	"fmt"
	"log"
	"net/http"

	"os"

	"github.com/graphql-go/handler"
	"github.com/mleonard87/merknera/games"
	"github.com/mleonard87/merknera/gameworker"
	"github.com/mleonard87/merknera/graphql"
	"github.com/mleonard87/merknera/repository"
	"github.com/mleonard87/merknera/schema"
	"github.com/mleonard87/merknera/security"
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

func registerLoginHandler() {
	lih := security.LoginHandler{}
	http.Handle("/login", lih)
	loh := security.LogoutHandler{}
	http.Handle("/logout", loh)
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
	awaitingMoves, err := repository.ListAwaitingMoves()
	if err != nil {
		log.Fatal(err)
	}

	for _, gameMove := range awaitingMoves {
		gb, err := gameMove.GameBot()
		if err != nil {
			log.Fatal(err)
		}

		g, err := gb.Game()
		if err != nil {
			log.Fatal(err)
		}

		gt, err := g.GameType()
		if err != nil {
			log.Fatal(err)
		}

		gameManagerConfig, err := games.GetGameManagerConfigByMnemonic(gt.Mnemonic)
		if err != nil {
			log.Fatal(err)
		}

		rpcMethodName, err := gameManagerConfig.GameManager.Resume(g)
		if err != nil {
			log.Fatal(err)
		}

		gameworker.QueueGameMove(rpcMethodName, gameMove)
	}
}

func main() {
	registerRPCHandler()
	registerGraphQLHandler()
	enableGraphiql := os.Getenv("MERKNERA_GRAPHIQL")
	if enableGraphiql == "TRUE" {
		registerStaticFileServerHandler()
	}
	registerAboutHandler()
	registerLoginHandler()

	gameworker.StartGameMoveDispatcher(4)

	//go verifyBotsAndQueueMoves()

	fmt.Println("Merknera is now listening on localhost:8080")
	http.ListenAndServe("localhost:8080", nil)
}
