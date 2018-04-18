package crypter

import (
	"fmt"
)

// IPrivateKey defines the interface of privatekey.
type IPrivateKey interface {
	Public() IPublicKey
	Bytes() []byte
}

// IPublicKey defines the interface of publickey.
type IPublicKey interface {
	Bytes() []byte
}

type ICrypter interface {
	Name() string
	GenerateKey() (IPrivateKey, IPublicKey, error)
	Sign(privateKey IPrivateKey, message []byte) ([]byte, error)
	Verify(publicKey IPublicKey, message, signature []byte) bool

	ToPublicKey(data []byte) IPublicKey
	ToPrivateKey(data []byte) IPrivateKey
}

var crypters = make(map[string]ICrypter)

func RegisterCrypter(name string, crypter ICrypter) {
	if _, ok := crypters[name]; !ok {
		crypters[name] = crypter
		return
	}
	panic(fmt.Sprintf("crypter %s already registered", name))
}

func UnRegisterCrypter(name string) {
	delete(crypters, name)
}

func Crypter(name string) (ICrypter, error) {
	if crypter, ok := crypters[name]; ok {
		return crypter, nil
	}
	return nil, fmt.Errorf("unsppoort crypter %s", name)
}

func MustCrypter(name string) ICrypter {
	if crypter, ok := crypters[name]; ok {
		return crypter
	}
	panic(fmt.Errorf("unsppoort crypter %s", name))
}
