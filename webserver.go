package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"path/filepath"
	"regexp"
	"strconv"

	//	"time"
	"github.com/gorilla/mux"
)

// SetMyCookie adds a simple cookie to the response
// Just for testing right now
func SetMyCookie(response http.ResponseWriter) {
	cookie := http.Cookie{Name: "testcookiename", Value: "testcookievalue"}
	http.SetCookie(response, &cookie)
}

// AboutHandler servers up the about page. Probably isn't nessesary :-\
func AboutHandler(response http.ResponseWriter, request *http.Request) {
	http.ServeFile(response, request, "about.html")
}

// HomeHandler respond to the URL /home with an html home page
func HomeHandler(response http.ResponseWriter, request *http.Request) {
	response.Header().Set("Content-type", "text/html")
	webpage, err := ioutil.ReadFile("home.html")
	if err != nil {
		http.Error(response, fmt.Sprintf("home.html file error %v", err), 500)
	}
	fmt.Fprint(response, string(webpage))
	fmt.Println("Sent response to /home")
}

// VoteGETHandler serves the vote.html file
func VoteGETHandler(response http.ResponseWriter, request *http.Request) {
	http.ServeFile(response, request, "vote.html")
}

// VoteMessage struct, as is sent by the client as a json file
type VoteMessage struct {
	ID    string `json:"Id"`
	Votes []uint `json:"Votes"`
}

// VotePOSTHandler This recieves votes as POST requests to /vote and records them to the database
func VotePOSTHandler(response http.ResponseWriter, request *http.Request) {

	votes := VoteMessage{}
	err := json.NewDecoder(request.Body).Decode(&votes)
	if err != nil {
		panic(err) // IdIoMaTiC gO eRrOr HaNdLiNg
	}

	fmt.Println(votes.ID, votes.Votes)
}

// ResHandler loads up files from the /res folder
// WARNING - ALL FILES IN THAT FOLDER WILL BE PUBLIC
func ResHandler(response http.ResponseWriter, request *http.Request) {
	resourceFolder := "res"
	// Only resources with characters from a-z, A-Z, 0-9, and the _ (underscore) character will be valid.
	var resURL = regexp.MustCompile(`^/res/(\w+\.\w+)$`)
	var resource = resURL.FindStringSubmatch(request.URL.Path)
	// resource is captured regex matches i.e. ["/res/file.txt", "file.txt"]

	if len(resource) == 0 { // If url could not be parsed, send 404
		fmt.Println("Could not parse /res request:", request.URL.Path)
		http.Error(response, "404 page not found", 404)
		return
	}

	// Everything's good, serve up the file
	http.ServeFile(response, request, filepath.Join(resourceFolder, resource[1]))
}

// PicHandler loads up files from the /res/pic folder
// WARNING - ALL FILES IN THAT FOLDER WILL BE PUBLIC
func PicHandler(response http.ResponseWriter, request *http.Request) {
	resourceFolder := "res/pic"
	// Only resources with characters from a-z, A-Z, 0-9, and the _ (underscore) character will be valid.
	var resURL = regexp.MustCompile(`^/res/pic/(\w+\.\w+)$`)
	var resource = resURL.FindStringSubmatch(request.URL.Path)
	// resource is captured regex matches i.e. ["/res/file.txt", "file.txt"]

	if len(resource) == 0 { // If url could not be parsed, send 404
		fmt.Println("Could not parse /res request:", request.URL.Path)
		http.Error(response, "404 page not found", 404)
		return
	}

	// Everything's good, serve up the file
	http.ServeFile(response, request, filepath.Join(resourceFolder, resource[1]))
}

// Route for a request matching path and method
type Route struct {
	path   string
	f      func(http.ResponseWriter, *http.Request)
	method string
}

// SetRoutes ..
func SetRoutes(routes []Route) *mux.Router {
	mux := mux.NewRouter()
	for _, v := range routes {
		mux.Handle(v.path, http.HandlerFunc(v.f)).Methods(v.method)
	}
	return mux
}

func main() {
	port := 8097
	portstring := strconv.Itoa(port)

	// We're using gorilla/mux as the router because
	// it's not garbage like the default one.
	mux := mux.NewRouter()

	mux.Handle("/about", http.HandlerFunc(AboutHandler)).Methods("GET")
	mux.Handle("/vote", http.HandlerFunc(VoteGETHandler)).Methods("GET")
	mux.Handle("/vote", http.HandlerFunc(VotePOSTHandler)).Methods("POST")
	mux.Handle("/res/{resource}", http.HandlerFunc(ResHandler)).Methods("GET")
	mux.Handle("/res/pic/{picture}", http.HandlerFunc(PicHandler)).Methods("GET")
	mux.Handle("/", http.HandlerFunc(HomeHandler)).Methods("GET")

	// Start listing on a given port with these routes on this server.
	log.Print("Listening on port " + portstring + " ... ")
	err := http.ListenAndServe(":"+portstring, mux)
	if err != nil {
		log.Fatal("ListenAndServe error: ", err)
	}
}
