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

package utils

import (
	"encoding/json"
	"github.com/yuin/gopher-lua"
)

func ApiDecode() lua.LGFunction {
	return func(L *lua.LState) int {
		str := L.CheckString(1)

		var value interface{}
		err := json.Unmarshal([]byte(str), &value)
		if err != nil {
			L.Push(lua.LNil)
			L.Push(lua.LString(err.Error()))
			return 2
		}
		L.Push(fromJSON(L, value))
		return 1
	}
}

func ApiEncode() lua.LGFunction {
	return func(L *lua.LState) int {
		value := L.CheckAny(1)

		visited := make(map[*lua.LTable]bool)
		data, err := toJSON(value, visited)
		if err != nil {
			L.Push(lua.LNil)
			L.Push(lua.LString(err.Error()))
			return 2
		}
		L.Push(lua.LString(string(data)))
		return 1
	}
}
