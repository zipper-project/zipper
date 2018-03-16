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

package sortedlinkedlist

import (
	"fmt"
	"testing"
)

type INT int

func (e INT) Compare(v interface{}) int {
	if e < v.(INT) {
		return -1
	} else if e > v.(INT) {
		return 1
	}
	return 0
}

func (e INT) Serialize() []byte {
	str := fmt.Sprintf("%d", e)
	return []byte(str)
}

func Test(t *testing.T) {
	list := NewSortedLinkedList()
	list.Clear()

	list.Add(INT(10))
	list.Add(INT(8))
	list.Add(INT(14))
	list.Add(INT(5))
	list.Add(INT(7))

	fmt.Println(list.isSorted())
	fmt.Println("len:", list.Len())

	next := list.Iter()
	for elem := next(); elem != nil; elem = next() {
		fmt.Println(elem.(INT))
	}

	fmt.Println(len(list.RemoveBefore(INT(12))))
	fmt.Println("len:", list.Len())

	next = list.Iter()
	for elem := next(); elem != nil; elem = next() {
		fmt.Println(elem.(INT))
	}
}
