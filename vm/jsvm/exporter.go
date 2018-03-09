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

package jsvm

import (
	"bytes"
	"time"

	"github.com/robertkrimen/otto"
	"github.com/zipper-project/zipper/common/log"
	"github.com/zipper-project/zipper/vm"
)

func exporter(ottoVM *otto.Otto, workerProc *vm.WorkerProc) (*otto.Object, error) {
	exporterFuncs, _ := ottoVM.Object(`ZIP = {
		toNumber: function(value, def) {
			if(typeof(value) == "undefined") {
				return def;
			} else {
				return value * 1;
			}
		}
	}`)

	exporterFuncs.Set("GetGlobalState", getGlobalStateFunc(workerProc))
	exporterFuncs.Set("PutGlobalState", putGlobalStateFunc(workerProc))
	exporterFuncs.Set("DelGlobalState", delGlobalStateFunc(workerProc))
	exporterFuncs.Set("GetState", getStateFunc(workerProc))
	exporterFuncs.Set("PutState", putStateFunc(workerProc))
	exporterFuncs.Set("DelState", delStateFunc(workerProc))

	//exporterFuncs.Set("GetByPrefix", getByPrefixFunc(workerProc))
	//exporterFuncs.Set("GetByRange", getByRangeFunc(workerProc))
	exporterFuncs.Set("ComplexQuery", complexQueryFunc(workerProc))
	exporterFuncs.Set("GetBalance", getBalanceFunc(workerProc))
	exporterFuncs.Set("GetBalances", getBalancesFunc(workerProc))
	exporterFuncs.Set("Account", accountFunc(workerProc))
	exporterFuncs.Set("TxInfo", txInfoFunc(workerProc))
	exporterFuncs.Set("Transfer", transferFunc(workerProc))
	exporterFuncs.Set("CurrentBlockHeight", currentBlockHeightFunc(workerProc))
	exporterFuncs.Set("Sleep", sleepFunc(workerProc))

	return exporterFuncs, nil
}

func getGlobalStateFunc(workerProc *vm.WorkerProc) interface{} {
	return func(fc otto.FunctionCall) otto.Value {
		if len(fc.ArgumentList) != 1 {
			log.Error("param illegality when invoke getGlobalState")
			return fc.Otto.MakeCustomError("getGlobalStateFunc", "param illegality when invoke getGlobalState")
		}

		key, err := fc.Argument(0).ToString()
		data, err := workerProc.CCallGetGlobalState(key)
		if err != nil {
			log.Errorf("getGlobalState error key:%s  err:%s", key, err)
			return fc.Otto.MakeCustomError("getGlobalStateFunc", "getGlobalState error:"+err.Error())
		}

		if data == nil {
			return otto.NullValue()
		}

		buf := bytes.NewBuffer(data)
		val, err := byteToJSvalue(buf, fc.Otto)
		if err != nil {
			log.Error("byteToJSvalue error", err)
			return fc.Otto.MakeCustomError("getGlobalStateFunc", "byteToJSvalue error:"+err.Error())
		}
		return val
	}
}

func putGlobalStateFunc(workerProc *vm.WorkerProc) interface{} {
	return func(fc otto.FunctionCall) otto.Value {
		if len(fc.ArgumentList) != 2 {
			log.Error("param illegality when invoke SetGlobalState")
			return fc.Otto.MakeCustomError("setGlobalStateFunc", "param illegality when invoke SetGlobalState")
		}

		key, err := fc.Argument(0).ToString()
		if err != nil {
			log.Error("get string key error", err)
			return fc.Otto.MakeCustomError("setGlobalStateFunc", "get string key error"+err.Error())
		}

		value := fc.Argument(1)
		data, err := jsvalueToByte(value)
		if err != nil {
			log.Errorf("jsvalueToByte error key:%s  err:%s", key, err)
			return fc.Otto.MakeCustomError("setGlobalStateFunc", "jsvalueToByte error:"+err.Error())
		}

		err = workerProc.CCallSetGlobalState(key, data)
		if err != nil {
			log.Errorf("SetGlobalState error key:%s  err:%s", key, err)
			return fc.Otto.MakeCustomError("setGlobalStateFunc", "SetGlobalState error:"+err.Error())
		}

		val, _ := otto.ToValue(true)
		return val
	}
}

