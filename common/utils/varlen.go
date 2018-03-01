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

package utils

import (
	"encoding/binary"

	//"fmt"
	"io"
	"math"
)

// VarInt converts n to varint bytes
// the raw protocol: https://en.bitcoin.it/wiki/Protocol_documentation#Variable_length_integer
func VarInt(n uint64) []byte {
	var (
		result []byte
	)

	// 1 - uint8
	if n < 0xFD {
		return []byte{byte(n)}
	}
	// 3 - uint16
	if n <= math.MaxUint16 {
		result = make([]byte, 3)
		result[0] = 0xFD
		binary.LittleEndian.PutUint16(result[1:], uint16(n))
		return result
	}
	// 5 - uint32
	if n <= math.MaxUint32 {
		result = make([]byte, 5)
		result[0] = 0xFE
		binary.LittleEndian.PutUint32(result[1:], uint32(n))
		return result
	}

	// 9 - uint64
	result = make([]byte, 9)
	result[0] = 0xFF
	binary.LittleEndian.PutUint64(result[1:], n)

	return result
}

// WriteVarInt writes varint bytes to stream
func WriteVarInt(w io.Writer, count uint64) {
	w.Write(VarInt(count))
}

// ReadVarInt decodes varint and returns real length
func ReadVarInt(r io.Reader) (uint64, error) {
	var (
		count = make([]byte, 1)
		buf   []byte
		err   error
	)

	if n, err := r.Read(count); err != nil {
		return uint64(n), err
	}

	switch count[0] {
	case 0xFD:
		buf := make([]byte, 2)
		_, err = io.ReadFull(r, buf)
		return (uint64)(binary.LittleEndian.Uint16(buf)), err
	case 0xFE:
		buf = make([]byte, 4)
		_, err = io.ReadFull(r, buf)
		return (uint64)(binary.LittleEndian.Uint32(buf)), err
	case 0xFF:
		buf = make([]byte, 8)
		_, err = io.ReadFull(r, buf)
		return binary.LittleEndian.Uint64(buf), err
	default:
		return uint64(uint8(count[0])), err
	}
}
