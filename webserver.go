package main

import (
	"Emoji-battle-royale/config"
	"Emoji-battle-royale/database"
	"Emoji-battle-royale/scheduler"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"time"

	//	"time"

	//"github.com/BurntSushi/toml"

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

// VotePOSTHandler This recieves votes as POST requests to /vote and records them to the database
func VotePOSTHandler(response http.ResponseWriter, request *http.Request) {

	t := database.Transaction{}
	err := json.NewDecoder(request.Body).Decode(&t)
	if err != nil {
		fmt.Println("Unable to parse transaction:", request.Body)
		http.Error(response, "422 unable to parse input", 422)
		return
	}

	db.StoreTransaction(t)
}

// VoteGETHandler returns a vote page based on the current phase
func VoteGETHandler(sched scheduler.Schedule) http.Handler {
	switch phase := sched.GetPhase(); phase {
	case scheduler.Before:
		return ServeSingleFileHandler("vote_before.html")
	case scheduler.During:
		return ServeSingleFileHandler("vote_during.html")
	case scheduler.After:
		return ServeSingleFileHandler("vote_after.html")
	}
	return http.NotFoundHandler()
}

/***** GLOBAL VARIABLES *****/

var db *database.Store

/***** MAIN *****/

func main() {

	configFile := "example_config.toml"

	conf, err := config.LoadConfig(configFile)
	if err != nil {
		log.Fatal("Unable to load config")
	}

	fmt.Printf("name: %s", conf.ElectionName)

	databaseFile := "blue.db" //TODO: move the database file into a separate folder

	resetDatabaseEachOpen := true
	if resetDatabaseEachOpen {
		db, err = database.CreateOrOverwriteDB(databaseFile)
	} else {
		db, err = database.OpenDB(databaseFile)
	}
	defer db.Close()

	if err != nil {
		panic(err) // could not open database. Unrecoverable error
	}

	r := mux.NewRouter()
	r.Handle("/about", ServeSingleFileHandler("about.html")).Methods("GET")
	r.Handle("/vote", ServeSingleFileHandler("vote_before.html")).Methods("GET")
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
