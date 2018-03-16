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
	"github.com/zipper-project/zipper/peer"
)

func ServerOption() *peer.Option {
	option := peer.NewDefaultOption()
	option.ListenAddress = getString("net.listenAddr", option.ListenAddress)
	option.ReconnectInterval = getDuration("net.connectTimeInterval", option.ReconnectInterval)
	option.ReconnectTimes = getInt("net.reconnectTimes", option.ReconnectTimes)
	option.KeepAliveInterval = getDuration("net.keepAliveInterval", option.KeepAliveInterval)
	option.KeepAliveTimes = getInt("net.keepAliveTimes", option.KeepAliveTimes)
	option.MaxPeers = getInt("net.maxPeers", option.MaxPeers)
	option.MinPeers = getInt("net.minPeers", option.MinPeers)
	option.PeerID = []byte(getString("blockchain.nodeId", string(option.PeerID)))
	option.BootstrapNodes = getStringSlice("net.bootstrapNodes", option.BootstrapNodes)
	return option
}
