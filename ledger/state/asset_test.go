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
