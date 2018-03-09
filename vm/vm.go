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

package vm

import (
	ltyes "github.com/zipper-project/zipper/ledger/balance"
	"github.com/zipper-project/zipper/ledger/state"
	"github.com/zipper-project/zipper/proto"
)

//blockchain should provide the implement to VM
type ISmartConstract interface {
	GetGlobalState(key string) ([]byte, error)

	PutGlobalState(key string, value []byte) error

	DelGlobalState(key string) error

	GetState(key string) ([]byte, error)

	PutState(key string, value []byte) error

	DelState(key string) error

	ComplexQuery(key string) ([]byte, error)

	GetBalance(addr string, assetID uint32) (int64, error)

	GetCurrentBlockHeight() uint32

	AddTransfer(fromAddr, toAddr string, assetID uint32, amount, fee int64) error

	Transfer(tx *proto.Transaction) error

	GetBalances(addr string) (*ltyes.Balance, error)

	//GetByPrefix(prefix string) ([]*db.KeyValue, error)

	//GetByRange(startKey, limitKey string) ([]*db.KeyValue, error)
	CallBack(response *state.CallBackResponse) error
}

type BVMEngine interface {
}
