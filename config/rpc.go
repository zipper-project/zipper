package config

import (
	"github.com/spf13/viper"
)

type RPCOption struct {
	Enabled  bool
	Port     string
	User     string
	PassWord string
}

func NewDefaultRPCOption() *RPCOption {
	option := &RPCOption{
		Enabled: true,
		Port:    "8000",
	}

	return option
}

func RPCConfig() *RPCOption {
	option := NewDefaultRPCOption()
	option.Enabled = viper.GetBool("jrpc.enabled")
	option.Port = getString("jrpc.port", option.Port)
	option.User = getString("jrpc.user", option.User)
	option.PassWord = getString("jrpc.password", option.PassWord)
	return option
}
