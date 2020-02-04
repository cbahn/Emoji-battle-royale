package main

import (
	"Emoji-battle-royale/database"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"path/filepath"
	"regexp"
	"strconv"

	"github.com/boltdb/bolt"

	//	"time"

	"github.com/gorilla/mux"
)

// ServeSingleFileHandler returns a handler which serves up a single static file from the public directory
func ServeSingleFileHandler(filename string) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		fileDirectory := "public"
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

// FileServeHandler hhh
// Example: "res/pic", `^/res/pic/(\w+\.\w+)$`
func FileServeHandler(path string, regexMatch string) func(http.ResponseWriter, *http.Request) {
	return func(response http.ResponseWriter, request *http.Request) {
		var resURL = regexp.MustCompile(regexMatch)
		var resource = resURL.FindStringSubmatch(request.URL.Path)
		// resource is captured regex matches i.e. ["/res/file.txt", "file.txt"]

		if len(resource) == 0 { // If url could not be parsed, send 404
			fmt.Println("Could not parse /res request:", request.URL.Path)
			http.Error(response, "404 page not found", 404)
			return
		}

		// Everything's good, serve up the file
		http.ServeFile(response, request, filepath.Join(path, resource[1]))
	}
}

/***** MAIN *****/

var db *bolt.DB

func main() {
	port := 8097
	candidateCount := 50

	mux := mux.NewRouter()
	mux.Handle("/about", ServeSingleFileHandler("about.html")).Methods("GET")
	mux.Handle("/vote", ServeSingleFileHandler("vote.html")).Methods("GET")
	mux.PathPrefix("/res/").Handler(http.StripPrefix("/res/", http.FileServer(http.Dir("public/res"))))
	mux.Handle("/", ServeSingleFileHandler("home.html")).Methods("GET")
	mux.Handle("/vote", http.HandlerFunc(VotePOSTHandler)).Methods("POST")

	databaseFile := "blue.db"
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

	log.Print("Listening on port " + strconv.Itoa(port) + " ... ")
	err = http.ListenAndServe(":"+strconv.Itoa(port), mux)
	if err != nil {
		log.Fatal("ListenAndServe error: ", err)
	}
}
