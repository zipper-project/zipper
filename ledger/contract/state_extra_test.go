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
package contract

import (
	"bytes"
	"strconv"
	"testing"
)

var stateExtra *StateExtra

func init() {
	stateExtra = NewStateExtra()
}

func TestSetAndGetAndDelete(t *testing.T) {
	stateExtra.set("contractAddr", "Tom", []byte("hello"))
	stateExtra.set("contractAddr", "Marry", []byte("world"))

	stateExtra.delete("contractAddr", "Marry")

	if !bytes.Equal([]byte("hello"), stateExtra.get("contractAddr", "Tom")) {
		t.Error("get result not equal set value ")
	}

	if stateExtra.get("contractAddr", "Marry") != nil {
		t.Error("get result not equal set value ")
	}

}

func TestStateGetByPrefix(t *testing.T) {
	for i := 0; i < 10; i++ {
		stateExtra.set("contractAddr", "Tom_"+strconv.Itoa(i), []byte("hello"+strconv.Itoa(i)))
		stateExtra.set("contractAddr", "Tom_1"+strconv.Itoa(i), []byte("hello_1"+strconv.Itoa(i)))

	}

	values := stateExtra.getByPrefix("contractAddr", "Tom_1")
	for _, v := range values {
		t.Log("key: ", string(v.Key), "value: ", string(v.Value))
	}

}

func TestStateGetByRange(t *testing.T) {
	for i := 0; i < 10; i++ {
		stateExtra.set("contractAddr", "Tom_"+strconv.Itoa(i), []byte("hello"+strconv.Itoa(i)))
	}

	values := stateExtra.getByRange("contractAddr", "Tom_1", "Tom_4")
	for _, v := range values {
		t.Log("key: ", string(v.Key), "value: ", string(v.Value))
	}
}
