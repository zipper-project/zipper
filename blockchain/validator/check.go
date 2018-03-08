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

package validator

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"path/filepath"
	"plugin"
	"strings"
	"sync"

	"github.com/zipper-project/zipper/common/log"
	"github.com/zipper-project/zipper/common/utils"
	"github.com/zipper-project/zipper/coordinate"
	"github.com/zipper-project/zipper/ledger/state"
	"github.com/zipper-project/zipper/params"
	"github.com/zipper-project/zipper/proto"
)

func (v *Verification) isOverCapacity() bool {
	return v.txpool.Len() > v.config.TxPoolCapacity
}

func (v *Verification) isExist(tx *proto.Transaction) bool {
	if _, ok := v.inTxs[tx.Hash()]; ok {
		return true
	}

	if ledgerTx, _ := v.ledger.GetTxByTxHash(tx.Hash().Bytes()); ledgerTx != nil {
		return true
	}

	return false
}

func (v *Verification) checkTransactionLegal(tx *proto.Transaction) error {
	if !(strings.Compare(tx.FromChain(), params.ChainID.String()) == 0 || (strings.Compare(tx.ToChain(), params.ChainID.String()) == 0)) {
		return fmt.Errorf("[validator] illegal transaction %s : fromCahin %s or toChain %s == params.ChainID %s", tx.Hash(), tx.FromChain(), tx.ToChain(), params.ChainID.String())
	}

	if tx.Amount().Sign() < 0 || tx.Fee().Sign() < 0 {
		return fmt.Errorf("[validator] illegal transaction %s : Amount must be >0 or Fee must bigger than 0", tx.Hash())
	}

	switch tx.GetType() {
	case proto.TypeAtomic:
		if strings.Compare(tx.FromChain(), tx.ToChain()) != 0 {
			return fmt.Errorf("[validator] illegal transaction %s : fromchain %s == tochain %s", tx.Hash(), tx.FromChain(), tx.ToChain())
		}
	case proto.TypeAcrossChain:
		if !(len(tx.FromChain()) == len(tx.ToChain()) && strings.Compare(tx.FromChain(), tx.ToChain()) != 0) {
			return fmt.Errorf("[validator] illegal transaction %s : wrong chain floor, fromchain %s ==  tochain %s", tx.Hash(), tx.FromChain(), tx.ToChain())
		}
	case proto.TypeDistribut:
		address := tx.Sender()
		fromChain := coordinate.HexToChainCoordinate(tx.FromChain())
		toChainParent := coordinate.HexToChainCoordinate(tx.ToChain()).ParentCoorinate()
		if !bytes.Equal(fromChain, toChainParent) || strings.Compare(address.String(), tx.Recipient().String()) != 0 {
			return fmt.Errorf("[validator] illegal transaction %s :wrong chain floor, fromChain %s - toChain %s = 1", tx.Hash(), tx.FromChain(), tx.ToChain())
		}
	case proto.TypeBackfront:
		address := tx.Sender()
		fromChainParent := coordinate.HexToChainCoordinate(tx.FromChain()).ParentCoorinate()
		toChain := coordinate.HexToChainCoordinate(tx.ToChain())
		if !bytes.Equal(fromChainParent, toChain) || strings.Compare(address.String(), tx.Recipient().String()) != 0 {
			return fmt.Errorf("[validator] illegal transaction %s :wrong chain floor, fromChain %s - toChain %s = 1", tx.Hash(), tx.FromChain(), tx.ToChain())
		}
	case proto.TypeMerged:
	// nothing to do
	case proto.TypeIssue, proto.TypeIssueUpdate:
		fromChain := coordinate.HexToChainCoordinate(tx.FromChain())
		toChain := coordinate.HexToChainCoordinate(tx.FromChain())

		// && strings.Compare(fromChain.String(), "00") == 0)
		if len(fromChain) != len(toChain) {
			return fmt.Errorf("[validator] illegal transaction %s: should issue chain floor, fromChain %s or toChain %s", tx.Hash(), tx.FromChain(), tx.ToChain())
		}

		if !v.isIssueTransaction(tx) {
			return fmt.Errorf("[validator] illegal transaction %s: valid issue tx public key fail", tx.Hash())
		}

		if len(tx.Payload) > 0 {
			if tp := tx.GetType(); tp == proto.TypeIssue {
				asset := &state.Asset{
					ID:     tx.AssetID(),
					Issuer: tx.Sender(),
					Owner:  tx.Recipient(),
				}
				if _, err := asset.Update(string(tx.Payload)); err != nil {
					return fmt.Errorf("[validator] illegal transaction %s: invalid issue coin(%s) - %s", tx.Hash(), string(tx.Payload), err)
				}
			} else if tp == proto.TypeIssueUpdate {
				asset := &state.Asset{
					ID:     tx.AssetID(),
					Issuer: tx.Sender(),
					Owner:  tx.Recipient(),
				}
				if _, err := asset.Update(string(tx.Payload)); err != nil {
					asset := &state.Asset{
						ID:     tx.AssetID(),
						Issuer: tx.Recipient(),
						Owner:  tx.Sender(),
					}
					if _, err := asset.Update(string(tx.Payload)); err != nil {
						return fmt.Errorf("[validator] illegal transaction %s: invalid issue coin(%s) - %s", tx.Hash(), string(tx.Payload), err)
					}
				}
			}
		}
	}

	return nil
}

