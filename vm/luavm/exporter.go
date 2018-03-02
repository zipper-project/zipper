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
	"bytes"
	"time"

	"github.com/yuin/gopher-lua"
	"github.com/zipper-project/zipper/vm"
	"github.com/zipper-project/zipper/common/log"
	luajson "github.com/zipper-project/zipper/vm/luavm/json"
)

func exporter(workerProc *vm.WorkerProc) map[string]lua.LGFunction {
	return map[string]lua.LGFunction{
		"GetGlobalState": getGlobalStateFunc(workerProc),
		"PutGlobalState": setGlobalStateFunc(workerProc),
		"DelGlobalState": delGlobalStateFunc(workerProc),
		"GetState":       getStateFunc(workerProc),
		"PutState":       putStateFunc(workerProc),
		"DelState":       delStateFunc(workerProc),

		//"GetByPrefix":  getByPrefixFunc(workerProc),
		//"GetByRange":   getByRangeFunc(workerProc),
		"ComplexQuery": complexQueryFunc(workerProc),

		"GetBalance":  getBalanceFunc(workerProc),
		"GetBalances": getBalancesFunc(workerProc),
		"Account":     accountFunc(workerProc),
		"TxInfo":      txInfo(workerProc),
		"Transfer":    transferFunc(workerProc),

		"CurrentBlockHeight": currentBlockHeightFunc(workerProc),
		"sleep":              sleepFunc(workerProc),
		"jsonEncode":         luajson.ApiEncode(),
		"jsonDecode":         luajson.ApiDecode(),
	}
}

func getGlobalStateFunc(workerProc *vm.WorkerProc) lua.LGFunction {
	return func(l *lua.LState) int {
		if l.GetTop() != 1 {
			l.RaiseError("param illegality when invoke GetGlobalState")
			return 1
		}

		key := l.CheckString(1)
		data, err := workerProc.CCallGetGlobalState(key)
		if err != nil {
			l.RaiseError("GetGlobalState error key:%s  err:%s", key, err)
			return 1
		}

		if data == nil {
			l.Push(lua.LNil)
			return 1
		}

		buf := bytes.NewBuffer(data)
		if lv, err := byteToLValue(buf); err != nil {
			l.RaiseError("byteToLValue error")
		} else {
			l.Push(lv)
		}

		return 1
	}
}

func setGlobalStateFunc(workerProc *vm.WorkerProc) lua.LGFunction {
	return func(l *lua.LState) int {
		if l.GetTop() != 2 {
			l.RaiseError("param illegality when invoke SetGlobalState")
			return 1
		}

		key := l.CheckString(1)
		value := l.Get(2)

		data := lvalueToByte(value)
		err := workerProc.CCallSetGlobalState(key, data)
		if err != nil {
			l.RaiseError("SetGlobalState error key:%s  err:%s", key, err)
		} else {
			l.Push(lua.LBool(true))
		}

		return 1
	}
}

func delGlobalStateFunc(workerProc *vm.WorkerProc) lua.LGFunction {
	return func(l *lua.LState) int {
		if l.GetTop() != 1 {
			l.RaiseError("param illegality when invoke DelGlobalState")
			return 1
		}

		key := l.CheckString(1)
		err := workerProc.CCallDelGlobalState(key)
		if err != nil {
			l.RaiseError("DelGlobalState error key:%s   err:%s", key, err)
		} else {
			l.Push(lua.LBool(true))
		}

		return 1
	}
}

func getStateFunc(workerProc *vm.WorkerProc) lua.LGFunction {
	return func(l *lua.LState) int {
		if l.GetTop() != 1 {
			l.RaiseError("param illegality when invoke GetState")
			return 1
		}

		key := l.CheckString(1)
		data, err := workerProc.CCallGetState(key)
		if err != nil {
			l.RaiseError("getState error key:%s  err:%s", key, err)
			return 1
		}
		if data == nil {
			l.Push(lua.LNil)
			return 1
		}

		buf := bytes.NewBuffer(data)
		if lv, err := byteToLValue(buf); err != nil {
			l.RaiseError("byteToLValue error")
		} else {
			l.Push(lv)
		}

		return 1
	}
}

