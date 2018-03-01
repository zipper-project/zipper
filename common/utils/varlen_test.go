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
	"fmt"
	"testing"
)

func TestVarInt(t *testing.T) {
	for _, i := range []uint64{0xFF, 0xFFFF, 0xFFFFFF} {
		res := VarInt(uint64(i))
		if i == 0xFF {
			if !bytes.Equal(res, []byte{0xfd, 0xff, 0x00}) {
				t.Errorf("VarInt calculates with 0xFF error, VarInt(%d) = %0x, except %s", i, []byte{0xfd, 0xff, 0x00}, fmt.Sprintf("%0x", res))
			}

			buf := bytes.NewBuffer(res)
			ri, _ := ReadVarInt(buf)
			if uint64(i) != ri {
				t.Errorf("VarInt(%d) != ReadVarInt(%d)", uint64(i), ri)
			}
		}
		if i == 0xFFFF {
			if bytes.Equal(res, []byte{0xfd, 0xff, 0x00}) {
				t.Errorf("VarInt calculates with 0xFF error! except %s", fmt.Sprintf("%0x", res))
			}
			buf := bytes.NewBuffer(res)
			ri, _ := ReadVarInt(buf)
			if uint64(i) != ri {
				t.Errorf("VarInt(%d) != ReadVarInt(%d)", uint64(i), ri)
			}
		}
		if i == 0xFFFFFF {
			if bytes.Equal(res, []byte{0xfd, 0xff, 0x00}) {
				t.Errorf("VarInt calculates with 0xFF error! except %s", fmt.Sprintf("%0x", res))
			}
			buf := bytes.NewBuffer(res)
			ri, _ := ReadVarInt(buf)
			if uint64(i) != ri {
				t.Errorf("VarInt(%d) != ReadVarInt(%d)", uint64(i), ri)
			}
		}
	}
}

func TestVarEncode_string(t *testing.T) {
	buf := new(bytes.Buffer)
	VarEncode(buf, "TestString")

	var res string
	VarDecode(buf, &res)
	if res != "TestString" {
		t.Log("VarEncode on string type Error")
	}
}

// func TestVarArray(t *testing.T) {

// 	var (
// 		testSha256HashStr = "2cf24dba5fb0a30e26e83b2ac5b9e29e1b161e5c1fa7425e73043362938b9824"
// 		testPrivateKey    = "289c2857d4598e37fb9647507e47a309d6133539bf21a8b9cb6df88fd5232032"

// 		buf = new(bytes.Buffer)
// 	)

// 	// Sha256 hash
// 	h := crypto.Sha256([]byte("hello"))
// 	if h.String() != testSha256HashStr {
// 		t.Errorf("Sha256(%s) = %s, except %s !", "hello", testSha256HashStr, h.String())
// 	}

// 	priv, _ := crypto.HexToECDSA(testPrivateKey)
// 	sig, _ := priv.Sign(h[:])

// 	VarEncode(buf, sig)
// 	fmt.Println(buf.Bytes())

// }
