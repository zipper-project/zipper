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
package db

import (
	"bytes"
	"fmt"
	"sync"

	"github.com/tecbot/gorocksdb"
	"github.com/zipper-project/zipper/common/log"
	"github.com/zipper-project/zipper/common/utils"
)

const (
	// OperationPut represents put operation
	OperationPut uint = iota
	// OperationDelete represents delete operation
	OperationDelete
)

var (
	deafultColumnfamilies = []string{"asset", "balance", "ledger", "peer", "index", "state", "block", "transaction", "storage", "scontract", "persistCacheTxs"}
	config                *Config
	dbInstance            *BlockchainDB
	once                  sync.Once

	rocksDBLogLevelMap = map[string]gorocksdb.InfoLogLevel{
		"debug": gorocksdb.DebugInfoLogLevel,
		"info":  gorocksdb.InfoInfoLogLevel,
		"warn":  gorocksdb.WarnInfoLogLevel,
		"error": gorocksdb.ErrorInfoLogLevel,
		"fatal": gorocksdb.FatalInfoLogLevel,
	}
)

//KeyValue key value
type KeyValue struct {
	Key   []byte
	Value []byte
}

// CfHandlerMap is columnfamilies handler set
type CfHandlerMap map[string]*gorocksdb.ColumnFamilyHandle

// BlockchainDB encapsulates rocksdb's structures
type BlockchainDB struct {
	DB         *gorocksdb.DB
	cfHandlers CfHandlerMap
}

// Config is the configuration of the gorocksdb
type Config struct {
	DbPath            string
	Columnfamilies    []string
	KeepLogFileNumber int
	MaxLogFileSize    int
	LogLevel          string
}

// WriteBatch wrappers batch operation
type WriteBatch struct {
	CfName    string
	Operation uint
	Key       []byte
	Value     []byte
	Typ       string
}

// DefaultConfig defines the default configuration of the rocksdb
func DefaultConfig() *Config {
	return &Config{
		DbPath:            "/tmp/rocksdb-test/",
		Columnfamilies:    deafultColumnfamilies,
		KeepLogFileNumber: 10,
		MaxLogFileSize:    10485760,
		LogLevel:          "warn",
	}
}

//GetDBInstance returns db instance
func GetDBInstance() *BlockchainDB {
	if dbInstance == nil {
		NewDB(DefaultConfig())
	}
	return dbInstance
}

// NewDB returns a basic db instance
func NewDB(c *Config) *BlockchainDB {
	once.Do(func() {
		config = c

		dbInstance = &BlockchainDB{}
		dbInstance.open()
	})
	return dbInstance
}

func (blockchainDB *BlockchainDB) open() {
	opts := gorocksdb.NewDefaultOptions()
	defer opts.Destroy()

	if config.MaxLogFileSize > 0 {
		opts.SetMaxLogFileSize(config.MaxLogFileSize)
	}

	if config.KeepLogFileNumber > 0 {
		opts.SetKeepLogFileNum(config.KeepLogFileNumber)
	}

	LogLevel, ok := rocksDBLogLevelMap[config.LogLevel]

	if ok {
		opts.SetInfoLogLevel(LogLevel)
	}

	// if the dir not exists, create a new db
	opts.SetCreateIfMissing(true)
	opts.SetCreateIfMissingColumnFamilies(true)

	var cfOpts []*gorocksdb.Options

	Columnfamilies := append(config.Columnfamilies, "default")
	for range Columnfamilies {
		cfOpts = append(cfOpts, opts)
	}

	db, cfHandlers, err := gorocksdb.OpenDbColumnFamilies(opts, config.DbPath, Columnfamilies, cfOpts)
	if err != nil {
		panic(fmt.Sprintf("failed to open db, error: [%s]", err))
	}

	blockchainDB.DB = db
	blockchainDB.cfHandlers = make(map[string]*gorocksdb.ColumnFamilyHandle)
	for index, cfName := range Columnfamilies {
		blockchainDB.cfHandlers[cfName] = cfHandlers[index]
	}
}

// Close releases all column family handles and closes rocksdb
func (blockchainDB *BlockchainDB) Close() {
	for cfName := range blockchainDB.cfHandlers {
		blockchainDB.cfHandlers[cfName].Destroy()
	}
	blockchainDB.DB.Close()
}

