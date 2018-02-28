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

package utils

import (
	"encoding/hex"
	"errors"
	"fmt"
)

var (
	ErrNonString = errors.New("cannot unmarshal non-string as hex data")
	ErrOddLength = errors.New("hex string has odd length")
	ErrSyntax    = errors.New("invalid hex")
)

// Bytes marshals/unmarshals as a JSON string .
type Bytes []byte

// MarshalText implements encoding.TextMarshaler
func (b Bytes) MarshalText() ([]byte, error) {
	result := make([]byte, len(b)*2)
	hex.Encode(result[:], b)
	return result, nil
}

// UnmarshalJSON implements json.Unmarshaler.
func (b *Bytes) UnmarshalJSON(input []byte) error {
	if !isString(input) {
		return ErrNonString
	}
	return b.UnmarshalText(input[1 : len(input)-1])
}

// UnmarshalText implements encoding.TextUnmarshaler.
func (b *Bytes) UnmarshalText(input []byte) error {
	raw, err := checkText(input)
	if err != nil {
		return err
	}
	dec := make([]byte, len(raw)/2)
	if _, err := hex.Decode(dec, raw); err != nil {
		return err
	}
	*b = dec
	return nil
}

func isString(input []byte) bool {
	return len(input) >= 2 && input[0] == '"' && input[len(input)-1] == '"'
}

func checkText(input []byte) ([]byte, error) {
	if len(input)%2 != 0 {
		return nil, ErrOddLength
	}
	return input, nil
}

// UnmarshalFixedText decodes the input as a string with 0x prefix. The length of out
// determines the required input length. This function is commonly used to implement the
// UnmarshalText method for fixed-size types.
func UnmarshalFixedText(input, out []byte) error {
	raw, err := checkText(input)
	if err != nil {
		return err
	}
	if len(raw)/2 != len(out) {
		return fmt.Errorf("hex string has length %d, want %d ", len(raw), len(out)*2)
	}
	// Pre-verify syntax before modifying out.
	for _, b := range raw {
		if decodeNibble(b) == badNibble {
			return ErrSyntax
		}
	}
	hex.Decode(out, raw)
	return nil
}

const badNibble = ^uint64(0)

func decodeNibble(in byte) uint64 {
	switch {
	case in >= '0' && in <= '9':
		return uint64(in - '0')
	case in >= 'A' && in <= 'F':
		return uint64(in - 'A' + 10)
	case in >= 'a' && in <= 'f':
		return uint64(in - 'a' + 10)
	default:
		return badNibble
	}
}
