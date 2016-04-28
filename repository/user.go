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
	Id       int
	Username string
}

// Token generator taken from https://stackoverflow.com/questions/22892120/how-to-generate-a-random-string-of-a-fixed-length-in-golang
const letterBytes = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ"
const (
	letterIdxBits = 6                    // 6 bits to represent a letter index
	letterIdxMask = 1<<letterIdxBits - 1 // All 1-bits, as many as letterIdxBits
	letterIdxMax  = 63 / letterIdxBits   // # of letter indices fitting in 63 bits
	tokenLength   = 50
)

var src = rand.NewSource(time.Now().UnixNano())

func GenerateToken(n int) string {
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

func CreateUser(username string) (User, error) {
	var userId int
	db := GetDB()
	token := GenerateToken(tokenLength)
	err := db.QueryRow(`
	INSERT INTO merknera_user (
	  username
	) VALUES (
	  $1
	) RETURNING id
	`, username, token).Scan(&userId)
	if err != nil {
		log.Printf("An error occurred in user.CreateUser():1:\n%s\n", err)
		return User{}, err
	}

	_, err = db.Exec(`
	INSERT INTO merknera_user_token (
	  merknera_user_id
	, token
	) VALUES (
	  $1
	, $2
	)
	`, userId, token)
	if err != nil {
		log.Printf("An error occurred in user.CreateUser():2:\n%s\n", err)
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
	, username
	FROM merknera_user
	WHERE id = $1
	`, id).Scan(&user.Id, &user.Username)
	if err != nil {
		log.Printf("An error occurred in user.GetUserById():\n%s\n", err)
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
	, mu.username
	FROM merknera_user_token mut
	JOIN merknera_user mu
	  ON mut.merknera_user_id = mu.id
	WHERE token = $1
	`, token).Scan(&user.Id, &user.Username)
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
