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

package noops

import (
	"time"
)

// NewDefaultOptions Create noops options with default
func NewDefaultOptions() *Options {
	options := &Options{}
	options.BatchSize = 2000
	options.BatchTimeout = 10 * time.Second
	options.BlockSize = 2000
	options.BlockTimeout = 10 * time.Second
	options.BufferSize = 100
	return options
}

// Options Define noops options
type Options struct {
	BatchSize    int
	BatchTimeout time.Duration
	BlockSize    int
	BlockTimeout time.Duration
	BufferSize   int
}
