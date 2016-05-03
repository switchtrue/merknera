package security

import (
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"
	"strconv"
	"time"
)

const (
	GOOGLE_TOKENINFO_ENDPOINT = "https://www.googleapis.com/oauth2/v3/tokeninfo?id_token=%s"
)

type TokenInfoResponse struct {
	Issuer             string `json:"iss"`
	Audience           string `json:"aud"`
	Subject            string `json:"sub"`
	Email              string `json:"email"`
	Name               string `json:"name"`
	Picture            string `json:"picture"`
	GivenName          string `json:"given_name"`
	FamilyName         string `json:"family_name"`
	IssuedAtUnix       string `json:"iat"`
	IssuedAt           time.Time
	ExpirationTimeUnix string `json:"exp"`
	ExpirationTime     time.Time
	ErrorDescription   string `json:"error_description"`
}

func ValidateGoogleIdToken(token string) (TokenInfoResponse, error) {
	tokenInfoUrl := fmt.Sprintf(GOOGLE_TOKENINFO_ENDPOINT, token)

	response, err := http.Get(tokenInfoUrl)
	if err != nil {
		fmt.Printf("Error reaching Google token info endpoint:\n%v\n", err)
		return TokenInfoResponse{}, err
	}
	defer response.Body.Close()
	contents, err := ioutil.ReadAll(response.Body)
	if err != nil {
		fmt.Printf("Error reading response from Google token info endpoint:\n%v\n", err)
		return TokenInfoResponse{}, err
	}

	tir := TokenInfoResponse{}

	err = json.Unmarshal(contents, &tir)
	if err != nil {
		fmt.Printf("Error unmarshalling response from Google token info endpoint:\n%v\n", err)
		return TokenInfoResponse{}, err
	}

	if tir.ErrorDescription != "" {
		em := fmt.Sprintf("Google token info endpoint responded: %s", tir.ErrorDescription)
		return TokenInfoResponse{}, errors.New(em)
	}

	tir.IssuedAt, err = parseUnixTime(tir.IssuedAtUnix)
	if err != nil {
		em := fmt.Sprintf("Error parsing IssuedAt time (iat) \"%s\":\n%s\n", tir.IssuedAtUnix, err)
		return TokenInfoResponse{}, errors.New(em)
	}

	tir.ExpirationTime, err = parseUnixTime(tir.ExpirationTimeUnix)
	if err != nil {
		em := fmt.Sprintf("Error parsing Expiration time (xp) \"%s\":\n%s\n", tir.ExpirationTimeUnix, err)
		return TokenInfoResponse{}, errors.New(em)
	}

	return tir, nil
}

func parseUnixTime(ut string) (time.Time, error) {
	i, err := strconv.ParseInt(ut, 10, 64)
	if err != nil {
		return time.Time{}, err
	}
	tm := time.Unix(i, 0)

	return tm, nil
}