func delGlobalStateFunc(workerProc *vm.WorkerProc) interface{} {
	return func(fc otto.FunctionCall) otto.Value {
		if len(fc.ArgumentList) != 1 {
			log.Error("param illegality when invoke DelGlobalState")
			return fc.Otto.MakeCustomError("delGlobalStateFunc", "param illegality when invoke DelGlobalState")
		}

		key, err := fc.Argument(0).ToString()
		err = workerProc.CCallDelGlobalState(key)
		if err != nil {
			log.Errorf("DelGlobalState error key:%s   err:%s", key, err)
			return fc.Otto.MakeCustomError("delGlobalStateFunc", "DelGlobalState error:"+err.Error())
		}

		val, _ := otto.ToValue(true)
		return val
	}
}

func getStateFunc(workerProc *vm.WorkerProc) interface{} {
	return func(fc otto.FunctionCall) otto.Value {
		if len(fc.ArgumentList) != 1 {
			log.Error("param illegality when invoke GetState")
			return fc.Otto.MakeCustomError("getStateFunc", "param illegality when invoke GetState")
		}

		key, err := fc.Argument(0).ToString()
		data, err := workerProc.CCallGetState(key)
		if err != nil {
			log.Errorf("getState error key:%s  err:%s", key, err)
			return fc.Otto.MakeCustomError("getStateFunc", "getState error:"+err.Error())
		}
		if data == nil {
			return otto.NullValue()
		}

		buf := bytes.NewBuffer(data)
		val, err := byteToJSvalue(buf, fc.Otto)
		if err != nil {
			log.Error("byteToJSvalue error", err)
			return fc.Otto.MakeCustomError("getStateFunc", "byteToJSvalue error:"+err.Error())
		}
		return val
	}
}

func putStateFunc(workerProc *vm.WorkerProc) interface{} {
	return func(fc otto.FunctionCall) otto.Value {
		if len(fc.ArgumentList) != 2 {
			log.Error("param illegality when invoke PutState")
			return fc.Otto.MakeCustomError("putStateFunc", "param illegality when invoke PutState")
		}

		key, err := fc.Argument(0).ToString()
		if err != nil {
			log.Error("get string key error", err)
			return fc.Otto.MakeCustomError("putStateFunc", "get string key error"+err.Error())
		}

		value := fc.Argument(1)
		data, err := jsvalueToByte(value)
		if err != nil {
			log.Errorf("jsvalueToByte error key:%s  err:%s", key, err)
			return fc.Otto.MakeCustomError("putStateFunc", "jsvalueToByte error:"+err.Error())
		}

		err = workerProc.CCallPutState(key, data)
		if err != nil {
			log.Errorf("putState error key:%s  err:%s", key, err)
			return fc.Otto.MakeCustomError("putStateFunc", "putState error:"+err.Error())
		}

		val, _ := otto.ToValue(true)
		return val
	}
}

func delStateFunc(workerProc *vm.WorkerProc) interface{} {
	return func(fc otto.FunctionCall) otto.Value {
		if len(fc.ArgumentList) != 1 {
			log.Error("param illegality when invoke DelState")
			return fc.Otto.MakeCustomError("delStateFunc", "param illegality when invoke DelState")
		}

		key, err := fc.Argument(0).ToString()
		err = workerProc.CCallDelState(key)
		if err != nil {
			log.Errorf("delState error key:%s   err:%s", key, err)
			return fc.Otto.MakeCustomError("delStateFunc", "delState error:"+err.Error())
		}

		val, _ := otto.ToValue(true)
		return val
	}
}

