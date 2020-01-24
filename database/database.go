/* This database code was a lot easier to write with the awesome reference at:
https://github.com/zupzup/boltdb-example/blob/master/main.go
*/

package database

import (
	"encoding/json"
	"fmt"
	"strconv"

	"github.com/boltdb/bolt"
)

// Transaction type
type Transaction struct {
	UserID string `json:"UserID"`
	// TimeStamp TimeDate
	Votes []uint32 `json:"Votes"`
}

/* integer storage utilites */
func bytesToUint(b []byte) uint32 {
	r, err := strconv.ParseInt(string(b), 10, 32)
	if err != nil {
		panic(err)
	}
	return uint32(r)
}

func uintToBytes(i uint32) []byte {
	str := fmt.Sprintf("%09d", i)
	return []byte(str)
}

/* database initialization */

func setupDB() (*bolt.DB, error) {
	db, err := bolt.Open("test.db", 0600, nil)
	if err != nil {
		return nil, fmt.Errorf("could not open db, %v", err)
	}

	err = db.Update(func(tx *bolt.Tx) error {

		// Make STATE bucket
		stateBucket, err := tx.CreateBucketIfNotExists([]byte("STATE"))
		if err != nil {
			return fmt.Errorf("could not create STATE bucket: %v", err)
		}

		// Create VOTES bucket fresh
		stateBucket.DeleteBucket([]byte("VOTES"))
		_, err = stateBucket.CreateBucketIfNotExists([]byte("VOTES"))
		if err != nil {
			return fmt.Errorf("could not create VOTES bucket: %v", err)
		}

		// Set the transactionSequence number at 0 if it's not set
		transactionSequence := stateBucket.Get([]byte("transactionSequence"))
		if transactionSequence == nil { // if trSeq isn't found..
			stateBucket.Put([]byte("transactionSequence"), uintToBytes(0)) // ..set to 0..
		} // ..otherwise leave it alone

		// Make TRANSACTIONS bucket
		_, err = tx.CreateBucketIfNotExists([]byte("TRANSACTIONS"))
		if err != nil {
			return fmt.Errorf("could not create TRANSACTIONS bucket: %v", err)
		}

		return nil
	})
	if err != nil {
		return nil, fmt.Errorf("could not set up buckets, %v", err)
	}
	return db, nil
}

func resetSequence(db *bolt.DB) error {
	// Delete the whole DB bucket
	err := db.Update(func(tx *bolt.Tx) error {

		// Set the transactionSequence number at 0
		tx.Bucket([]byte("STATE")).Put([]byte("transactionSequence"), uintToBytes(0))
		return nil
	})
	if err != nil {
		return fmt.Errorf("could not reset DB, %v", err)
	}
	return nil
}

func addTransaction(db *bolt.DB, tr Transaction) error {
	err := db.Update(func(tx *bolt.Tx) error {

		// Increase VOTES counts
		votesBucket := tx.Bucket([]byte("STATE")).Bucket([]byte("VOTES"))
		var currentVotes uint32
		for index, count := range tr.Votes {

			candidateNumber := uint32(index)

			// We don't have to increase the count for candidates who received no votes
			if count > 0 {
				currentVotesByte := votesBucket.Get(uintToBytes(candidateNumber))
				if currentVotesByte != nil {
					currentVotes = bytesToUint(currentVotesByte)
				} else {
					currentVotes = 0
				}

				// Update count and store value
				currentVotes += count
				votesBucket.Put(uintToBytes(candidateNumber), uintToBytes(currentVotes))
			}
		}

		// Retieve and update transaction sequence number
		trSequenceBytes := tx.Bucket([]byte("STATE")).Get([]byte("transactionSequence"))
		trSequence := bytesToUint(trSequenceBytes)
		trSequence++

		// Store transaction
		transactionBytes, err := json.Marshal(tr)
		if err != nil {
			return fmt.Errorf("Unable to marshal transaction, %v", err)
		}
		tx.Bucket([]byte("TRANSACTIONS")).Put(uintToBytes(trSequence), transactionBytes)

		// Store updated sequence number
		tx.Bucket([]byte("STATE")).Put([]byte("transactionSequence"), uintToBytes(trSequence))

		return nil
	})
	return err
}

func getTransaction(db *bolt.DB, trNumber uint32) (Transaction, error) {
	var bytes []byte
	var tr Transaction

	err := db.View(func(tx *bolt.Tx) error {
		bytes = tx.Bucket([]byte("TRANSACTIONS")).Get(uintToBytes(trNumber))

		if bytes == nil { // Could not find key
			return fmt.Errorf("Unable to find transaction number %09d", trNumber)
		}
		return nil
	})
	if err != nil {
		return tr, fmt.Errorf("Unable to retrieve transaction, %v", err)
	}

	err = json.Unmarshal(bytes, &tr)
	if err != nil {
		return tr, fmt.Errorf("Could not unmarshal data, %v", err)
	}
	return tr, nil
}
