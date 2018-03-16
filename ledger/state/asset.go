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
	"bytes"
	"encoding/json"
	"fmt"

	"github.com/zipper-project/zipper/account"
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

//Update update asset
func (asset *Asset) Update(jsonStr string) (*Asset, error) {
	if len(jsonStr) == 0 || !json.Valid([]byte(jsonStr)) {
		return asset, nil
	}

	tAsset := &Asset{}
	if err := json.Unmarshal([]byte(jsonStr), tAsset); err != nil {
		return nil, fmt.Errorf("asset update failed: invalid json string for asset - %s", err)
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
		return nil, fmt.Errorf("asset update failed: id, issuer, owner are readonly attribute, can't modified -- %#v to %#v", asset, newAsset)
	}
	return newAsset, nil
}
