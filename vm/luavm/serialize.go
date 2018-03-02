// Copyright (C) 2017, Zipper Team.  All rights reserved.
//
// This file is part of zipper
//
// The zipper is free software: you can redistribute it and/or modify
// it under the terms of the GNU General Public License as published by
// the Free Software Foundation, either version 3 of the License, or
// (at your option) any later version.
//
// The zipper is distributed in the hope that it will be useful,
// but WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
//
// GNU General Public License for more details.
//
// You should have received a copy of the GNU General Public License
// along with this program.  If not, see <http://www.gnu.org/licenses/>.

package luavm

import (
	"bytes"
	"errors"

	lua "github.com/yuin/gopher-lua"
	"github.com/zipper-project/zipper/common/utils"
	ltyes "github.com/zipper-project/zipper/ledger/types"
)


const (
	lstringType = iota
	lboolType
	lnumberType
	ltableTYpe
)

func lvalueToByte(value lua.LValue) []byte {
	buf := new(bytes.Buffer)

	switch value.(type) {
	case lua.LString:
		buf.WriteByte(lstringType)
		data := []byte(value.String())
		lenByte := utils.VarInt(uint64(len(data)))
		buf.Write(lenByte)
		buf.Write(data)
		return buf.Bytes()

	case lua.LBool:
		buf.WriteByte(lboolType)
		bl := bool(value.(lua.LBool))
		if bl {
			buf.WriteByte(1)
		} else {
			buf.WriteByte(0)
		}
		return buf.Bytes()

	case lua.LNumber:
		buf.WriteByte(lnumberType)
		f := float64(value.(lua.LNumber))
		buf.Write(utils.Float64ToByte(f))
		return buf.Bytes()

	case *lua.LTable:
		buf.WriteByte(ltableTYpe)
		tb := value.(*lua.LTable)
		count := tb.ElementCount()
		buf.Write(utils.VarInt(uint64(count)))

		tb.ForEach(func(k lua.LValue, v lua.LValue) {
			buf.Write(lvalueToByte(k))
			buf.Write(lvalueToByte(v))
		})

		return buf.Bytes()
	}

	return nil
}

func byteToLValue(buf *bytes.Buffer) (lua.LValue, error) {
	tp, err := buf.ReadByte()
	if err != nil {
		return nil, err
	}

	switch tp {
	case lstringType:
		len, err := utils.ReadVarInt(buf)
		if err != nil {
			return nil, err
		}
		data := make([]byte, len)
		buf.Read(data)
		return lua.LString(string(data)), nil
	case lboolType:
		bl, err := buf.ReadByte()
		if err != nil {
			return nil, err
		}
		if 1 == bl {
			return lua.LBool(true), nil
		}
		return lua.LBool(false), nil
	case lnumberType:
		data := make([]byte, 8)
		if n, err := buf.Read(data); n != 8 || err != nil {
			return nil, errors.New("buf stream error")
		}
		return lua.LNumber(utils.ByteToFloat64(data)), nil
	case ltableTYpe:
		tb := new(lua.LTable)
		count, err := utils.ReadVarInt(buf)
		if err != nil {
			return nil, err
		}

		for i := 0; i < int(count); i++ {
			k, err := byteToLValue(buf)
			if err != nil {
				return nil, err
			}

			v, err := byteToLValue(buf)
			if err != nil {
				return nil, err
			}

			tb.RawSet(k, v)
		}

		return tb, nil
	}

	return nil, errors.New("not support data type")
}

func objToLValue(balance *ltyes.Balance) lua.LValue {
	//tb := new(lua.LTable)
	amountTb := new(lua.LTable)
	for k, v := range balance.Amounts {
		amountTb.RawSetInt(int(k), lua.LNumber(v))
	}
	//tb.RawSet(lua.LString("Amounts"), amountTb)
	//tb.RawSet(lua.LString("Nonce"), lua.LNumber(balance.Nonce))
	//return tb

	return amountTb
}
