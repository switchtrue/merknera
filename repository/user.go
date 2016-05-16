package repository

import (
	"database/sql"
	"errors"
	"fmt"
	"log"
	"math/rand"
	"time"
)

type User struct {
	Id       int `json:"id"`
	Name     string
	Email    string
	ImageUrl sql.NullString
}

// Token generator taken from https://stackoverflow.com/questions/22892120/how-to-generate-a-random-string-of-a-fixed-length-in-golang
const letterBytes = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"
const (
	letterIdxBits = 6                    // 6 bits to represent a letter index
	letterIdxMask = 1<<letterIdxBits - 1 // All 1-bits, as many as letterIdxBits
	letterIdxMax  = 63 / letterIdxBits   // # of letter indices fitting in 63 bits
	tokenLength   = 50
)

func (u *User) Update(name string, imageUrl string) error {
	db := GetDB()
	_, err := db.Exec(`
	UPDATE merknera_user
	SET
	  name = $1
	, image_url = $2
	WHERE id = $3
	`, name, imageUrl, u.Id)
	if err != nil {
		log.Printf("An error occurred in user.Update():\n%s\n", err)
		return err
	}

	u.Name = name
	u.ImageUrl = sql.NullString{String: imageUrl, Valid: true}

	return nil
}

var src = rand.NewSource(time.Now().UnixNano())

func generateToken(n int) string {
	b := make([]byte, n)
	// A src.Int63() generates 63 random bits, enough for letterIdxMax characters!
	for i, cache, remain := n-1, src.Int63(), letterIdxMax; i >= 0; {
		if remain == 0 {
			cache, remain = src.Int63(), letterIdxMax
		}
		if idx := int(cache & letterIdxMask); idx < len(letterBytes) {
			b[i] = letterBytes[idx]
			i--
		}
		cache >>= letterIdxBits
		remain--
	}

	return string(b)
}

func (u *User) Tokens() ([]UserToken, error) {
	db := GetDB()
	rows, err := db.Query(`
	SELECT
	  id
	, token
	, description
	, status
	FROM merknera_user_token
	WHERE merknera_user_id = $1
	AND status = $2
	ORDER BY description
	`, u.Id, string(USER_TOKEN_STATUS_CURRENT))
	if err != nil {
		log.Printf("An error occurred in user.Tokens():1:\n%s\n", err)
		return []UserToken{}, err
	}

	var utList []UserToken
	for rows.Next() {
		var ut UserToken
		var status string
		err := rows.Scan(&ut.Id, &ut.Token, &ut.Description, &status)
		if err != nil {
			log.Printf("An error occurred in user.Tokens():2:\n%s\n", err)
			return utList, err
		}
		ut.Status = UserTokenStatus(status)
		utList = append(utList, ut)
	}

	return utList, nil
}

func (u *User) CreateToken(description string) (UserToken, error) {
	var tokenId int
	db := GetDB()
	token := generateToken(50)
	err := db.QueryRow(`
	INSERT INTO merknera_user_token (
	  merknera_user_id
	, token
	, description
	) VALUES (
	  $1
	, $2
	, $3
	) RETURNING id
	`, u.Id, token, description).Scan(&tokenId)
	if err != nil {
		log.Printf("An error occurred in user.CreateToken():1:\n%s\n", err)
		return UserToken{}, err
	}

	userToken, err := GetUserTokenById(tokenId)
	if err != nil {
		log.Printf("An error occurred in user.CreateToken():2:\n%s\n", err)
		return UserToken{}, err
	}
	return userToken, nil
}

func (u *User) RevokeToken(tokenId int) error {
	db := GetDB()
	_, err := db.Exec(`
	UPDATE merknera_user_token
	SET status = $1
	WHERE id = $2
	AND merknera_user_id = $3
	`, string(USER_TOKEN_STATUS_REVOKED), tokenId, u.Id)
	if err != nil {
		log.Printf("An error occurred in user.RevokeToken():\n%s\n", err)
		return err
	}

	return nil
}

