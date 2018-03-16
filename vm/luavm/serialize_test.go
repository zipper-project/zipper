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

package luavm

import (
	"testing"
	"bytes"
	lua "github.com/yuin/gopher-lua"
)

func TestLValueConvert(t *testing.T) {
	ls := lua.LString("hello")
	data := lvalueToByte(ls)
	buf := bytes.NewBuffer(data)
	v, err := byteToLValue(buf)
	if err != nil || string(v.(lua.LString)) != "hello" {
		t.Error("convert string error")
	}

	lb := lua.LBool(true)
	data = lvalueToByte(lb)
	buf = bytes.NewBuffer(data)
	v, err = byteToLValue(buf)
	if err != nil || bool(v.(lua.LBool)) != true {
		t.Error("convert bool error")
	}

	ln := lua.LNumber(float64(123456789.4321))
	data = lvalueToByte(ln)
	buf = bytes.NewBuffer(data)
	v, err = byteToLValue(buf)
	if err != nil || float64(v.(lua.LNumber)) != 123456789.4321 {
		t.Error("convert number error")
	}

	ltb := new(lua.LTable)
	ltb.RawSetString("str", ls)
	ltb.RawSetInt(1, lb)
	ltb.RawSet(lb, lb)

	lctb := new(lua.LTable)
	ltb.RawSetInt(10, ls)

	ltb.RawSet(lctb, lctb)

	data = lvalueToByte(ltb)
	buf = bytes.NewBuffer(data)
	v, err = byteToLValue(buf)
	if err != nil {
		t.Error("convert table error")
	}
	ntb := v.(*lua.LTable)
	ntb.ForEach(func(key lua.LValue, value lua.LValue) {
		t.Log("key： ", key, "value：", value)

	})
}
