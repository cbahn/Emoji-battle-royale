package database

import (
	//	"fmt"

	"log"
	"testing"

	"github.com/boltdb/bolt"
)

func TestCreateThenOpenDatabase(t *testing.T) {
	var databaseName string = "test.db"

	db1, err := createOrOverwriteDB(databaseName, 10)
	if err != nil {
		t.Errorf("Couldn't create database: %v", err)
	}

	db1.Close()

	db2, err := openDB(databaseName)
	if err != nil {
		t.Errorf("Could not open database: %v", err)
	}
	defer db2.Close()

	err = db2.View(func(tx *bolt.Tx) error {
		candidate0Byte := tx.Bucket([]byte("STATE")).Bucket([]byte("VOTES")).Get(uintToBytes(0))

		if candidate0Byte == nil {
			t.Errorf("Could not read value of VOTES[0]")
		}

		if 0 != int(bytesToUint(candidate0Byte)) {
			t.Errorf("Newly created VOTES[0] expected 0, has %d", int(bytesToUint((candidate0Byte))))
		}
		return nil
	})
	if err != nil {
		t.Errorf("Could not complete view-only transaction: %v", err)
	}
}

func TestSetGetTransaction(t *testing.T) {
	db, err := createOrOverwriteDB("test.db", 10)
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
	db, err := createOrOverwriteDB("test.db", 10)
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()

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
