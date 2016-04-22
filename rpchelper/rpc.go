package rpchelper

import (
	"bytes"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
)

const (
	PING_METHOD_NAME = "Status.Ping"
)

type clientRequest struct {
	Method string      `json:"method"`
	Params interface{} `json:"params"`
	Id     int         `json:"id"`
}

func Ping(rpcEndpoint string) error {
	cr := new(clientRequest)
	cr.Id = 1
	cr.Method = PING_METHOD_NAME

	jsonBody, err := json.Marshal(*cr)
	if err != nil {
		return err
	}

	client := &http.Client{}

	req, err := http.NewRequest("POST", rpcEndpoint, bytes.NewBuffer(jsonBody))
	if err != nil {
		return err
	}
	req.Header.Add("Content-Type", "application/json")

	_, err = client.Do(req)
	if err != nil {
		return err
	}

	return nil
}

func Call(rpcEndpoint string, method string, args interface{}) {
	cr := new(clientRequest)
	cr.Id = 1
	cr.Method = method
	cr.Params = args

	jsonBody, err := json.Marshal(*cr)
	if err != nil {
		log.Fatal(err)
	}

	client := &http.Client{}

	req, err := http.NewRequest("POST", rpcEndpoint, bytes.NewBuffer(jsonBody))
	if err != nil {
		log.Fatal(err)
	}
	req.Header.Add("Content-Type", "application/json")
	req.Header.Add("Accept", "application/json")

	res, err := client.Do(req)
	fmt.Println(res)

}
