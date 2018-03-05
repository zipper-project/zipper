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
package contract

import (
	"github.com/zipper-project/zipper/account"
	"github.com/zipper-project/zipper/vm"
)

var (
	// DefaultAdminAddr is the default value of admin address.
	DefaultAdminAddr = account.Address{
		0x29, 0x76, 0x3b, 0xb3, 0x68, 0xf2, 0xd4, 0xf6, 0x24, 0x16,
		0xa1, 0xd7, 0xa8, 0x2d, 0x16, 0x88, 0x5c, 0x20, 0x6a, 0x36,
	}

	// DefaultGlobalContract is the default value of global contract.
	DefaultGlobalContract = vm.ContractCode{
		Type: "luavm",
		Code: []byte(
			`--[[
			global 合约。
			--]]
			
			local L0 = require("L0")
			
			function L0Init(args)
				return true
			end
			
			function L0Invoke(funcName, args)
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
					L0.SetGlobalState(key, value)
					return true
				elseif funcName == "DelGlobalState" then
					L0.DelGlobalState(key)
					return true
				end
				return false
			end
			
			function L0Query(args)
				if type(args) ~= "table" then
					return ""
				end
			
				local key = args[0]
				if type(key) ~= "string" then
					return ""
				end
			
				return L0.GetGlobalState(key)
			end`),
	}
)
