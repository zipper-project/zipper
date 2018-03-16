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
package state

import (
	"bytes"
	"errors"
	"fmt"
	"strconv"
	"strings"
	"sync"

	"github.com/zipper-project/zipper/common/crypto"
	"github.com/zipper-project/zipper/common/db"
	"github.com/zipper-project/zipper/common/log"
	"github.com/zipper-project/zipper/common/treap"
	"github.com/zipper-project/zipper/common/utils"
	"github.com/zipper-project/zipper/ledger/balance"
	pb "github.com/zipper-project/zipper/proto"
)

// NewBLKRWSet create object
func NewBLKRWSet(db *db.BlockchainDB) *BLKRWSet {
	return &BLKRWSet{
		chainCodeSet: NewKVRWSet(),
		balanceSet:   NewKVRWSet(),
		assetSet:     NewKVRWSet(),
		chainCodeCF:  "scontract",
		balanceCF:    "balance",
		assetCF:      "asset",
		dbHandler:    db,
		exit:         make(chan struct{}, 1),
	}
}

var assetIDKeyPrefix = "asset"
var assetIDKeySuffix = "$"

//BLKRWSet encapsulates the read-write set during transactions of block simulation
type BLKRWSet struct {
	chainCodeSet *KVRWSet
	chainCodeRW  sync.RWMutex
	balanceSet   *KVRWSet
	balanceRW    sync.RWMutex
	assetSet     *KVRWSet
	assetRW      sync.RWMutex

	dbHandler   *db.BlockchainDB
	chainCodeCF string
	balanceCF   string
	assetCF     string

	txs         pb.Transactions
	transferTxs pb.Transactions
	errTxs      pb.Transactions

	BlockIndex uint32
	TxIndex    uint32
	curTxIndex uint32

	waiting   bool
	waitingRW sync.RWMutex
	exit      chan struct{}
}

// GetChainCodeState get state for chaincode address and key. If committed is false, this first looks in memory
// and if missing, pulls from db.  If committed is true, this pulls from the db only.
func (blk *BLKRWSet) GetChainCodeState(chaincodeAddr string, key string, committed bool) ([]byte, error) {
	blk.chainCodeRW.RLock()
	defer blk.chainCodeRW.RUnlock()
	ckey := ConstructCompositeKey(chaincodeAddr, key)
	if !committed {
		if kvw, ok := blk.chainCodeSet.Writes[ckey]; ok {
			return kvw.Value, nil
		}

		if kvr, ok := blk.chainCodeSet.Reads[ckey]; ok {
			return kvr.Value, nil
		}
	}
	return blk.dbHandler.Get(blk.chainCodeCF, []byte(ckey))
}

// GetChainCodeStateByRange get state for chaincode address and key. If committed is false, this first looks in memory
// and if missing, pulls from db.  If committed is true, this pulls from the db only.
func (blk *BLKRWSet) GetChainCodeStateByRange(chaincodeAddr string, startKey string, endKey string, committed bool) (map[string][]byte, error) {
	blk.chainCodeRW.RLock()
	defer blk.chainCodeRW.RUnlock()
	chaincodePrefix := ConstructCompositeKey(chaincodeAddr, "")
	ckeyStart := ConstructCompositeKey(chaincodeAddr, startKey)
	ckeyEnd := ConstructCompositeKey(chaincodeAddr, endKey)
	ret := make(map[string][]byte)
	if len(endKey) > 0 {
		dbValues := blk.dbHandler.GetByRange(blk.chainCodeCF, []byte(ckeyStart), []byte(ckeyEnd))
		for _, kv := range dbValues {
			_, key := DecodeCompositeKey(string(kv.Key))
			ret[key] = kv.Value
		}
	} else {
		dbValues := blk.dbHandler.GetByPrefix(blk.chainCodeCF, []byte(ckeyStart))
		for _, kv := range dbValues {
			_, key := DecodeCompositeKey(string(kv.Key))
			ret[key] = kv.Value
		}
	}

	if !committed {
		cache := treap.NewImmutable()
		for ckey, kvr := range blk.chainCodeSet.Reads {
			if strings.HasPrefix(ckey, chaincodePrefix) {
				cache = cache.Put([]byte(ckey), kvr.Value)
			}
		}
		for ckey, kvw := range blk.chainCodeSet.Writes {
			if strings.HasPrefix(ckey, chaincodePrefix) {
				cache = cache.Put([]byte(ckey), kvw.Value)
			}
		}
		if len(endKey) > 0 {
			for iter := cache.Iterator([]byte(ckeyStart), []byte(ckeyEnd)); iter.Next(); {
				if val := iter.Value(); val != nil {
					_, key := DecodeCompositeKey(string(iter.Key()))
					ret[key] = val
				}
			}
		} else {
			for iter := cache.Iterator([]byte(ckeyStart), nil); iter.Next(); {
				if !bytes.HasPrefix(iter.Key(), []byte(ckeyStart)) {
					break
				}
				if val := iter.Value(); val != nil {
					_, key := DecodeCompositeKey(string(iter.Key()))
					ret[key] = val
				}
			}
		}
	}
	return ret, nil
}

