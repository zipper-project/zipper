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

package scip

import (
	"encoding/json"
	"fmt"
	"strings"
	"sync"
	"time"

	"github.com/zipper-project/zipper/common/log"
	"github.com/zipper-project/zipper/common/utils"
	"github.com/zipper-project/zipper/consensus"
	"github.com/zipper-project/zipper/proto"
)

//MINQUORUM  Define min quorum
const MINQUORUM = 3

//EMPTYREQUEST empty request id
const EMPTYREQUEST = 1136160000

//NewScip Create scip consenter
func NewScip(options *Options, stack consensus.IStack) *Scip {
	scip := &Scip{
		options:    options,
		stack:      stack,
		testing:    true,
		testChan:   make(chan struct{}),
		statistics: make(map[string]time.Duration),

		recvConsensusMsgChan: make(chan *Message, options.BufferSize),
		outputTxsChan:        make(chan *consensus.OutputTxs, options.BufferSize),
		broadcastChan:        make(chan *consensus.BroadcastConsensus, options.BufferSize),
	}
	scip.primaryHistory = make(map[string]int64)

	scip.vcStore = make(map[string]*viewChangeList)
	scip.coreStore = make(map[string]*scipCore)
	scip.committedRequests = make(map[uint32]*Committed)

	scip.blockTimer = time.NewTimer(scip.options.BlockTimeout)
	scip.blockTimer.Stop()

	if scip.options.N < MINQUORUM {
		scip.options.N = MINQUORUM
	}
	if scip.options.ResendViewChange < scip.options.ViewChange {
		scip.options.ResendViewChange = scip.options.ViewChange
	}
	if scip.options.BlockTimeout < scip.options.BatchTimeout {
		scip.options.BlockTimeout = scip.options.BatchTimeout
	}

	if scip.options.ViewChangePeriod > 0*time.Second && scip.options.ViewChangePeriod <= scip.options.ViewChange {
		scip.options.ViewChangePeriod = 60 * 3 * scip.options.ViewChange / 2
	}
	return scip
}

//Scip Define scip consenter
type Scip struct {
	sync.RWMutex
	options       *Options
	stack         consensus.IStack
	testing       bool
	testChan      chan struct{}
	statistics    map[string]time.Duration
	statisticsCnt int

	function func(int, proto.Transactions)

	recvConsensusMsgChan chan *Message
	outputTxsChan        chan *consensus.OutputTxs
	broadcastChan        chan *consensus.BroadcastConsensus

	height                uint32
	seqNo                 uint32
	execHeight            uint32
	execSeqNo             uint32
	priority              int64
	primaryHistory        map[string]int64
	primaryID             string
	lastPrimaryID         string
	newViewTimer          *time.Timer
	viewChangePeriodTimer *time.Timer

	rvc               *ViewChange
	vcStore           map[string]*viewChangeList
	rwVcStore         sync.RWMutex
	coreStore         map[string]*scipCore
	committedRequests map[uint32]*Committed

	fetched []*Committed

	blockTimer *time.Timer
	exit       chan struct{}
}

func (scip *Scip) Name() string {
	return "scip"
}

func (scip *Scip) String() string {
	bytes, _ := json.Marshal(scip.options)
	return string(bytes)
}

//Options
func (scip *Scip) Options() consensus.IOptions {
	return scip.options
}

//Start Start consenter serverice
func (scip *Scip) Start() {
	if scip.exit != nil {
		log.Warnf("Replica %s consenter already started", scip.options.ID)
		return
	}
	if scip.testing {
		//scip.testConsensus()
	}
	log.Infof("scip : %s", scip)
	log.Infof("Replica %s consenter started", scip.options.ID)
	scip.height = scip.stack.GetBlockchainInfo().Height
	scip.seqNo = scip.stack.GetBlockchainInfo().LastSeqNo
	scip.execHeight = scip.height
	scip.execSeqNo = scip.seqNo
	scip.priority = time.Now().UnixNano()
	scip.exit = make(chan struct{})
	for {
		select {
		case <-scip.exit:
			scip.exit = nil
			log.Infof("Replica %s consenter stopped", scip.options.ID)
			return
		case msg := <-scip.recvConsensusMsgChan:
			for msg != nil {
				msg = scip.processConsensusMsg(msg)
			}
		case <-scip.blockTimer.C:
			scip.sendEmptyRequest()
		}
	}
}

