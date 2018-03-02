// Copyright (C) 2017, Zipper Team.  All rights reserved.
//
// This file is part of zipper
//
// The zipper is free software: you can redistribute it and/or modify
// it under the terms of the GNU General Public License as published by
// the Free Software Foundation, either version 3 of the License, or
// (at your option) any later version.
//
// The zipper is distributed in the hope that it will be useful,
// but WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
//
// GNU General Public License for more details.
//
// You should have received a copy of the GNU General Public License
// along with this program.  If not, see <http://www.gnu.org/licenses/>.

package luavm

import (
	"os"
	"fmt"
	"sync"
	"errors"
	"testing"
	"io/ioutil"
	"github.com/zipper-project/zipper/vm"
	"github.com/zipper-project/zipper/proto"
	"github.com/zipper-project/zipper/account"
	"github.com/zipper-project/zipper/common/crypto"
	ltyes "github.com/zipper-project/zipper/ledger/types"
	"time"
)

type MockHandler struct {
	sync.Mutex
	cache map[string][]byte
}

func NewMockHandler() *MockHandler {
	return &MockHandler{
		cache: make(map[string][]byte),
	}
}

func (hd *MockHandler)GetGlobalState(key string) ([]byte, error) {
	hd.Lock()
	defer hd.Unlock()

	if value, ok := hd.cache[key]; ok {
		return value, nil
	}
	return []byte{}, errors.New("Not found")
}

func (hd *MockHandler)PutGlobalState(key string, value []byte) error {
	hd.Lock()
	defer hd.Unlock()

	hd.cache[key] = value
	return nil
}

func (hd *MockHandler)DelGlobalState(key string) error {
	hd.Lock()
	defer hd.Unlock()

	delete(hd.cache, key)
	return nil
}

func (hd *MockHandler) GetState(key string) ([]byte, error) {
	hd.Lock()
	defer hd.Unlock()

	if value, ok := hd.cache[key]; ok {
		return value, nil
	}
	return []byte{}, errors.New("Not found")
}

func (hd *MockHandler) PutState(key string, value []byte) error {
	hd.Lock()
	defer hd.Unlock()

	hd.cache[key] = value
	return nil
}

func (hd *MockHandler) DelState(key string) error {
	hd.Lock()
	defer hd.Unlock()

	delete(hd.cache, key)
	return nil
}

func (hd *MockHandler) ComplexQuery(key string) ([]byte, error) {
	return []byte{}, errors.New("Not found")
}

//func (hd *MockHandler) GetByPrefix(prefix string) ([]*db.KeyValue, error) {
//	return []*db.KeyValue{}, nil
//}
//
//func (hd *MockHandler) GetByRange(startKey, limitKey string) ([]*db.KeyValue, error) {
//	return []*db.KeyValue{}, nil
//}

func (hd *MockHandler) GetBalance(addr string, assetID uint32) (int64, error) {
	return int64(100), nil
}

func (hd *MockHandler) GetBalances(addr string) (*ltyes.Balance, error) {
	hd.Lock()
	defer hd.Unlock()

	balance := ltyes.NewBalance()
	balance.Amounts[0] = int64(100)
	balance.Amounts[1] = int64(50)
	return balance, nil
}

func (hd *MockHandler) GetCurrentBlockHeight() uint32 {
	return 100
}

func (hd *MockHandler) AddTransfer(fromAddr, toAddr string, assetID uint32, amount, fee int64) error {
	hd.Lock()
	defer hd.Unlock()
	fmt.Printf("AddTransfer from:%s to:%s amount:%d txType:%d", fromAddr, toAddr, amount, fee)
	return nil
}

func (hd *MockHandler) Transfer(tx *proto.Transaction) error {
	return nil
}

func (hd *MockHandler) SmartContractFailed() {

}

func (hd *MockHandler) SmartContractCommitted() {

}

func (hd *MockHandler) CombineAndValidRwSet(interface{}) interface{} {
	return nil
}

func (hd *MockHandler) CallBack(response *vm.MockerCallBackResponse) error {
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
				fmt.Println("read file failed ....")
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
	tx.ContractSpec = CreateContractSpec(args, "coin.lua")

	return vm.NewContractData(tx)
}


func TestLuaWorker(t *testing.T) {
	vm.VMConf = vm.DefaultConfig()
	luaWorker := NewLuaWorker(vm.DefaultConfig())
	workerProc := &vm.WorkerProc{
		ContractData: CreateContractData([]string{}),
		SCHandler: NewMockHandler(),
	}

	luaWorker.VmJob(&vm.WorkerProcWithCallback{
		WorkProc: workerProc,
		Idx: 1,
	})

	time.Sleep(time.Second)
}