package database

import (
	//	"fmt"
	"fmt"
	"log"
	"testing"

	"github.com/boltdb/bolt"
)

func setupTestDB() (*bolt.DB, error) {
	// Create database
	db, err := bolt.Open("testing.db", 0600, nil)
	if err != nil {
		return nil, fmt.Errorf("could not open db, %v", err)
	}

	// Make DB.TRANSACTIONS bucket
	err = db.Update(func(tx *bolt.Tx) error {
		root, err := tx.CreateBucketIfNotExists([]byte("DB"))
		if err != nil {
			return fmt.Errorf("could not create root bucket: %v", err)
		}
		_, err = root.CreateBucketIfNotExists([]byte("TRANSACTIONS"))
		if err != nil {
			return fmt.Errorf("could not create weight bucket: %v", err)
		}
		return nil
	})
	if err != nil {
		return nil, fmt.Errorf("could not set up buckets, %v", err)
	}

	fmt.Println("DB Setup Done")
	return db, nil
}

func TestSimpleDatabase(t *testing.T) {
	db, _ := setupTestDB()

	// Add "apple":"good"
	_ = db.Update(func(tx *bolt.Tx) error {
		err := tx.Bucket([]byte("DB")).Bucket([]byte("TRANSACTIONS")).Put([]byte("apple"), []byte("good"))

		if err != nil {
			return fmt.Errorf("could not insert weight: %v", err)
		}
		return nil
	})

	// read value
	db.View(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte("DB")).Bucket([]byte("TRANSACTIONS"))
		v := b.Get([]byte("apple"))

		if string(v) != "good" {
			t.Errorf("Expected apple, got %s\n", v)
		}
		return nil
	})

}

func TestSetGetTransaction(t *testing.T) {
	db, err := setupDB()
	resetSequence(db)

	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()

	tr := Transaction{"Jonny38275", []uint32{1, 2, 3, 5, 8}}

	err = addTransaction(db, tr)
	if err != nil {
		panic(err)
	}

	tr2, err := getTransaction(db, 1)
	if err != nil {
		panic(err)
	}

	if tr.UserID != tr2.UserID {
		t.Errorf("Expected Jonny38275, got %s", tr2.UserID)
	}

	if tr2.Votes[1] != tr.Votes[1] {
		t.Errorf("Votes[1] doesn't match")
	}
}

func TestVoteCountAfterTransaction(t *testing.T) {
	// Initalize
	db, err := setupDB()
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()
	resetSequence(db)

	// Add a transaction
	tr := Transaction{"Jonny38275", []uint32{1, 123456, 0, 5}}
	err = addTransaction(db, tr)
	if err != nil {
		panic(err)
	}

	// Retrieve values from STATE.VOTES[0 to 3]
	v := []uint32{0, 0, 0, 0}
	err = db.View(func(tx *bolt.Tx) error {
		votesBucket := tx.Bucket([]byte("STATE")).Bucket([]byte("VOTES"))
		for i := range v {
			vBytes := votesBucket.Get(uintToBytes(uint32(i)))
			if vBytes != nil {
				v[i] = bytesToUint(vBytes)
			} else {
				v[i] = 0
			}
		}
		return nil
	})
	if err != nil {
		t.Errorf("View-only Transaction failed")
	}

	for i := range v {
		if v[i] != tr.Votes[i] {
			t.Errorf("Stored vote[%d] as %d, got %d", i, tr.Votes[i], v[i])
		}
	}
}