func (scip *Scip) sendEmptyRequest() {
	if scip.isPrimary() {
		scip.blockTimer.Stop()
		req := &Request{
			ID:     EMPTYREQUEST,
			Time:   uint32(time.Now().UnixNano()),
			Height: scip.height,
			Txs:    nil,
		}
		// scip.recvConsensusMsgChan <- &Message{
		// 	Type:    MESSAGEREQUEST,
		// 	Payload: utils.Serialize(req),
		// }
		scip.recvRequest(req)
	}
}

//Stop Stop consenter serverice
func (scip *Scip) Stop() {
	if scip.exit == nil {
		log.Warnf("Replica %s consenter already stopped", scip.options.ID)
		return
	}
	close(scip.exit)
}

// Quorum num of quorum
func (scip *Scip) Quorum() int {
	return scip.options.Q
}

// BatchSize size of batch
func (scip *Scip) BatchSize() int {
	return scip.options.BatchSize
}

// PendingSize size of batch pending
func (scip *Scip) PendingSize() int {
	return len(scip.coreStore)
}

// BatchTimeout size of batch timeout
func (scip *Scip) BatchTimeout() time.Duration {
	return scip.options.BatchTimeout
}

//ProcessBatches
func (scip *Scip) ProcessBatch(txs proto.Transactions, function func(int, proto.Transactions)) {
	scip.function = function
	if len(txs) == 0 {
		return
	}
	scip.startNewViewTimer()
	if !scip.isPrimary() {
		scip.function(0, txs)
		return
	}
	scip.function(1, txs)
	scip.function(3, txs)
	req := &Request{
		ID:   time.Now().UnixNano(),
		Time: uint32(time.Now().Unix()),
		Txs:  txs,
	}
	log.Debugf("Replica %s send Request for consensus %s", scip.options.ID, req.Name())
	scip.recvConsensusMsgChan <- &Message{
		Type:    MESSAGEREQUEST,
		Payload: utils.Serialize(req),
	}
}

//RecvConsensus Receive consensus data for consenter
func (scip *Scip) RecvConsensus(payload []byte) {
	msg := &Message{}
	if err := utils.Deserialize(payload, msg); err != nil {
		log.Errorf("Replica %s receive consensus message : unkown %v", scip.options.ID, err)
		return
	}
	scip.recvConsensusMsgChan <- msg
}

//BroadcastConsensusChannel Broadcast consensus data
func (scip *Scip) BroadcastConsensusChannel() <-chan *consensus.BroadcastConsensus {
	return scip.broadcastChan
}

//OutputTxsChannel Commit block data
func (scip *Scip) OutputTxsChannel() <-chan *consensus.OutputTxs {
	if scip.testing {
		return nil
	}
	return scip.outputTxsChan
}

func (scip *Scip) broadcast(to string, msg *Message) {
	scip.broadcastChan <- &consensus.BroadcastConsensus{
		To:      to,
		Payload: utils.Serialize(msg),
	}
}

func (scip *Scip) isPrimary() bool {
	return strings.Compare(scip.options.ID, scip.primaryID) == 0
}

func (scip *Scip) hasPrimary() bool {
	return strings.Compare("", scip.primaryID) != 0
}

func (scip *Scip) processConsensusMsg(msg *Message) *Message {
	log.Debugf("scip handle consensus message type %v ", msg.Type)
	switch tp := msg.Type; tp {
	case MESSAGEREQUEST:
		if request := msg.GetRequest(); request != nil {
			return scip.recvRequest(request)
		}
	case MESSAGEPREPREPARE:
		if preprepare := msg.GetPrePrepare(); preprepare != nil {
			return scip.recvPrePrepare(preprepare)
		}
	case MESSAGEPREPARE:
		if prepare := msg.GetPrepare(); prepare != nil {
			return scip.recvPrepare(prepare)
		}
	case MESSAGECOMMIT:
		if commit := msg.GetCommit(); commit != nil {
			return scip.recvCommit(commit)
		}
	case MESSAGECOMMITTED:
		if committed := msg.GetCommitted(); committed != nil {
			return scip.recvCommitted(committed)
		}
	case MESSAGEFETCHCOMMITTED:
		if fct := msg.GetFetchCommitted(); fct != nil {
			return nil
		}
	case MESSAGEVIEWCHANGE:
		if vc := msg.GetViewChange(); vc != nil {
			return scip.recvViewChange(vc)
		}
	default:
		log.Warnf("unsupport consensus message type %v ", tp)
	}
	return nil
}

