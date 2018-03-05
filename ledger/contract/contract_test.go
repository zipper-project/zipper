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
package contract

import (
	"os"
	"strconv"
	"testing"
	"time"

	"github.com/zipper-project/zipper/account"
	"github.com/zipper-project/zipper/common/db"
	"github.com/zipper-project/zipper/common/utils"
	"github.com/zipper-project/zipper/ledger/balance"
	pb "github.com/zipper-project/zipper/proto"
)

var (
	testDb        = db.NewDB(db.DefaultConfig())
	testSCAddr    = "0xa032277be213f56221b6140998c03d860a60e2f8"
	testSender    = account.HexToAddress("0xa032277be213f56221b6140998c03d860a60e1f8")
	testReciepent = account.HexToAddress("0xa032277be213f56221b6140998c03d860a60e3f8")
)

var smartContract *SmartConstract

type TestLedger struct {
}

func newTestLedger() *TestLedger {
	return &TestLedger{}
}

func (ledger *TestLedger) GetTmpBalance(addr account.Address) (*balance.Balance, error) {
	b := balance.NewBalance()
	b.Set(0, int64(20))
	return b, nil
}
func (ledger *TestLedger) Height() (uint32, error) {
	return uint32(10), nil
}

func makeSmartContract() *SmartConstract {
	return NewSmartConstract(testDb, newTestLedger())
}

func TestInitEnv(t *testing.T) {
	smartContract = makeSmartContract()
}

func TestSmartConstract_StartConstract(t *testing.T) {
	ht, _ := smartContract.ledgerHandler.Height()
	smartContract.StartConstract(ht)
}

func TestSmartConstract_ExecTransaction(t *testing.T) {
	tx := &pb.Transaction{
		TxData: pb.TxData{
			Header: &pb.TxHeader{
				FromChain:  account.NewChainCoordinate([]byte("0")),
				ToChain:    account.NewChainCoordinate([]byte("0")),
				Type:       pb.TransactionType_LuaContractInit,
				Nonce:      uint32(1),
				Sender:     testSender.String(),
				Recipient:  "00000000000000000000",
				AssetID:    uint32(0),
				Amount:     int64(10e11),
				Fee:        int64(1),
				CreateTime: uint32(time.Now().Unix()),
			},
		},
	}
	smartContract.ExecTransaction(tx, testSCAddr)
}

func TestSmartConstract_AddState(t *testing.T) {
	smartContract.AddState("hello", []byte("world"))
	smartContract.AddState("Lucy", []byte("sweet"))
}

func TestSmartConstract_DelState(t *testing.T) {
	smartContract.DelState("hello")
}

func TestSmartConstract_GetState(t *testing.T) {
	value, err := smartContract.GetState("hello")
	t.Log(" hello value: ", string(value), " err: ", err)
	value, err = smartContract.GetState("Lucy")
	t.Log(" Lucy value: ", string(value), " err: ", err)
}

func TestSmartConstract_CurrentBlockHeight(t *testing.T) {
	ht, _ := smartContract.ledgerHandler.Height()
	t.Log(" block height: ", ht)
}

func TestSmartConstract_GetBalances(t *testing.T) {
	balance, _ := smartContract.ledgerHandler.GetTmpBalance(account.HexToAddress("123456789"))
	t.Log(" account balance: ", balance)
}

func TestSmartConstract_AddTransfer(t *testing.T) {
	smartContract.AddTransfer("11000000000000000000", "22000000000000000000", uint32(0), int64(10), uint32(0))
}

func TestSmartConstract_SmartContractCommitted(t *testing.T) {
	smartContract.SmartContractCommitted()
}

func TestSmartConstract_FinishContractTransaction(t *testing.T) {
	txs, err := smartContract.FinishContractTransaction()
	t.Log(" txs: ", txs, " err: ", err)
}

func TestSmartConstract_AddChangesForPersistence(t *testing.T) {
	for i := 0; i < 10; i++ {
		smartContract.AddState("Tom_"+strconv.Itoa(i), []byte("hello"+strconv.Itoa(i)))
		smartContract.AddState("Tom_1"+strconv.Itoa(i), []byte("hello_1"+strconv.Itoa(i)))

	}

	var writeBatchs []*db.WriteBatch
	writeBatchs, _ = smartContract.AddChangesForPersistence(writeBatchs)
	if err := testDb.AtomicWrite(writeBatchs); err != nil {
		t.Error(err)
	}
}

func TestSmartConstract_StopContract(t *testing.T) {
	ht, _ := smartContract.ledgerHandler.Height()
	smartContract.StopContract(ht)
}

func TestSmartConstract_StartConstract2(t *testing.T) {
	ht, _ := smartContract.ledgerHandler.Height()
	smartContract.StartConstract(ht)
}

func TestSmartConstract_GetState2(t *testing.T) {
	value, _ := smartContract.GetState("Lucy")
	utils.AssertEquals(t, string(value), "sweet")
}

func TestGetByPrefix(t *testing.T) {

	for i := 0; i < 10; i++ {
		smartContract.AddState("Tom_1"+strconv.Itoa(i), []byte("hello_11"+strconv.Itoa(i)))
	}

	values := smartContract.GetByPrefix("Tom_1")

	value, _ := smartContract.GetState("Tom_1")
	t.Log(string(value))
	for _, v := range values {
		t.Log("key: ", string(v.Key), "value: ", string(v.Value))
	}

}

func TestGetByRange(t *testing.T) {

	values := smartContract.GetByRange("Tom_1", "Tom_8")
	for _, v := range values {
		t.Log("key: ", string(v.Key), "value: ", string(v.Value))
	}
	os.RemoveAll("/tmp/rocksdb-test/")

}
