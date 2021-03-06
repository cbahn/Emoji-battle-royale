/* This database code was a lot easier to write with the awesome reference at:
https://github.com/zupzup/boltdb-example
*/

package database

import (
	"encoding/binary"
	"encoding/csv"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"os"
	"strings"

	"github.com/boltdb/bolt"
)

type Votes map[string]int

// Transaction type
type Transaction struct {
	UserID string `json:"Id"`
	// TimeStamp TimeDate
	Votes Votes `json:"Votes"`
}

type Store struct {
	db *bolt.DB
}

var expectedBuckets = [...]string{"TRANSACTIONS", "VOTES", "CANDIDATES"}

// itob returns an 8-byte big endian representation of v.
func itob(v int) []byte {
	b := make([]byte, 8)
	binary.BigEndian.PutUint64(b, uint64(v))
	return b
}

// btoi converts an 8-byte big endian byte array to an int
func btoi(b []byte) int {
	return int(binary.BigEndian.Uint64(b))
}

func booltobyte(b bool) []byte {
	if b {
		return []byte{byte(1)}
	}
	return []byte{byte(0)}
}

func bytetobool(b []byte) bool {
	return int(b[0]) == 1
}

// OpenDB loads a database and verifies that it contains the expected buckets
func OpenDB(filename string) (*Store, error) {
	db, err := bolt.Open("test.db", 0600, nil)
	if err != nil {
		return nil, fmt.Errorf("could not open db file %s: %v", filename, err)
	}

	// Check that DB state is correct
	err = db.View(func(tx *bolt.Tx) error {

		// Check that exected buckets exist
		for _, v := range expectedBuckets {
			if nil == tx.Bucket([]byte(v)) {
				return fmt.Errorf("%s bucket not found", v)
			}
		}

		return nil
	})

	if err != nil {
		return &Store{db: db}, fmt.Errorf("could not open database file %s: %v", filename, err)
	}
	return &Store{db: db}, nil
}

// CreateOrOverwriteDB will create a new database or delete and re-create one if the filename already exists
func CreateOrOverwriteDB(filename string) (*Store, error) {

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

		// Create buckets
		for _, v := range expectedBuckets {
			_, err := tx.CreateBucket([]byte(v))
			if err != nil {
				return fmt.Errorf("Could not create %s bucket: %v", v, err)
			}
		}

		return nil
	})
	if err != nil {
		return nil, fmt.Errorf("Could not initialize database: %v", err)
	}
	return &Store{db: db}, nil
}

// Close the database connection
func (s *Store) Close() {
	s.db.Close()
}

// InitializeCandidates populates the CANDIDATES and VOTES buckets
// WARNING: calling this on an already initialized database will not cause any errors
func (s *Store) InitializeCandidates(candidates []string) {
	s.db.Update(func(tx *bolt.Tx) error {
		bCAN := tx.Bucket([]byte("CANDIDATES"))
		bVOT := tx.Bucket([]byte("VOTES"))

		for _, can := range candidates {
			// Add all candidates to the candidate list and set their value to true
			bCAN.Put([]byte(can), booltobyte(true))

			// Add all candidates to the vote list and set their number to 0
			bVOT.Put([]byte(can), itob(0))
		}
		return nil
	})
}

// StoreTransaction saves the transaction to the database
func (s *Store) StoreTransaction(t Transaction) error {
	return s.db.Update(func(tx *bolt.Tx) error {
		// Retrieve buckets
		bTRN := tx.Bucket([]byte("TRANSACTIONS"))
		bCAN := tx.Bucket([]byte("CANDIDATES"))
		bVOT := tx.Bucket([]byte("VOTES"))

		// Increase the total vote count for each candidate voted for
		for candidate, voteCount := range t.Votes {
			// Confirm that the candidate exists and is active
			candidateStatus := bCAN.Get([]byte(candidate))
			if candidateStatus == nil {
				return fmt.Errorf("Transaction contains invalid candidate name: %s", candidate)
			}
			if !bytetobool(candidateStatus) {
				return fmt.Errorf("Cannot vote for eliminated candidate %s", candidate)
			}

			v := bVOT.Get([]byte(candidate))
			if v == nil {
				return fmt.Errorf("Transaction contains invalid candidate name: %s", candidate)
			}

			bVOT.Put([]byte(candidate), itob(voteCount+btoi(v)))
		}

		// Generate ID for this trasaction
		// This returns an error only if the Tx is closed or not writeable.
		// That can't happen in an Update() call so I ignore the error check.
		id, _ := bTRN.NextSequence()

		// Marshal transaction into bytes.
		buf, err := json.Marshal(t)
		if err != nil {
			return err
		}

		//fmt.Printf("Storing transaction: %s\n", string(buf))

		// Persist bytes to bucket
		return bTRN.Put(itob(int(id)), buf)
	})
}

