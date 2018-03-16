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
package crypto

import (
	"bytes"
	"encoding/hex"
	"testing"
)

const (
	testPrivateKey = "289c2857d4598e37fb9647507e47a309d6133539bf21a8b9cb6df88fd5232032"
)

func TestSign(t *testing.T) {
	priv, _ := HexToECDSA(testPrivateKey)
	msg := Sha256([]byte("hello"))

	sig, err := priv.Sign(msg[:])
	if err != nil {
		t.Errorf("Sign Error %s", err)
	}

	if !sig.Validate() {
		t.Errorf("Signature Validate error %s", err)
	}

	pub, err := sig.RecoverPublicKey(msg[:])
	if err != nil {
		t.Errorf("SigToPub Error %v - %v - %s", sig, pub, err)
	}

	pub2 := priv.Public()
	if !bytes.Equal(pub.Bytes(), pub2.Bytes()) {
		t.Errorf("public key not match! %0x - %0x ", pub.Bytes(), pub2.Bytes())
	}
}

func TestHash(t *testing.T) {
	var (
		testSha256HashStr    = "2cf24dba5fb0a30e26e83b2ac5b9e29e1b161e5c1fa7425e73043362938b9824"
		testRipemd160HashStr = "8eb208f7e05d987a9b044a8e98c6b087f15a0bfc"
	)
	// Ripemd160 hash
	h := Ripemd160([]byte("abc"))
	if hex.EncodeToString(h) != testRipemd160HashStr {
		t.Errorf("Rimped160(%s) = %s, except %s !", "abc", testRipemd160HashStr, hex.EncodeToString(h))
	}

	// Sha256 hash
	h = Sha256([]byte("hello")).Bytes()
	hash := NewHash(h)
	if hash.String() != testSha256HashStr {
		t.Errorf("Sha256(%s) = %s, except %s !", "hello", testSha256HashStr, hash.String())
	}

	// Setbytes
	h2 := Sha256([]byte("hello"))
	hash2 := new(Hash)
	hash2.SetBytes(h2[:])
	if hash2.String() != testSha256HashStr {
		t.Errorf("Sha256(%s) = %s, except %s !", "hello", testSha256HashStr, hash.String())
	}
}

func TestLoadAndSaveECDSA(t *testing.T) {
	priv, _ := HexToECDSA(testPrivateKey)
	priv.SaveECDSA("nodekey")
	priv2, _ := LoadECDSA("nodekey")
	if !bytes.Equal(priv.SecretBytes(), priv2.SecretBytes()) {
		t.Errorf("load and save private key error %s != %s", priv, priv2)
	}
}

func BenchmarkSign(b *testing.B) {
	for i := 0; i < b.N; i++ {
		priv, _ := HexToECDSA(testPrivateKey)
		msg := Sha256([]byte("hello"))

		sig, err := priv.Sign(msg[:])
		if err != nil {
			b.Errorf("Sign Error %s", err)
		}

		if !sig.Validate() {
			b.Errorf("Signature Validate error %s", err)
		}

		pub, err := sig.RecoverPublicKey(msg[:])
		if err != nil {
			b.Errorf("SigToPub Error %v - %v - %s", sig, pub, err)
		}

		pub2 := priv.Public()
		if !bytes.Equal(pub.Bytes(), pub2.Bytes()) {
			b.Errorf("public key not match! %0x - %0x ", pub.Bytes(), pub2.Bytes())
		}

	}
}

func BenchmarkSha256(b *testing.B) {
	var (
		testSha256HashStr = "2cf24dba5fb0a30e26e83b2ac5b9e29e1b161e5c1fa7425e73043362938b9824"
	)

	for i := 0; i < b.N; i++ {
		h := Sha256([]byte("hello")).Bytes()
		hash := NewHash(h)
		if hash.String() != testSha256HashStr {
			b.Errorf("Sha256(%s) = %s, except %s !", "hello", testSha256HashStr, hash.String())
		}
	}

}

func BenchmarkRipemd160(b *testing.B) {
	var (
		testRipemd160HashStr = "8eb208f7e05d987a9b044a8e98c6b087f15a0bfc"
	)

	for i := 0; i < b.N; i++ {
		// Ripemd160 hash
		h := Ripemd160([]byte("abc"))
		if hex.EncodeToString(h) != testRipemd160HashStr {
			b.Errorf("Rimped160(%s) = %s, except %s !", "abc", testRipemd160HashStr, hex.EncodeToString(h))
		}
	}

}
