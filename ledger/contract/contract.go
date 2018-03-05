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
	"encoding/json"
	"errors"
	"fmt"
	"strings"

	"sync"

	"github.com/zipper-project/zipper/account"
	"github.com/zipper-project/zipper/common/db"
	"github.com/zipper-project/zipper/common/log"
	"github.com/zipper-project/zipper/common/utils"
	"github.com/zipper-project/zipper/config"
	"github.com/zipper-project/zipper/ledger/balance"
	pb "github.com/zipper-project/zipper/proto"
)

const (
	// ColumnFamily is the column family of contract state in db.
	ColumnFamily = "scontract"

	// permissionPrefix is the prefix of data permission key.
	permissionPrefix = "permission."
)

type ILedgerSmartContract interface {
	GetTmpBalance(addr account.Address) (*balance.Balance, error)
	Height() (uint32, error)
}

// SmartConstract represents the account state
type SmartConstract struct {
	dbHandler     *db.BlockchainDB
	balancePrefix []byte
	columnFamily  string
	ledgerHandler ILedgerSmartContract
	stateExtra    *StateExtra

	height           uint32
	scAddr           string
	committed        bool
	currentTx        *pb.Transaction
	smartContractTxs pb.Transactions
	sync.Mutex
}

//var sctx *SmartConstract

// NewSmartConstract returns a new State
func NewSmartConstract(db *db.BlockchainDB, ledgerHandler ILedgerSmartContract) *SmartConstract {
	sctx := &SmartConstract{
		dbHandler:     db,
		balancePrefix: []byte("sc_"),
		columnFamily:  ColumnFamily,
		ledgerHandler: ledgerHandler,
		stateExtra:    NewStateExtra(),
	}
	return sctx
}

//GetColumnFamily return smart constract columnFamily
func (sctx *SmartConstract) GetColumnFamily() string { return sctx.columnFamily }

// StartConstract start constract
func (sctx *SmartConstract) StartConstract(blockHeight uint32) {
	log.Debugf("startConstract() for blockHeight [%d]", blockHeight)
	if !sctx.InProgress() {
		log.Errorf("A tx [%d] is already in progress. Received call for begin of another smartcontract [%d]", sctx.height, blockHeight)
	}
	sctx.height = blockHeight
}

// StopContract start contract
func (sctx *SmartConstract) StopContract(blockHeight uint32) {
	log.Debugf("stopConstract() for blockHeight [%d]", blockHeight)
	if sctx.height != blockHeight {
		log.Errorf("Different blockHeight in contract-begin [%d] and contract-finish [%d]", sctx.height, blockHeight)
	}

	sctx.height = 0
	sctx.stateExtra = NewStateExtra()
}

// ExecTransaction exec transaction
func (sctx *SmartConstract) ExecTransaction(tx *pb.Transaction, scAddr string) {
	sctx.committed = false
	sctx.currentTx = tx
	sctx.scAddr = scAddr
	sctx.smartContractTxs = make(pb.Transactions, 0)
}

// GetGlobalState returns the global state.
func (sctx *SmartConstract) GetGlobalState(key string) ([]byte, error) {
	if !sctx.InProgress() {
		log.Errorf("State can be changed only in context of a block.")
	}

	log.Debugf("GetGlobalState key=[%s]", key)
	return sctx.GetStateInOneAddr(config.GlobalStateKey, key)
}

func (sctx *SmartConstract) verifyPermission(key string) error {
	var dataAdmin []byte
	var err error
	if key == config.AdminKey || key == config.GlobalContractKey {
		dataAdmin, err = sctx.GetContractStateData(config.GlobalStateKey, config.AdminKey)
		if err != nil {
			return err
		}
	} else {
		var permissionKey string
		if strings.Contains(key, permissionPrefix) {
			permissionKey = key
		} else {
			permissionKey = permissionPrefix + key
		}

		dataAdmin, err = sctx.GetContractStateData(config.GlobalStateKey, permissionKey)
		if err != nil {
			return err
		}

		if len(dataAdmin) == 0 {
			dataAdmin, err = sctx.GetContractStateData(config.GlobalStateKey, config.AdminKey)
			if err != nil {
				return err
			}
		}
	}

	sender := sctx.currentTx.Sender().Bytes()
	if len(dataAdmin) > 0 {
		var dataAdminAddr account.Address
		err = json.Unmarshal(dataAdmin, &dataAdminAddr)
		if err != nil {
			return nil
		}

		if !bytes.Equal(sender, dataAdminAddr[:]) {
			log.Errorf("change global state, permission denied, \n%#v\n%#v\n",
				sender, dataAdminAddr[:])
			return fmt.Errorf("change global state, permission denied")
		}
	}

	return nil
}

