package rpchelper

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"time"
)

const (
	PING_METHOD_NAME = "Status.Ping"
)

type RPCClientRequest struct {
	JsonRpcVersion string      `json:"jsonrpc,omitempty"`
	Method         string      `json:"method"`
	Params         interface{} `json:"params"`
	Id             int         `json:"id"`
}

type RPCServerResponse struct {
	JsonRpcVersion string      `json:"jsonrpc,omitempty"`
	Result         interface{} `json:"result,omitempty"`
	Error          string      `json:"error,omitempty"`
	Id             int         `json:"id"`
}

func Ping(rpcEndpoint string) error {
	rcr := new(RPCClientRequest)
	rcr.JsonRpcVersion = "2.0"
	rcr.Id = 1
	rcr.Method = PING_METHOD_NAME

	jsonBody, err := json.Marshal(*rcr)
	if err != nil {
		return err
	}

	timeout := time.Duration(30 * time.Second)
	client := &http.Client{
		Timeout: timeout,
	}

	req, err := http.NewRequest("POST", rpcEndpoint, bytes.NewBuffer(jsonBody))
	if err != nil {
		return err
	}
	req.Header.Add("Content-Type", "application/json")

	res, err := client.Do(req)
	if err != nil {
		return err
	}

	if res.StatusCode != http.StatusOK {
		em := fmt.Sprintf("Response status not OK (200). Received %d", res.Status)
		return errors.New(em)
	}

	return nil
}

func Call(rpcEndpoint string, method string, params interface{}) (*RPCServerResponse, error) {
	rcr := new(RPCClientRequest)
	rcr.JsonRpcVersion = "2.0"
	rcr.Id = 1
	rcr.Method = method
	rcr.Params = params

	jsonBody, err := json.Marshal(*rcr)
	if err != nil {
		return &RPCServerResponse{}, err
	}

	client := &http.Client{}

	req, err := http.NewRequest("POST", rpcEndpoint, bytes.NewBuffer(jsonBody))
	if err != nil {
		return &RPCServerResponse{}, err
	}
	req.Header.Add("Content-Type", "application/json")
	req.Header.Add("Accept", "application/json")

	res, err := client.Do(req)
	if err != nil {
		return &RPCServerResponse{}, err
	}

	defer res.Body.Close()
	nextMoveResponse, err := ioutil.ReadAll(res.Body)
	if err != nil {
		return &RPCServerResponse{}, err
	}

	var rpcResult RPCServerResponse
	err = json.Unmarshal(nextMoveResponse, &rpcResult)
	if err != nil {
		return &RPCServerResponse{}, err
	}

	return &rpcResult, nil
}

func Notify(rpcEndpoint string, method string, args interface{}) error {
	rcr := new(RPCClientRequest)
	rcr.JsonRpcVersion = "2.0"
	rcr.Id = 1
	rcr.Method = method
	rcr.Params = args

	jsonBody, err := json.Marshal(*rcr)
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

	client.Do(req)

	return nil
}
