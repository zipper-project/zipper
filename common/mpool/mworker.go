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

package mpool

import (
	"time"
	"sync/atomic"
)

type VmWorker interface {
	//called for job, adn returned synchronously
	VmJob(interface{}) (interface{}, error)

	//wait for to execute the next job
	VmReady() bool
}

type VmExtendedWorker interface {
	// when the mechine is opened and closed, the will be implemented.
	VmInitialize()
	VmTerminate()
}

type VmInterruptableWorker interface {
	//called by the client that will be killed this worker
	VmInterruptable()
}

// the vm DefaultWorker
type VmDefaultWorker struct {
	job *func(interface{}) interface{}
}

func (worker *VmDefaultWorker) VmJob(data interface{}) (interface{}, error) {
	return (*worker.job)(data), nil
}

func (worker *VmDefaultWorker) VmReady() bool {
	return true
}

// the external worker
type workerWrapper struct {
	readyChan chan int
	jobChan  chan interface{}
	outputChan chan interface{}
	workerMechine uint32
	worker VmWorker
	jobCnt int
}

func (ww *workerWrapper) Open() {
	if extWorker, ok := ww.worker.(VmExtendedWorker); ok {
		extWorker.VmInitialize()
	}

	ww.readyChan = make(chan int)
	ww.jobChan = make(chan interface{})
	ww.outputChan = make(chan interface{})

	atomic.SwapUint32(&ww.workerMechine, 1)
	go ww.Loop()
}

func (ww *workerWrapper) Close() {
	close(ww.jobChan)
	atomic.SwapUint32(&ww.workerMechine, 0)
}

func (ww *workerWrapper) Loop() {
	// TODO: now wait for the next job to come through sleep
	//thread := rand.Int()
	waitNextJob := func() {
		for !ww.worker.VmReady() {
			if atomic.LoadUint32(&ww.workerMechine) == 0 {
				break
			}

			time.Sleep(5 * time.Millisecond)
		}

		ww.readyChan <- 1
	}

	waitNextJob()
	for data := range ww.jobChan {
		//ww.outputChan <- ww.worker.VmJob(data)
		ww.worker.VmJob(data)
		ww.jobCnt ++
		waitNextJob()
	}
	close(ww.readyChan)
	close(ww.outputChan)
}

func (ww *workerWrapper) Join() {
	for {
		_, readyChan := <-ww.readyChan
		_, outputChan := <-ww.outputChan
		if !readyChan && !outputChan {
			break
		}
	}

	if extWorker, ok  := ww.worker.(VmExtendedWorker); ok {
		extWorker.VmTerminate()
	}
}

func (ww *workerWrapper) Interrupt() {
	if interrupttWorker, ok := ww.worker.(VmInterruptableWorker); ok {
		interrupttWorker.VmInterruptable()
	}
}
