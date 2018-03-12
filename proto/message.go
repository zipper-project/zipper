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