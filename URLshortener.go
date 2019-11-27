package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log"
	"net/http"
)

// simple home page
var homePage = []byte(`
<html>
	<body>
	   Home page of URLshortener
	</body>
</html>
`)

var tokenDB *TokenDB

/* test:
curl -i -v http://localhost:8080/
*/

// Home shows simple home page
// It can be used to check the service health.
func home(w http.ResponseWriter) {
	if tokenDB.DB.Ping() != nil {
		w.WriteHeader(500)
		return
	}
	w.Write(homePage)
}

/* test:
curl -i -v http://localhost:8080/<token>
*/

// Redirect handles redirection to URL that was stored for the specified token
func redirect(w http.ResponseWriter, r *http.Request, sToken string) {
	longURL, err := tokenDB.Get(sToken)
	if err != nil {
		// send 404 response
		http.NotFound(w, r)
		fmt.Printf("URL fror token '%s' was not found\n", sToken)
		return
	}
	// make redirect response
	fmt.Println("Redirest to ", longURL)
	http.Redirect(w, r, longURL, 301)
}

/* test:
curl -v POST -H "Content-Type: application/json" -d '{"url":"https://www.w3schools.com/html/html_forms.asp","exp":"10"}' http://localhost:8080/token
*/

// getNewToken handle the new token creation for passed url and sets expiration for it
func getNewToken(w http.ResponseWriter, r *http.Request) {
	// ????: check some authorisation???

	// parameters structure
	var params struct {
		URL string `json:"url"`
		Exp int    `json:"exp,string,omitempty"`
	}

	// read the request body
	buf := make([]byte, r.ContentLength)
	_, err := r.Body.Read(buf)
	if err != nil && !errors.Is(err, io.EOF) {
		fmt.Printf("request body reading error: %v", err)
		w.WriteHeader(400)
		return
	}

	// read JSON parameters
	err = json.Unmarshal(buf, &params)
	if err != nil || params.URL == "" {
		fmt.Printf("bad request")
		w.WriteHeader(400)
		return
	}

	// set the default expiration if it is not passed
	if params.Exp == 0 {
		params.Exp = 1 // TODO: should be default expiration from config
	}

	// create new token
	sToken, err := tokenDB.New(params.URL, params.Exp)
	if err != nil {
		fmt.Printf("new token creation error: %v\n", err)
		w.WriteHeader(504)
		return
	}

	// return new token and short URL
	fmt.Printf("[ %s | %s | %d ]\n", sToken, params.URL, params.Exp)

	// prepare response body
	resp, err := json.Marshal(
		struct {
			Token string `json:"token"`
			URL   string `json:"url"`
		}{
			Token: sToken,
			URL:   "localhost:8080/" + sToken, // TODO: read server name from config or from request
		})
	if err != nil {
		fmt.Printf("response body JSON marshaling error: %v\n", err)
		w.WriteHeader(500)
		return
	}

	// send response
	w.Write(resp)
}

// myMUX
func myMUX(w http.ResponseWriter, r *http.Request) {
	path := r.URL.Path
	switch path {
	case "/": // request for health-check
		fmt.Println("Home")
		home(w)
	case "/token": // request for new short url/token
		fmt.Println("Request for token")
		getNewToken(w, r)
	case "/favicon.ico": // I have no idea why the chromium make such requests together with request for redirect
		return
	default: // all the rest are requests for redirect
		fmt.Println("Request for redirect")
		redirect(w, r, path[1:])
	}
}

func main() {

	var err error

	// create new TokenDB interface
	tokenDB, err = TokenDBNew()
	if err != nil {
		panic(err)
	}

	// register the handler
	http.HandleFunc("/", myMUX)

	// start server
	// TODO: read host:port from config file
	fmt.Println("starting server at :8080")
	log.Fatal(http.ListenAndServe(":8080", nil))
}
