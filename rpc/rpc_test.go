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
	"testing"
	"time"
)

var (
	rpcURL = "http://192.168.8.222:8881"
)

func TestSign(t *testing.T) {
	testcase := map[string]string{
		"ST_ACCOUNT_018": `{"id": 1, "method": "Account.Sign", "params": [{"OriginTx": "140001030504000103020101143951c0b0975beb85801f79aa415f38c8bf7ef076030f4240016400fedd011f59", "Addr": "0xf451adb835da32f3c9d40118ef1d02da8c544b81", "Pass": "12345"}]}`,
		"ST_ACCOUNT_019": `{"id": 1, "method": "Account.Sign", "params": [{"OriginTx": "1040001030504000103020101143951c0b0975beb85801f79aa415f38c8bf7ef076030f4240016400fedd011f59", "Addr": "0xf451adb835da32f3c9d40118ef1d02da8c544b81", "Pass": "12345"}]}`,
		"ST_ACCOUNT_020": `{"id": 1, "method": "Account.Sign", "params": [{"OriginTx": "40001030504000103020101143951c0b0975beb85801f79aa415f38c8bf7ef076030f4240016400fedd011f59", "Addr": "0xf451adb835da32f3c9d40118ef1d02da8c544b81", "Pass": "12345"}]}`,
		"ST_ACCOUNT_021": `{"id": 1, "method": "Account.Sign", "params": [{"OriginTx": "040001030504000103020101143951c0b0975beb85801f79aa415f38c8bf7ef076030f4240016400fedd011f59", "Addr": "0x723056db1ba8497a1a00a79a252a51df34e029b66", "Pass": "12345"}]}`,
		"ST_ACCOUNT_022": `{"id": 1, "method": "Account.Sign", "params": [{"OriginTx": "040001030504000103020101143951c0b0975beb85801f79aa415f38c8bf7ef076030f4240016400fedd011f59", "Addr": "0x723056db1ba8497a1a00a79a252a51df34e029b", "Pass": "12345"}]}`,
		"ST_ACCOUNT_023": `{"id": 1, "method": "Account.Sign", "params": [{"OriginTx": "040001030504000103020101143951c0b0975beb85801f79aa415f38c8bf7ef076030f4240016400fedd011f59", "Addr": "0x723056db1ba8497a1a00a79a252a51df34e029b6", "Pass": "12345"}]}`,
		"ST_ACCOUNT_024": `{"id": 1, "method": "Account.Sign", "params": [{"OriginTx": "040001030504000103020101143951c0b0975beb85801f79aa415f38c8bf7ef076030f4240016400fedd011f59", "Addr": "0xf451adb835da32f3c9d40118ef1d02da8c544b81", "Pass": "12345"}]}`,
		"ST_ACCOUNT_025": `{"id": 1, "method": "Account.Sign", "params": [{"OriginTx": "040001030504000103020101143951c0b0975beb85801f79aa415f38c8bf7ef076030f4240016400fedd011f59", "Addr": "0xf451adb835da32f3c9d40118ef1d02da8c544b81", "Pass": ""}]}`,
		"ST_ACCOUNT_026": `{"id": 1, "method": "Account.Sign", "params": [{"OriginTx": "040001030504000103020101143951c0b0975beb85801f79aa415f38c8bf7ef076030f4240016400fedd011f59", "Addr": "", "Pass": "123456"}]}`,
		"ST_ACCOUNT_027": `{"id": 1, "method": "Account.Sign", "params": [{"OriginTx": "", "Addr": "0x723056db1ba8497a1a00a79a252a51df34e029b6", "Pass": "123456"}]}`,
		"ST_ACCOUNT_028": `{"id": 1, "method": "Account.Sign", "params": [{"OriginTx": "", "Addr": "", "Pass": "123456"}]}`,
	}

	for k, v := range testcase {
		now := time.Now()
		resp := sendRPC(v, rpcURL)
		_ = resp
		t.Logf("TestCase [%s]  Used Time:(%v)   Result: %s ", k, time.Now().Sub(now), string(resp))
	}
}

