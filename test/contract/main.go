// Copyright (C) 2017, Zipper Team.  All rights reserved.
//
// This file is part of zipper
//
// The zipper is free software: you can use, copy, modify,
// and distribute this software for any purpose with or
// without fee is hereby granted, provided that the above
// copyright notice and this permission notice appear in all copies.
//
// The zipper is distributed in the hope that it will be useful,
// but WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
// ISC License for more details.
//
// You should have received a copy of the ISC License
// along with this program.  If not, see <https://opensource.org/licenses/isc>.

package main

import (
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/zipper-project/zipper/common/crypto"
	"github.com/zipper-project/zipper/account"
	"github.com/zipper-project/zipper/proto"
	"math/rand"
)

var (
	fromChain = []byte{0}
	toChain   = []byte{0}

	issuePriKeyHex = "496c663b994c3f6a8e99373c3308ee43031d7ea5120baf044168c95c45fbcf83"
	privkeyHex     = "596c663b994c3f6a8e99373c3308ee43031d7ea5120baf044168c95c45fbcf83"
	privkey, _     = crypto.HexToECDSA(privkeyHex)
	sender         = account.PublicKeyToAddress(*privkey.Public())

	senderAccount = "abc123"
	accountInfo   = map[string]string{
		"from_user": senderAccount,
		"to_user":   senderAccount,
	}

	txChan = make(chan *proto.Transaction, 5)

	httpClient = &http.Client{
		Transport: &http.Transport{
			MaxIdleConnsPerHost: 100,
		},
		Timeout: time.Second * 500,
	}
	url = "http://localhost:8881"
)

// contract lang
type contractLang string

func (lang contractLang) ConvertInitTxType() proto.TransactionType {
	switch lang {
	case langLua:
		return proto.TransactionType_LuaContractInit
	case langJS:
		return proto.TransactionType_JSContractInit
	}

	return proto.TransactionType(0)
}

const (
	langLua = "lua"
	langJS  = "js"
)

// contract config
type contractConf struct {
	path       string
	lang       contractLang
	isGlobal   bool
	initArgs   []string
	invokeArgs []string
}

func newContractConf(path string, lang contractLang, isGlobal bool, initArgs, invokeArgs []string) *contractConf {
	return &contractConf{
		path:       path,
		lang:       lang,
		isGlobal:   isGlobal,
		initArgs:   initArgs,
		invokeArgs: invokeArgs,
	}
}

var (
	voteLua = newContractConf(
		"./l0vote.lua",
		langLua,
		false,
		nil,
		[]string{"vote", "chain", "england"})

	coinLua = newContractConf(
		"./template.lua",
		langLua,
		false,
		nil,
		[]string{"transfer", "8ce1bb0858e71b50d603ebe4bec95b11d8833e68", "100"})

	coinJS = newContractConf(
		"./template.js",
		langJS,
		false,
		[]string{"hello", "world"},
		//[]string{"transfer", "8ce1bb0858e71b50d603ebe4bec95b11d8833e68", "100"})
		[]string{"testwrite", "8ce1bb0858e71b50d603ebe4bec95b11d8833e68", "100"})

	globalSetAccountLua = newContractConf(
		"./global.lua",
		langLua,
		true,
		nil,
		[]string{
			"SetGlobalState",
			"account." + senderAccount,
			fmt.Sprintf(`{"addr":"%s", "uid":"%s", "frozened":false}`, sender.String(), senderAccount),
		})

	securityPluginName = "security.so"
)

func main() {
	go sendTransaction()
	time.Sleep(1 * time.Microsecond)

	issueTX()
	time.Sleep(10)
	//transferTx()
	//testSecurityContract()
	for i:=0; i<0; i++ {
		time.Sleep(time.Second * 1)
		deploySmartContractTX(coinJS)
		time.Sleep(time.Second * 1)
		deploySmartContractTX(coinLua)
	// time.Sleep(10 * time.Second)
	// execSmartContractTX(coinJS)
	}
	ch := make(chan struct{})
	<-ch
}

func httpPost(postForm string, resultHandler func(result map[string]interface{})) {
	req, _ := http.NewRequest("POST", url, strings.NewReader(postForm))
	req.Header.Set("Content-Type", "application/json")
	resp, err := httpClient.Do(req)
	if err != nil {
		panic(err)
	}
	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		panic(fmt.Errorf("Couldn't parse response body. %+v", err))
	}

	var result map[string]interface{}
	json.Unmarshal(body, &result)
	if resultHandler == nil {
		fmt.Println("http response:", result)
	} else {
		resultHandler(result)
	}
}

