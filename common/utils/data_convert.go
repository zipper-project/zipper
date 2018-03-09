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

import (
	"bytes"
	"encoding/binary"
	"encoding/hex"
	"fmt"
	"io"
	"math"
	"time"

	"github.com/golang/protobuf/proto"
)

func MinimizeSilce(src []byte) []byte {
	dest := make([]byte, len(src))
	copy(dest, src)
	return dest
}

func BytesToHex(byteSlice []byte) string {
	return hex.EncodeToString(byteSlice)
}

func HexToBytes(s string) []byte {
	h, _ := hex.DecodeString(s)
	return h
}

func Uint32ToBytes(src uint32) []byte {
	buf := new(bytes.Buffer)
	binary.Write(buf, binary.LittleEndian, src)
	return buf.Bytes()
}

func BytesToUint32(src []byte) uint32 {
	return binary.LittleEndian.Uint32(src)
}

func CurrentTimestamp() uint32 {
	return uint32(time.Now().Unix())
}

func Uint32ArrayToBytes(arrary []uint32) []byte {
	var (
		buf = new(bytes.Buffer)
	)
	VarEncode(buf, (uint64)(len(arrary)))
	for _, v := range arrary {
		VarEncode(buf, (uint64)(len(Uint32ToBytes(v))))
		buf.Write(Uint32ToBytes(v))
	}
	return buf.Bytes()
}

func BytesToUint32Arrary(src []byte) []uint32 {
	var (
		array = make([]uint32, 0)
	)
	r := bytes.NewBuffer(src)
	len, _ := ReadVarInt(r)
	for i := uint64(0); i < len; i++ {
		uint32Len, _ := ReadVarInt(r)
		buf := make([]byte, uint32Len)
		io.ReadFull(r, buf)
		array = append(array, BytesToUint32(buf))
	}
	return array
}

type Times []uint32

func (t Times) Len() int           { return len(t) }
func (t Times) Swap(i, j int)      { t[i], t[j] = t[j], t[i] }
func (t Times) Less(i, j int) bool { return t[i] < t[j] }

func DecodeUint32(bytes []byte, cnt uint32) ([]uint32, error) {
	ret := make([]uint32, cnt)
	b := proto.NewBuffer(bytes)
	for index := uint32(0); index < cnt; index++ {
		number, err := b.DecodeVarint()
		if err != nil {
			return nil, fmt.Errorf("Failed to decode uint64 --- %s", err)
		}
		ret[index] = uint32(number)
	}
	return ret, nil
}

func Float64ToByte(src float64) []byte {
	bits := math.Float64bits(src)
	bytes := make([]byte, 8)
	binary.LittleEndian.PutUint64(bytes, bits)
	return bytes
}

func ByteToFloat64(src []byte) float64 {
	bits := binary.LittleEndian.Uint64(src)
	return math.Float64frombits(bits)
}

// func Int64ToByte(src int64) []byte {
// 	buf := new(bytes.Buffer)
// 	binary.Write(buf, binary.LittleEndian, src)
// 	return buf.Bytes()
// }

// func ByteToInt64(src []byte) int64 {

// 	return
// }
