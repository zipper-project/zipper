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

package validator

import (
	"bytes"
	"fmt"
	"sync"
	"time"

	"github.com/zipper-project/zipper/account"
	"github.com/zipper-project/zipper/common/crypto"
	"github.com/zipper-project/zipper/common/log"
	"github.com/zipper-project/zipper/common/utils/sortedlinkedlist"
	"github.com/zipper-project/zipper/consensus"
	"github.com/zipper-project/zipper/coordinate"
	"github.com/zipper-project/zipper/ledger"
	"github.com/zipper-project/zipper/ledger/balance"
	"github.com/zipper-project/zipper/ledger/state"
	"github.com/zipper-project/zipper/params"
	"github.com/zipper-project/zipper/proto"
)

type Validator interface {
	Start()
	ProcessTransaction(tx *proto.Transaction) error
	VerifyTxs(txs proto.Transactions) (proto.Transactions, proto.Transactions)
	UpdateAccount(tx *proto.Transaction) bool
	RollBackAccount(tx *proto.Transaction)
	RemoveTxsInVerification(txs proto.Transactions)
	GetTransactionByHash(txHash crypto.Hash) (*proto.Transaction, bool)
	GetAsset(id uint32) *state.Asset
	GetBalance(addr account.Address) *balance.Balance
}

type Verification struct {
	config             *Config
	ledger             *ledger.Ledger
	consenter          consensus.Consenter
	txpool             *sortedlinkedlist.SortedLinkedList
	requestBatchSignal chan int
	requestBatchTimer  *time.Timer
	blacklist          map[string]time.Time
	rwBlacklist        sync.RWMutex
	accounts           map[string]*balance.Balance
	rwAccount          sync.RWMutex
	assets             map[uint32]*state.Asset
	inTxs              map[crypto.Hash]*proto.Transaction
	rwInTxs            sync.RWMutex
	sync.RWMutex
	//static map[crypto.Hash]time.Duration
}

func NewVerification(config *Config, ledger *ledger.Ledger, consenter consensus.Consenter) *Verification {
	return &Verification{
		config:             config,
		ledger:             ledger,
		consenter:          consenter,
		txpool:             sortedlinkedlist.NewSortedLinkedList(),
		requestBatchSignal: make(chan int),
		requestBatchTimer:  time.NewTimer(consenter.BatchTimeout()),
		blacklist:          make(map[string]time.Time),
		accounts:           make(map[string]*balance.Balance),
		assets:             make(map[uint32]*state.Asset),
		inTxs:              make(map[crypto.Hash]*proto.Transaction),
	}
}

func (v *Verification) Start() {
	log.Info("validator start ...")
	go v.processLoop()
}

func (v *Verification) makeRequestBatch() proto.Transactions {
	var requestBatch proto.Transactions
	var to string
	v.requestBatchTimer.Reset(v.consenter.BatchTimeout())
	v.txpool.IterElement(func(element sortedlinkedlist.IElement) bool {
		tx := element.(*proto.Transaction)
		if to == "" {
			to = tx.ToChain()
		}
		if tx.ToChain() == to && len(requestBatch) <= v.consenter.BatchSize() {
			requestBatch = append(requestBatch, tx)
		} else {
			return true
		}
		return false
	})

	return requestBatch
}

func (v *Verification) processLoop() {
	ticker := time.NewTicker(v.config.BlacklistDur)
	for {
		select {
		case <-ticker.C:
			v.rwBlacklist.Lock()
			for address, created := range v.blacklist {
				if created.Add(v.config.BlacklistDur).Before(time.Now()) {
					delete(v.blacklist, address)
				}
			}
			v.rwBlacklist.Unlock()
		case cnt := <-v.requestBatchSignal:
			if cnt >= (v.config.TxPoolDelay + v.consenter.BatchSize()) {
				requestBatch := v.makeRequestBatch()
				log.Debugf("request Batch: %d ", len(requestBatch))
				v.consenter.ProcessBatch(requestBatch, v.consensusFailed)
			}
		case <-v.requestBatchTimer.C:
			if requestBatch := v.makeRequestBatch(); len(requestBatch) != 0 {
				log.Debugf("request Batch Timeout: %d ", len(requestBatch))
				v.consenter.ProcessBatch(requestBatch, v.consensusFailed)
			}
		}
	}
}

