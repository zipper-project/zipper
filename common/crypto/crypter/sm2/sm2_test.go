package sm2

import (
	"crypto/rand"
	"crypto/sha256"
	"testing"
)

func TestKeyGeneration(t *testing.T) {
	p256 := P256()
	priv, err := GenerateKey(p256, rand.Reader)
	if err != nil {
		t.Errorf("error: %s", err)
		return
	}

	if !p256.IsOnCurve(priv.PublicKey.X, priv.PublicKey.Y) {
		t.Errorf("public key invalid: %s", err)
	}
}

func TestSignAndVerify(t *testing.T) {
	p256 := P256()
	priv, _ := GenerateKey(p256, rand.Reader)

	msg := []byte("testing")

	dig := sha256.Sum256(msg)
	hashed := dig[:]
	r, s, err := Sign(rand.Reader, priv, hashed)
	if err != nil {
		t.Errorf("error signing: %s", err)
		return
	}

	if !Verify(&priv.PublicKey, hashed, r, s) {
		t.Errorf("Verify failed")
	}

	msg[0] ^= 0xff
	dig = sha256.Sum256(msg)
	hashed = dig[:]
	if Verify(&priv.PublicKey, hashed, r, s) {
		t.Errorf("Verify always works!")
	}
}
