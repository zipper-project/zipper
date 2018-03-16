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
	"github.com/zipper-project/zipper/blockchain/validator"
)

func ValidatorConfig() *validator.Config {
	var config = validator.DefaultConfig()

	config.IsValid = getbool("validator.status", config.IsValid)
	config.BlacklistDur = getDuration("validator.blacklisttimeout", config.BlacklistDur)
	config.TxPoolCapacity = getInt("validator.txpool.capacity", config.TxPoolCapacity)
	config.TxPoolTimeOut = getDuration("validator.txpool.timeout", config.TxPoolTimeOut)
	config.TxPoolDelay = getInt("validator.txpool.txdelay", config.TxPoolDelay)
	config.TxPoolDelay = getInt("validator.txpool.txdelay", config.TxPoolDelay)
	config.PublicAddresses = getStringSlice("validator.issueaddr", []string{})
	return config
}
