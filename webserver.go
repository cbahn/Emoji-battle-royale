package main

import (
	"fmt"
	"io/ioutil"
	"strconv"
	"log"
	"net/http"
	"regexp"
	"encoding/json"
	"time"
)

func SetMyCookie(response http.ResponseWriter){
	// Add a simplistic cookie to the response.
	cookie := http.Cookie{Name: "testcookiename", Value:"testcookievalue"}
	http.SetCookie(response, &cookie)
}

// Respond to URLs of the form /generic/...
func GenericHandler(response http.ResponseWriter, request *http.Request){

	// Set cookie and MIME type in the HTTP headers.
	SetMyCookie(response)
	response.Header().Set("Content-type", "text/plain")

	// Parse URL and POST data into the request.Form
	err := request.ParseForm()
	if err != nil {
		http.Error(response, fmt.Sprintf("error parsing url %v", err), 500)
	}

	// Send the text diagnostics to the client.
	fmt.Fprint(response,  "FooWebHandler says ... \n")
	fmt.Fprintf(response, " request.Method     '%v'\n", request.Method)
	fmt.Fprintf(response, " request.RequestURI '%v'\n", request.RequestURI)
	fmt.Fprintf(response, " request.URL.Path   '%v'\n", request.URL.Path)
	fmt.Fprintf(response, " request.Form       '%v'\n", request.Form)
	fmt.Fprintf(response, " request.Cookies()  '%v'\n", request.Cookies())
}

// Respond to the URL /home with an html home page
func HomeHandler(response http.ResponseWriter, request *http.Request){
	response.Header().Set("Content-type", "text/html")
	webpage, err := ioutil.ReadFile("home.html")
	if err != nil { 
		http.Error(response, fmt.Sprintf("home.html file error %v", err), 500)
	}
	fmt.Fprint(response, string(webpage));
	fmt.Println("Sent response to /home")
}

// Respond to URLs of the form /item/...
func ItemHandler(response http.ResponseWriter, request *http.Request){

	// Set cookie and MIME type in the HTTP headers.
	SetMyCookie(response)
	response.Header().Set("Content-type", "application/json")

	// Some sample data to be sent back to the client.
	data := map[string]string { "what" : "item", "name" : "" }

	// Was the URL of the form /item/name ?
	var itemURL = regexp.MustCompile(`^/item/(\w+)$`)
	var itemMatches = itemURL.FindStringSubmatch(request.URL.Path)
	// itemMatches is captured regex matches i.e. ["/item/which", "which"]
	if len(itemMatches) > 0 {
		// Yes, so send the JSON to the client.
		// Send the data appended to the current time
		t := time.Now()
		data["name"] = t.Format("03:04:05")
		json_bytes, _ := json.Marshal(data)
		fmt.Fprintf(response, "%s\n", json_bytes)

	} else {
		// No, so send "page not found."
		http.Error(response, "404 page not found", 404)
	}
}

// Respond to URLs of the form /item/...
func PutHandler(response http.ResponseWriter, request *http.Request){

	// Check that this is actually a POST request or 404
	if request.Method == "POST" {

		// Parse the request with ParseForm()
		if err := request.ParseForm(); err != nil {
			fmt.Fprintf(response, "ParseForm() err: %v", err)
			return
		}
		fmt.Println("post recieved: %v",request.PostForm)
		fmt.Fprintf(response, "Post from website! request.PostForm = %v\n", request.PostForm)

	} else {
		// No, so send "page not found."
		http.Error(response, "404 page not found", 404)
	}
}


// Loads up files from the /res folder when.
// WARNING - ALL FILES IN THAT FOLDER WILL BE PUBLIC
func ResHandler(response http.ResponseWriter, request *http.Request){

	// Only resources with characters from a-z, A-Z, 0-9, and the _ (underscore) character will be valid.
	var resURL = regexp.MustCompile(`^/res/(\w+\.\w+)$`) 
	var resource = resURL.FindStringSubmatch(request.URL.Path)
	// resource is captured regex matches i.e. ["/res/file.txt", "file.txt"]

	if len(resource) == 0 { // If url could not be parsed, send 404
		fmt.Println("Could not parse /res request:", request.URL.Path)
		http.Error(response, "404 page not found", 404)
		return
	}

	_, err := ioutil.ReadFile("res/" + resource[1])

	if err != nil { // File read error, send 404
		fmt.Println("Error processing response ",request.URL.Path,err)
		http.Error(response, "404 page not found", 404)
		return
	}

	// Everything's good, serve up the file
	http.ServeFile(response, request, "res/" + resource[1])
}


func main(){
	port := 8097
	portstring := strconv.Itoa(port)

	// Register request handlers for two URL patterns.
	// (The docs are unclear on what a 'pattern' is, 
	// but seems be the start of the URL, ending in a /).
	// See gorilla/mux for a more powerful matching system.
	// Note that the "/" pattern matches all request URLs.
	mux := http.NewServeMux()
	mux.Handle("/home", 	http.HandlerFunc( HomeHandler ))
	mux.Handle("/item/",	http.HandlerFunc( ItemHandler ))
	mux.Handle("/generic/", http.HandlerFunc( GenericHandler ))
	mux.Handle("/put/", 	http.HandlerFunc( PutHandler ))
	mux.Handle("/res/",		http.HandlerFunc( ResHandler ))

	// Start listing on a given port with these routes on this server.
	// (I think the server name can be set here too , i.e. "foo.org:8080")
	log.Print("Listening on port " + portstring + " ... ")
	err := http.ListenAndServe(":" + portstring, mux)
	if err != nil {
		log.Fatal("ListenAndServe error: ", err)
	}
}

