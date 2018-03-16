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
	PublicAddresses     []string
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
		PublicAddresses: []string{},
	}
}
