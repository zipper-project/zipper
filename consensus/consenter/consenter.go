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
