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

	// profile
	CPUFile  string
	ProfPort string
}

func NewDefaultNodeOption() *Option {
	return &NodeOption{
		LogLevel: "debug",
		LogFile:  "zipper.log",
	}
}

func NodeOption() *Option {
	option := NewDefaultNodeOption()
	option.ProfPort = getString("blockchain.profPort", option.ProfPort)
	option.CPUFile = getString("blockchain.cpuprofile", option.CPUFile)
	option.LogLevel = getString("log.level", option.LogLevel)
	option.LogFile = filepath.Join(DataDir, LogDirName, option.LogFile)
	return option
}
