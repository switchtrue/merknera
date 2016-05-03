package security

import (
	"crypto/rsa"
	"errors"
	"fmt"
	"io/ioutil"
	"log"

	"github.com/dgrijalva/jwt-go"
)

const (
	privateKeyPath  = "keys/app.rsa"     // openssl genrsa -out app.rsa keysize
	publicKeyPath   = "keys/app.rsa.pub" // openssl rsa -in app.rsa -pubout > app.rsa.pub
	JWT_COOKIE_NAME = "merknerajwt"
)

var (
	verifyKey *rsa.PublicKey
	signKey   *rsa.PrivateKey
)

// read the key files before starting http handlers
func init() {
	signBytes, err := ioutil.ReadFile(privateKeyPath)
	if err != nil {
		log.Fatalf("1:\n%v\n", err)
	}

	signKey, err = jwt.ParseRSAPrivateKeyFromPEM(signBytes)
	if err != nil {
		log.Fatalf("2:\n%v\n", err)
	}

	verifyBytes, err := ioutil.ReadFile(publicKeyPath)
	if err != nil {
		log.Fatalf("3:\n%v\n", err)
	}

	verifyKey, err = jwt.ParseRSAPublicKeyFromPEM(verifyBytes)
	if err != nil {
		log.Fatalf("4:\n%v\n", err)
	}
}

func NewJWTToken(userId int) (string, error) {
	token := jwt.New(jwt.SigningMethodRS256)
	token.Claims["userId"] = userId
	tokenString, err := token.SignedString(signKey)
	if err != nil {
		return "", err
	}

	return tokenString, nil
}

func ValidateToken(tokenStr string) (*jwt.Token, error) {
	token, err := jwt.Parse(tokenStr, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodRSA); !ok {
			fmt.Printf("Unexpected signing method: %v", token.Header["alg"])
			return nil, fmt.Errorf("Unexpected signing method: %v", token.Header["alg"])
		}
		return verifyKey, nil
	})
	if err != nil {
		return &jwt.Token{}, err
	}

	if token.Valid {
		return token, nil
	} else {
		return &jwt.Token{}, errors.New("Unable to validate JWT.")
	}
}
