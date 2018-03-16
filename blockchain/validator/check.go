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

package validator

import (
	"bytes"
	"fmt"

	"strings"

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

	if tx.Amount() < 0 || tx.Fee() < 0 {
		return fmt.Errorf("[validator] illegal transaction %s : Amount must be >0 or Fee must bigger than 0", tx.Hash())
	}

	switch tx.GetType() {
	case proto.TransactionType_Atomic:
		if strings.Compare(tx.FromChain(), tx.ToChain()) != 0 {
			return fmt.Errorf("[validator] illegal transaction %s : fromchain %s == tochain %s", tx.Hash(), tx.FromChain(), tx.ToChain())
		}
	case proto.TransactionType_AcrossChain:
		if !(len(tx.FromChain()) == len(tx.ToChain()) && strings.Compare(tx.FromChain(), tx.ToChain()) != 0) {
			return fmt.Errorf("[validator] illegal transaction %s : wrong chain floor, fromchain %s ==  tochain %s", tx.Hash(), tx.FromChain(), tx.ToChain())
		}
	case proto.TransactionType_Distribut:
		address := tx.Sender()
		fromChain := coordinate.HexToChainCoordinate(tx.FromChain())
		toChainParent := coordinate.HexToChainCoordinate(tx.ToChain()).ParentCoorinate()
		if !bytes.Equal(fromChain, toChainParent) || strings.Compare(address.String(), tx.Recipient().String()) != 0 {
			return fmt.Errorf("[validator] illegal transaction %s :wrong chain floor, fromChain %s - toChain %s = 1", tx.Hash(), tx.FromChain(), tx.ToChain())
		}
	case proto.TransactionType_Backfront:
		address := tx.Sender()
		fromChainParent := coordinate.HexToChainCoordinate(tx.FromChain()).ParentCoorinate()
		toChain := coordinate.HexToChainCoordinate(tx.ToChain())
		if !bytes.Equal(fromChainParent, toChain) || strings.Compare(address.String(), tx.Recipient().String()) != 0 {
			return fmt.Errorf("[validator] illegal transaction %s :wrong chain floor, fromChain %s - toChain %s = 1", tx.Hash(), tx.FromChain(), tx.ToChain())
		}
	case proto.TransactionType_Merged:
	// nothing to do
	case proto.TransactionType_Issue, proto.TransactionType_IssueUpdate:
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
			if tp := tx.GetType(); tp == proto.TransactionType_Issue {
				asset := &state.Asset{
					ID:     tx.AssetID(),
					Issuer: tx.Sender(),
					Owner:  tx.Recipient(),
				}
				if _, err := asset.Update(string(tx.Payload)); err != nil {
					return fmt.Errorf("[validator] illegal transaction %s: invalid issue coin(%s) - %s", tx.Hash(), string(tx.Payload), err)
				}
			} else if tp == proto.TransactionType_IssueUpdate {
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
	for _, addr := range v.config.PublicAddresses {
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

	address, err := tx.Verfiy()
	if err != nil || !bytes.Equal(address.Bytes(), tx.Sender().Bytes()) {
		return fmt.Errorf("[validator] illegal transaction %s: invalid signature", tx.Hash())
	}

	return nil
}
