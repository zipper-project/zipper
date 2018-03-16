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


package proto

import (
	"fmt"
	"math/big"
	"testing"

	"github.com/zipper-project/zipper/account"
	"github.com/zipper-project/zipper/common/crypto"
	"github.com/zipper-project/zipper/common/utils"
)

var testTx = getTestTransaction()
var testTxHex = fmt.Sprintf("%0x", testTx.Serialize())

func getTestTransaction() *Transaction {
	sender := account.HexToAddress("0xc9bc867a613381f35b4430a6cb712eff8bb50311")
	address := account.HexToAddress("0xc9bc867a613381f35b4430a6cb712eff8bb50310")
	fromChain := account.NewChainCoordinate([]byte{0, 1, 3})
	toChain := account.NewChainCoordinate([]byte{0, 1, 1})
	nonce := uint32(10000)
	tx := &Transaction{
		Header: &TxHeader{
			FromChain:  fromChain,
			ToChain:    toChain,
			Type:       TransactionType_Atomic,
			Nonce:      nonce,
			Sender:     sender.String(),
			Recipient:  address.String(),
			AssetID:    1,
			Amount:     big.NewInt(1100).Int64(),
			Fee:        big.NewInt(110).Int64(),
			CreateTime: utils.CurrentTimestamp(),
		},
	}
	tx.Payload = []byte("123456")
	return tx
}

func TestTxDeserialize(t *testing.T) {
	txBytes := utils.HexToBytes(testTxHex)
	tx := new(Transaction)
	tx.Deserialize(txBytes)
	utils.AssertEquals(t, tx.Serialize(), txBytes)
}

func TestTxSender(t *testing.T) {
	var (
		priv, _ = crypto.GenerateKey()
		addr    = account.PublicKeyToAddress(*priv.Public())
	)

	tx := &Transaction{
		Header: &TxHeader{
			FromChain:  nil,
			ToChain:    nil,
			Type:       TransactionType_Atomic,
			Nonce:      1,
			Sender:     addr.String(),
			Recipient:  addr.String(),
			AssetID:    1,
			Amount:     big.NewInt(1100).Int64(),
			Fee:        big.NewInt(110).Int64(),
			CreateTime: utils.CurrentTimestamp(),
		},
	}

	sig, _ := priv.Sign(tx.SignHash().Bytes())
	tx.GetHeader().Signature = sig.Bytes()

	tx2 := new(Transaction)
	tx2.Deserialize(tx.Serialize())
	utils.AssertEquals(t, tx.Serialize(), tx2.Serialize())

}