func putStateFunc(workerProc *vm.WorkerProc) lua.LGFunction {
	return func(l *lua.LState) int {
		if l.GetTop() != 2 {
			l.RaiseError("param illegality when invoke PutState")
			return 1
		}

		key := l.CheckString(1)
		value := l.Get(2)

		data := lvalueToByte(value)

		//log.Debugf("workerProc: %+p", workerProc)
		err := workerProc.CCallPutState(key, data)
		if err != nil {
			l.RaiseError("putState error key:%s  err:%s", key, err)
		} else {
			l.Push(lua.LBool(true))
		}

		return 1
	}
}

func delStateFunc(workerProc *vm.WorkerProc) lua.LGFunction {
	return func(l *lua.LState) int {
		if l.GetTop() != 1 {
			l.RaiseError("param illegality when invoke DelState")
			return 1
		}

		key := l.CheckString(1)
		err := workerProc.CCallDelState(key)
		if err != nil {
			l.RaiseError("delState error key:%s   err:%s", key, err)
		} else {
			l.Push(lua.LBool(true))
		}

		return 1
	}
}
//
//func getByPrefixFunc(workerProc *vm.WorkerProc) lua.LGFunction {
//	return func(l *lua.LState) int {
//		if l.GetTop() != 1 {
//			l.RaiseError("param illegality when invoke getByPrefix")
//			return 1
//		}
//
//		key := l.CheckString(1)
//
//		values, err := workerProc.CCallGetByPrefix(key)
//
//		if err != nil {
//			l.RaiseError("getByPrefix error key:%s  err:%s", key, err)
//		}
//		if values == nil {
//			l.Push(lua.LNil)
//			return 1
//		}
//
//		if lv, err := kvsToLValue(values); err != nil {
//			l.RaiseError("byteToLValue error")
//		} else {
//			l.Push(lv)
//		}
//
//		return 1
//	}
//}
//
//func getByRangeFunc(workerProc *vm.WorkerProc) lua.LGFunction {
//	return func(l *lua.LState) int {
//		if l.GetTop() != 2 {
//			l.RaiseError("param illegality when invoke getByRange")
//			return 1
//		}
//
//		startKey := l.CheckString(1)
//		limitKey := l.CheckString(2)
//		values, err := workerProc.CCallGetByRange(startKey, limitKey)
//		if err != nil {
//			l.RaiseError("getByRange error startKey:%s ,limitKsy:%s ,err:%s", startKey, limitKey, err)
//		}
//		if values == nil {
//			l.Push(lua.LNil)
//			return 1
//		}
//
//		lv, err := kvsToLValue(values)
//		if err != nil {
//			l.RaiseError("byteToLValue error")
//		} else {
//			l.Push(lv)
//		}
//
//		return 1
//	}
//}

func complexQueryFunc(workerProc *vm.WorkerProc) lua.LGFunction {
	return func(l *lua.LState) int {
		if l.GetTop() != 1 {
			l.RaiseError("param illegality when invoke complexQuery")
			return 1
		}

		key := l.CheckString(1)
		data, err := workerProc.CCallComplexQuery(key)
		if err != nil {
			l.RaiseError("complexQuery error  key:%s  err:%s", key, err)
			return 1
		}
		if data == nil {
			l.Push(lua.LNil)
			return 1
		}

		lv := lua.LString(string(data))
		l.Push(lv)
		return 1
	}
}

func getBalanceFunc(workerProc *vm.WorkerProc) lua.LGFunction {
	return func(l *lua.LState) int {
		if l.GetTop() != 2 {
			l.RaiseError("param illegality when invoke getByRange")
			return 1
		}

		addr := l.CheckString(1)
		assetID := l.CheckInt64(2)
		res, err := workerProc.CCallGetBalance(addr, uint32(assetID))
		if err != nil {
			l.RaiseError("getBalance error addr:%s, assetID: %d, err:%s", addr, assetID, err)
			return 1
		}

		val := lua.LNumber(res)
		l.Push(val)
		return 1
	}
}

func getBalancesFunc(workerProc *vm.WorkerProc) lua.LGFunction {
	return func(l *lua.LState) int {
		if l.GetTop() != 1 {
			l.RaiseError("param illegality when invoke getByRange")
			return 1
		}

		addr := l.CheckString(1)
		res, err := workerProc.CCallGetBalances(addr)
		if err != nil {
			l.RaiseError("getBalance error addr:%s, err:%s", addr, err)
			return 1
		}

		l.Push(objToLValue(res))
		return 1
	}
}

