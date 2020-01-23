package database

import (
	//	"fmt"
	"fmt"
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

func destroyTestDB(db *bolt.DB) error {
	// Delete the whole DB bucket
	err := db.Update(func(tx *bolt.Tx) error {
		err := tx.DeleteBucket([]byte("DB"))
		if err != nil {
			return fmt.Errorf("could not delete root bucket: %v", err)
		}
		return nil
	})
	if err != nil {
		return fmt.Errorf("could not delete testing buckets, %v", err)
	}
	return nil
}

func TestSimpleDatabase(t *testing.T) {
	db, _ := setupTestDB()
	defer destroyTestDB(db)

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
		fmt.Printf("The answer is: %s\n", v)
		return nil
	})

}
