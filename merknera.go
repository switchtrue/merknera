package main

import (
	"fmt"
	"log"
	"net/http"

	"github.com/mleonard87/merknera/gameworker"
	"github.com/mleonard87/merknera/repository"
	"github.com/mleonard87/merknera/services"
	"github.com/mleonard87/rpc"
	"github.com/mleonard87/rpc/json"
)

func Init() {
	s := rpc.NewServer()
	s.RegisterCodec(json.NewCodec(), "application/json")
	s.RegisterService(new(services.RegistrationService), "")
	s.RegisterService(new(services.UserService), "")
	http.Handle("/rpc", s)
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprint(w, "<h1>Welcome to Merknera</h1>")
	})

	gameworker.StartGameMoveDispatcher(4)

	botList, err := repository.ListBots()
	if err != nil {
		log.Fatal(err)
	}
	for _, b := range botList {
		b.Ping()
	}

	fmt.Println("Merknera is now listening on localhost:8080")
	http.ListenAndServe(":8080", nil)
}

func main() {
	Init()
}
