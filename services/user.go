package services

import (
	"net/http"

	"github.com/mleonard87/merknera/repository"
)

type UserArgs struct {
	Username string
}

type UserReply struct {
	UserId int
	Token  string
}

type UserService struct{}

func (h *UserService) Create(r *http.Request, args *UserArgs, reply *UserReply) error {
	user, err := repository.CreateUser(args.Username)
	if err != nil {
		return err
	}

	reply.UserId = user.Id
	reply.Token = user.Token

	return nil
}
