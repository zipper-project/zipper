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

package params

import "github.com/zipper-project/zipper/coordinate"

const (
	// ProtocolName represents the name of the p2p protocol
	ProtocolName = "L0-NETWORK"
	// ProtocolVersion represents the version of the p2p protocol
	ProtocolVersion = "0.0.1"
)

// ChainID  chain ID
var (
	ChainID       = coordinate.NewChainCoordinate([]byte{0, 1, 3})
	PeerID        string
	PublicAddress []string
	ConnNums      int
	LocalIp       string
	Nvp           bool
	Mongodb       bool
	MaxOccurs     int
)
