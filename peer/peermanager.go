// Copyright (C) 2017, Zipper Team Technology Co.,Ltd.  All rights reserved.
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

package p2p

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
	pm.RLock()
	defer pm.RUnlock()
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
	if peer, ok := pm.peers[conn]; ok {
		peer.Stop()
	}
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

	if len(pm.peers) >= option.MaxPeers {
		log.Warnf("connected peer more than max peers.")
		return
	}
	if bytes.Equal(option.PeerID, peer.ID) {
		log.Warnf("can ont connect self[%s]", peer.ID)
		return
	}
	if pm.Contains(peer.ID) {
		log.Warnf("peer [%s] already connected", peer.ID)
		return
	}
	if peer.Address == "" || strings.HasPrefix(peer.Address, ":") {
		log.Warnf("wrong peer address %s.", peer.Address)
		return
	}

	go func() {
		i := 0
		log.Debugf("peer manager try connect : %s", peer.Address)
		for {
			if pm.Contains(peer.ID) || i > option.ReconnectTimes {
				break
			}
			if conn, err := net.Dial("tcp4", peer.Address); err == nil {
				pm.Add(conn, protocol)
				break
			}
			t := time.NewTimer(option.ReconnectInterval)
			<-t.C
			i++
		}
	}()
}
