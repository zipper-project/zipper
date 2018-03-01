package proto

import "github.com/golang/protobuf/proto"

type IIMsg interface {
	Marshal() ([]byte, error)
	Unmarshal(data []byte) error
}

type IMsg interface {
	Marshal() ([]byte, error)
	Unmarshal(data []byte) error
	SetPayload(data []byte)
}

func (m *Message) Marshal() ([]byte, error) {
	return proto.Marshal(m)
}

func (m *Message) Unmarshal(data []byte) error {
	return proto.Unmarshal(data, m)
}

func (m *Message) SetPayload(data []byte)  {
	m.Payload = data
}

func (m *GetBlocksMsg) Marshal() ([]byte, error) {
	return proto.Marshal(m)
}

func (m *GetBlocksMsg) Unmarshal(data []byte) error {
	return proto.Unmarshal(data, m)
}

func (m *GetInvMsg) Marshal() ([]byte, error) {
	return proto.Marshal(m)
}

func (m *GetInvMsg) Unmarshal(data []byte) error {
	return proto.Unmarshal(data, m)
}

func (m *GetDataMsg) Marshal() ([]byte, error) {
	return proto.Marshal(m)
}

func (m *GetDataMsg) Unmarshal(data []byte) error {
	return proto.Unmarshal(data, m)
}

func (m *OnBlockMsg) Marshal() ([]byte, error) {
	return proto.Marshal(m)
}

func (m *OnBlockMsg) Unmarshal(data []byte) error {
	return proto.Unmarshal(data, m)
}
