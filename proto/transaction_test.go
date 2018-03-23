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

package proto

import (
	"bytes"
	"fmt"
	"testing"
	"time"
)

func TestTransaction(t *testing.T) {
	txHeader := NewTxHeader(0, uint32(time.Now().Unix()), 0, TransactionType_Atomic)
	tx := NewTransaction(txHeader, nil, nil, nil, nil)

	fmt.Println("tx hash", tx.Hash())
	fmt.Println("tx signhash", tx.SignHash())
	txData := tx.Serialize()
	testTx := &Transaction{}
	testTx.Deserialize(txData)
	if !bytes.Equal(txData, testTx.Serialize()) {
		t.Errorf("tx.Serialize error")
	}
}
