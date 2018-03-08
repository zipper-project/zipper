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

package validator

import "time"

type Config struct {
	IsValid           bool
	TxPoolCapacity    int
	TxPoolDelay       int
	MaxWorker         int
	MaxQueue          int
	TxPoolTimeOut     time.Duration
	BlacklistDur      time.Duration
	SecurityPluginDir string
}

func DefaultConfig() *Config {
	return &Config{
		IsValid:        true,
		TxPoolCapacity: 200000,
		TxPoolDelay:    5000,
		MaxWorker:      10,
		MaxQueue:       2000,
		TxPoolTimeOut:  30 * time.Minute,
		BlacklistDur:   1 * time.Minute,
	}
}
