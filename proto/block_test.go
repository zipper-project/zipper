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

package proto

import (
	"bytes"
	"fmt"
	"testing"
	"time"

	"github.com/zipper-project/zipper/common/crypto"
)

func TestBlock(t *testing.T) {
	prviousHash := crypto.DoubleSha256([]byte("xxxx"))
	txsRootHash := crypto.DoubleSha256([]byte("xxxx"))
	stateHash := crypto.DoubleSha256([]byte("xxxx"))

	header := NewBlockHeader(prviousHash, txsRootHash, stateHash, uint32(time.Now().Unix()), uint32(100), 0)
	fmt.Println("BlockHeader hash", header.Hash())
	headerData := header.Serialize()
	testHeader := &BlockHeader{}
	testHeader.Deserialize(headerData)
	if !bytes.Equal(headerData, testHeader.Serialize()) {
		t.Errorf("BlockHeader.Serialize error")
	}

	block := NewBlock(header, nil)
	fmt.Println("Block hash", block.Hash())
	blkData := block.Serialize()
	testBlock := &Block{}
	testBlock.Deserialize(blkData)
	if !bytes.Equal(blkData, testBlock.Serialize()) {
		t.Errorf("Block.Serialize error")
	}
}
