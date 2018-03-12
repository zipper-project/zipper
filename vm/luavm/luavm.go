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

package luavm

import (
	"encoding/json"
	"errors"
	"fmt"
	"math/rand"
	"strconv"
	"strings"

	"github.com/yuin/gopher-lua"
	"github.com/yuin/gopher-lua/parse"
	"github.com/zipper-project/zipper/common/log"
	"github.com/zipper-project/zipper/params"
	"github.com/zipper-project/zipper/proto"
	"github.com/zipper-project/zipper/vm"
)

// Start start vm process
type LuaWorker struct {
	isInit     bool
	isCanRedo  bool
	workerFlag int

	L          *lua.LState
	VMConf     *vm.Config
	workerProc *vm.WorkerProc
	luaProto   map[string]*lua.FunctionProto
	luaLFunc   map[string]*lua.LFunction
}

// Create Lua worker
func NewLuaWorker(conf *vm.Config) *LuaWorker {
	worker := &LuaWorker{isInit: false}
	worker.workerInit(true, conf)

	return worker
}

// VmJob handler main work
func (worker *LuaWorker) VmJob(data interface{}) (interface{}, error) {
	worker.isCanRedo = false
	return worker.ExecJob(data)
}

// Exec worker
func (worker *LuaWorker) ExecJob(data interface{}) (interface{}, error) {
	workerProcWithCallback := data.(*vm.WorkerProcWithCallback)
	result, err := worker.requestHandle(workerProcWithCallback.WorkProc)
	return result, err
}

func (worker *LuaWorker) VmReady() bool {
	return true
}

func (worker *LuaWorker) VmInitialize() {
	if !worker.isInit {
		worker.workerInit(true, vm.DefaultConfig())
	}
}

func (worker *LuaWorker) VmTerminate() {
	worker.L.Close()
}

func (worker *LuaWorker) requestHandle(wp *vm.WorkerProc) (interface{}, error) {
	txType := wp.ContractData.Transaction.GetHeader().GetType()
	if txType == proto.TransactionType_LuaContractInit {
		return worker.InitContract(wp)
	} else if txType == proto.TransactionType_ContractInvoke {
		return worker.InvokeExecute(wp)
	} else if txType == proto.TransactionType_ContractQuery {
		return worker.QueryContract(wp)
	}

	return nil, errors.New(fmt.Sprintf("luavm no method match transaction type: %d", txType))
}

func (worker *LuaWorker) InitContract(wp *vm.WorkerProc) (interface{}, error) {
	worker.resetProc(wp)
	err := worker.txTransfer()
	if err != nil {
		return nil, err
	}

	err = worker.StoreContractCode()
	if err != nil {
		return false, err
	}

	ok, err := worker.execContract(wp.ContractData, "Init")
	if err != nil {
		return false, err
	}

	if _, ok := ok.(bool); !ok {
		return false, errors.New("InitContract execContract result type is not bool")
	}

	err = worker.workerProc.CCallCommit()

	if err != nil {
		log.Errorf("commit all change error contractAddr:%s, errmsg:%s\n", worker.workerProc.ContractData.ContractAddr, err.Error())
		return false, err
	}

	return ok, err
}

func (worker *LuaWorker) InvokeExecute(wp *vm.WorkerProc) (interface{}, error) {
	worker.resetProc(wp)
	err := worker.txTransfer()
	if err != nil {
		return nil, err
	}

	if len(wp.ContractData.ContractCode) == 0 {
		code, err := worker.GetContractCode()
		if err != nil {
			return nil, errors.New("can't get contract code")
		}
		wp.ContractData.ContractCode = string(code)
	}

	ok, err := worker.execContract(wp.ContractData, "Invoke")
	if err != nil {
		return false, err
	}

	if _, ok := ok.(bool); !ok {
		return false, errors.New("RealExecute execContract result type is not bool")
	}

	err = worker.workerProc.CCallCommit()

	if err != nil {
		log.Errorf("commit all change error contractAddr:%s, errmsg:%s\n", worker.workerProc.ContractData.ContractAddr, err.Error())
		return false, err
	}

	return ok, err
}

func (worker *LuaWorker) QueryContract(wp *vm.WorkerProc) ([]byte, error) {
	worker.resetProc(wp)
	value, err := worker.execContract(wp.ContractData, "Query")
	if err != nil {
		return nil, err
	}

	result, ok := value.(string)
	if !ok {
		return nil, errors.New("QueryContract execContract result type is not string")
	}

	return []byte(result), nil
}

