package rpcclient

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/Jeffail/gabs"
)

var (
	httpClient = &http.Client{
		Transport: &http.Transport{
			MaxIdleConnsPerHost: 100,
		},
		Timeout: time.Second * 500,
	}
)

type RPCRequest struct {
	Jsonrpc string        `json:"jsonrpc"`
	Method  string        `json:"method"`
	Params  []interface{} `json:"params"`
	ID      int           `json:"id"`
}

func NewRPCRequest(jsonrpc string, method string, param ...interface{}) *RPCRequest {
	r := new(RPCRequest)
	r.Jsonrpc = jsonrpc
	r.Method = method
	r.Params = make([]interface{}, 0)
	r.Params = append(r.Params, param...)
	return r
}

func SendRPCRequst(host string, rpcRequest *RPCRequest) (*gabs.Container, error) {
	var buff bytes.Buffer
	err := json.NewEncoder(&buff).Encode(rpcRequest)
	if err != nil {
		return nil, fmt.Errorf("SendRPCRequst EncodeRequest error --- %s", err)
	}

	req, _ := http.NewRequest("POST", host, &buff)
	req.Header.Set("Content-Type", "application/json")
	resp, err := httpClient.Do(req)
	//resp, err := http.Post(host, "application/json", &buff)
	if err != nil {
		return nil, fmt.Errorf("SendRPCRequst Post error --- %s", err)
	}
	defer resp.Body.Close()

	jsonParsed, err := gabs.ParseJSONBuffer(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("SendRPCRequst ParseJSONBuffer error --- %s", err)
	}
	return jsonParsed, nil
}
