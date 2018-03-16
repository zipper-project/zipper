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

package bsvm

import (
	"encoding/json"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/zipper-project/zipper/common/log"
	"github.com/zipper-project/zipper/ledger/state"
	"github.com/zipper-project/zipper/params"
	"github.com/zipper-project/zipper/proto"
	"github.com/zipper-project/zipper/vm"
	"github.com/zipper-project/zipper/vm/jsvm"
	"github.com/zipper-project/zipper/vm/luavm"
)

type WorkerInfo struct {
	allTxsCnt    int
	redoTxsCnt   int
	workerTxCnt  int
	allExecTime  time.Duration
	allWaitTime  time.Duration
	allMergeTime time.Duration
}

type BsWorker struct {
	isCanRedo bool
	workerID  int

	workerInfo *WorkerInfo
	jsWorker   *jsvm.JsWorker
	luaWorker  *luavm.LuaWorker
}

func NewBsWorker(conf *vm.Config, idx int) *BsWorker {
	bsWorker := &BsWorker{
		workerID:   idx,
		workerInfo: &WorkerInfo{},
		jsWorker:   jsvm.NewJsWorker(conf),
		luaWorker:  luavm.NewLuaWorker(conf),
	}

	return bsWorker
}

func (worker *BsWorker) FetchContractType(workerProcWithCallback *vm.WorkerProcWithCallback) string {
	var err error
	txType := "unknown"

	if workerProcWithCallback.WorkProc.ContractData.Transaction.GetHeader().GetType() == proto.TransactionType_ContractInvoke {
		txType, err = worker.GetInvokeType(workerProcWithCallback)
		if err != nil {
			log.Errorf("ThreadId: %+v, can't execute contract, tx_hash: %s, tx_idx: %+v, err_msg: %+v, can Redo: %+v", worker.workerID,
				workerProcWithCallback.WorkProc.ContractData.Transaction.Hash().String(), workerProcWithCallback.Idx, err.Error(), worker.isCanRedo)
		}
	} else {
		txType = worker.GetInitType(workerProcWithCallback)
	}

	return txType
}

func (worker *BsWorker) ExecJob(workerProcWithCallback *vm.WorkerProcWithCallback) error {
	var res interface{}
	var err error

	execTime := time.Now()
	if worker.isCommonTransaction(workerProcWithCallback) {
		err = worker.ExecCommonTransaction(workerProcWithCallback)
	} else {
		txType := worker.FetchContractType(workerProcWithCallback)
		if strings.Contains(txType, "lua") {
			res, err = worker.luaWorker.VmJob(workerProcWithCallback)
		} else if strings.Contains(txType, "js") {
			res, err = worker.jsWorker.VmJob(workerProcWithCallback)
		} else {
			log.Errorf("can't find tx type: %+v, %+v",
				workerProcWithCallback.WorkProc.ContractData.Transaction.Hash().String(),
				workerProcWithCallback.WorkProc.ContractData.Transaction.GetHeader().GetType())
			err = errors.New("find contract type fail ...")
		}
	}

	waitTime := time.Now()

	if workerProcWithCallback.Idx != 0 {
		if !worker.isCanRedo {
			vm.Txsync.Wait(workerProcWithCallback.Idx % vm.VMConf.BsWorkerCnt)
		}
	}

	mergeTime := time.Now()
	res = res
	cerr := workerProcWithCallback.WorkProc.SCHandler.CallBack(&state.CallBackResponse{
		IsCanRedo: !worker.isCanRedo,
		Err:       err,
		//Result: res.(string),
	})

	nowTime := time.Now()

	rexecTime := waitTime.Sub(execTime)
	rwaitTime := mergeTime.Sub(waitTime)
	rmergeTime := nowTime.Sub(mergeTime)

	worker.workerInfo.workerTxCnt++
	worker.workerInfo.allExecTime += rexecTime
	worker.workerInfo.allWaitTime += rwaitTime
	worker.workerInfo.allMergeTime += rmergeTime

	if worker.workerInfo.workerTxCnt%1000 == 0 {
		averExec := worker.workerInfo.allExecTime / time.Duration(worker.workerInfo.workerTxCnt)
		waitExec := worker.workerInfo.allWaitTime / time.Duration(worker.workerInfo.workerTxCnt)
		mergeExec := worker.workerInfo.allMergeTime / time.Duration(worker.workerInfo.workerTxCnt)

		log.Infof("worker id: %d, execTime: %s, waitTime: %s, mergeTime: %s", worker.workerID, averExec, waitExec, mergeExec)
	}

	return cerr
}

