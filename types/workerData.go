package types

import (
	"github.com/zipper-project/zipper/peer/proto"
	"github.com/zipper-project/zipper/peer"
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