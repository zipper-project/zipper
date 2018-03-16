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
	"strconv"
	"strings"
	"sync"

	"github.com/zipper-project/zipper/common/log"
	"github.com/zipper-project/zipper/common/treap"
	"github.com/zipper-project/zipper/common/utils"
	pb "github.com/zipper-project/zipper/proto"
)

// NewTXRWSet create object
func NewTXRWSet(blk *BLKRWSet, tx *pb.Transaction, txIndex uint32) *TXRWSet {
	return &TXRWSet{
		chainCodeSet: NewKVRWSet(),
		balanceSet:   NewKVRWSet(),
		assetSet:     NewKVRWSet(),
		block:        blk,
		currentTx:    tx,
		TxIndex:      txIndex,
	}
}

// TXRWSet encapsulates the read-write set during transaction simulation
type TXRWSet struct {
	chainCodeSet *KVRWSet
	chainCodeRW  sync.RWMutex
	balanceSet   *KVRWSet
	balanceRW    sync.RWMutex
	assetSet     *KVRWSet
	assetRW      sync.RWMutex

	block       *BLKRWSet
	currentTx   *pb.Transaction
	transferTxs pb.Transactions
	TxIndex     uint32
}

// GetChainCodeState get state for chaincode address and key. If committed is false, this first looks in memory
// and if missing, pulls from db.  If committed is true, this pulls from the db only.
func (tx *TXRWSet) GetChainCodeState(chaincodeAddr string, key string, committed bool) ([]byte, error) {
	tx.chainCodeRW.RLock()
	defer tx.chainCodeRW.RUnlock()
	ckey := ConstructCompositeKey(chaincodeAddr, key)
	if !committed {
		if kvw, ok := tx.chainCodeSet.Writes[ckey]; ok {
			return kvw.Value, nil
		}

		if kvr, ok := tx.chainCodeSet.Reads[ckey]; ok {
			return kvr.Value, nil
		}
	}
	val, err := tx.block.GetChainCodeState(chaincodeAddr, key, committed)
	if val != nil {
		tx.chainCodeSet.Reads[ckey] = &KVRead{
			Value: val,
		}
	}
	return val, err
}

// GetChainCodeStateByRange get state for chaincode address and key. If committed is false, this first looks in memory
// and if missing, pulls from db.  If committed is true, this pulls from the db only.
func (tx *TXRWSet) GetChainCodeStateByRange(chaincodeAddr string, startKey string, endKey string, committed bool) (map[string][]byte, error) {
	tx.chainCodeRW.RLock()
	defer tx.chainCodeRW.RUnlock()
	chaincodePrefix := ConstructCompositeKey(chaincodeAddr, "")
	ckeyStart := ConstructCompositeKey(chaincodeAddr, startKey)
	ckeyEnd := ConstructCompositeKey(chaincodeAddr, endKey)
	ret, err := tx.block.GetChainCodeStateByRange(chaincodeAddr, startKey, endKey, committed)
	if err != nil {
		return nil, err
	}
	if !committed {
		cache := treap.NewImmutable()
		for ckey, kvr := range tx.chainCodeSet.Reads {
			if strings.HasPrefix(ckey, chaincodePrefix) {
				cache.Put([]byte(ckey), kvr.Value)
			}
		}
		for ckey, kvw := range tx.chainCodeSet.Writes {
			if strings.HasPrefix(ckey, chaincodePrefix) {
				cache.Put([]byte(ckey), kvw.Value)
			}
		}

		if len(endKey) > 0 {
			for iter := cache.Iterator([]byte(ckeyStart), []byte(ckeyEnd)); iter.Next(); {
				if val := iter.Value(); val != nil {
					ret[string(iter.Key())] = val
				}
			}
		} else {
			for iter := cache.Iterator([]byte(startKey), nil); iter.Next(); {
				if !bytes.HasPrefix(iter.Key(), []byte(startKey)) {
					break
				}
				if val := iter.Value(); val != nil {
					ret[string(iter.Key())] = val
				}
			}
		}
	}
	return ret, nil
}

// SetChainCodeState set state to given value for chaincode address and key. Does not immideatly writes to DB
func (tx *TXRWSet) SetChainCodeState(chaincodeAddr string, key string, value []byte) error {
	tx.chainCodeRW.Lock()
	defer tx.chainCodeRW.Unlock()
	ckey := ConstructCompositeKey(chaincodeAddr, key)
	tx.chainCodeSet.Writes[ckey] = &KVWrite{
		Value:    value,
		IsDelete: false,
	}
	return nil
}

