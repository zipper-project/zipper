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

type IMsg interface {
	Serialize() ([]byte, error)
	Deserialize(data []byte) error
}

func (m *StatusMsg) Serialize() ([]byte, error) {
	return proto.Marshal(m)
}

func (m *StatusMsg) Deserialize(data []byte) error {
	return proto.Unmarshal(data, m)
}

func (m *GetBlocksMsg) Serialize() ([]byte, error) {
	return proto.Marshal(m)
}

func (m *GetBlocksMsg) Deserialize(data []byte) error {
	return proto.Unmarshal(data, m)
}

func (m *GetInvMsg) Serialize() ([]byte, error) {
	return proto.Marshal(m)
}

func (m *GetInvMsg) Deserialize(data []byte) error {
	return proto.Unmarshal(data, m)
}

func (m *GetDataMsg) Serialize() ([]byte, error) {
	return proto.Marshal(m)
}

func (m *GetDataMsg) Deserialize(data []byte) error {
	return proto.Unmarshal(data, m)
}

func (m *OnBlockMsg) Serialize() ([]byte, error) {
	return proto.Marshal(m)
}

func (m *OnBlockMsg) Deserialize(data []byte) error {
	return proto.Unmarshal(data, m)
}

func (m *OnTransactionMsg) Serialize() ([]byte, error) {
	return proto.Marshal(m)
}

func (m *OnTransactionMsg) Deserialize(data []byte) error {
	return proto.Unmarshal(data, m)
}
