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

package jsvm

import (
	"bytes"
	"errors"
	"strconv"

	"github.com/robertkrimen/otto"
	"github.com/zipper-project/zipper/common/utils"
	ltyes "github.com/zipper-project/zipper/ledger/balance"
)

const (
	stringType = iota
	boolType
	numberType
	objectType
	nullType
)

func jsvalueToByte(value otto.Value) ([]byte, error) {
	buf := new(bytes.Buffer)
	if value.IsBoolean() {
		buf.WriteByte(boolType)
		v, err := value.ToBoolean()
		if err != nil {
			return nil, err
		}
		if v {
			buf.WriteByte(1)
		} else {
			buf.WriteByte(0)
		}
		return buf.Bytes(), nil
	} else if value.IsNull() {
		buf.WriteByte(nullType)
		return buf.Bytes(), nil
	} else if value.IsNumber() {
		buf.WriteByte(numberType)
		v, err := value.ToFloat()
		if err != nil {
			return nil, err
		}
		buf.Write(utils.Float64ToByte(v))
		return buf.Bytes(), nil
	} else if value.IsString() {
		buf.WriteByte(stringType)
		v, err := value.ToString()
		if err != nil {
			return nil, err
		}

		data := []byte(v)
		lenByte := utils.VarInt(uint64(len(data)))
		buf.Write(lenByte)
		buf.Write(data)
		return buf.Bytes(), nil
	} else if value.IsObject() {
		obj := value.Object()
		keys := obj.Keys()

		buf.WriteByte(objectType)
		buf.Write(utils.VarInt(uint64(len(keys))))

		for _, k := range keys {
			v, _ := obj.Get(k)
			// write key
			keyByte := []byte(k)
			lenByte := utils.VarInt(uint64(len(keyByte)))
			buf.Write(lenByte)
			buf.Write(keyByte)

			valueByte, err := jsvalueToByte(v)
			if err != nil {
				return nil, err
			}
			buf.Write(valueByte)
		}
		return buf.Bytes(), nil
	}

	return nil, nil
}

func byteToJSvalue(buf *bytes.Buffer, ottoVM *otto.Otto) (otto.Value, error) {
	tp, err := buf.ReadByte()
	if err != nil {
		return otto.NullValue(), err
	}

	switch tp {
	case stringType:
		len, err := utils.ReadVarInt(buf)
		if err != nil {
			return otto.NullValue(), err
		}
		data := make([]byte, len)
		buf.Read(data)
		return otto.ToValue(string(data))
	case boolType:
		b, err := buf.ReadByte()
		if err != nil {
			return otto.NullValue(), err
		}
		if b == 1 {
			return otto.ToValue(true)
		}
		return otto.ToValue(false)
	case numberType:
		data := make([]byte, 8)
		if n, err := buf.Read(data); n != 8 || err != nil {
			return otto.NullValue(), errors.New("buf stream error")
		}
		return otto.ToValue(utils.ByteToFloat64(data))
	case nullType:
		return otto.NullValue(), nil
	case objectType:
		count, err := utils.ReadVarInt(buf)
		if err != nil {
			return otto.NullValue(), err
		}

		mp := make(map[string]interface{}, count)
		if err != nil {
			return otto.NullValue(), err
		}
		for i := 0; i < int(count); i++ {
			len, err := utils.ReadVarInt(buf)
			if err != nil {
				return otto.NullValue(), err
			}

			data := make([]byte, len)
			buf.Read(data)
			k := string(data)

			v, err := byteToJSvalue(buf, ottoVM)
			if err != nil {
				return otto.NullValue(), err
			}

			mp[k] = v
		}
		return ottoVM.ToValue(mp)
	}

	return otto.NullValue(), nil
}

//func kvsToJSValue(kvs []*db.KeyValue, ottoVM *otto.Otto) (otto.Value, error) {
//	mp := make(map[string]interface{})
//	for _, v := range kvs {
//		buf := bytes.NewBuffer(v.Value)
//		value, err := byteToJSvalue(buf, ottoVM)
//		if err != nil {
//			return otto.NullValue(), err
//		}
//		mp[string(v.Key)] = value
//	}
//	return ottoVM.ToValue(mp)
//}
//
func objToLValue(balance *ltyes.Balance, ottoVM *otto.Otto) (otto.Value, error) {
	amountsMp := make(map[string]interface{})
	for k, v := range balance.Amounts {
		value, err := ottoVM.ToValue(v)
		if err != nil {
			return otto.NullValue(), err
		}
		amountsMp[strconv.Itoa(int(k))] = value
	}

	amounts, err := ottoVM.ToValue(amountsMp)
	if err != nil {
		return otto.NullValue(), err
	}

	return amounts, nil
}
