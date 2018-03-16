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
package balance

import (
	"sync"

	"github.com/zipper-project/zipper/common/utils"
)

// Balance Contain all asset amounts and nonce
type Balance struct {
	Amounts map[uint32]int64
	rw      sync.RWMutex
}

// NewBalance Create an balance object
func NewBalance() *Balance {
	return &Balance{
		Amounts: make(map[uint32]int64),
	}
}

//Set Set amount of asset
func (b *Balance) Set(id uint32, amount int64) {
	b.rw.Lock()
	defer b.rw.Unlock()
	b.Amounts[id] = amount
}

//Get Get amount of asset
func (b *Balance) Get(id uint32) int64 {
	if amount, ok := b.Amounts[id]; ok {
		return amount
	}
	return 0
}

//Add Set amount of asset to  sum +y and return.
func (b *Balance) Add(id uint32, amount int64) int64 {
	b.rw.Lock()
	defer b.rw.Unlock()
	if !utils.CheckInt64Border(b.Amounts[id], amount) {
		panic("balance amount value out of range.")
	}
	b.Amounts[id] = b.Amounts[id] + amount
	return b.Amounts[id]
}

//Serialize returns the serialized bytes of a balance
func (b *Balance) Serialize() []byte {
	return utils.Serialize(b)
}

//Deserialize deserializes bytes to a balance
func (b *Balance) Deserialize(balanceBytes []byte) error {
	return utils.Deserialize(balanceBytes, b)
}
