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
	"path/filepath"
	"time"

	"github.com/spf13/viper"
	"github.com/zipper-project/zipper/common/utils"
)

var (
	DataDir          = "."
	LogDirName       = "logs"
	ChainDataDirName = "chaindata"
	PluginDirname    = "plugin"
)

//ReadInConfig
func ReadInConfig(cfgFile string) (err error) {
	viper.SetConfigFile(cfgFile)
	if err := viper.ReadInConfig(); err != nil {
		return err
	}
	DataDir = getString("blockchain.datadir", DataDir)
	utils.OpenDir(filepath.Join(DataDir, ChainDataDirName))
	utils.OpenDir(filepath.Join(DataDir, LogDirName))
	utils.OpenDir(filepath.Join(DataDir, PluginDirname))
	return nil
}

func getInt(key string, defaultValue int) int {
	var (
		value int
	)
	if value = viper.GetInt(key); value == 0 {
		return defaultValue
	}
	return value
}

func getString(key string, defaultValue string) string {
	var (
		value string
	)
	if value = viper.GetString(key); value == "" {
		return defaultValue
	}
	return value
}

func getStringSlice(key string, defaultValue []string) []string {
	var (
		value []string
	)
	if value = viper.GetStringSlice(key); len(value) == 0 {
		return defaultValue
	}
	return value
}

func getDuration(key string, defaultValue time.Duration) time.Duration {
	var (
		value string
	)
	if value = viper.GetString(key); value == "" {
		return defaultValue
	}
	if duration, err := time.ParseDuration(value); err == nil {
		return duration
	}
	return defaultValue
}

func getbool(key string, defaultValue bool) bool {
	var (
		value bool
	)
	if value = viper.GetBool(key); value == false {
		return defaultValue
	}
	return value
}