// EliminateCandidate turns the CANDIDATES(candidate) value to false
//  so that they can no longer recieve votes
func (s *Store) EliminateCandidate(candidate string) error {
	return s.db.Update(func(tx *bolt.Tx) error {
		// Retrieve buckets
		bCAN := tx.Bucket([]byte("CANDIDATES"))

		c := bCAN.Get([]byte(candidate))

		if c == nil {
			return fmt.Errorf("Cannot eliminate %s, candidate not found", candidate)
		}

		if !bytetobool(c) {
			return fmt.Errorf("Cannot eliminate %s, candidate already eliminted", candidate)
		}

		bCAN.Put([]byte(candidate), booltobyte(false))

		return nil
	})
}

// GetAllTransactions returns a map of all Transactions by transactionID
func (s *Store) GetAllTransactions() map[int]Transaction {
	m := make(map[int]Transaction)

	s.db.View(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte("TRANSACTIONS"))

		c := b.Cursor()
		for k, v := c.First(); k != nil; k, v = c.Next() {

			// It's important that the struct is cleared between each new transaction
			var t Transaction

			if err := json.Unmarshal(v, &t); err != nil {
				return fmt.Errorf("Unable to unmarshal transaction %d", btoi(k))
			}

			m[btoi(k)] = t
		}
		return nil
	})
	return m
}

// GetVotes returns a map of current candidates vote totals from all transactions
func (s *Store) GetVotes() Votes {
	votes := make(map[string]int)

	s.db.View(func(tx *bolt.Tx) error {
		bVOT := tx.Bucket([]byte("VOTES"))

		c := bVOT.Cursor()
		for k, v := c.First(); k != nil; k, v = c.Next() {
			votes[string(k)] = btoi(v)
		}

		return nil
	})

	return votes
}

// GetCandidateList returns a list of all candidates
// if includeEliminatedCandidates is false, only active candidates will be returned
func (s *Store) GetCandidateList(includeEliminatedCandidates bool) []string {
	var candidateList []string
	s.db.View(func(tx *bolt.Tx) error {
		bCAN := tx.Bucket([]byte("CANDIDATES"))
		c := bCAN.Cursor()
		for k, v := c.First(); k != nil; k, v = c.Next() {
			if includeEliminatedCandidates || bytetobool(v) {
				candidateList = append(candidateList, string(k))
			}
		}
		return nil
	})
	return candidateList
}

// ExportAllTransactionsAsCSV will export a complete list of all transactions in CSV format
// to the writer w
func (s *Store) ExportAllTransactionsAsCSV(w io.Writer) {
	candidates := s.GetCandidateList(true)

	header := append([]string{"Transaction Number", "UserId"}, candidates...)

	var data = [][]string{header}

	transactions := s.GetAllTransactions()

	for trasactionNumber, transaction := range transactions {
		// Start building the line with the Transaction Number and the UserId
		var line = []string{fmt.Sprintf("%d", trasactionNumber), transaction.UserID}

		for _, can := range candidates {

			if val, ok := transaction.Votes[can]; ok {
				//Candidate found, set votes
				line = append(line, fmt.Sprintf("%d", val))
			} else {
				// Candidate not found, votes==0
				line = append(line, "0")
			}
		}

		data = append(data, line)
	}

	csvw := csv.NewWriter(w)
	csvw.WriteAll(data)

	if err := csvw.Error(); err != nil {
		log.Fatalln("error writing csv:", err)
	}
}
