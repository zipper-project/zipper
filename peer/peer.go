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
	"context"
	"encoding/binary"
	"encoding/hex"
	"errors"
	"fmt"
	"io"
	"net"
	"net/url"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/zipper-project/zipper/common/log"
	"github.com/zipper-project/zipper/peer/proto"
)

var (
	scheme            = "encode"
	delimiter         = "&"
	maxMsgSize uint32 = 1024 * 1024 * 100
)

const (
	VP  uint32 = 1
	NVP uint32 = 2
	ALL uint32 = VP | NVP
)

var TypeName = map[uint32]string{
	VP:  "VP",
	NVP: "NVP",
	ALL: "ALL",
}

// PeerID represents the peer identity
type PeerID []byte

func (p PeerID) String() string {
	return fmt.Sprintf("%s", hex.EncodeToString(p))
}

const (
	BASE         = 0
	HANDSHAKE    = iota + BASE*100 // 1
	HANDSHAKEACK                   // 2
	PING                           // 3
	PONG                           // 4
	PEERS                          // 5
	PEERSACK                       // 6
)

// Peer represents a peer in blockchain
type Peer struct {
	cancel    context.CancelFunc
	waitGroup sync.WaitGroup

	handshaked     bool
	lastActiveTime time.Time
	sendChannel    chan *proto.Message

	conn    net.Conn
	ID      PeerID
	Address string
	Type    uint32

	protocol    IProtocolManager
	peerManager *PeerManager
}

func NewPeer(conn net.Conn, protocol IProtocolManager, pm *PeerManager) *Peer {
	return &Peer{
		lastActiveTime: time.Now(),
		sendChannel:    make(chan *proto.Message, 100),
		conn:           conn,
		protocol:       protocol,
		peerManager:    pm,
	}
}

func (peer *Peer) Start() {
	if peer.cancel != nil {
		log.Warnf("Peer %s(%s->%s) already started.", peer.String(), peer.conn.LocalAddr().String(), peer.conn.RemoteAddr().String())
		return
	}
	ctx, cancel := context.WithCancel(context.Background())
	peer.cancel = cancel
	peer.waitGroup.Add(2)
	go peer.recv(ctx)
	go peer.send(ctx)
	log.Infof("Peer %s(%s->%s) Started", peer.String(), peer.conn.LocalAddr().String(), peer.conn.RemoteAddr().String())
}

func (peer *Peer) Stop() {
	peer.peerManager.remove(peer.conn)
	peer.conn.Close()
}

func (peer *Peer) stop() {
	if peer.cancel == nil {
		log.Warnf("Peer %s(%s->%s) already stopped.", peer.String(), peer.conn.LocalAddr().String(), peer.conn.RemoteAddr().String())
		return
	}
	peer.peerManager.Remove(peer.conn)
	peer.sendChannel = make(chan *proto.Message)
	peer.cancel()
	peer.waitGroup.Wait()
	log.Infof("Peer %s(%s->%s) Stopped", peer.String(), peer.conn.LocalAddr().String(), peer.conn.RemoteAddr().String())
	peer.conn = nil
	peer.cancel = nil
}

func (peer *Peer) SendMsg(msg *proto.Message) error {
	if msg.Header.ProtoID != BASE {
		log.Debugf("Peer %s(%s->%s) send msg %d/%d", peer, peer.conn.LocalAddr().String(), peer.conn.RemoteAddr().String(), msg.Header.ProtoID, msg.Header.MsgID)
	}
	select {
	case peer.sendChannel <- msg:
		return nil
	default:
		return fmt.Errorf("Peer %s(%s->%s) conn send channel fully ", peer.String(), peer.conn.LocalAddr().String(), peer.conn.RemoteAddr().String())
	}
}

// String is the representation of a peer as a URL.
func (peer *Peer) String() string {
	u := url.URL{Scheme: scheme}
	u.User = url.User(peer.ID.String())
	u.Host = peer.Address
	return u.String() + delimiter + strconv.FormatUint(uint64(peer.Type), 10)
}

// ParsePeer parses a peer designator.
func (peer *Peer) ParsePeer(rawurl string) error {
	urlAndType := strings.Split(rawurl, delimiter)
	peerURL := urlAndType[0]
	typeStr := urlAndType[1]
	u, err := url.Parse(peerURL)
	if err != nil {
		return err
	}
	if u.Scheme != scheme {
		return fmt.Errorf("invalid URL scheme, want \"%s\"", scheme)
	}
	// Parse the PeerID from the user portion.
	if u.User == nil {
		return errors.New("does not contain peer ID")
	}
	id, _ := hex.DecodeString(u.User.String())
	peerType, _ := strconv.ParseUint(typeStr, 10, 64)

	peer.ID = id
	peer.Address = u.Host
	peer.Type = uint32(peerType)
	return nil
}