func TestCreateTransaction(t *testing.T) {
	testcase := map[string]string{
		"ST_ACCOUNT_029":   `{"id": 1, "method": "Transaction.Create", "params": [{"FromChain": ",0,1", "ToChain": "0,1", "Recipient": "0x27c649b7c4f66cfaedb99d6b38527db4deda6f41", "Amount": 1000000, "Fee": 100, "TxType": 1}]}`,
		"ST_ACCOUNT_029_1": `{"id": 1, "method": "Transaction.Create", "params": [{"FromChain": "0,1000", "ToChain": "1000,1", "Recipient": "0x27c649b7c4f66cfaedb99d6b38527db4deda6f41", "Amount": 1000000, "Fee": 100, "TxType": 1}]}`,
		"ST_ACCOUNT_030":   `{"id": 1, "method": "Transaction.Create", "params": [{"FromChain": "0,1", "ToChain": "0,1", "Recipient": "0x27c649b7c4f66cfaedb99d6b38527db4deda6f41", "Amount": 1000000, "Fee": 100, "TxType": 1}]}`,
		"ST_ACCOUNT_031":   `{"id": 1, "method": "Transaction.Create", "params": [{"FromChain": "0,1,3,5", "ToChain": "0,1,3,2", "Recipient": "0x27c649b7c4f66cfaedb99d6b38527db4deda6f", "Amount": 1000000, "Fee": 100, "TxType": 1}]}`,
		"ST_ACCOUNT_032":   `{"id": 1, "method": "Transaction.Create", "params": [{"FromChain": "0,1,3,5", "ToChain": "0,1,3,2", "Recipient": "0x27c649b7c4f66cfaedb99d6b38527db4deda6f41", "Amount": 1000000, "Fee": 100, "TxType": 1}]}`,
		"ST_ACCOUNT_033":   `{"id": 1, "method": "Transaction.Create", "params": [{"FromChain": "0,1,3,5", "ToChain": "0,1,3,2", "Recipient": "", "Amount": 1000000, "Fee": 100, "TxType": 1}]}`,
		"ST_ACCOUNT_034":   `{"id": 1, "method": "Transaction.Create", "params": [{"FromChain": "0,1,3,5", "ToChain": "0,1,3,2", "Recipient": "0x27c649b7c4f66cfaedb99d6b38527db4deda6f41", "Amount": 1000000, "Fee": 100, "TxType": 1}]}`,
		"ST_ACCOUNT_035":   `{"id": 1, "method": "Transaction.Create", "params": [{"FromChain": "0,1,3,5", "ToChain": "0,1,3,2", "Recipient": "0x27c649b7c4f66cfaedb99d6b38527db4deda6f41", "Amount": -1000000, "Fee": 100, "TxType": 1}]}`,
		"ST_ACCOUNT_036":   `{"id": 1, "method": "Transaction.Create", "params": [{"FromChain": "0,1,3,5", "ToChain": "0,1,3,2", "Recipient": "0x27c649b7c4f66cfaedb99d6b38527db4deda6f41", "Amount": 1000000, "Fee": 1000000, "TxType": 1}]}`,
		"ST_ACCOUNT_036_1": `{"id": 1, "method": "Transaction.Create", "params": [{"FromChain": "0,1,3,5", "ToChain": "0,1,3,2", "Recipient": "0x27c649b7c4f66cfaedb99d6b38527db4deda6f41", "Amount": 0, "Fee": 0, "TxType": 1}]}`,
		"ST_ACCOUNT_036_2": `{"id": 1, "method": "Transaction.Create", "params": [{"FromChain": "0,1,3,5", "ToChain": "0,1,3,2", "Recipient": "0x27c649b7c4f66cfaedb99d6b38527db4deda6f41", "Amount": -1, "Fee": -1, "TxType": 1}]}`,
		"ST_ACCOUNT_036_3": `{"id": 1, "method": "Transaction.Create", "params": [{"FromChain": "0,1,3,5", "ToChain": "0,1,3,2", "Recipient": "0x27c649b7c4f66cfaedb99d6b38527db4deda6f41", "Amount": 100000000000000000000000000000, "Fee": 10000000000000000000000000000000000, "TxType": 1}]}`,
		"ST_ACCOUNT_037":   `{"id": 1, "method": "Transaction.Create", "params": [{"FromChain": "0,1,3,5", "ToChain": "0,1,3,2", "Recipient": "0x27c649b7c4f66cfaedb99d6b38527db4deda6f41", "Amount": 0, "Fee": 100, "TxType": 1}]}`,
		"ST_ACCOUNT_038":   `{"id": 1, "method": "Transaction.Create", "params": [{"FromChain": "0,1,3,5", "ToChain": "0,1,3,2", "Recipient": "0x27c649b7c4f66cfaedb99d6b38527db4deda6f41", "Amount": 1000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000, "Fee": 100, "TxType": 1}]}`,
		"ST_ACCOUNT_039":   `{"id": 1, "method": "Transaction.Create", "params": [{"FromChain": "0,1,3,5", "ToChain": "0,1,3,2", "Recipient": "0x27c649b7c4f66cfaedb99d6b38527db4deda6f41", "Amount": 1, "Fee": 100, "TxType": 1}]}`,
		"ST_ACCOUNT_040":   `{"id": 1, "method": "Transaction.Create", "params": [{"FromChain": "0,1,3,5", "ToChain": "0,1,3,2", "Recipient": "0x27c649b7c4f66cfaedb99d6b38527db4deda6f41", "Amount": 1000000, "Fee": 100, "TxType": -1}]}`,
		"ST_ACCOUNT_041":   `{"id": 1, "method": "Transaction.Create", "params": [{"FromChain": "0,1,3,5", "ToChain": "0,1,3,2", "Recipient": "0x27c649b7c4f66cfaedb99d6b38527db4deda6f41", "Amount": 1000000, "Fee": 100, "TxType": 0}]}`,
		"ST_ACCOUNT_042":   `{"id": 1, "method": "Transaction.Create", "params": [{"FromChain": "0,1,3,5", "ToChain": "0,1,3,2", "Recipient": "0x27c649b7c4f66cfaedb99d6b38527db4deda6f41", "Amount": 1000000, "Fee": 100, "TxType": 1}]}`,
		"ST_ACCOUNT_043":   `{"id": 1, "method": "Transaction.Create", "params": [{"FromChain": "0,1,3,5", "ToChain": "0,1,3,2", "Recipient": "0x27c649b7c4f66cfaedb99d6b38527db4deda6f41", "Amount": 1000000, "Fee": 100, "TxType": 2}]}`,
		"ST_ACCOUNT_044":   `{"id": 1, "method": "Transaction.Create", "params": [{"FromChain": "0,1,3,5", "ToChain": "0,1,3,2", "Recipient": "0x27c649b7c4f66cfaedb99d6b38527db4deda6f41", "Amount": 1000000, "Fee": 100, "TxType": 3}]}`,
		"ST_ACCOUNT_045":   `{"id": 1, "method": "Transaction.Create", "params": [{"FromChain": "0,1,3,5", "ToChain": "0,1,3,2", "Recipient": "0x27c649b7c4f66cfaedb99d6b38527db4deda6f41", "Amount": 1000000, "Fee": 100, "TxType": 4}]}`,
		"ST_ACCOUNT_046":   `{"id": 1, "method": "Transaction.Create", "params": [{"FromChain": "0,1,3,5", "ToChain": "0,1,3,2", "Recipient": "0x27c649b7c4f66cfaedb99d6b38527db4deda6f41", "Amount": 1000000, "Fee": 100, "TxType": 5}]}`,
		"ST_ACCOUNT_047":   `{"id": 1, "method": "Transaction.Create", "params": [{"FromChain": "0,1,3,5", "ToChain": "0,1,3,2", "Recipient": "0x27c649b7c4f66cfaedb99d6b38527db4deda6f41", "Amount": 1000000, "Fee": 100, "TxType": 6}]}`,
		"ST_ACCOUNT_048":   `{"id": 1, "method": "Transaction.Create", "params": [{"FromChain": "0,1,3,5", "ToChain": "0,1,3,2", "Recipient": "0x27c649b7c4f66cfaedb99d6b38527db4deda6f41", "Amount": 1000000, "Fee": 100, "TxType": 7}]}`,
		"ST_ACCOUNT_049":   `{"id": 1, "method": "Transaction.Create", "params": [{"FromChain": "0,1,3,5", "ToChain": "0,1,3,2", "Recipient": "0x27c649b7c4f66cfaedb99d6b38527db4deda6f41", "Amount": 1000000, "Fee": 100, "TxType": 8}]}`,
		"ST_ACCOUNT_050":   `{"id": 1, "method": "Transaction.Create", "params": [{"FromChain": "0,1,3,5", "ToChain": "0,1,3,2", "Recipient": "0x27c649b7c4f66cfaedb99d6b38527db4deda6f41", "Amount": 1000000, "Fee": 100, "TxType": 9}]}`,
		"ST_ACCOUNT_051":   `{"id": 1, "method": "Transaction.Create", "params": [{"FromChain": "0,1,3,5", "ToChain": "0,1,3,2", "Recipient": "0x27c649b7c4f66cfaedb99d6b38527db4deda6f41", "Amount": 1000000, "Fee": 100, "TxType": 100000000000000000000000000000000000000000000000000}]}`,
	}

	for k, v := range testcase {
		now := time.Now()
		resp := sendRPC(v, rpcURL)
		_ = resp
		t.Logf("TestCase [%s]  Used Time:(%v)   Result: %s ", k, time.Now().Sub(now), string(resp))
	}
}

