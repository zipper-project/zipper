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

package types

import (
	"time"

	"github.com/zipper-project/zipper/common/crypto"
)

// Transaction represents the basic transaction that contained in blocks
type Transaction struct {
}

// Hash returns the hash of a transaction
func (tx *Transaction) Hash() crypto.Hash {
	return crypto.Hash{}
}

// CreateTime returns the create time of the transaction
func (tx *Transaction) CreateTime() uint32 {
	return uint32(time.Now().Unix())
}

// Transactions represents transaction slice type for basic sorting.
type Transactions []*Transaction
