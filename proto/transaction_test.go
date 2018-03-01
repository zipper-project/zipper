package proto

import (
	"bytes"
	"fmt"
	"math/big"
	"testing"

	"github.com/zipper-project/zipper/account"
	"github.com/zipper-project/zipper/common/crypto"
	"github.com/zipper-project/zipper/common/utils"
)

var testTx = getTestTransaction()
var testTxHex = fmt.Sprintf("%0x", testTx.Serialize())

func getTestTransaction() *Transaction {
	sender := account.HexToAddress("0xc9bc867a613381f35b4430a6cb712eff8bb50311")
	address := account.HexToAddress("0xc9bc867a613381f35b4430a6cb712eff8bb50310")
	fromChain := account.NewChainCoordinate([]byte{0, 1, 3})
	toChain := account.NewChainCoordinate([]byte{0, 1, 1})
	nonce := uint32(10000)

	tx := &Transaction{
		TxData: TxData{
			Header: &TxHeader{
				FromChain:  fromChain,
				ToChain:    toChain,
				Type:       TransactionType_Atomic,
				Nonce:      nonce,
				Sender:     sender.String(),
				Recipient:  address.String(),
				AssetID:    1,
				Amount:     big.NewInt(1100).Int64(),
				Fee:        big.NewInt(110).Int64(),
				CreateTime: utils.CurrentTimestamp(),
			},
		},
	}
	tx.Payload = []byte("123456")
	return tx
}

func TestTxDeserialize(t *testing.T) {
	txBytes := utils.HexToBytes(testTxHex)
	tx := new(Transaction)
	tx.Deserialize(txBytes)
	if !bytes.Equal(tx.Serialize(), txBytes) {
		t.Errorf("Tx Deserialize error! %v != %v", tx.Serialize(), txBytes)
	}
}

func TestTxSender(t *testing.T) {
	var (
		priv, _ = crypto.GenerateKey()
		addr    = account.PublicKeyToAddress(*priv.Public())
	)

	tx := &Transaction{
		TxData: TxData{
			Header: &TxHeader{
				FromChain:  nil,
				ToChain:    nil,
				Type:       TransactionType_Atomic,
				Nonce:      1,
				Sender:     addr.String(),
				Recipient:  addr.String(),
				AssetID:    1,
				Amount:     big.NewInt(1100).Int64(),
				Fee:        big.NewInt(110).Int64(),
				CreateTime: utils.CurrentTimestamp(),
			},
		},
	}

	sig, _ := priv.Sign(tx.SignHash().Bytes())
	tx.WithSignature(sig)

	tx2 := new(Transaction)
	tx2.Deserialize(tx.Serialize())
	if !bytes.Equal(tx.Serialize(), tx2.Serialize()) {
		t.Errorf("Deserialize error with Signature, %0x != %0x", tx.Serialize(), tx2.Serialize())
	}
}
