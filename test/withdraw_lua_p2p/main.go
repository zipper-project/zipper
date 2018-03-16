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
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"time"

	"github.com/zipper-project/zipper/account"
	"github.com/zipper-project/zipper/common/crypto"
	"github.com/zipper-project/zipper/common/utils"
	"github.com/zipper-project/zipper/proto"
)

var (
	fromChain      = []byte{0}
	toChain        = []byte{0}
	txChan         = make(chan *proto.Transaction, 1)
	issuePriKeyHex = "496c663b994c3f6a8e99373c3308ee43031d7ea5120baf044168c95c45fbcf83"
)

func main() {
	srv.Start()
	time.Sleep(time.Second)
	go func() {
		for {
			select {
			case tx := <-txChan:
				fmt.Println(time.Now().Format("2006-01-02 15:04:05"), "Hash:", tx.Hash(), "Sender:", tx.Sender(), " Nonce: ", tx.Nonce(), "Asset: ", tx.AssetID(), " Type:", tx.GetType(), "txChan size:", len(txChan))
				Relay(tx)
			}
		}
	}()
	//模拟交易所提现场景
	systemPriv, _ := crypto.GenerateKey()
	systemAddr := account.PublicKeyToAddress(*systemPriv.Public())
	feePriv, _ := crypto.GenerateKey()
	feeAddr := account.PublicKeyToAddress(*feePriv.Public())
	//assetID := uint32(time.Now().UnixNano())

	userPriv, _ := crypto.GenerateKey()
	userAddr := account.PublicKeyToAddress(*userPriv.Public())

	for {
		assetID := uint32(time.Now().UnixNano())
		//1.发行资产系统账户	系统账户=10000
		issueTx(systemAddr, assetID, int64(10000))
		//2.转账给提现账户, 以完成提现操作 		提现账户=5000 系统账户=5000 合约账户=0
		atomicTx(systemPriv, userAddr, assetID, int64(5000))
		//3.部署提现合约
		initArgs := []string{}
		initArgs = append(initArgs, systemAddr.String())
		initArgs = append(initArgs, feeAddr.String())
		contractAddr := deployTx(systemPriv, assetID, int64(0), "./withdraw.lua", initArgs)
		//4.发起提现请求 1000		提现账户=4000 系统账户=5000 合约账户=1000
		invokeArgs := []string{}
		invokeArgs = append(invokeArgs, "launch")
		invokeArgs = append(invokeArgs, "D0001")
		invokeTx(userPriv, assetID, int64(1000), contractAddr, invokeArgs)
		//5.发起撤销提现请求		提现账户=5000 系统账户=5000 合约账户=0
		invokeArgs = []string{}
		invokeArgs = append(invokeArgs, "cancel")
		invokeArgs = append(invokeArgs, "D0001")
		invokeTx(userPriv, assetID, int64(0), contractAddr, invokeArgs)
		//4.发起提现请求		提现账户=4000 系统账户=5000 合约账户=1000
		invokeArgs = []string{}
		invokeArgs = append(invokeArgs, "launch")
		invokeArgs = append(invokeArgs, "D0002")
		invokeTx(userPriv, assetID, int64(1000), contractAddr, invokeArgs)
		//6.系统账户发起提现成功		提现账户=4000 系统账户=5900 合约账户=0 手续费账户=100
		invokeArgs = []string{}
		invokeArgs = append(invokeArgs, "succeed")
		invokeArgs = append(invokeArgs, "D0002")
		invokeArgs = append(invokeArgs, "100")
		invokeTx(systemPriv, assetID, int64(0), contractAddr, invokeArgs)
		//4.发起提现请求		提现账户=3000 系统账户=5900 合约账户=1000 手续费账户=100
		invokeArgs = []string{}
		invokeArgs = append(invokeArgs, "launch")
		invokeArgs = append(invokeArgs, "D0003")
		invokeTx(userPriv, assetID, int64(1000), contractAddr, invokeArgs)
		//7.系统账户发起提现失败			提现账户=4000 系统账户=5900 合约账户=0 手续费账户=100
		invokeArgs = []string{}
		invokeArgs = append(invokeArgs, "fail")
		invokeArgs = append(invokeArgs, "D0003")
		invokeTx(systemPriv, assetID, int64(0), contractAddr, invokeArgs)
	}
}