// DelChainCodeState tracks the deletion of state for chaincode address and key. Does not immediately writes to DB
func (tx *TXRWSet) DelChainCodeState(chaincodeAddr string, key string) {
	tx.chainCodeRW.Lock()
	defer tx.chainCodeRW.Unlock()
	ckey := ConstructCompositeKey(chaincodeAddr, key)
	tx.chainCodeSet.Writes[ckey] = &KVWrite{
		Value:    nil,
		IsDelete: true,
	}
}

// GetBalanceState get balance for address and assetID. If committed is false, this first looks in memory
// and if missing, pulls from db.  If committed is true, this pulls from the db only.
func (tx *TXRWSet) GetBalanceState(addr string, assetID uint32, committed bool) (int64, error) {
	tx.balanceRW.RLock()
	defer tx.balanceRW.RUnlock()
	var amount int64
	ckey := ConstructCompositeKey(addr, strconv.FormatUint(uint64(assetID), 10)+assetIDKeySuffix)
	if !committed {
		if kvw, ok := tx.balanceSet.Writes[ckey]; ok {
			if kvw.IsDelete {
				return 0, nil
			}
			utils.Deserialize(kvw.Value, amount)
			return amount, nil
		}

		if kvr, ok := tx.balanceSet.Reads[ckey]; ok {
			utils.Deserialize(kvr.Value, amount)

			return amount, nil
		}
	}
	val, err := tx.block.GetBalanceState(addr, assetID, committed)
	if err != nil {
		return 0, err
	}
	tx.chainCodeSet.Reads[ckey] = &KVRead{
		Value: utils.Serialize(val),
	}
	return val, nil
}

