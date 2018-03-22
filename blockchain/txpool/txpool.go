package txpool

import (
	"container/list"
	"fmt"
	"sync"
	"sync/atomic"
	"time"

	"github.com/zipper-project/zipper/common/crypto"
	"github.com/zipper-project/zipper/ledger/state"

	"github.com/zipper-project/zipper/common/log"
	"github.com/zipper-project/zipper/proto"
)

const (
	// MaxAmount is the maximum transaction amount allowed .
	MaxAmount = 21e6 * 1e8

	// MaxOrphanTxs is the maximum number of orphan transactions
	// that can be queued.
	MaxOrphanTxs = 1000

	// orphanTTL is the maximum amount of time an orphan is allowed to
	// stay in the orphan pool before it expires and is evicted during the
	// next scan.
	orphanTTL = 15 * time.Minute

	// orphanExpireScanInterval is the minimum amount of time in between
	// scans of the orphan pool to evict expired transactions.
	orphanExpireScanInterval = 5 * time.Minute
)

// Config is a descriptor containing the memory pool configuration.
type Config struct {
	// FetchUtxoView defines the function to use to fetch unspent
	// transaction output information.
	FetchUtxoView func(*proto.Transaction) (*state.UtxoViewpoint, error)
}

// TxDesc is a descriptor containing a transaction in the mempool along with
// additional metadata.
type TxDesc struct {
	// Tx is the transaction associated with the entry.
	Tx *proto.Transaction
	// Added is the time when the entry was added to the source pool.
	Added time.Time
}

// orphanTx is normal transaction that references an ancestor transaction
// that is not yet available.  It also contains additional information related
// to it such as an expiration time to help prevent caching the orphan forever.
type orphanTx struct {
	// Tx is the transaction associated with the entry.
	tx *proto.Transaction
	// expiration is the time to help prevent caching the orphan forever.
	expiration time.Time
}

// OutPoint defines a bitcoin data type that is used to track previous
// transaction outputs.
type OutPoint struct {
	Hash  crypto.Hash
	Index uint32
}

func NewOutPoint(out *proto.OutPoint) *OutPoint {
	return &OutPoint{
		Hash:  crypto.NewHash(out.TxHash),
		Index: out.Index,
	}
}

// TxPool is used as a source of transactions that need to be mined into blocks
// and relayed to other peers.  It is safe for concurrent access from multiple
// peers.
type TxPool struct {

	// The following variables must only be used atomically.
	lastUpdated int64 // last time pool was updated

	sync.RWMutex
	cfg           *Config
	pool          map[crypto.Hash]*TxDesc
	outpoints     map[OutPoint]*proto.Transaction
	orphans       map[crypto.Hash]*orphanTx
	orphansByPrev map[OutPoint]map[crypto.Hash]*proto.Transaction

	// nextExpireScan is the time after which the orphan pool will be
	// scanned in order to evict orphans.  This is NOT a hard deadline as
	// the scan will only run when an orphan is added to the pool as opposed
	// to on an unconditional timer.
	nextExpireScan time.Time
}

// removeOrphan removes the passed orphan transaction from the orphan pool and
// previous orphan index.
func (mp *TxPool) removeOrphan(tx *proto.Transaction, removeRedeemers bool) {
	// Nothing to do if passed tx is not an orphan.
	txHash := tx.Hash()
	_, exists := mp.orphans[txHash]
	if !exists {
		return
	}

	// Remove the reference from the previous orphan index.
	for _, txIn := range tx.Inputs {
		orphans, exists := mp.orphansByPrev[*NewOutPoint(txIn.PreviousOutPoint)]
		if exists {
			delete(orphans, txHash)

			// Remove the map entry altogether if there are no
			// longer any orphans which depend on it.
			if len(orphans) == 0 {
				delete(mp.orphansByPrev, *NewOutPoint(txIn.PreviousOutPoint))
			}
		}
	}

	// Remove any orphans that redeem outputs from this one if requested.
	if removeRedeemers {
		prevOut := OutPoint{Hash: txHash}
		for txOutIdx := range tx.Outputs {
			prevOut.Index = uint32(txOutIdx)
			for _, orphan := range mp.orphansByPrev[prevOut] {
				mp.removeOrphan(orphan, true)
			}
		}
	}

	// Remove the transaction from the orphan pool.
	delete(mp.orphans, txHash)
}

