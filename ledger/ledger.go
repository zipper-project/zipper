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
	"bytes"
	"encoding/json"
	"io/ioutil"
	"path/filepath"

	"fmt"

	"errors"
	"strconv"
	"strings"

	"github.com/zipper-project/zipper/account"
	"github.com/zipper-project/zipper/common/crypto"
	"github.com/zipper-project/zipper/common/db"
	"github.com/zipper-project/zipper/common/log"
	"github.com/zipper-project/zipper/common/utils"
	"github.com/zipper-project/zipper/config"
	"github.com/zipper-project/zipper/ledger/balance"
	"github.com/zipper-project/zipper/ledger/blockstorage"
	"github.com/zipper-project/zipper/ledger/contract"
	"github.com/zipper-project/zipper/ledger/state"
	pb "github.com/zipper-project/zipper/proto"
)

var (
	ledgerInstance *Ledger
)

// Ledger represents the ledger in blockchain
type Ledger struct {
	dbHandler *db.BlockchainDB
	block     *blockstorage.Blockchain
	state     *state.State
	contract  *contract.SmartConstract
	conf      *Config
	mdbChan   chan []*db.WriteBatch
}

// NewLedger returns the ledger instance
func NewLedger(kvdb *db.BlockchainDB, conf *Config) *Ledger {
	if ledgerInstance == nil {
		ledgerInstance = &Ledger{
			dbHandler: kvdb,
			block:     blockstorage.NewBlockchain(kvdb),
			state:     state.NewState(kvdb),
		}
		ledgerInstance.contract = contract.NewSmartConstract(kvdb, ledgerInstance)
		_, err := ledgerInstance.Height()
		if err != nil {
			ledgerInstance.init()
		}
	}
	return ledgerInstance
}

func (ledger *Ledger) DBHandler() *db.BlockchainDB {
	return ledger.dbHandler
}

func (ledger *Ledger) reOrgBatches(batches []*db.WriteBatch) map[string][]*db.WriteBatch {
	reBatches := make(map[string][]*db.WriteBatch)
	for _, batch := range batches {
		columnName := batch.CfName
		batchKey := batch.Key
		if 0 == strings.Compare(batch.CfName, ledger.contract.GetColumnFamily()) {
			keys := strings.SplitN(string(batch.Key), "|", 2)
			if len(keys) != 2 {
				continue
			}
			columnName = "contract|" + keys[0]
			batchKey = []byte(keys[1])
		}

		if _, ok := reBatches[columnName]; !ok {
			reBatches[columnName] = make([]*db.WriteBatch, 0)
		}

		reBatches[columnName] = append(reBatches[columnName], db.NewWriteBatch(columnName, batch.Operation, batchKey, batch.Value, batch.Typ))
	}

	return reBatches
}

// VerifyChain verifys the blockchain data
func (ledger *Ledger) VerifyChain() {
	height, err := ledger.Height()
	if err != nil {
		panic(err)
	}
	currentBlockHeader, err := ledger.block.GetBlockByNumber(height)
	for i := height; i >= 1; i-- {
		previousBlockHeader, err := ledger.block.GetBlockByNumber(i - 1) // storage
		if previousBlockHeader != nil && err != nil {

			log.Debug("get block err")
			panic(err)
		}
		// verify previous block
		if previousBlockHeader.Hash().String() != currentBlockHeader.PreviousHash {
			panic(fmt.Errorf("block [%d], veifychain breaks", i))
		}
		currentBlockHeader = previousBlockHeader
	}
}

// GetGenesisBlock returns the genesis block of the ledger
func (ledger *Ledger) GetGenesisBlock() *pb.BlockHeader {

	genesisBlockHeader, err := ledger.GetBlockByNumber(0)
	if err != nil {
		panic(err)
	}
	return genesisBlockHeader
}

// AppendBlock appends a new block to the ledger,flag = true pack up block ,flag = false sync block
func (ledger *Ledger) AppendBlock(block *pb.Block, flag bool) error {
	var (
		txWriteBatchs []*db.WriteBatch
		txs           pb.Transactions
	)

	bh, _ := ledger.Height()
	ledger.contract.StartConstract(bh)

	txWriteBatchs, txs, errTxs = ledger.executeTransactions(block.GetTxDatas(), flag)
	block.Header.TxsMerkleHash = merkleRootHash(txs).String()

	writeBatchs := ledger.block.AppendBlock(block)
	writeBatchs = append(writeBatchs, txWriteBatchs...)
	writeBatchs = append(writeBatchs, ledger.state.WriteBatchs()...)
	if err := ledger.dbHandler.AtomicWrite(writeBatchs); err != nil {
		return err
	}

	ledger.contract.StopContract(bh)

	return nil
}

