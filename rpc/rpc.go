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
package rpc

import (
	"encoding/hex"
	"errors"

	"github.com/zipper-project/zipper/blockchain"
	"github.com/zipper-project/zipper/proto"
)

type RPCTransaction struct {
	bc *blockchain.Blockchain
}

func NewRPCTransaction(bc *blockchain.Blockchain) *RPCTransaction {
	return &RPCTransaction{
		bc: bc,
	}
}

func (rt *RPCTransaction) Broadcast(txHex string, reply *string) error {
	if len(txHex) < 1 {
		return errors.New("Invalid Params: len(txSerializeData) must be >0 ")
	}

	tx := new(proto.Transaction)
	txByte, _ := hex.DecodeString(txHex)
	err := tx.Deserialize(txByte)
	if err != nil {
		return err
	}

	//TODO verfiy transaction

	rt.bc.Relay(tx)
	return nil
}
