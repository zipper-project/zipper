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

package consenter

import (
	"fmt"
	"strings"

	"github.com/zipper-project/zipper/consensus"
	"github.com/zipper-project/zipper/consensus/scip"
	"github.com/zipper-project/zipper/consensus/noops"
)

// NewConsenter Create consenter of plugin
func NewConsenter(option *Options, stack consensus.IStack) (consenter consensus.Consenter) {
	plugin := strings.ToLower(option.Plugin)
	if plugin == "scip" {
		consenter = scip.NewScip(option.Scip, stack)
	} else if plugin == "noops" {
		consenter = noops.NewNoops(option.Noops, stack)
	} else {
		panic(fmt.Sprintf("Unspport consenter of plugin %s", plugin))
	}
	//go consenter.Start()
	return consenter
}

// NewDefaultOptions Create consenter options with default value
func NewDefaultOptions() *Options {
	options := &Options{
		Plugin: "noops",
		Noops:  noops.NewDefaultOptions(),
		Scip:   scip.NewDefaultOptions(),
	}
	return options
}

// Options Define consenter options
type Options struct {
	Noops  *noops.Options
	Scip   *scip.Options
	Plugin string
}
