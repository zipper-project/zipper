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
package account

import "github.com/zipper-project/zipper/common/utils"

// ChainCoordinate represents the coordinate of the blockchain
type ChainCoordinate []byte

// NewChainCoordinate returns an instance of ChainCoordinate
func NewChainCoordinate(c []byte) ChainCoordinate {
	var cc = make(ChainCoordinate, len(c))
	copy(cc, c)
	return cc
}

// String returns string format of chain coordinate
func (cc ChainCoordinate) String() string {
	return utils.BytesToHex(cc)
}

// Bytes return []byte
func (cc ChainCoordinate) Bytes() []byte {
	return cc[:]
}

// MarshalText returns the hex representation of a.
func (cc ChainCoordinate) MarshalText() ([]byte, error) {
	return utils.Bytes(cc[:]).MarshalText()
}

// UnmarshalText parses a hash in hex syntax.
func (cc ChainCoordinate) UnmarshalText(input []byte) error {
	return utils.UnmarshalFixedText(input, cc[:])
}

// HexToChainCoordinate returns chain coordinate via the hex
func HexToChainCoordinate(hex string) ChainCoordinate {
	return NewChainCoordinate(utils.HexToBytes(hex))
}

// ParentCoorinate returns parent coorinate
func (cc ChainCoordinate) ParentCoorinate() ChainCoordinate {
	b := cc.Bytes()
	pb := b[:len(b)-1]
	return NewChainCoordinate(pb)
}
