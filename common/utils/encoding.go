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

import (
	"bytes"
	"fmt"
	"io"
	"math/big"
	"reflect"
)

// VarEncode encodes the input data
// TODO: val support all types
func VarEncode(w io.Writer, val interface{}) {
	// value - type - kind - Interface()
	s := reflect.ValueOf(val)
	if s.Kind() == reflect.Ptr {
		s = s.Elem()
	}
	recursiveEncode(w, s)
}

func recursiveEncode(w io.Writer, s reflect.Value) {
	switch s.Kind() {
	case reflect.Struct:
		numField := s.NumField()
		for i := 0; i < numField; i++ {
			recursiveEncode(w, s.Field(i))
		}
	case reflect.Ptr:
		ptrEncode(w, s)
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64, reflect.Uintptr:
		WriteVarInt(w, s.Uint())
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		WriteVarInt(w, uint64(s.Int()))
	case reflect.Slice:
		sliceEncode(w, s)
	case reflect.Array:
		arrayEncode(w, s)
	case reflect.String:
		WriteVarInt(w, uint64(len(s.String())))
		w.Write([]byte(s.String()))
	case reflect.Bool:
		if s.Bool() {
			w.Write([]byte{1})
		} else {
			w.Write([]byte{0})
		}
	case reflect.Map:
		mapEncode(w, s)
	default:
		break
	}
}

func ptrEncode(w io.Writer, s reflect.Value) {
	elemValue := s.Elem()
	ptrW := new(bytes.Buffer)

	// Check if the pointer is nil
	if !elemValue.IsValid() {
		WriteVarInt(w, 0)
		return
	}

	v := s.Interface()
	switch v.(type) {
	case *big.Int:
		bigVal := v.(*big.Int)
		// WriteVarInt(w, (uint64)(len(bigVal.Bytes())))
		// w.Write(bigVal.Bytes())
		ptrW.Write(bigVal.Bytes())
	default:
		recursiveEncode(ptrW, elemValue)
	}
	WriteVarInt(w, (uint64)(ptrW.Len()))
	w.Write(ptrW.Bytes())
}

func sliceEncode(w io.Writer, s reflect.Value) {
	vType := s.Type().Elem()
	switch vType.Kind() {
	case reflect.Uint8:
		buf := s.Bytes()
		WriteVarInt(w, (uint64)(len(buf)))
		w.Write(buf)
	default:
		vlen := s.Len()
		WriteVarInt(w, (uint64)(vlen))
		for i := 0; i < vlen; i++ {
			recursiveEncode(w, s.Index(i))
		}
		break
	}
}

func arrayEncode(w io.Writer, v reflect.Value) {
	etp := v.Type().Elem()
	switch etp.Kind() {
	case reflect.Uint8:
		if !v.CanAddr() {
			cpy := reflect.New(v.Type()).Elem()
			cpy.Set(v)
			v = cpy
		}
		size := v.Len()
		slice := v.Slice(0, size).Bytes()

		WriteVarInt(w, (uint64)(size))
		w.Write(slice)
	default:
		// fmt.Println(v.Bytes())
	}
}

func mapEncode(w io.Writer, v reflect.Value) {
	mapW := new(bytes.Buffer)
	for _, key := range v.MapKeys() {
		recursiveEncode(mapW, key)
		recursiveEncode(mapW, v.MapIndex(key))
	}
	WriteVarInt(w, (uint64)(v.Len()))
	w.Write(mapW.Bytes())
}

// VarDecode decodes the data to val, val mustbe pointer
func VarDecode(r io.Reader, val interface{}) error {
	rv := reflect.ValueOf(val)
	if rv.Kind() != reflect.Ptr || rv.IsNil() {
		return fmt.Errorf("vardecode error, val mustbe pointer") //&InvalidVarDecodeError{reflect.TypeOf(v)}
	}
	return recursiveDecode(r, rv.Elem())
}

