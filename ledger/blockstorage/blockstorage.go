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
	"errors"

	"github.com/zipper-project/zipper/common/crypto"
	"github.com/zipper-project/zipper/common/db"
	"github.com/zipper-project/zipper/common/utils"
	pb "github.com/zipper-project/zipper/proto"
)

const (
	heightKey string = "blockLastHeight"
)

// Blockchain represents block
type Blockchain struct {
	dbHandler               *db.BlockchainDB
	txPrefix                []byte
	blockColumnFamily       string
	transactionColumnFamily string
	indexColumnFamily       string
}

// NewBlockchain initialization
func NewBlockchain(db *db.BlockchainDB) *Blockchain {
	return &Blockchain{
		dbHandler:               db,
		txPrefix:                []byte("tx_"),
		blockColumnFamily:       "block",
		transactionColumnFamily: "transaction",
		indexColumnFamily:       "index",
	}
}

// GetBlockByHash gets block by block hash
func (blockchain *Blockchain) GetBlockByHash(blockHash []byte) (*pb.BlockHeader, error) {
	blockHeaderBytes, err := blockchain.dbHandler.Get(blockchain.blockColumnFamily, blockHash)
	if err != nil {
		return nil, err
	}

	if len(blockHeaderBytes) == 0 {
		return nil, errors.New("not found block. ")
	}

	blockHeader := new(pb.BlockHeader)
	if err := blockHeader.Deserialize(blockHeaderBytes); err != nil {
		return nil, err
	}

	return blockHeader, nil

}

//GetTransactionHashList get transaction hash list by block height
func (blockchain *Blockchain) GetTransactionHashList(blockHeight uint32) ([]byte, error) {
	txHashsBytes, err := blockchain.dbHandler.Get(blockchain.indexColumnFamily, prependKeyPrefix(blockchain.txPrefix, utils.Uint32ToBytes(blockHeight)))
	if err != nil {
		return nil, err
	}
	if len(txHashsBytes) == 0 && blockHeight != 0 {
		return nil, errors.New("not found transactions")
	}

	return txHashsBytes, nil
}

// GetBlockByNumber gets block by block height number
func (blockchain *Blockchain) GetBlockByNumber(blockNum uint32) (*pb.BlockHeader, error) {
	blockHashBytes, err := blockchain.GetBlockHashByNumber(blockNum)
	if err != nil {
		return nil, err
	}
	return blockchain.GetBlockByHash(blockHashBytes)
}

// GetTransactionsByNumber by block height number
func (blockchain *Blockchain) GetTransactionsByNumber(blockNum uint32, transactionType uint32) (pb.Transactions, error) {
	txHashsBytes, err := blockchain.GetTransactionHashList(blockNum)
	if err != nil {
		return nil, err
	}

	txHashs := []crypto.Hash{}

	utils.Deserialize(txHashsBytes, &txHashs)

	return blockchain.getTransactionsByHashList(txHashs, transactionType)
}

// GetTransactionsByHash by block hash
func (blockchain *Blockchain) GetTransactionsByHash(blockHash []byte, transactionType uint32) (pb.Transactions, error) {

	blockHeader, err := blockchain.GetBlockByHash(blockHash)
	if err != nil {
		return nil, err
	}

	txHashsBytes, err := blockchain.GetTransactionHashList(blockHeader.Height)
	if err != nil {
		return nil, err
	}

	txHashs := []crypto.Hash{}

	utils.Deserialize(txHashsBytes, &txHashs)

	return blockchain.getTransactionsByHashList(txHashs, transactionType)

}

// GetTransactionByTxHash gets transaction by transaction hash
func (blockchain *Blockchain) GetTransactionByTxHash(txHash []byte) (*pb.Transaction, error) {
	txBytes, err := blockchain.dbHandler.Get(blockchain.transactionColumnFamily, txHash)
	if err != nil {
		return nil, err
	}

	if len(txBytes) == 0 {
		return nil, errors.New("not found transaction by txHash")
	}

	tx := new(pb.Transaction)

	if err := tx.Deserialize(txBytes); err != nil {
		return nil, err
	}
	return tx, nil
}

// GetBlockchainHeight gets blockchain height
func (blockchain *Blockchain) GetBlockchainHeight() (uint32, error) {
	heightBytes, _ := blockchain.dbHandler.Get(blockchain.indexColumnFamily, []byte(heightKey))
	if len(heightBytes) == 0 {
		return 0, errors.New("failed to get the height")
	}
	height := utils.BytesToUint32(heightBytes)
	return height, nil
}

