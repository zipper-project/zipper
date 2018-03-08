package proto

import (
	"github.com/golang/protobuf/proto"
	"github.com/zipper-project/zipper/common/crypto"
)

// IInventory defines interface that broadcast data should implements
type IInventory interface {
	Hash() crypto.Hash
	Serialize() []byte
}

// Serialize returns the serialized bytes of a blockheader
func (bh *BlockHeader) Serialize() []byte {
	bytes, _ := proto.Marshal(bh)
	return bytes
}

// Deserialize deserialize the input data to header
func (bh *BlockHeader) Deserialize(data []byte) error {
	return proto.Unmarshal(data, bh)
}

// Hash returns the hash of the blockheader
func (bh *BlockHeader) Hash() crypto.Hash {
	return crypto.DoubleSha256(bh.Serialize())
}

// Serialize block data marshal
func (b *Block) Serialize() []byte {
	bytes, _ := proto.Marshal(b)
	return bytes
}

// Deserialize deserializes bytes to Block
func (b *Block) Deserialize(data []byte) error {
	return proto.Unmarshal(data, b)
}

// Hash returns the hash of the blockheader in block
func (b *Block) Hash() crypto.Hash {
	return b.Header.Hash()
}
