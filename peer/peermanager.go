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

package peer

import (
	"bytes"
	"fmt"
	"net"
	"strings"
	"sync"
	"time"

	"github.com/zipper-project/zipper/common/log"
	"github.com/zipper-project/zipper/peer/proto"
)

type PeerManager struct {
	sync.RWMutex
	peers map[net.Conn]*Peer
}

func NewPeerManager() *PeerManager {
	return &PeerManager{
		peers: make(map[net.Conn]*Peer),
	}
}

func (pm *PeerManager) Stop() {
	pm.Lock()
	defer pm.Unlock()
	for _, peer := range pm.peers {
		peer.Stop()
	}
}

func (pm *PeerManager) Broadcast(msg *proto.Message, tp uint32) {
	pm.RLock()
	defer pm.RUnlock()
	for _, peer := range pm.peers {
		if peer.Type&tp > 0 {
			peer.SendMsg(msg)
		}
	}
}

func (pm *PeerManager) Unicast(msg *proto.Message, peerID []byte) {
	pm.RLock()
	defer pm.RUnlock()
	for _, peer := range pm.peers {
		if bytes.Equal(peer.ID, peerID) {
			peer.SendMsg(msg)
			break
		}
	}
}

func (pm *PeerManager) IterFunc(function func(peer *Peer)) {
	pm.RLock()
	defer pm.RUnlock()
	for _, peer := range pm.peers {
		function(peer)
	}
}

func (pm *PeerManager) Add(conn net.Conn, protocol IProtocolManager) (*Peer, error) {
	pm.Lock()
	defer pm.Unlock()

	if _, ok := pm.peers[conn]; ok {
		return nil, fmt.Errorf("conn alreay exist")
	}
	peer := NewPeer(conn, protocol, pm)
	pm.peers[conn] = peer
	peer.Start()
	return peer, nil
}

func (pm *PeerManager) Remove(conn net.Conn) {
	pm.Lock()
	defer pm.Unlock()
	delete(pm.peers, conn)
}

func (pm *PeerManager) remove(conn net.Conn) {
	delete(pm.peers, conn)
}

func (pm *PeerManager) Contains(id PeerID) bool {
	pm.RLock()
	defer pm.RUnlock()
	for _, peer := range pm.peers {
		if bytes.Equal(peer.ID, id) {
			return true
		}
	}
	return false
}

func (pm *PeerManager) Connect(peer *Peer, protocol IProtocolManager) {
	pm.RLock()
	defer pm.RUnlock()

	if bytes.Equal(option.PeerID, peer.ID) {
		return
	}

	if len(pm.peers) >= option.MaxPeers {
		log.Warnf("connected peer more than max peers.")
		return
	}
	if peer.Address == "" || strings.HasPrefix(peer.Address, ":") {
		log.Warnf("wrong peer address %s.", peer.Address)
		return
	}

	go func() {
		i := 0
		for {
			if pm.Contains(peer.ID) || i > option.ReconnectTimes {
				break
			}
			log.Debugf("peer manager try connect : %s %s(%d)", peer.ID, peer.Address, i+1)
			if conn, err := net.Dial("tcp4", peer.Address); err == nil {
				if _, err := pm.Add(conn, protocol); err != nil {
					log.Warnf("peer manager try connect : %s(%d) --- %s", peer.Address, i+1, err)
					conn.Close()
				}
				break
			}
			t := time.NewTimer(option.ReconnectInterval)
			<-t.C
			i++
		}
	}()
}
