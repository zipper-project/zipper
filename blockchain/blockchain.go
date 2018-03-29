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

package blockchain

import (
	"bytes"
	"container/list"
	"sync"
	"time"

	"github.com/willf/bloom"
	"github.com/zipper-project/zipper/blockchain/txpool"
	"github.com/zipper-project/zipper/common/crypto"
	"github.com/zipper-project/zipper/common/db"
	"github.com/zipper-project/zipper/common/log"
	"github.com/zipper-project/zipper/config"
	"github.com/zipper-project/zipper/consensus"
	"github.com/zipper-project/zipper/consensus/consenter"
	"github.com/zipper-project/zipper/ledger"
	"github.com/zipper-project/zipper/peer"
	p2p "github.com/zipper-project/zipper/peer/proto"
	"github.com/zipper-project/zipper/proto"
)

type IInventory interface {
	Hash() crypto.Hash
	Serialize() []byte
}

// Blockchain is blockchain instance
type Blockchain struct {
	sync.RWMutex
	//ledger
	*ledger.Ledger

	// txpool
	txPool *txpool.TxPool

	// consensus
	consenter consensus.Consenter

	// server
	server *peer.Server

	currentBlockHeader *proto.BlockHeader

	filter *bloom.BloomFilter

	requestBatchSignal chan int

	quitCh chan bool

	orphans *list.List

	// 0 respresents sync block, 1 respresents sync done
	synced bool
	// 0 respresents service not start , 1 respresents service started
	started bool
}

var (
	filterN       uint = 1000000
	falsePositive      = 0.000001
)

// load loads local blockchain data
func (bc *Blockchain) load() {
	t := time.Now()
	bc.Ledger.VerifyChain()
	delay := time.Since(t)

	height, err := bc.Height()

	if err != nil {
		log.Error("GetBlockHeight error", err)
		return
	}
	bc.currentBlockHeader, err = bc.GetBlockByNumber(height)

	if bc.currentBlockHeader == nil || err != nil {
		log.Errorf("GetBlockByNumber error %v ", err)
		panic(err)
	}

	log.Debugf("Load blockchain data, bestblockhash: %s height: %d load delay : %v ", bc.currentBlockHeader.Hash(), height, delay)
}

// NewBlockchain returns a fully initialised blockchain service using input data
func NewBlockchain(pm peer.IProtocolManager) *Blockchain {
	bc := &Blockchain{
		filter:             bloom.NewWithEstimates(filterN, falsePositive),
		orphans:            list.New(),
		requestBatchSignal: make(chan int),
		quitCh:             make(chan bool),
	}

	log.Debugf("start: db.NewDB...")
	chainDb := db.NewDB(config.DBConfig())

	log.Debugf("start: ledger.NewLedger...")
	bc.Ledger = ledger.NewLedger(chainDb)
	bc.load()

	log.Debugf("start: consenter.NewConsenter...")
	bc.consenter = consenter.NewConsenter(config.ConsenterOptions(), bc)

	log.Debugf("start: peer.NewServer...")
	bc.server = peer.NewServer(config.ServerOption(), pm)

	return bc
}

