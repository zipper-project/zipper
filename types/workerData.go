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


package types

import (
	"github.com/zipper-project/zipper/peer/proto"
	"github.com/zipper-project/zipper/peer"
)

const (
	SyncWorkerProtoID = 1  // include transaction handler
	ConsensusWorkerProtoID = 2
)

type WorkerData struct {
	msg *proto.Message
	sendPeer *peer.Peer
}

func (wd *WorkerData) GetMsg() *proto.Message {
	return wd.msg
}

func (wd *WorkerData) GetSendPeer() *peer.Peer {
	return wd.sendPeer
}

func NewWorkerData(sendPeer *peer.Peer, msg *proto.Message) *WorkerData {
	return &WorkerData{
		sendPeer: sendPeer,
		msg: msg,
	}
}