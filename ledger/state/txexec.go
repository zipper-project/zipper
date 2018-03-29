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
	"errors"

	"github.com/zipper-project/zipper/account"
	"github.com/zipper-project/zipper/common/db"
	"github.com/zipper-project/zipper/common/log"
	"github.com/zipper-project/zipper/ledger/balance"
	pb "github.com/zipper-project/zipper/proto"
)

var permissionPrefix = "permission."

// func (tx *TXRWSet) verifyPermission(key string) error {
// 	var dataAdmin []byte
// 	var err error
// 	if key == params.AdminKey || key == params.GlobalContractKey {
// 		dataAdmin, err = tx.GetChainCodeState(params.GlobalStateKey, params.AdminKey, false)
// 		if err != nil {
// 			return err
// 		}
// 	} else {
// 		var permissionKey string
// 		if strings.Contains(key, permissionPrefix) {
// 			permissionKey = key
// 		} else {
// 			permissionKey = permissionPrefix + key
// 		}

// 		dataAdmin, err = tx.GetChainCodeState(params.GlobalStateKey, permissionKey, false)
// 		if err != nil {
// 			return err
// 		}

// 		if len(dataAdmin) == 0 {
// 			dataAdmin, err = tx.GetChainCodeState(params.GlobalStateKey, params.AdminKey, false)
// 			if err != nil {
// 				return err
// 			}
// 		}
// 	}

// 	sender := tx.currentTx.Sender().Bytes()
// 	if len(dataAdmin) > 0 {
// 		var dataAdminAddr account.Address
// 		err = json.Unmarshal(dataAdmin, &dataAdminAddr)
// 		if err != nil {
// 			return nil
// 		}

// 		if !bytes.Equal(sender, dataAdminAddr[:]) {
// 			log.Errorf("change global state, permission denied, \n%#v\n%#v\n",
// 				sender, dataAdminAddr[:])
// 			return fmt.Errorf("change global state, permission denied")
// 		}
// 	}
// 	return nil
// }

// func (tx *TXRWSet) GetGlobalState(key string) ([]byte, error) {
// 	log.Debugf("GetGlobalState key=[%s]", key)
// 	//return tx.GetChainCodeState(params.GlobalStateKey, key, false)
// 	return nil,nil
// }

// func (tx *TXRWSet) PutGlobalState(key string, value []byte) error {
// 	// if err := tx.verifyPermission(key); err != nil {
// 	// 	return err
// 	// }
// 	log.Debugf("SetGlobalState key=[%s], value=[%#v]", key, value)
// 	//return tx.SetChainCodeState(params.GlobalStateKey, key, value)
// 	return nil,nil

// }

// func (tx *TXRWSet) DelGlobalState(key string) error {
// 	// if err := tx.verifyPermission(key); err != nil {
// 	// 	return err
// 	// }
// 	log.Debugf("DelGlobalState key=[%s]", key)
// 	tx.DelChainCodeState(params.GlobalStateKey, key)
// 	return nil
// }

func (tx *TXRWSet) ComplexQuery(key string) ([]byte, error) {
	chaincodeAddr := account.NewAddress(tx.currentTx.ContractSpec.Addr).String()
	log.Debugf("ComplexQuery chaincode=[%s], key=[%s]", chaincodeAddr, key)
	return nil, errors.New("vp can't support complex qery")
}

func (tx *TXRWSet) GetState(key string) ([]byte, error) {
	chaincodeAddr := account.NewAddress(tx.currentTx.ContractSpec.Addr).String()
	log.Debugf("GetState chaincode=[%s], key=[%s]", chaincodeAddr, key)
	return tx.GetChainCodeState(chaincodeAddr, key, false)
}

func (tx *TXRWSet) PutState(key string, value []byte) error {
	chaincodeAddr := account.NewAddress(tx.currentTx.ContractSpec.Addr).String()
	log.Debugf("SetState chaincode=[%s], key=[%s], value=[%#v]", chaincodeAddr, key, value)
	return tx.SetChainCodeState(chaincodeAddr, key, value)
}

func (tx *TXRWSet) DelState(key string) error {
	chaincodeAddr := account.NewAddress(tx.currentTx.ContractSpec.Addr).String()
	log.Debugf("DelState chaincode=[%s], key=[%s]", chaincodeAddr, key)
	tx.DelChainCodeState(chaincodeAddr, key)
	return nil
}

func (tx *TXRWSet) GetByPrefix(key string) ([]*db.KeyValue, error) {
	chaincodeAddr := account.NewAddress(tx.currentTx.ContractSpec.Addr).String()
	log.Debugf("GetByPrefix chaincode=[%s], key=[%s]", chaincodeAddr, key)
	ret, err := tx.GetChainCodeStateByRange(chaincodeAddr, key, "", false)
	if err != nil {
		return nil, err
	}
	kvs := []*db.KeyValue{}
	for k, v := range ret {
		kvs = append(kvs, &db.KeyValue{
			[]byte(k),
			v,
		})
	}
	return kvs, nil
}

func (tx *TXRWSet) GetByRange(startKey, endKey string) ([]*db.KeyValue, error) {
	chaincodeAddr := account.NewAddress(tx.currentTx.ContractSpec.Addr).String()
	log.Debugf("GetByRange chaincode=[%s], startKey=[%s], endKey=[%s]", chaincodeAddr, startKey, endKey)
	ret, err := tx.GetChainCodeStateByRange(chaincodeAddr, startKey, endKey, false)
	if err != nil {
		return nil, err
	}
	kvs := []*db.KeyValue{}
	for k, v := range ret {
		kvs = append(kvs, &db.KeyValue{
			[]byte(k),
			v,
		})
	}
	return kvs, nil
}

func (tx *TXRWSet) GetBalance(addr string, assetID uint32) (int64, error) {
	log.Debugf("GetBalance addr=[%s], assetID=[%d]", addr, assetID)
	return tx.GetBalanceState(addr, assetID, false)
}

func (tx *TXRWSet) GetBalances(addr string) (*balance.Balance, error) {
	log.Debugf("GetBalances addr=[%s]", addr)
	ret, err := tx.GetBalanceStates(addr, false)
	return &balance.Balance{Amounts: ret}, err
}

func (tx *TXRWSet) GetCurrentBlockHeight() uint32 {
	log.Debugf("GetCurrentBlockHeight")
	return tx.block.BlockIndex
}

func (tx *TXRWSet) AddTransfer(fromAddr, toAddr string, assetID uint32, amount, fee int64) error {
	log.Debugf("AddTransfer from=[%s], to=[%s], assetID=[%d], amount=[%s], fee=[%s]", fromAddr, toAddr, assetID, amount, fee)
	return nil
}

func (tx *TXRWSet) Transfer(ttx *pb.Transaction) error {
	log.Debugf("TXRWSet Transfer")
	err := tx.block.view.connectTransaction(ttx, tx.block.BlockIndex)
	if err != nil {
		return err
	}

	return nil
}

type CallBackResponse struct {
	IsCanRedo bool
	Err       error
	Result    interface{}
}

func (tx *TXRWSet) CallBack(res *CallBackResponse) error {
	log.Debugf("TXRWSet CallBack txIndex: %d %v", tx.TxIndex, res)
	if res.Err != nil {
		tx.transferTxs = nil
		if res.IsCanRedo {
			tx.assetSet = NewKVRWSet()
			tx.balanceSet = NewKVRWSet()
			tx.chainCodeSet = NewKVRWSet()
			return res.Err
		}
		tx.assetSet = nil
		tx.balanceSet = nil
		tx.chainCodeSet = nil
	}
	return tx.ApplyChanges()
}
