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
package ledger

import (
	"encoding/json"
	"os"
	"testing"

	"github.com/zipper-project/zipper/account"
	"github.com/zipper-project/zipper/common/crypto"
	"github.com/zipper-project/zipper/common/db"
	"github.com/zipper-project/zipper/common/utils"
	pb "github.com/zipper-project/zipper/proto"
	"github.com/zipper-project/zipper/vm"
)

var (
	issueReciepent     = account.HexToAddress("0xa032277be213f56221b6140998c03d860a60e1f8")
	atmoicReciepent    = account.HexToAddress("0xa132277be213f56221b6140998c03d860a60e1f8")
	acrossReciepent    = account.HexToAddress("0xa232277be213f56221b6140998c03d860a60e1f8")
	distributReciepent = account.HexToAddress("0xa332277be213f56221b6140998c03d860a60e1f8")
	backfrontReciepent = account.HexToAddress("0xa432277be213f56221b6140998c03d860a60e1f8")

	issueAmount = int64(100)
	Amount      = int64(1)
	fee         = int64(0)
)

func TestExecuteIssueTx(t *testing.T) {
	vm.VMConf = vm.DefaultConfig()
	testDb := db.NewDB(db.DefaultConfig())
	li := NewLedger(testDb)
	defer os.RemoveAll("/tmp/rocksdb-test")
	//config.ChainID = []byte{byte(0)}

	issueTxKeypair, _ := crypto.GenerateKey()
	addr := account.PublicKeyToAddress(*issueTxKeypair.Public())
	issueTx := pb.NewTransaction(account.NewChainCoordinate([]byte{byte(0)}),
		account.NewChainCoordinate([]byte{byte(0)}),
		pb.TransactionType_Issue,
		uint32(0),
		addr,
		issueReciepent,
		uint32(0),
		issueAmount,
		fee,
		utils.CurrentTimestamp())
	issueCoin := make(map[string]interface{})
	issueCoin["id"] = 0
	issueTx.Payload, _ = json.Marshal(issueCoin)
	signature, _ := issueTxKeypair.Sign(issueTx.Hash().Bytes())
	issueTx.GetHeader().Signature = signature.Bytes()

	li.AppendBlock(&pb.Block{Transactions: []*pb.Transaction{issueTx},
		Header: &pb.BlockHeader{}}, true)

	sender := issueTx.Sender()
	t.Logf("sender address: %s \n", sender.String())
	t.Log(li.GetBalance(sender))
	t.Log(li.GetBalance(issueReciepent))
}

// func TestExecuteAtmoicTx(t *testing.T) {

// 	testDb := db.NewDB(db.DefaultConfig())
// 	li := NewLedger(testDb, &Config{"file"})
// 	defer os.RemoveAll("/tmp/rocksdb-test")

// 	config.ChainID = []byte{byte(0)}

// 	atmoicTxKeypair, _ := crypto.GenerateKey()
// 	addr := account.PublicKeyToAddress(*atmoicTxKeypair.Public())

// 	atmoicTx := pb.NewTransaction(account.NewChainaccount([]byte{byte(0)}),
// 		account.NewChainaccount([]byte{byte(0)}),
// 		pb.TypeAtomic,
// 		uint32(0),
// 		addr,
// 		atmoicReciepent,
// 		uint32(0),
// 		Amount,
// 		fee,
// 		utils.CurrentTimestamp())

// 	atmoicCoin := make(map[string]interface{})
// 	atmoicCoin["id"] = 0
// 	atmoicTx.Payload, _ = json.Marshal(atmoicCoin)

// 	signature, _ := atmoicTxKeypair.Sign(atmoicTx.Hash().Bytes())
// 	atmoicTx.WithSignature(signature)
// 	atmoicSender := atmoicTx.Sender()

// 	issueTxKeypair, _ := crypto.GenerateKey()
// 	addr = account.PublicKeyToAddress(*issueTxKeypair.Public())

// 	issueTx := pb.NewTransaction(account.NewChainaccount([]byte{byte(0)}),
// 		account.NewChainaccount([]byte{byte(0)}),
// 		pb.TypeIssue,
// 		uint32(0),
// 		addr,
// 		atmoicSender,
// 		uint32(0),
// 		issueAmount,
// 		fee,
// 		utils.CurrentTimestamp())
// 	issueCoin := make(map[string]interface{})
// 	issueCoin["id"] = 0
// 	issueTx.Payload, _ = json.Marshal(issueCoin)
// 	signature1, _ := issueTxKeypair.Sign(issueTx.Hash().Bytes())
// 	issueTx.WithSignature(signature1)

