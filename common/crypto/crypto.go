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
package crypto

import (
	"bytes"
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"encoding/hex"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"math/big"
	"os"

	"crypto/sha256"

	"github.com/ethereum/go-ethereum/common/math"
	"github.com/ethereum/go-ethereum/crypto/secp256k1"
	"github.com/zipper-project/zipper/common/utils"
)

const (
	// SignatureSize represents the signature length
	SignatureSize = 65
)

var (
	// N is secp256k1 N
	N = S256().Params().N
	// halfN is N / 2
	halfN          = new(big.Int).Rsh(N, 1)
	emptySignature = Signature{}
)

type (
	// PrivateKey represents the ecdsa privatekey
	PrivateKey ecdsa.PrivateKey
	// PublicKey represents the ecdsa publickey
	PublicKey ecdsa.PublicKey
	// Signature represents the ecdsa_signcompact signature
	// data format [r - s - v]
	Signature [SignatureSize]byte
)

// Keccak256 calculates and returns the Keccak256 hash of the input data.
func Keccak256(data ...[]byte) []byte {
	//d := sha3.NewKeccak256()
	d := sha256.New()
	for _, b := range data {
		d.Write(b)
	}
	return d.Sum(nil)
}

// S256 returns an instance of the secp256k1 curve
func S256() elliptic.Curve {
	return secp256k1.S256()
}

// GenerateKey returns a random PrivateKey
func GenerateKey() (*PrivateKey, error) {
	priv, err := ecdsa.GenerateKey(S256(), rand.Reader)
	if err != nil {
		return nil, err
	}
	return (*PrivateKey)(priv), err
}

// SecretBytes returns the actual bytes of ecdsa privatekey
func (priv *PrivateKey) SecretBytes() []byte {
	return math.PaddedBigBytes(priv.D, priv.Params().BitSize/8)
}

// Sign signs the hash and returns the signature
func (priv *PrivateKey) Sign(hash []byte) (sig *Signature, err error) {
	if len(hash) != 32 {
		return &emptySignature, fmt.Errorf("hash is required to be exactly 32 bytes (%d)", len(hash))
	}
	secretKey := priv.SecretBytes()
	defer utils.ZeroMemory(secretKey)

	rawSig, err := secp256k1.Sign(hash, secretKey)

	sig = new(Signature)
	sig.SetBytes(rawSig, false)
	return sig, err
}

// Bytes returns the bytes of the signature
func (sig *Signature) Bytes() []byte {
	return sig[:]
}

// Public returns the public key corresponding to priv.
func (priv *PrivateKey) Public() *PublicKey {
	return (*PublicKey)(&priv.PublicKey)
}

// SaveECDSA saves a private key to the given file
func (priv *PrivateKey) SaveECDSA(file string) error {
	ioutil.WriteFile(file, []byte(hex.EncodeToString(priv.SecretBytes())), 0600)
	return nil
}

// LoadECDSA loads a private key from the given file
func LoadECDSA(file string) (*PrivateKey, error) {
	buf := make([]byte, 64)
	fd, err := os.Open(file)
	if err != nil {
		return nil, err
	}
	defer fd.Close()

	if _, err = io.ReadFull(fd, buf); err != nil {
		return nil, err
	}

	return HexToECDSA(string(buf))
}

// HexToECDSA parses a secp256k1 private key
func HexToECDSA(hexkey string) (*PrivateKey, error) {
	b, err := hex.DecodeString(hexkey)
	if err != nil {
		return nil, errors.New("invalid hex string")
	}
	if len(b) != 32 {
		return nil, errors.New("invalid length, need 256 bits")
	}
	return ToECDSA(b), nil
}

// ToECDSA creates a private key with the given D value.
func ToECDSA(prv []byte) *PrivateKey {
	if len(prv) == 0 {
		return nil
	}

	priv := new(ecdsa.PrivateKey)
	priv.PublicKey.Curve = S256()
	priv.D = new(big.Int).SetBytes(prv)
	priv.PublicKey.X, priv.PublicKey.Y = priv.PublicKey.Curve.ScalarBaseMult(prv)
	return (*PrivateKey)(priv)
}

// Bytes returns the ecdsa PublickKey to bytes
func (pub *PublicKey) Bytes() []byte {
	if pub == nil || pub.X == nil || pub.Y == nil {
		return nil
	}
	return elliptic.Marshal(secp256k1.S256(), pub.X, pub.Y)
}

// SetBytes returns a format signature according the raw signature
func (sig *Signature) SetBytes(data []byte, compressed bool) {
	if len(data) == 65 {
		copy(sig[:], data[:])
	}

	sig[64] += 27
	if compressed {
		sig[64] += 4
	}
}

// MarshalText returns the hex representation of h.
func (sig Signature) MarshalText() ([]byte, error) {
	return utils.Bytes(sig[:]).MarshalText()
}

// UnmarshalText parses a hash in hex syntax.
func (sig *Signature) UnmarshalText(input []byte) error {
	return utils.UnmarshalFixedText(input, sig[:])
}

// Ecrecover recovers publick key
func (sig *Signature) Ecrecover(hash []byte) ([]byte, error) {
	data := make([]byte, SignatureSize)
	copy(data[:], sig[:])
	data[64] = (data[64] - 27) & ^byte(4)
	return Ecrecover(hash, data[:])
}

// RecoverPublicKey recovers public key and also verifys the signature
func (sig *Signature) RecoverPublicKey(hash []byte) (*PublicKey, error) {
	s, err := sig.Ecrecover(hash)
	if err != nil {
		return nil, err
	}

	return ToECDSAPub(s), nil
}

// Verify verifys the signature with public key
func (sig *Signature) Verify(hash []byte, pub *PublicKey) bool {
	sigPub, err := sig.Ecrecover(hash)
	if err != nil {
		return false
	}

	return bytes.Equal(sigPub, pub.Bytes())
}

// VRS returns the v r s values
func (sig *Signature) VRS() (v byte, r, s *big.Int) {
	return (sig[64] - 27) & ^byte(4), new(big.Int).SetBytes(sig[:32]), new(big.Int).SetBytes(sig[32:64])
}

// Validate validates whether the signature values are valid
func (sig *Signature) Validate() bool {
	v, r, s := sig.VRS()
	one := big.NewInt(1)

	if r.Cmp(one) < 0 || s.Cmp(one) < 0 {
		return false
	}

	if s.Cmp(halfN) > 0 {
		return false
	}

	return r.Cmp(N) < 0 && s.Cmp(N) < 0 && (v == 0 || v == 1)
}

// SigToPub recovers public key from the input data to the ecdsa public key
func SigToPub(hash, sig []byte) (*PublicKey, error) {
	s, err := Ecrecover(hash, sig)
	if err != nil {
		return nil, err
	}

	return ToECDSAPub(s), nil
}

// Ecrecover recovers publick key
func Ecrecover(hash, sig []byte) ([]byte, error) {
	return secp256k1.RecoverPubkey(hash, sig)
}

// ToECDSAPub returns ecdsa public key according the input data
func ToECDSAPub(pub []byte) *PublicKey {
	if len(pub) == 0 {
		return nil
	}
	x, y := elliptic.Unmarshal(S256(), pub)
	return (*PublicKey)(&ecdsa.PublicKey{Curve: S256(), X: x, Y: y})
}

// ZeroKey clean private key
func ZeroKey(k *PrivateKey) {
	b := k.D.Bits()
	utils.ZeroMemory(b)
}
