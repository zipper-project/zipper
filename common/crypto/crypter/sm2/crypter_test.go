package sm2

import (
	"encoding/hex"
	"fmt"
	"testing"
)

func TestSM2_1(t *testing.T) {
	crypter := &Crypter{}
	priv, pub, err := crypter.GenerateKey()
	if err != nil {
		panic(err)
	}
	fmt.Println("org:", priv, pub)
	fmt.Println("trs:", crypter.ToPrivateKey(priv.Bytes()), crypter.ToPublicKey(pub.Bytes()))

}
func TestSM2_2(t *testing.T) {
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

	fmt.Println("priv", hex.EncodeToString(priv.Bytes()))
	fmt.Println("pub ", hex.EncodeToString(pub.Bytes()))
	fmt.Println("sign ", hex.EncodeToString(sign))
	if !crypter.Verify(pub, msg, sign) {
		panic("err")
	}
}