func (scip *Scip) startNewViewTimer() {
	scip.Lock()
	defer scip.Unlock()
	if scip.newViewTimer == nil {
		id := time.Now().Truncate(scip.options.Request).Format("2006-01-02 15:04:05")
		scip.newViewTimer = time.AfterFunc(scip.options.Request, func() {
			scip.Lock()
			defer scip.Unlock()
			vc := &ViewChange{
				ID:            "scip-" + id,
				Priority:      scip.priority,
				PrimaryID:     scip.options.ID,
				SeqNo:         scip.execSeqNo,
				Height:        scip.execHeight,
				OptHash:       scip.options.Hash(),
				LastPrimaryID: scip.primaryID,
				ReplicaID:     scip.options.ID,
				Chain:         scip.options.Chain,
			}
			scip.sendViewChange(vc, fmt.Sprintf("request timeout(%s)", scip.options.Request))
			scip.newViewTimer = nil
		})
	}
}

func (scip *Scip) stopNewViewTimer() {
	scip.Lock()
	defer scip.Unlock()
	if scip.newViewTimer != nil {
		scip.newViewTimer.Stop()
		scip.newViewTimer = nil
	}
}

func (scip *Scip) startViewChangePeriodTimer() {
	if scip.options.ViewChangePeriod > 0*time.Second && scip.viewChangePeriodTimer == nil {
		scip.viewChangePeriodTimer = time.AfterFunc(scip.options.ViewChangePeriod, func() {
			vc := &ViewChange{
				ID:            "scip-period",
				Priority:      scip.priority,
				PrimaryID:     scip.options.ID,
				SeqNo:         scip.execSeqNo,
				Height:        scip.execHeight,
				OptHash:       scip.options.Hash(),
				LastPrimaryID: scip.primaryID,
				ReplicaID:     scip.options.ID,
				Chain:         scip.options.Chain,
			}
			scip.sendViewChange(vc, fmt.Sprintf("period timemout(%v)", scip.options.ViewChangePeriod))
		})
	}
}

func (scip *Scip) stopViewChangePeriodTimer() {
	if scip.viewChangePeriodTimer != nil {
		scip.viewChangePeriodTimer.Stop()
		scip.viewChangePeriodTimer = nil
	}
}

func (scip *Scip) recvFetchCommitted(fct *FetchCommitted) *Message {
	if fct.Chain != scip.options.Chain {
		log.Errorf("Replica %s received FetchCommitted(%d) from %s: ingnore, diff chain (%s-%s)", scip.options.ID, fct.SeqNo, fct.ReplicaID, scip.options.Chain, fct.Chain)
		return nil
	}

	log.Debugf("Replica %s received FetchCommitted(%d) from %s", scip.options.ID, fct.SeqNo, fct.ReplicaID)

	if request, ok := scip.committedRequests[fct.SeqNo]; ok {
		ctt := &Committed{
			SeqNo:     fct.SeqNo,
			Height:    request.Height,
			Digest:    request.Digest,
			Txs:       request.Txs,
			ErrTxs:    request.ErrTxs,
			Chain:     scip.options.Chain,
			ReplicaID: scip.options.ID,
		}
		scip.broadcast(scip.options.Chain, &Message{
			Type:    MESSAGECOMMITTED,
			Payload: utils.Serialize(ctt),
		})
	} else {
		log.Warnf("Replica %s received FetchCommitted(%d) from %s : ignore missing ", scip.options.ID, fct.SeqNo, fct.ReplicaID)
	}
	return nil
}