// Get returns the value for the given column family and key
func (blockchainDB *BlockchainDB) Get(cfName string, key []byte) ([]byte, error) {
	blockchainDB.checkIfColumnExists(cfName)

	opt := gorocksdb.NewDefaultReadOptions()
	defer opt.Destroy()

	slice, err := blockchainDB.DB.GetCF(opt, blockchainDB.cfHandlers[cfName], key)
	if err != nil {
		return nil, err
	}
	defer slice.Free()
	if slice.Data() == nil {
		return nil, nil
	}
	data := utils.MinimizeSilce(slice.Data())
	return data, nil
}

// GetByPrefix for bulk reads
func (blockchainDB *BlockchainDB) GetByPrefix(cfName string, prefix []byte) []*KeyValue {
	blockchainDB.checkIfColumnExists(cfName)

	var keyValues []*KeyValue

	ro := gorocksdb.NewDefaultReadOptions()
	ro.SetFillCache(true)
	defer ro.Destroy()
	it := blockchainDB.DB.NewIteratorCF(ro, blockchainDB.cfHandlers[cfName])
	defer it.Close()

	for it.Seek(prefix); it.ValidForPrefix(prefix); it.Next() {
		keyValues = append(keyValues, &KeyValue{Key: utils.MinimizeSilce(it.Key().Data()), Value: utils.MinimizeSilce(it.Value().Data())})
		it.Key().Free()
		it.Value().Free()
	}
	return keyValues
}

//GetByRange Range query
func (blockchainDB *BlockchainDB) GetByRange(cfName string, startkey []byte, limitKey []byte) []*KeyValue {
	blockchainDB.checkIfColumnExists(cfName)

	var keyValues []*KeyValue
	ro := gorocksdb.NewDefaultReadOptions()
	ro.SetFillCache(true)
	defer ro.Destroy()
	it := blockchainDB.DB.NewIteratorCF(ro, blockchainDB.cfHandlers[cfName])
	defer it.Close()

	for it.Seek(startkey); it.Valid(); it.Next() {
		keyValues = append(keyValues, &KeyValue{Key: utils.MinimizeSilce(it.Key().Data()), Value: utils.MinimizeSilce(it.Value().Data())})
		if bytes.Equal(it.Key().Data(), limitKey) {
			it.Key().Free()
			it.Value().Free()
			break
		}
		it.Key().Free()
		it.Value().Free()
	}
	return keyValues
}

// Put saves the key/value in the given column family
func (blockchainDB *BlockchainDB) Put(cfName string, key []byte, value []byte) error {
	blockchainDB.checkIfColumnExists(cfName)

	opt := gorocksdb.NewDefaultWriteOptions()
	defer opt.Destroy()
	err := blockchainDB.DB.PutCF(opt, blockchainDB.cfHandlers[cfName], key, value)
	if err != nil {
		return err
	}
	return nil
}

// Delete deletes the given key in the specified column family
func (blockchainDB *BlockchainDB) Delete(cfName string, key []byte) error {
	blockchainDB.checkIfColumnExists(cfName)

	opt := gorocksdb.NewDefaultWriteOptions()
	defer opt.Destroy()
	err := blockchainDB.DB.DeleteCF(opt, blockchainDB.cfHandlers[cfName], key)
	if err != nil {
		return err
	}
	return nil
}

// AtomicWrite writes batch
func (blockchainDB *BlockchainDB) AtomicWrite(writeBatchs []*WriteBatch) error {
	wb := gorocksdb.NewWriteBatch()
	defer wb.Destroy()
	for _, writeBatch := range writeBatchs {
		switch writeBatch.Operation {
		case OperationPut:
			wb.PutCF(blockchainDB.cfHandlers[writeBatch.CfName], writeBatch.Key, writeBatch.Value)
		case OperationDelete:
			wb.DeleteCF(blockchainDB.cfHandlers[writeBatch.CfName], writeBatch.Key)
		}
	}

	wo := gorocksdb.NewDefaultWriteOptions()
	defer wo.Destroy()

	return blockchainDB.DB.Write(wo, wb)
}

func (blockchainDB *BlockchainDB) checkIfColumnExists(cfName string) {
	if _, ok := blockchainDB.cfHandlers[cfName]; !ok {
		log.Errorf("column family does not exist %s", cfName)
		panic("column family does not exist")
	}
}

// NewWriteBatch returns a writebatch instance
func NewWriteBatch(cfName string, operation uint, key, value []byte, typ string) *WriteBatch {
	return &WriteBatch{
		CfName:    cfName,
		Operation: operation,
		Key:       key,
		Value:     value,
		Typ:       typ,
	}
}
