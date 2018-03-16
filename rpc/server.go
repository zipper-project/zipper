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

package rpc

import (
	"io"
	"net"

	"github.com/zipper-project/zipper/blockchain"
	"github.com/zipper-project/zipper/common/log"
	"github.com/zipper-project/zipper/common/rpc"
	"github.com/zipper-project/zipper/config"
)

type HttpConn struct {
	in  io.Reader
	out io.Writer
}

func NewHttConn(in io.Reader, out io.Writer) *HttpConn {
	return &HttpConn{
		in:  in,
		out: out,
	}
}

func (c *HttpConn) Read(p []byte) (n int, err error)  { return c.in.Read(p) }
func (c *HttpConn) Write(d []byte) (n int, err error) { return c.out.Write(d) }
func (c *HttpConn) Close() error                      { return nil }

func NewServer(bc *blockchain.Blockchain) *rpc.Server {
	server := rpc.NewServer()

	server.Register(NewRPCTransaction(bc))
	server.Register(NewRPCLedger(bc))
	return server
}

// StartServer with Test instance as a service
func StartServer(server *rpc.Server, option *config.RPCOption) {
	if option.Enabled == false {
		return
	}
	var (
		listener net.Listener
		err      error
	)
	if listener, err = net.Listen("tcp", ":"+option.Port); err != nil {
		log.Errorf("RPC Server error %+v", err)
	}
	rpc.NewHTTPServer(server, []string{"*"}).Serve(listener)
}