// SetChainCodeState set state to given value for chaincode address and key. Does not immideatly writes to DB
func (blk *BLKRWSet) SetChainCodeState(chaincodeAddr string, key string, value []byte) error {
	blk.chainCodeRW.Lock()
	defer blk.chainCodeRW.Unlock()
	ckey := ConstructCompositeKey(chaincodeAddr, key)
	blk.chainCodeSet.Writes[ckey] = &KVWrite{
		Value:    value,
		IsDelete: false,
	}
	return nil
}

// DelChainCodeState tracks the deletion of state for chaincode address and key. Does not immediately writes to DB
func (blk *BLKRWSet) DelChainCodeState(chaincodeAddr string, key string) {
	blk.chainCodeRW.Lock()
	defer blk.chainCodeRW.Unlock()
	ckey := ConstructCompositeKey(chaincodeAddr, key)
	blk.chainCodeSet.Writes[ckey] = &KVWrite{
		Value:    nil,
		IsDelete: true,
	}
}

// GetBalanceState get balance for address and assetID. If committed is false, this first looks in memory
// and if missing, pulls from db.  If committed is true, this pulls from the db only.
func (blk *BLKRWSet) GetBalanceState(addr string, assetID uint32, committed bool) (int64, error) {
	blk.balanceRW.RLock()
	defer blk.balanceRW.RUnlock()
	var amount int64
	ckey := ConstructCompositeKey(addr, strconv.FormatUint(uint64(assetID), 10)+assetIDKeySuffix)
	if !committed {
		if kvw, ok := blk.balanceSet.Writes[ckey]; ok {
			if kvw.IsDelete {
				return 0, nil
			}
			utils.Deserialize(kvw.Value, &amount)
			return amount, nil
		}

		if kvr, ok := blk.balanceSet.Reads[ckey]; ok {
			utils.Deserialize(kvr.Value, &amount)
			return amount, nil
		}
	}
	value, err := blk.dbHandler.Get(blk.balanceCF, []byte(ckey))
	if err != nil {
		return 0, err
	}
	utils.Deserialize(value, &amount)
	return amount, nil
}

// GetBalanceStates get balances for address. If committed is false, this first looks in memory
// and if missing, pulls from db.  If committed is true, this pulls from the db only.
func (blk *BLKRWSet) GetBalanceStates(addr string, committed bool) (map[uint32]int64, error) {
	blk.balanceRW.RLock()
	defer blk.balanceRW.RUnlock()
	prefix := ConstructCompositeKey(addr, "")
	ret := make(map[string][]byte)
	dbValues := blk.dbHandler.GetByPrefix(blk.balanceCF, []byte(prefix))
	for _, kv := range dbValues {
		ret[string(kv.Key)] = kv.Value
	}
	if !committed {
		cache := treap.NewImmutable()
		for ckey, kvr := range blk.balanceSet.Reads {
			if strings.HasPrefix(ckey, prefix) {
				cache = cache.Put([]byte(ckey), kvr.Value)
			}
		}
		for ckey, kvw := range blk.balanceSet.Writes {
			if strings.HasPrefix(ckey, prefix) {
				cache = cache.Put([]byte(ckey), kvw.Value)
			}
		}

		for iter := cache.Iterator([]byte(prefix), nil); iter.Next(); {
			if !bytes.HasPrefix(iter.Key(), []byte(prefix)) {
				break
			}
			if val := iter.Value(); val != nil {
				ret[string(iter.Key())] = val
			}
		}
	}

	balances := make(map[uint32]int64)
	for k, v := range ret {
		if v != nil {
			_, key := DecodeCompositeKey(k)
			assetID, err := strconv.ParseUint(strings.TrimSuffix(key, assetIDKeySuffix), 10, 32)
			if err != nil {
				return nil, err
			}
			var amount int64
			if err := utils.Deserialize(v, &amount); err != nil {
				return nil, err
			}
			balances[uint32(assetID)] = amount
		}
	}
	return balances, nil
}

