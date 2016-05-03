package security

import (
	"log"
	"net/http"

	"database/sql"

	"github.com/mleonard87/merknera/repository"
)

type LoginHandler struct{}

func (l LoginHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	// allow cross domain AJAX requests
	w.Header().Set("Access-Control-Allow-Origin", "http://localhost:3000")
	w.Header().Set("Access-Control-Allow-Headers", "content-type,authorization")
	w.Header().Set("Access-Control-Allow-Credentials", "true")

	idToken := r.FormValue("id_token")

	tir, err := ValidateGoogleIdToken(idToken)
	if err != nil {
		log.Printf("security/handler: Error validating id_token:\n%v\n", err)
		w.WriteHeader(http.StatusUnauthorized)
		w.Write([]byte("401 Unauthorized"))
		return
	}

	user, err := getOrCreateUserFromTokenInfo(tir)

	tokenString, err := NewJWTToken(user.Id)
	if err != nil {
		log.Printf("security/handler: Error generating JWT token:\n%v\n", err)
		w.WriteHeader(http.StatusUnauthorized)
		w.Write([]byte("401 Unauthorized"))
		return
	}

	jwtCookie := http.Cookie{
		Name:  JWT_COOKIE_NAME,
		Value: tokenString,
	}

	http.SetCookie(w, &jwtCookie)
	w.Write([]byte("OK"))
}

func getOrCreateUserFromTokenInfo(tir TokenInfoResponse) (repository.User, error) {
	var user repository.User

	user, err := repository.GetUserByEmail(tir.Email)
	if err != nil {
		// If we found no rows then bootstrap a user account now.
		if err == sql.ErrNoRows {
			user, err = repository.CreateUser(tir.Name, tir.Email, tir.Picture)
			if err != nil {
				return repository.User{}, err
			}
			return user, nil
		}
		return repository.User{}, err
	}

	err = user.Update(tir.Name, tir.Picture)
	if err != nil {
		return user, err
	}

	return user, nil
}
