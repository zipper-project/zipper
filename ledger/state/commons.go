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
	"strings"
)

var stateKeyDelimiter = string([]byte{0x00})

// ConstructCompositeKey returns a []byte that uniquely represents a given prefix and key.
// This assumes that prefix does not contain a 0x00 byte, but the key may
func ConstructCompositeKey(prefix string, key string) string {
	return strings.Join([]string{prefix, key}, stateKeyDelimiter)
}

// DecodeCompositeKey decodes the compositeKey constructed by ConstructCompositeKey method
// back to the original prefix and key form
func DecodeCompositeKey(compositeKey string) (string, string) {
	split := strings.SplitN(compositeKey, stateKeyDelimiter, 2)
	return split[0], split[1]
}

// Copy returns a copy of given bytes
func Copy(src []byte) []byte {
	dest := make([]byte, len(src))
	copy(dest, src)
	return dest
}
