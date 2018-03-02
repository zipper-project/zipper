package utils
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