// SetBalacneState set balance to given value for chaincode address and key. Does not immideatly writes to DB
func (blk *BLKRWSet) SetBalacneState(addr string, assetID uint32, amount int64) error {
	blk.balanceRW.Lock()
	defer blk.balanceRW.Unlock()
	value := utils.Serialize(amount)
	ckey := ConstructCompositeKey(addr, strconv.FormatUint(uint64(assetID), 10)+assetIDKeySuffix)
	blk.balanceSet.Writes[ckey] = &KVWrite{
		Value:    value,
		IsDelete: false,
	}
	return nil
}

// DelBalanceState tracks the deletion of balance for chaincode address and key. Does not immediately writes to DB
func (blk *BLKRWSet) DelBalanceState(addr string, assetID uint32) {
	blk.balanceRW.Lock()
	defer blk.balanceRW.Unlock()
	ckey := ConstructCompositeKey(addr, strconv.FormatUint(uint64(assetID), 10)+assetIDKeySuffix)
	blk.balanceSet.Writes[ckey] = &KVWrite{
		Value:    nil,
		IsDelete: true,
	}
}

// GetAssetState get asset for assetID. If committed is false, this first looks in memory
// and if missing, pulls from db.  If committed is true, this pulls from the db only.
func (blk *BLKRWSet) GetAssetState(assetID uint32, committed bool) (*Asset, error) {
	blk.assetRW.RLock()
	defer blk.assetRW.RUnlock()
	assetInfo := &Asset{}
	ckey := ConstructCompositeKey(assetIDKeyPrefix, strconv.FormatUint(uint64(assetID), 10)+assetIDKeySuffix)
	if !committed {
		if kvw, ok := blk.assetSet.Writes[ckey]; ok {
			if kvw.IsDelete {
				return nil, nil
			}
			if err := utils.Deserialize(kvw.Value, assetInfo); err != nil {
				return nil, err
			}
			return assetInfo, nil
		}

		if kvr, ok := blk.assetSet.Reads[ckey]; ok {
			if err := utils.Deserialize(kvr.Value, assetInfo); err != nil {
				return nil, err
			}
			return assetInfo, nil
		}
	}
	value, err := blk.dbHandler.Get(blk.assetCF, []byte(ckey))
	if err != nil {
		return nil, err
	}
	if value == nil {
		return nil, nil
	}
	if err := utils.Deserialize(value, assetInfo); err != nil {
		return nil, err
	}
	return assetInfo, nil
}

// GetAssetStates get assets. If committed is false, this first looks in memory
// and if missing, pulls from db.  If committed is true, this pulls from the db only.
func (blk *BLKRWSet) GetAssetStates(committed bool) (map[uint32]*Asset, error) {
	blk.assetRW.RLock()
	defer blk.assetRW.RUnlock()
	prefix := ConstructCompositeKey(assetIDKeyPrefix, "")
	ret := make(map[string][]byte)
	dbValues := blk.dbHandler.GetByPrefix(blk.assetCF, []byte(prefix))
	for _, kv := range dbValues {
		ret[string(kv.Key)] = kv.Value
	}
	if !committed {
		cache := treap.NewImmutable()
		for ckey, kvr := range blk.assetSet.Reads {
			if strings.HasPrefix(ckey, prefix) {
				cache = cache.Put([]byte(ckey), kvr.Value)
			}
		}
		for ckey, kvw := range blk.assetSet.Writes {
			if strings.HasPrefix(ckey, prefix) {
				cache = cache.Put([]byte(ckey), kvw.Value)
			}
		}

		for iter := cache.Iterator([]byte(prefix), nil); iter.Next(); {
			if !bytes.HasPrefix(iter.Key(), []byte(prefix)) {
				break
			}
			if val := iter.Value(); val != nil {
				ret[string(iter.Key())] = val
			}
		}
	}

	assets := make(map[uint32]*Asset)
	for _, v := range ret {
		if v != nil {
			assetInfo := &Asset{}
			if err := utils.Deserialize(v, assetInfo); err != nil {
				return nil, err
			}
			assets[assetInfo.ID] = assetInfo
		}
	}
	return assets, nil
}

