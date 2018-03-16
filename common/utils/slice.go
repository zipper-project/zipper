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

package utils

import "reflect"

func Contain(obj interface{}, target interface{}) bool {
	targetValue := reflect.ValueOf(target)
	switch reflect.TypeOf(target).Kind() {
	case reflect.Slice, reflect.Array:
		for i := 0; i < targetValue.Len(); i++ {
			if targetValue.Index(i).Interface() == obj {
				return true
			}
		}
	case reflect.Map:
		if targetValue.MapIndex(reflect.ValueOf(obj)).IsValid() {
			return true
		}
	}
	return false
}

func DelStringFromSlice(val string, slice *[]string) {
	s := make([]string, len(*slice)-1)
	for n, str := range *slice {
		if str == val {
			copy(s[:], (*slice)[:n])
			copy(s[n:], (*slice)[n+1:])
		}
	}
	*slice = s
}

// ZeroMemory erases a slice memory
func ZeroMemory(s interface{}) {
	if v, ok := s.([]interface{}); ok {
		for i := range v {
			v[i] = 0
		}
	}
}
