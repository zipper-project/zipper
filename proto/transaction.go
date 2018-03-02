package proto

import (
	"errors"
	"math/big"
	"strings"
	"sync/atomic"

	"github.com/golang/protobuf/proto"
	"github.com/zipper-project/zipper/account"
	"github.com/zipper-project/zipper/common/crypto"
	"github.com/zipper-project/zipper/common/utils"
)

var (
	// ErrEmptySignature represents no signature
	ErrEmptySignature = errors.New("Signature Empty Error")
)

//Transaction
type Transaction struct {
	TxData
	hash   atomic.Value
	sender atomic.Value
}

//Balance
type Balance struct {
	ID        uint32
	Sender    *big.Int
	Recipient *big.Int
	Callback  func(interface{})
}

// Hash returns the hash of a transaction
func (tx *Transaction) Hash() crypto.Hash {
	if hash := tx.hash.Load(); hash != nil {
		//log.Debugf("Tx Hash %s", hash.(crypto.Hash))
		return hash.(crypto.Hash)
	}
	v := crypto.DoubleSha256(tx.Serialize())
	tx.hash.Store(v)
	//log.Debugf("Tx Hash %s", v)
	return v
}

// SignHash returns the hash of a raw transaction before sign
func (tx *Transaction) SignHash() crypto.Hash {
	rawTx := &Transaction{
		TxData: TxData{
			Header: &TxHeader{
				FromChain:  tx.Header.FromChain,
				ToChain:    tx.Header.ToChain,
				Type:       tx.Header.Type,
				Nonce:      tx.Header.Nonce,
				Sender:     tx.Header.Sender,
				Recipient:  tx.Header.Recipient,
				AssetID:    tx.Header.AssetID,
				Amount:     tx.Header.Amount,
				Fee:        tx.Header.Fee,
				CreateTime: tx.Header.CreateTime,
			},
			Payload:      tx.Payload,
			Meta:         tx.Meta,
			ContractSpec: tx.ContractSpec,
		},
	}
	return rawTx.Hash()
}

// Marshal marshal txData proto message
func (tx *TxData) Marshal() []byte {
	bytes, _ := proto.Marshal(tx)
	return bytes
}

// Serialize returns the serialized bytes of a transaction
func (tx *Transaction) Serialize() []byte {
	return tx.Marshal()
}

// Deserialize deserializes bytes to a transaction
func (tx *Transaction) Deserialize(data []byte) error {
	txData := &TxData{}
	err := proto.Unmarshal(data, txData)
	if err != nil {
		return err
	}
	tx.TxData = *txData
	return nil
}

// Verfiy Also can use this method verify signature
func (tx *Transaction) Verfiy() (account.Address, error) {
	var (
		a   account.Address
		err error
	)
	switch tx.Header.GetType() {
	case TransactionType_Atomic, TransactionType_AcrossChain, TransactionType_Backfront, TransactionType_Distribut, TransactionType_IssueUpdate,
		TransactionType_JSContractInit, TransactionType_LuaContractInit, TransactionType_ContractInvoke, TransactionType_ContractQuery, TransactionType_Security:
		fallthrough
	case TransactionType_Issue:
		if tx.Header.Signature != nil {
			if sender := tx.sender.Load(); sender != nil {
				return sender.(account.Address), nil
			}
			sig := &crypto.Signature{}
			sig.SetBytes(tx.Header.Signature, false)
			p, err := sig.RecoverPublicKey(tx.SignHash().Bytes())
			if err != nil {
				return a, err
			}
			a = account.PublicKeyToAddress(*p)
			tx.sender.Store(a)
		} else {
			err = ErrEmptySignature
		}

	case TransactionType_Merged:
		a = account.ChainCoordinateToAddress(account.HexToChainCoordinate(tx.FromChain()))
	}
	return a, err
}

// Sender returns the address of the sender.
func (tx *Transaction) Sender() account.Address {
	return account.HexToAddress(tx.Header.Sender)
}

// FromChain returns the chain coordinate of the sender
func (tx *Transaction) FromChain() string { return utils.BytesToHex(tx.Header.FromChain) }

// ToChain returns the chain coordinate of the recipient
func (tx *Transaction) ToChain() string { return utils.BytesToHex(tx.Header.ToChain) }

// IsLocalChain returns whether or not local chain
func (tx *Transaction) IsLocalChain() bool { return strings.Compare(tx.FromChain(), tx.ToChain()) == 0 }

// Recipient returns the address of the recipient
func (tx *Transaction) Recipient() account.Address {
	return account.HexToAddress(tx.Header.Recipient)
}

// Amount returns the transfer amount of the transaction
func (tx *Transaction) Amount() int64 { return tx.Header.Amount }

// Fee returns the nonce of the transaction
func (tx *Transaction) Fee() int64 { return tx.Header.Fee }

// AssetID returns the asset id of the transaction
func (tx *Transaction) AssetID() uint32 { return tx.Header.AssetID }

// WithSignature returns a new transaction with the given signature.
func (tx *Transaction) WithSignature(sig *crypto.Signature) {
	//TODO: sender cache
	tx.Header.Signature = sig.Bytes()
}

//WithPayload returns a new transaction with the given data
func (tx *Transaction) WithPayload(data []byte) {
	tx.Payload = data
}

// CreateTime returns the create time of the transaction
func (tx *Transaction) CreateTime() uint32 {
	return tx.Header.CreateTime
}

// Compare implements interface consensus need
func (tx *Transaction) Compare(v interface{}) int {
	if tx.CreateTime() >= v.(*Transaction).CreateTime() {
		return 1
	}
	return 0
}

// GetType returns transaction type
func (tx *Transaction) GetType() TransactionType { return tx.Header.Type }

// Transactions represents transaction slice type for basic sorting.
type Transactions []*Transaction

// Len returns the length of s
func (s Transactions) Len() int { return len(s) }

// Swap swaps the i'th and the j'th element in s
func (s Transactions) Swap(i, j int) { s[i], s[j] = s[j], s[i] }

// Less compares nonce of the i'th and the j'th element in s
func (s Transactions) Less(i, j int) bool { return s[i].Header.Nonce < s[j].Header.Nonce }
