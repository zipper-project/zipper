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

package vm

import (
	"encoding/hex"
	"errors"
	"strconv"

	"github.com/zipper-project/zipper/account"
)

func CheckStateKey(key string) error {
	//if ContractCodeKey == key {
	//	return errors.New("state key illegal:" + key)
	//}

	if len(key) > VMConf.ExecLimitMaxStateKeyLength {
		return errors.New("state key size illegal " +
			strconv.Itoa(len(key)) +
			" , max length is:" + strconv.Itoa(VMConf.ExecLimitMaxStateKeyLength))
	}
	return nil
}

func CheckStateValue(value []byte) error {
	if value == nil {
		return nil
	}

	if len(value) > VMConf.ExecLimitMaxStateValueSize {
		return errors.New("state value size illegal " +
			strconv.Itoa(len(value)) +
			"  max size is:" + strconv.Itoa(VMConf.ExecLimitMaxStateValueSize))
	}

	return nil
}

func CheckStateKeyValue(key string, value []byte) error {
	if err := CheckStateKey(key); err != nil {
		return err
	}

	if err := CheckStateValue(value); err != nil {
		return err
	}

	return nil
}

func CheckAddr(addr string) error {
	if addr[0:2] == "0x" {
		addr = addr[2:]
	}

	addrByte, err := hex.DecodeString(addr)
	if err != nil {
		return errors.New("account address illegal")
	}

	if len(addrByte) != account.AddressLength {
		return errors.New("account address illegal")
	}

	return nil
}

func CheckVmMem(mem int) error {
	if mem < 200 {
		VMConf.VMMaxMem = 200
		return errors.New("if maxMem < 200 ,maxMem use the default value 200MB")
	}
	return nil
}