// SetAssetState set balance to given value for assetID. Does not immideatly writes to DB
func (blk *BLKRWSet) SetAssetState(assetID uint32, assetInfo *Asset) error {
	blk.assetRW.Lock()
	defer blk.assetRW.Unlock()
	value := utils.Serialize(assetInfo)
	ckey := ConstructCompositeKey(assetIDKeyPrefix, strconv.FormatUint(uint64(assetID), 10)+assetIDKeySuffix)
	blk.assetSet.Writes[ckey] = &KVWrite{
		Value:    value,
		IsDelete: false,
	}
	return nil
}

// DelAssetState tracks the deletion of asset for assetID. Does not immediately writes to DB
func (blk *BLKRWSet) DelAssetState(assetID uint32) {
	blk.assetRW.Lock()
	defer blk.assetRW.Unlock()
	ckey := ConstructCompositeKey(assetIDKeyPrefix, strconv.FormatUint(uint64(assetID), 10)+assetIDKeySuffix)
	blk.assetSet.Writes[ckey] = &KVWrite{
		Value:    nil,
		IsDelete: true,
	}
}

// ApplyChanges merges delta
func (blk *BLKRWSet) ApplyChanges() ([]*db.WriteBatch, pb.Transactions, pb.Transactions, error) {
	blk.wait()
	log.Debugf("BLKRWSet ApplyChanges blockHeight:%d, txNum:%d", blk.BlockIndex, blk.TxIndex)
	blk.chainCodeRW.RLock()
	defer blk.chainCodeRW.RUnlock()
	blk.assetRW.RLock()
	defer blk.assetRW.RUnlock()
	blk.balanceRW.RLock()
	defer blk.balanceRW.RUnlock()

	writeBatchs := make([]*db.WriteBatch, 0)
	for ckey, wset := range blk.chainCodeSet.Writes {
		if wset.IsDelete {
			writeBatchs = append(writeBatchs, db.NewWriteBatch(blk.chainCodeCF, db.OperationDelete, []byte(ckey), nil, blk.chainCodeCF))
		} else {
			writeBatchs = append(writeBatchs, db.NewWriteBatch(blk.chainCodeCF, db.OperationPut, []byte(ckey), wset.Value, blk.chainCodeCF))
		}
	}

	for ckey, wset := range blk.assetSet.Writes {
		if wset.IsDelete {
			writeBatchs = append(writeBatchs, db.NewWriteBatch(blk.assetCF, db.OperationDelete, []byte(ckey), nil, blk.assetCF))
		} else {
			writeBatchs = append(writeBatchs, db.NewWriteBatch(blk.assetCF, db.OperationPut, []byte(ckey), wset.Value, blk.assetCF))
		}
	}

	for ckey, wset := range blk.balanceSet.Writes {
		if wset.IsDelete {
			writeBatchs = append(writeBatchs, db.NewWriteBatch(blk.balanceCF, db.OperationDelete, []byte(ckey), nil, blk.balanceCF))
		} else {
			writeBatchs = append(writeBatchs, db.NewWriteBatch(blk.balanceCF, db.OperationPut, []byte(ckey), wset.Value, blk.balanceCF))
		}
	}

	errTxs := blk.errTxs
	txs := blk.txs
	txs = append(txs, blk.transferTxs...)
	return writeBatchs, txs, errTxs, nil
}

