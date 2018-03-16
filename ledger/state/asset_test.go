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
	"testing"

	"github.com/zipper-project/zipper/account"
	"github.com/zipper-project/zipper/common/utils"
)

func TestUpdate(t *testing.T) {
	asset := &Asset{
		ID:         123,
		Name:       "house",
		Descr:      "The house is for sale",
		Precision:  1,
		Expiration: 1520503284,
		Issuer:     account.HexToAddress("0xa032277be213f56221b6140998c03d860a60e1f8"),
		Owner:      account.HexToAddress("0xa132277be213f56221b6140998c03d860a60e1f8"),
	}

	result, err := asset.Update(`{"id":123,"name":"house","descr":"The house is for sale","precision":1,"expiration":1520503284,"issuer":"a032277be213f56221b6140998c03d860a60e1f8","owner":"a132277be213f56221b6140998c03d860a60e1f8"}`)
	if err != nil {
		t.Error(err)
	}

	utils.AssertEquals(t, result, asset)
}
