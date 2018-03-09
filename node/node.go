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

package node

import (
	"net/http"
	_ "net/http/pprof"
	"os"
	"os/signal"
	"runtime"
	"runtime/pprof"

	"syscall"

	"github.com/zipper-project/zipper/blockchain"
	"github.com/zipper-project/zipper/common/log"
	"github.com/zipper-project/zipper/config"
)

// Node represents the blockchain zipper
type Node struct {
	bc  *blockchain.Blockchain
	cfg *config.Option
}

// NewNode returns node daemon instance
func NewNode(cfgFile string) *Node {
	if err := config.ReadInConfig(cfgFile); err != nil {
		panicf("loadConfig error %v", err)
		return nil
	}

	return &Node{
		bc: blockchain.NewBlockChain(),
	}
}

// Start starts the blockchain service
func (l *Lcnd) Start() {
	runtime.GOMAXPROCS(runtime.NumCPU())

	abort := make(chan os.Signal, 1)
	signal.Notify(abort, os.Interrupt, os.Kill, syscall.SIGHUP, syscall.SIGTERM, syscall.SIGKILL)
	go func() {
		<-abort
		l.bc.Stop()
		os.Exit(0)
	}()

	log.New(l.cfg.LogFile)
	log.SetLevel(l.cfg.LogLevel)
	if l.cfg.ProfPort != "" {
		go func() {
			err := http.ListenAndServe(":"+l.cfg.ProfPort, nil)
			if err != nil {
				log.Errorf("Prof Server start error=%v", err)
			} else {
				log.Debugf("Prof Server start on port=%s", l.cfg.ProfPort)
			}
		}()
	}

	if l.cfg.CPUFile != "" {
		cpuFile := l.cfg.CPUFile
		cpuProfile, _ := os.Create(cpuFile)
		pprof.StartCPUProfile(cpuProfile)

		defer func() {
			memFile := cpuFile + ".mem"
			pprof.StopCPUProfile()
			memProfile, _ := os.Create(memFile)
			pprof.WriteHeapProfile(memProfile)
			memProfile.Close()
			cpuProfile.Close()
		}()
	}

	l.bc.Start()
}