// SetGlobalState sets the global state.
func (sctx *SmartConstract) SetGlobalState(key string, value []byte) error {
	if !sctx.InProgress() {
		log.Errorf("State can be changed only in context of a block.")
	}

	err := sctx.verifyPermission(key)
	if err != nil {
		return err
	}

	log.Debugf("SetGlobalState key=[%s], value=[%#v]", key, value)
	sctx.stateExtra.set(config.GlobalStateKey, key, value)
	return nil
}

// DelGlobalState deletes the global state.
func (sctx *SmartConstract) DelGlobalState(key string) error {
	if !sctx.InProgress() {
		log.Errorf("State can be changed only in context of a block.")
	}

	err := sctx.verifyPermission(key)
	if err != nil {
		return err
	}

	log.Debugf("DelGlobalState key=[%s]", key)
	sctx.stateExtra.delete(config.GlobalStateKey, key)
	return nil
}

// GetState get value
func (sctx *SmartConstract) GetState(key string) ([]byte, error) {
	if !sctx.InProgress() {
		log.Errorf("State can be changed only in context of a block.")
	}

	return sctx.GetStateInOneAddr(sctx.scAddr, key)
}

func (sctx *SmartConstract) GetStateInOneAddr(scAddr, key string) ([]byte, error) {
	value := sctx.stateExtra.get(scAddr, key)
	if len(value) == 0 {
		var err error
		scAddrkey := EnSmartContractKey(scAddr, key)
		log.Debugf("sctx.scAddr: %s,%s,%s", scAddr, key, scAddrkey)
		value, err = sctx.dbHandler.Get(sctx.GetColumnFamily(), []byte(scAddrkey))
		if err != nil {
			return nil, fmt.Errorf("can't get data from db err: %v", err)
		}
	}

	return value, nil
}

// AddState put key-value into cache
func (sctx *SmartConstract) AddState(key string, value []byte) {
	log.Debugf("PutState smartcontract=[%x], key=[%s], value=[%#v]", sctx.scAddr, key, value)
	if !sctx.InProgress() {
		log.Errorf("State can be changed only in context of a block.")
	}

	sctx.stateExtra.set(sctx.scAddr, key, value)
}

// DelState remove key-value
func (sctx *SmartConstract) DelState(key string) {
	if !sctx.InProgress() {
		log.Errorf("State can be changed only in context of a block.")
	}

	sctx.stateExtra.delete(sctx.scAddr, key)
}

func (sctx *SmartConstract) GetByPrefix(prefix string) []*db.KeyValue {
	if !sctx.InProgress() {
		log.Errorf("State can be changed only in context of a block.")
	}
	scAddrkey := EnSmartContractKey(sctx.scAddr, prefix)
	cacheValues := sctx.stateExtra.getByPrefix(sctx.scAddr, prefix)
	dbValues := sctx.dbHandler.GetByPrefix(sctx.GetColumnFamily(), []byte(scAddrkey))

	return sctx.getKeyValues(cacheValues, dbValues)
}

func (sctx *SmartConstract) GetByRange(startKey, limitKey string) []*db.KeyValue {
	if !sctx.InProgress() {
		log.Errorf("State can be changed only in context of a block.")
	}

	scAddrStartKey := EnSmartContractKey(sctx.scAddr, startKey)
	scAddrlimitKey := EnSmartContractKey(sctx.scAddr, limitKey)
	cacheValues := sctx.stateExtra.getByRange(sctx.scAddr, startKey, limitKey)
	dbValues := sctx.dbHandler.GetByRange(sctx.GetColumnFamily(), []byte(scAddrStartKey), []byte(scAddrlimitKey))

	return sctx.getKeyValues(cacheValues, dbValues)
}

func (sctx *SmartConstract) getKeyValues(cacheKVs []*db.KeyValue, dbKVS []*db.KeyValue) []*db.KeyValue {
	var keyValues []*db.KeyValue

	cacheValuesMap := make(map[string]*db.KeyValue)
	for _, v := range cacheKVs {
		_, key := DeSmartContractKey(string(v.Key))
		cacheValuesMap[key] = v
		v.Key = []byte(key)
	}

	for _, v := range dbKVS {
		_, key := DeSmartContractKey(string(v.Key))
		if _, ok := cacheValuesMap[key]; !ok {
			v.Key = []byte(key)
			keyValues = append(keyValues, v)
		}
	}
	return append(keyValues, cacheKVs...)
}

// GetBalances get balance
func (sctx *SmartConstract) GetBalances(addr string) (*balance.Balance, error) {
	return sctx.ledgerHandler.GetTmpBalance(account.HexToAddress(addr))
}

// CurrentBlockHeight get currentBlockHeight
func (sctx *SmartConstract) CurrentBlockHeight() uint32 {
	height, err := sctx.ledgerHandler.Height()
	if err == nil {
		log.Errorf("can't read blockchain height")
	}

	return height
}

// SmartContractFailed execute smartContract fail
func (sctx *SmartConstract) SmartContractFailed() {
	sctx.committed = false
	log.Errorf("VM can't put state into L0")
}