func (worker *LuaWorker) txTransfer() error {
	err := worker.workerProc.SCHandler.Transfer(worker.workerProc.ContractData.Transaction)
	if err != nil {
		return errors.New(fmt.Sprintf("Transfer failed..., err_msg: %s", err))
	}

	return nil
}

func (worker *LuaWorker) resetProc(wp *vm.WorkerProc) {
	worker.workerProc = wp
	worker.workerProc.StateChangeQueue = vm.NewStateQueue()
	worker.workerProc.TransferQueue = vm.NewTransferQueue()
	worker.L = worker.newState()
	loader := func(L *lua.LState) int {
		mod := L.SetFuncs(L.NewTable(), exporter(worker.workerProc)) // register functions to the table
		L.Push(mod)
		return 1
	}
	worker.L.PreloadModule("ZIP", loader)
}

func (worker *LuaWorker) workerInit(isInit bool, vmconf *vm.Config) {
	worker.isInit = true
	worker.VMConf = vmconf
	worker.luaProto = make(map[string]*lua.FunctionProto)
	worker.luaLFunc = make(map[string]*lua.LFunction)
	worker.workerFlag = rand.Int()
}

// execContract start a lua vm and execute smart contract script
func (worker *LuaWorker) execContract(cd *vm.ContractData, funcName string) (interface{}, error) {
	defer func() {
		if e := recover(); e != nil {
			log.Error("LuaVM exec contract code error ", e)
		}
	}()

	if err := worker.CheckContractCode(cd.ContractCode); err != nil {
		return false, err
	}

	//worker.L = worker.newState()
	//defer worker.L.Close()
	//
	////ctx, cancel := context.WithTimeout(context.Background(), time.Duration(worker.VMConf.ExecLimitMaxRunTime)*time.Millisecond)
	////defer cancel()
	////
	////worker.L.SetContext(ctx)
	//
	//ctx, cancel := context.WithCancel(context.Background())
	//worker.L.SetContext(ctx)
	//timeOut := time.Duration(worker.VMConf.ExecLimitMaxRunTime) * time.Millisecond
	//timeOutChann := make(chan bool, 1)
	//defer func() {
	//	timeOutChann <- true
	//}()
	//
	//go func() {
	//	gog.Add(1)
	//	defer gog.Done()
	//	select {
	//	case <-timeOutChann:
	//		worker.L.RemoveContext()
	//	case <-time.After(timeOut):
	//		cancel()
	//	}
	//}()

	//startTime := time.Now()
	//loader := func(L *lua.LState) int {
	//	mod := L.SetFuncs(L.NewTable(), exporter(worker.workerProc)) // register functions to the table
	//	L.Push(mod)
	//	return 1
	//}
	//worker.L.PreloadModule("ZIP", loader)

	_, ok := worker.luaProto[cd.ContractAddr]
	if !ok {
		chunk, err := parse.Parse(strings.NewReader(cd.ContractCode), "<string>")
		if err != nil {
			return nil, err
		}
		proto, err := lua.Compile(chunk, "<string>")
		if err != nil {
			return nil, err
		}
		worker.luaProto[cd.ContractAddr] = proto
	}

	fn := &lua.LFunction{
		IsG: false,
		Env: worker.L.Env,

		Proto:     worker.luaProto[cd.ContractAddr],
		GFunction: nil,
		Upvalues:  make([]*lua.Upvalue, 0)}
	worker.L.Push(fn)

	if err := worker.L.PCall(0, lua.MultRet, nil); err != nil {
		return false, err
	}

	callLuaFuncResult, err := worker.callLuaFunc(worker.L, funcName, cd.ContractParams...)

	return callLuaFuncResult, err
}

func (worker *LuaWorker) GetContractCode() (string, error) {
	var err error
	cc := new(vm.ContractCode)
	var code []byte
	if len(worker.workerProc.ContractData.ContractAddr) == 0 {
		code, err = worker.workerProc.SCHandler.GetGlobalState(params.GlobalContractKey)
	} else {
		code, err = worker.workerProc.SCHandler.GetState(vm.ContractCodeKey)
	}

	if len(code) != 0 && err == nil {
		contractCode, err := vm.DoContractStateData(code)
		if err != nil {
			return "", fmt.Errorf("cat't find contract code in db, err: %+v", err)
		}
		err = json.Unmarshal(contractCode, cc)
		if err != nil {
			return "", fmt.Errorf("cat't find contract code in db, err: %+v", err)
		}

		return string(cc.Code), nil
	} else {
		return "", errors.New("cat't find contract code in db")
	}
}

