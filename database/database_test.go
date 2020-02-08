package database

import (
	"testing"
)

func TestAddTransactions(t *testing.T) {
	var databaseName string = "TestAddTransactions.db"

	db1, err := CreateOrOverwriteDB(databaseName)
	if err != nil {
		t.Errorf("Couldn't create database: %v", err)
	}

	db1.InitializeCandidates([]string{"ted", "jeb", "hil"})

	t1 := Transaction{
		userID: "jonny",
		votes: Votes{
			"ted": 7,
			"jeb": 15,
		},
	}

	t2 := Transaction{
		userID: "billy",
		votes: Votes{
			"jeb": 70,
			"hil": 153,
		},
	}

	if err := db1.StoreTransaction(t1); err != nil {
		t.Errorf("Error adding transaction 1: %v", err)
	}

	if err := db1.StoreTransaction(t2); err != nil {
		t.Errorf("Error adding transaction 2: %v", err)
	}

	if len(db1.GetAllTransactions()) != 2 {
		t.Errorf("Expected 2 transactions, got %d", len(db1.GetAllTransactions()))
	}

	receivedVotes := db1.GetVotes()
	expectedVotes := Votes{
		"ted": 7,
		"jeb": 15 + 70,
		"hil": 153,
	}
	for can, vo := range expectedVotes {
		if receivedVotes[can] != vo {
			t.Errorf("Expected %s to have %d votes, got %d", can, vo, receivedVotes[can])
		}
	}
}

func TestInvalidTransaction(t *testing.T) {
	var databaseName string = "TestInvalidTransactions.db"

	db1, err := CreateOrOverwriteDB(databaseName)
	if err != nil {
		t.Errorf("Couldn't create database: %v", err)
	}

	db1.InitializeCandidates([]string{"ted", "jeb", "hil"})

	err = db1.StoreTransaction(Transaction{
		userID: "jonny",
		votes: Votes{
			"ted": 7,
			"jeb": 15,
		},
	})
	if err != nil {
		t.Errorf("Could not store jonny's valid transaction")
	}

	_ = db1.EliminateCandidate("jeb") // Please clap

	if len(db1.GetCandidateList(true)) != 3 {
		t.Errorf("GetCandidateList returned %d results, 3 expected", len(db1.GetCandidateList(true)))
	}

	if len(db1.GetCandidateList(false)) != 2 {
		t.Errorf("GetEliminatedCandidates returned %d results, 2 expected", len(db1.GetCandidateList(false)))
	}

	err = db1.StoreTransaction(Transaction{
		userID: "billy",
		votes: Votes{
			"jeb": 70,
			"hil": 153,
		},
	})
	if err == nil {
		t.Errorf("No error when storing invalid transaction 1")
	}

	err = db1.StoreTransaction(Transaction{
		userID: "billy",
		votes: Votes{
			"ted":      1,
			"Ron Paul": 999,
		},
	})
	if err == nil {
		t.Errorf("No error when storing invalid transaction 2")
	}

	if len(db1.GetAllTransactions()) != 1 {
		t.Errorf("Expected 1 transactions, got %d", len(db1.GetAllTransactions()))
	}

	receivedVotes := db1.GetVotes()
	expectedVotes := Votes{
		"ted": 7,
		"jeb": 15,
		"hil": 0,
	}
	for can, vo := range expectedVotes {
		if receivedVotes[can] != vo {
			t.Errorf("Expected %s to have %d votes, got %d", can, vo, receivedVotes[can])
		}
	}
}