func recursiveDecode(r io.Reader, s reflect.Value) error {
	var (
		err error
		l   uint64
	)
	switch s.Kind() {
	case reflect.Ptr:
		err = ptrDecode(r, s)
	case reflect.Struct:
		for i := 0; i < s.NumField(); i++ {
			recursiveDecode(r, s.Field(i))
		}
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		if l, err = ReadVarInt(r); l > 0 {
			s.SetUint(l)
		}
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		if l, err = ReadVarInt(r); l > 0 {
			s.SetInt((int64)(l))
		}
	case reflect.Slice:
		// only support []byte,
		// TODO: other types support
		err = sliceDecode(r, s)
	case reflect.Array:
		l, err = ReadVarInt(r)
		if l > 0 && int(l) == s.Len() {
			buf := make([]byte, l)
			n, rErr := io.ReadFull(r, buf)

			if n > 0 && rErr == nil {
				reflect.Copy(s, reflect.ValueOf(buf))
			}
			err = rErr
		} else {
			err = fmt.Errorf("array decode length error")
		}
	case reflect.String:
		if l, err = ReadVarInt(r); l > 0 && s.CanSet() {
			buf := make([]byte, l)
			io.ReadFull(r, buf)
			s.SetString(string(buf))
		}
	case reflect.Bool:
		b := make([]byte, 1)
		if _, err = r.Read(b); b[0] == 1 {
			s.SetBool(true)
		}
	case reflect.Map:
		err = mapDecode(r, s)
	default:
		// fmt.Println("defualt", s)
	}
	return err
}

func sliceDecode(r io.Reader, s reflect.Value) error {
	var (
		l   uint64
		err error
	)
	vType := s.Type().Elem()
	l, err = ReadVarInt(r)

	if l > 0 {
		newVal := reflect.MakeSlice(s.Type(), (int)(l), int(l))

		switch vType.Kind() {
		case reflect.Uint8:
			buf := make([]byte, l)
			n, rErr := io.ReadFull(r, buf)
			if n == int(l) && rErr == nil && s.CanSet() {
				reflect.Copy(newVal, reflect.ValueOf(buf))
				s.Set(newVal)
			}
			err = rErr
		default:
			for i := 0; i < int(l); i++ {
				err = recursiveDecode(r, newVal.Index(i))
			}
			s.Set(newVal)
		}
	}

	return err
}

func ptrDecode(r io.Reader, s reflect.Value) error {
	var (
		l   uint64
		n   int
		err error
	)
	v := s.Interface()
	l, err = ReadVarInt(r)
	switch v.(type) {
	case *big.Int:
		bigVal := new(big.Int)
		if l > 0 {
			buf := make([]byte, l)
			n, err = io.ReadFull(r, buf)
			if n == int(l) && s.CanSet() {
				bigVal.SetBytes(buf)
				s.Set(reflect.ValueOf(bigVal))
			}
		}
		if l == 0 {
			s.Set(reflect.ValueOf(bigVal))
		}
	default:
		if s.Kind() == reflect.Ptr && l > 0 {
			if !s.IsNil() || !s.IsValid() {
				err = recursiveDecode(r, s.Elem())
				break
			}

			val := reflect.New(s.Type().Elem())
			if err = recursiveDecode(r, val.Elem()); err == nil && s.CanSet() {
				s.Set(val)
			}
		}
	}

	return err
}

func mapDecode(r io.Reader, s reflect.Value) error {
	var (
		l   uint64
		err error
	)
	l, err = ReadVarInt(r)
	if l > 0 {
		t := s.Type()
		newVal := reflect.MakeMap(t)
		for i := 0; i < int(l); i++ {
			key := reflect.New(t.Key())
			value := reflect.New(t.Elem())
			err = recursiveDecode(r, key.Elem())
			err = recursiveDecode(r, value.Elem())
			newVal.SetMapIndex(key.Elem(), value.Elem())
		}
		s.Set(newVal)
	}
	return err
}

// Serialize serializes an object to bytes
func Serialize(obj interface{}) []byte {
	buf := new(bytes.Buffer)
	VarEncode(buf, obj)
	return buf.Bytes()
}

// Deserialize deserializes bytes to object
func Deserialize(data []byte, obj interface{}) error {
	buf := bytes.NewBuffer(data)
	return VarDecode(buf, obj)
}
