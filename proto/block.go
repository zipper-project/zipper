package proto

import (
	"github.com/golang/protobuf/proto"
	"github.com/zipper-project/zipper/common/crypto"
)

// NewBlock returns an new block
func NewBlock(prvHash, stateHash crypto.Hash,
	timeStamp, height, nonce uint32,
	txsHash crypto.Hash,
	Txs Transactions) *Block {
	return &Block{
		Header: NewBlockHeader(prvHash, stateHash, timeStamp, height, nonce, txsHash),
	}
}

// NewBlockHeader returns a blockheader
func NewBlockHeader(prvHash, stateHash crypto.Hash, timeStamp, height, nonce uint32, txsHash crypto.Hash) *BlockHeader {
	return &BlockHeader{
		prvHash.String(),
		stateHash.String(),
		timeStamp,
		nonce,
		txsHash.String(),
		height,
	}
}

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

// Height returns the block height
func (b *Block) Height() uint32 { return b.Header.Height }

// PreviousHash returns the previous hash of the block
func (b *Block) PreviousHash() string {
	return b.Header.PreviousHash
}