//
//func getByPrefixFunc(workerProc *vm.WorkerProc) interface{} {
//	return func(fc otto.FunctionCall) otto.Value {
//		if len(fc.ArgumentList) != 1 {
//			log.Error("param illegality when invoke GetByPrefix")
//			return fc.Otto.MakeCustomError("getByPrefixFunc", "param illegality when invoke GetByPrefix")
//		}
//
//		perfix, err := fc.Argument(0).ToString()
//		values, err := workerProc.CCallGetByPrefix(perfix)
//		if err != nil {
//			log.Errorf("getByPrefix error key:%s  err:%s", perfix, err)
//			return fc.Otto.MakeCustomError("getByPrefixFunc", "getByPrefix error:"+err.Error())
//		}
//		if values == nil {
//			return otto.NullValue()
//		}
//
//		val, err := kvsToJSValue(values, fc.Otto)
//		if err != nil {
//			log.Error("byteToJSvalue error", err)
//			return fc.Otto.MakeCustomError("getByPrefixFunc", "byteToJSvalue error:"+err.Error())
//		}
//		return val
//	}
//}
//
//func getByRangeFunc(workerProc *vm.WorkerProc) interface{} {
//	return func(fc otto.FunctionCall) otto.Value {
//		if len(fc.ArgumentList) != 2 {
//			log.Error("param illegality when invoke GetByRange")
//			return fc.Otto.MakeCustomError("getByRangeFunc", "param illegality when invoke GetByRange")
//		}
//
//		startKey, err := fc.Argument(0).ToString()
//		limitKey, err := fc.Argument(1).ToString()
//
//		values, err := workerProc.CCallGetByRange(startKey, limitKey)
//		if err != nil {
//			log.Errorf("getByRange error startKey:%s  limitKey:%s  err:%s", startKey, limitKey, err)
//			return fc.Otto.MakeCustomError("getByRangeFunc", "getByRange error:"+err.Error())
//		}
//		if values == nil {
//			return otto.NullValue()
//		}
//
//		val, err := kvsToJSValue(values, fc.Otto)
//		if err != nil {
//			log.Error("byteToJSvalue error", err)
//			return fc.Otto.MakeCustomError("getByRangeFunc", "byteToJSvalue error:"+err.Error())
//		}
//		return val
//	}
//}

func complexQueryFunc(workerProc *vm.WorkerProc) interface{} {
	return func(fc otto.FunctionCall) otto.Value {
		if len(fc.ArgumentList) != 1 {
			log.Error("param illegality when invoke complexQuery")
			return fc.Otto.MakeCustomError("complexQueryFunc", "param illegality when invoke complexQuery")
		}
		key, err := fc.Argument(0).ToString()
		data, err := workerProc.CCallComplexQuery(key)
		if err != nil {
			log.Errorf("complexQuery error key:%s  err:%s", key, err)
			return fc.Otto.MakeCustomError("complexQueryFunc", "complexQuery error:"+err.Error())
		}
		if data == nil {
			return otto.NullValue()
		}

		val, err := otto.ToValue(string(data))
		if err != nil {
			log.Error("byteToJSvalue error", err)
			return fc.Otto.MakeCustomError("complexQueryFunc", "byteToJSvalue error:"+err.Error())
		}
		return val
	}
}

func getBalanceFunc(workerProc *vm.WorkerProc) interface{} {
	return func(fc otto.FunctionCall) otto.Value {
		if len(fc.ArgumentList) != 2 {
			log.Error("param illegality when invoke getBalanceFunc")
			return fc.Otto.MakeCustomError("getBalanceFunc", "param illegality when invoke getBalanceFunc")
		}

		addr, err := fc.Argument(0).ToString()
		assetID, err := fc.Argument(1).ToInteger()
		if err != nil {
			log.Errorf("getBalance error addr:%s, assetID: %d,  err:%s", addr, assetID, err)
			return fc.Otto.MakeCustomError("getBalance", "getBalance error:"+err.Error())
		}

		res, err := workerProc.CCallGetBalance(addr, uint32(assetID))
		if err != nil {
			log.Errorf("CCallGetBalance Error, addr: %s, assetID: %d, err: %s", addr, assetID, err)
			return fc.Otto.MakeCustomError("getBalance", "getBalance error:"+err.Error())
		}

		val, err := otto.ToValue(res)
		if err != nil {
			log.Errorf("CCallGetBalance Error, addr: %s, assetID: %d, err: %s", addr, assetID, err)
			return fc.Otto.MakeCustomError("getBalance", "getBalance error:"+err.Error())
		}

		return val
	}
}

