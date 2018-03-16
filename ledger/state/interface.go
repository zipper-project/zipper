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
	"math/big"

	"github.com/zipper-project/zipper/common/db"
)

// IKVRWSet encapsulates the function performed during transaction simulation
type IKVRWSet interface {

	// GetChainCodeState get state for chaincode address and key. If committed is false, this first looks in memory
	// and if missing, pulls from db.  If committed is true, this pulls from the db only.
	GetChainCodeState(chaincodeAddr string, key string, committed bool) ([]byte, error)
	// GetChainCodeStateByRange get state for chaincode address and key. If committed is false, this first looks in memory
	// and if missing, pulls from db.  If committed is true, this pulls from the db only.
	GetChainCodeStateByRange(chaincodeAddr string, startKey string, endKey string, committed bool) (map[string][]byte, error)
	// SetChainCodeState set state to given value for chaincode address and key. Does not immideatly writes to DB
	SetChainCodeState(chaincodeAddr string, key string, value []byte) error
	// DelChainCodeState tracks the deletion of state for chaincode address and key. Does not immediately writes to DB
	DelChainCodeState(chaincodeAddr string, key string)

	// GetBalanceState get balance for address and assetID. If committed is false, this first looks in memory
	// and if missing, pulls from db.  If committed is true, this pulls from the db only.
	GetBalanceState(addr string, assetID uint32, committed bool) (*big.Int, error)
	// GetBalances get balances for address. If committed is false, this first looks in memory
	// and if missing, pulls from db.  If committed is true, this pulls from the db only.
	GetBalanceStates(addr string, committed bool) (map[uint32]*big.Int, error)
	// SetBalacneState set balance to given value for chaincode address and key. Does not immideatly writes to DB
	SetBalacneState(addr string, assetID uint32, amount *big.Int) error
	// DelBalanceState tracks the deletion of balance for chaincode address and key. Does not immediately writes to DB
	DelBalanceState(addr string, assetID uint32)

	// GetAssetState get asset for assetID. If committed is false, this first looks in memory
	// and if missing, pulls from db.  If committed is true, this pulls from the db only.
	GetAssetState(assetID uint32, committed bool) (*Asset, error)
	// GetAssets get assets. If committed is false, this first looks in memory
	// and if missing, pulls from db.  If committed is true, this pulls from the db only.
	GetAssetStates(committed bool) (map[uint32]*Asset, error)
	// SetAssetState set balance to given value for assetID. Does not immideatly writes to DB
	SetAssetState(assetID uint32, assetInfo *Asset) error
	// DelAssetState tracks the deletion of asset for assetID. Does not immediately writes to DB
	DelAssetState(assetID uint32)

	// ApplyChanges merges delta
	ApplyChanges() ([]*db.WriteBatch, error)
}