// GetBalanceStates get balances for address. If committed is false, this first looks in memory
// and if missing, pulls from db.  If committed is true, this pulls from the db only.
func (tx *TXRWSet) GetBalanceStates(addr string, committed bool) (map[uint32]int64, error) {
	tx.balanceRW.RLock()
	defer tx.balanceRW.RUnlock()
	balances, err := tx.block.GetBalanceStates(addr, committed)
	if err != nil {
		return nil, err
	}

	prefix := ConstructCompositeKey(addr, "")
	ret := make(map[string][]byte)
	if !committed {
		cache := treap.NewImmutable()
		for ckey, kvr := range tx.balanceSet.Reads {
			if strings.HasPrefix(ckey, prefix) {
				cache.Put([]byte(ckey), kvr.Value)
			}
		}
		for ckey, kvw := range tx.balanceSet.Writes {
			if strings.HasPrefix(ckey, prefix) {
				cache.Put([]byte(ckey), kvw.Value)
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

	for k, v := range ret {
		if v != nil {
			assetID, err := strconv.ParseUint(k, 10, 32)
			if err != nil {
				return nil, err
			}
			utils.Deserialize(v, balances[uint32(assetID)])
		}
	}
	return balances, nil
}

// SetBalacneState set balance to given value for chaincode address and key. Does not immideatly writes to DB
func (tx *TXRWSet) SetBalacneState(addr string, assetID uint32, amount int64) error {
	tx.balanceRW.Lock()
	defer tx.balanceRW.Unlock()
	ckey := ConstructCompositeKey(addr, strconv.FormatUint(uint64(assetID), 10)+assetIDKeySuffix)
	tx.balanceSet.Writes[ckey] = &KVWrite{
		Value:    utils.Serialize(amount),
		IsDelete: false,
	}
	return nil
}

// DelBalanceState tracks the deletion of balance for chaincode address and key. Does not immediately writes to DB
func (tx *TXRWSet) DelBalanceState(addr string, assetID uint32) {
	tx.balanceRW.Lock()
	defer tx.balanceRW.Unlock()
	ckey := ConstructCompositeKey(addr, strconv.FormatUint(uint64(assetID), 10)+assetIDKeySuffix)
	tx.balanceSet.Writes[ckey] = &KVWrite{
		Value:    nil,
		IsDelete: true,
	}
}

// GetAssetState get asset for assetID. If committed is false, this first looks in memory
// and if missing, pulls from db.  If committed is true, this pulls from the db only.
func (tx *TXRWSet) GetAssetState(assetID uint32, committed bool) (*Asset, error) {
	tx.assetRW.RLock()
	defer tx.assetRW.RUnlock()
	assetInfo := &Asset{}
	ckey := ConstructCompositeKey(assetIDKeyPrefix, strconv.FormatUint(uint64(assetID), 10)+assetIDKeySuffix)
	if !committed {
		if kvw, ok := tx.assetSet.Writes[ckey]; ok {
			if kvw.IsDelete {
				return nil, nil
			}
			if err := utils.Deserialize(kvw.Value, assetInfo); err != nil {
				return nil, err
			}
			return assetInfo, nil
		}

		if kvr, ok := tx.assetSet.Reads[ckey]; ok {
			if err := utils.Deserialize(kvr.Value, assetInfo); err != nil {
				return nil, err
			}
			return assetInfo, nil
		}
	}
	val, err := tx.block.GetAssetState(assetID, committed)
	if val != nil {
		tx.chainCodeSet.Reads[ckey] = &KVRead{
			Value: utils.Serialize(val),
		}
	}
	return val, err
}

// GetAssetStates get assets. If committed is false, this first looks in memory
// and if missing, pulls from db.  If committed is true, this pulls from the db only.
func (tx *TXRWSet) GetAssetStates(committed bool) (map[uint32]*Asset, error) {
	tx.assetRW.RLock()
	defer tx.assetRW.RUnlock()

	assets, err := tx.block.GetAssetStates(committed)
	if err != nil {
		return nil, err
	}
	prefix := ConstructCompositeKey(assetIDKeyPrefix, "")
	ret := make(map[string][]byte)
	if !committed {
		cache := treap.NewImmutable()
		for ckey, kvr := range tx.assetSet.Reads {
			if strings.HasPrefix(ckey, prefix) {
				cache.Put([]byte(ckey), kvr.Value)
			}
		}
		for ckey, kvw := range tx.assetSet.Writes {
			if strings.HasPrefix(ckey, prefix) {
				cache.Put([]byte(ckey), kvw.Value)
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

	assetInfo := &Asset{}
	for _, v := range ret {
		if v != nil {
			if err := utils.Deserialize(v, assetInfo); err != nil {
				return nil, err
			}
			assets[assetInfo.ID] = assetInfo
		}
	}
	return assets, nil
}

// SetAssetState set balance to given value for assetID. Does not immideatly writes to DB
func (tx *TXRWSet) SetAssetState(assetID uint32, assetInfo *Asset) error {
	tx.assetRW.Lock()
	defer tx.assetRW.Unlock()
	value := utils.Serialize(assetInfo)
	ckey := ConstructCompositeKey(assetIDKeyPrefix, strconv.FormatUint(uint64(assetID), 10)+assetIDKeySuffix)
	tx.assetSet.Writes[ckey] = &KVWrite{
		Value:    value,
		IsDelete: false,
	}
	return nil
}

// DelAssetState tracks the deletion of asset for assetID. Does not immediately writes to DB
func (tx *TXRWSet) DelAssetState(assetID uint32) {
	tx.assetRW.Lock()
	defer tx.assetRW.Unlock()
	ckey := ConstructCompositeKey(assetIDKeyPrefix, strconv.FormatUint(uint64(assetID), 10)+assetIDKeySuffix)
	tx.assetSet.Writes[ckey] = &KVWrite{
		Value:    nil,
		IsDelete: true,
	}
}

// ApplyChanges merges delta
func (tx *TXRWSet) ApplyChanges() error {
	tx.chainCodeRW.RLock()
	defer tx.chainCodeRW.RUnlock()
	tx.assetRW.RLock()
	defer tx.assetRW.RUnlock()
	tx.balanceRW.RLock()
	defer tx.balanceRW.RUnlock()
	log.Debugf("TXRWSet ApplyChanges txIndex: %d ", tx.TxIndex)
	err := tx.block.merge(tx.chainCodeSet, tx.assetSet, tx.balanceSet, tx.currentTx, tx.transferTxs, tx.TxIndex)

	// tx.assetSet = NewKVRWSet()
	// tx.balanceSet = NewKVRWSet()
	// tx.chainCodeSet = NewKVRWSet()
	// tx.transferTxs = nil
	return err
}
