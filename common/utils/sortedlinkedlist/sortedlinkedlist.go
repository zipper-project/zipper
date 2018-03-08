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

package sortedlinkedlist

import (
	"container/list"
	"crypto/sha256"
	"encoding/hex"
	"sync"
)

//IElement interface for element of sortedLinkedList
type IElement interface {
	Compare(v interface{}) int
	Serialize() []byte
}

var sortedLinkedList *SortedLinkedList

func init() {
	sortedLinkedList = NewSortedLinkedList()
}

//NewSortedLinkedList Create sortedLinkedList instance
func NewSortedLinkedList() *SortedLinkedList {
	list := &SortedLinkedList{}
	list.Clear()
	return list
}

//SortedLinkedList Define sortedLinkedList struct
type SortedLinkedList struct {
	list    list.List
	mapping map[string]*list.Element
	sync.RWMutex
}

//Clear Initialize
func Clear() { sortedLinkedList.Clear() }

//Clear Initialize
func (sll *SortedLinkedList) Clear() {
	sll.Lock()
	defer sll.Unlock()
	sll.mapping = make(map[string]*list.Element)
	sll.list.Init()
}

//Len Len of elements
func Len() int { return sortedLinkedList.Len() }

//Len Len of elements
func (sll *SortedLinkedList) Len() int {
	sll.RLock()
	defer sll.RUnlock()
	return sll.list.Len()
}

//Add Add element
func Add(element IElement) { sortedLinkedList.Add(element) }

//Add Add element
func (sll *SortedLinkedList) Add(element IElement) {
	sll.Lock()
	defer sll.Unlock()
	key := sll.key(element)
	if _, ok := sll.mapping[key]; ok {
		return
	}

	// var indexElement *list.Element
	// for elem := sll.list.Front(); elem != nil; elem = elem.Next() {
	// 	if element.Compare(elem.Value.(IElement)) == -1 {
	// 		indexElement = elem
	// 		break
	// 	}
	// }
	// if indexElement != nil {
	// 	sll.mapping[key] = sll.list.InsertBefore(element, indexElement)
	// } else {
	// 	sll.mapping[key] = sll.list.PushBack(element)
	// }

	var indexElement *list.Element
	for elem := sll.list.Back(); elem != nil; elem = elem.Prev() {
		if element.Compare(elem.Value.(IElement)) == 1 {
			indexElement = elem
			break
		}
	}
	if indexElement != nil {
		sll.mapping[key] = sll.list.InsertAfter(element, indexElement)
	} else {
		sll.mapping[key] = sll.list.PushFront(element)
	}
}

//Remove Remove element
func Remove(element IElement) { sortedLinkedList.Remove(element) }

//Remove Remove element
func (sll *SortedLinkedList) Remove(element IElement) {
	sll.Lock()
	defer sll.Unlock()
	key := sll.key(element)
	elem, ok := sll.mapping[key]
	if ok {
		sll.list.Remove(elem)
		delete(sll.mapping, key)
	}
}

//Removes Remove element
func Removes(elements []IElement) { sortedLinkedList.Removes(elements) }

//Removes Remove element
func (sll *SortedLinkedList) Removes(elements []IElement) {
	sll.Lock()
	defer sll.Unlock()
	for _, element := range elements {
		key := sll.key(element)
		elem, ok := sll.mapping[key]
		if ok {
			sll.list.Remove(elem)
			delete(sll.mapping, key)
		}
	}
}

//RemoveBefore Remove elements before element
func RemoveBefore(element IElement) (elements []IElement) {
	return sortedLinkedList.RemoveBefore(element)
}

//RemoveBefore Remove elements before element
func (sll *SortedLinkedList) RemoveBefore(element IElement) (elements []IElement) {
	sll.Lock()
	defer sll.Unlock()
	for elem := sll.list.Front(); elem != nil; elem = elem.Next() {
		if element.Compare(elem.Value.(IElement)) != -1 {
			elements = append(elements, elem.Value.(IElement))
		}
	}
	for _, element := range elements {
		key := sll.key(element)
		elem, _ := sll.mapping[key]
		sll.list.Remove(elem)
		delete(sll.mapping, key)
	}
	return
}

//RemoveAll Remove all elements
func RemoveAll() (elements []IElement) { return sortedLinkedList.RemoveAll() }

//RemoveAll Remove all elements
func (sll *SortedLinkedList) RemoveAll() (elements []IElement) {
	sll.Lock()
	defer sll.Unlock()
	for elem := sll.list.Front(); elem != nil; elem = elem.Next() {
		elements = append(elements, elem.Value.(IElement))
	}
	for _, element := range elements {
		key := sll.key(element)
		elem, _ := sll.mapping[key]
		sll.list.Remove(elem)
		delete(sll.mapping, key)
	}
	return
}

//IterElement Iter, thread safe
func IterElement(function func(element IElement) bool) {
	sortedLinkedList.IterElement(function)
}

//IterElement Iter, thread safe
func (sll *SortedLinkedList) IterElement(function func(element IElement) bool) {
	sll.RLock()
	defer sll.RUnlock()
	for elem := sll.list.Front(); elem != nil; elem = elem.Next() {
		if function(elem.Value.(IElement)) {
			break
		}
	}
}

//Iter Iter, not thread safe
func Iter() func() IElement { return sortedLinkedList.Iter() }

//Iter Iter, not thread safe
func (sll *SortedLinkedList) Iter() func() IElement {
	elem := sll.list.Front()
	return func() IElement {
		if elem != nil {
			element := elem.Value.(IElement)
			elem = elem.Next()
			return element
		}
		return nil
	}
}

func (sll *SortedLinkedList) key(element IElement) string {
	hash := sha256.Sum256(element.Serialize())
	return hex.EncodeToString(hash[:])
}

func (sll *SortedLinkedList) isSorted() bool {
	var pre IElement
	for elem := sll.list.Front(); elem != nil; elem = elem.Next() {
		if pre == nil {
			pre = elem.Value.(IElement)
		} else {
			if pre.Compare(elem.Value.(IElement)) == 1 {
				return false
			}
			pre = elem.Value.(IElement)
		}
	}
	return true
}

//IsExist elemeng if exist
func IsExist(element IElement) bool { return sortedLinkedList.IsExist(element) }

//IsExist elemeng if exist
func (sll *SortedLinkedList) IsExist(element IElement) bool {
	sll.RLock()
	defer sll.RUnlock()
	key := sll.key(element)
	_, ok := sll.mapping[key]
	return ok
}

//RemoveFront remove front
func RemoveFront() { sortedLinkedList.RemoveFront() }

//RemoveFront remove front
func (sll *SortedLinkedList) RemoveFront() IElement {
	sll.Lock()
	defer sll.Unlock()
	elem := sll.list.Front()
	sll.list.Remove(elem)
	key := sll.key(elem.Value.(IElement))
	delete(sll.mapping, key)
	return elem.Value.(IElement)
}

//GetIElementByKey get element
func GetIElementByKey(key string) IElement { return sortedLinkedList.GetIElementByKey(key) }

//GetIElementByKey get element
func (sll *SortedLinkedList) GetIElementByKey(key string) IElement {
	sll.RLock()
	defer sll.RUnlock()
	elem, ok := sll.mapping[key]
	if !ok {
		return nil
	}
	return elem.Value.(IElement)
}