func (blk *BLKRWSet) merge(chainCodeSet *KVRWSet, assetSet *KVRWSet, balanceSet *KVRWSet, tx *pb.Transaction, ttxs pb.Transactions, txIndex uint32) error {
	blk.chainCodeRW.Lock()
	defer blk.chainCodeRW.Unlock()
	blk.assetRW.Lock()
	defer blk.assetRW.Unlock()
	blk.balanceRW.Lock()
	defer blk.balanceRW.Unlock()

	if chainCodeSet == nil && assetSet == nil && balanceSet == nil && ttxs == nil {
		blk.errTxs = append(blk.errTxs, tx)
	} else {
		for ckey, rset := range chainCodeSet.Reads {
			if trset, ok := blk.chainCodeSet.Reads[ckey]; ok {
				if bytes.Compare(trset.Value, rset.Value) != 0 {
					chaincodeAddr, key := DecodeCompositeKey(ckey)
					return fmt.Errorf("chaincode readset conflict -- %s %s", chaincodeAddr, key)
				}
			}
		}

		for ckey, rset := range assetSet.Reads {
			if trset, ok := blk.assetSet.Reads[ckey]; ok {
				if bytes.Compare(trset.Value, rset.Value) != 0 {
					_, key := DecodeCompositeKey(ckey)
					return fmt.Errorf("asset readset conflict -- %s", key)
				}
			}
		}

		for ckey, rset := range balanceSet.Reads {
			if trset, ok := blk.balanceSet.Reads[ckey]; ok {
				if bytes.Compare(trset.Value, rset.Value) != 0 {
					addr, key := DecodeCompositeKey(ckey)
					return fmt.Errorf("balance readset conflict -- %s %s", addr, key)
				}
			}
		}

		for ckey, wset := range chainCodeSet.Writes {
			blk.chainCodeSet.Writes[ckey] = wset
		}

		for ckey, wset := range assetSet.Writes {
			blk.assetSet.Writes[ckey] = wset
		}

		for ckey, wset := range balanceSet.Writes {
			blk.balanceSet.Writes[ckey] = wset
		}
		blk.transferTxs = append(blk.transferTxs, ttxs...)

		blk.txs = append(blk.txs, tx)
	}

	blk.waitingRW.Lock()
	blk.curTxIndex = txIndex + 1
	log.Debugf("BLKRWSet merge lock blockHeight:%d, txIndex:%d", blk.BlockIndex, blk.curTxIndex)
	if blk.waiting && blk.TxIndex == blk.curTxIndex {
		blk.exit <- struct{}{}
	}
	blk.waitingRW.Unlock()
	log.Debugf("BLKRWSet merge blockHeight:%d, txIndex:%d", blk.BlockIndex, blk.curTxIndex)
	return nil
}

func (blk *BLKRWSet) GetChainCodeCF() string {
	return blk.chainCodeCF
}

func (blk *BLKRWSet) GetAssetCF() string {
	return blk.assetCF
}

func (blk *BLKRWSet) GetBalanceCF() string {
	return blk.balanceCF
}

func (blk *BLKRWSet) ComplexQuery(key string) ([]byte, error) {
	return nil, errors.New("vp can't support complex qery")
}

func (blk *BLKRWSet) GetBalances(addr string) (*balance.Balance, error) {
	ret, err := blk.GetBalanceStates(addr, false)
	return &balance.Balance{Amounts: ret}, err
}

func (blk *BLKRWSet) GetAsset(assetID uint32) (*Asset, error) {
	ret, err := blk.GetAssetState(assetID, false)
	return ret, err
}

func (blk *BLKRWSet) GetAssets() (map[uint32]*Asset, error) {
	ret, err := blk.GetAssetStates(false)
	return ret, err
}

func (blk *BLKRWSet) wait() {
	blk.waitingRW.Lock()
	if blk.TxIndex != blk.curTxIndex {
		blk.waiting = true
	} else {
		blk.waiting = false
	}
	blk.waitingRW.Unlock()
	if blk.waiting {
		<-blk.exit
	}
}

func (blk *BLKRWSet) SetBlock(blkIndex, txNum uint32) {
	log.Debugf("BLKRWSet SetBlock blockHeight:%d, txNum:%d", blkIndex, txNum)
	blk.BlockIndex = blkIndex
	blk.TxIndex = txNum
	blk.curTxIndex = 0
	blk.exit = make(chan struct{}, 1)
	blk.assetSet = NewKVRWSet()
	blk.balanceSet = NewKVRWSet()
	blk.chainCodeSet = NewKVRWSet()
	blk.txs = nil
	blk.errTxs = nil
	blk.transferTxs = nil
}

func (blk *BLKRWSet) RootHash() crypto.Hash {
	hashs := make([]crypto.Hash, 3)
	hashs[0] = crypto.DoubleSha256(utils.Serialize(blk.chainCodeSet))
	hashs[1] = crypto.DoubleSha256(utils.Serialize(blk.assetSet))
	hashs[2] = crypto.DoubleSha256(utils.Serialize(blk.balanceSet))
	return crypto.ComputeMerkleHash(hashs)[0]
}