// AppendBlock appends a block
func (blockchain *Blockchain) AppendBlock(block *pb.Block) []*db.WriteBatch {
	blockHashBytes := block.Hash().Bytes()
	height := block.GetHeader().GetHeight()
	blockHeightBytes := utils.Uint32ToBytes(height)

	// storage
	var (
		writeBatchs []*db.WriteBatch
		txHashs     []crypto.Hash
	)

	writeBatchs = append(writeBatchs, db.NewWriteBatch(blockchain.blockColumnFamily, db.OperationPut, blockHashBytes, block.Header.Serialize(), blockchain.blockColumnFamily)) // block hash => block
	writeBatchs = append(writeBatchs, db.NewWriteBatch(blockchain.indexColumnFamily, db.OperationPut, blockHeightBytes, blockHashBytes, blockchain.indexColumnFamily))         // height => block hash
	writeBatchs = append(writeBatchs, db.NewWriteBatch(blockchain.indexColumnFamily, db.OperationPut, []byte(heightKey), blockHeightBytes, blockchain.indexColumnFamily))      // update block height

	//storage  tx hash
	for _, txData := range block.GetTxDatas() {
		tx := &pb.Transaction{TxData: *txData}
		txHashs = append(txHashs, tx.Hash())
		writeBatchs = append(writeBatchs, db.NewWriteBatch(blockchain.transactionColumnFamily, db.OperationPut, tx.Hash().Bytes(), tx.Serialize(), blockchain.transactionColumnFamily)) // tx hash => tx detail
	}
	writeBatchs = append(writeBatchs, db.NewWriteBatch(blockchain.indexColumnFamily, db.OperationPut, prependKeyPrefix(blockchain.txPrefix, blockHeightBytes), utils.Serialize(txHashs), string(blockchain.txPrefix))) // prefix + blockheight  => all tx hash

	return writeBatchs
}

//GetBlockHashByNumber get block hash by block number
func (blockchain *Blockchain) GetBlockHashByNumber(blockNum uint32) ([]byte, error) {
	currentHeight, err := blockchain.GetBlockchainHeight()

	if err != nil {
		return nil, err
	}
	if blockNum > currentHeight {
		return nil, errors.New("exceeds the max height")
	}
	blockHashBytes, err := blockchain.dbHandler.Get(blockchain.indexColumnFamily, utils.Uint32ToBytes(blockNum))
	if err != nil {
		return nil, err
	}

	if len(blockHashBytes) == 0 {
		return nil, errors.New("not found block Hash")
	}
	return blockHashBytes, nil
}

func (blockchain *Blockchain) getTransactionsByHashList(txHashs []crypto.Hash, transactionType uint32) (pb.Transactions, error) {
	var (
		txs pb.Transactions
	)
	for _, txHash := range txHashs {
		tx, err := blockchain.GetTransactionByTxHash(txHash.Bytes())
		if err != nil {
			return nil, err
		}

		if uint32(100) == transactionType {
			txs = append(txs, tx)
		} else {
			if tx.GetType() == pb.TransactionType(transactionType) {
				txs = append(txs, tx)
			}
		}
	}
	return txs, nil
}

//GetBlockCF return block columnFamily
func (blockchain *Blockchain) GetBlockCF() string {
	return blockchain.blockColumnFamily
}

//GetTransactionCF return transaction columnFamily
func (blockchain *Blockchain) GetTransactionCF() string {
	return blockchain.transactionColumnFamily
}

//GetIndexCF return index of block and transaction columnFamily
func (blockchain *Blockchain) GetIndexCF() string {
	return blockchain.indexColumnFamily
}

//GetTransactionInBlock return transactions of block
func (blockchain *Blockchain) GetTransactionInBlock(data []byte, typ string) (map[string][]crypto.Hash, bool) {
	var txHashes []crypto.Hash
	txsMap := make(map[string][]crypto.Hash)
	switch typ {
	case string(blockchain.txPrefix):
		if data != nil {
			utils.Deserialize(data, &txHashes)
		}
		txsMap["txs"] = txHashes
		return txsMap, true
	}
	return nil, false
}

func removeKeyPrefix(data []byte, prefix []byte) []byte {
	prefixLen := len(prefix)
	return data[prefixLen:]
}

func prependKeyPrefix(prefix []byte, key []byte) []byte {
	modifiedKey := []byte{}
	modifiedKey = append(modifiedKey, prefix...)
	modifiedKey = append(modifiedKey, key...)
	return modifiedKey
}