// 	_, _, errtxs := li.executeTransactions(pb.Transactions{issueTx, atmoicTx}, false)
// 	if len(errtxs) > 0 {
// 		t.Error("error")
// 	}

// 	li.dbHandler.AtomicWrite(li.state.WriteBatchs())

// 	sender := issueTx.Sender()
// 	b, _ := li.GetBalanceFromDB(sender)
// 	t.Logf("issueTx sender : %v", b.Amounts[0].Sign())
// 	t.Log(li.GetBalanceFromDB(atmoicSender))
// 	t.Log(li.GetBalanceFromDB(atmoicReciepent))

// }

// func TestExecuteAcossTx1(t *testing.T) {
// 	testDb := db.NewDB(db.DefaultConfig())
// 	li := NewLedger(testDb, &Config{"file"})
// 	defer os.RemoveAll("/tmp/rocksdb-test")

// 	config.ChainID = []byte{byte(0)}

// 	acrossTxKeypair, _ := crypto.GenerateKey()
// 	addr := account.PublicKeyToAddress(*acrossTxKeypair.Public())

// 	acrossTx := pb.NewTransaction(account.NewChainaccount([]byte{byte(0)}),
// 		account.NewChainaccount([]byte{byte(1)}),
// 		pb.TypeAcrossChain,
// 		uint32(0),
// 		addr,
// 		acrossReciepent,
// 		uint32(0),
// 		Amount,
// 		fee,
// 		utils.CurrentTimestamp())
// 	acrossCoin := make(map[string]interface{})
// 	acrossCoin["id"] = 0
// 	acrossTx.Payload, _ = json.Marshal(acrossCoin)
// 	signature, _ := acrossTxKeypair.Sign(acrossTx.Hash().Bytes())
// 	acrossTx.WithSignature(signature)
// 	acrossSender := acrossTx.Sender()

// 	issueTxKeypair, _ := crypto.GenerateKey()
// 	addr = account.PublicKeyToAddress(*issueTxKeypair.Public())
// 	issueTx := pb.NewTransaction(account.NewChainaccount([]byte{byte(0)}),
// 		account.NewChainaccount([]byte{byte(0)}),
// 		pb.TypeIssue,
// 		uint32(0),
// 		addr,
// 		acrossSender,
// 		uint32(0),
// 		issueAmount,
// 		fee,
// 		utils.CurrentTimestamp())
// 	issueCoin := make(map[string]interface{})
// 	issueCoin["id"] = 0
// 	issueTx.Payload, _ = json.Marshal(issueCoin)
// 	signature1, _ := issueTxKeypair.Sign(issueTx.Hash().Bytes())
// 	issueTx.WithSignature(signature1)

// 	_, _, errtxs := li.executeTransactions(pb.Transactions{issueTx, acrossTx}, false)
// 	if len(errtxs) > 0 {
// 		t.Error("error")
// 	}

// 	li.dbHandler.AtomicWrite(li.state.WriteBatchs())

// 	sender := issueTx.Sender()
// 	t.Log(li.GetBalanceFromDB(sender))
// 	t.Log(li.GetBalanceFromDB(acrossSender))
// 	t.Log(li.GetBalanceFromDB(acrossReciepent))
// }

// func TestExecuteAcossTx2(t *testing.T) {
// 	testDb := db.NewDB(db.DefaultConfig())
// 	li := NewLedger(testDb, &Config{"file"})
// 	defer os.RemoveAll("/tmp/rocksdb-test")

// 	config.ChainID = []byte{byte(0)}

// 	acrossTxKeypair, _ := crypto.GenerateKey()
// 	addr := account.PublicKeyToAddress(*acrossTxKeypair.Public())

// 	acrossTx := pb.NewTransaction(account.NewChainaccount([]byte{byte(1)}),
// 		account.NewChainaccount([]byte{byte(0)}),
// 		pb.TypeAcrossChain,
// 		uint32(0),
// 		addr,
// 		acrossReciepent,
// 		uint32(0),
// 		Amount,
// 		fee,
// 		utils.CurrentTimestamp())
// 	acrossCoin := make(map[string]interface{})
// 	acrossCoin["id"] = 0
// 	acrossTx.Payload, _ = json.Marshal(acrossCoin)
// 	signature, _ := acrossTxKeypair.Sign(acrossTx.Hash().Bytes())
// 	acrossTx.WithSignature(signature)
// 	acrossSender := acrossTx.Sender()

// 	issueTxKeypair, _ := crypto.GenerateKey()
// 	addr = account.PublicKeyToAddress(*issueTxKeypair.Public())