// GetBlockByNumber gets the block by the given number
func (ledger *Ledger) GetBlockByNumber(number uint32) (*pb.BlockHeader, error) {
	return ledger.block.GetBlockByNumber(number)
}

// GetBlockByHash returns the block detail by hash
func (ledger *Ledger) GetBlockByHash(blockHashBytes []byte) (*pb.BlockHeader, error) {
	return ledger.block.GetBlockByHash(blockHashBytes)
}

//GetTransactionHashList returns transactions hash list by block number
func (ledger *Ledger) GetTransactionHashList(number uint32) ([]crypto.Hash, error) {

	txHashsBytes, err := ledger.block.GetTransactionHashList(number)
	if err != nil {
		return nil, err
	}

	txHashs := []crypto.Hash{}

	utils.Deserialize(txHashsBytes, &txHashs)

	return txHashs, nil
}

// Height returns height of ledger
func (ledger *Ledger) Height() (uint32, error) {
	return ledger.block.GetBlockchainHeight()
}

//GetLastBlockHash returns last block hash
func (ledger *Ledger) GetLastBlockHash() (crypto.Hash, error) {
	height, err := ledger.block.GetBlockchainHeight()
	if err != nil {
		return crypto.Hash{}, err
	}
	lastBlock, err := ledger.block.GetBlockByNumber(height)
	if err != nil {
		return crypto.Hash{}, err
	}
	return lastBlock.Hash(), nil
}

//GetBlockHashByNumber returns block hash by block number
func (ledger *Ledger) GetBlockHashByNumber(blockNum uint32) (crypto.Hash, error) {

	hashBytes, err := ledger.block.GetBlockHashByNumber(blockNum)
	if err != nil {
		return crypto.Hash{}, err
	}

	blockHash := new(crypto.Hash)

	blockHash.SetBytes(hashBytes)

	return *blockHash, err
}

// GetTxsByBlockHash returns transactions  by block hash and transactionType
func (ledger *Ledger) GetTxsByBlockHash(blockHashBytes []byte, transactionType uint32) (pb.Transactions, error) {
	return ledger.block.GetTransactionsByHash(blockHashBytes, transactionType)
}

//GetTxsByBlockNumber returns transactions by blcokNumber and transactionType
func (ledger *Ledger) GetTxsByBlockNumber(blockNumber uint32, transactionType uint32) (pb.Transactions, error) {
	return ledger.block.GetTransactionsByNumber(blockNumber, transactionType)
}

//GetTxByTxHash returns transaction by tx hash []byte
func (ledger *Ledger) GetTxByTxHash(txHashBytes []byte) (*pb.Transaction, error) {
	return ledger.block.GetTransactionByTxHash(txHashBytes)
}

// GetBalanceFromDB returns balance by account
func (ledger *Ledger) GetBalanceFromDB(addr account.Address) (*balance.Balance, error) {
	return ledger.state.GetBalance(addr)
}

// GetAssetFromDB returns asset
func (ledger *Ledger) GetAssetFromDB(id uint32) (*state.Asset, error) {
	return ledger.state.GetAsset(id)
}

//QueryContract processes new contract query transaction
func (ledger *Ledger) QueryContract(tx *pb.Transaction) ([]byte, error) {
	return ledger.contract.QueryContract(tx)
}

// init generates the genesis block
func (ledger *Ledger) init() error {

	// genesis block
	blockHeader := new(pb.BlockHeader)
	blockHeader.TimeStamp = uint32(0)
	blockHeader.Nonce = uint32(100)
	blockHeader.Height = 0

	genesisBlock := new(pb.Block)
	genesisBlock.Header = blockHeader
	writeBatchs := ledger.block.AppendBlock(genesisBlock)
	if err := ledger.state.UpdateAsset(0, account.Address{}, account.Address{}, "{}"); err != nil {
		panic(err)
	}
	writeBatchs = append(writeBatchs, ledger.state.WriteBatchs()...)

	// admin address
	buf, err := contract.ConcrateStateJson(contract.DefaultAdminAddr)
	if err != nil {
		return err
	}

	writeBatchs = append(writeBatchs,
		db.NewWriteBatch(contract.ColumnFamily,
			db.OperationPut,
			[]byte(contract.EnSmartContractKey(config.GlobalStateKey, config.AdminKey)),
			buf.Bytes(), contract.ColumnFamily))

	// global contract
	buf, err = contract.ConcrateStateJson(&contract.DefaultGlobalContract)
	if err != nil {
		return err
	}

	writeBatchs = append(writeBatchs,
		db.NewWriteBatch(contract.ColumnFamily,
			db.OperationPut,
			[]byte(contract.EnSmartContractKey(config.GlobalStateKey, config.GlobalContractKey)),
			buf.Bytes(), contract.ColumnFamily))

	return ledger.dbHandler.AtomicWrite(writeBatchs)

}