func (u *User) ListBots() ([]Bot, error) {
	db := GetDB()
	rows, err := db.Query(`
	SELECT
	  b.id
	, b.name
	, b.version
	, b.game_type_id
	, b.user_id
	, b.rpc_endpoint
	, b.programming_language
	, b.website
	, b.description
	, b.status
	, b.last_online_datetime
	FROM bot b
	WHERE b.user_id = $1
	AND b.status != $2
	ORDER BY b.name, b.version
	`, u.Id, string(BOT_STATUS_SUPERSEDED))
	if err != nil {
		return []Bot{}, err
	}

	var botList []Bot
	for rows.Next() {
		var bot Bot
		var status string
		err := rows.Scan(&bot.Id, &bot.Name, &bot.Version, &bot.gameTypeId, &bot.userId, &bot.RPCEndpoint, &bot.ProgrammingLanguage, &bot.Website, &bot.Description, &status, &bot.LastOnlineDateTime)
		if err != nil {
			log.Printf("An error occurred in user.ListBots():\n%s\n", err)
			return botList, err
		}
		bot.Status = BotStatus(status)
		botList = append(botList, bot)
	}

	return botList, nil
}

func CreateUser(name, email, imageUrl string) (User, error) {
	var userId int
	db := GetDB()
	err := db.QueryRow(`
	INSERT INTO merknera_user (
	  name
	, email
	, image_url
	) VALUES (
	  $1
	, $2
	, $3
	) RETURNING id
	`, name, email, imageUrl).Scan(&userId)
	if err != nil {
		log.Printf("An error occurred in user.CreateUser():1:\n%s\n", err)
		return User{}, err
	}

	user, err := GetUserById(userId)
	if err != nil {
		log.Printf("An error occurred in user.CreateUser():3:\n%s\n", err)
		return User{}, err
	}
	return user, nil
}

func GetUserById(id int) (User, error) {
	var user User
	db := GetDB()
	err := db.QueryRow(`
	SELECT
	  id
	, name
	, email
	, image_url
	FROM merknera_user
	WHERE id = $1
	`, id).Scan(&user.Id, &user.Name, &user.Email, &user.ImageUrl)
	if err != nil {
		log.Printf("An error occurred in user.GetUserById():\n%s\n", err)
		return User{}, err
	}

	return user, nil
}

func GetUserByEmail(email string) (User, error) {
	var user User
	db := GetDB()
	err := db.QueryRow(`
	SELECT
	  id
	, name
	, email
	, image_url
	FROM merknera_user
	WHERE email = $1
	`, email).Scan(&user.Id, &user.Name, &user.Email, &user.ImageUrl)
	if err != nil {
		log.Printf("An error occurred in user.GetUserByEmail():\n%s\n", err)
		return User{}, err
	}

	return user, nil
}

func GetUserByToken(token string) (User, error) {
	var user User
	db := GetDB()
	err := db.QueryRow(`
	SELECT
	  mu.id
	, mu.name
	, mu.email
	, mu.image_url
	FROM merknera_user_token mut
	JOIN merknera_user mu
	  ON mut.merknera_user_id = mu.id
	WHERE mut.token = $1
	  AND mut.status = $2
	`, token, string(USER_TOKEN_STATUS_CURRENT)).Scan(&user.Id, &user.Name, &user.Email, &user.ImageUrl)
	if err != nil {
		if err == sql.ErrNoRows {
			em := fmt.Sprintf("User with Token \"%s\" is not currently registered with Merknera", token)
			return User{}, errors.New(em)
		}
		log.Printf("An error occurred in user.GetUserByToken():\n%s\n", err)
		return User{}, err
	}

	return user, nil
}

func ListUsers() ([]User, error) {
	db := GetDB()
	rows, err := db.Query(`
	SELECT
	  u.id
	, u.name
	, u.email
	, u.image_url
	FROM merknera_user u
	ORDER BY u.name
	`)
	if err != nil {
		log.Printf("An error occurred in user.ListUsers():1:\n%s\n", err)
		return []User{}, err
	}

	var userList []User
	for rows.Next() {
		var user User
		err := rows.Scan(&user.Id, &user.Name, &user.Email, &user.ImageUrl)
		if err != nil {
			log.Printf("An error occurred in user.ListUsers():2:\n%s\n", err)
			return userList, err
		}
		userList = append(userList, user)
	}

	return userList, nil
}
