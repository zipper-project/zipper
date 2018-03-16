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

package keystore

import (
	crand "crypto/rand"
	"errors"
	"os"
	"path/filepath"
	"strings"
	"sync"

	"github.com/zipper-project/zipper/common/crypto"
	"github.com/zipper-project/zipper/common/db"
	"github.com/zipper-project/zipper/common/utils"

	"github.com/zipper-project/zipper/account"
	pb "github.com/zipper-project/zipper/proto"
)

var (
	ErrNoMatch = errors.New("no key for given address or file")
	ErrDecrypt = errors.New("could not decrypt key with given passphrase")
)

var columnFamily = "account"
var KeyStoreScheme = "keystore"
var ksInstance *KeyStore
var ksPInstance *KeyStore
var once sync.Once

const (
	ScryptN = 2
	ScryptP = 1
)

// KeyStore definition
type KeyStore struct {
	storage keyStore
	db      *db.BlockchainDB
	ksDir   string
}

// NewKeyStore new a KeyStore instance
func NewKeyStore(db *db.BlockchainDB, keydir string, scryptN, scryptP int) *KeyStore {
	once.Do(func() {
		keydir, err := filepath.Abs(keydir)
		if err != nil {
			panic(err)
		}
		ksInstance = &KeyStore{storage: &keyStorePassphrase{keydir, scryptN, scryptP}}
		ksInstance.db = db
		ksInstance.ksDir = keydir
	})
	return ksInstance
}

// NewPlaintextKeyStore new a PlaintextKeyStore instance
func NewPlaintextKeyStore(db *db.BlockchainDB, keydir string) *KeyStore {
	once.Do(func() {
		keydir, err := filepath.Abs(keydir)
		if err != nil {
			panic(err)
		}
		ksPInstance = &KeyStore{storage: &keyStorePlain{keydir}}
		ksPInstance.db = db
		ksPInstance.ksDir = keydir
	})
	return ksPInstance
}

// HasAddress returns if current node has the specified addr
func (ks *KeyStore) HasAddress(addr account.Address) bool {
	a, _ := ks.db.Get(columnFamily, addr.Bytes())
	if len(a) == 0 {
		return false
	}
	return true
}

func (ks *KeyStore) Find(addr account.Address) *account.Account {
	var account account.Account
	a, _ := ks.db.Get(columnFamily, addr.Bytes())
	if len(a) == 0 {
		return &account
	}
	account.Deserialize(a)
	return &account
}

func (ks *KeyStore) Accounts() ([]string, error) {
	var res []string
	err := filepath.Walk(ks.ksDir, func(path string, f os.FileInfo, err error) error {
		if f == nil {
			return err
		}
		if f.IsDir() {
			return nil
		}
		hexStr := strings.Split(f.Name(), "--")[2]
		address := account.HexToAddress(hexStr) //account.NewAddress([]byte(hexStr[2]))
		res = append(res, address.String())

		return nil
	})
	if err != nil {
		return res, err
	}
	return res, err
}

// NewAccount creates a new account
func (ks *KeyStore) NewAccount(passphrase string, accountType uint32) (account.Account, error) {
	_, a, err := storeNewKey(ks.storage, crand.Reader, passphrase)
	if err != nil {
		return account.Account{}, err
	}
	a.AccountType = accountType
	err = ks.db.Put(columnFamily, a.Address.Bytes(), a.Serialize())
	if err != nil {
		return account.Account{}, err
	}
	return a, nil
}

// Delete removes the speciified account
func (ks *KeyStore) Delete(a account.Account, passphrase string) error {
	a, key, err := ks.getDecryptedKey(a, passphrase)
	if key != nil {
		crypto.ZeroKey((*crypto.PrivateKey)(key.PrivateKey))
	}
	if err != nil {
		return err
	}
	err = os.Remove(a.URL.Path)
	if err != nil {
		return err
	}
	err = ks.db.Delete(columnFamily, a.Address.Bytes())
	return err
}

// Update update the specified account
func (ks *KeyStore) Update(a account.Account, passphrase, newPassphrase string) error {
	a, key, err := ks.getDecryptedKey(a, passphrase)
	if err != nil {
		return err
	}
	return ks.storage.StoreKey(a.URL.Path, key, newPassphrase)
}

// SignTx sign the specified transaction
func (ks *KeyStore) SignTx(a account.Account, tx *pb.Transaction, pass string) (*pb.Transaction, error) {
	_, key, err := ks.getDecryptedKey(a, pass)
	if err != nil {
		return nil, err
	}

	genTxSender(tx, key.PrivateKey.Public())

	sig, err := key.PrivateKey.Sign(tx.Hash().Bytes())
	if err != nil {
		return nil, err
	}
	tx.GetHeader().Signature = sig.Bytes()
	return tx, nil
}

func genTxSender(tx *pb.Transaction, publicKey *crypto.PublicKey) {
	//generated sender address by PublicKey
	tx.GetHeader().Sender = account.PublicKeyToAddress(*publicKey).String()

	//contract init transaction generated contract address  by sender address and contract code
	if tx.GetType() == pb.TransactionType_LuaContractInit || tx.GetType() == pb.TransactionType_JSContractInit {
		contractSpec := new(pb.ContractSpec)
		utils.Deserialize(tx.Payload, contractSpec)
		var a account.Address
		pubBytes := []byte(tx.GetHeader().Sender + string(contractSpec.Code))
		a.SetBytes(crypto.Keccak256(pubBytes[1:])[12:])
		contractSpec.Addr = a.Bytes()
		tx.ContractSpec = contractSpec

		//generated recipient address by contract address
		tx.GetHeader().Recipient = account.NewAddress(a.Bytes()).String()
	}
}

// SignHashWithPassphrase signs hash if the private key matching the given address
// can be decrypted with the given passphrase. The produced signature is in the
// [R || S || V] format where V is 0 or 1.
func (ks *KeyStore) SignHashWithPassphrase(a account.Account, passphrase string, hash []byte) (signature []byte, err error) {
	_, key, err := ks.getDecryptedKey(a, passphrase)
	if err != nil {
		return []byte{}, err
	}
	defer crypto.ZeroKey(key.PrivateKey)
	sig, err := key.PrivateKey.Sign(hash)
	if err != nil {
		return []byte{}, err
	}
	return sig.Bytes(), nil
}

func (ks *KeyStore) getDecryptedKey(a account.Account, auth string) (account.Account, *Key, error) {
	addr := account.PublicKeyToAddress(*a.PublicKey)
	key, err := ks.storage.GetKey(addr, a.URL.Path, auth)
	return a, key, err
}
