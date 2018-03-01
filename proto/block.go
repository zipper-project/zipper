package proto

import (
	"errors"
	"sync/atomic"

	"github.com/golang/protobuf/proto"
	"github.com/zipper-project/zipper/common/crypto"
)

//Block block
type Block struct {
	BlockData
	// caches
	hash atomic.Value
}

// IInventory defines interface that broadcast data should implements
type IInventory interface {
	Hash() crypto.Hash
	Serialize() []byte
}

// Serialize returns the serialized bytes of a blockheader
func (h *BlockHeader) Serialize() []byte {
	bytes, _ := proto.Marshal(h)
	return bytes
}

// Deserialize deserialize the input data to header
func (h *BlockHeader) Deserialize(data []byte) error {
	return proto.Unmarshal(data, h)
}

// Hash returns the hash of the blockheader
func (h *BlockHeader) Hash() crypto.Hash {
	return crypto.DoubleSha256(h.Serialize())
}

//Marshal block data marshal
func (b *BlockData) Marshal() []byte {
	bytes, _ := proto.Marshal(b)
	return bytes
}

// PreviousHash returns the previous hash of the block
func (b *Block) PreviousHash() crypto.Hash {
	return crypto.HexToHash(b.Header.PreviousHash)
}

// Hash returns the hash of the blockheader in block
func (b *Block) Hash() crypto.Hash {
	if hash := b.hash.Load(); hash != nil {
		return hash.(crypto.Hash)
	}
	v := b.Header.Hash()
	b.hash.Store(v)
	return v
}

// Serialize serializes the all data in block
func (b *Block) Serialize() []byte {
	return b.Marshal()
}

//GetTransactions get Transactions by Type
func (b *Block) GetTransactions(txType TransactionType) (Transactions, error) {
	var txs Transactions
	if txType < 0 || txType > 5 {
		return nil, errors.New("transaction type is not support")
	}
	for _, td := range b.TxDatas {
		if td.GetHeader().GetType() == txType {
			txs = append(txs, &Transaction{TxData: *td})
		}
	}
	return txs, nil
}

// Deserialize deserializes bytes to Block
func (b *Block) Deserialize(data []byte) error {
	bInfo := &BlockData{}
	err := proto.Unmarshal(data, bInfo)
	if err != nil {
		return err
	}
	b.BlockData = *bInfo
	return nil
}
