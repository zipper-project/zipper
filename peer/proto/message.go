package proto

import "github.com/golang/protobuf/proto"

func NewMessage(header *Header, payload []byte) *Message {
	//header.Magic = ""
	return &Message{
		Header:  header,
		Payload: payload,
	}
}

func (msg *Message) MarshalMsg() ([]byte, error) {
	return proto.Marshal(msg)
}

func (msg *Message) UnmarshalMsg(data []byte) error {
	return proto.Unmarshal(data, msg)
}

func (handshake *HandShake) MarshalMsg() ([]byte, error) {
	return proto.Marshal(handshake)
}

func (handshake *HandShake) UnmarshalMsg(data []byte) error {
	return proto.Unmarshal(data, handshake)
}
