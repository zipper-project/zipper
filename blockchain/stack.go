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

package blockchain

import (
	"github.com/zipper-project/zipper/consensus"
	"github.com/zipper-project/zipper/proto"
)

//VerifyTxs verify
func (bc *Blockchain) VerifyTxs(txs proto.Transactions) (proto.Transactions, proto.Transactions) {
	return bc.validator.VerifyTxs(txs)
}

func (bc *Blockchain) GetBlockchainInfo() *consensus.BlockchainInfo {
	height, _ := bc.ledger.Height()
	return &consensus.BlockchainInfo{
		Height: height,
	}
}
