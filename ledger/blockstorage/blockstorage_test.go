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
package blockstorage

import (
	"math/big"
	"math/rand"
	"os"
	"testing"
	"time"

	"github.com/zipper-project/zipper/account"
	"github.com/zipper-project/zipper/common/crypto"
	"github.com/zipper-project/zipper/common/db"
	"github.com/zipper-project/zipper/common/utils"
	pb "github.com/zipper-project/zipper/proto"
)

var (
	testDb = db.NewDB(db.DefaultConfig())

	lastid    uint32
	sender    = account.HexToAddress("0xa122277be213f56221b6140998c03d860a60e1f8")
	reciepent = account.HexToAddress("0x27c649b7c4f66cfaedb99d6b38527db4deda6f41")
	amount    = big.NewInt(521000)
	fee       = big.NewInt(200)

	testTxHash       crypto.Hash
	blockHeaderBytes []byte
)

func addGenesisblock(b *Blockchain) error {
	blockHeader := new(pb.BlockHeader)
	blockHeader.TimeStamp = utils.CurrentTimestamp()
	blockHeader.Nonce = uint32(100)
	blockHeader.Height = 0

	genesisBlock := new(pb.Block)
	genesisBlock.Header = blockHeader
	writeBatchs := b.AppendBlock(genesisBlock)
	if err := b.dbHandler.AtomicWrite(writeBatchs); err != nil {
		return err
	}
	return nil
}

func TestAppendBlock(t *testing.T) {

	b := NewBlockchain(testDb)
	if err := addGenesisblock(b); err != nil {
		t.Error(err)
	}
	var previousHash crypto.Hash
	for i := 1; i < 3; i++ {
		header := new(pb.BlockHeader)
		header.TimeStamp = uint32(time.Now().Unix())
		header.Nonce = rand.Uint32()
		header.Height = uint32(i)

		header.PreviousHash = previousHash.String()

		nb := new(pb.Block)
		nb.Header = header

		// transaction
		var hashSlice []crypto.Hash
		for j := 0; j < 1; j++ {
			tx := &pb.Transaction{
				Header: &pb.TxHeader{
					FromChain:  account.NewChainCoordinate([]byte{byte(i + j)}),
					ToChain:    account.NewChainCoordinate([]byte{byte(i + j)}),
					Type:       pb.TransactionType_Atomic,
					Nonce:      rand.Uint32(),
					Sender:     sender.String(),
					Recipient:  reciepent.String(),
					AssetID:    1,
					Amount:     amount.Int64(),
					Fee:        fee.Int64(),
					CreateTime: utils.CurrentTimestamp(),
				},
			}

			//sing tx
			keypair, _ := crypto.GenerateKey()
			s, _ := keypair.Sign(tx.SignHash().Bytes())
			tx.GetHeader().Signature = s.Bytes()
			nb.Transactions = append(nb.Transactions, tx)
			hashSlice = append(hashSlice, tx.Hash())

			testTxHash = tx.Hash()
		}

		merkleHash := crypto.GetMerkleHash(hashSlice) //  utils.ComputeMerkleHash(hashSlice)
		nb.Header.TxsMerkleHash = merkleHash.String()

		writeBatchs := b.AppendBlock(nb)

		if err := b.dbHandler.AtomicWrite(writeBatchs); err != nil {
			t.Error(err)
		}

		previousHash = nb.Hash()
		t.Log("tx len: ", len(nb.Transactions), "block height:", nb.GetHeader().GetHeight())
		blockHeaderBytes = nb.Header.Serialize()
	}
	return
}

func TestGetBlockchainHeight(t *testing.T) {
	b := NewBlockchain(testDb)
	blockHeader, err := b.GetBlockByNumber(2)
	if err != nil {
		t.Error(err)
	}
	height, err := b.GetBlockchainHeight()
	if err != nil {
		t.Error(err)
	}
	t.Log("height", blockHeader.GetHeight(), height)
	utils.AssertEquals(t, height, blockHeader.GetHeight())
}

func TestGetTransactionsByNumber(t *testing.T) {
	b := NewBlockchain(testDb)
	txs, err := b.GetTransactionsByNumber(1, 0)
	if err != nil {
		t.Error(err)
	}
	t.Log("transactions len:", len(txs))

	blockHeader, err := b.GetBlockByNumber(2)
	if err != nil {
		t.Error(err)
	}
	utils.AssertEquals(t, blockHeader.Serialize(), blockHeaderBytes)
}

func TestGetTransactionByTxHash(t *testing.T) {
	b := NewBlockchain(testDb)
	tx, err := b.GetTransactionByTxHash(testTxHash.Bytes())
	if err != nil {
		t.Error(err, testTxHash)
		return
	}
	utils.AssertEquals(t, tx.Hash(), testTxHash)
	os.RemoveAll("/tmp/rocksdb-test")
}
