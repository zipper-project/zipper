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

package scip

import (
	"bytes"
	"fmt"
	"sort"
	"strings"
	"sync"
	"time"

	"github.com/zipper-project/zipper/common/crypto"
	"github.com/zipper-project/zipper/common/log"
	"github.com/zipper-project/zipper/common/utils"
	"github.com/zipper-project/zipper/consensus"
	"github.com/zipper-project/zipper/proto"
)

func merkleRootHash(txs []*proto.Transaction) crypto.Hash {
	if len(txs) > 0 {
		hashs := make([]crypto.Hash, 0)
		for _, tx := range txs {
			hashs = append(hashs, tx.Hash())
		}
		return crypto.ComputeMerkleHash(hashs)[0]
	}
	return crypto.Hash{}
}

type scipCore struct {
	digest       string
	txs          proto.Transactions
	errTxs       proto.Transactions
	prePrepare   *PrePrepare
	prepare      []*Prepare
	passPrepare  bool
	commit       []*Commit
	passCommit   bool
	newViewTimer *time.Timer

	startTime time.Time
	endTime   time.Time
	sync.RWMutex
}

func (scip *Scip) getscipCore(digest string) *scipCore {
	core, ok := scip.coreStore[digest]
	if ok {
		return core
	}

	core = &scipCore{
		digest: digest,
	}
	core.startTime = time.Now()
	scip.coreStore[digest] = core
	return core
}

func (scip *Scip) startNewViewTimerForCore(core *scipCore, replica string) {
	scip.stopNewViewTimer()
	scip.stopNewViewTimerForCore(core)
	scip.rwVcStore.Lock()
	defer scip.rwVcStore.Unlock()
	for k, vcl := range scip.vcStore {
		if strings.Contains(k, "resend") {
			continue
		}
		if /*(scip.hasPrimary() && strings.Contains(k, "scip")) ||*/ k == core.digest {
			vcs := []*ViewChange{}
			for _, vc := range vcl.vcs {
				if vc.ReplicaID == replica {
					continue
				}
				vcs = append(vcs, vc)
			}
			if len(vcs) == 0 {
				vcl.stop()
				delete(scip.vcStore, core.digest)
			} else {
				vcl.vcs = vcs
			}
		}
	}

	core.Lock()
	defer core.Unlock()
	if core.newViewTimer == nil && scip.hasPrimary() {
		core.newViewTimer = time.AfterFunc(scip.options.Request, func() {
			core.Lock()
			defer core.Unlock()
			vc := &ViewChange{
				ID:            core.digest,
				Priority:      scip.priority,
				PrimaryID:     scip.options.ID,
				SeqNo:         scip.execSeqNo,
				Height:        scip.execHeight,
				OptHash:       scip.options.Hash(),
				LastPrimaryID: scip.lastPrimaryID,
				ReplicaID:     scip.options.ID,
				Chain:         scip.options.Chain,
			}
			scip.sendViewChange(vc, fmt.Sprintf("%s request timeout(%s)", core.digest, scip.options.Request))
			core.newViewTimer = nil
		})
	}
}

func (scip *Scip) stopNewViewTimerForCore(core *scipCore) {
	core.Lock()
	defer core.Unlock()
	if core.newViewTimer != nil {
		core.newViewTimer.Stop()
		core.newViewTimer = nil
	}
}

func (scip *Scip) maybePassPrepare(core *scipCore) bool {
	q := 0
	nq := 0
	self := false
	hasPrimary := false
	for _, prepare := range core.prepare {
		if core.prePrepare.SeqNo != prepare.SeqNo || core.prePrepare.PrimaryID != prepare.PrimaryID ||
			core.prePrepare.Height != prepare.Height || core.prePrepare.OptHash != prepare.OptHash {
			continue
		}
		if prepare.ReplicaID == scip.options.ID {
			self = true
		}
		if prepare.ReplicaID == prepare.PrimaryID {
			hasPrimary = true
		}
		q++
		nq = prepare.Quorum
	}
	log.Debugf("Replica %s received Prepare for consensus %s, voted: %d(%d/%d,%v)", scip.options.ID, core.digest, len(core.prepare), q, nq, self)
	return hasPrimary && self && q >= nq
}

