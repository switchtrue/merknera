package services

import (
	"net/http"

	"github.com/mleonard87/merknera/repository"
)

type UserArgs struct {
	Username string `json:"username"`
}

type UserReply struct {
	UserId int    `json:"userid"`
	Token  string `json:"token"`
}

type UserService struct{}

func (h *UserService) Create(r *http.Request, args *UserArgs, reply *UserReply) error {
	db := repository.NewTransaction()

	user, err := repository.CreateUser(db, args.Username)
	if err != nil {
		//db.Rollback()
		return err
	}

	//db.Commit()

	reply.UserId = user.Id
	reply.Token = user.Token

	return nil
}