func (v *Verification) ProcessTransaction(tx *proto.Transaction) error {
	startTime := time.Now()
	if err := v.checkTransaction(tx); err != nil {
		return err
	}

	v.rwInTxs.Lock()
	if v.isExist(tx) {
		v.rwInTxs.Unlock()
		return fmt.Errorf("transaction %s already existed", tx.Hash())
	}

	if v.isOverCapacity() {
		elem := v.txpool.RemoveFront()
		delete(v.inTxs, elem.(*proto.Transaction).Hash())
		log.Warnf("[validator]  excess capacity, remove front transaction")
	}

	v.txpool.Add(tx)
	v.inTxs[tx.Hash()] = tx
	cnt := v.txpool.Len()
	v.rwInTxs.Unlock()
	if cnt == 1 {
		v.requestBatchTimer.Reset(v.consenter.BatchTimeout())
	}
	v.requestBatchSignal <- cnt
	log.Debugf("ProcessTransaction, tx_hash: %+v time: %s", tx.Hash(), time.Now().Sub(startTime))
	log.Debugf("[txPool] add transaction success, tx_hash: %s, sender: %s, receiver: %s, assetID %d, amount: %s,txpool_len: %d",
		tx.Hash().String(), tx.Sender(), tx.Recipient(), tx.AssetID(), tx.Amount(), cnt)
	return nil
}

func (v *Verification) consensusFailed(flag int, txs proto.Transactions) {
	if len(txs) == 0 {
		return
	}
	switch flag {
	case 0: // not used, do nothing
		log.Debug("[validator] not primary replica ...")
		log.Debugf("process batch size: %d", len(txs))
		for _, tx := range txs {
			log.Debugf("process tx_hash %s", tx.Hash().String())
		}
	case 1: //used, do nothing
		log.Debug("[validator] primary replica ...")
	case 2: // add
		v.rwInTxs.Lock()
		defer v.rwInTxs.Unlock()
		for _, tx := range txs {
			v.txpool.Add(tx)
		}
	case 3: // remove
		v.rwInTxs.Lock()
		defer v.rwInTxs.Unlock()
		var elems []sortedlinkedlist.IElement
		//d:=time.Now().Sub(time.Unix(0,0))
		for _, tx := range txs {
			//v.static[tx.Hash()] =  d- v.static[tx.Hash()]
			elems = append(elems, tx)
		}
		v.txpool.Removes(elems)
	case 4: //update account
		v.rwAccount.Lock()
		defer v.rwAccount.Unlock()
		for _, tx := range txs {
			if !v.updateAccount(tx) {
				panic("balance is not enough")
			}
		}
	case 5: // rollback account
		v.rwAccount.Lock()
		defer v.rwAccount.Unlock()
		for _, tx := range txs {
			v.rollBackAccount(tx)
		}
	case 6: // remove when verify error
		v.rwInTxs.Lock()
		defer v.rwInTxs.Unlock()
		var elems []sortedlinkedlist.IElement
		for _, tx := range txs {
			delete(v.inTxs, tx.Hash())
			elems = append(elems, tx)
		}
		v.txpool.Removes(elems)
	default:
		log.Error("[validator] not support this flag ...")
	}
}

// VerifyTxs consensus verify txs
func (v *Verification) VerifyTxs(txs proto.Transactions) (ttxs proto.Transactions, etxs proto.Transactions) {
	if len(txs) == 0 || !v.config.IsValid {
		return txs, etxs
	}
	v.rwInTxs.Lock()
	v.rwAccount.Lock()
	defer v.rwInTxs.Unlock()
	defer v.rwAccount.Unlock()
	for _, tx := range txs {
		if !v.isExist(tx) {
			if err := v.checkTransaction(tx); err != nil {
				etxs = append(etxs, tx)
				log.Errorf("[validator] tx_hash: %s illegal, err %s", tx.Hash().String(), err)
				continue
			}
		}

		assetID := tx.AssetID()
		asset, ok := v.assets[assetID]
		if !ok {
			asset, _ = v.ledger.GetAsset(assetID)
		}
		if tx.GetType() != proto.TransactionType_Issue {
			if asset == nil {
				etxs = append(etxs, tx)
				log.Errorf("[validator] tx_hash: %s, asset %d not exist", tx.Hash().String(), tx.AssetID())
				continue
			}
			if tx.GetType() == proto.TransactionType_IssueUpdate && len(tx.Payload) > 0 {
				newAsset, err := asset.Update(string(tx.Payload))
				if err != nil {
					etxs = append(etxs, tx)
					log.Errorf("[validator] tx_hash: %s, update asset %d(%s) --- %s", tx.Hash().String(), assetID, string(tx.Payload), err)
					continue
				}
				v.assets[assetID] = newAsset
			}
		} else {
			if asset == nil {
				asset := &state.Asset{
					ID:     assetID,
					Issuer: tx.Sender(),
					Owner:  tx.Recipient(),
				}
				newAsset, err := asset.Update(string(tx.Payload))
				if err != nil {
					etxs = append(etxs, tx)
					log.Errorf("[validator] tx_hash: %s, new issue asset %d(%s) --- %s", tx.Hash().String(), assetID, string(tx.Payload), err)
					continue
				}
				v.assets[assetID] = newAsset
			} else {
				etxs = append(etxs, tx)
				log.Errorf("[validator] tx_hash: %s, new issue asset %d(%s) --- already exist", tx.Hash().String(), assetID, string(tx.Payload))
				continue
			}
		}

		// remove balance is negative tx
		if !v.updateAccount(tx) {
			etxs = append(etxs, tx)
			log.Errorf("[validator] tx_hash: %s, asset %d balance is not enough", tx.Hash().String(), tx.AssetID())
			continue
		}
		ttxs = append(ttxs, tx)
	}

	return ttxs, etxs
}

