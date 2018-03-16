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

package main

import (
	"github.com/zipper-project/zipper/peer"
	p2p "github.com/zipper-project/zipper/peer/proto"
	"github.com/zipper-project/zipper/proto"
)

var (
	srv        *peer.Server
	targetPeer *peer.Peer
)

func init() {
	option := peer.NewDefaultOption()
	option.ListenAddress = ":20066"
	option.BootstrapNodes = []string{"encode://303030315f616263@127.0.0.1:20166&1"}
	option.PeerID = []byte("0005_abc")
	srv = peer.NewServer(option, nil)
	targetPeer = &peer.Peer{}
	targetPeer.ParsePeer("encode://303030315f616263@127.0.0.1:20166&1")
}

func Relay(tx *proto.Transaction) {
	header := &p2p.Header{}
	header.ProtoID = uint32(proto.ProtoID_SyncWorker)
	header.MsgID = uint32(proto.MsgType_BC_OnTransactionMsg)
	txm := &proto.OnTransactionMsg{Transaction: tx}
	bts, _ := txm.MarshalMsg()
	msg := p2p.NewMessage(header, bts)
	srv.Unicast(msg, targetPeer.ID)
}