// RemoveOrphan This function is safe for concurrent access.
func (mp *TxPool) RemoveOrphan(tx *proto.Transaction, removeRedeemers bool) {
	mp.Lock()
	defer mp.Unlock()
	mp.removeOrphan(tx, removeRedeemers)
}

// limitNumOrphans limits the number of orphan transactions by evicting a random
// orphan if adding a new one would cause it to overflow the max allowed.
func (mp *TxPool) limitNumOrphans() error {
	// Scan through the orphan pool and remove any expired orphans when it's
	// time.  This is done for efficiency so the scan only happens
	// periodically instead of on every orphan added to the pool.
	if now := time.Now(); now.After(mp.nextExpireScan) {
		origNumOrphans := len(mp.orphans)
		for _, otx := range mp.orphans {
			if now.After(otx.expiration) {
				// Remove redeemers too because the missing
				// parents are very unlikely to ever materialize
				// since the orphan has already been around more
				// than long enough for them to be delivered.
				mp.removeOrphan(otx.tx, true)
			}
		}

		// Set next expiration scan to occur after the scan interval.
		mp.nextExpireScan = now.Add(orphanExpireScanInterval)

		numOrphans := len(mp.orphans)
		if numExpired := origNumOrphans - numOrphans; numExpired > 0 {
			// pickNoun returns the singular or plural form of a noun depending
			// on the count n.
			pickNoun := func(n int, singular, plural string) string {
				if n == 1 {
					return singular
				}
				return plural
			}
			log.Debugf("Expired %d %s (remaining: %d)", numExpired,
				pickNoun(numExpired, "orphan", "orphans"),
				numOrphans)
		}
	}

	// Nothing to do if adding another orphan will not cause the pool to
	// exceed the limit.
	if len(mp.orphans)+1 <= MaxOrphanTxs {
		return nil
	}

	// Remove a random entry from the map.  For most compilers, Go's
	// range statement iterates starting at a random item although
	// that is not 100% guaranteed by the spec.  The iteration order
	// is not important here because an adversary would have to be
	// able to pull off preimage attacks on the hashing function in
	// order to target eviction of specific entries anyways.
	for _, otx := range mp.orphans {
		// Don't remove redeemers in the case of a random eviction since
		// it is quite possible it might be needed again shortly.
		mp.removeOrphan(otx.tx, false)
		break
	}

	return nil
}

// addOrphan adds an orphan transaction to the orphan pool.
func (mp *TxPool) addOrphan(tx *proto.Transaction) {
	// Nothing to do if no orphans are allowed.
	if MaxOrphanTxs <= 0 {
		return
	}

	// Limit the number orphan transactions to prevent memory exhaustion.
	// This will periodically remove any expired orphans and evict a random
	// orphan if space is still needed.
	mp.limitNumOrphans()

	txHash := tx.Hash()
	mp.orphans[txHash] = &orphanTx{
		tx:         tx,
		expiration: time.Now().Add(orphanTTL),
	}
	for _, txIn := range tx.Inputs {
		if _, exists := mp.orphansByPrev[*NewOutPoint(txIn.PreviousOutPoint)]; !exists {
			mp.orphansByPrev[*NewOutPoint(txIn.PreviousOutPoint)] =
				make(map[crypto.Hash]*proto.Transaction)
		}
		mp.orphansByPrev[*NewOutPoint(txIn.PreviousOutPoint)][txHash] = tx
	}

	log.Debugf("Stored orphan transaction %v (total: %d)", txHash,
		len(mp.orphans))
}

// removeOrphanDoubleSpends removes all orphans which spend outputs spent by the
// passed transaction from the orphan pool.  Removing those orphans then leads
// to removing all orphans which rely on them, recursively.  This is necessary
// when a transaction is added to the main pool because it may spend outputs
// that orphans also spend.
func (mp *TxPool) removeOrphanDoubleSpends(tx *proto.Transaction) {
	for _, txIn := range tx.Inputs {
		for _, orphan := range mp.orphansByPrev[*NewOutPoint(txIn.PreviousOutPoint)] {
			mp.removeOrphan(orphan, true)
		}
	}
}

