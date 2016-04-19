package repository

import (
	"math/rand"
	"time"

	_ "github.com/lib/pq"
)

type User struct {
	Id       int
	Username string
	Token    string
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
	db := GetDatabaseConnection()
	defer db.Close()

	var userId int
	err := db.QueryRow(`
	INSERT INTO merknera_user (
	  username
	, token
	) VALUES (
	  $1
	, $2
	) RETURNING id
	`, username, GenerateToken(tokenLength)).Scan(&userId)
	if err != nil {
		return User{}, err
	}

	user, err := GetUserById(userId)
	if err != nil {
		return User{}, err
	}
	return user, nil
}

func GetUserById(id int) (User, error) {
	db := GetDatabaseConnection()
	defer db.Close()

	var user User
	err := db.QueryRow(`
	SELECT
	  id
	, username
	, token
	FROM merknera_user
	WHERE id = $1
	`, id).Scan(&user.Id, &user.Username, &user.Token)
	if err != nil {
		return User{}, err
	}

	return user, nil
}

func GetUserByToken(token string) (User, error) {
	db := GetDatabaseConnection()
	defer db.Close()

	var user User
	err := db.QueryRow(`
	SELECT
	  id
	, username
	, token
	FROM merknera_user
	WHERE token = $1
	`, token).Scan(&user.Id, &user.Username, &user.Token)
	if err != nil {
		return User{}, err
	}

	return user, nil
}
