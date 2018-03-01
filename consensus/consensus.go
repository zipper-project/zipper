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

package consensus

import (
	"time"

	"github.com/zipper-project/zipper/consensus/types"
)

//BroadcastConsensus Define consensus data for broadcast
type BroadcastConsensus struct {
	To      string
	Payload []byte
}

//IOptions
type IOptions interface {
}

// OutputTxs Consensus output object
type OutputTxs struct {
	Txs    types.Transactions
	SeqNos []uint32
	Time   uint32
	Height uint32
}

// Consenter Interface for plugin consenser
type Consenter interface {
	Options() IOptions

	Start()
	Stop()
	Name() string

	Quorum() int
	BatchSize() int
	PendingSize() int
	BatchTimeout() time.Duration
	ProcessBatch(request types.Transactions, function func(int, types.Transactions))

	RecvConsensus([]byte)
	BroadcastConsensusChannel() <-chan *BroadcastConsensus
	OutputTxsChannel() <-chan *OutputTxs
}

// BlockchainInfo information of block chain
type BlockchainInfo struct {
	Height    uint32
	LastSeqNo uint32
}

// IStack Interface for other function for plugin consenser
type IStack interface {
	VerifyTxs(request types.Transactions) (types.Transactions, types.Transactions)
	GetBlockchainInfo() *BlockchainInfo
}
