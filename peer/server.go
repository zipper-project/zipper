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
package peer

import (
	"context"
	"crypto"
	"net"
	"strings"
	"sync"
	"time"

	"github.com/zipper-project/zipper/common/log"
	"github.com/zipper-project/zipper/peer/proto"
)

// Option is the p2p network configuration
type Option struct {
	ListenAddress     string
	DeadLine          time.Duration
	PrivateKey        *crypto.PrivateKey
	ReconnectInterval time.Duration
	ReconnectTimes    int
	KeepAliveInterval time.Duration
	KeepAliveTimes    int
	MaxPeers          int
	MinPeers          int
	Cores             int
	BootstrapNodes    []string
	CAPath            string
	ChainID           []byte
	PeerID            []byte
	NVP               bool
}

//option defines the default network configuration
var option = &Option{
	ListenAddress:     ":20166",
	DeadLine:          time.Second,
	PrivateKey:        nil,
	ReconnectInterval: 30 * time.Second,
	ReconnectTimes:    3,
	KeepAliveInterval: 15 * time.Second,
	KeepAliveTimes:    30,
	MaxPeers:          8,
	MinPeers:          3,
	Cores:             1,
	BootstrapNodes:    nil,
	CAPath:            "",
}

func NewDefaultOption() *Option {
	return option
}

// Server represent a p2p network server
type Server struct {
	cancel    context.CancelFunc
	waitGroup sync.WaitGroup

	protocol    IProtocolManager
	peerManager *PeerManager

	RootCertificate []byte
	Certificate     []byte

	option *Option
}

// NewServer return a new p2p server
func NewServer(option *Option) *Server {
	srv := &Server{
		peerManager: NewPeerManager(),
		option:      option,
	}
	return srv
}

// Start start p2p network run as goroutine
func (srv *Server) Start() {
	if srv.cancel != nil {
		log.Warnf("Server already started.")
		return
	}

	ctx, cancel := context.WithCancel(context.Background())
	srv.cancel = cancel
	addrs := strings.Split(srv.option.ListenAddress, ",")
	for _, addr := range addrs {
		srv.waitGroup.Add(1)
		go srv.listen(ctx, addr)
	}
	log.Infoln("Server Started")

	for _, bNode := range srv.option.BootstrapNodes {
		peer := &Peer{}
		if err := peer.ParsePeer(bNode); err != nil {
			continue
		}
		srv.peerManager.Connect(peer, srv.protocol)
	}
}

func (srv *Server) listen(ctx context.Context, addr string) (err error) {
	defer srv.waitGroup.Done()

	var (
		listener *net.TCPListener
		naddr    *net.TCPAddr
		conn     net.Conn
	)

	if naddr, err = net.ResolveTCPAddr("tcp4", addr); err != nil {
		log.Errorf("net.ResolveTCPAddr(\"tcp4\", \"%s\") error(%v)", addr, err)
		return
	}
	if listener, err = net.ListenTCP("tcp4", naddr); err != nil {
		log.Errorf("net.ListenTCP(\"tcp4\", \"%s\") error(%v)", addr, err)
		return
	}

	defer listener.Close()
	listener.SetDeadline(time.Now().Add(srv.option.DeadLine))
	for {
		select {
		case <-ctx.Done():
			return
		default:
		}

		if conn, err = listener.AcceptTCP(); err != nil {
			if opErr, ok := err.(*net.OpError); ok && opErr.Timeout() {
				continue
			} else {
				log.Errorf("listener.Accept(\"%s\") error(%v)", listener.Addr().String(), err)
				return
			}
		}
		// handle requests
		log.Debugf("Accept connection %s, %v", conn.RemoteAddr(), conn)
		if peer, err := srv.peerManager.Add(conn, srv.protocol); err == nil {
			peer.SendMsg(NewHandshakeMessage())
		}
	}
}

// Stop stop p2p network
func (srv *Server) Stop() {
	if srv.cancel == nil {
		log.Warnf("Server already stopped.")
		return
	}

	srv.cancel()
	srv.waitGroup.Wait()
	srv.cancel = nil
	srv.peerManager.Stop()
	log.Infoln("Server Stopped")
}

// Broadcast broadcasts message to remote peers
func (srv *Server) Broadcast(msg *proto.Message, tp uint32) {
	srv.peerManager.Broadcast(msg, tp)
}
