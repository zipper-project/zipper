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


package protoManager

import (
	"fmt"
	"sync"

	"github.com/zipper-project/zipper/common/mpool"
	"github.com/zipper-project/zipper/peer"
	"github.com/zipper-project/zipper/peer/proto"
	mproto "github.com/zipper-project/zipper/proto"
	"github.com/zipper-project/zipper/types"
	"github.com/zipper-project/zipper/common/log"
	"github.com/zipper-project/zipper/blockchain"
)

type ProtoManager struct {
	bc *blockchain.Blockchain
	sync.Mutex
	wm map[mproto.ProtoID]*mpool.VirtualMachine
}

func NewProtoManager() *ProtoManager {
	return &ProtoManager{
		wm: make(map[mproto.ProtoID]*mpool.VirtualMachine),
	}
}

func (pm *ProtoManager) SetBlockChain(bc *blockchain.Blockchain) {
	pm.bc = bc
}

func (pm *ProtoManager) RegisterWorker(protocalID mproto.ProtoID, workers []mpool.VmWorker) error {
	pm.Lock()
	defer pm.Unlock()

	if _, ok := pm.wm[protocalID]; ok {
		return fmt.Errorf("wm: %s have beed registered", protocalID)
	}

	vm := mpool.CreateCustomVM(workers)
	_, err := vm.Open(string(uint32(protocalID)))
	if err != nil {
		return err
	}
	pm.wm[protocalID] = vm
	log.Debugf("===========>>>ProtoID: %+v, is Running(vm): %+v, map Running(vm): %+v, vm: %p", protocalID, vm.IsRunning(), pm.wm[protocalID].IsRunning(), vm)
	return nil
}

func (pm *ProtoManager) CreateStatusMsg() (*proto.Message, error) {
	statusMsg := &mproto.StatusMsg{
		Version: 1,
		StartHeight: pm.bc.CurrentHeight(),
	}

	statusMsgData, err := statusMsg.MarshalMsg()
	if err != nil {
		return nil, err
	}

	return &proto.Message{
		Header:&proto.Header{
			ProtoID: uint32(mproto.ProtoID_SyncWorker),
			MsgID: uint32(mproto.MsgType_BC_OnStatusMSg),
		},
		Payload: statusMsgData,
	}, nil
}

func (pm *ProtoManager) Handle(sendPeer *peer.Peer, msg *proto.Message) error {
	pm.Lock()
	defer pm.Unlock()


	err := pm.wm[mproto.ProtoID(msg.Header.ProtoID)].SendWorkCleanAsync(types.NewWorkerData(sendPeer, msg))
	log.Debugf("ProtoManager recv, ProtoID: %+v, Msg: %+v", msg.Header.ProtoID, msg.Header.MsgID)
	return err
}

func (pm *ProtoManager) InitAndRegisterWorker() {
	//pm.RegisterWorker(params.BlockSyncIdx, blocksync.NewSyncWorker())

}
