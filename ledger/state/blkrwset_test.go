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
package state

import (
	"os"
	"testing"

	"github.com/zipper-project/zipper/account"
	"github.com/zipper-project/zipper/common/db"
	"github.com/zipper-project/zipper/common/utils"
)

var (
	chaincodeAddr = "0xa032277be213f56221b6140998c03d860a60e1f8"
	balanceAddr   = "0xa132277be213f56221b6140998c03d860a60e1f8"
)

func TestSetChainCodeStateAndGetChainCodeState(t *testing.T) {
	testDB := db.NewDB(db.DefaultConfig())
	defer os.RemoveAll("/tmp/rocksdb-test")
	testKey := "test"
	testValue := []byte("TestSetChainCodeState")
	b := NewBLKRWSet(testDB)
	if err := b.SetChainCodeState(chaincodeAddr, testKey, testValue); err != nil {
		t.Error(err)
	}

	value, err := b.GetChainCodeState(chaincodeAddr, testKey, false)
	if err != nil {
		t.Error(err)
	}

	utils.AssertEquals(t, testValue, value)
}

func TestGetChainCodeStateByRange(t *testing.T) {
	testDB := db.NewDB(db.DefaultConfig())
	defer os.RemoveAll("/tmp/rocksdb-test")
	testKey := "test"
	testValue := []byte("TestSetChainCodeState")

	testKey1 := "test1"
	testValue1 := []byte("TestSetChainCodeState1")

	b := NewBLKRWSet(testDB)
	if err := b.SetChainCodeState(chaincodeAddr, testKey, testValue); err != nil {
		t.Error(err)
	}
	if err := b.SetChainCodeState(chaincodeAddr, testKey1, testValue1); err != nil {
		t.Error(err)
	}

	values, err := b.GetChainCodeStateByRange(chaincodeAddr, testKey, "", false)
	if err != nil {
		t.Error(err)
	}

	if len(values) == 0 {
		t.Error("values if not exist")
	}
	for k, v := range values {
		switch k {
		case testKey:
			utils.AssertEquals(t, v, testValue)
		case testKey1:
			utils.AssertEquals(t, v, testValue1)
		default:
			t.Errorf("have not set key: %s,value: %s.", k, v)
		}
	}

}

func TestDelChainCodeState(t *testing.T) {
	testDB := db.NewDB(db.DefaultConfig())
	defer os.RemoveAll("/tmp/rocksdb-test")
	testKey := "test"
	testValue := []byte("TestSetChainCodeState")
	b := NewBLKRWSet(testDB)
	if err := b.SetChainCodeState(chaincodeAddr, testKey, testValue); err != nil {
		t.Error(err)
	}
	b.DelChainCodeState(chaincodeAddr, testKey)
	value, err := b.GetChainCodeState(chaincodeAddr, testKey, false)
	if err != nil {
		t.Error(err)
	}

	if value != nil {
		t.Errorf("value must be nil,not %v", value)
	}
}

func TestSetBalacneStateAndGetBalanceState(t *testing.T) {
	testDB := db.NewDB(db.DefaultConfig())
	defer os.RemoveAll("/tmp/rocksdb-test")
	testAssetID := uint32(123456)
	testAmount := int64(10000)
	b := NewBLKRWSet(testDB)
	if err := b.SetBalacneState(balanceAddr, testAssetID, testAmount); err != nil {
		t.Error(err)
	}
	amount, err := b.GetBalanceState(balanceAddr, testAssetID, false)
	if err != nil {
		t.Error(err)
	}
	utils.AssertEquals(t, testAmount, amount)

}

func TestGetBalanceStates(t *testing.T) {
	testDB := db.NewDB(db.DefaultConfig())
	defer os.RemoveAll("/tmp/rocksdb-test")
	testAssetID := uint32(123456)
	testAmount := int64(10000)
	testAssetID1 := uint32(1111)
	testAmount1 := int64(10000)
	b := NewBLKRWSet(testDB)
	if err := b.SetBalacneState(balanceAddr, testAssetID, testAmount); err != nil {
		t.Error(err)
	}
	if err := b.SetBalacneState(balanceAddr, testAssetID1, testAmount1); err != nil {
		t.Error(err)
	}

	amounts, err := b.GetBalanceStates(balanceAddr, false)
	if err != nil {
		t.Error(err)
	}

	if len(amounts) == 0 {
		t.Error("amounts if not exist")
	}
	for k, v := range amounts {
		switch k {
		case testAssetID:
			utils.AssertEquals(t, v, testAmount)
		case testAssetID1:
			utils.AssertEquals(t, v, testAmount1)
		default:
			t.Errorf("have not set assetID: %v,amount: %v.", k, v)
		}
	}
}

