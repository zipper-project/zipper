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
	"bytes"
	"encoding/json"
	"fmt"

	"github.com/zipper-project/zipper/account"
	"github.com/zipper-project/zipper/common/log"
)

//Asset Attributes
type Asset struct {
	ID         uint32 `json:"id"`         // id
	Name       string `json:"name"`       // name
	Descr      string `json:"descr"`      // description
	Precision  uint64 `json:"precision"`  // divisible, precision
	Expiration uint32 `json:"expiration"` // expriation datetime

	Issuer account.Address `json:"issuer"` // issuer address
	Owner  account.Address `json:"owner"`  // owner address
}

//Update
func (asset *Asset) Update(jsonStr string) (*Asset, error) {
	if len(jsonStr) == 0 {
		return asset, nil
	}
	tAsset := &Asset{}
	if err := json.Unmarshal([]byte(jsonStr), tAsset); err != nil {
		return nil, fmt.Errorf("invalid json string for asset - %s", err)
	}

	var newVal map[string]interface{}
	json.Unmarshal([]byte(jsonStr), &newVal)

	oldJSONAStr, _ := json.Marshal(asset)
	var oldVal map[string]interface{}
	json.Unmarshal(oldJSONAStr, &oldVal)

	for k, val := range newVal {
		if _, ok := oldVal[k]; ok {
			oldVal[k] = val
		}
	}

	bts, _ := json.Marshal(oldVal)
	newAsset := &Asset{}
	json.Unmarshal(bts, newAsset)

	if asset.ID != newAsset.ID ||
		!bytes.Equal(asset.Issuer.Bytes(), newAsset.Issuer.Bytes()) ||
		!bytes.Equal(asset.Owner.Bytes(), newAsset.Owner.Bytes()) {

		log.Errorf("asset update failed, attribute mismatch, from %#v to %#v",
			asset, newAsset)
		return nil, fmt.Errorf("id, issuer, owner are readonly attribute, can't modified")
	}

	return newAsset, nil
}