// 	issueTx := pb.NewTransaction(account.NewChainaccount([]byte{byte(0)}),
// 		account.NewChainaccount([]byte{byte(0)}),
// 		pb.TypeIssue,
// 		uint32(0),
// 		addr,
// 		acrossReciepent,
// 		uint32(0),
// 		issueAmount,
// 		fee,
// 		utils.CurrentTimestamp())
// 	issueCoin := make(map[string]interface{})
// 	issueCoin["id"] = 0
// 	issueTx.Payload, _ = json.Marshal(issueCoin)
// 	signature1, _ := issueTxKeypair.Sign(issueTx.Hash().Bytes())
// 	issueTx.WithSignature(signature1)

// 	_, _, errtxs := li.executeTransactions(pb.Transactions{issueTx, acrossTx}, false)
// 	if len(errtxs) > 0 {
// 		t.Error("error")
// 	}
// 	li.dbHandler.AtomicWrite(li.state.WriteBatchs())

// 	sender := issueTx.Sender()
// 	t.Log(li.GetBalanceFromDB(sender))
// 	t.Log(li.GetBalanceFromDB(acrossSender))
// 	t.Log(li.GetBalanceFromDB(acrossReciepent))
// }

// func TestExecuteMergedTx(t *testing.T) {
// 	testDb := db.NewDB(db.DefaultConfig())
// 	li := NewLedger(testDb, &Config{"file"})
// 	defer os.RemoveAll("/tmp/rocksdb-test")

// 	config.ChainID = []byte{byte(0)}

// 	from := account.NewChainaccount([]byte{byte(0)})
// 	to := account.NewChainaccount([]byte{byte(0)})
// 	sender := account.NewChainaccount([]byte{byte(0), byte(0)})
// 	reciepent := account.NewChainaccount([]byte{byte(0), byte(1)})

// 	mergedTx := pb.NewTransaction(from,
// 		to,
// 		pb.TypeMerged,
// 		uint32(0),
// 		account.ChainaccountToAddress(sender),
// 		account.ChainaccountToAddress(reciepent),
// 		uint32(0),
// 		Amount,
// 		fee,
// 		utils.CurrentTimestamp())
// 	mergeCoin := make(map[string]interface{})
// 	mergeCoin["id"] = 0
// 	mergedTx.Payload, _ = json.Marshal(mergeCoin)

// 	senderAddress := account.ChainaccountToAddress(sender)
// 	sig := &crypto.Signature{}
// 	copy(sig[:], senderAddress[:])
// 	mergedTx.WithSignature(sig)

// 	issueTxKeypair, _ := crypto.GenerateKey()
// 	addr := account.PublicKeyToAddress(*issueTxKeypair.Public())

// 	issueTx := pb.NewTransaction(account.NewChainaccount([]byte{byte(0)}),
// 		account.NewChainaccount([]byte{byte(0)}),
// 		pb.TypeIssue,
// 		uint32(0),
// 		addr,
// 		senderAddress,
// 		uint32(0),
// 		issueAmount,
// 		fee,
// 		utils.CurrentTimestamp())

// 	issueCoin := make(map[string]interface{})
// 	issueCoin["id"] = 0
// 	issueTx.Payload, _ = json.Marshal(issueCoin)

// 	signature1, _ := issueTxKeypair.Sign(issueTx.Hash().Bytes())
// 	issueTx.WithSignature(signature1)

// 	_, _, errtxs := li.executeTransactions(pb.Transactions{issueTx, mergedTx}, false)
// 	if len(errtxs) > 0 {
// 		t.Error("error")
// 	}
// 	li.dbHandler.AtomicWrite(li.state.WriteBatchs())

// 	issueSenderaddress := issueTx.Sender()

// 	t.Log(li.GetBalanceFromDB(issueSenderaddress))
// 	t.Log(li.GetBalanceFromDB(account.ChainaccountToAddress(sender)))
// 	t.Log(li.GetBalanceFromDB(account.ChainaccountToAddress(reciepent)))
// }

// func TestExecuteDistributTx(t *testing.T) {
// 	testDb := db.NewDB(db.DefaultConfig())
// 	li := NewLedger(testDb, &Config{"file"})
// 	defer os.RemoveAll("/tmp/rocksdb-test")

// 	config.ChainID = []byte{byte(0)}

// 	distributTxKeypair, _ := crypto.GenerateKey()
// 	addr := account.PublicKeyToAddress(*distributTxKeypair.Public())

// 	distributTx := pb.NewTransaction(account.NewChainaccount([]byte{byte(0)}),
// 		account.NewChainaccount([]byte{byte(1)}),
// 		pb.TypeDistribut,
// 		uint32(1),
// 		addr,
// 		distributReciepent,
// 		uint32(0),
// 		Amount,
// 		fee,
// 		utils.CurrentTimestamp())

