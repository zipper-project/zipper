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
package state

import (
	"bytes"
	"encoding/json"
	"errors"

	"github.com/zipper-project/zipper/account"
	"github.com/zipper-project/zipper/common/utils"
)

var (
	// DefaultAdminAddr is the default value of admin address.
	DefaultAdminAddr = account.Address{
		0x29, 0x76, 0x3b, 0xb3, 0x68, 0xf2, 0xd4, 0xf6, 0x24, 0x16,
		0xa1, 0xd7, 0xa8, 0x2d, 0x16, 0x88, 0x5c, 0x20, 0x6a, 0x36,
	}

	// DefaultGlobalContract is the default value of global contract.
	DefaultGlobalContractType = "luavm"

	DefaultGlobalContractCode = []byte(
		`--[[
			global 合约。
			--]]

			local ZIP = require("ZIP")

			function Init(args)
				return true
			end

			function Invoke(funcName, args)
				if type(args) ~= "table" then
					return false
				end

				local key = args[0]
				if type(key) ~= "string" then
					return false
				end

				if funcName == "SetGlobalState" then
					local value = args[1]
					if not(value) then
						return false
					end
					ZIP.SetGlobalState(key, value)
					return true
				elseif funcName == "DelGlobalState" then
					ZIP.DelGlobalState(key)
					return true
				end
				return false
			end

			function Query(args)
				if type(args) ~= "table" then
					return ""
				end

				local key = args[0]
				if type(key) ~= "string" then
					return ""
				end

				return ZIP.GetGlobalState(key)
			end`)
)

const (
	stringType = iota
)

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