func (bc *Blockchain) Start() {
	go bc.server.Start()
	blockChainLoop := func() {
		requestBatchTimer := time.NewTimer(bc.consenter.BatchTimeout())
		for {
			select {
			case <-bc.quitCh:
				return
			case cnt := <-bc.requestBatchSignal:
				if cnt >= bc.consenter.BatchSize() {
					requestBatch := bc.txPool.Txs()
					log.Debugf("request Batch: %d ", len(requestBatch))
					bc.consenter.ProcessBatch(requestBatch, bc.consensusFailed)
				} else if cnt == 1 {
					requestBatchTimer.Reset(bc.consenter.BatchTimeout())
				}
			case <-requestBatchTimer.C:
				requestBatch := bc.txPool.Txs()
				log.Debugf("request Batch Timeout: %d ", len(requestBatch))
				bc.consenter.ProcessBatch(requestBatch, bc.consensusFailed)
			case broadcastConsensusData := <-bc.consenter.BroadcastConsensusChannel():
				//TODO
				header := &p2p.Header{}
				header.ProtoID = uint32(proto.ProtoID_ConsensusWorker)
				header.MsgID = uint32(proto.MsgType_BC_OnConsensusMSg)
				msg := p2p.NewMessage(header, broadcastConsensusData.Payload)
				bc.server.Broadcast(msg, peer.VP)
			case commitedTxs := <-bc.consenter.OutputTxsChannel():
				//add lo
				log.Infof("Outputs StartConsensusService len=%d", len(commitedTxs.Txs))

				height, _ := bc.Ledger.Height()
				height++
				if commitedTxs.Height == height {
					bc.Lock()
					if !bc.synced {
						bc.synced = true
					}
					bc.Unlock()
					bc.processConsensusOutput(commitedTxs)
				} else if commitedTxs.Height > height {
					//orphan
					bc.orphans.PushBack(commitedTxs)
					for elem := bc.orphans.Front(); elem != nil; elem = elem.Next() {
						ocommitedTxs := elem.Value.(*consensus.OutputTxs)
						if ocommitedTxs.Height < height {
							bc.orphans.Remove(elem)
						} else if ocommitedTxs.Height == height {
							bc.orphans.Remove(elem)
							bc.processConsensusOutput(ocommitedTxs)
							height++
						} else {
							break
						}
					}
					if bc.orphans.Len() > 100 {
						bc.orphans.Remove(bc.orphans.Front())
					}
				} /*else if bc.synced {
					log.Panicf("Height %d already exist in ledger", commitedTxs.Height)
				}*/
			}
		}
	}
	go blockChainLoop()
	if bc.consenter.Name() == "noops" {
		bc.StartServices()
	}
}

func (bc *Blockchain) Stop() {
	bc.consenter.Stop()
	close(bc.quitCh)
	bc.server.Stop()
	bc.txPool = nil
	bc.orphans = list.New()
}

func (bc *Blockchain) GetConsenter() consensus.Consenter {
	return bc.consenter
}

// CurrentHeight returns current heigt of the current block
func (bc *Blockchain) CurrentHeight() uint32 {
	return bc.currentBlockHeader.Height
}

// CurrentBlockHash returns current block hash of the current block
func (bc *Blockchain) CurrentBlockHash() crypto.Hash {
	return bc.currentBlockHeader.Hash()
}

// Start starts blockchain services
func (bc *Blockchain) StartServices() {
	bc.Lock()
	defer bc.Unlock()
	log.Debug("BlockChain Service start")
	// start consesnus
	go bc.consenter.Start()
	// start txpool
	bc.txPool = txpool.NewTxPool()
	bc.started = true
}

func (bc *Blockchain) Started() bool {
	bc.RLock()
	defer bc.RUnlock()
	return bc.started
}

func (bc *Blockchain) Synced() bool {
	bc.RLock()
	defer bc.RUnlock()
	return bc.synced
}

// ProcessTransaction processes new transaction from the network
func (bc *Blockchain) ProcessTransaction(tx *proto.Transaction, needNotify bool) bool {
	// step 1: validate and mark transaction
	// step 2: add transaction to txPool
	// if atomic.LoadUint32(&bc.synced) == 0 {
	//log.Debugf("[Blockchain] new tx, tx_hash: %v, tx_sender: %v, tx_nonce: %v", tx.Hash().String(), tx.Sender().String(), tx.Nonce())
	if bc.txPool == nil {
		return true
	}

	// err := bc.txPool.ProcessTransaction(tx, true)
	// log.Debugf("[Blockchain] new tx, tx_hash: %v, tx_sender: %v, tx_nonce: %v, end", tx.Hash().String(), tx.Sender().String(), tx.Nonce())
	// if err != nil {
	// 	log.Errorf(fmt.Sprintf("process transaction %v failed, %v", tx.Hash(), err))
	// 	return false
	// }
	bc.requestBatchSignal <- bc.txPool.Count()
	return true
}

func (bc *Blockchain) processConsensusOutput(output *consensus.OutputTxs) {
	blk := bc.GenerateBlock(output.Txs, output.Time)
	if blk.GetHeader().GetHeight() == output.Height {
		bc.Relay(blk)
	}
}