func (worker *LuaWorker) StoreContractCode() error {
	code, err := vm.ConcrateStateJson(&vm.ContractCode{Code: []byte(worker.workerProc.ContractData.ContractCode), Type: "luavm"})
	if err != nil {
		log.Errorf("Can't concrate contract code")
	}

	if len(worker.workerProc.ContractData.ContractAddr) == 0 {
		err = worker.workerProc.CCallPutState(params.GlobalContractKey, code.Bytes())
	} else {
		err = worker.workerProc.CCallPutState(vm.ContractCodeKey, code.Bytes()) // add js contract code into state
	}

	return err
}

func (worker *LuaWorker) CheckContractCode(code string) error {
	if len(code) == 0 || len(code) > worker.VMConf.ExecLimitMaxScriptSize {
		return errors.New("contract script code size illegal " +
			strconv.Itoa(len(code)) +
			"byte , max size is:" +
			strconv.Itoa(worker.VMConf.ExecLimitMaxScriptSize) + " byte")
	}

	return nil
}

// newState create a lua vm
func (worker *LuaWorker) newState() *lua.LState {
	opt := lua.Options{
		SkipOpenLibs:        true,
		CallStackSize:       worker.VMConf.VMCallStackSize,
		RegistrySize:        worker.VMConf.VMRegistrySize,
		MaxAllowOpCodeCount: worker.VMConf.ExecLimitMaxOpcodeCount,
	}
	L := lua.NewState(opt)

	// forbid: lua.IoLibName, lua.OsLibName, lua.DebugLibName, lua.ChannelLibName, lua.CoroutineLibName
	worker.openLib(L, lua.LoadLibName, lua.OpenPackage)
	worker.openLib(L, lua.BaseLibName, lua.OpenBase)
	worker.openLib(L, lua.TabLibName, lua.OpenTable)
	worker.openLib(L, lua.StringLibName, lua.OpenString)
	worker.openLib(L, lua.MathLibName, lua.OpenMath)

	return L
}

// openLib loads the built-in libraries. It is equivalent to running OpenLoad,
// then OpenBase, then iterating over the other OpenXXX functions in any order.
func (worker *LuaWorker) openLib(L *lua.LState, libName string, libFunc lua.LGFunction) {
	L.Push(L.NewFunction(libFunc))
	L.Push(lua.LString(libName))
	L.Call(1, 0)
}

// call lua function(Init, Invoke)
func (worker *LuaWorker) callLuaFunc(L *lua.LState, funcName string, params ...string) (interface{}, error) {
	p := lua.P{
		Fn:      L.GetGlobal(funcName),
		NRet:    1,
		Protect: true,
	}

	//log.Debugf("callLuaFunc, funcName: %+v, Parms: %+v", funcName, params)
	var err error
	l := len(params)
	var lvparams []lua.LValue
	if "Invoke" == funcName {
		if l == 0 {
			lvparams = []lua.LValue{lua.LNil, lua.LNil}
		} else if l == 1 {
			lvparams = []lua.LValue{lua.LString(params[0]), lua.LNil}
		} else if l > 1 {
			tb := new(lua.LTable)
			for i := 1; i < l; i++ {
				tb.RawSet(lua.LNumber(i-1), lua.LString(params[i]))
			}
			lvparams = []lua.LValue{lua.LString(params[0]), tb}
		}
	} else {
		if l == 0 {
			lvparams = []lua.LValue{}
		} else if l > 0 {
			tb := new(lua.LTable)
			for i := 0; i < l; i++ {
				tb.RawSet(lua.LNumber(i), lua.LString(params[i]))
			}
			lvparams = []lua.LValue{tb}
		}
	}

	err = L.CallByParam(p, lvparams...)
	if err != nil {
		return false, err
	}

	if _, ok := L.Get(-1).(lua.LBool); ok {
		ret := L.ToBool(-1)
		L.Pop(1) // remove received value
		return ret, nil
	}

	queryResult := L.ToString(-1)
	L.Pop(1) // remove received value
	return queryResult, nil
}
