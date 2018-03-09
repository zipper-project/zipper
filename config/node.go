package config

import "path/filepath"

type Option struct {
	// log
	LogLevel string
	LogFile  string

	// profile
	CPUFile  string
	ProfPort string
}

func NewDefaultOption() *Option {
	return &Option{
		LogLevel: "debug",
		LogFile:  "node.log",
	}
}

func NodeOption() *Option {
	option := NewDefaultOption()
	option.ProfPort = getString("blockchain.profPort", option.ProfPort)
	option.CPUFile = getString("blockchain.cpuprofile", option.CPUFile)
	option.LogLevel = getString("log.level", option.LogLevel)
	option.LogFile = filepath.Join(DataDir, LogDirName, option.LogFile)
	return option
}