// GenerateBlock gets transactions from consensus service and generates a new block
func (bc *Blockchain) GenerateBlock(txs proto.Transactions, createTime uint32) *proto.Block {
	var (
		// default value is empty hash
		merkleRootHash crypto.Hash
		stateRootHash  crypto.Hash
	)

	blk := proto.NewBlock(&proto.BlockHeader{bc.currentBlockHeader.Hash().Bytes(),
		merkleRootHash.Bytes(),
		stateRootHash.Bytes(),
		createTime,
		bc.currentBlockHeader.Height + 1,
		uint32(100),
	},
		txs,
	)
	return blk
}

// ProcessBlock processes new block from the network,flag = true pack up block ,flag = false sync block
func (bc *Blockchain) ProcessBlock(blk *proto.Block, flag bool) bool {
	log.Debugf("block previoushash %s, currentblockhash %s,len %d", blk.GetHeader().GetPreviousHash(), bc.CurrentBlockHash(), len(blk.Transactions))
	if bytes.Equal(blk.GetHeader().GetPreviousHash(), bc.CurrentBlockHash().Bytes()) {
		bc.AppendBlock(blk, flag)
		log.Infof("New Block  %s, height: %d Transaction Number: %d", blk.Hash(), blk.GetHeader().GetHeight(), len(blk.Transactions))
		bc.currentBlockHeader = blk.Header
		return true
	}
	return false
}

func (bc *Blockchain) Relay(inv interface{}) {
	var (
		msg *p2p.Message
	)
	switch inv.(type) {
	case *proto.Transaction:
		tx := inv.(*proto.Transaction)
		if bc.filter.TestAndAdd(tx.Serialize()) {
			log.Debugf("Bloom Test is true, txHash: %+v", tx.Hash())
			return
		}
		if bc.ProcessTransaction(tx, true) {
			log.Debugf("ProcessTransaction, tx_hash: %+v", tx.Hash())
			header := &p2p.Header{}
			header.ProtoID = uint32(proto.ProtoID_SyncWorker)
			header.MsgID = uint32(proto.MsgType_BC_OnTransactionMsg)
			txMsg := &proto.OnTransactionMsg{
				Transaction: tx,
			}

			txMsgData, _ := txMsg.Serialize()
			msg = p2p.NewMessage(header, txMsgData)
			bc.server.Broadcast(msg, peer.VP)
		}
	case *proto.Block:
		block := inv.(*proto.Block)
		if bc.filter.TestAndAdd(block.Serialize()) {
			log.Debugf("Bloom Test is true, BlockHash: %+v", block.Hash())
			return
		}

		if bc.ProcessBlock(block, true) {
			log.Debugf("ProcessTransaction, blk_hash: %+v", block.Hash())
			invMsg := &proto.GetInvMsg{}
			invMsg.Type = proto.InvType_block
			invMsg.Hashs = []string{block.Hash().String()}

			header := &p2p.Header{}
			header.ProtoID = uint32(proto.ProtoID_SyncWorker)
			header.MsgID = uint32(proto.MsgType_BC_GetInvMsg)
			imsgByte, _ := invMsg.Serialize()
			msg = p2p.NewMessage(header, imsgByte)
			bc.server.Broadcast(msg, peer.NVP)
		}
	}
}

func (bc *Blockchain) GetNextBlockHash(hash crypto.Hash) (crypto.Hash, error) {
	//TODO
	return crypto.Hash{}, nil
}
func (bc *Blockchain) consensusFailed(flag int, txs proto.Transactions) {
	if len(txs) == 0 {
		return
	}
	switch flag {
	case 0: // not used, do nothing
		log.Debug("[validator] not primary replica ...")
	case 1: //used, do nothing
		log.Debug("[validator] primary replica ...")
	case 2: // add
		for _, tx := range txs {
			bc.txPool.ProcessTransaction(tx, true)
		}
	case 3: // remove
		for _, tx := range txs {
			bc.txPool.RemoveTransaction(tx, false)
		}
	case 4: // remove
		for _, tx := range txs {
			bc.txPool.RemoveTransaction(tx, true)
		}
	default:
		log.Error("[validator] not support this flag ...")
	}
}
