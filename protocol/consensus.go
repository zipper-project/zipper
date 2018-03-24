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

package protocol

import (
	"github.com/zipper-project/zipper/common/log"
	"github.com/zipper-project/zipper/common/mpool"
)

type ConsensusWorker struct {
	consenter Consenter
}

func (worker *ConsensusWorker) VmJob(data interface{}) (interface{}, error) {
	workerData := data.(*WorkerData)
	msg := workerData.GetMsg()

	log.Debugf("======= ConsensusWorker recv proto: %+v, msg: %+v", msg.Header.ProtoID, msg.Header.MsgID)
	worker.consenter.RecvConsensus(msg.Payload)
	return nil, nil
}

func (worker *ConsensusWorker) VmReady() bool {
	return true
}

func NewConsensusWorker(consenter Consenter) *ConsensusWorker {
	return &ConsensusWorker{
		consenter: consenter,
	}
}

func GetConsensusWorkers(workerNums int, consenter Consenter) []mpool.VmWorker {
	cssWorkers := make([]mpool.VmWorker, 0)
	for i := 0; i < workerNums; i++ {
		cssWorkers = append(cssWorkers, NewConsensusWorker(consenter))
	}

	return cssWorkers
}