func getBalancesFunc(workerProc *vm.WorkerProc) interface{} {
	return func(fc otto.FunctionCall) otto.Value {
		if len(fc.ArgumentList) != 2 {
			log.Error("param illegality when invoke getBalancesFunc")
			return fc.Otto.MakeCustomError("getBalancesFunc", "param illegality when invoke getBalancesFunc")
		}

		addr, err := fc.Argument(0).ToString()
		if err != nil {
			log.Errorf("getBalances error addr:%s, err:%s", addr, err)
			return fc.Otto.MakeCustomError("getBalances", "getBalances error:"+err.Error())
		}

		res, err := workerProc.CCallGetBalances(addr)
		if err != nil {
			log.Errorf("CCallGetBalance Error, addr: %s, err: %s", addr, err)
			return fc.Otto.MakeCustomError("getBalance", "getBalance error:"+err.Error())
		}

		val, err := objToLValue(res, fc.Otto)
		if err != nil {
			log.Errorf("CCallGetBalances Error, addr: %s, err: %s", addr, err)
			return fc.Otto.MakeCustomError("getBalances", "getBalances error:"+err.Error())
		}

		return val
	}
}

func txInfoFunc(workerProc *vm.WorkerProc) interface{} {
	return func(fc otto.FunctionCall) otto.Value {
		var addr, sender, recipient string
		var amount, fee int64
		var err error
		if len(fc.ArgumentList) == 1 {
			addr, err = fc.Argument(0).ToString()
		} else {
			addr = workerProc.ContractData.ContractAddr
		}

		sender = workerProc.ContractData.Transaction.Sender().String()
		recipient = workerProc.ContractData.Transaction.Recipient().String()

		amount = workerProc.ContractData.Transaction.GetHeader().GetAmount()
		amountValue, err := fc.Otto.ToValue(amount)
		if err != nil {
			log.Error("accountFunc -> call amount ToLValue error", err)
			return fc.Otto.MakeCustomError("accountFunc", "call call amount ToLValue error:"+err.Error())
		}

		fee = workerProc.ContractData.Transaction.GetHeader().GetFee()
		feeValue, err := fc.Otto.ToValue(fee)
		if err != nil {
			log.Error("accountFunc -> call amount ToLValue error", err)
			return fc.Otto.MakeCustomError("accountFunc", "call call amount ToLValue error:"+err.Error())
		}

		assetID := workerProc.ContractData.Transaction.GetHeader().GetFee()
		assetIDValue, err := fc.Otto.ToValue(assetID)
		if err != nil {
			log.Error("accountFunc -> call amount ToLValue error", err)
			return fc.Otto.MakeCustomError("accountFunc", "call call amount ToLValue error:"+err.Error())
		}

		mp := make(map[string]interface{}, 3)
		mp["Address"] = addr
		mp["Sender"] = sender
		mp["Recipient"] = recipient
		mp["Amount"] = amountValue
		mp["Fee"] = feeValue
		mp["AssetID"] = assetIDValue

		val, err := fc.Otto.ToValue(mp)
		if err != nil {
			log.Error("accountFunc -> otto ToValue error", err)
			return fc.Otto.MakeCustomError("accountFunc", "otto ToValue error:"+err.Error())
		}
		return val
	}
}

