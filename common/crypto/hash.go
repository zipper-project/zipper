// Copyright (C) 2017, Zipper Team Technology Co.,Ltd.  All rights reserved.
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
package crypto

import (
	"crypto/sha256"
	"encoding/hex"

	"github.com/zipper-project/zipper/common/utils"
	"golang.org/x/crypto/ripemd160"
)

const (
	// HashSize represents the hash length
	HashSize = 32
)

type (
	// Hash represents the 32 byte hash of arbitrary data
	Hash [HashSize]byte
)

// String returns the string respresentation of the hash
func (h Hash) String() string { return hex.EncodeToString(h[:]) }

// Bytes returns the bytes respresentation of the hash
func (h Hash) Bytes() []byte { return h[:] }

// Xor calculates the h ^ h1. returns resulting hash
func (h Hash) Xor(h1 Hash) Hash {
	d := Hash{}
	for i := 0; i < HashSize; i++ {
		d[i] = h[i] ^ h1[i]
	}
	return d
}

// MarshalText returns the hex representation of h.
func (h Hash) MarshalText() ([]byte, error) {
	return utils.Bytes(h[:]).MarshalText()
}

// UnmarshalText parses a hash in hex syntax.
func (h *Hash) UnmarshalText(input []byte) error {
	return utils.UnmarshalFixedText(input, h[:])
}

// SetBytes sets the hash to the value of b
func (h *Hash) SetBytes(b []byte) {
	if len(*h) == len(b) {
		copy(h[:], b[:HashSize])
	}
}

// SetString set string `s` to h
func (h *Hash) SetString(s string) { h.SetBytes([]byte(s)) }

// Reverse sets self byte-reversed hash
func (h *Hash) Reverse() Hash {
	for i, b := range h[:HashSize/2] {
		h[i], h[HashSize-1-i] = h[HashSize-1-i], b
	}
	return *h
}

// PrefixLen returns the length of the zero prefix
func (h Hash) PrefixLen() int {
	for i := 0; i < HashSize; i++ {
		for j := 0; j < 8; j++ {
			if (h[i]>>uint(7-j))&0x1 != 0 {
				return i*8 + j
			}
		}
	}
	return HashSize*8 - 1
}

// Equal reports whether h1 equal h
func (h Hash) Equal(h1 Hash) bool {
	for i := 0; i < HashSize; i++ {
		if h[i] != h1[i] {
			return false
		}
	}
	return true
}

// NewHash constructs a hash use the giving bytes
func NewHash(data []byte) Hash {
	h := Hash{}
	h.SetBytes(data)
	return h
}

// HexToHash coverts string `s` to hash
func HexToHash(s string) Hash {
	buf, _ := hex.DecodeString(s)
	return NewHash(buf)
}

// CalcHash calculates and returns the hash of the a + b
func CalcHash(a, b Hash) Hash {
	buf := make([]byte, 0)

	buf = append(buf, a.Reverse().Bytes()...)
	buf = append(buf, b.Reverse().Bytes()...)

	h := Sha256(buf)

	return h
}

// ComputeMerkleHash returns the merkle root hash of the hash lists
func ComputeMerkleHash(data []Hash) []Hash {
	length := len(data)
	if length <= 1 {
		return data
	}

	digests := make([]Hash, 0)
	for i := 0; i < length/2*2; i += 2 {
		h := CalcHash(data[i], data[i+1])
		h.Reverse()
		digests = append(digests, h)
	}

	if length%2 == 1 {
		h := CalcHash(data[length-1], data[length-1])
		h.Reverse()
		digests = append(digests, h)
	}
	data = digests

	return ComputeMerkleHash(data)
}

// GetMerkleHash returns the final hash
func GetMerkleHash(data []Hash) Hash {
	return ComputeMerkleHash(data)[0]
}

// Sha256 calculates and returns sha256 hash of the input data
func Sha256(data []byte) Hash {
	h := sha256.Sum256(data)

	return NewHash(h[:])
}

// DoubleSha256 calculates and returns double sha256 hash of the input data
func DoubleSha256(data []byte) Hash {
	h := sha256.Sum256(data)
	h = sha256.Sum256(h[:])

	return NewHash(h[:])
}

// Ripemd160 calculates the returns ripemd160 hash of the input data
func Ripemd160(data []byte) []byte {
	ripemd := ripemd160.New()
	ripemd.Write(data)

	return ripemd.Sum(nil)
}
