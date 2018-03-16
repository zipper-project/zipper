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

package vm

import (
	"bytes"
	"encoding/json"
	"errors"
	"sync"

	"github.com/zipper-project/zipper/common/utils"
)

const (
	stringType = iota
)

// For handle the transacton concurrency processing, to ensure the order of the transaction
type TxSync struct {
	workerChans sync.Map
}

var Txsync *TxSync

func NewTxSync(workersCnt int) *TxSync {
	Txsync = &TxSync{}
	for i := 0; i < workersCnt; i++ {
		Txsync.workerChans.Store(i, make(chan struct{}, 1))
	}

	return Txsync
}

func (ts *TxSync) Notify(idx int) {
	value, _ := ts.workerChans.Load(idx)
	value.(chan struct{}) <- struct{}{}
}

func (ts *TxSync) Wait(idx int) {
	value, _ := ts.workerChans.Load(idx)
	<-value.(chan struct{})
}

func DoContractStateData(src []byte) ([]byte, error) {
	if len(src) == 0 {
		return nil, nil
	}

	buf := bytes.NewBuffer(src)
	tp, err := buf.ReadByte()
	if err != nil {
		return nil, err
	}

	switch tp {
	case stringType:
		len, err := utils.ReadVarInt(buf)
		if err != nil {
			return nil, err
		}
		data := make([]byte, len)
		buf.Read(data)
		return data, nil
	default:
		return nil, errors.New("not support states")
	}
}

func ConcrateStateJson(v interface{}) (*bytes.Buffer, error) {
	data, err := json.Marshal(v)
	if err != nil {
		return &bytes.Buffer{}, err
	}

	buf := new(bytes.Buffer)
	buf.WriteByte(stringType)
	lenByte := utils.VarInt(uint64(len(data)))
	buf.Write(lenByte)
	buf.Write(data)

	return buf, nil
}
