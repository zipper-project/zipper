package proto

import (
	"bytes"
	"encoding/hex"
	"fmt"
	"math/big"
	"testing"
	"time"

	"github.com/zipper-project/zipper/account"
	"github.com/zipper-project/zipper/common/crypto"
	"github.com/zipper-project/zipper/common/utils"
)

var (
	testHashStr string
)

func TestBlockSerialize(t *testing.T) {
	var (
		testBlock = Block{
			BlockData: BlockData{
				Header: &BlockHeader{
					PreviousHash: crypto.DoubleSha256([]byte("xxxx")).String(),
					TimeStamp:    uint32(time.Now().Unix()),
					Nonce:        uint32(100),
				},
			},
		}
	)
	Txs := make([]*TxData, 0)
	hashs := make([]crypto.Hash, 0)
	reciepent := account.HexToAddress("0xbf6080eaae18a6eb4d9d3b9ef08a8bdf02e3caa8")
	for i := 1; i < 3; i++ {

		tx := &TxData{
			Header: &TxHeader{
				FromChain:  account.NewChainCoordinate([]byte{byte(i)}),
				ToChain:    account.NewChainCoordinate([]byte{byte(i)}),
				Type:       TransactionType_Atomic,
				Nonce:      1,
				Sender:     reciepent.String(),
				Recipient:  reciepent.String(),
				AssetID:    1,
				Amount:     big.NewInt(1100).Int64(),
				Fee:        big.NewInt(110).Int64(),
				CreateTime: utils.CurrentTimestamp(),
			},
		}
		Txs = append(Txs, tx)
		ttx := &Transaction{TxData: *tx}
		hashs = append(hashs, ttx.Hash())
	}

	testBlock.TxDatas = Txs
	testBlock.Header.TxsMerkleHash = crypto.ComputeMerkleHash(hashs)[0].String()

	fmt.Println("Block hash", testBlock.Hash())
	fmt.Printf("Block Raw {'previousHash':%v,\n 'MerkleHash':%v,\n  'Nonce':%v,\n TimeStamp':%v,\n Txs:%v.\n",
		testBlock.Header.PreviousHash,
		testBlock.Header.TxsMerkleHash,
		testBlock.Header.Nonce,
		testBlock.Header.TimeStamp,
		testBlock.TxDatas,
	)
	fmt.Println("Block Header serialize()", testBlock.Header.Serialize())
	fmt.Println("Block AtomicTxs", testBlock.TxDatas, len(testBlock.TxDatas))
	fmt.Println("Block serialize()", hex.EncodeToString(testBlock.Serialize()))
	testHashStr = hex.EncodeToString(testBlock.Serialize())
}

func TestBlockDeserialize(t *testing.T) {
	testBlock := &Block{}
	data, _ := hex.DecodeString(testHashStr)
	fmt.Println("------------block deseriailze---------")
	testBlock.Deserialize(data)
	blkData := testBlock.Serialize()
	if !bytes.Equal(blkData, data) {
		t.Errorf("Block.Serialize error, %0x != %0x ", data, blkData)
	}
}
