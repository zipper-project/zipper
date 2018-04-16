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
	"bytes"
	"io/ioutil"
	"net/http"
	"testing"
	"time"
)

var (
	rpcURLList = []string{
		"http://192.168.8.222:8881",
		"http://192.168.8.222:8882",
		"http://192.168.8.222:8883",
		"http://192.168.8.222:8884",
	}
)

func TestLedgerBlockHeight(t *testing.T) {
	testcase := map[string]string{
		"ST_BLOCK_001": `{"id": 2, "method": "Ledger.Height", "params":[]}`,
	}

	for k, v := range testcase {
		n := time.Now()
		resp := sendRPC(v, rpcURLList...)
		t.Logf("TestCase %s  Used Time:%v  Result: %s ", k, time.Now().Sub(n), string(resp))
	}
}

func BenchmarkBlockHeight(b *testing.B) {
	for i := 0; i < b.N; i++ {
		resp := sendRPC(`{"id": 2, "method": "Ledger.Height", "params":[]}`, rpcURLList...)
		_ = resp
		b.Logf("Result: %s ", string(resp))
	}
}

func BenchmarkBlance(b *testing.B) {
	for i := 0; i < b.N; i++ {
		resp := sendRPC(`{"id": 2, "method": "Ledger.GetBalance", "params":["0xe63b4a97f30fd67f703758e3b14eb6ec9514ece1"]}`, rpcURLList...)
		_ = resp
		b.Logf("Result: %s ", string(resp))
	}
}

func TestLedgerGetLastBlockHash(t *testing.T) {
	testcase := map[string]string{
		"ST_BLOCK_003": `{"id": 2, "method": "Ledger.GetLastBlockHash", "params":[]}`,
	}

	for k, v := range testcase {
		n := time.Now()
		resp := sendRPC(v, rpcURLList...)
		t.Logf("TestCase %s  Used Time:%v  Result: %s ", k, time.Now().Sub(n), string(resp))
	}
}

func BenchmarkGetLastBlockHash(b *testing.B) {
	for i := 0; i < b.N; i++ {
		resp := sendRPC(`{"id": 2, "method": "Ledger.GetLastBlockHash", "params":[]}`, rpcURLList...)
		_ = resp
		b.Logf("Result: %s ", string(resp))
	}
}

func TestBlance(t *testing.T) {
	testcase := map[string]string{
		"ST_BLOCK_005": `{"id": 2, "method": "Ledger.GetBalance", "params":["0xe63b4a97f30fd67f703758e3b14eb6ec9514ece1"]}`,
		"ST_BLOCK_007": `{"id": 2, "method": "Ledger.GetBalance", "params":["0x27c649b7c4f6"]}`,
		"ST_BLOCK_008": `{"id": 2, "method": "Ledger.GetBalance", "params":["0x8bcb81b09016f8f4b33231fe36814651666b3ca3"]}`,
		"ST_BLOCK_009": `{"id": 2, "method": "Ledger.GetBalance", "params":["0x27c649b7c4f66cfaedb99d6b38527db4deda6f41"]}`,
		// "ST_BLOCK_010": `{"id": 2, "method": "Ledger.GetBalance", "params":["0x27c649b7c4f66cfaedb99d6b38527db4deda6f41"]}`,
		// "ST_BLOCK_011": `{"id": 2, "method": "Ledger.GetBalance", "params":["0x27c649b7c4f66cfaedb99d6b38527db4deda6f41"]}`,
	}

	for k, v := range testcase {
		n := time.Now()
		resp := sendRPC(v, rpcURLList...)
		_ = resp
		t.Logf("TestCase [%s]  Used Time:(%v)  Result: %s ", k, time.Now().Sub(n), string(resp))
	}
}

func TestHeight(t *testing.T) {
	testcase := map[string]string{
		"ST_BLOCK_012": `{"id": 2, "method": "Ledger.GetBlockByNumber", "params":[-1]}`,
		"ST_BLOCK_013": `{"id": 2, "method": "Ledger.GetBlockByNumber", "params":[0]}`,
		"ST_BLOCK_014": `{"id": 2, "method": "Ledger.GetBlockByNumber", "params":[18446744073709551615]}`,
		"ST_BLOCK_015": `{"id": 2, "method": "Ledger.GetBlockByNumber", "params":[199999999999999999999999999999999999999999]}`,
	}
	for k, v := range testcase {
		n := time.Now()
		resp := sendRPC(v, rpcURLList...)
		_ = resp
		t.Logf("TestCase [%s]  Used Time:(%v)  Result: %s ", k, time.Now().Sub(n), string(resp))
	}
}

func BenchmarkHeight(b *testing.B) {
	for i := 0; i < b.N; i++ {
		resp := sendRPC(`{"id": 2, "method": "Ledger.GetBlockByNumber", "params":[100]}`, rpcURLList...)
		_ = resp
		b.Logf("Result: %s ", string(resp))
	}
}

func TestGetBlockByHash(t *testing.T) {
	testcase := map[string]string{
		"ST_BLOCK_019": `{"id": 2, "method": "Ledger.GetBlockByHash", "params":[]}`,
		"ST_BLOCK_020": `{"id": 2, "method": "Ledger.GetBlockByHash", "params":["111110274a42707bc8a198cc165bdefc8451d743a7144b0dba8178bc54d962dd"]}`,
		"ST_BLOCK_021": `{"id": 2, "method": "Ledger.GetBlockByHash", "params":["88f720274a42efc8451d743a7144b0dba8178bc54d962dd"]}`,
		"ST_BLOCK_022": `{"id": 2, "method": "Ledger.GetBlockByHash", "params":["4805ee708aab60ae1c8650c495cfb1b8802f6b99177e455e7b303745017fc1f9"]}`,
	}
	for k, v := range testcase {
		n := time.Now()
		resp := sendRPC(v, rpcURLList...)
		t.Logf("TestCase [%s]  Used Time:(%v)  Result: %s ", k, time.Now().Sub(n), string(resp))
	}
}