// isTransactionInPool returns whether or not the passed transaction already
// exists in the main pool.
func (mp *TxPool) isTransactionInPool(hash *crypto.Hash) bool {
	if _, exists := mp.pool[*hash]; exists {
		return true
	}
	return false
}

// IsTransactionInPool This function is safe for concurrent access.
func (mp *TxPool) IsTransactionInPool(hash *crypto.Hash) bool {
	mp.RLock()
	defer mp.RUnlock()
	return mp.isTransactionInPool(hash)
}

// isOrphanInPool returns whether or not the passed transaction already exists
// in the orphan pool.
func (mp *TxPool) isOrphanInPool(hash *crypto.Hash) bool {
	if _, exists := mp.orphans[*hash]; exists {
		return true
	}
	return false
}

// IsOrphanInPool This function is safe for concurrent access.
func (mp *TxPool) IsOrphanInPool(hash *crypto.Hash) bool {
	mp.RLock()
	defer mp.RUnlock()
	return mp.isOrphanInPool(hash)
}

// haveTransaction returns whether or not the passed transaction already exists
// in the main pool or in the orphan pool.
func (mp *TxPool) haveTransaction(hash *crypto.Hash) bool {
	return mp.isTransactionInPool(hash) || mp.isOrphanInPool(hash)
}

// HaveTransaction This function is safe for concurrent access.
func (mp *TxPool) HaveTransaction(hash *crypto.Hash) bool {
	mp.RLock()
	defer mp.RUnlock()
	return mp.haveTransaction(hash)
}

// removeTransaction removes the passed transaction from the mempool. When the
// removeRedeemers flag is set, any transactions that redeem outputs from the
// removed transaction will also be removed recursively from the mempool, as
// they would otherwise become orphans.
func (mp *TxPool) removeTransaction(tx *proto.Transaction, removeRedeemers bool) {
	txHash := tx.Hash()
	if removeRedeemers {
		// Remove any transactions which rely on this one.
		for i := uint32(0); i < uint32(len(tx.Outputs)); i++ {
			prevOut := OutPoint{Hash: txHash, Index: i}
			if txRedeemer, exists := mp.outpoints[prevOut]; exists {
				mp.removeTransaction(txRedeemer, true)
			}
		}
	}

	// Remove the transaction if needed.
	if txDesc, exists := mp.pool[txHash]; exists {
		// Mark the referenced outpoints as unspent by the pool.
		for _, txIn := range txDesc.Tx.Inputs {
			delete(mp.outpoints, *NewOutPoint(txIn.PreviousOutPoint))
		}
		delete(mp.pool, txHash)
		atomic.StoreInt64(&mp.lastUpdated, time.Now().Unix())
	}
}

// RemoveTransaction This function is safe for concurrent access.
func (mp *TxPool) RemoveTransaction(tx *proto.Transaction, removeRedeemers bool) {
	mp.Lock()
	defer mp.Unlock()
	mp.removeTransaction(tx, removeRedeemers)
}

// removeDoubleSpends removes all transactions which spend outputs spent by the
// passed transaction from the memory pool.  Removing those transactions then
// leads to removing all transactions which rely on them, recursively.  This is
// necessary when a block is connected to the main chain because the block may
// contain transactions which were previously unknown to the memory pool.
func (mp *TxPool) removeDoubleSpends(tx *proto.Transaction) {
	for _, txIn := range tx.Inputs {
		if txRedeemer, ok := mp.outpoints[*NewOutPoint(txIn.PreviousOutPoint)]; ok {
			if !txRedeemer.Hash().Equal(tx.Hash()) {
				mp.removeTransaction(txRedeemer, true)
			}
		}
	}
}

// RemoveDoubleSpends This function is safe for concurrent access.
func (mp *TxPool) RemoveDoubleSpends(tx *proto.Transaction) {
	mp.Lock()
	defer mp.Unlock()
	mp.removeDoubleSpends(tx)
}

