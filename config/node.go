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

import "path/filepath"

type NodeOption struct {
	// log
	LogLevel string
	LogFile  string
	LogName  string

	// profile
	CPUFile  string
	ProfPort string

	DataDir string
}

func NewDefaultNodeOption() *NodeOption {
	return &NodeOption{
		LogLevel: "debug",
		LogFile:  "zipper.log",
	}
}

func NewNodeOption() *NodeOption {
	option := NewDefaultNodeOption()
	option.ProfPort = getString("blockchain.profPort", option.ProfPort)
	option.CPUFile = getString("blockchain.cpuprofile", option.CPUFile)
	option.LogLevel = getString("log.level", option.LogLevel)
	option.DataDir = getString("log.datadir", option.DataDir)
	option.LogName = getString("log.logdirname", option.LogName)
	option.LogFile = filepath.Join(option.DataDir, option.LogName, option.LogFile)
	return option
}
