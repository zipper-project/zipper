// Copyright (C) 2017, Zipper Team Technology Co.,Ltd.  All rights reserved.
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

import (
	"testing"
	"reflect"
	"runtime"
	"fmt"
)

func AssertSame(t testing.TB, actual interface{}, expected interface{}) {
	t.Logf("%s: AssertSame [%#v] and [%#v]", getCallerInfo(), actual, expected)
	if actual != expected {
		t.Fatalf("Values actual=[%#v] and expected=[%#v] do not point to same object. %s", actual, expected, getCallerInfo())
	}
}

func AssertEquals(t testing.TB, actual interface{}, expected interface{}) {
	t.Logf("%s: AssertEquals [%#v] and [%#v]", getCallerInfo(), actual, expected)
	if expected == nil && isNil(actual) {
		return
	}
	if !reflect.DeepEqual(actual, expected) {
		t.Fatalf("Values are not equal.\n Actual=[%#v], \n Expected=[%#v]\n %s", actual, expected, getCallerInfo())
	}
}

func AssertNotEquals(t testing.TB, actual interface{}, expected interface{}) {
	if reflect.DeepEqual(actual, expected) {
		t.Fatalf("Values are not supposed to be equal. Actual=[%#v], Expected=[%#v]\n %s", actual, expected, getCallerInfo())
	}
}

func isNil(in interface{}) bool {
	return in == nil || reflect.ValueOf(in).IsNil() || (reflect.TypeOf(in).Kind() == reflect.Slice && reflect.ValueOf(in).Len() == 0)
}

func getCallerInfo() string {
	_, file, line, ok := runtime.Caller(2)
	if !ok {
		return "Could not retrieve caller's info"
	}
	return fmt.Sprintf("CallerInfo = [%s:%d]", file, line)
}