// addTransaction adds the passed transaction to the memory pool.  It should
// not be called directly as it doesn't perform any validation.  This is a
// helper for maybeAcceptTransaction.
func (mp *TxPool) addTransaction(tx *proto.Transaction) *TxDesc {
	// Add the transaction to the pool and mark the referenced outpoints
	// as spent by the pool.
	txD := &TxDesc{
		Tx:    tx,
		Added: time.Now(),
	}
	mp.pool[tx.Hash()] = txD

	for _, txIn := range tx.Inputs {
		mp.outpoints[*NewOutPoint(txIn.PreviousOutPoint)] = tx
	}
	atomic.StoreInt64(&mp.lastUpdated, time.Now().Unix())

	return txD
}

// checkPoolDoubleSpend checks whether or not the passed transaction is
// attempting to spend coins already spent by other transactions in the pool.
// Note it does not check for double spends against transactions already in the
// main chain.
func (mp *TxPool) checkPoolDoubleSpend(tx *proto.Transaction) error {
	for _, txIn := range tx.Inputs {
		if txR, exists := mp.outpoints[*NewOutPoint(txIn.PreviousOutPoint)]; exists {
			err := fmt.Errorf("output %v already spent by "+
				"transaction %v in the memory pool",
				*NewOutPoint(txIn.PreviousOutPoint), txR.Hash())
			return err
		}
	}

	return nil
}

// Count returns the number of transactions in the main pool.  It does not
// include the orphan pool.
func (mp *TxPool) Count() int {
	mp.RLock()
	defer mp.RUnlock()
	count := len(mp.pool)
	return count
}

// LastUpdated returns the last time a transaction was added to or removed from
// the main pool.  It does not include the orphan pool.
func (mp *TxPool) LastUpdated() time.Time {
	return time.Unix(atomic.LoadInt64(&mp.lastUpdated), 0)
}

// processOrphans determines if there are any orphans which depend on the passed
// transaction hash (it is possible that they are no longer orphans) and
// potentially accepts them to the memory pool.  It repeats the process for the
// newly accepted transactions (to detect further orphans which may no longer be
// orphans) until there are no more.
//
// It returns a slice of transactions added to the mempool.  A nil slice means
// no transactions were moved from the orphan pool to the mempool.
func (mp *TxPool) processOrphans(acceptedTx *proto.Transaction) []*TxDesc {
	var acceptedTxns []*TxDesc

	// Start with processing at least the passed transaction.
	processList := list.New()
	processList.PushBack(acceptedTx)
	for processList.Len() > 0 {
		// Pop the transaction to process from the front of the list.
		firstElement := processList.Remove(processList.Front())
		processItem := firstElement.(*proto.Transaction)

		prevOut := OutPoint{Hash: processItem.Hash()}
		for txOutIdx := range processItem.Outputs {
			// Look up all orphans that redeem the output that is
			// now available.  This will typically only be one, but
			// it could be multiple if the orphan pool contains
			// double spends.  While it may seem odd that the orphan
			// pool would allow this since there can only possibly
			// ultimately be a single redeemer, it's important to
			// track it this way to prevent malicious actors from
			// being able to purposely constructing orphans that
			// would otherwise make outputs unspendable.
			//
			// Skip to the next available output if there are none.
			prevOut.Index = uint32(txOutIdx)
			orphans, exists := mp.orphansByPrev[prevOut]
			if !exists {
				continue
			}

			// Potentially accept an orphan into the tx pool.
			for _, tx := range orphans {
				missing, txD, err := mp.maybeAcceptTransaction(
					tx, false)
				if err != nil {
					// The orphan is now invalid, so there
					// is no way any other orphans which
					// redeem any of its outputs can be
					// accepted.  Remove them.
					mp.removeOrphan(tx, true)
					break
				}

				// Transaction is still an orphan.  Try the next
				// orphan which redeems this output.
				if len(missing) > 0 {
					continue
				}

				// Transaction was accepted into the main pool.
				//
				// Add it to the list of accepted transactions
				// that are no longer orphans, remove it from
				// the orphan pool, and add it to the list of
				// transactions to process so any orphans that
				// depend on it are handled too.
				acceptedTxns = append(acceptedTxns, txD)
				mp.removeOrphan(tx, false)
				processList.PushBack(tx)

				// Only one transaction for this outpoint can be
				// accepted, so the rest are now double spends
				// and are removed later.
				break
			}
		}
	}

	// Recursively remove any orphans that also redeem any outputs redeemed
	// by the accepted transactions since those are now definitive double
	// spends.
	mp.removeOrphanDoubleSpends(acceptedTx)
	for _, txD := range acceptedTxns {
		mp.removeOrphanDoubleSpends(txD.Tx)
	}

	return acceptedTxns
}

