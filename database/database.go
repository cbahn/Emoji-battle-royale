/* This database code was a lot easier to write with the awesome reference at:
https://github.com/zupzup/boltdb-example/blob/master/main.go
*/

package database

import (
	"encoding/json"
	"fmt"
	"os"
	"strconv"
	"strings"

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
func openDB(filename string) (*bolt.DB, error) {
	db, err := bolt.Open("test.db", 0600, nil)
	if err != nil {
		return nil, fmt.Errorf("could not open db file %s: %v", filename, err)
	}

	// Check that DB state is correct
	err = db.View(func(tx *bolt.Tx) error {
		stateBucket := tx.Bucket([]byte("STATE"))
		if stateBucket == nil {
			return fmt.Errorf("STATE bucket does not exist")
		}

		if nil == stateBucket.Get([]byte("transactionSequence")) {
			return fmt.Errorf("transactionSequence not set")
		}

		candidateCountBytes := stateBucket.Get([]byte("candidateCount"))
		if candidateCountBytes == nil {
			return fmt.Errorf("candidateCount not set")
		}

		candidateCount := int(bytesToUint(candidateCountBytes))
		for i := 0; i < candidateCount; i++ {
			if nil == stateBucket.Bucket([]byte("VOTES")).Get(uintToBytes(uint32(i))) {
				return fmt.Errorf("Could not retrieve VOTES[%d]", i)
			}
		}

		if nil == tx.Bucket([]byte("TRANSACTIONS")) {
			return fmt.Errorf("TRANSACTIONS bucket does not exist")
		}
		return nil
	})

	if err != nil {
		return db, fmt.Errorf("could not open database file %s: %v", filename, err)
	}
	return db, nil
}

func createOrOverwriteDB(filename string, candidateCount int) (*bolt.DB, error) {

	if !strings.HasSuffix(filename, ".db") {
		return nil, fmt.Errorf("New database filename must end with .db")
	}

	// does the database file exist? get rid of it
	if _, err := os.Stat(filename); err == nil {
		// DB file already exists. Delete it
		err = os.Remove(filename)
		if err != nil {
			return nil, fmt.Errorf("Could not remove old file %s: %v", filename, err)
		}
	} else if os.IsNotExist(err) {
		// No DB found. we're good to continue
	} else {
		return nil, fmt.Errorf("Trouble overwriting %s: %v", filename, err)
	}

	// Create new database file
	db, err := bolt.Open(filename, 0600, nil)
	if err != nil {
		return nil, fmt.Errorf("could not open/create db, %v", err)
	}

	err = db.Update(func(tx *bolt.Tx) error {

		// Make STATE bucket
		stateBucket, err := tx.CreateBucket([]byte("STATE"))
		if err != nil {
			return fmt.Errorf("could not create STATE bucket: %v", err)
		}

		// Set the transactionSequence number to 0
		// NOTE: this means that the next transaction created will be 1 and there will never
		//  be a 0 transaction
		stateBucket.Put([]byte("transactionSequence"), uintToBytes(0))

		stateBucket.Put([]byte("candidateCount"), uintToBytes(uint32(candidateCount)))

		// Create VOTES bucket
		votesBucket, err := stateBucket.CreateBucket([]byte("VOTES"))
		if err != nil {
			return fmt.Errorf("could not create VOTES bucket: %v", err)
		}

		// Set all vote counts to 0
		for i := 0; i < candidateCount; i++ {
			votesBucket.Put(uintToBytes(uint32(i)), uintToBytes(0))
		}

		// Create TRANSACTIONS bucket
		_, err = tx.CreateBucketIfNotExists([]byte("TRANSACTIONS"))
		if err != nil {
			return fmt.Errorf("could not create TRANSACTIONS bucket: %v", err)
		}

		return nil
	})
	if err != nil {
		return nil, fmt.Errorf("Could not initialize database: %v", err)
	}
	return db, nil
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

func getVotes(db *bolt.DB) ([]uint32, error) {

	var votes []uint32

	err := db.View(func(tx *bolt.Tx) error {

		candidateCount := int(bytesToUint(tx.Bucket([]byte("STATE")).Get([]byte("candidateCount"))))

		votes = make([]uint32, candidateCount)

		votesBucket := tx.Bucket([]byte("STATE")).Bucket([]byte("VOTES"))
		for i := 0; i < candidateCount; i++ {
			votes[i] = bytesToUint(votesBucket.Get(uintToBytes(uint32(i))))
		}
		return nil
	})
	if err != nil {
		return nil, fmt.Errorf("View transaction failed")
	}
	return votes, nil
}
