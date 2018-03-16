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

package consensus

import (
	"time"

	"github.com/zipper-project/zipper/proto"
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
	Txs    proto.Transactions
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
	ProcessBatch(request proto.Transactions, function func(int, proto.Transactions))

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
	VerifyTxs(request proto.Transactions) (proto.Transactions, proto.Transactions)
	GetBlockchainInfo() *BlockchainInfo
}