// ProcessOrphans This function is safe for concurrent access.
func (mp *TxPool) ProcessOrphans(acceptedTx *proto.Transaction) []*TxDesc {
	mp.Lock()
	defer mp.Unlock()
	return mp.processOrphans(acceptedTx)
}

// fetchInputUtxos loads utxo details about the input transactions referenced by
// the passed transaction.  First, it loads the details form the viewpoint of
// the main chain, then it adjusts them based upon the contents of the
// transaction pool.
func (mp *TxPool) fetchInputUtxos(tx *proto.Transaction) (*state.UtxoViewpoint, error) {
	utxoView, err := mp.cfg.FetchUtxoView(tx)
	if err != nil {
		return nil, err
	}

	// Attempt to populate any missing inputs from the transaction pool.
	for originHash, entry := range utxoView.Entries() {
		if entry != nil && !entry.IsFullySpent() {
			continue
		}

		if poolTxDesc, exists := mp.pool[originHash]; exists {
			utxoView.AddTxOuts(poolTxDesc.Tx, 0)
		}
	}
	return utxoView, nil
}

// maybeAcceptTransaction is the main workhorse for handling insertion of new
// free-standing transactions into a memory pool.  It includes functionality
// such as rejecting duplicate transactions, ensuring transactions follow all
// rules, detecting orphan transactions, and insertion into the memory pool.
//
// If the transaction is an orphan (missing parent transactions), the
// transaction is NOT added to the orphan pool, but each unknown referenced
// parent is returned.  use processTransaction instead if new orphans should
// be added to the orphan pool.
func (mp *TxPool) maybeAcceptTransaction(tx *proto.Transaction, rejectDupOrphans bool) ([]*crypto.Hash, *TxDesc, error) {
	txHash := tx.Hash()

	// Don't accept the transaction if it already exists in the pool.  This
	// applies to orphan transactions as well when the reject duplicate
	// orphans flag is set.  This check is intended to be a quick check to
	// weed out duplicates.
	if mp.isTransactionInPool(&txHash) || (rejectDupOrphans && mp.isOrphanInPool(&txHash)) {
		err := fmt.Errorf("already have transaction %v in pool", txHash)
		return nil, nil, err
	}

	// Perform preliminary sanity checks on the transaction.  This makes
	// use of blockchain which contains the invariant rules for what
	// transactions are allowed into blocks.
	err := CheckTransactionSanity(tx)
	if err != nil {
		return nil, nil, err
	}

	// The transaction may not use any of the same outputs as other
	// transactions already in the pool as that would ultimately result in a
	// double spend.  This check is intended to be quick and therefore only
	// detects double spends within the transaction pool itself.  The
	// transaction could still be double spending coins from the main chain
	// at this point.  There is a more in-depth check that happens later
	// after fetching the referenced transaction inputs from the main chain
	// which examines the actual spend data and prevents double spends.
	err = mp.checkPoolDoubleSpend(tx)
	if err != nil {
		return nil, nil, err
	}

	// Fetch all of the unspent transaction outputs referenced by the inputs
	// to this transaction.  This function also attempts to fetch the
	// transaction itself to be used for detecting a duplicate transaction
	// without needing to do a separate lookup.
	utxoView, err := mp.fetchInputUtxos(tx)
	if err != nil {
		return nil, nil, err
	}

	// Don't allow the transaction if it exists in the main chain and is not
	// not already fully spent.
	txEntry := utxoView.LookupEntry(&txHash)
	if txEntry != nil && !txEntry.IsFullySpent() {
		return nil, nil, fmt.Errorf("transaction %v already exists in db", txHash)
	}
	delete(utxoView.Entries(), txHash)

	// Transaction is an orphan if any of the referenced input transactions
	// don't exist.  Adding orphans to the orphan pool is not handled by
	// this function, and the caller should use addOrphan if this
	// behavior is desired.
	var missingParents []*crypto.Hash
	for originHash, entry := range utxoView.Entries() {
		if entry == nil || entry.IsFullySpent() {
			// Must make a copy of the hash here since the iterator
			// is replaced and taking its address directly would
			// result in all of the entries pointing to the same
			// memory location and thus all be the final hash.
			hashCopy := originHash
			missingParents = append(missingParents, &hashCopy)
		}
	}
	if len(missingParents) > 0 {
		return missingParents, nil, nil
	}

	// Perform several checks on the transaction inputs using the invariant
	// rules in blockchain for what transactions are allowed into blocks.
	// Also returns the fees associated with the transaction which will be
	// used later.
	txFee, err := CheckTransactionInputs(tx, utxoView)
	if err != nil {
		return nil, nil, err
	}
	_ = txFee

	// Verify crypto signatures for each input and reject the transaction if
	// any don't verify.
	err = ValidateTransactionScripts(tx, utxoView)
	if err != nil {
		return nil, nil, err
	}

	// Add to transaction pool.
	txD := mp.addTransaction(tx)

	log.Debugf("Accepted transaction %v (pool size: %v)", txHash,
		len(mp.pool))

	return nil, txD, nil
}

