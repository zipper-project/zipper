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

package blocksync

import (
	"errors"
	"fmt"

	"github.com/zipper-project/zipper/blockchain"
	"github.com/zipper-project/zipper/common/crypto"
	"github.com/zipper-project/zipper/common/log"
	"github.com/zipper-project/zipper/common/mpool"
	"github.com/zipper-project/zipper/ledger"
	"github.com/zipper-project/zipper/peer"
	msgProto "github.com/zipper-project/zipper/peer/proto"
	"github.com/zipper-project/zipper/proto"
	"github.com/zipper-project/zipper/types"
)

type SyncWorker struct {
	ledger         *ledger.Ledger
	bc             *blockchain.Blockchain
	expectedHeight uint32
}

func (worker *SyncWorker) VmJob(data interface{}) (interface{}, error) {
	workerData := data.(*types.WorkerData)
	msg := workerData.GetMsg()
	log.Debugf("======> syncWorker ProtoID: %+v, MsgID: %+v", msg.Header.MsgID, msg.Header.MsgID)
	switch proto.MsgType(msg.Header.MsgID) {
	case proto.MsgType_BC_OnStatusMSg:
		worker.OnStatus(workerData)
	case proto.MsgType_BC_GetBlocksMsg:
		worker.OnGetBlocks(workerData)
	case proto.MsgType_BC_GetInvMsg:
		worker.OnGetInv(workerData)
	case proto.MsgType_BC_GetDataMsg:
		worker.OnGetData(workerData)
	case proto.MsgType_BC_OnBlockMsg:
		worker.OnBlock(workerData)
	case proto.MsgType_BC_OnTransactionMsg:
		worker.OnTransaction(workerData)
	default:
		return nil, errors.New("Not support this type of message")
	}

	return nil, nil
}

func (worker *SyncWorker) VmReady() bool {
	return true
}

func NewSyncWorker(ledger *ledger.Ledger, bc *blockchain.Blockchain) *SyncWorker {
	return &SyncWorker{
		ledger: ledger,
		bc:     bc,
	}
}

func GetSyncWorkers(workerNums int, bc *blockchain.Blockchain) []mpool.VmWorker {
	cssWorkers := make([]mpool.VmWorker, 0)
	for i := 0; i < workerNums; i++ {
		cssWorkers = append(cssWorkers, NewSyncWorker(bc.GetLedger(), bc))
	}

	return cssWorkers
}

func (worker *SyncWorker) OnStatus(workerData *types.WorkerData) {
	statusMsg := proto.StatusMsg{}
	if err := statusMsg.Deserialize(workerData.GetMsg().Payload); err != nil {
		log.Errorf("Get invalid StatusMsg: %+v", workerData.GetMsg().Payload)
		return
	}

	if worker.bc.CurrentHeight() < statusMsg.StartHeight {
		if statusMsg.StartHeight > worker.expectedHeight {
			worker.expectedHeight = statusMsg.StartHeight
		}

		getBlocksMsg := &proto.GetBlocksMsg{
			LocatorHashes: []string{worker.bc.CurrentBlockHash().String()},
			HashStop:      crypto.Hash{}.String(),
		}

		worker.SendMsg(workerData.GetSendPeer(), proto.ProtoID_SyncWorker, proto.MsgType_BC_GetBlocksMsg, getBlocksMsg)
	} else if !worker.bc.Started() {
		worker.bc.StartServices()
	}
}

func (worker *SyncWorker) OnGetBlocks(workerData *types.WorkerData) {
	getBlocksMsg := proto.GetBlocksMsg{}
	if err := getBlocksMsg.Deserialize(workerData.GetMsg().Payload); err != nil {
		log.Errorf("Get invalid GetBlocksMsg: %+v", workerData.GetMsg().Payload)
		return
	}

	//check all hash for get, and get the starting hash
	var startHash string
	for _, h := range getBlocksMsg.LocatorHashes {
		_, err := worker.ledger.GetBlockByHash([]byte(h))
		if err != nil {
			break
		}

		startHash = h
	}

	hash := crypto.NewHash([]byte(startHash))
	hashes := []string{}
	var err error
	for {
		hash, err = worker.bc.GetNextBlockHash(hash)
		if err != nil || hash.Equal(crypto.Hash{}) {
			break
		} else {
			hashes = append(hashes, hash.String())
		}
	}

	if len(hashes) > 0 {
		getInvMsg := &proto.GetInvMsg{
			Type:  proto.InvType_block,
			Hashs: hashes,
		}

		worker.SendMsg(workerData.GetSendPeer(), proto.ProtoID_SyncWorker, proto.MsgType_BC_GetInvMsg, getInvMsg)
	}
}

