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
	testConfig = &db.Config{
		DbPath:            "/tmp/rocksdb-test/",
		Columnfamilies:    []string{"balance"},
		KeepLogFileNumber: 10,
		MaxLogFileSize:    10485760,
		LogLevel:          "warn",
	}
)

func TestInitAndUpdateBalance(t *testing.T) {

	testDb := db.NewDB(testConfig)
	s := NewState(testDb)
	a := account.HexToAddress("0xa122277be213f56221b6140998c03d860a60e1f8")

	id := uint32(0)
	amount := int64(1024)
	if err := s.UpdateBalance(a, id, 1024); err != nil {
		t.Error("update balance err:", err)
	}

	writeBatchs := s.WriteBatchs()
	s.dbHandler.AtomicWrite(writeBatchs)

	balance, err := s.GetBalance(a)
	if err != nil {
		t.Error(err)
	}
	t.Log("get balance after update", amount, balance.Get(0))

	utils.AssertEquals(t, balance.Get(0), amount)

	os.RemoveAll("/tmp/rocksdb-test")
}
