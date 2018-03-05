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
package contract

import (
	"bytes"

	"github.com/zipper-project/zipper/common/db"
	"github.com/zipper-project/zipper/common/log"
	"github.com/zipper-project/zipper/common/treap"
	"github.com/zipper-project/zipper/common/utils"
)

type StateExtra struct {
	ContractStateDeltas map[string]*ContractStateDelta
	success             bool
}

func NewStateExtra() *StateExtra {
	return &StateExtra{make(map[string]*ContractStateDelta), false}
}

func (stateExtra *StateExtra) get(scAddr string, key string) []byte {
	contractStateDelta, ok := stateExtra.ContractStateDeltas[scAddr]
	if ok {
		return contractStateDelta.get(EnSmartContractKey(scAddr, key))
	}
	return nil
}

func (stateExtra *StateExtra) set(scAddr string, key string, value []byte) {
	contractStateDelta := stateExtra.getOrCreateContractStateDelta(scAddr)
	contractStateDelta.set(EnSmartContractKey(scAddr, key), value)
	return
}

func (stateExtra *StateExtra) delete(scAddr string, key string) {
	contractStateDelta := stateExtra.getOrCreateContractStateDelta(scAddr)
	contractStateDelta.remove(EnSmartContractKey(scAddr, key))
	return
}

func (stateExtra *StateExtra) getByPrefix(scAddr string, prefix string) []*db.KeyValue {
	contractStateDelta := stateExtra.getOrCreateContractStateDelta(scAddr)
	return contractStateDelta.getByPrefix(EnSmartContractKey(scAddr, prefix))
}

func (stateExtra *StateExtra) getByRange(scAddr, startKey, limitKey string) []*db.KeyValue {
	contractStateDelta := stateExtra.getOrCreateContractStateDelta(scAddr)
	return contractStateDelta.getByRange(EnSmartContractKey(scAddr, startKey), EnSmartContractKey(scAddr, limitKey))

}

func (stateExtra *StateExtra) getOrCreateContractStateDelta(scAddr string) *ContractStateDelta {
	contractStateDelta, ok := stateExtra.ContractStateDeltas[scAddr]
	if !ok {
		contractStateDelta = newContractStateDelta(scAddr)
		stateExtra.ContractStateDeltas[scAddr] = contractStateDelta
	}
	return contractStateDelta
}

func (stateExtra *StateExtra) getUpdatedContractStateDelta() map[string]*ContractStateDelta {
	return stateExtra.ContractStateDeltas
}

type ContractStateDelta struct {
	contract string
	cache    *treap.Immutable
}

func newContractStateDelta(scAddr string) *ContractStateDelta {
	return &ContractStateDelta{scAddr, treap.NewImmutable()}
}

func (csd *ContractStateDelta) get(key string) []byte {
	value := csd.cache.Get([]byte(key))

	if value != nil {
		cv := &CacheKVs{}
		cv.deserialize(value)
		if cv.Optype != db.OperationDelete {
			return cv.Value
		}
	}
	return nil
}

func (csd *ContractStateDelta) set(key string, value []byte) {
	cv := newCacheKVs(db.OperationPut, key, value)
	csd.cache = csd.cache.Put([]byte(key), cv.serialize())
}

func (csd *ContractStateDelta) remove(key string) {
	cv := newCacheKVs(db.OperationDelete, key, nil)
	csd.cache = csd.cache.Put([]byte(key), cv.serialize())
}

func (csd *ContractStateDelta) getByPrefix(prefix string) []*db.KeyValue {
	var values []*db.KeyValue

	for iter := csd.cache.Iterator([]byte(prefix), nil); iter.Next(); {
		if !bytes.HasPrefix(iter.Key(), []byte(prefix)) {
			break
		}
		cv := &CacheKVs{}
		cv.deserialize(iter.Value())
		if cv.Optype != db.OperationDelete {
			values = append(values, &db.KeyValue{Key: []byte(cv.Key), Value: cv.Value})
		}
	}
	return values
}

func (csd *ContractStateDelta) getByRange(startKey string, limitkey string) []*db.KeyValue {
	var values []*db.KeyValue

	for iter := csd.cache.Iterator([]byte(startKey), nil); iter.Next(); {
		// if bytes.HasPrefix(iter.Key(), []byte(startKey)) && len(iter.Key()) != len([]byte(startKey)) {
		// 	continue
		// }
		cv := &CacheKVs{}
		cv.deserialize(iter.Value())
		if cv.Optype != db.OperationDelete {
			values = append(values, &db.KeyValue{Key: []byte(cv.Key), Value: cv.Value})
			if bytes.Equal(iter.Key(), []byte(limitkey)) {
				break
			}
		}

	}
	return values

}

type CacheKVs struct {
	Optype uint
	Key    string
	Value  []byte
}

func newCacheKVs(typ uint, key string, value []byte) *CacheKVs {
	return &CacheKVs{Optype: typ, Key: key, Value: value}
}

func (c *CacheKVs) serialize() []byte {
	return utils.Serialize(c)
}

func (c *CacheKVs) deserialize(cacheKVsBytes []byte) {
	if err := utils.Deserialize(cacheKVsBytes, c); err != nil {
		log.Errorln("CacheKVs deserialize error", err)
	}
}