// SmartContractCommitted execute smartContract successfully
func (sctx *SmartConstract) SmartContractCommitted() {
	sctx.committed = true
}

// AddTransfer add transfer to make new transaction
func (sctx *SmartConstract) AddTransfer(fromAddr, toAddr string, assetID uint32, amount int64, txType uint32) {
	tx := &pb.Transaction{
		TxData: pb.TxData{
			Header: &pb.TxHeader{
				FromChain:  sctx.currentTx.GetHeader().GetFromChain(),
				ToChain:    sctx.currentTx.GetHeader().GetToChain(),
				Type:       pb.TransactionType(txType),
				Nonce:      sctx.currentTx.GetHeader().GetNonce(),
				Sender:     fromAddr,
				Recipient:  toAddr,
				AssetID:    assetID,
				Amount:     amount,
				Fee:        sctx.currentTx.GetHeader().GetFee(),
				CreateTime: sctx.currentTx.GetHeader().GetCreateTime(),
			},
		},
	}

	sctx.smartContractTxs = append(sctx.smartContractTxs, tx)
}

// InProgress
func (sctx *SmartConstract) InProgress() bool {
	return true
}

// FinishContractTransaction finish contract transaction
func (sctx *SmartConstract) FinishContractTransaction() (pb.Transactions, error) {
	if !sctx.committed {
		return nil, errors.New("Execute VM Fail")
	}

	return sctx.smartContractTxs, nil
}

// AddChangesForPersistence put cache data into db
func (sctx *SmartConstract) AddChangesForPersistence(writeBatch []*db.WriteBatch) ([]*db.WriteBatch, error) {
	updateContractStateDelta := sctx.stateExtra.getUpdatedContractStateDelta()
	for _, smartContract := range updateContractStateDelta {
		smartContract.cache.ForEach(func(key, value []byte) bool {
			cv := &CacheKVs{}
			cv.deserialize(value)
			if cv.Optype == db.OperationDelete {
				log.Debugln("Contract Del: ", string(key))
				writeBatch = append(writeBatch, db.NewWriteBatch(sctx.GetColumnFamily(), db.OperationDelete, key, cv.Value, sctx.GetColumnFamily()))
			} else if cv.Optype == db.OperationPut {
				log.Debugln("Contract Put: ", string(key), string(cv.Value))
				writeBatch = append(writeBatch, db.NewWriteBatch(sctx.GetColumnFamily(), db.OperationPut, key, cv.Value, sctx.GetColumnFamily()))
			} else {
				log.Errorf("invalid method ...")
			}
			return true
		})
	}

	return writeBatch, nil
}

func (sctx *SmartConstract) GetContractStateData(scAddr string, key string) ([]byte, error) {
	srcValue, err := sctx.GetStateInOneAddr(scAddr, key)
	if err != nil {
		return nil, err
	}

	value, err := DoContractStateData(srcValue)
	if err != nil {
		log.Errorf("can't handle state data, err: %+v, value: %+v, src: %+v", err, srcValue, string(srcValue))
		return nil, err
	}

	return value, nil
}

func (sctx *SmartConstract) ExecuteSmartContractTx(tx *pb.Transaction) (pb.Transactions, error) {
	sctx.Lock()
	defer sctx.Unlock()

	sctx.ExecTransaction(tx, utils.BytesToHex(tx.ContractSpec.GetAddr()))
	// ok, err := vm.RealExecute(tx, contractSpec, sctx)
	// if err != nil {
	// 	return nil, fmt.Errorf("contract execute failed : %v ", err)
	// }

	// if !ok {
	// 	return nil, fmt.Errorf("contract execute failed : error in contract")
	// }

	smartContractTxs, err := sctx.FinishContractTransaction()
	if err != nil {
		log.Error("FinishContractTransaction: ", err)
		return nil, err
	}

	return smartContractTxs, nil
}

func (sctx *SmartConstract) ExecuteRequireContract(tx *pb.Transaction, scAddr string) (bool, error) {
	sctx.Lock()
	defer sctx.Unlock()
	tx.ContractSpec.Addr = utils.HexToBytes(scAddr)
	sctx.ExecTransaction(tx, utils.BytesToHex(tx.ContractSpec.GetAddr()))

	var ok bool
	// ok, err := vm.RealExecuteRequire(tx, contractSpec, sctx)
	// if err != nil {
	// 	return ok, fmt.Errorf("contract execute failed : %v ", err)
	// }

	return ok, nil
}

func (sctx *SmartConstract) QueryContract(tx *pb.Transaction) ([]byte, error) {
	sctx.Lock()
	defer sctx.Unlock()
	sctx.ExecTransaction(tx, utils.BytesToHex(tx.ContractSpec.GetAddr()))
	var result []byte
	// result, err := vm.Query(tx, contractSpec, sctx)
	// if err != nil {
	// 	log.Errorf("contract query execute failed: %v ", err)
	// 	return nil, fmt.Errorf("contract query execute failed : %v ", err)
	// }
	return result, nil
}