func (scip *Scip) maybePassCommit(core *scipCore) bool {
	q := 0
	nq := 0
	self := false
	hasPrimary := false
	for _, commit := range core.commit {
		if core.prePrepare.SeqNo != commit.SeqNo || core.prePrepare.PrimaryID != commit.PrimaryID ||
			core.prePrepare.Height != commit.Height || core.prePrepare.OptHash != commit.OptHash {
			continue
		}
		if commit.ReplicaID == scip.options.ID {
			self = true
		}
		if commit.ReplicaID == commit.PrimaryID {
			hasPrimary = true
		}
		q++
		nq = commit.Quorum
	}
	log.Debugf("Replica %s received Commit for consensus %s, voted: %d(%d/%d,%v)", scip.options.ID, core.digest, len(core.commit), q, nq, self)
	return hasPrimary && self && q >= nq
}

func (scip *Scip) recvRequest(request *Request) *Message {
	digest := request.Name()
	if scip.isPrimary() {
		if _, ok := scip.vcStore[digest]; ok {
			return nil
		}
		core := scip.getscipCore(digest)
		var txs, etxs proto.Transactions
		txs, etxs = scip.stack.VerifyTxs(request.Txs)
		core.txs = txs
		core.errTxs = etxs
		scip.seqNo++

		log.Debugf("Replica %s received Request for consensus %s", scip.options.ID, digest)
		request.Height = scip.height
		preprepare := &PrePrepare{
			PrimaryID:  scip.primaryID,
			SeqNo:      scip.seqNo,
			Height:     scip.height,
			OptHash:    scip.options.Hash(),
			MerkleRoot: string(merkleRootHash(core.errTxs).Bytes()),
			//Digest:    digest,
			Quorum:    scip.Quorum(),
			Request:   request,
			Chain:     scip.options.Chain,
			ReplicaID: scip.options.ID,
		}

		log.Debugf("Replica %s send PrePrepare for consensus %s", scip.options.ID, digest)
		scip.broadcast(scip.options.Chain, &Message{
			Type:    MESSAGEPREPREPARE,
			Payload: utils.Serialize(preprepare),
		})
		scip.recvPrePrepare(preprepare)
		scip.height++
	} else {
		log.Debugf("Replica %s received Request for consensus %s: ignore, backup", scip.options.ID, digest)
	}
	return nil
}

