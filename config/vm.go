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

package config

import (
	"github.com/zipper-project/zipper/vm"
)

//VMConfig returns vm configuration
func VMConfig(logFile, logLevel string) *vm.Config {
	var config = vm.DefaultConfig()
	config.LogFile = logFile
	config.LogLevel = logLevel
	config.VMRegistrySize = getInt("vm.registrySize", config.VMRegistrySize)
	config.VMCallStackSize = getInt("vm.callStackSize", config.VMCallStackSize)
	config.VMMaxMem = getInt("vm.maxMem", config.VMMaxMem)
	config.ExecLimitStackDepth = getInt("vm.execLimitStackDepth", config.ExecLimitStackDepth)
	config.ExecLimitMaxOpcodeCount = getInt("vm.execLimitMaxOpcodeCount", config.ExecLimitMaxOpcodeCount)
	config.ExecLimitMaxRunTime = getInt("vm.execLimitMaxRunTime", config.ExecLimitMaxRunTime)
	config.ExecLimitMaxScriptSize = getInt("vm.execLimitMaxScriptSize", config.ExecLimitMaxScriptSize)
	config.ExecLimitMaxStateValueSize = getInt("vm.execLimitMaxStateValueSize", config.ExecLimitMaxStateValueSize)
	config.ExecLimitMaxStateItemCount = getInt("vm.execLimitMaxStateItemCount", config.ExecLimitMaxStateItemCount)
	config.ExecLimitMaxStateKeyLength = getInt("vm.execLimitMaxStateKeyLength", config.ExecLimitMaxStateKeyLength)
	config.LuaVMExeFilePath = getString("vm.luaVMExeFilePath", config.LuaVMExeFilePath)
	config.JSVMExeFilePath = getString("vm.jsVMExeFilePath", config.JSVMExeFilePath)
	config.BsWorkerCnt = getInt("vm.BsWorkerCnt", config.BsWorkerCnt)
	return config
}