func (v *Verification) isIssueTransaction(tx *proto.Transaction) bool {
	address := tx.Sender()
	addressHex := utils.BytesToHex(address.Bytes())
	for _, addr := range params.PublicAddress {
		if strings.Compare(addressHex, addr) == 0 {
			return true
		}
	}
	return false
}

func (v *Verification) checkTransaction(tx *proto.Transaction) error {
	if err := v.checkTransactionLegal(tx); err != nil {
		return err
	}
	if err := v.checkTransactionSecurity(tx); err != nil {
		return fmt.Errorf("[validator] illegal transaction %+v: err: %v", tx.Hash(), err)
	}
	address, err := tx.Verfiy()
	if err != nil || !bytes.Equal(address.Bytes(), tx.Sender().Bytes()) {
		return fmt.Errorf("[validator] illegal transaction %s: invalid signature", tx.Hash())
	}

	return nil
}

// SecurityPluginDir returns the directory of security plugin.
func (v *Verification) SecurityPluginDir() string {
	return v.config.SecurityPluginDir
}

// SecurityVerifier is the security plugin verifier.
type SecurityVerifier func(*proto.Transaction, func(key string) ([]byte, error)) error

// SecurityVerifierManager managers the security plugin verifier.
type SecurityVerifierManager struct {
	sync.Mutex
	securityPath string
	verifier     SecurityVerifier
}

var securityVerifierMnger SecurityVerifierManager

func (v *Verification) getSecurityVerifier() (SecurityVerifier, error) {
	securityVerifierMnger.Lock()
	defer securityVerifierMnger.Unlock()

	securityPathData, err := v.sctx.GetContractStateData(params.GlobalStateKey, params.SecurityContractKey)
	if err != nil {
		log.Errorf("get security plugin path failed, %v", err)
		return nil, fmt.Errorf("get security plugin path failed, %v", err)
	}
	if len(securityPathData) == 0 {
		return nil, nil
	}

	var securityPath string
	err = json.Unmarshal(securityPathData, &securityPath)
	if err != nil {
		log.Errorf("unmarshal security plugin path failed, %v", err)
		return nil, fmt.Errorf("unmarshal security plugin path failed, %v", err)
	}

	if securityPath == securityVerifierMnger.securityPath {
		return securityVerifierMnger.verifier, nil
	}

	security, err := plugin.Open(filepath.Join(v.SecurityPluginDir(), securityPath))
	if err != nil {
		log.Errorf("load security plugin failed, %v", err)
		return nil, fmt.Errorf("load security plugin failed, %v", err)
	}

	verifyFn, err := security.Lookup("Verify")
	if err != nil {
		log.Errorf("can't find security plugin verifier, %v", err)
		return nil, fmt.Errorf("can't find security plugin verifier, %v", err)
	}

	verifier, ok := verifyFn.(func(*proto.Transaction, func(key string) ([]byte, error)) error)
	if !ok {
		log.Error("invalid security plugin verifier format")
		return nil, errors.New("invalid security plugin verifier format")
	}

	securityVerifierMnger.verifier = SecurityVerifier(verifier)
	securityVerifierMnger.securityPath = securityPath
	return securityVerifierMnger.verifier, nil
}

func (v *Verification) checkTransactionSecurity(tx *proto.Transaction) error {
	verifier, err := v.getSecurityVerifier()
	if err != nil {
		return err
	}

	if verifier == nil {
		return nil
	}

	if err := verifier(tx, func(key string) ([]byte, error) {
		data, err := v.sctx.GetContractStateData(params.GlobalStateKey, key)
		if err != nil {
			return nil, err
		}
		return data, nil
	}); err != nil {
		return err
	}

	log.Infof("security verify success of transaction %v", tx.Hash())
	return nil
}

func (v *Verification) checkTransactionSecurityByContract(tx *proto.Transaction) bool {
	securityAddrData, err := v.sctx.GetContractStateData(params.GlobalStateKey, params.SecurityContractKey)
	if err != nil {
		log.Errorf("get security contract address failed, %v", err)
		return true
	}

	if len(securityAddrData) == 0 {
		log.Info("there is no security contract yet")
		return true
	}

	var addr string
	err = json.Unmarshal(securityAddrData, &addr)
	if err != nil {
		log.Errorf("unmarshal security contract address failed, %v", err)
		return true
	}

	bh, _ := v.ledger.Height()
	v.sctx.StartConstract(bh)
	ok, err := v.sctx.ExecuteRequireContract(tx, addr)
	v.sctx.StopContract(bh)

	if err != nil {
		log.Errorf("execute security contract failed, %v", err)
		return true
	}

	return ok
}