func (v *Verification) RemoveTxsInVerification(txs proto.Transactions) {
	v.rwInTxs.Lock()
	defer v.rwInTxs.Unlock()
	for _, tx := range txs {
		log.Debugf("[validator] remove transaction in verification ,tx_hash: %s ,txpool_len: %d", tx.Hash(), v.txpool.Len())
		delete(v.inTxs, tx.Hash())
		v.txpool.Remove(tx)
	}
}

func (v *Verification) fetchAccount(address account.Address) *balance.Balance {
	account, ok := v.accounts[address.String()]
	if !ok {
		account, _ = v.ledger.GetBalance(address)
		v.accounts[address.String()] = account
	}
	return account
}

func (v *Verification) updateAccount(tx *proto.Transaction) bool {
	assetID := tx.AssetID()
	plusAmount := tx.Amount()
	plusFee := tx.Fee()
	subAmount := -plusAmount
	subFee := -plusFee

	if fromChain := coordinate.HexToChainCoordinate(tx.FromChain()).Bytes(); bytes.Equal(fromChain, params.ChainID) {
		senderAccont := v.fetchAccount(tx.Sender())
		if senderAccont != nil {
			senderAccont.Add(assetID, subAmount)
			senderAccont.Add(assetID, subFee)
			//	log.Debugln("[validator] updateAccount sender: ", tx.Sender(), "amount: ", senderAccont.amount)
			if (tx.GetType() != proto.TransactionType_Issue && tx.GetType() != proto.TransactionType_IssueUpdate) && senderAccont.Get(assetID) < 0 {
				senderAccont.Add(assetID, plusAmount)
				senderAccont.Add(assetID, plusFee)
				return false
			}
		}
	}

	if toChain := coordinate.HexToChainCoordinate(tx.ToChain()).Bytes(); bytes.Equal(toChain, params.ChainID) {
		receiverAccount := v.fetchAccount(tx.Recipient())
		if receiverAccount != nil {
			receiverAccount.Add(assetID, plusAmount)
			receiverAccount.Add(assetID, plusFee)
			//	log.Debugln("[validator] updateAccount Recipient: ", tx.Recipient(), "amount: ", receiverAccount.amount)
			if receiverAccount.Get(assetID) < 0 {
				receiverAccount.Add(assetID, subAmount)
				receiverAccount.Add(assetID, subFee)
				return false
			}
		}
	}
	return true
}

func (v *Verification) rollBackAccount(tx *proto.Transaction) {
	assetID := tx.AssetID()
	plusAmount := tx.Amount()
	plusFee := tx.Fee()
	subAmount := -plusAmount
	subFee := -plusFee

	if fromChain := coordinate.HexToChainCoordinate(tx.FromChain()).Bytes(); bytes.Equal(fromChain, params.ChainID) {
		senderAccont := v.fetchAccount(tx.Sender())
		if senderAccont != nil {
			senderAccont.Add(assetID, plusAmount)
			senderAccont.Add(assetID, plusFee)
		}
	}

	if toChain := coordinate.HexToChainCoordinate(tx.ToChain()).Bytes(); bytes.Equal(toChain, params.ChainID) {
		receiverAccount := v.fetchAccount(tx.Recipient())
		if receiverAccount != nil {
			receiverAccount.Add(assetID, subAmount)
			receiverAccount.Add(assetID, subFee)
		}
	}
}

func (v *Verification) UpdateAccount(tx *proto.Transaction) bool {
	v.rwAccount.Lock()
	defer v.rwAccount.Unlock()
	return v.updateAccount(tx)
}

//RollBackAccount roll back account balance
func (v *Verification) RollBackAccount(tx *proto.Transaction) {
	v.rwAccount.Lock()
	defer v.rwAccount.Unlock()
	v.rollBackAccount(tx)
}

func (v *Verification) GetTransactionByHash(txHash crypto.Hash) (*proto.Transaction, bool) {
	if elem := v.txpool.GetIElementByKey(txHash.String()); elem != nil {
		return elem.(*proto.Transaction), true
	}
	return nil, false
}

func (v *Verification) GetBalance(addr account.Address) *balance.Balance {
	v.rwAccount.Lock()
	defer v.rwAccount.Unlock()
	acconut := v.fetchAccount(addr)
	return acconut
}

func (v *Verification) GetAsset(id uint32) *state.Asset {
	v.rwAccount.Lock()
	defer v.rwAccount.Unlock()
	asset, ok := v.assets[id]
	if !ok {
		asset, _ = v.ledger.GetAsset(id)
	}
	return asset
}