func (worker *SyncWorker) OnGetInv(workerData *types.WorkerData) {
	if worker.bc.Synced() {
		return
	}

	getInvMsg := &proto.GetInvMsg{}
	if err := getInvMsg.Deserialize(workerData.GetMsg().Payload); err != nil {
		log.Errorf("Get invalid GetInvMsg: %+v", workerData.GetMsg().Payload)
		return
	}

	hashes := []string{}
	switch getInvMsg.Type {
	case proto.InvType_block:
		for _, h := range getInvMsg.Hashs {
			if tx, _ := worker.bc.GetTransaction(crypto.NewHash([]byte(h))); tx != nil {
				hashes = append(hashes, h)
			}
		}
	case proto.InvType_transaction:
		for _, h := range getInvMsg.Hashs {
			if block, _ := worker.ledger.GetBlockByHash([]byte(h)); block == nil {
				hashes = append(hashes, h)
			}
		}

	default:
		log.Errorf("Not support this type: %+v in GetInvMsg", getInvMsg.Type)
	}

	if len(hashes) > 0 {
		getDataMsg := &proto.GetDataMsg{}
		getDataMsg.InvList[0].Hashs = hashes

		worker.SendMsg(workerData.GetSendPeer(), proto.ProtoID_SyncWorker, proto.MsgType_BC_GetDataMsg, getDataMsg)
	}
}

func (worker *SyncWorker) OnGetData(workerData *types.WorkerData) {
	getDataMsg := &proto.GetDataMsg{}
	if err := getDataMsg.Deserialize(workerData.GetMsg().Payload); err != nil {
		log.Errorf("Get invalid GetDataMsg: %+v", workerData.GetMsg().Payload)
		return
	}

	for _, inv := range getDataMsg.InvList {
		switch inv.Type {
		case proto.InvType_block:
			for _, h := range inv.Hashs {
				if header, err := worker.ledger.GetBlockByHash([]byte(h)); err == nil && header != nil {
					txs, err := worker.ledger.GetTxsByBlockHash([]byte(h), 100)
					if err != nil {
						log.Errorf("Can't GetTxsByBlockHash, err: %+v", err)
						break
					}

					blockMsg := &proto.OnBlockMsg{
						Block: &proto.Block{
							Header:       header,
							Transactions: txs,
						},
					}

					worker.SendMsg(workerData.GetSendPeer(), proto.ProtoID_SyncWorker, proto.MsgType_BC_OnBlockMsg, blockMsg)
				}
			}
		case proto.InvType_transaction:
			for _, h := range inv.Hashs {
				if tx, _ := worker.bc.GetTransaction(crypto.NewHash([]byte(h))); tx != nil {
					txMsg := &proto.OnTransactionMsg{
						Transaction: tx,
					}

					worker.SendMsg(workerData.GetSendPeer(), proto.ProtoID_SyncWorker, proto.MsgType_BC_OnTransactionMsg, txMsg)
				}
			}
		default:
			log.Errorf("Not support this type: %+v in OnGetDataMsg", inv.Type)
		}
	}

}

func (worker *SyncWorker) OnBlock(workerData *types.WorkerData) {
	blockMsg := &proto.OnBlockMsg{}
	if err := blockMsg.Deserialize(workerData.GetMsg().Payload); err != nil {
		log.Errorf("Get invalid OnBlockMsg: %+v", workerData.GetMsg().Payload)
		return
	}

	if worker.bc.CurrentHeight()+1 < blockMsg.Block.Header.Height {
		getBlocksMsg := &proto.GetBlocksMsg{
			LocatorHashes: []string{worker.bc.CurrentBlockHash().String()},
			HashStop:      crypto.Hash{}.String(),
		}

		worker.SendMsg(workerData.GetSendPeer(), proto.ProtoID_SyncWorker, proto.MsgType_BC_GetBlocksMsg, getBlocksMsg)
	} else {
		worker.bc.Relay(blockMsg.Block)
		if !worker.bc.Started() && worker.bc.CurrentHeight() == worker.expectedHeight {
			worker.bc.StartServices()
		}
	}
}

func (worker *SyncWorker) OnTransaction(workerData *types.WorkerData) {
	txMsg := &proto.OnTransactionMsg{}
	if err := txMsg.Deserialize(workerData.GetMsg().Payload); err != nil {
		log.Errorf("Get invalid OnTransactionMsg: %+v", workerData.GetMsg().Payload)
		return
	}

	worker.bc.Relay(txMsg.Transaction)
}

func (worker *SyncWorker) SendMsg(peer *peer.Peer, protoID proto.ProtoID, msgID proto.MsgType, imsg proto.IMsg) error {
	msg := &msgProto.Message{}
	msg.Header.ProtoID = uint32(protoID)
	msg.Header.MsgID = uint32(msgID)

	imsgByte, err := imsg.Serialize()
	if err != nil {
		return fmt.Errorf("to create invalid msg: %+v", imsg)
	}

	msg.Payload = imsgByte
	err = peer.SendMsg(msg)
	if err != nil {
		return fmt.Errorf("can't send msg[%+v] to peer[%+v]", msg, peer.ID)
	}
	return nil
}