func (scip *Scip) recvPrePrepare(preprepare *PrePrepare) *Message {
	if preprepare.Request == nil {
		return nil
	}
	digest := preprepare.Request.Name()
	if preprepare.Chain != scip.options.Chain {
		log.Errorf("Replica %s received PrePrepare from %s for consensus %s: ignore, diff chain (%s==%s)", scip.options.ID, preprepare.ReplicaID, digest, preprepare.Chain, scip.options.Chain)
		return nil
	}
	if len(scip.primaryID) == 0 && len(scip.lastPrimaryID) == 0 {
		scip.primaryID = preprepare.ReplicaID
		scip.seqNo = preprepare.SeqNo
		scip.execSeqNo = scip.seqNo
	}
	if preprepare.ReplicaID != scip.primaryID {
		log.Errorf("Replica %s received PrePrepare from %s for consensus %s: ignore, diff primayID (%s==%s)", scip.options.ID, preprepare.ReplicaID, digest, preprepare.PrimaryID, scip.primaryID)
		return nil
	}

	core := scip.getscipCore(digest)
	if core.prePrepare != nil {
		log.Errorf("Replica %s received PrePrepare from %s for consensus %s: already exist ", scip.options.ID, preprepare.ReplicaID, digest)
		vc := &ViewChange{
			ID:            digest,
			Priority:      scip.priority,
			PrimaryID:     scip.options.ID,
			SeqNo:         scip.execSeqNo,
			Height:        scip.execHeight,
			OptHash:       scip.options.Hash(),
			LastPrimaryID: scip.lastPrimaryID,
			ReplicaID:     scip.options.ID,
			Chain:         scip.options.Chain,
		}
		scip.sendViewChange(vc, fmt.Sprintf("already exist"))
		return nil
	}

	if !scip.isPrimary() {

		if preprepare.SeqNo != scip.seqNo+1 {
			log.Errorf("Replica %s received PrePrepare from %s for consensus %s: ignore, wrong seqNo (%d==%d)", scip.options.ID, preprepare.ReplicaID, digest, preprepare.SeqNo, scip.seqNo)
			vc := &ViewChange{
				ID:            digest,
				Priority:      scip.priority,
				PrimaryID:     scip.options.ID,
				SeqNo:         scip.execSeqNo,
				Height:        scip.execHeight,
				OptHash:       scip.options.Hash(),
				LastPrimaryID: scip.lastPrimaryID,
				ReplicaID:     scip.options.ID,
				Chain:         scip.options.Chain,
			}
			scip.sendViewChange(vc, fmt.Sprintf("wrong seqNo (%d==%d)", preprepare.SeqNo, scip.seqNo+1))
			return nil
		}
		if preprepare.Height != scip.height {
			log.Errorf("Replica %s received PrePrepare from %s for consensus %s: ignore, wrong height (%d==%d)", scip.options.ID, preprepare.ReplicaID, digest, preprepare.Height, scip.height)
			vc := &ViewChange{
				ID:            digest,
				Priority:      scip.priority,
				PrimaryID:     scip.options.ID,
				SeqNo:         scip.execSeqNo,
				Height:        scip.execHeight,
				OptHash:       scip.options.Hash(),
				LastPrimaryID: scip.lastPrimaryID,
				ReplicaID:     scip.options.ID,
				Chain:         scip.options.Chain,
			}
			scip.sendViewChange(vc, fmt.Sprintf("wrong seqNo (%d==%d)", preprepare.SeqNo, scip.seqNo+1))
			return nil
		}
		var txs, etxs proto.Transactions
		txs, etxs = scip.stack.VerifyTxs(preprepare.Request.Txs)
		if !bytes.Equal(merkleRootHash(etxs).Bytes(), []byte(preprepare.MerkleRoot)) {
			log.Errorf("Replica %s received PrePrepare from %s for consensus %s: failed to verify", scip.options.ID, preprepare.ReplicaID, digest)
			vc := &ViewChange{
				ID:            digest,
				Priority:      scip.priority,
				PrimaryID:     scip.options.ID,
				SeqNo:         scip.execSeqNo,
				Height:        scip.execHeight,
				OptHash:       scip.options.Hash(),
				LastPrimaryID: scip.lastPrimaryID,
				ReplicaID:     scip.options.ID,
				Chain:         scip.options.Chain,
			}
			scip.sendViewChange(vc, fmt.Sprintf("failed to verify"))
			return nil
		}
		core.txs = txs
		core.errTxs = etxs
		scip.seqNo++
		scip.height++
	}

	log.Debugf("Replica %s received PrePrepare from %s for consensus %s, seqNo %d", scip.options.ID, preprepare.ReplicaID, digest, preprepare.SeqNo)

	scip.startNewViewTimerForCore(core, preprepare.ReplicaID)
	core.prePrepare = preprepare
	prepare := &Prepare{
		PrimaryID: scip.primaryID,
		SeqNo:     preprepare.SeqNo,
		Height:    preprepare.Height,
		OptHash:   scip.options.Hash(),
		Digest:    digest,
		Quorum:    scip.Quorum(),
		Chain:     scip.options.Chain,
		ReplicaID: scip.options.ID,
	}

	log.Debugf("Replica %s send Prepare for consensus %s", scip.options.ID, prepare.Digest)
	scip.broadcast(scip.options.Chain, &Message{Type: MESSAGEPREPARE, Payload: utils.Serialize(prepare)})
	scip.recvPrepare(prepare)
	return nil
}

func (scip *Scip) recvPrepare(prepare *Prepare) *Message {
	if _, ok := scip.committedRequests[prepare.SeqNo]; ok || prepare.SeqNo <= scip.execSeqNo {
		log.Debugf("Replica %s received Prepare from %s for consensus %s: ignore delay(%d<=%d)", scip.options.ID, prepare.ReplicaID, prepare.Digest, prepare.SeqNo, scip.execSeqNo)
		return nil
	}

	core := scip.getscipCore(prepare.Digest)
	if prepare.Chain != scip.options.Chain {
		log.Errorf("Replica %s received Prepare from %s for consensus %s: ignore, diff chain (%s==%s)", scip.options.ID, prepare.ReplicaID, prepare.Digest, prepare.Chain, scip.options.Chain)
		return nil
	}

	log.Debugf("Replica %s received Prepare from %s for consensus %s", scip.options.ID, prepare.ReplicaID, prepare.Digest)

	scip.startNewViewTimerForCore(core, prepare.ReplicaID)
	core.prepare = append(core.prepare, prepare)
	if core.prePrepare == nil {
		log.Debugf("Replica %s received Prepare for consensus %s, voted: %d", scip.options.ID, prepare.Digest, len(core.prepare))
		return nil
	}
	if core.passPrepare || !scip.maybePassPrepare(core) {
		return nil
	}
	core.passPrepare = true
	commit := &Commit{
		PrimaryID: scip.primaryID,
		SeqNo:     core.prePrepare.SeqNo,
		Height:    core.prePrepare.Height,
		OptHash:   scip.options.Hash(),
		Digest:    prepare.Digest,
		Quorum:    scip.Quorum(),
		Chain:     scip.options.Chain,
		ReplicaID: scip.options.ID,
	}

	log.Debugf("Replica %s send Commit for consensus %s", scip.options.ID, commit.Digest)
	scip.broadcast(scip.options.Chain, &Message{Type: MESSAGECOMMIT, Payload: utils.Serialize(commit)})
	scip.recvCommit(commit)
	return nil
}

