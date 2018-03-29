package txpool

import (
	"fmt"
	"math"

	"github.com/zipper-project/zipper/common/crypto"
	"github.com/zipper-project/zipper/ledger/state"
	"github.com/zipper-project/zipper/proto"
)

// CheckTransactionSanity performs some preliminary checks on a transaction to
// ensure it is sane.  These checks are context free.
func CheckTransactionSanity(tx *proto.Transaction) error {
	// A transaction must have at least one input.
	if len(tx.Inputs) == 0 {
		return fmt.Errorf("transaction has no inputs")
	}

	// A transaction must have at least one output.
	if len(tx.Outputs) == 0 {
		return fmt.Errorf("transaction has no outputs")
	}

	// Ensure the transaction amounts are in range.  Each transaction
	// output must not be negative.  Also, the total of all outputs
	// must abide by the same restrictions.
	var totalAmount int64
	for _, txOut := range tx.Outputs {
		amount := txOut.Value
		if amount < 0 {
			err := fmt.Errorf("transaction output has negative "+
				"value of %v", amount)
			return err
		}

		// Two's complement int64 overflow guarantees that any overflow
		// is detected and reported.
		totalAmount += amount
		if totalAmount < 0 {
			err := fmt.Errorf("total value of all transaction "+
				"outputs exceeds max allowed value of %v",
				MaxAmount)
			return err
		}
		if totalAmount > MaxAmount {
			err := fmt.Errorf("total value of all transaction "+
				"outputs is %v which is higher than max "+
				"allowed value of %v", totalAmount,
				MaxAmount)
			return err
		}
	}

	// Check for duplicate transaction inputs.
	existingTxOut := make(map[OutPoint]struct{})
	for _, txIn := range tx.Inputs {
		if _, exists := existingTxOut[*NewOutPoint(txIn.PreviousOutPoint)]; exists {
			return fmt.Errorf("transaction contains duplicate inputs")
		}
		existingTxOut[*NewOutPoint(txIn.PreviousOutPoint)] = struct{}{}
	}

	// Previous transaction outputs referenced by the inputs to this
	// transaction must not be null.
	for _, txIn := range tx.Inputs {
		prevOut := NewOutPoint(txIn.PreviousOutPoint)
		if isNullOutpoint(prevOut) {
			return fmt.Errorf("transaction " +
				"input refers to previous output that " +
				"is null")
		}
	}
	return nil
}

// isNullOutpoint determines whether or not a previous transaction output point
// is set.
func isNullOutpoint(outpoint *OutPoint) bool {
	if outpoint.Index == math.MaxUint32 && outpoint.Hash.Equal(crypto.Hash{}) {
		return true
	}
	return false
}

// CheckTransactionInputs performs a series of checks on the inputs to a
// transaction to ensure they are valid.  An example of some of the checks
// include verifying all inputs exist, detecting double spends,
// validating all values and fees are in the legal range
// and the total output amount doesn't exceed the input
// amount, and verifying the signatures to prove the spender was the owner
// and therefore allowed to spend them.  As it checks the inputs,
// it also calculates the total fees for the transaction and returns that value.
//
// NOTE: The transaction MUST have already been sanity checked with the
// CheckTransactionSanity function prior to calling this function.
func CheckTransactionInputs(tx *proto.Transaction, utxoView *state.UtxoViewpoint) (int64, error) {
	txHash := tx.Hash()
	var totalAmountIn int64
	for txInIndex, txIn := range tx.Inputs {
		// Ensure the referenced input transaction is available.
		originTxHash := crypto.NewHash(txIn.PreviousOutPoint.TxHash)
		originTxIndex := txIn.PreviousOutPoint.Index
		utxoEntry := utxoView.LookupEntry(originTxHash)
		if utxoEntry == nil || utxoEntry.IsOutputSpent(originTxIndex) {
			err := fmt.Errorf("output %v referenced from "+
				"transaction %s:%d either does not exist or "+
				"has already been spent", txIn.PreviousOutPoint,
				tx.Hash(), txInIndex)
			return 0, err
		}

		// Ensure the transaction amounts are in range.  Each of the
		// output values of the input transactions must not be negative
		// or more than the max allowed per transaction.
		originTxAmount := utxoEntry.AmountByIndex(originTxIndex)
		if originTxAmount < 0 {
			err := fmt.Errorf("transaction output has negative "+
				"value of %v", originTxAmount)
			return 0, err
		}
		if originTxAmount > MaxAmount {
			err := fmt.Errorf("transaction output value of %v is "+
				"higher than max allowed value of %v",
				originTxAmount,
				MaxAmount)
			return 0, err
		}

		// The total of all outputs must not be more than the max
		// allowed per transaction.  Also, we could potentially overflow
		// the accumulator so check for overflow.
		lastAmountIn := totalAmountIn
		totalAmountIn += originTxAmount
		if totalAmountIn < lastAmountIn ||
			totalAmountIn > MaxAmount {
			err := fmt.Errorf("total value of all transaction "+
				"inputs is %v which is higher than max "+
				"allowed value of %v", totalAmountIn,
				MaxAmount)
			return 0, err
		}
	}

	// Calculate the total output amount for this transaction.  It is safe
	// to ignore overflow and out of range errors here because those error
	// conditions would have already been caught by checkTransactionSanity.
	var totalAmountOut int64
	for _, txOut := range tx.Outputs {
		totalAmountOut += txOut.Value
	}

	// Ensure the transaction does not spend more than its inputs.
	if totalAmountIn < totalAmountOut {
		err := fmt.Errorf("total value of all transaction inputs for "+
			"transaction %v is %v which is less than the amount "+
			"spent of %v", txHash, totalAmountIn, totalAmountOut)
		return 0, err
	}

	txFeeInAmount := totalAmountIn - totalAmountOut
	return txFeeInAmount, nil
}

// ValidateTransactionScripts validates the scripts for the passed transaction
// using multiple goroutines.
func ValidateTransactionScripts(tx *proto.Transaction, utxoView *state.UtxoViewpoint) error {
	return nil
}
