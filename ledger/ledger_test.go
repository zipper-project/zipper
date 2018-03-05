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
package ledger

import (
	"encoding/json"
	"os"
	"testing"
	"time"

	"github.com/zipper-project/zipper/account"
	"github.com/zipper-project/zipper/common/crypto"
	"github.com/zipper-project/zipper/common/db"
	pb "github.com/zipper-project/zipper/proto"
)

var (
	issueReciepent  = account.HexToAddress("0xa032277be213f56221b6140998c03d860a60e1f8")
	atmoicReciepent = account.HexToAddress("0xa132277be213f56221b6140998c03d860a60e1f8")

	issueAmount = int64(100)
	Amount      = int64(1)
	fee         = int64(0)
)

func TestExecuteAtmoicTx(t *testing.T) {
	testDb := db.NewDB(db.DefaultConfig())
	li := NewLedger(testDb, &Config{"file"})
	defer os.RemoveAll("/tmp/rocksdb-test")

	//config.ChainID = []byte{byte(0)}

	atmoicTxKeypair, _ := crypto.GenerateKey()
	addr := account.PublicKeyToAddress(*atmoicTxKeypair.Public())

	atmoicTx := &pb.Transaction{
		TxData: pb.TxData{
			Header: &pb.TxHeader{
				FromChain:  account.NewChainCoordinate([]byte("0")),
				ToChain:    account.NewChainCoordinate([]byte("0")),
				Type:       pb.TransactionType_Atomic,
				Nonce:      uint32(0),
				Sender:     addr.String(),
				Recipient:  atmoicReciepent.String(),
				AssetID:    uint32(0),
				Amount:     Amount,
				Fee:        fee,
				CreateTime: uint32(time.Now().Unix()),
			},
		},
	}
	atmoicCoin := make(map[string]interface{})
	atmoicCoin["id"] = 0
	atmoicTx.Payload, _ = json.Marshal(atmoicCoin)

	signature, _ := atmoicTxKeypair.Sign(atmoicTx.Hash().Bytes())
	atmoicTx.WithSignature(signature)
	atmoicSender := atmoicTx.Sender()

	issueTxKeypair, _ := crypto.GenerateKey()
	addr = account.PublicKeyToAddress(*issueTxKeypair.Public())

	issueTx := &pb.Transaction{
		TxData: pb.TxData{
			Header: &pb.TxHeader{
				FromChain:  account.NewChainCoordinate([]byte("0")),
				ToChain:    account.NewChainCoordinate([]byte("0")),
				Type:       pb.TransactionType_Issue,
				Nonce:      uint32(0),
				Sender:     addr.String(),
				Recipient:  atmoicSender.String(),
				AssetID:    uint32(0),
				Amount:     issueAmount,
				Fee:        fee,
				CreateTime: uint32(time.Now().Unix()),
			},
		},
	}
	issueCoin := make(map[string]interface{})
	issueCoin["id"] = 0
	issueTx.Payload, _ = json.Marshal(issueCoin)
	signature1, _ := issueTxKeypair.Sign(issueTx.Hash().Bytes())
	issueTx.WithSignature(signature1)

	// _, _, errtxs := li.executeTransactions(pb.Transactions{issueTx, atmoicTx}, false)
	// if len(errtxs) > 0 {
	// 	t.Error("error")
	// }

	li.dbHandler.AtomicWrite(li.state.WriteBatchs())

	sender := issueTx.Sender()
	b, _ := li.GetBalanceFromDB(sender)
	t.Logf("issueTx sender : %v", b.Amounts[0])
	t.Log(li.GetBalanceFromDB(atmoicSender))
	t.Log(li.GetBalanceFromDB(atmoicReciepent))
}
