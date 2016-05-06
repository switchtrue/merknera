package repository

import "log"

type UserTokenStatus string

const (
	USER_TOKEN_STATUS_CURRENT UserTokenStatus = "CURRENT"
	USER_TOKEN_STATUS_REVOKED UserTokenStatus = "REVOKED"
)

type UserToken struct {
	Id          int
	user_id     int
	Token       string
	Description string
	Status      UserTokenStatus
}

func (ut *UserToken) User() (User, error) {
	return GetUserById(ut.user_id)
}

func GetUserTokenById(id int) (UserToken, error) {
	var ut UserToken
	var status string
	db := GetDB()
	err := db.QueryRow(`
	SELECT
	  id
	, merknera_user_id
	, token
	, description
	, status
	FROM merknera_user_token
	WHERE id = $1
	`, id).Scan(&ut.Id, &ut.user_id, &ut.Token, &ut.Description, &status)
	if err != nil {
		log.Printf("An error occurred in user_token.GetUserTokenById():\n%s\n", err)
		return UserToken{}, err
	}

	ut.Status = UserTokenStatus(status)

	return ut, nil
}
