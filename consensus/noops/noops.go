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

package noops

import (
	"fmt"
	"time"

	"encoding/json"

	"github.com/zipper-project/zipper/common/log"
	"github.com/zipper-project/zipper/consensus"
	"github.com/zipper-project/zipper/proto"
)

// NewNoops Create Noops
func NewNoops(options *Options, stack consensus.IStack) *Noops {
	noops := &Noops{
		options:       options,
		stack:         stack,
		pendingChan:   make(chan *batchRequest, options.BufferSize),
		outputTxsChan: make(chan *consensus.OutputTxs, options.BufferSize),
		broadcastChan: make(chan *consensus.BroadcastConsensus, options.BufferSize),
	}
	noops.seqNo = noops.stack.GetBlockchainInfo().LastSeqNo
	noops.height = noops.stack.GetBlockchainInfo().Height

	noops.blockTimer = time.NewTimer(noops.options.BlockTimeout)
	noops.blockTimer.Stop()
	return noops
}

type batchRequest struct {
	Txs      proto.Transactions
	Function func(int, proto.Transactions)
}

// Noops Define Noops
type Noops struct {
	options *Options
	stack   consensus.IStack
	seqNo   uint32
	height  uint32

	pendingChan   chan *batchRequest
	outputTxsChan chan *consensus.OutputTxs
	broadcastChan chan *consensus.BroadcastConsensus

	blockTimer *time.Timer
	exit       chan struct{}
}

func (noops *Noops) Name() string {
	return "noops"
}

func (noops *Noops) String() string {
	bytes, _ := json.Marshal(noops.options)
	return string(bytes)
}

// IsRunning Noops consenter serverice already started
func (noops *Noops) IsRunning() bool {
	return noops.exit != nil
}

//Options
func (noops *Noops) Options() consensus.IOptions {
	return noops.options
}

// Start Start consenter serverice of Noops
func (noops *Noops) Start() {
	if noops.IsRunning() {
		return
	}
	log.Infof("noops : %s", noops)
	noops.exit = make(chan struct{})
	outputTxs := make(proto.Transactions, 0)
	seqNos := make([]uint32, 0)
	noops.blockTimer.Reset(noops.options.BlockTimeout)
	for {
		select {
		case <-noops.exit:
			noops.exit = nil
			noops.processBlock(outputTxs, seqNos, fmt.Sprintf("exit"))
			outputTxs = make(proto.Transactions, 0)
			seqNos = make([]uint32, 0)
			return
		case batchReq := <-noops.pendingChan:
			batchReq.Function(1, batchReq.Txs)
			txs, errTxs := noops.stack.VerifyTxs(batchReq.Txs)
			batchReq.Function(3, txs)
			batchReq.Function(6, errTxs)
			if len(txs) > 0 {
				if len(outputTxs) == 0 {
					noops.blockTimer.Reset(noops.options.BlockTimeout)
				}
				noops.seqNo++
				seqNos = append(seqNos, noops.seqNo)
				outputTxs = append(outputTxs, txs...)
				if len(outputTxs) >= noops.options.BlockSize {
					noops.processBlock(outputTxs, seqNos, fmt.Sprintf("size %d", noops.options.BlockSize))
					outputTxs = make(proto.Transactions, 0)
					seqNos = make([]uint32, 0)
				}

			}
		case <-noops.blockTimer.C:
			noops.processBlock(outputTxs, seqNos, fmt.Sprintf("timeout %s", noops.options.BlockTimeout))
			outputTxs = make(proto.Transactions, 0)
			seqNos = make([]uint32, 0)
		}
	}
}

func (noops *Noops) processBlock(txs proto.Transactions, seqNos []uint32, reason string) {
	noops.blockTimer.Stop()
	if len(seqNos) != 0 {
		noops.height++
		log.Infof("Noops write block %d (%d transactions)  %v : %s", noops.height, len(txs), seqNos, reason)
		noops.outputTxsChan <- &consensus.OutputTxs{Txs: txs, SeqNos: seqNos, Time: uint32(time.Now().Unix()), Height: noops.height}
	}
}

// Stop Stop consenter serverice of Noops
func (noops *Noops) Stop() {
	if noops.IsRunning() {
		close(noops.exit)
	}
}

// Quorum num of quorum
func (noops *Noops) Quorum() int {
	return 1
}

// BatchSize size of batch
func (noops *Noops) BatchSize() int {
	return noops.options.BatchSize
}

// PendingSize size of batch pending
func (noops *Noops) PendingSize() int {
	return len(noops.pendingChan)
}

// BatchTimeout size of batch timeout
func (noops *Noops) BatchTimeout() time.Duration {
	return noops.options.BatchTimeout
}

//ProcessBatches
func (noops *Noops) ProcessBatch(request proto.Transactions, function func(int, proto.Transactions)) {
	log.Infof("Noops ProcessBatch %d transactions", len(request))
	noops.pendingChan <- &batchRequest{Txs: request, Function: function}
	function(1, request)
}

// RecvConsensus Receive consensus data
func (noops *Noops) RecvConsensus(payload []byte) {
	panic("unspport in noops")
}

// BroadcastConsensusChannel Broadcast consensus data
func (noops *Noops) BroadcastConsensusChannel() <-chan *consensus.BroadcastConsensus {
	return noops.broadcastChan
}

// OutputTxsChannel Commit block data
func (noops *Noops) OutputTxsChannel() <-chan *consensus.OutputTxs {
	return noops.outputTxsChan
}
