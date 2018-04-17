package rpcclient

import (
	"fmt"
	"math/big"
)

var (
	rpchost = "http://localhost:8881"
)

const (
	methodGetBlockNumber = "Ledger.Height"
)

func getBlockNumber() (*big.Int, error) {
	request := NewRPCRequest("2.0", methodGetBlockNumber)
	jsonParsed, err := SendRPCRequst(rpchost, request)
	if err != nil {
		return big.NewInt(0), fmt.Errorf("getBlockNumber SendRPCRequst error --- %s", err)
	}

	if /*value*/ _, ok := jsonParsed.Path("error.code").Data().(float64); ok /*&& value > 0*/ {
		msg, _ := jsonParsed.Path("error.message").Data().(string)
		return nil, fmt.Errorf("getBlockNumber error --- %s", msg)
	}

	r, ok := jsonParsed.Path("result").Data().(string)
	if !ok {
		return big.NewInt(0), fmt.Errorf("getBlockNumber Path('result') interface error --- %s", err)
	}

	var ret = big.NewInt(0)
	ret.UnmarshalJSON([]byte(r))
	return ret, nil
}
