package p2p

import "github.com/zipper-project/zipper/peer/proto"

type IProtocolManager interface {
	Handle(*Peer, *proto.Message) error
}