func TestBroadcastTransaction(t *testing.T) {
	testcase := map[string]string{
		"ST_ACCOUNT_052": `{"id":1, "method":"Transaction.Broadcast", "params":["0000000114c19dffb10ebd0adf25855cf674f3cd6e072cb939eb499ae0069d8574ba748521ea075b802f8a8227738277814171155f86e80127767dc765776e6d51ab3b63461fed41921b01"]}`,
		"ST_ACCOUNT_053": `{"id":1, "method":"Transaction.Broadcast", "params":["0000000114c19dffb10ea84077bc0a402282dfe91fe2b717970502540be400043b9aca0041f0bd0adf25855cf674f3cd6e072cb939eb499ae0069d8574ba748521ea075b802f8a8227710293841203478910238938277814171155f86e80127767dc765776e6d51ab3b63461fed41921b01"]}`,
		"ST_ACCOUNT_054": `{"id":1, "method":"Transaction.Broadcast", "params":["0000000114c19dffb10ea84077bc0a402282dfe91fe2b717970502540be400043b9aca0041f0bd0adf25855cf674f3cd6e072cb939eb499ae0069d8574ba748521ea075b802f8a8227738277814171155f86e80127767dc765776e6d51ab3b63461fed41921b01"]}`,
		"ST_ACCOUNT_055": `{"id":1, "method":"Transaction.Broadcast", "params":["123946871823769487123964871239647116982374613246123649182374689123746013274301246012378460123764901378246032789r6078a12123410123946876190238746071236407912334598230451203498671023987460123976512736547176108"]}`,
		"ST_ACCOUNT_056": `{"id":1, "method":"Transaction.Broadcast", "params":[""]}`,
		"ST_ACCOUNT_057": `{"id":1, "method":"Transaction.Broadcast", "params":["0000000114c19dffb10ea84077bc0a402282dfe91fe2b717970502540be400043b9aca0041f0bd0adf25855cf674f3cd6e072cb939eb499ae0069d8574ba748521ea075b802f8a8227738277814171155f86e80127767dc765776e6d51ab3b63461fed41921b01"]}`,
	}

	for k, v := range testcase {
		now := time.Now()

		resp := sendRPC(v, "http://192.168.8.222:8882")
		_ = resp
		t.Logf("TestCase [%s]  Used Time:(%v)   Result: %s ", k, time.Now().Sub(now), string(resp))
	}
}