func (scip *Scip) recvCommit(commit *Commit) *Message {
	if _, ok := scip.committedRequests[commit.SeqNo]; ok || commit.SeqNo <= scip.execSeqNo {
		log.Debugf("Replica %s received Commit from %s for consensus %s: ignore delay(%d<=%d)", scip.options.ID, commit.ReplicaID, commit.Digest, commit.SeqNo, scip.execSeqNo)
		return nil
	}

	core := scip.getscipCore(commit.Digest)
	if commit.Chain != scip.options.Chain {
		log.Errorf("Replica %s received Commit from %s for consensus %s: ignore, diff chain (%s==%s)", scip.options.ID, commit.ReplicaID, commit.Digest, commit.Chain, scip.options.Chain)
		return nil
	}

	log.Debugf("Replica %s received Commit from %s for consensus %s", scip.options.ID, commit.ReplicaID, commit.Digest)

	scip.startNewViewTimerForCore(core, commit.ReplicaID)
	core.commit = append(core.commit, commit)
	if core.prePrepare == nil {
		log.Debugf("Replica %s received Commit for consensus %s, voted: %d", scip.options.ID, commit.Digest, len(core.commit))
		return nil
	}
	if core.passCommit || !scip.maybePassCommit(core) {
		return nil
	}
	scip.stopNewViewTimerForCore(core)
	core.passCommit = true
	core.endTime = time.Now()
	committed := &Committed{
		SeqNo:     core.prePrepare.SeqNo,
		Height:    core.prePrepare.Height,
		Digest:    commit.Digest,
		Txs:       core.txs,
		ErrTxs:    core.errTxs,
		Chain:     scip.options.Chain,
		ReplicaID: scip.options.ID,
	}

	log.Debugf("Replica %s send Committed for consensus %s", scip.options.ID, commit.Digest)
	scip.broadcast(scip.options.Chain, &Message{Type: MESSAGECOMMITTED, Payload: utils.Serialize(committed)})
	scip.recvCommitted(committed)
	return nil
}

func (scip *Scip) recvCommitted(committed *Committed) *Message {
	if committed.Chain != scip.options.Chain {
		log.Debugf("Replica %s received Committed from %s for consensus %s: ignore diff chain", scip.options.ID, committed.ReplicaID, committed.Digest)
		return nil
	}
	if _, ok := scip.committedRequests[committed.SeqNo]; ok || committed.SeqNo <= scip.execSeqNo {
		log.Debugf("Replica %s received Committed from %s for consensus %s: ignore delay(%d<=%d)", scip.options.ID, committed.ReplicaID, committed.Digest, committed.SeqNo, scip.execSeqNo)
		return nil
	}

	digest := committed.Digest
	if committed.ReplicaID == scip.options.ID {
		log.Debugf("Replica %s received Committed from %s for consensus %s", scip.options.ID, committed.ReplicaID, digest)
		//scip.committedRequests[committed.SeqNo] = committed.Request
	} else {
		fetched := []*Committed{}
		for _, c := range scip.fetched {
			if c.SeqNo == committed.SeqNo && c.ReplicaID == committed.ReplicaID {
				continue
			}
			if c.SeqNo > scip.execSeqNo {
				fetched = append(fetched, c)
			}
		}
		scip.fetched = fetched
		scip.fetched = append(scip.fetched, committed)

		q := 0
		for _, c := range scip.fetched {
			if c.SeqNo == committed.SeqNo {
				q++
			}
		}
		log.Debugf("Replica %s received Committed from %s for consensus %s, vote: %d/%d", scip.options.ID, committed.ReplicaID, digest, scip.Quorum(), q)
		if q >= scip.Quorum() {
			//scip.committedRequests[committed.SeqNo] = committed.Request
		} else {
			return nil
		}
	}
	if scip.seqNo == 0 {
		scip.seqNo = committed.SeqNo
		scip.execSeqNo = scip.seqNo
	}
	scip.committedRequests[committed.SeqNo] = committed
	d, _ := time.ParseDuration("0s")
	if core, ok := scip.coreStore[digest]; ok {
		scip.stopNewViewTimerForCore(core)
		delete(scip.coreStore, digest)
		d = core.endTime.Sub(core.startTime)
		if core.txs != nil {
			scip.function(3, core.txs)
		}
	}
	//remove invalid ViewChange
	scip.rwVcStore.Lock()
	keys := []string{}
	for key, vcl := range scip.vcStore {
		if vcl.vcs[0].SeqNo > committed.SeqNo {
			continue
		}
		vcl.stop()
		keys = append(keys, key)
	}
	for _, key := range keys {
		delete(scip.vcStore, key)
	}
	scip.rwVcStore.Unlock()
	log.Infof("Replica %s execute for consensus %s: seqNo:%d height:%d, duration: %s", scip.options.ID, committed.Digest, committed.SeqNo, committed.Height, d)
	scip.execute()

	// for _, core := range scip.coreStore {
	// 	if core.prePrepare != nil {
	// 		preprepare := core.prePrepare
	// 		if preprepare.SeqNo <= scip.execSeqNo {
	// 			scip.stopNewViewTimerForCore(core)
	// 			delete(scip.coreStore, core.digest)
	// 		}
	// 	} else if len(core.prepare) > 0 {
	// 		prepare := core.prepare[0]
	// 		if prepare.SeqNo <= scip.execSeqNo {
	// 			scip.stopNewViewTimerForCore(core)
	// 			delete(scip.coreStore, core.digest)
	// 		}
	// 	} else if len(core.commit) > 0 {
	// 		commit := core.commit[0]
	// 		if commit.SeqNo <= scip.execSeqNo {
	// 			scip.stopNewViewTimerForCore(core)
	// 			delete(scip.coreStore, core.digest)
	// 		}
	// 	}
	// }
	return nil
}

