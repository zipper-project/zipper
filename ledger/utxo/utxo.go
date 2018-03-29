package utxo

import (
	"github.com/zipper-project/zipper/common/crypto"
	"github.com/zipper-project/zipper/common/db"
)

type Utxo struct {
	dbHandler        *db.BlockchainDB
	utxoColumnFamily string
}

func NewUtxo(db *db.BlockchainDB) *Utxo {
	return &Utxo{
		dbHandler:        db,
		utxoColumnFamily: "utxo",
	}
}

func (u *Utxo) GetUtxoEntryBySet(txSet map[crypto.Hash]struct{}) (map[crypto.Hash][]byte, error) {
	utxos := make(map[crypto.Hash][]byte)
	for hash := range txSet {
		value, err := u.dbHandler.Get(u.utxoColumnFamily, hash.Bytes())
		if err != nil {
			return nil, err
		}
		utxos[hash] = value
	}
	return utxos, nil
}

func (u *Utxo) GetUtxoEntryByHash(hash crypto.Hash) ([]byte, error) {
	value, err := u.dbHandler.Get(u.utxoColumnFamily, hash.Bytes())
	if err != nil {
		return nil, err
	}
	return value, nil
}

func (u *Utxo) PutUtxo(utxos map[crypto.Hash][]byte) []*db.WriteBatch {
	var writeBatchs []*db.WriteBatch
	for hash, utxo := range utxos {
		writeBatchs = append(writeBatchs, db.NewWriteBatch(u.utxoColumnFamily, db.OperationPut, hash.Bytes(), utxo, u.utxoColumnFamily))
	}
	return writeBatchs
}