func (worker *BsWorker) VmJob(data interface{}) (interface{}, error) {
	workerProcWithCallback := data.(*vm.WorkerProcWithCallback)
	log.Debugf("worker thread id: %+v, start tx: %+v, tx_idx: %+v", worker.workerID, workerProcWithCallback.WorkProc.ContractData.Transaction.Hash().String(), workerProcWithCallback.Idx)
	defer log.Debugf("worker thread id: %+v, finish tx: %+v, tx_idx: %+v", worker.workerID, workerProcWithCallback.WorkProc.ContractData.Transaction.Hash().String(), workerProcWithCallback.Idx)
	worker.isCanRedo = false
	err := worker.ExecJob(workerProcWithCallback)

	if err != nil && !worker.isCanRedo {
		log.Errorf("worker thread id: %+v, to tx redo, tx_hash: %+v, tx_idx: %+v, cause: %+v",
			worker.workerID, workerProcWithCallback.WorkProc.ContractData.Transaction.Hash().String(), workerProcWithCallback.Idx, err)
		worker.isCanRedo = true
		err := worker.ExecJob(workerProcWithCallback)
		if err != nil {
			log.Errorf("worker thread id: %+v, tx redo failed, tx_hash: %+v, tx_idx: %+v, cause: %+v",
				worker.workerID, workerProcWithCallback.WorkProc.ContractData.Transaction.Hash().String(), workerProcWithCallback.Idx, err)
		}
	}

	vm.Txsync.Notify((workerProcWithCallback.Idx + 1) % vm.VMConf.BsWorkerCnt)
	return nil, nil
}

func (worker *BsWorker) VmReady() bool {
	return true
}

func (worker *BsWorker) VmInitialize() {
	//pass
}

func (worker *BsWorker) VmTerminate() {
	//pass
}

func (worker *BsWorker) GetInvokeType(wpwc *vm.WorkerProcWithCallback) (string, error) {
	var err error
	cc := new(vm.ContractCode)
	var code []byte
	if len(wpwc.WorkProc.ContractData.ContractAddr) == 0 {
		code, err = wpwc.WorkProc.SCHandler.GetGlobalState(params.GlobalContractKey)
	} else {
		code, err = wpwc.WorkProc.SCHandler.GetState(vm.ContractCodeKey)
	}

	if len(code) != 0 && err == nil {
		contractCode, err := vm.DoContractStateData(code)
		if err != nil {
			return "", fmt.Errorf("cat't find contract code in db(1), err: %+v, contract_addr: %+v, len(code): %+v",
				err, wpwc.WorkProc.ContractData.ContractAddr, len(code))
		}
		err = json.Unmarshal(contractCode, cc)
		if err != nil {
			return "", fmt.Errorf("cat't find contract code in db(2), err: %+v, contract_addr: %+v, len(code): %+v", err, wpwc.WorkProc.ContractData.ContractAddr, len(code))
		}
		wpwc.WorkProc.ContractData.ContractCode = string(cc.Code)
		return cc.Type, nil
	} else {
		return "", errors.New(fmt.Sprintf("can't find contract code in db,err: %+v, addr: %+v, len(code): %+v",
			err, wpwc.WorkProc.ContractData.ContractAddr, len(code)))
	}
}

func (worker *BsWorker) GetInitType(wpwc *vm.WorkerProcWithCallback) string {
	txType := wpwc.WorkProc.ContractData.Transaction.GetHeader().GetType()
	if txType == proto.TransactionType_LuaContractInit {
		return "lua"
	} else {
		return "js"
	}
}

func (worker *BsWorker) isCommonTransaction(wpwc *vm.WorkerProcWithCallback) bool {
	txType := wpwc.WorkProc.ContractData.Transaction.GetHeader().GetType()
	if txType == proto.TransactionType_LuaContractInit || txType == proto.TransactionType_ContractInvoke ||
		txType == proto.TransactionType_JSContractInit || txType == proto.TransactionType_ContractQuery {
		return false
	}

	return true
}

func (worker *BsWorker) ExecCommonTransaction(wpwc *vm.WorkerProcWithCallback) error {
	return wpwc.WorkProc.SCHandler.Transfer(wpwc.WorkProc.ContractData.Transaction)
}
