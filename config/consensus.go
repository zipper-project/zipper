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
	"github.com/zipper-project/zipper/common/crypto"
	"github.com/zipper-project/zipper/common/utils"
	"github.com/zipper-project/zipper/consensus/consenter"
	"github.com/zipper-project/zipper/consensus/noops"
	"github.com/zipper-project/zipper/consensus/scip"
)

func ConsenterOptions() *consenter.Options {
	option := consenter.NewDefaultOptions()
	option.Plugin = getString("consensus.plugin", option.Plugin)
	option.Noops = NoopsOptions()
	option.Scip = SCIPOptions()
	return option
}

func NoopsOptions() *noops.Options {
	option := noops.NewDefaultOptions()
	option.BatchSize = getInt("consensus.noops.batchSize", option.BatchSize)
	option.BatchTimeout = getDuration("consensus.noops.batchTimeout", option.BatchTimeout)
	option.BlockSize = getInt("consensus.noops.blockSize", option.BlockSize)
	option.BlockTimeout = getDuration("consensus.noops.blockTimeout", option.BlockTimeout)
	return option
}

func SCIPOptions() *scip.Options {
	option := scip.NewDefaultOptions()
	option.Chain = getString("blockchain.chainId", option.Chain)
	option.ID = option.Chain + ":" + utils.BytesToHex(crypto.Ripemd160([]byte(getString("blockchain.nodeId", option.ID)+option.Chain)))
	option.N = getInt("consensus.scip.N", option.N)
	option.Q = getInt("consensus.scip.Q", option.Q)
	option.K = getInt("consensus.scip.K", option.K)
	option.BatchSize = getInt("consensus.scip.batchSize", option.BatchSize)
	option.BatchTimeout = getDuration("consensus.scip.batchTimeout", option.BatchTimeout)
	option.BlockSize = getInt("consensus.scip.blockSize", option.BlockSize)
	option.BlockTimeout = getDuration("consensus.scip.blockTimeout", option.BlockTimeout)
	option.Request = getDuration("consensus.scip.request", option.Request)
	option.ViewChange = getDuration("consensus.scip.viewChange", option.ViewChange)
	option.ResendViewChange = getDuration("consensus.scip.resendViewChange", option.ViewChange)
	option.ViewChangePeriod = getDuration("consensus.scip.viewChangePeriod", option.ViewChangePeriod)
	return option
}
