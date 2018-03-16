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
	"testing"

	"github.com/zipper-project/zipper/common/utils"
)

func TestBalance(t *testing.T) {
	b := NewBalance()
	utils.AssertEquals(t, int64(0), b.Get(0))

	b.Set(0, 100)
	utils.AssertEquals(t, int64(100), b.Get(0))

	b.Add(0, 100)
	utils.AssertEquals(t, int64(200), b.Get(0))

	b.Add(0, -1000)
	utils.AssertEquals(t, int64(-800), b.Get(0))
}

func TestSerializeAndDeserialize(t *testing.T) {
	b := NewBalance()
	b.Set(0, -100)
	b.Set(1, 200)
	b.Set(3, 300)
	balanceBytes := b.Serialize()

	tb := NewBalance()
	err := tb.Deserialize(balanceBytes)
	if err != nil {
		t.Error(err)
	}

	utils.AssertEquals(t, b.Get(0), tb.Get(0))
	utils.AssertEquals(t, b.Get(2), tb.Get(2))
	utils.AssertEquals(t, b.Get(3), tb.Get(3))
}
