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

	//repository.InitializeConnectionPool()
	gameworker.StartGameMoveDispatcher(4)

	db := repository.GetDB()
	botList, err := repository.ListBots(db)
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
