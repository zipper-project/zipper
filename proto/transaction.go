// Copyright (C) 2017, Zipper Team.  All rights reserved.
//
// This file is part of zipper
//
// The zipper is free software: you can use, copy, modify,
// and distribute this software for any purpose with or
// without fee is hereby granted, provided that the above
// copyright notice and this permission notice appear in all copies.
//
// The zipper is distributed in the hope that it will be useful,
// but WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
// ISC License for more details.
//
// You should have received a copy of the ISC License
// along with this program.  If not, see <https://opensource.org/licenses/isc>.

package proto

import (
	"errors"

	"github.com/golang/protobuf/proto"
	"github.com/zipper-project/zipper/common/crypto"
)

var (
	// ErrEmptySignature represents no signature
	ErrEmptySignature = errors.New("Signature Empty Error")
)

func NewTxHeader(version, createtime, nonce uint32, txType TransactionType) *TxHeader {
	txHeader := &TxHeader{
		Version:    version,
		CreateTime: createtime,
		Nonce:      nonce,
		Type:       txType,
	}
	return txHeader
}

func NewTransaction(txHeader *TxHeader, inputs []*TxIn, outputs []*TxOut, contractSpec *ContractSpec, payload []byte) *Transaction {
	tx := &Transaction{
		Header:       txHeader,
		Inputs:       inputs,
		Outputs:      outputs,
		ContractSpec: contractSpec,
		Payload:      payload,
	}
	return tx
}

// Hash returns the hash of a transaction
func (tx *Transaction) Hash() crypto.Hash {
	return crypto.DoubleSha256(tx.Serialize())
}

// SignHash returns the hash of a raw transaction before sign
func (tx *Transaction) SignHash() crypto.Hash {
	inputs := make([]*TxIn, 0)
	for _, input := range tx.Inputs {
		inputs = append(inputs, &TxIn{
			PreviousOutPoint: input.PreviousOutPoint,
			TxWitness:        nil,
		})
	}
	rawTx := NewTransaction(tx.Header, inputs, tx.Outputs, tx.ContractSpec, tx.Payload)
	return rawTx.Hash()
}

// Serialize marshal txData proto message
func (tx *Transaction) Serialize() []byte {
	bytes, _ := proto.Marshal(tx)
	return bytes
}

// Deserialize deserializes bytes to a transaction
func (tx *Transaction) Deserialize(data []byte) error {
	return proto.Unmarshal(data, tx)
}

// Transactions represents transaction slice type for basic sorting.
type Transactions []*Transaction

func MerkleRootHash(txs Transactions) crypto.Hash {
	if len(txs) > 0 {
		hashs := make([]crypto.Hash, 0)
		for _, tx := range txs {
			hashs = append(hashs, tx.Hash())
		}
		return crypto.ComputeMerkleHash(hashs)[0]
	}
	return crypto.Hash{}
}
