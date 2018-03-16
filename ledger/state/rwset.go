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

// Version encapsulates the version of a Key
type Version struct {
	BlockNum uint64
	TxNum    uint64
}

// KVRead captures a read operation performed during transaction simulation
type KVRead struct {
	Value   []byte
	Version *Version
}

// KVWrite captures a write (update/delete) operation performed during transaction simulation
type KVWrite struct {
	Value    []byte
	IsDelete bool
}

// KVRWSet encapsulates the read-write operation performed during transaction simulation
type KVRWSet struct {
	Reads  map[string]*KVRead
	Writes map[string]*KVWrite
}

//NewKVRWSet initialization
func NewKVRWSet() *KVRWSet {
	return &KVRWSet{
		Reads:  make(map[string]*KVRead),
		Writes: make(map[string]*KVWrite),
	}
}
