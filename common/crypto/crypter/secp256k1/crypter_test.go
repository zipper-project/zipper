package secp256k1

import (
	"fmt"
	"testing"
)

func TestSecp256k1_1(t *testing.T) {
	crypter := &Crypter{}
	priv, pub, err := crypter.GenerateKey()
	if err != nil {
		panic(err)
	}
	fmt.Println("org:", priv, pub)
	fmt.Println("trs:", crypter.ToPrivateKey(priv.Bytes()), crypter.ToPublicKey(pub.Bytes()))
}

func TestSecp256k1_2(t *testing.T) {
	crypter := &Crypter{}
	priv, pub, err := crypter.GenerateKey()
	if err != nil {
		panic(err)
	}

	msg := []byte("testing")
	sign, err := crypter.Sign(priv, msg)
	if err != nil {
		panic(err)
	}

	if !crypter.Verify(pub, msg, sign) {
		panic("err")
	}
}