// 	distributCoin := make(map[string]interface{})
// 	distributCoin["id"] = 0
// 	distributTx.Payload, _ = json.Marshal(distributCoin)

// 	signature, _ := distributTxKeypair.Sign(distributTx.Hash().Bytes())
// 	distributTx.WithSignature(signature)
// 	distributAddress := distributTx.Sender()

// 	issueTxKeypair, _ := crypto.GenerateKey()
// 	addr = account.PublicKeyToAddress(*issueTxKeypair.Public())
// 	issueTx := pb.NewTransaction(account.NewChainaccount([]byte{byte(0)}),
// 		account.NewChainaccount([]byte{byte(0)}),
// 		pb.TypeIssue,
// 		uint32(0),
// 		addr,
// 		distributAddress,
// 		uint32(0),
// 		issueAmount,
// 		fee,
// 		utils.CurrentTimestamp())
// 	issueCoin := make(map[string]interface{})
// 	issueCoin["id"] = 0
// 	issueTx.Payload, _ = json.Marshal(issueCoin)

// 	signature1, _ := issueTxKeypair.Sign(issueTx.Hash().Bytes())
// 	issueTx.WithSignature(signature1)

// 	_, _, errtxs := li.executeTransactions(pb.Transactions{issueTx, distributTx}, false)
// 	if len(errtxs) > 0 {
// 		t.Error("error")
// 	}
// 	li.dbHandler.AtomicWrite(li.state.WriteBatchs())

// 	sender := issueTx.Sender()
// 	t.Log(li.GetBalanceFromDB(sender))
// 	t.Log(li.GetBalanceFromDB(account.ChainaccountToAddress(account.NewChainaccount([]byte{byte(1)}))))
// 	t.Log(li.GetBalanceFromDB(distributAddress))
// 	t.Log(li.GetBalanceFromDB(distributReciepent))
// }

// func TestExecuteBackfrontTx(t *testing.T) {
// 	testDb := db.NewDB(db.DefaultConfig())
// 	li := NewLedger(testDb, &Config{"file"})
// 	defer os.RemoveAll("/tmp/rocksdb-test")

// 	config.ChainID = []byte{byte(1)}

// 	backfrontTxKeypair, _ := crypto.GenerateKey()
// 	addr := account.PublicKeyToAddress(*backfrontTxKeypair.Public())

// 	backfrontTx := pb.NewTransaction(account.NewChainaccount([]byte{byte(0)}),
// 		account.NewChainaccount([]byte{byte(1)}),
// 		pb.TypeBackfront,
// 		uint32(1),
// 		addr,
// 		backfrontReciepent,
// 		uint32(0),
// 		Amount,
// 		fee,
// 		utils.CurrentTimestamp())
// 	backforntCoin := make(map[string]interface{})
// 	backforntCoin["id"] = 0
// 	backfrontTx.Payload, _ = json.Marshal(backforntCoin)
// 	signature, _ := backfrontTxKeypair.Sign(backfrontTx.Hash().Bytes())
// 	backfrontTx.WithSignature(signature)
// 	backfrontAddrress := backfrontTx.Sender()

// 	issueTxKeypair, _ := crypto.GenerateKey()
// 	addr = account.PublicKeyToAddress(*issueTxKeypair.Public())
// 	issueTx := pb.NewTransaction(account.NewChainaccount([]byte{byte(0)}),
// 		account.NewChainaccount([]byte{byte(0)}),
// 		pb.TypeIssue,
// 		uint32(0),
// 		addr,
// 		backfrontAddrress,
// 		uint32(0),
// 		issueAmount,
// 		fee,
// 		utils.CurrentTimestamp())
// 	issueCoin := make(map[string]interface{})
// 	issueCoin["id"] = 0
// 	issueTx.Payload, _ = json.Marshal(issueCoin)
// 	signature1, _ := issueTxKeypair.Sign(issueTx.Hash().Bytes())
// 	issueTx.WithSignature(signature1)

// 	_, _, errtxs := li.executeTransactions(pb.Transactions{issueTx}, false)
// 	if len(errtxs) > 0 {
// 		t.Error("error")
// 	}

// 	li.dbHandler.AtomicWrite(li.state.WriteBatchs())

// 	sender := issueTx.Sender()
// 	t.Log(li.GetBalanceFromDB(sender))
// 	t.Log(li.GetBalanceFromDB(account.ChainaccountToAddress(account.NewChainaccount([]byte{byte(0)}))))
// 	t.Log(li.GetBalanceFromDB(backfrontAddrress))
// 	t.Log(li.GetBalanceFromDB(backfrontReciepent))

// }
