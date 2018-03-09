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
	option.BootstrapNodes = getStringSlice("net.bootstrapNodes", option.BootstrapNodes)
	option.MaxPeers = getInt("net.maxPeers", option.MaxPeers)
	option.MinPeers = getInt("net.minPeers", option.MinPeers)
	return option
}
