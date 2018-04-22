package sm2c

/*
#cgo CFLAGS: -I./libsm2
#cgo CFLAGS: -I./openssl/include
// #cgo LDFLAGS: ./openssl/libcrypto.a
#cgo LDFLAGS: -L./openssl -lcrypto
#include "./openssl/include/openssl/ossl_typ.h"
#include "./openssl/include/openssl/crypto.h"
#include "./openssl/crypto/ec/ec_lcl.h"
#include "./libsm2/sm2.c"

extern EC_KEY* SM2_private(const char *zHex);
extern EC_KEY* SM2_public(const char *xHex, const char *yHex);
extern int SM2_sign(int type, const unsigned char *dgst, int dlen, unsigned char *sig, unsigned int *siglen, EC_KEY *eckey);
extern int SM2_verify(int type, const unsigned char *dgst, int dgst_len,const unsigned char *sigbuf, int sig_len, EC_KEY *eckey);
*/
import "C"

import (
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"math/big"
	"unsafe"

	"github.com/ethereum/go-ethereum/common/math"
	"github.com/zipper-project/zipper/common/crypto/crypter"
)

// Curve returns an instance of the sm2 curve
func P256() elliptic.Curve {
	// See http://www.oscca.gov.cn/UpFile/2010122214836668.pdf
	p256 := &elliptic.CurveParams{Name: "SM2-P-256"}
	p256.P, _ = new(big.Int).SetString("FFFFFFFEFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFF00000000FFFFFFFFFFFFFFFF", 16)
	p256.N, _ = new(big.Int).SetString("FFFFFFFEFFFFFFFFFFFFFFFFFFFFFFFF7203DF6B21C6052B53BBF40939D54123", 16)
	p256.B, _ = new(big.Int).SetString("28E9FA9E9D9F5E344D5A9E4BCF6509A7F39789F515AB8F92DDBCBD414D940E93", 16)
	p256.Gx, _ = new(big.Int).SetString("32C4AE2C1F1981195F9904466A39C9948FE30BBFF2660BE1715A4589334C74C7", 16)
	p256.Gy, _ = new(big.Int).SetString("BC3736A2F4F6779C59BDCEE36B692153D0A9877CC62A474002DF32E52139F0A0", 16)
	p256.BitSize = 256
	return p256
}

type PrivateKey ecdsa.PrivateKey

type PublicKey ecdsa.PublicKey

func (priv *PrivateKey) Bytes() []byte {
	return math.PaddedBigBytes(priv.D, priv.Params().BitSize/8)
}

func (priv *PrivateKey) Public() crypter.IPublicKey {
	return (*PublicKey)(((*ecdsa.PrivateKey)(priv)).Public().(*ecdsa.PublicKey))
}

func (pub *PublicKey) Bytes() []byte {
	return elliptic.Marshal(P256(), pub.X, pub.Y)
}

type Crypter struct {
}

func (this *Crypter) Name() string {
	return "sm2c_double256"
}

func (this *Crypter) GenerateKey() (crypter.IPrivateKey, crypter.IPublicKey, error) {
	private, err := ecdsa.GenerateKey(P256(), rand.Reader)
	if err != nil {
		return nil, nil, err
	}
	return (*PrivateKey)(private), (*PublicKey)(private.Public().(*ecdsa.PublicKey)), nil
}

func (this *Crypter) Sign(privateKey crypter.IPrivateKey, message []byte) ([]byte, error) {
	sign := this.DoubleSha256(message)
	zHex := C.CString(hex.EncodeToString(privateKey.Bytes()))
	defer C.free(unsafe.Pointer(zHex))
	var eckey *C.EC_KEY = C.SM2_private(zHex)
	defer C.free(unsafe.Pointer(eckey))
	var sigLen C.uint = C.uint(C.ECDSA_size(eckey))
	signature := make([]byte, sigLen)
	if C.SM2_sign(0, (*C.uchar)(unsafe.Pointer(&sign[0])), C.int(len(sign)), (*C.uchar)(unsafe.Pointer(&signature[0])), &sigLen, eckey) == 0 {
		return nil, nil
	}
	return signature, nil
}

func (this *Crypter) Verify(publicKey crypter.IPublicKey, message, sig []byte) bool {
	sign := this.DoubleSha256(message)
	pub := publicKey.(*PublicKey)
	xHex := C.CString(hex.EncodeToString(pub.X.Bytes()))
	defer C.free(unsafe.Pointer(xHex))
	yHex := C.CString(hex.EncodeToString(pub.Y.Bytes()))
	defer C.free(unsafe.Pointer(yHex))
	var eckey *C.EC_KEY = C.SM2_public(xHex, yHex)
	defer C.free(unsafe.Pointer(eckey))
	return C.SM2_verify(0, (*C.uchar)(unsafe.Pointer(&sign[0])), C.int(len(sign)), (*C.uchar)(unsafe.Pointer(&sig[0])), C.int(len(sig)), eckey) == 1
}

func (this *Crypter) ToPrivateKey(data []byte) crypter.IPrivateKey {
	priv := new(PrivateKey)
	priv.PublicKey.Curve = P256()
	priv.D = new(big.Int).SetBytes(data)
	priv.PublicKey.X, priv.PublicKey.Y = priv.PublicKey.Curve.ScalarBaseMult(priv.D.Bytes())
	return priv
}

func (this *Crypter) ToPublicKey(data []byte) crypter.IPublicKey {
	x, y := elliptic.Unmarshal(P256(), data)
	return &PublicKey{
		Curve: P256(),
		X:     x,
		Y:     y,
	}
}

func (this *Crypter) DoubleSha256(data []byte) []byte {
	h := sha256.Sum256(data)
	h = sha256.Sum256(h[:])
	return h[:]
}
