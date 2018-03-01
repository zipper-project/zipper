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
package account

import (
	"fmt"

	"github.com/zipper-project/zipper/common/crypto"
	"github.com/zipper-project/zipper/common/utils"
)

const (
	addressPrefix = "0x"
	AddressLength = 20
)

// Address definition
type Address [AddressLength]byte

// NewAddress new a var of Address type
func NewAddress(b []byte) Address {
	var a Address
	if len(b) > AddressLength {
		copy(a[:], b[:20])
	} else {
		copy(a[:], b[:])
	}
	return a
}

// String returns address string
func (self Address) String() string {
	return fmt.Sprintf("%s%x", addressPrefix, self[:])
}

//Bytes returns address bytes
func (self Address) Bytes() []byte {
	return self[:]
}

// HexToAddress creates address from address string
func HexToAddress(hex string) Address {
	var a Address
	var b []byte
	if len(hex) > 1 {
		if hex[0:2] == addressPrefix {
			hex = hex[2:]
		}
		b = utils.HexToBytes(hex)
	}
	a.SetBytes(b)
	return a
}

// SetBytes set the Address var's bytes
func (a *Address) SetBytes(b []byte) {
	if len(b) > len(a) {
		b = b[len(b)-AddressLength:]
	}
	copy(a[AddressLength-len(b):], b)
}

// MarshalText returns the hex representation of a.
func (a Address) MarshalText() ([]byte, error) {
	return utils.Bytes(a[:]).MarshalText()
}

// UnmarshalText parses a hash in hex syntax.
func (a *Address) UnmarshalText(input []byte) error {
	return utils.UnmarshalFixedText(input, a[:])
}

// Equal reports whether a1 equal a
func (a Address) Equal(a1 Address) bool {
	for i := 0; i < AddressLength; i++ {
		if a[i] != a1[i] {
			return false
		}
	}
	return true
}

// PublicKeyToAddress generate address from the public key
func PublicKeyToAddress(p crypto.PublicKey) Address {
	pubBytes := p.Bytes()
	var a Address
	a.SetBytes(crypto.Keccak256(pubBytes[1:])[12:])
	return a
}

// ChainCoordinateToAddress return the publicaccount address of the specified chain by chaincoordinate
func ChainCoordinateToAddress(cc ChainCoordinate) Address {
	b := cc.Bytes()
	chainIndex := b[len(b)-1:]
	var a Address
	addrBytes := append(chainIndex, crypto.Keccak256(b)[13:]...)
	a.SetBytes(addrBytes)
	return a
}