func (peer *Peer) recv(ctx context.Context) {
	defer peer.stop()
	defer peer.waitGroup.Done()
	headerSize := 4
	for {
		select {
		case <-ctx.Done():
			return
		default:
		}

		//head
		peer.conn.SetReadDeadline(time.Now().Add(option.DeadLine))
		headerBytes := make([]byte, headerSize)
		if n, err := peer.conn.Read(headerBytes); err != nil {
			if err == io.EOF {
				log.Debugf("Peer %s(%s->%s) received close --- %s", peer, peer.conn.LocalAddr().String(), peer.conn.RemoteAddr().String(), err)
				return
			} else if opErr, ok := err.(*net.OpError); ok && opErr.Timeout() {
				continue
			}
			log.Errorf("Peer %s(%s->%s) conn read header --- %s", peer, peer.conn.LocalAddr().String(), peer.conn.RemoteAddr().String(), err)
			return
		} else if n != headerSize {
			err := fmt.Errorf("missing (expect %v, actual %v)", headerSize, n)
			log.Errorf("Peer %s(%s->%s) conn read header --- %s", peer, peer.conn.LocalAddr().String(), peer.conn.RemoteAddr().String(), err)
			return
		}
		//data
		dataSize := binary.LittleEndian.Uint32(headerBytes)
		if dataSize > maxMsgSize {
			err := fmt.Errorf("message too big")
			log.Errorf("Peer %s(%s->%s) conn read datasize --- %s", peer, peer.conn.LocalAddr().String(), peer.conn.RemoteAddr().String(), err)
			return
		}
		data := make([]byte, dataSize)
		peer.conn.SetReadDeadline(time.Now().Add(option.DeadLine))
		if n, err := io.ReadFull(peer.conn, data); err != nil {
			log.Errorf("Peer %s(%s->%s) conn read data --- %s", peer, peer.conn.LocalAddr().String(), peer.conn.RemoteAddr().String(), err)
			return
		} else if uint32(n) != dataSize {
			err := fmt.Errorf("missing (expect %v, actual %v)", dataSize, n)
			log.Errorf("Peer %s(%s->%s) conn read data --- %s", peer, peer.conn.LocalAddr().String(), peer.conn.RemoteAddr().String(), err)
			return
		}
		peer.lastActiveTime = time.Now()

		msg := &proto.Message{}
		if err := msg.Deserialize(data); err != nil {
			log.Errorf("Peer %s(%s->%s) conn read data --- %s", peer, peer.conn.LocalAddr().String(), peer.conn.RemoteAddr().String(), err)
			return
		}
		if msg.Header.ProtoID != BASE {
			log.Debugf("Peer %s(%s->%s) handle msg %d/%d", peer, peer.conn.LocalAddr().String(), peer.conn.RemoteAddr().String(), msg.Header.ProtoID, msg.Header.MsgID)
		}

		if !peer.handshaked {
			switch msg.Header.MsgID {
			case HANDSHAKE:
				handshake := &proto.HandShake{}
				if err := handshake.Deserialize(msg.Payload); err != nil {
					return
				}
				if !verifyHandShake(handshake) {
					log.Errorf("Peer %s(%s->%s) handle msg %d,  handshake --- failed to verify %#v", peer, peer.conn.LocalAddr().String(), peer.conn.RemoteAddr().String(), msg.Header.MsgID, msg)
					return
				}
				if peer.peerManager.Contains(handshake.Id) {
					log.Errorf("Peer %s(%s->%s) handle msg %d,  handshake --- id already exist", peer, peer.conn.LocalAddr().String(), peer.conn.RemoteAddr().String(), msg.Header.MsgID)
					return
				}
				peer.Address = handshake.Address
				peer.Type = handshake.Type
				peer.ID = handshake.Id
				peer.handshaked = true
				peer.SendMsg(NewHandshakeAckMessage())
				if peer.protocol != nil {
					msg, _ := peer.protocol.CreateStatusMsg()
					peer.SendMsg(msg)
				}
				go peer.loop(ctx)
			case HANDSHAKEACK:
				handshake := &proto.HandShake{}
				if err := handshake.Deserialize(msg.Payload); err != nil {
					return
				}
				if !verifyHandShake(handshake) {
					log.Errorf("Peer %s(%s->%s) handle msg %d,  handshake --- failed to verify", peer, peer.conn.LocalAddr().String(), peer.conn.RemoteAddr().String(), msg.Header.MsgID)
					return
				}
				if peer.peerManager.Contains(handshake.Id) {
					log.Errorf("Peer %s(%s->%s) handle msg %d,  handshake --- id already exist", peer, peer.conn.LocalAddr().String(), peer.conn.RemoteAddr().String(), msg.Header.MsgID)
					return
				}
				peer.Address = handshake.Address
				peer.Type = handshake.Type
				peer.ID = handshake.Id
				peer.handshaked = true
				if peer.protocol != nil {
					msg, _ := peer.protocol.CreateStatusMsg()
					peer.SendMsg(msg)
				}
				go peer.loop(ctx)
			default:
				log.Errorf("Peer %s(%s->%s) handle msg %d, no handshake", peer, peer.conn.LocalAddr().String(), peer.conn.RemoteAddr().String(), msg.Header.MsgID)
				return
			}
		} else {
			if msg.Header.ProtoID == BASE {
				switch msg.Header.MsgID {
				case HANDSHAKE, HANDSHAKEACK:
					log.Errorf("Peer %s(%s->%s) handle msg %d, already handshake", peer, peer.conn.LocalAddr().String(), peer.conn.RemoteAddr().String(), msg.Header.MsgID)
					return
				case PING:
					header := &proto.Header{
						ProtoID: BASE,
						MsgID:   PONG,
					}

					peer.SendMsg(proto.NewMessage(header, nil))
				case PONG:
				case PEERS:
					var peers []string
					peer.peerManager.IterFunc(func(peer *Peer) {
						if peer.handshaked {
							peers = append(peers, peer.String())
						}
					})
					header := &proto.Header{
						ProtoID: BASE,
						MsgID:   PEERSACK,
					}
					payload := strings.Join(peers, ",")

					peer.SendMsg(proto.NewMessage(header, []byte(payload)))
				case PEERSACK:
					for _, peerURL := range strings.Split(string(msg.Payload), ",") {
						if peerURL == "" {
							continue
						}
						tpeer := NewPeer(nil, nil, nil)
						if err := tpeer.ParsePeer(peerURL); err != nil {
							continue
						}
						peer.peerManager.Connect(tpeer, peer.protocol)
					}
				default:
					log.Errorf("Peer %s(%s->%s) handle msg %d/%d --- not support", peer, peer.conn.LocalAddr().String(), peer.conn.RemoteAddr().String(), msg.Header.ProtoID, msg.Header.MsgID)
					return
				}
			} else if peer.protocol != nil {
				if err := peer.protocol.Handle(peer, msg); err != nil {
					log.Errorf("Peer %s(%s->%s) handle msg %d/%d --- %s", peer, peer.conn.LocalAddr().String(), peer.conn.RemoteAddr().String(), msg.Header.ProtoID, msg.Header.MsgID, err)
					return
				}
			}
		}
	}
}