func txInfo(workerProc *vm.WorkerProc) lua.LGFunction {
	return func(l *lua.LState) int {
		var sender, recipient string
		var amount, fee int64

		sender = workerProc.ContractData.Transaction.Sender().String()
		amount = workerProc.ContractData.Transaction.Amount()
		fee = workerProc.ContractData.Transaction.Fee()
		recipient = workerProc.ContractData.Transaction.Recipient().String()
		assetID := workerProc.ContractData.Transaction.AssetID()
		hash := workerProc.ContractData.Transaction.Hash().String()
		tb := l.NewTable()
		log.Debugf("worker_lua tx_hash: %+v, tx_sender: %+v, wp: %p, cpar: %#v, tx: %#v", workerProc.ContractData.Transaction.Hash(), sender, workerProc, workerProc.ContractData.ContractParams, workerProc.ContractData.Transaction)
		tb.RawSetString("Sender", lua.LString(sender))
		tb.RawSetString("Recipient", lua.LString(recipient))
		tb.RawSetString("AssetID", lua.LNumber(assetID))
		tb.RawSetString("Amount", lua.LNumber(amount))
		tb.RawSetString("Fee", lua.LNumber(fee))
		tb.RawSetString("Hash", lua.LString(hash))
		l.Push(tb)
		return 1
	}
}

func accountFunc(workerProc *vm.WorkerProc) lua.LGFunction {
	return func(l *lua.LState) int {
		var addr, sender, recipient string
		var amount int64
		if l.GetTop() == 1 {
			addr = l.CheckString(1)
		} else {
			addr = workerProc.ContractData.ContractAddr
		}

		balances, err := workerProc.CCallGetBalances(addr)
		if err != nil {
			l.RaiseError("get balances error addr:%s  err:%s", addr, err)
			return 1
		}
		sender = workerProc.ContractData.Transaction.Sender().String()
		amount = workerProc.ContractData.Transaction.Amount()
		recipient = workerProc.ContractData.Transaction.Recipient().String()
		tb := l.NewTable()
		tb.RawSetString("Sender", lua.LString(sender))
		tb.RawSetString("Address", lua.LString(addr))
		tb.RawSetString("Recipient", lua.LString(recipient))
		tb.RawSetString("Amount", lua.LNumber(amount))
		tb.RawSetString("Balances", objToLValue(balances))
		l.Push(tb)
		return 1
	}
}

func transferFunc(workerProc *vm.WorkerProc) lua.LGFunction {
	return func(l *lua.LState) int {
		if l.GetTop() != 4 {
			l.RaiseError("param illegality when invoke Transfer")
			return 1
		}

		recipientAddr := l.CheckString(1)
		id := int64(float64(l.CheckNumber(2)))
		amout := int64(float64(l.CheckNumber(3)))
		fee := int64(float64(l.CheckNumber(4)))
		err := workerProc.CCallTransfer(recipientAddr, id, amout, fee)
		if err != nil {
			l.RaiseError("contract do transfer error recipientAddr:%s,id:%d, amout:%d, fee:%d  err:%s", recipientAddr, id, amout, fee, err)
			return 1
		}

		l.Push(lua.LBool(false))
		return 1
	}
}

func sleepFunc(workerProc *vm.WorkerProc) lua.LGFunction {
	return func(l *lua.LState) int {
		if l.GetTop() != 1 {
			l.RaiseError("param illegality when invoke sleep")
			return 1
		}
		n := time.Duration(int64(float64(l.CheckNumber(1))))
		time.Sleep(n * time.Millisecond)
		l.Push(lua.LBool(true))
		return 1
	}
}

func currentBlockHeightFunc(workerProc *vm.WorkerProc) lua.LGFunction {
	return func(l *lua.LState) int {
		height, err := workerProc.CCallCurrentBlockHeight()
		if err != nil {
			l.RaiseError("get currentBlockHeight error")
			return 1
		}

		l.Push(lua.LNumber(height))
		return 1
	}
}
