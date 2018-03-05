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
package state

import (
	"sync"

	"github.com/zipper-project/zipper/account"
	"github.com/zipper-project/zipper/common/db"
	"github.com/zipper-project/zipper/common/utils"
	"github.com/zipper-project/zipper/ledger/balance"
)

// State represents the account state
type State struct {
	dbHandler     *db.BlockchainDB
	balancePrefix []byte
	balanceCF     string
	tmpBalance    map[string]*balance.Balance
	assetPrefix   []byte
	assetCF       string
	tmpAsset      map[uint32]*Asset
	sync.RWMutex
}

// NewState returns a new State
func NewState(db *db.BlockchainDB) *State {
	return &State{
		dbHandler:     db,
		balancePrefix: []byte("bl_"),
		balanceCF:     "balance",
		tmpBalance:    make(map[string]*balance.Balance),
		assetPrefix:   []byte("ast_"),
		assetCF:       "asset",
		tmpAsset:      make(map[uint32]*Asset),
	}
}

//WriteBatch atomic writeBatchs
func (state *State) WriteBatchs() []*db.WriteBatch {
	state.Lock()
	defer state.Unlock()
	writeBatchs := make([]*db.WriteBatch, 0)
	for a, b := range state.tmpBalance {
		key := append(state.balancePrefix, []byte(a)...)
		writeBatchs = append(writeBatchs, db.NewWriteBatch(state.balanceCF, db.OperationPut, key, b.Serialize(), state.balanceCF))
	}
	state.tmpBalance = make(map[string]*balance.Balance)
	for a, b := range state.tmpAsset {
		id := utils.Serialize(a)
		key := append(state.assetPrefix, id...)
		writeBatchs = append(writeBatchs, db.NewWriteBatch(state.assetCF, db.OperationPut, key, utils.Serialize(b), state.assetCF))
	}
	state.tmpAsset = make(map[uint32]*Asset)
	return writeBatchs
}

// GetBalance returns balance by account
func (state *State) GetBalance(a account.Address) (*balance.Balance, error) {
	key := append(state.balancePrefix, []byte(a.String())...)
	balanceBytes, err := state.dbHandler.Get(state.balanceCF, key)
	if err != nil {
		return nil, err
	}
	balance := balance.NewBalance()
	if len(balanceBytes) != 0 {
		balance.Deserialize(balanceBytes)
	}
	return balance, nil
}

//GetTmpBalance get tmpBalance When the block is not packaged
func (state *State) GetTmpBalance(addr account.Address) (*balance.Balance, error) {
	state.Lock()
	defer state.Unlock()
	balance, ok := state.tmpBalance[addr.String()]
	if !ok {
		balance, err := state.GetBalance(addr)
		if err != nil {
			return nil, err
		}
		state.tmpBalance[addr.String()] = balance
		return balance, nil
	}
	return balance, nil
}

// UpdateBalance updates the account balance
func (state *State) UpdateBalance(a account.Address, id uint32, amount int64) error {
	tmpBalance, err := state.GetTmpBalance(a)
	if err != nil {
		return err
	}

	tmpBalance.Add(id, amount)

	return nil
}

// GetAsset returns asset by id
func (state *State) GetAsset(id uint32) (*Asset, error) {
	key := append(state.assetPrefix, utils.Serialize(id)...)
	bytes, err := state.dbHandler.Get(state.assetCF, key)
	if err != nil {
		return nil, err
	}
	if len(bytes) != 0 {
		asset := &Asset{}
		utils.Deserialize(bytes, asset)
		return asset, nil
	}
	return nil, nil
}

//GetTmpAsset get tmpAsset When the block is not packaged
func (state *State) GetTmpAsset(id uint32) (*Asset, error) {
	state.Lock()
	defer state.Unlock()
	asset, ok := state.tmpAsset[id]
	if !ok {
		asset, err := state.GetAsset(id)
		if err != nil {
			return nil, err
		}
		if asset != nil {
			state.tmpAsset[id] = asset
		}
		return asset, nil
	}
	return asset, nil
}

// UpdateAsset updates asset
func (state *State) UpdateAsset(id uint32, issur, owner account.Address, jsonStr string) error {
	state.Lock()
	defer state.Unlock()
	asset, ok := state.tmpAsset[id]
	if !ok {
		asset = &Asset{
			ID:     id,
			Issuer: issur,
			Owner:  owner,
		}
	}
	newAsset, err := asset.Update(jsonStr)
	if err != nil {
		return err
	}
	state.tmpAsset[id] = newAsset
	return nil
}

func (state *State) GetAssetCF() string {
	return state.assetCF
}

func (state *State) GetBalanceCF() string {
	return state.balanceCF
}
