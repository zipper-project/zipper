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

package scip

import "time"
import "github.com/zipper-project/zipper/common/utils"
import "encoding/hex"
import "crypto/sha256"

//NewDefaultOptions Create nbft options with default value
func NewDefaultOptions() *Options {
	options := &Options{}
	options.Chain = "0"
	options.ID = "0"
	options.N = 4
	options.Q = 3
	options.K = 20

	options.BatchSize = 2000
	options.BatchTimeout = 10 * time.Second
	options.BlockSize = 2000
	options.BlockTimeout = 10 * time.Second
	options.Request = 20 * time.Second
	options.BufferSize = 100

	options.ViewChange = 2 * time.Second
	options.ResendViewChange = 2 * time.Second
	options.ViewChangePeriod = 300 * time.Second
	return options
}

//Options Define nbft options
type Options struct {
	Chain string
	ID    string
	N     int
	Q     int
	K     int

	BatchSize    int
	BatchTimeout time.Duration
	BlockSize    int
	BlockTimeout time.Duration
	Request      time.Duration
	BufferSize   int

	ViewChange       time.Duration
	ResendViewChange time.Duration
	ViewChangePeriod time.Duration
}

func (this *Options) Hash() string {
	opt := &Options{}
	opt.Chain = this.Chain
	//opt.ID = this.ID
	opt.N = this.N
	opt.Q = this.Q
	opt.K = this.K

	// opt.BatchSize = this.BatchSize
	// opt.BatchTimeout = this.BatchTimeout
	// opt.BlockSize = this.BlockSize
	// opt.BlockTimeout = this.BlockTimeout
	opt.Request = this.Request
	opt.BufferSize = this.BufferSize

	opt.ViewChange = this.ViewChange
	opt.ResendViewChange = this.ResendViewChange
	opt.ViewChangePeriod = this.ViewChangePeriod
	h := sha256.Sum256(utils.Serialize(opt))
	return hex.EncodeToString(h[:])
}
