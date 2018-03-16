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
	MarshalMsg() ([]byte, error)
	UnmarshalMsg(data []byte) error
}

func (m *StatusMsg) MarshalMsg() ([]byte, error) {
	return proto.Marshal(m)
}

func (m *StatusMsg) UnmarshalMsg(data []byte) error {
	return proto.Unmarshal(data, m)
}

func (m *GetBlocksMsg) MarshalMsg() ([]byte, error) {
	return proto.Marshal(m)
}

func (m *GetBlocksMsg) UnmarshalMsg(data []byte) error {
	return proto.Unmarshal(data, m)
}

func (m *GetInvMsg) MarshalMsg() ([]byte, error) {
	return proto.Marshal(m)
}

func (m *GetInvMsg) UnmarshalMsg(data []byte) error {
	return proto.Unmarshal(data, m)
}

func (m *GetDataMsg) MarshalMsg() ([]byte, error) {
	return proto.Marshal(m)
}

func (m *GetDataMsg) UnmarshalMsg(data []byte) error {
	return proto.Unmarshal(data, m)
}

func (m *OnBlockMsg) MarshalMsg() ([]byte, error) {
	return proto.Marshal(m)
}

func (m *OnBlockMsg) UnmarshalMsg(data []byte) error {
	return proto.Unmarshal(data, m)
}

func (m *OnTransactionMsg) MarshalMsg() ([]byte, error) {
	return proto.Marshal(m)
}

func (m *OnTransactionMsg) UnmarshalMsg(data []byte) error {
	return proto.Unmarshal(data, m)
}