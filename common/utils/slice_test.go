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

package utils

import "testing"

func TestSlice(t *testing.T) {
	arr := []string{"Test1", "Test2", "Test3", "Test4"}
	arr1 := []string{}
	arr2 := []string{""}
	if Contain("Test", arr) {
		t.Error("error")
	}
	DelStringFromSlice("Test1", &arr)

	t.Log(arr)

	if len(arr) != 3 {
		t.Log("len ", len(arr))
		t.Error("error")
	}

	if Contain("Test1", arr) {
		t.Error("error")
	}
	if Contain("Test", arr1) {
		t.Error("error")
	}
	if !Contain("", arr2) {
		t.Error("error")
	}

}