func accountFunc(workerProc *vm.WorkerProc) interface{} {
	return func(fc otto.FunctionCall) otto.Value {
		var addr, sender, recipient string
		var amount int64
		var err error
		if len(fc.ArgumentList) == 1 {
			addr, err = fc.Argument(0).ToString()
		} else {
			addr = workerProc.ContractData.ContractAddr
		}

		balances, err := workerProc.CCallGetBalances(addr)
		if err != nil {
			log.Error("accountFunc -> call CCallGetBalances error", err)
			return fc.Otto.MakeCustomError("accountFunc", "call CCallGetBalances error:"+err.Error())
		}

		balancesValue, err := objToLValue(balances, fc.Otto)
		if err != nil {
			log.Error("accountFunc -> call objToLValue error", err)
			return fc.Otto.MakeCustomError("accountFunc", "call objToLValue error:"+err.Error())
		}

		sender = workerProc.ContractData.Transaction.Sender().String()
		recipient = workerProc.ContractData.Transaction.Recipient().String()

		amount = workerProc.ContractData.Transaction.GetHeader().GetAmount()
		amountValue, err := fc.Otto.ToValue(amount)
		if err != nil {
			log.Error("accountFunc -> call amount ToLValue error", err)
			return fc.Otto.MakeCustomError("accountFunc", "call call amount ToLValue error:"+err.Error())
		}

		mp := make(map[string]interface{}, 3)
		mp["Address"] = addr
		mp["Balances"] = balancesValue
		mp["Sender"] = sender
		mp["Recipient"] = recipient
		mp["Amount"] = amountValue

		val, err := fc.Otto.ToValue(mp)
		if err != nil {
			log.Error("accountFunc -> otto ToValue error", err)
			return fc.Otto.MakeCustomError("accountFunc", "otto ToValue error:"+err.Error())
		}
		return val
	}
}

func transferFunc(workerProc *vm.WorkerProc) interface{} {
	return func(fc otto.FunctionCall) otto.Value {
		if len(fc.ArgumentList) != 4 {
			log.Error("transferFunc -> param illegality when invoke Transfer")
			return fc.Otto.MakeCustomError("transferFunc", "param illegality when invoke Transfer")
		}

		recipientAddr, err := fc.Argument(0).ToString()
		if err != nil {
			log.Errorf("transferFunc -> get recipientAddr arg error")
			return fc.Otto.MakeCustomError("transferFunc", err.Error())
		}

		id, err := fc.Argument(1).ToInteger()
		if err != nil {
			log.Errorf("transferFunc -> get id arg error")
			return fc.Otto.MakeCustomError("transferFunc", err.Error())
		}

		amout, err := fc.Argument(2).ToInteger()
		if err != nil {
			log.Errorf("transferFunc -> get amout arg error")
			return fc.Otto.MakeCustomError("transferFunc", err.Error())
		}

		fee, err := fc.Argument(3).ToInteger()
		if err != nil {
			log.Errorf("transferFunc -> get fee arg error")
			return fc.Otto.MakeCustomError("transferFunc", err.Error())
		}

		err = workerProc.CCallTransfer(recipientAddr, id, amout, fee)
		if err != nil {
			log.Errorf("transferFunc -> contract do transfer error recipientAddr:%s, amout:%d, fee:%d  err:%s", recipientAddr, amout, fee, err)
			return fc.Otto.MakeCustomError("transferFunc", err.Error())
		}

		val, _ := otto.ToValue(true)
		return val
	}
}

func sleepFunc(workerProc *vm.WorkerProc) interface{} {
	return func(fc otto.FunctionCall) otto.Value {
		if len(fc.ArgumentList) != 1 {
			log.Error("sleepFunc -> param illegality when invoke Transfer")
			return fc.Otto.MakeCustomError("sleepFunc", "param illegality when invoke Sleep")
		}

		n, err := fc.Argument(0).ToInteger()
		if err != nil {
			log.Errorf("sleepFunc -> get duration error")
			return fc.Otto.MakeCustomError("sleepFunc", err.Error())
		}
		time.Sleep(time.Duration(n) * time.Millisecond)
		val, _ := otto.ToValue(true)
		return val
	}
}

func currentBlockHeightFunc(workerProc *vm.WorkerProc) interface{} {
	return func(fc otto.FunctionCall) otto.Value {
		height, err := workerProc.CCallCurrentBlockHeight()
		if err != nil {
			log.Error("currentBlockHeightFunc -> get currentBlockHeight error")
			return fc.Otto.MakeCustomError("currentBlockHeightFunc", "get currentBlockHeight error:"+err.Error())
		}

		val, err := otto.ToValue(height)
		if err != nil {
			return otto.NullValue()
		}
		return val
	}
}