func TestGetTXByHash(t *testing.T) {
	testcase := map[string]string{
		"ST_BLOCK_024": `{"id": 2, "method": "Ledger.GetTxByHash", "params":["ffbd9b86623c0af028bf2876633e5c5250110105cf2176004181133b77597345"]}`,
		"ST_BLOCK_025": `{"id": 2, "method": "Ledger.GetTxByHash", "params":["c471a3e53b873c0e40c2af2036daaf74570799e9ee52fbda1b6579d3c1825cdb"]}`,
		"ST_BLOCK_026": `{"id": 2, "method": "Ledger.GetTxByHash", "params":["dd1ac5abed28ac12b8b8430060863b5438963b63dee1dc52a1f4be1146954326"]}`,
		"ST_BLOCK_027": `{"id": 2, "method": "Ledger.GetTxByHash", "params":["00bcd88a43e8fdca61f6a3df289df1c82d82b48451e05f51aa9adb0300b4cac2"]}`,
		"ST_BLOCK_028": `{"id": 2, "method": "Ledger.GetTxByHash", "params":[]}`,
		"ST_BLOCK_029": `{"id": 2, "method": "Ledger.GetTxByHash", "params":["ffbd9b86623c0af028bf28766376004181133b77597345"]}`,
	}

	for k, v := range testcase {
		n := time.Now()
		resp := sendRPC(v, rpcURLList...)
		t.Logf("TestCase [%s]  Used Time:(%v)  Result: %s ", k, time.Now().Sub(n), string(resp))
	}
}

func BenchmarkGetTXByHash_2(t *testing.B) {
	for i := 0; i < t.N; i++ {
		resp := sendRPC(`{"id": 2, "method": "Ledger.GetTxByHash", "params":["00bcd88a43e8fdca61f6a3df289df1c82d82b48451e05f51aa9adb0300b4cac2"]}`, rpcURLList...)
		_ = resp
		// t.Logf("TestCase Result: %s ", string(resp))
	}

}
func BenchmarkGetTXByHash_1(t *testing.B) {
	for i := 0; i < t.N; i++ {
		resp := sendRPC(`{"id": 2, "method": "Ledger.GetTxByHash", "params":["dd1ac5abed28ac12b8b8430060863b5438963b63dee1dc52a1f4be1146954326"]}`, rpcURLList...)
		_ = resp
		// t.Logf("TestCase Result: %s ", string(resp))
	}
}

func TestGetTXsByBlockNumber(t *testing.T) {
	testcase := map[string]string{
		"ST_BLOCK_030": `{"id": 2, "method": "Ledger.GetTxsByBlockHash", "params":[{"BlockHash":"","TxType":1}]}`,
		"ST_BLOCK_031": `{"id": 2, "method": "Ledger.GetTxsByBlockHash", "params":[{"BlockHash":"ffbd9b86623c0af028bf2876633e5c5250110105cf2176004181133b77597345","TxType":1}]}`,
		"ST_BLOCK_032": `{"id": 2, "method": "Ledger.GetTxsByBlockHash", "params":[{"BlockHash":"88f720274a42707bc8a198cc165bdefc8451d744b0dba8178bc54d962dd","TxType":1}]}`,
		"ST_BLOCK_033": `{"id": 2, "method": "Ledger.GetTxsByBlockHash", "params":[{"BlockHash":"649c76565b4e30fc0533ae0a834cc4d5bc8c97812602d85a75232f5410d8fc50","TxType":1}]}`,
		"ST_BLOCK_034": `{"id": 2, "method": "Ledger.GetTxsByBlockHash", "params":[{"BlockHash":"649c76565b4e30fc0533ae0a834cc4d5bc8c97812602d85a75232f5410d8fc50","TxType":-1}]}`,
		"ST_BLOCK_035": `{"id": 2, "method": "Ledger.GetTxsByBlockHash", "params":[{"BlockHash":"649c76565b4e30fc0533ae0a834cc4d5bc8c97812602d85a75232f5410d8fc50","TxType":0}]}`,
		"ST_BLOCK_036": `{"id": 2, "method": "Ledger.GetTxsByBlockHash", "params":[{"BlockHash":"649c76565b4e30fc0533ae0a834cc4d5bc8c97812602d85a75232f5410d8fc50","TxType":1}]}`,
		"ST_BLOCK_037": `{"id": 2, "method": "Ledger.GetTxsByBlockHash", "params":[{"BlockHash":"649c76565b4e30fc0533ae0a834cc4d5bc8c97812602d85a75232f5410d8fc50","TxType":10000000000000000}]}`,
		"ST_BLOCK_038": `{"id": 2, "method": "Ledger.GetTxsByBlockHash", "params":[{"BlockHash":"649c76565b4e30fc0533ae0a834cc4d5bc8c97812602d85a75232f5410d8fc50"}]}`,
	}

	for k, v := range testcase {
		n := time.Now()
		resp := sendRPC(v, rpcURLList...)
		t.Logf("TestCase [%s]  Used Time:(%v)  Result: %s ", k, time.Now().Sub(n), string(resp))
	}
}

func sendRPC(params string, address ...string) []byte {

	buf := bytes.NewBuffer(nil)

	for _, addr := range address {
		req, err := http.NewRequest("POST", addr, bytes.NewBufferString(params))
		if err != nil {
			panic(err)
		}

		resp, err := client.Do(req)
		if err != nil {
			panic(err)
		}

		defer resp.Body.Close()

		res, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			panic(err)
		}
		buf.Write(res)
	}

	return buf.Bytes()
}
