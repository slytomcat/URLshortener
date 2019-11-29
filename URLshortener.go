package main

// URLshortener is a microservice to shorten long URLs
// and to handle the redirection by generated short URLs.
//
// Request for short URL:
// URL: <host>[:<port>]/token
// Method: POST
// Body: JSON with following parameters:
//   url - URL to shorten, mandatory
//   exp - short URL expiration in days, optional
// Response: JSON with following parameters:
//   token - token for short URL
//   url - short URL
//
// Redirect to long URL:
// URL: <host>[:<port>]/<token> - URL from response on request for short URL
// Method: GET
// No parameters
// Response contain the redirection to long URL
//
// Helth-check:
// URL: <host>[:<port>]/
// Method: GET
// No parameters
// Response: simple home page and HTTP 200 OK in case of good service health
// or HTTP 500 Server error in case of bad service health

import (
	"encoding/json"
	"errors"
	"io"
	"log"
	"net/http"
	"strings"
)

// simple home page to display on health check request
var (
	homePage = []byte(`
<html>
	<body>
	   <h1>Home page of URLshortener</h1>

	   See sources at <a href="https://github.com/slytomcat/URLshortener">https://github.com/slytomcat/URLshortener</a>
	</body>
</html>
`)
	tokenDB *TokenDB
	Server  *http.Server
)

/* test for test env:
curl -i -v http://localhost:8080/
*/

// Home shows simple home page
// But before showing home page it make full selftest
func home(w http.ResponseWriter) {

	// Perform self-test

	// make the request for new token
	resp, err := http.Post("http://"+CONFIG.ListenHostPort+"/token", "application/json",
		strings.NewReader(`{"url": "http://`+CONFIG.ShortDomain+`/favicon.ico", "exp": "1"}`))
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
	}
	defer resp.Body.Close()

	// read response body
	buf := make([]byte, resp.ContentLength)
	_, err = resp.Body.Read(buf)
	if err != nil && !errors.Is(err, io.EOF) {
		w.WriteHeader(http.StatusInternalServerError)
	}

	// decode response body
	var rep struct {
		URL   string `json:"url"`
		Token string `json:"token"`
	}
	err = json.Unmarshal(buf, &rep)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
	}

	// try to make the redirect by received short URL
	resp2, err := http.Get("http://" + rep.URL)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
	}
	defer resp2.Body.Close()

	// expire received token and check the responce status
	if err := tokenDB.Expire(rep.Token); err != nil || (resp2.StatusCode != http.StatusOK) {
		w.WriteHeader(http.StatusInternalServerError)
	}

	// finally, if self-test was successfully passed, show the home page
	w.Write(homePage)
}

/* test for test env:
curl -i -v http://localhost:8080/<token>
*/

// Redirect handles redirection to URL that was stored for the specified token
func redirect(w http.ResponseWriter, r *http.Request, sToken string) {
	longURL, err := tokenDB.Get(sToken)
	if err != nil {
		// send 404 response
		http.NotFound(w, r)
		log.Printf("URL for token '%s' was not found\n", sToken)
		return
	}
	// make redirect response
	log.Println("Redirest to ", longURL)
	http.Redirect(w, r, longURL, http.StatusMovedPermanently)
}

/* test for test env:
curl -v POST -H "Content-Type: application/json" -d '{"url":"https://www.w3schools.com/html/html_forms.asp","exp":"10"}' http://localhost:8080/token
*/

// getNewToken handle the new token creation for passed url and sets expiration for it
func getNewToken(w http.ResponseWriter, r *http.Request) {
	// ????: check some authorisation???

	// parameters structure
	var params struct {
		URL string `json:"url"`                  // long URL
		Exp int    `json:"exp,string,omitempty"` // Expiration
	}

	// read the request body
	buf := make([]byte, r.ContentLength)
	_, err := r.Body.Read(buf)
	if err != nil && !errors.Is(err, io.EOF) {
		log.Printf("request body reading error: %v", err)
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	// parse JSON to parameters structure
	err = json.Unmarshal(buf, &params)
	if err != nil || params.URL == "" {
		log.Printf("bad request")
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	// set the default expiration if it is not passed
	if params.Exp == 0 {
		params.Exp = CONFIG.DefaultExp
	}

	// create new token
	sToken, err := tokenDB.New(params.URL, params.Exp)
	if err != nil {
		log.Printf("new token creation error: %v\n", err)
		w.WriteHeader(http.StatusGatewayTimeout)
		return
	}

	// prepare response body
	resp, err := json.Marshal(
		struct {
			Token string `json:"token"` // token
			URL   string `json:"url"`   // short URL
		}{
			Token: sToken,
			URL:   CONFIG.ShortDomain + "/" + sToken,
		})
	if err != nil {
		log.Printf("response body JSON marshaling error: %v\n", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	// log new token record
	log.Printf("Saved token:\n[ %s | %s | %d ]\n", sToken, params.URL, params.Exp)

	// send response
	w.Write(resp)
}

// myMUX selects the handler function according to request URL
func myMUX(w http.ResponseWriter, r *http.Request) {
	path := r.URL.Path
	switch path {
	case "/":
		// request for health-check
		log.Println("health-check")
		home(w)
	case "/token":
		// request for new short url/token
		log.Println("request for token")
		getNewToken(w, r)
	case "/favicon.ico":
		// Chromium make such requests together with request for redirect to show the site icon on tab header
		// In this code it is used for health check
		return
	default:
		// all the rest are requests for redirect (probably)
		log.Println("request for redirect")
		redirect(w, r, path[1:])
	}
}

func main() {

	var err error

	log.SetPrefix("URLshortener: ")

	// get the configuratin variables
	err = readConfig(".cnf.json")
	if err != nil {
		log.Fatal(err)
	}

	// create new TokenDB interface
	tokenDB, err = TokenDBNew()
	if err != nil {
		log.Fatal(err)
	}

	// register the handler
	http.HandleFunc("/", myMUX)

	// start server
	log.Println("starting server at", CONFIG.ListenHostPort)
	Server = &http.Server{
		Addr:    CONFIG.ListenHostPort,
		Handler: nil}
	log.Println(Server.ListenAndServe())
}
