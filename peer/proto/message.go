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

package proto

import "github.com/golang/protobuf/proto"

func NewMessage(header *Header, payload []byte) *Message {
	//header.Magic = ""
	return &Message{
		Header:  header,
		Payload: payload,
	}
}

func (msg *Message) Serialize() ([]byte, error) {
	return proto.Marshal(msg)
}

func (msg *Message) Deserialize(data []byte) error {
	return proto.Unmarshal(data, msg)
}

func (handshake *HandShake) Serialize() ([]byte, error) {
	return proto.Marshal(handshake)
}

func (handshake *HandShake) Deserialize(data []byte) error {
	return proto.Unmarshal(data, handshake)
}
