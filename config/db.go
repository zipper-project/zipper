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

	"github.com/zipper-project/zipper/common/db"
	"github.com/zipper-project/zipper/params"
)

// DBConfig returns configurations for database
func DBConfig() *db.Config {
	dbConfig := db.DefaultConfig()
	dbConfig.DbPath = filepath.Join(params.DataDir, params.ChainDataDirName)
	dbConfig.Columnfamilies = getStringSlice("db.columnfamilies", dbConfig.Columnfamilies)
	dbConfig.KeepLogFileNumber = getInt("db.keepLogFileNumber", dbConfig.KeepLogFileNumber)
	dbConfig.MaxLogFileSize = getInt("db.maxLogFileSize", dbConfig.MaxLogFileSize)
	dbConfig.LogLevel = getString("db.loglevel", dbConfig.LogLevel)
	return dbConfig
}