func sendTransaction() {
	for {
		select {
		case tx := <-txChan:
			fmt.Printf("hash: %s, type: %v, nonce: %v, amount: %v\n",
				tx.Hash().String(), tx.GetType(), tx.Nonce(), tx.Amount())

			httpPost(`{"id":1,"method":"RPCTransaction.Broadcast","params":["`+hex.EncodeToString(tx.Serialize())+`"]}`, nil)
		}
	}
}

func issueTX() {
	issueKey, _ := crypto.HexToECDSA(issuePriKeyHex)
	nonce := rand.Uint32()
	issueSender := account.PublicKeyToAddress(*issueKey.Public())

	tx := proto.NewTransaction(
		account.NewChainCoordinate(fromChain),
		account.NewChainCoordinate(toChain),
		proto.TransactionType_Issue,
		nonce,
		issueSender,
		sender,
		1,
		1000000000,
		1,
		uint32(time.Now().Unix()),
	)

	issueCoin := make(map[string]interface{})
	issueCoin["id"] = 1
	tx.Payload, _ = json.Marshal(issueCoin)
	tx.Meta, _ = json.Marshal(map[string]map[string]string{
		"account": accountInfo,
	})

	fmt.Println("issueSender address: ", issueSender.String(), " receriver: ", sender.String())
	sig, _ := issueKey.Sign(tx.SignHash().Bytes())
	tx.WithSignature(sig)

	txChan <- tx
}

func transferTx() {
	nonce := rand.Uint32()
	privateKey, _ := crypto.GenerateKey()
	receiver := account.PublicKeyToAddress(*privateKey.Public())
	tx := proto.NewTransaction(
		account.NewChainCoordinate(fromChain),
		account.NewChainCoordinate(toChain),
		proto.TransactionType_Atomic,
		nonce,
		sender,
		receiver,
		1,
		10,
		1,
		uint32(time.Now().Unix()),
	)

	sig, _ := privkey.Sign(tx.SignHash().Bytes())
	tx.WithSignature(sig)

	txChan <- tx
}


func deploySmartContractTX(conf *contractConf) []byte {
	contractSpec := new(proto.ContractSpec)
	contractSpec.Params = conf.initArgs
	nonce := rand.Uint32()
	f, _ := os.Open(conf.path)
	buf, _ := ioutil.ReadAll(f)
	contractSpec.Code = buf

	if !conf.isGlobal {
		var a account.Address
		pubBytes := []byte(sender.String() + string(buf))
		a.SetBytes(crypto.Keccak256(pubBytes[1:])[12:])
		contractSpec.Addr = a.Bytes()
	}

	tx := proto.NewTransaction(
		account.NewChainCoordinate(fromChain),
		account.NewChainCoordinate(toChain),
		conf.lang.ConvertInitTxType(),
		nonce,
		sender,
		account.NewAddress(contractSpec.Addr),
		1,
		10,
		0,
		uint32(time.Now().Unix()),
	)


	tx.ContractSpec = contractSpec

	sig, _ := privkey.Sign(tx.SignHash().Bytes())
	tx.WithSignature(sig)
	fmt.Println("> deploy ContractAddr:", account.NewAddress(contractSpec.Addr).String())

	txChan <- tx
	return contractSpec.Addr
}

func execSmartContractTX(conf *contractConf) {
	contractSpec := new(proto.ContractSpec)
	contractSpec.Params = conf.invokeArgs

	if !conf.isGlobal {
		f, _ := os.Open(conf.path)
		buf, _ := ioutil.ReadAll(f)

		var a account.Address
		pubBytes := []byte(sender.String() + string(buf))
		a.SetBytes(crypto.Keccak256(pubBytes[1:])[12:])

		contractSpec.Addr = a.Bytes()
	}

	nonce := 2
	tx := proto.NewTransaction(
		account.NewChainCoordinate(fromChain),
		account.NewChainCoordinate(toChain),
		proto.TransactionType_ContractInvoke,
		uint32(nonce),
		sender,
		account.NewAddress(contractSpec.Addr),
		0,
		0,
		0,
		uint32(time.Now().Unix()),
	)

	fmt.Println("> exe ContractAddr:", account.NewAddress(contractSpec.Addr).String())
	tx.ContractSpec = contractSpec

	sig, _ := privkey.Sign(tx.SignHash().Bytes())
	tx.WithSignature(sig)
	txChan <- tx
}

func queryGlobalContract(key string) {
	form := `{"id": 2, "method": "Transaction.Query", "params":[{"ContractAddr":"","ContractParams":["` + key + `"]}]}`
	httpPost(form, func(result map[string]interface{}) {
		if result != nil {
			fmt.Printf("> query result: %s\n", result["result"])
		} else {
			fmt.Println("> query failed")
		}
	})
}
