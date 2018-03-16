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
	"bytes"
	"encoding/json"
	"time"

	"github.com/zipper-project/zipper/account"
	"github.com/zipper-project/zipper/common/crypto"
	"github.com/zipper-project/zipper/common/db"
	"github.com/zipper-project/zipper/common/log"
	"github.com/zipper-project/zipper/common/mpool"
	"github.com/zipper-project/zipper/common/utils"
	"github.com/zipper-project/zipper/ledger/balance"
	"github.com/zipper-project/zipper/ledger/blockstorage"
	"github.com/zipper-project/zipper/ledger/state"
	"github.com/zipper-project/zipper/params"
	pb "github.com/zipper-project/zipper/proto"
	"github.com/zipper-project/zipper/vm"
	"github.com/zipper-project/zipper/vm/bsvm"
)

var (
	ledgerInstance *Ledger
)

// Ledger represents the ledger in blockchain
type Ledger struct {
	dbHandler *db.BlockchainDB
	block     *blockstorage.Blockchain
	state     *state.BLKRWSet
	mdbChan   chan []*db.WriteBatch
	vmEnv     map[string]*mpool.VirtualMachine
}

// NewLedger returns the ledger instance
func NewLedger(kvdb *db.BlockchainDB) *Ledger {
	if ledgerInstance == nil {
		ledgerInstance = &Ledger{
			dbHandler: kvdb,
			block:     blockstorage.NewBlockchain(kvdb),
			state:     state.NewBLKRWSet(kvdb),
		}
		ledgerInstance.init()
		_, err := ledgerInstance.Height()
		if err != nil {
			log.Error(err)
		}
		ledgerInstance.initVmEnv()
	}
	return ledgerInstance
}

func (ledger *Ledger) DBHandler() *db.BlockchainDB {
	return ledger.dbHandler
}

func (ledger *Ledger) initVmEnv() {
	ledger.vmEnv = make(map[string]*mpool.VirtualMachine)
	bsWorkers := make([]mpool.VmWorker, vm.VMConf.BsWorkerCnt)
	for i := 0; i < vm.VMConf.BsWorkerCnt; i++ {
		bsWorkers[i] = bsvm.NewBsWorker(vm.VMConf, i)
	}
	addNewEnv := func(name string, worker []mpool.VmWorker) *mpool.VirtualMachine {
		env := mpool.CreateCustomVM(worker)
		env.Open(name)
		ledger.vmEnv[name] = env

		return env
	}
	addNewEnv("bs", bsWorkers)
}

// VerifyChain verifys the blockchain data
func (ledger *Ledger) VerifyChain() {
	height, err := ledger.Height()
	if err != nil {
		log.Panicf("VerifyChain -- Height %s", err)
	}
	currentBlockHeader, err := ledger.block.GetBlockByNumber(height)
	for i := height; i >= 1; i-- {
		previousBlockHeader, err := ledger.block.GetBlockByNumber(i - 1) // storage
		if previousBlockHeader != nil && err != nil {
			log.Panicf("VerifyChain -- GetBlockByNumber %s", err)
		}
		// verify previous block
		if previousBlockHeader.Hash().String() != currentBlockHeader.PreviousHash {
			log.Panicf("VerifyChain -- block [%d] mismatch, veifychain breaks", i)
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

	//go ledger.Validator.RemoveTxsInVerification(block.Transactions)

	ledger.state.SetBlock(block.GetHeader().GetHeight(), uint32(len(block.Transactions)))

	wokerData := func(tx *pb.Transaction, txIdx int) *vm.WorkerProc {
		return &vm.WorkerProc{
			ContractData: vm.NewContractData(tx),
			SCHandler:    state.NewTXRWSet(ledger.state, tx, uint32(txIdx)),
		}
	}

	//log.Debugf("appendBlock cnt: %+v ...........", len(block.Transactions))
	startTime := time.Now()
	vm.NewTxSync(vm.VMConf.BsWorkerCnt)
	for idx, tx := range block.Transactions {
		ledger.vmEnv["bs"].SendWorkCleanAsync(&vm.WorkerProcWithCallback{
			WorkProc: wokerData(tx, idx),
			Idx:      idx,
		})
	}

	writeBatches, oktxs, errtxs, err := ledger.state.ApplyChanges()
	if err != nil || len(errtxs) != 0 {
		//TODO
		log.Errorf("AppendBlock Err: %+v, errtxs: %+v", err, len(errtxs))
	}

	execTime := time.Now().Sub(startTime)
	blkHt, _ := ledger.Height()
	log.Warnf("appendBlock cnt: %+v, oktxs: %+v, errtxs: %+v, blkht: %+v, execTime: %s ...........", len(block.Transactions), len(oktxs), len(errtxs), blkHt, execTime)

	block.Transactions = oktxs
	block.Header.TxsMerkleHash = merkleRootHash(block.Transactions).String()
	block.Header.StateHash = ledger.state.RootHash().String()
	blkWriteBatches := ledger.block.AppendBlock(block)
	writeBatches = append(writeBatches, blkWriteBatches...)
	return ledger.dbHandler.AtomicWrite(writeBatches)
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

//ComplexQuery com
func (ledger *Ledger) ComplexQuery(key string) ([]byte, error) {
	return ledger.state.ComplexQuery(key)
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

// GetBalance returns balance by account
func (ledger *Ledger) GetBalance(addr account.Address) (*balance.Balance, error) {
	return ledger.state.GetBalances(addr.String())
}

// GetAsset returns asset
func (ledger *Ledger) GetAsset(id uint32) (*state.Asset, error) {
	return ledger.state.GetAsset(id)
}

// GetAssets returns assets
func (ledger *Ledger) GetAssets() (map[uint32]*state.Asset, error) {
	return ledger.state.GetAssets()
}

//QueryContract processes new contract query transaction
func (ledger *Ledger) QueryContract(tx *pb.Transaction) ([]byte, error) {
	//return ledger.state.QueryContract(tx)
	return nil, nil
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

	// admin address
	buf, err := state.ConcrateStateJson(state.DefaultAdminAddr)
	if err != nil {
		return err
	}

	writeBatchs = append(writeBatchs,
		db.NewWriteBatch(ledger.state.GetChainCodeCF(),
			db.OperationPut,
			[]byte(state.ConstructCompositeKey(params.GlobalStateKey, params.AdminKey)),
			buf.Bytes(), ledger.state.GetChainCodeCF()))

	// global contract
	buf, err = state.ConcrateStateJson(&vm.ContractCode{
		state.DefaultGlobalContractCode,
		state.DefaultGlobalContractType,
	})
	if err != nil {
		return err
	}

	writeBatchs = append(writeBatchs,
		db.NewWriteBatch(ledger.state.GetChainCodeCF(),
			db.OperationPut,
			[]byte(state.ConstructCompositeKey(params.GlobalStateKey, params.GlobalContractKey)),
			buf.Bytes(), ledger.state.GetChainCodeCF()))

	err = ledger.dbHandler.AtomicWrite(writeBatchs)
	if err != nil {
		return err
	}
	return err
}

func (ledger *Ledger) checkCoordinate(tx *pb.Transaction) bool {
	fromChainID := tx.GetHeader().GetFromChain()
	toChainID := tx.GetHeader().GetToChain()
	if bytes.Equal(fromChainID, toChainID) {
		return true
	}
	return false
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
