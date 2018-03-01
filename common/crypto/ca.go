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
	"crypto"
	"crypto/rand"
	"crypto/rsa"
	"crypto/sha256"
	"crypto/x509"
	"crypto/x509/pkix"
	"encoding/pem"
	"fmt"
	"math/big"
	rd "math/rand"
	"time"
)

func init() {
	rd.Seed(time.Now().UnixNano())
}

type CertInformation struct {
	Country            []string
	Organization       []string
	OrganizationalUnit []string
	EmailAddress       []string
	Province           []string
	Locality           []string
	CommonName         string
	IsCA               bool
	Names              []pkix.AttributeTypeAndValue
}

func Parse(cert, key []byte) (rootCertificate *x509.Certificate, rootPrivateKey *rsa.PrivateKey, err error) {
	rootCertificate, err = ParseCrt(cert)
	if err != nil {
		return
	}
	rootPrivateKey, err = ParseKey(key)
	return
}

func ParseCrt(data []byte) (*x509.Certificate, error) {
	p, _ := pem.Decode(data)
	return x509.ParseCertificate(p.Bytes)
}

func ParseKey(data []byte) (*rsa.PrivateKey, error) {
	p, _ := pem.Decode(data)
	return x509.ParsePKCS1PrivateKey(p.Bytes)
}

func GenerateRootCertificateBytes(rootCertificate *x509.Certificate, rootPrivateKey *rsa.PrivateKey) ([]byte, error) {
	cert, err := x509.CreateCertificate(rand.Reader, rootCertificate, rootCertificate, &rootPrivateKey.PublicKey, rootPrivateKey)
	if err != nil {
		return nil, fmt.Errorf("create root cert error: %s", err)
	}

	buffer := new(bytes.Buffer)

	if err := pem.Encode(buffer, &pem.Block{Bytes: cert, Type: "CERTIFICATE"}); err != nil {
		return nil, err
	}
	return buffer.Bytes(), nil
}

func GeneratePrivateKeyBytes(privateKey *rsa.PrivateKey) ([]byte, error) {
	buffer := new(bytes.Buffer)
	key := x509.MarshalPKCS1PrivateKey(privateKey)
	if err := pem.Encode(buffer, &pem.Block{Bytes: key, Type: "PRIVATE KEY"}); err != nil {
		return nil, err
	}
	return buffer.Bytes(), nil
}

// VerifyCertificate use root Certificate  to verify remote Certificate
func VerifyCertificate(rootCertificate, Certificate *x509.Certificate) error {
	return Certificate.CheckSignatureFrom(rootCertificate)
}

func SignRsa(privateKey *rsa.PrivateKey, data []byte) ([]byte, error) {
	hashed := sha256.Sum256(data)
	return rsa.SignPKCS1v15(rand.Reader, privateKey, crypto.SHA256, hashed[:])
}

func VerifySign(hashed [sha256.Size]byte, sign []byte, Certificate *x509.Certificate) error {
	return rsa.VerifyPKCS1v15(Certificate.PublicKey.(*rsa.PublicKey), crypto.SHA256, hashed[:], sign)
}

func NewCertificate(info CertInformation) *x509.Certificate {
	return &x509.Certificate{
		SerialNumber: big.NewInt(rd.Int63()),
		Subject: pkix.Name{
			Country:            info.Country,
			Organization:       info.Organization,
			OrganizationalUnit: info.OrganizationalUnit,
			Province:           info.Province,
			CommonName:         info.CommonName,
			Locality:           info.Locality,
			ExtraNames:         info.Names,
		},
		NotBefore:             time.Now(),
		NotAfter:              time.Now().AddDate(20, 0, 0),
		BasicConstraintsValid: true,
		IsCA:           info.IsCA,
		ExtKeyUsage:    []x509.ExtKeyUsage{x509.ExtKeyUsageClientAuth, x509.ExtKeyUsageServerAuth},
		KeyUsage:       x509.KeyUsageDigitalSignature | x509.KeyUsageCertSign,
		EmailAddresses: info.EmailAddress,
	}
}
