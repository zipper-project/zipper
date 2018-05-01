package crypto

import (
	"encoding/hex"
	"io"
	"io/ioutil"
	"os"

	"github.com/zipper-project/zipper/common/crypto/crypter"
)

func SaveCrypter(file string, name string, priv crypter.IPrivateKey) error {
	privHex := hex.EncodeToString(priv.Bytes())
	ioutil.WriteFile(file, []byte(privHex), 0600)
	return nil
}

func LoadCrypter(file string, name string) (crypter.IPrivateKey, error) {
	buf := make([]byte, 64)
	fd, err := os.Open(file)
	if err != nil {
		return nil, err
	}
	defer fd.Close()

	if _, err = io.ReadFull(fd, buf); err != nil {
		return nil, err
	}

	privBytes, err := hex.DecodeString(string(buf))
	if err != nil {
		return nil, err
	}

	crypter, err := crypter.Crypter(name)
	if err != nil {
		return nil, err
	}
	return crypter.ToPrivateKey(privBytes), nil
}

// // Keccak256 calculates and returns the Keccak256 hash of the input data.
// func Keccak256(data ...[]byte) []byte {
// 	//d := sha3.NewKeccak256()
// 	d := sha256.New()
// 	for _, b := range data {
// 		d.Write(b)
// 	}
// 	return d.Sum(nil)
// }