func (ledger *Ledger) executeTransactions(txDatas []*pb.TxData, flag bool) ([]*db.WriteBatch, pb.Transactions, pb.Transactions) {
	var (
		err                error
		adminData          []byte
		errTxs             pb.Transactions
		syncTxs            pb.Transactions
		syncContractGenTxs pb.Transactions
		writeBatchs        []*db.WriteBatch
	)

	for _, txData := range txDatas {
		tx := &pb.Transaction{TxData: *txData}
		switch tp := tx.GetType(); tp {
		case pb.TransactionType_JSContractInit, pb.TransactionType_LuaContractInit, pb.TransactionType_ContractInvoke:
			if err = ledger.executeTransaction(tx, false); err != nil {
				errTxs = append(errTxs, tx)

			}
			var ttxs pb.Transactions
			ttxs, err = ledger.executeSmartContractTx(tx)
			if err != nil {
				errTxs = append(errTxs, tx)

			} else {
				var tttxs pb.Transactions
				for _, tt := range ttxs {
					if err = ledger.executeTransaction(tt, false); err != nil {
						break
					}
					tttxs = append(tttxs, tt)
				}
				if len(tttxs) != len(ttxs) {
					for _, tt := range tttxs {
						ledger.executeTransaction(tt, true)
					}
					errTxs = append(errTxs, tx)

				}
				syncContractGenTxs = append(syncContractGenTxs, tttxs...)
			}
			syncTxs = append(syncTxs, tx)
		default:
			if err = ledger.executeTransaction(tx, false); err != nil {
				errTxs = append(errTxs, tx)

			}
			syncTxs = append(syncTxs, tx)
		}
		continue
	}

	writeBatchs, err = ledger.contract.AddChangesForPersistence(writeBatchs)
	if err != nil {
		panic(err)
	}
	if flag {
		syncTxs = append(syncTxs, syncContractGenTxs...)
	}
	return writeBatchs, syncTxs, errTxs
}

func (ledger *Ledger) executeTransaction(tx *pb.Transaction, rollback bool) error {

	return nil
}

func (ledger *Ledger) executeSmartContractTx(tx *pb.Transaction) (pb.Transactions, error) {
	return ledger.contract.ExecuteSmartContractTx(tx)
}

func (ledger *Ledger) checkCoordinate(tx *pb.Transaction) bool {
	fromChainID := account.HexToChainCoordinate(tx.FromChain()).Bytes()
	toChainID := account.HexToChainCoordinate(tx.ToChain()).Bytes()
	if bytes.Equal(fromChainID, toChainID) {
		return true
	}
	return false
}

//GetTmpBalance get balance
func (ledger *Ledger) GetTmpBalance(addr account.Address) (*balance.Balance, error) {
	balance, err := ledger.state.GetTmpBalance(addr)
	if err != nil {
		log.Error("can't get balance from db")
	}

	return balance, err
}

func (ledger *Ledger) writeBlock(data interface{}) error {
	var bvalue []byte
	switch data.(type) {
	case []*db.WriteBatch:
		orgData := data.([]*db.WriteBatch)
		bvalue = utils.Serialize(orgData)
	}

	height, err := ledger.Height()
	if err != nil {
		log.Errorf("can't get blockchain height")
	}

	path := ledger.conf.ExceptBlockDir + string(filepath.Separator)
	fileName := path + strconv.Itoa(int(height))
	if utils.FileExist(fileName) {
		log.Infof("BlockChan have error, please check ...")
		return errors.New("except block have existed")
	}

	err = ioutil.WriteFile(fileName, bvalue, 0666)
	if err != nil {
		log.Errorf("write file: %s fail err: %+v", fileName, err)
	}
	return err
}

func merkleRootHash(txs []*pb.Transaction) crypto.Hash {
	if len(txs) > 0 {
		hashs := make([]crypto.Hash, 0)
		for _, tx := range txs {
			hashs = append(hashs, tx.Hash())
		}
		return crypto.ComputeMerkleHash(hashs)[0]
	}
	return crypto.Hash{}
}

func IsJson(src []byte) bool {
	var value interface{}
	return json.Unmarshal(src, &value) == nil
}
