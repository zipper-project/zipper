package proto

import "github.com/golang/protobuf/proto"

type IIMsg interface {
	Serialize() ([]byte, error)
	Deserialize(data []byte) error
}

type IMsg interface {
	Serialize() ([]byte, error)
	Deserialize(data []byte) error
	SetPayload(data []byte)
}

func (m *Message) Serialize() ([]byte, error) {
	return proto.Marshal(m)
}

func (m *Message) Deserialize(data []byte) error {
	return proto.Unmarshal(data, m)
}

func (m *Message) SetPayload(data []byte) {
	m.Payload = data
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
