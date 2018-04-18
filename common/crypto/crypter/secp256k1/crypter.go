package secp256k1

import (
	"bytes"
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"crypto/sha256"
	"math/big"

	"github.com/ethereum/go-ethereum/common/math"
	"github.com/ethereum/go-ethereum/crypto/secp256k1"
	"github.com/zipper-project/zipper/common/crypto/crypter"
)

type PrivateKey ecdsa.PrivateKey

type PublicKey ecdsa.PublicKey

func (priv *PrivateKey) Bytes() []byte {
	return math.PaddedBigBytes(priv.D, priv.Params().BitSize/8)
}

func (priv *PrivateKey) Public() crypter.IPublicKey {
	return (*PublicKey)(((*ecdsa.PrivateKey)(priv)).Public().(*ecdsa.PublicKey))
}

func (pub *PublicKey) Bytes() []byte {
	return elliptic.Marshal(S256(), pub.X, pub.Y)
}

type Crypter struct {
}

func S256() elliptic.Curve {
	return secp256k1.S256()
}

func (this *Crypter) Name() string {
	return "secp256k1"
}

func (this *Crypter) GenerateKey() (crypter.IPrivateKey, crypter.IPublicKey, error) {
	private, err := ecdsa.GenerateKey(S256(), rand.Reader)
	if err != nil {
		return nil, nil, err
	}
	return (*PrivateKey)(private), (*PublicKey)(private.Public().(*ecdsa.PublicKey)), nil
}

func (this *Crypter) Sign(privateKey crypter.IPrivateKey, message []byte) ([]byte, error) {
	hash := this.DoubleSha256(message)
	sig, err := secp256k1.Sign(hash, privateKey.Bytes())
	if err != nil {
		return nil, err
	}
	sig[64] += 27
	return sig, nil
}

func (this *Crypter) Verify(publicKey crypter.IPublicKey, message, sig []byte) bool {
	hash := this.DoubleSha256(message)
	data := make([]byte, len(sig))
	copy(data[:], sig[:])
	data[64] = (data[64] - 27) & ^byte(4)
	sigPub, err := secp256k1.RecoverPubkey(hash, data)
	if err != nil {
		panic(err)
	}
	return bytes.Equal(sigPub, publicKey.Bytes())
}

func (this *Crypter) ToPrivateKey(data []byte) crypter.IPrivateKey {
	priv := new(PrivateKey)
	priv.PublicKey.Curve = S256()
	priv.D = new(big.Int).SetBytes(data)
	priv.PublicKey.X, priv.PublicKey.Y = priv.PublicKey.Curve.ScalarBaseMult(priv.D.Bytes())
	return priv
}

func (this *Crypter) ToPublicKey(data []byte) crypter.IPublicKey {
	x, y := elliptic.Unmarshal(S256(), data)
	return &PublicKey{
		Curve: S256(),
		X:     x,
		Y:     y,
	}
}

func (this *Crypter) DoubleSha256(data []byte) []byte {
	h := sha256.Sum256(data)
	h = sha256.Sum256(h[:])
	return h[:]
}