func (peer *Peer) send(ctx context.Context) {
	defer peer.waitGroup.Done()
	headerSize := 4
	for {
		select {
		case <-ctx.Done():
			return
		case msg := <-peer.sendChannel:
			dataBytes, err := msg.Serialize()
			if err != nil {
				log.Errorf("msg marshal error --- %s", err)
				continue
			}
			//headdata
			headerBytes := make([]byte, headerSize)
			dataSize := len(dataBytes)
			binary.LittleEndian.PutUint32(headerBytes, uint32(dataSize))
			var buf bytes.Buffer
			if num, err := buf.Write(headerBytes); num != headerSize || err != nil {
				log.Errorf("%s(%s->%s) conn send header --- %s", peer, peer.conn.LocalAddr().String(), peer.conn.RemoteAddr().String(), err)
				continue
			}
			if num, err := buf.Write(dataBytes); num != dataSize && err != nil {
				log.Errorf("%s(%s->%s) conn send header --- %s", peer, peer.conn.LocalAddr().String(), peer.conn.RemoteAddr().String(), err)
				continue
			}
			//send
			num, err := peer.conn.Write(buf.Bytes())
			if err != nil || buf.Len() != num {
				log.Errorf("%s(%s->%s) conn send header & data --- %s", peer, peer.conn.LocalAddr().String(), peer.conn.RemoteAddr().String(), err)
				continue
			}
		}

	}
}

func (peer *Peer) loop(ctx context.Context) {
	keepAliveTimer := time.NewTimer(option.KeepAliveInterval)
	for {
		select {
		case <-ctx.Done():
			return
		case <-keepAliveTimer.C:
			log.Infof("Peers %d", len(peer.peerManager.peers))
			keepAliveTimer.Stop()
			if time.Now().Sub(peer.lastActiveTime) > option.KeepAliveInterval {
				header := &proto.Header{
					ProtoID: BASE,
					MsgID:   PING,
				}
				msg := proto.NewMessage(header, nil)
				peer.SendMsg(msg)
			}

			header := &proto.Header{
				ProtoID: BASE,
				MsgID:   PEERS,
			}
			msg := proto.NewMessage(header, nil)
			peer.SendMsg(msg)
			keepAliveTimer.Reset(option.KeepAliveInterval)
		}
	}
}
