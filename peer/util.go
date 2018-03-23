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
	"net"

	"github.com/zipper-project/zipper/peer/proto"
)

// GetLocalIP returns the non loopback local IP of the host
func GetLocalIP() string {
	addrs, err := net.InterfaceAddrs()
	if err != nil {
		return ""
	}
	for _, address := range addrs {
		// check the address type and if it is not a loopback then display it
		if ipnet, ok := address.(*net.IPNet); ok && !ipnet.IP.IsLoopback() {
			if ipnet.IP.To4() != nil {
				return ipnet.IP.String()
			}
		}
	}
	return ""
}

var (
	baseProtocolName    = "zipper-base-protocol"
	baseProtocolVersion = "0.0.1"
	handshake           *proto.HandShake
)

func GetHandshake() *proto.HandShake {
	if handshake != nil {
		return handshake
	}

	handshake := &proto.HandShake{}
	handshake.Name = baseProtocolName
	handshake.Version = baseProtocolVersion
	handshake.Id = option.PeerID

	address := option.ListenAddress
	ip, port, _ := net.SplitHostPort(option.ListenAddress)
	if ip == "" {
		address = net.JoinHostPort(GetLocalIP(), port)
	}
	handshake.Address = address

	tp := VP
	if option.NVP {
		tp = NVP
	}
	handshake.Type = tp

	// handshake.Cert =
	// handshake.Signature =
	return handshake
}

func verifyHandShake(handshake *proto.HandShake) bool {
	if handshake.Name != baseProtocolName || handshake.Version != baseProtocolVersion {
		return false
	}
	if bytes.Equal(handshake.Id, option.PeerID) {
		return false
	}
	return true
}

func NewHandshakeMessage() *proto.Message {
	header := &proto.Header{
		ProtoID: BASE,
		MsgID:   HANDSHAKE,
	}

	handshake := GetHandshake()
	payload, _ := handshake.Serialize()
	return proto.NewMessage(header, payload)
}

func NewHandshakeAckMessage() *proto.Message {
	header := &proto.Header{
		ProtoID: BASE,
		MsgID:   HANDSHAKEACK,
	}

	handshake := GetHandshake()
	payload, _ := handshake.Serialize()
	return proto.NewMessage(header, payload)
}