func (scip *Scip) sendViewChange(vc *ViewChange, reason string) {
	log.Infof("Replica %s send ViewChange(%s) for voter %s: %s", scip.options.ID, vc.ID, vc.PrimaryID, reason)
	msg := &Message{
		Type:    MESSAGEVIEWCHANGE,
		Payload: utils.Serialize(vc),
	}
	scip.recvConsensusMsgChan <- msg
	//scip.recvViewChange(vc)
	scip.broadcast(scip.options.Chain, msg)
}

type viewChangeList struct {
	vcs          []*ViewChange
	timeoutTimer *time.Timer
	resendTimer  *time.Timer
}

func (vcl *viewChangeList) start(scip *Scip) {
	vcl.timeoutTimer = time.AfterFunc(scip.options.ViewChange, func() {
		scip.rwVcStore.Lock()
		vcs := vcl.vcs
		delete(scip.vcStore, vcs[0].ID)
		scip.rwVcStore.Unlock()
		if len(vcs) >= scip.Quorum() {
			var tvc *ViewChange
			for _, v := range vcs {
				if v.PrimaryID == scip.lastPrimaryID {
					continue
				}
				if (scip.execSeqNo != 0 && v.SeqNo <= scip.execSeqNo) || v.Height < scip.execHeight || v.OptHash != scip.options.Hash() {
					continue
				}
				if p, ok := scip.primaryHistory[v.PrimaryID]; ok && p != v.Priority {
					continue
				}
				if tvc == nil {
					tvc = v
				} else if tvc.Priority > v.Priority {
					tvc = v
				}
			}
			log.Infof("Replica %s ViewChange(%s) timeout %s : voter %v", scip.options.ID, vcs[0].ID, scip.options.ViewChange, tvc)
			if tvc != nil && scip.rvc == nil {
				scip.rvc = tvc
				//vcl.resendTimer = time.AfterFunc(scip.options.ResendViewChange, func() {
				scip.rvc.ID += ":resend-" + tvc.PrimaryID
				scip.rvc.Chain = scip.options.Chain
				scip.rvc.ReplicaID = scip.options.ID
				scip.sendViewChange(scip.rvc, fmt.Sprintf("resend timeout(%s) - %s", scip.options.ResendViewChange, tvc.ID))
				scip.rvc = nil
				//})
			}
		} else {
			log.Debugf("Replica %s ViewChange(%s) timeout %s : %d", scip.options.ID, vcs[0].ID, scip.options.ViewChange, len(vcs))
		}
	})
}

func (vcl *viewChangeList) stop() {
	if vcl.timeoutTimer != nil {
		vcl.timeoutTimer.Stop()
		vcl.timeoutTimer = nil
	}
	if vcl.resendTimer != nil {
		vcl.resendTimer.Stop()
		vcl.resendTimer = nil
	}
}