func issueTx(owner account.Address, assetID uint32, amount int64) {
	issueKey, _ := crypto.HexToECDSA(issuePriKeyHex)
	issueSender := account.PublicKeyToAddress(*issueKey.Public())
	tx := proto.NewTransaction(
		account.NewChainCoordinate(fromChain),
		account.NewChainCoordinate(toChain),
		proto.TransactionType_Issue,
		uint32(time.Now().UnixNano()),
		issueSender,
		owner,
		assetID,
		amount,
		0,
		uint32(time.Now().Unix()),
	)
	issueCoin := make(map[string]interface{})
	issueCoin["id"] = assetID
	tx.Payload, _ = json.Marshal(issueCoin)
	sig, _ := issueKey.Sign(tx.SignHash().Bytes())
	tx.WithSignature(sig)

	fmt.Println("> issuer :", owner.String())
	sendTransaction(tx)
}

func atomicTx(privkey *crypto.PrivateKey, owner account.Address, assetID uint32, amount int64) {
	sender := account.PublicKeyToAddress(*privkey.Public())
	tx := proto.NewTransaction(
		account.NewChainCoordinate(fromChain),
		account.NewChainCoordinate(toChain),
		proto.TransactionType_Atomic,
		uint32(time.Now().UnixNano()),
		sender,
		owner,
		assetID,
		amount,
		0,
		uint32(time.Now().Unix()),
	)
	sig, _ := privkey.Sign(tx.SignHash().Bytes())
	tx.WithSignature(sig)

	fmt.Println("> atomic :", owner.String())
	sendTransaction(tx)
}

func deployTx(privkey *crypto.PrivateKey, assetID uint32, amount int64, path string, args []string) account.Address {
	sender := account.PublicKeyToAddress(*privkey.Public())

	contractSpec := new(proto.ContractSpec)
	f, _ := os.Open(path)
	buf, _ := ioutil.ReadAll(f)
	contractSpec.Code = buf

	var a account.Address
	pubBytes := []byte(sender.String() + string(buf))
	a.SetBytes(crypto.Keccak256(pubBytes[1:])[12:])
	contractSpec.Addr = a.Bytes()

	contractSpec.Params = args

	tx := proto.NewTransaction(
		account.NewChainCoordinate(fromChain),
		account.NewChainCoordinate(toChain),
		proto.TransactionType_LuaContractInit,
		uint32(time.Now().UnixNano()),
		sender,
		account.NewAddress(contractSpec.Addr),
		assetID,
		amount,
		0,
		uint32(time.Now().Unix()),
	)
	tx.Payload = utils.Serialize(contractSpec)
	sig, _ := privkey.Sign(tx.SignHash().Bytes())
	tx.WithSignature(sig)
	fmt.Println("> deploy :", account.NewAddress(contractSpec.Addr).String(), contractSpec.Params)
	sendTransaction(tx)

	return a
}

func invokeTx(privkey *crypto.PrivateKey, assetID uint32, amount int64, contractAddr account.Address, args []string) {
	sender := account.PublicKeyToAddress(*privkey.Public())

	contractSpec := new(proto.ContractSpec)
	contractSpec.Code = contractAddr.Bytes()

	contractSpec.Params = args

	tx := proto.NewTransaction(
		account.NewChainCoordinate(fromChain),
		account.NewChainCoordinate(toChain),
		proto.TransactionType_ContractInvoke,
		uint32(time.Now().UnixNano()),
		sender,
		account.NewAddress(contractSpec.Addr),
		assetID,
		amount,
		int64(0),
		uint32(time.Now().Unix()),
	)

	tx.Payload = utils.Serialize(contractSpec)
	sig, _ := privkey.Sign(tx.SignHash().Bytes())
	tx.WithSignature(sig)
	fmt.Println("> invoke :", account.NewAddress(contractSpec.Addr).String(), contractSpec.Params)
	sendTransaction(tx)
}

func sendTransaction(tx *proto.Transaction) {
	txChan <- tx
}