// ProcessTransaction is the main workhorse for handling insertion of new
// free-standing transactions into the memory pool.  It includes functionality
// such as rejecting duplicate transactions, ensuring transactions follow all
// rules, orphan transaction handling, and insertion into the memory pool.
//
// It returns a slice of transactions added to the mempool.  When the
// error is nil, the list will include the passed transaction itself along
// with any additional orphan transaactions that were added as a result of
// the passed one being accepted.
func (mp *TxPool) ProcessTransaction(tx *proto.Transaction, allowOrphan bool) ([]*TxDesc, error) {
	mp.Lock()
	defer mp.Unlock()
	log.Debugf("Processing transaction %v", tx.Hash())
	// Potentially accept the transaction to the memory pool.
	missingParents, txD, err := mp.maybeAcceptTransaction(tx, true)
	if err != nil {
		return nil, err
	}

	if len(missingParents) == 0 {
		// Accept any orphan transactions that depend on this
		// transaction (they may no longer be orphans if all inputs
		// are now available) and repeat for those accepted
		// transactions until there are no more.
		newTxs := mp.processOrphans(tx)
		acceptedTxs := make([]*TxDesc, len(newTxs)+1)

		// Add the parent transaction first so remote nodes
		// do not add orphans.
		acceptedTxs[0] = txD
		copy(acceptedTxs[1:], newTxs)

		return acceptedTxs, nil
	}

	// The transaction is an orphan (has inputs missing).  Reject
	// it if the flag to allow orphans is not set.
	if !allowOrphan {
		// Only use the first missing parent transaction in
		// the error message.
		//
		// NOTE: RejectDuplicate is really not an accurate
		// reject code here, but it matches the reference
		// implementation and there isn't a better choice due
		// to the limited number of reject codes.  Missing
		// inputs is assumed to mean they are already spent
		// which is not really always the case.
		err := fmt.Errorf("orphan transaction %v references "+
			"outputs of unknown or fully-spent "+
			"transaction %v", tx.Hash(), missingParents[0])
		return nil, err
	}

	// Potentially add the orphan transaction to the orphan pool.
	mp.addOrphan(tx)
	return nil, nil
}

// New returns a new memory pool for validating and storing standalone
// transactions until they are mined into a block.
func New(cfg *Config) *TxPool {
	return &TxPool{
		cfg:            cfg,
		pool:           make(map[crypto.Hash]*TxDesc),
		outpoints:      make(map[OutPoint]*proto.Transaction),
		orphans:        make(map[crypto.Hash]*orphanTx),
		orphansByPrev:  make(map[OutPoint]map[crypto.Hash]*proto.Transaction),
		nextExpireScan: time.Now().Add(orphanExpireScanInterval),
	}
}
