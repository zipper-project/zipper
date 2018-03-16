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

import "encoding/json"

const (
	ContractCodeKey = "__CONTRACT_CODE_KEY__"
)

var VMConf *Config

// Config vm config struct
type Config struct {
	LogFile                    string
	LogLevel                   string
	VMRegistrySize             int
	VMCallStackSize            int
	VMMaxMem                   int // vm maximum memory size (MB)
	ExecLimitStackDepth        int
	ExecLimitMaxOpcodeCount    int // maximum allow execute opcode count
	ExecLimitMaxRunTime        int // the contract maximum run time (millisecond)
	ExecLimitMaxScriptSize     int // contract script(lua source code) maximum size (byte)
	ExecLimitMaxStateValueSize int // the max state value size (byte)
	ExecLimitMaxStateItemCount int // the max state count in one contract
	ExecLimitMaxStateKeyLength int // max state key length
	LuaVMExeFilePath           string
	JSVMExeFilePath            string
	BsWorkerCnt                int
	JsWorkerCnt                int
	LuaWorkerCnt               int
}

// DefaultConfig default vm config
func DefaultConfig() *Config {
	return &Config{
		VMRegistrySize:             256,
		VMCallStackSize:            64,
		VMMaxMem:                   200,
		ExecLimitStackDepth:        100,
		ExecLimitMaxOpcodeCount:    10000000000000,
		ExecLimitMaxRunTime:        2000,
		ExecLimitMaxScriptSize:     10240, //5K
		ExecLimitMaxStateValueSize: 10240, //5K
		ExecLimitMaxStateItemCount: 10000000000,
		ExecLimitMaxStateKeyLength: 256,
		LuaVMExeFilePath:           "bin/luavm",
		JSVMExeFilePath:            "bin/jsvm",
		BsWorkerCnt:                1,
	}
}

func (c *Config) String() string {
	data, _ := json.Marshal(c)
	return string(data)
}

func (c *Config) SetString(src string) error {
	return json.Unmarshal([]byte(src), c)
}
