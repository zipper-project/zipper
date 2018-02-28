package proto

import "github.com/golang/protobuf/proto"

func NewMessage(header *Header, payload []byte) *Message {
	//header.Magic = ""
	return &Message{
		Header:  header,
		Payload: payload,
	}
}

func (msg *Message) Marshal() ([]byte, error) {
	return proto.Marshal(msg)
}

func (msg *Message) Unmarshal(data []byte) error {
	return proto.Unmarshal(data, msg)
}

func (handshake *HandShake) Marshal() ([]byte, error) {
	return proto.Marshal(handshake)
}

func (handshake *HandShake) Unmarshal(data []byte) error {
	return proto.Unmarshal(data, handshake)
}