func TestDelBalanceState(t *testing.T) {
	testDB := db.NewDB(db.DefaultConfig())
	defer os.RemoveAll("/tmp/rocksdb-test")
	testAssetID := uint32(123456)
	testAmount := int64(10000)
	b := NewBLKRWSet(testDB)
	if err := b.SetBalacneState(balanceAddr, testAssetID, testAmount); err != nil {
		t.Error(err)
	}
	b.DelBalanceState(balanceAddr, testAssetID)

	amount, err := b.GetBalanceState(balanceAddr, testAssetID, false)
	if err != nil {
		t.Error(err)
	}
	if amount != 0 {
		t.Errorf("amount must be 0 ,not %v", amount)
	}

}

func TestSetAssetStateAndGetAssetState(t *testing.T) {
	testDB := db.NewDB(db.DefaultConfig())
	defer os.RemoveAll("/tmp/rocksdb-test")
	b := NewBLKRWSet(testDB)
	testAssetID := uint32(123456)
	testAssetInfo := &Asset{
		ID:         123456,
		Name:       "house",
		Descr:      "The house is for sale",
		Precision:  1,
		Expiration: 1520503284,
		Issuer:     account.HexToAddress("0xa032277be213f56221b6140998c03d860a60e1f8"),
		Owner:      account.HexToAddress("0xa132277be213f56221b6140998c03d860a60e1f8"),
	}
	if err := b.SetAssetState(testAssetID, testAssetInfo); err != nil {
		t.Error(err)
	}
	info, err := b.GetAssetState(testAssetID, false)
	if err != nil {
		t.Error(err)
	}
	utils.AssertEquals(t, info, testAssetInfo)
}

func TestGetAssetStates(t *testing.T) {
	testDB := db.NewDB(db.DefaultConfig())
	defer os.RemoveAll("/tmp/rocksdb-test")
	b := NewBLKRWSet(testDB)
	testAssetID := uint32(123456)
	testAssetInfo := &Asset{
		ID:         123456,
		Name:       "house",
		Descr:      "The house is for sale",
		Precision:  1,
		Expiration: 1520503284,
		Issuer:     account.HexToAddress("0xa032277be213f56221b6140998c03d860a60e1f8"),
		Owner:      account.HexToAddress("0xa132277be213f56221b6140998c03d860a60e1f8"),
	}
	testAssetID1 := uint32(111111)
	testAssetInfo1 := &Asset{
		ID:         111111,
		Name:       "house1",
		Descr:      "The house is for sale",
		Precision:  1,
		Expiration: 1520503284,
		Issuer:     account.HexToAddress("0xa032277be213f56221b6140998c03d860a60e1f8"),
		Owner:      account.HexToAddress("0xa132277be213f56221b6140998c03d860a60e1f8"),
	}
	if err := b.SetAssetState(testAssetID, testAssetInfo); err != nil {
		t.Error(err)
	}
	if err := b.SetAssetState(testAssetID1, testAssetInfo1); err != nil {
		t.Error(err)
	}
	infos, err := b.GetAssetStates(false)
	if err != nil {
		t.Error(err)
	}

	if len(infos) != 2 {
		t.Error("infos number not equal ")
	}

	for k, v := range infos {
		switch k {
		case testAssetID:
			utils.AssertEquals(t, v, testAssetInfo)
		case testAssetID1:
			utils.AssertEquals(t, v, testAssetInfo1)
		default:
			t.Errorf("have not set assetID: %v,info: %v.", k, v)
		}
	}

}

func TestDelAssetState(t *testing.T) {
	testDB := db.NewDB(db.DefaultConfig())
	defer os.RemoveAll("/tmp/rocksdb-test")
	b := NewBLKRWSet(testDB)
	testAssetID := uint32(123456)
	testAssetInfo := &Asset{
		ID:         123456,
		Name:       "house",
		Descr:      "The house is for sale",
		Precision:  1,
		Expiration: 1520503284,
		Issuer:     account.HexToAddress("0xa032277be213f56221b6140998c03d860a60e1f8"),
		Owner:      account.HexToAddress("0xa132277be213f56221b6140998c03d860a60e1f8"),
	}
	if err := b.SetAssetState(testAssetID, testAssetInfo); err != nil {
		t.Error(err)
	}
	b.DelAssetState(testAssetID)
	info, err := b.GetAssetState(testAssetID, false)
	if err != nil {
		t.Error(err)
	}
	if info != nil {
		t.Errorf("info must be nil,not %v", info)
	}
}
