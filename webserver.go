package main

import (
	"Emoji-battle-royale/database"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/boltdb/bolt"

	//	"time"

	"github.com/gorilla/mux"
)

// ServeSingleFileHandler returns a handler which serves up a single static file from the public directory
func ServeSingleFileHandler(filename string) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		fileDirectory := "public"
		//log.Printf("About to serve %s\n", fileDirectory+"/"+filename)
		http.ServeFile(w, r, fileDirectory+"/"+filename)
	})
}

// VoteMessage struct, as is sent by the client as a json file
type VoteMessage struct {
	ID    string `json:"Id"`
	Votes []uint `json:"Votes"`
}

// VotePOSTHandler This recieves votes as POST requests to /vote and records them to the database
func VotePOSTHandler(response http.ResponseWriter, request *http.Request) {

	votes := database.Transaction{}
	err := json.NewDecoder(request.Body).Decode(&votes)
	if err != nil {
		fmt.Println("Unable to parse transaction:", request.Body)
		http.Error(response, "422 unable to parse input", 422)
		return
	}

	database.AddTransaction(db, votes)
	getvotes, _ := database.GetVotes(db)
	fmt.Println(getvotes)
}

// Route for a request matching path and method
type Route struct {
	path   string
	f      func(http.ResponseWriter, *http.Request)
	method string
}

/***** MAIN *****/

var db *bolt.DB

func main() {
	candidateCount := 50

	databaseFile := "blue.db" //Todo, move this to dat/
	resetDatabaseEachOpen := true

	var err error
	if resetDatabaseEachOpen {
		db, err = database.CreateOrOverwriteDB(databaseFile, candidateCount)
	} else {
		db, err = database.OpenDB(databaseFile)
	}
	defer db.Close()
	if err != nil {
		panic(err) // could not open database. Unrecoverable error
	}

	r := mux.NewRouter()
	r.Handle("/about", ServeSingleFileHandler("about.html")).Methods("GET")
	r.Handle("/vote", ServeSingleFileHandler("vote.html")).Methods("GET")
	r.PathPrefix("/res/").Handler(http.StripPrefix("/res/", http.FileServer(http.Dir("public/res"))))
	r.Handle("/", ServeSingleFileHandler("home.html")).Methods("GET")
	r.Handle("/vote", http.HandlerFunc(VotePOSTHandler)).Methods("POST")

	port := "8080"
	srv := &http.Server{
		Handler:      r,
		Addr:         "127.0.0.1:" + port,
		WriteTimeout: 15 * time.Second,
		ReadTimeout:  15 * time.Second,
	}

	log.Print("Listening on port " + port + " ... ")
	log.Fatal(srv.ListenAndServe())
}
