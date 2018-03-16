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

package bsvm

import (
	"errors"
	"fmt"
	"io/ioutil"
	"os"
	"sync"
	"testing"
	"time"

	"github.com/zipper-project/zipper/account"
	"github.com/zipper-project/zipper/common/crypto"
	ltyes "github.com/zipper-project/zipper/ledger/balance"
	"github.com/zipper-project/zipper/proto"
	"github.com/zipper-project/zipper/vm"
)

type MockerHandler struct {
	sync.Mutex
	cache map[string][]byte
}

func NewMockerHandler() *MockerHandler {
	return &MockerHandler{
		cache: make(map[string][]byte),
	}
}

func (hd *MockerHandler) GetGlobalState(key string) ([]byte, error) {
	hd.Lock()
	defer hd.Unlock()

	if value, ok := hd.cache[key]; ok {
		return value, nil
	}
	return []byte{}, errors.New("Not found")
}

func (hd *MockerHandler) PutGlobalState(key string, value []byte) error {
	hd.Lock()
	defer hd.Unlock()

	hd.cache[key] = value
	return nil
}

func (hd *MockerHandler) DelGlobalState(key string) error {
	hd.Lock()
	defer hd.Unlock()

	delete(hd.cache, key)
	return nil
}

func (hd *MockerHandler) GetState(key string) ([]byte, error) {
	hd.Lock()
	defer hd.Unlock()

	if value, ok := hd.cache[key]; ok {
		return value, nil
	}
	return []byte{}, errors.New("Not found")
}

func (hd *MockerHandler) PutState(key string, value []byte) error {
	hd.Lock()
	defer hd.Unlock()

	hd.cache[key] = value
	return nil
}

func (hd *MockerHandler) DelState(key string) error {
	hd.Lock()
	defer hd.Unlock()

	delete(hd.cache, key)
	return nil
}

func (hd *MockerHandler) ComplexQuery(key string) ([]byte, error) {
	return []byte{}, errors.New("Not found")
}

//func (hd *MockerHandler) GetByPrefix(prefix string) ([]*db.KeyValue, error) {
//	return []*db.KeyValue{}, nil
//}
//
//func (hd *MockerHandler) GetByRange(startKey, limitKey string) ([]*db.KeyValue, error) {
//	return []*db.KeyValue{}, nil
//}

func (hd *MockerHandler) GetBalance(addr string, assetID uint32) (int64, error) {
	return int64(100), nil
}

func (hd *MockerHandler) GetBalances(addr string) (*ltyes.Balance, error) {
	hd.Lock()
	defer hd.Unlock()

	balance := ltyes.NewBalance()
	balance.Amounts[0] = int64(100)
	balance.Amounts[1] = int64(50)
	return balance, nil
}

func (hd *MockerHandler) GetCurrentBlockHeight() uint32 {
	return 100
}

func (hd *MockerHandler) AddTransfer(fromAddr, toAddr string, assetID uint32, amount, fee int64) error {
	hd.Lock()
	defer hd.Unlock()
	return nil
}

func (hd *MockerHandler) Transfer(tx *proto.Transaction) error {
	return nil
}

func (hd *MockerHandler) SmartContractFailed() {

}

func (hd *MockerHandler) SmartContractCommitted() {

}

func (hd *MockerHandler) CombineAndValidRwSet(interface{}) interface{} {
	return nil
}

func (hd *MockerHandler) CallBack(response *vm.MockerCallBackResponse) error {
	return nil
}

var fileMap = make(map[string][]byte)
var fileLock sync.Mutex

func CreateContractSpec(args []string, fileName string) *proto.ContractSpec {
	contractSpec := &proto.ContractSpec{}
	contractSpec.Params = args
	var fileBuf []byte

	fileLock.Lock()
	defer fileLock.Unlock()
	if _, ok := fileMap[fileName]; ok {
		fileBuf = fileMap[fileName]
	} else {
		var err error
		f, _ := os.Open(fileName)
		fileBuf, err = ioutil.ReadAll(f)
		if err != nil {
			fmt.Println("read file failed ....", fileName)
			os.Exit(-1)
		}
		fileMap[fileName] = fileBuf
	}

	contractSpec.Code = fileBuf

	var a account.Address
	pubBytes := []byte("sender" + string(fileBuf))
	a.SetBytes(crypto.Keccak256(pubBytes[1:])[12:])
	contractSpec.Addr = a.Bytes()

	return contractSpec
}

func CreateContractData(args []string) *vm.ContractData {
	tx := &proto.Transaction{}
	tx.Header = &proto.TxHeader{
		Type: proto.TransactionType_LuaContractInit,
	}

	tx.ContractSpec = CreateContractSpec(args, "../luavm/coin.lua")
	return vm.NewContractData(tx)
}

func CreateContractDataWithFileName(args []string, name string, txType uint32) *vm.ContractData {
	tx := &proto.Transaction{}
	tx.Header.Type = proto.TransactionType(txType)
	return vm.NewContractData(tx)
}

func TestBsWorker(t *testing.T) {
	vm.NewTxSync(1)
	vm.VMConf = vm.DefaultConfig()
	bsWorker := NewBsWorker(vm.DefaultConfig(), 1)
	workerProc := &vm.WorkerProc{
		ContractData: CreateContractData([]string{}),
		SCHandler:    NewMockerHandler(),
	}

	bsWorker.VmJob(&vm.WorkerProcWithCallback{
		WorkProc: workerProc,
		Idx:      0,
	})

	time.Sleep(time.Second)
}
