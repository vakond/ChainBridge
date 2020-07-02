// Copyright 2020 ChainSafe Systems
// SPDX-License-Identifier: LGPL-3.0-only

package trie

import (
	"container/list"
	"errors"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/ethdb/leveldb"
	"github.com/ethereum/go-ethereum/rlp"
	"github.com/ethereum/go-ethereum/trie"
)

type Trie struct {
	trie     *trie.Trie
	txStored int
}

type TxTries struct {
	// TODO: the memory allocated for these is hard to get back, look for better way to have a queue
	txTries      []*Trie
	txRoots      []common.Hash
	triesToStore int
}

var (
	emptyRoot = common.HexToHash("56e81f171bcc55a6ff8345e692c0f86e5b48e01b996cadc001622fb5e363b421")
)

func newTxTries(t int) *TxTries {
	txTrie := &TxTries{
		triesToStore: t,
	}
	return txTrie

}

func (t *TxTries) updateTriesAndRoots(trie *Trie, root common.Hash) {
	if len(t.txTries) >= t.triesToStore {
		t.txTries = append(t.txTries, trie)
		// delete contents of trie from database
		t.txTries[0].deleteTrie(t.txRoots[0], t.txTries[0].txStored)
		t.txTries = t.txTries[1:]

		t.txRoots = append(t.txRoots, root)
		t.txRoots = t.txRoots[1:]

	} else {
		t.txTries = append(t.txTries, trie)
		t.txRoots = append(t.txRoots, root)
	}

}

func (t *TxTries) indexOfRoot(root common.Hash) int {
	for i, r := range t.txRoots {
		if root == r {
			return i
		}
	}
	return -1

}

// AddTrie creates a new instance of a trie object
func (t *TxTries) AddTrie(root common.Hash, db *leveldb.Database, transactions []common.Hash, transactionsRoot common.Hash) (*Trie, error) {
	// TODO: look into cache values
	// this creates a new trie database with our KVDB as the diskDB for node storage
	newTrie, err := trie.New(root, trie.NewDatabaseWithCache(db, 0))

	if err != nil {
		return nil, err
	}

	trie := &Trie{
		trie: newTrie,
	}

	err = trie.updateTrie(transactions, transactionsRoot)

	if err != nil {
		return nil, errors.New("could not update trie")
	}

	t.updateTriesAndRoots(trie, root)

	return trie, nil
}

func (t *Trie) updateTrie(transactions []common.Hash, transactionRoot common.Hash) error {

	for i, tx := range transactions {
		key, err := rlp.EncodeToBytes(i)
		if err != nil {
			return err
		}

		t.trie.Update(key, tx.Bytes())
	}

	// check if the root hash of the trie matches the transactionRoot
	if t.trie.Hash() != transactionRoot {
		return errors.New("transaction roots don't match")
	}

	//update the root in the
	return nil
}

func (t *Trie) deleteTrie(root common.Hash, txStored int) error {
	for i := 0; i < 10; i++ {
		// keys for all transactions are the rlp encoding of their position in the block.
		key, err := rlp.EncodeToBytes(i)
		if err != nil {
			return err
		}

		err = t.trie.TryDelete(key)
		if err != nil {
			return err
		}
	}

	return nil

}

// retrieves a Proof for a value at key in trie with root root
func (t *TxTries) RetrieveProof(root common.Hash, key []byte) (*ProofDatabase, error) {
	index := t.indexOfRoot(root)

	if index == -1 {
		return nil, errors.New("transaction trie for this transaction root does not exist")
	}

	return t.txTries[index].retrieveProof(root, key)
}

// retrieves a proof from a trie object, given the root of the trie it is contained in and the key
func (t *Trie) retrieveProof(root common.Hash, key []byte) (*ProofDatabase, error) {
	var proof = NewProofDatabase()
	err := t.trie.Prove(key, 0, proof)
	if err != nil {
		return nil, err
	}

	return proof, nil
}