type Uint32Slice []uint32

func (us Uint32Slice) Len() int {
	return len(us)
}
func (us Uint32Slice) Less(i, j int) bool {
	return us[i] < us[j]
}
func (us Uint32Slice) Swap(i, j int) {
	us[i], us[j] = us[j], us[i]
}

func (scip *Scip) execute() {
	keys := Uint32Slice{}
	for seqNo := range scip.committedRequests {
		keys = append(keys, seqNo)
	}
	sort.Sort(keys)

	nextExec := scip.execSeqNo + 1
	for seqNo, request := range scip.committedRequests {
		if nextExec > seqNo && nextExec-seqNo > uint32(scip.options.K*3) {
			delete(scip.committedRequests, seqNo)
		} else if seqNo == nextExec {
			scip.execSeqNo = nextExec
			if scip.seqNo < scip.execSeqNo {
				scip.seqNo = scip.execSeqNo
			}
			if scip.height < request.Height {
				scip.height = request.Height
			}
			if scip.execHeight != request.Height {
				panic(fmt.Sprintf("noreachable(%d +2 == %d)", scip.execHeight, request.Height))
			}
			scip.function(3, request.Txs)
			scip.function(4, request.ErrTxs)
			scip.execHeight = request.Height + 1
			var seqNos []uint32
			seqNos = append(seqNos, seqNo)
			scip.processBlock(request.Txs, seqNos, fmt.Sprintf("block timeout(%s), block size(%d)", scip.options.BatchTimeout, scip.options.BatchSize))
			nextExec = seqNo + 1
		} else if seqNo > nextExec {
			if seqNo-nextExec > uint32(scip.options.K) {
				log.Debugf("Replica %s need seqNo %d ", scip.options.ID, nextExec)
				for n, r := range scip.committedRequests {
					log.Debugf("Replica %s seqNo %d : %s", scip.options.ID, n, r.Digest)
				}
				log.Panicf("Replica %s fallen behind over %d", scip.options.ID, scip.options.K)
			}
			log.Warnf("Replica %s fetch committed %d ", scip.options.ID, nextExec)
			fc := &FetchCommitted{
				ReplicaID: scip.options.ID,
				Chain:     scip.options.Chain,
				SeqNo:     nextExec,
			}
			scip.broadcast(scip.options.Chain, &Message{Type: MESSAGEFETCHCOMMITTED, Payload: utils.Serialize(fc)})
			break
		}
	}
}

func (scip *Scip) processBlock(txs proto.Transactions, seqNos []uint32, reason string) {
	scip.blockTimer.Stop()
	if len(seqNos) != 0 {
		log.Infof("Replica %s write block %d (%d transactions)  %v : %s", scip.options.ID, scip.execHeight, len(txs), seqNos, reason)
		t := uint32(time.Now().Unix())
		if n := len(txs); n > 0 {
			t = txs[len(txs)-1].CreateTime()
		}
		scip.outputTxsChan <- &consensus.OutputTxs{Txs: txs, SeqNos: seqNos, Time: t, Height: scip.execHeight}
	} else {
		panic("unreachable")
	}
}
