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

package jsvm

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
	fmt.Printf("AddTransfer from:%s to:%s amount:%d txType:%d", fromAddr, toAddr, amount, fee)
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

var fileBuf []byte

func CreateContractSpec(args []string) *proto.ContractSpec {
	contractSpec := &proto.ContractSpec{}
	contractSpec.Params = args

	if len(fileBuf) == 0 {
		var err error
		f, _ := os.Open("./coin.js")
		fileBuf, err = ioutil.ReadAll(f)
		if err != nil {
			fmt.Println("read file failed ....")
			os.Exit(-1)
		}
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
		Type: proto.TransactionType_JSContractInit,
	}
	tx.ContractSpec = CreateContractSpec(args)
	return vm.NewContractData(tx)
}

func TestJsWorker(t *testing.T) {
	vm.VMConf = vm.DefaultConfig()
	jsWorker := NewJsWorker(vm.DefaultConfig())
	workerProc := &vm.WorkerProc{
		ContractData: CreateContractData([]string{}),
		SCHandler:    NewMockerHandler(),
	}

	_, err := jsWorker.VmJob(&vm.WorkerProcWithCallback{
		WorkProc: workerProc,
		Idx:      1,
	})

	fmt.Println(err)
	time.Sleep(time.Second)
}