func (scip *Scip) recvViewChange(vc *ViewChange) *Message {
	if vc.Chain != scip.options.Chain {
		log.Errorf("Replica %s received ViewChange(%s) from %s: ingnore, diff chain (%s-%s)", scip.options.ID, vc.ID, vc.ReplicaID, scip.options.Chain, vc.Chain)
		return nil
	}

	// if len(scip.primaryID) != 0 && vc.LastPrimaryID != scip.primaryID {
	// 	log.Errorf("Replica %s received ViewChange(%s) from %s: ingnore, diff primaryID (%s-%s)", scip.options.ID, vc.ID, vc.ReplicaID, scip.primaryID, vc.LastPrimaryID)
	// 	return nil
	// }

	scip.rwVcStore.Lock()
	defer scip.rwVcStore.Unlock()
	vcl, ok := scip.vcStore[vc.ID]
	if !ok {
		vcl = &viewChangeList{}
		scip.vcStore[vc.ID] = vcl
		vcl.start(scip)
	} else {
		for _, v := range vcl.vcs {
			if v.Chain == vc.Chain && v.ReplicaID == vc.ReplicaID {
				log.Warningf("Replica %s received ViewChange(%s) from %s: ingnore, duplicate, size %d", scip.options.ID, vc.ID, vc.ReplicaID, len(vcl.vcs))
				//scip.rwVcStore.Unlock()
				return nil
			}
		}
	}
	vcl.vcs = append(vcl.vcs, vc)
	vcs := vcl.vcs
	//scip.rwVcStore.Unlock()

	// if _, ok := scip.primaryHistory[vc.PrimaryID]; !ok && vc.PrimaryID == vc.ReplicaID {
	// 	scip.primaryHistory[vc.PrimaryID] = vc.Priority
	// }
	log.Infof("Replica %s received ViewChange(%s) from %s,  voter: %s %d %d %s, self: %d %d %s, size %d", scip.options.ID, vc.ID, vc.ReplicaID, vc.PrimaryID, vc.SeqNo, vc.Height, vc.OptHash, scip.execSeqNo, scip.execHeight, scip.options.Hash(), len(vcs))

	if len(vcs) >= scip.Quorum() {
		scip.stopNewViewTimer()
		// if len(vcs) == scip.Quorum() {
		// 	if scip.primaryID != "" {
		// 		scip.lastPrimaryID = scip.primaryID
		// 		scip.primaryID = ""
		// 		log.Infof("Replica %s ViewChange(%s) over : clear PrimaryID %s - %s", scip.options.ID, vcs[0].ID, scip.lastPrimaryID, vcs[0].ID)
		// 	}
		// }
		q := 0
		var tvc *ViewChange
		for _, v := range vcs {
			if v.PrimaryID == scip.lastPrimaryID {
				continue
			}
			if (scip.execSeqNo != 0 && v.SeqNo <= scip.execSeqNo) || v.Height < scip.execHeight || v.OptHash != scip.options.Hash() {
				continue
			}
			if p, ok := scip.primaryHistory[v.PrimaryID]; ok && p != v.Priority {
				continue
			}
			if tvc == nil {
				tvc = v
			} else if tvc.Priority > v.Priority {
				tvc = v
			}
		}
		for _, v := range vcs {
			if v.PrimaryID == scip.lastPrimaryID {
				continue
			}
			if (scip.execSeqNo != 0 && v.SeqNo <= scip.execSeqNo) || v.Height < scip.execHeight || v.OptHash != scip.options.Hash() {
				continue
			}
			if p, ok := scip.primaryHistory[v.PrimaryID]; ok && p != v.Priority {
				continue
			}
			if v.PrimaryID != tvc.PrimaryID {
				continue
			}
			q++
		}
		if q >= scip.Quorum() && scip.primaryID == "" {
			if scip.primaryID != "" {
				scip.lastPrimaryID = scip.primaryID
				scip.primaryID = ""
				log.Infof("Replica %s ViewChange(%s) over : clear PrimaryID %s - %s", scip.options.ID, vcs[0].ID, scip.lastPrimaryID, vcs[0].ID)
			}
			scip.newView(tvc)
		}
	}
	return nil
}

func (scip *Scip) newView(vc *ViewChange) {
	log.Infof("Replica %s vote new PrimaryID %s (%d %d) --- %s", scip.options.ID, vc.PrimaryID, vc.SeqNo, vc.Height, vc.ID)
	scip.primaryID = vc.PrimaryID
	scip.seqNo = vc.SeqNo
	scip.height = vc.Height
	scip.execSeqNo = scip.seqNo
	scip.execHeight = scip.height
	delete(scip.primaryHistory, scip.primaryID)
	if scip.primaryID == scip.options.ID {
		scip.priority = time.Now().UnixNano()
	}
	scip.stopViewChangePeriodTimer()
	scip.startViewChangePeriodTimer()

	for _, vcl := range scip.vcStore {
		vcl.stop()
	}
	scip.vcStore = make(map[string]*viewChangeList)
	for _, core := range scip.coreStore {
		scip.stopNewViewTimerForCore(core)
		if core.prePrepare != nil {
			scip.function(5, core.txs)
		}
	}
	scip.coreStore = make(map[string]*scipCore)

	for seqNo, req := range scip.committedRequests {
		if req.Height > scip.execHeight || seqNo > scip.execSeqNo {
			delete(scip.committedRequests, seqNo)
			scip.function(5, req.Txs)
		}
	}
}
