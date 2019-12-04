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
	"fmt"
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

func healthCheck() error {
	url := "http://" + CONFIG.ShortDomain + "/favicon.ico"
	var repl struct {
		URL   string `json:"url"`
		Token string `json:"token"`
	}
	var err error

	// self-test part 1: get short URL
	if CONFIG.Mode == 2 {
		// use tokenDB inteface as web-interface is locked in this mode
		if repl.Token, err = tokenDB.New(url, 1); err != nil {
			return err
		}
		repl.URL = CONFIG.ShortDomain + "/" + repl.Token
	} else {
		// make the HTTP request for new token
		resp, err := http.Post("http://"+CONFIG.ListenHostPort+"/token", "application/json",
			strings.NewReader(`{"url": "`+url+`", "exp": "1"}`))
		if err != nil {
			return err
		}
		defer resp.Body.Close()

		// read response body
		buf := make([]byte, resp.ContentLength)
		_, err = resp.Body.Read(buf)
		if err != nil && !errors.Is(err, io.EOF) {
			return err
		}

		// parse response body
		if err = json.Unmarshal(buf, &repl); err != nil {
			return err
		}
	}

	// self-test part 2: check redirect
	if CONFIG.Mode == 1 {
		// use tokenDB interface as web-interface is locked in this mode
		if _, err = tokenDB.Get(repl.Token); err != nil {
			return err
		}
	} else {
		// try to make the HTTP request for redirect by short URL
		resp2, err := http.Get("http://" + repl.URL)
		if err != nil {
			return err
		}
		defer resp2.Body.Close()

		// check redirect response status
		if resp2.StatusCode != http.StatusOK {
			return err
		}
	}

	// self-test part 3: expire received token
	if err := tokenDB.Expire(repl.Token); err != nil {
		return err
	}
	return nil
}

/* test for test env:
curl -i -v http://localhost:8080/
*/

// Home shows simple home page
func home(w http.ResponseWriter, r *http.Request) {
	log.Printf("health-check request from %s (%s)\n", r.RemoteAddr, r.Referer())
	// Perform self-test
	if err := healthCheck(); err != nil {
		// report error
		w.WriteHeader(http.StatusInternalServerError)
	} else {
		// show the home page if self-test was successfully passed
		w.Write(homePage)
	}
}

/* test for test env:
curl -i -v http://localhost:8080/<token>
*/

// Redirect handles redirection to URL that was stored for the specified token
func redirect(w http.ResponseWriter, r *http.Request) {
	sToken := r.URL.Path[1:]
	rMess := fmt.Sprintf("redirect request from %s (%s), token: %s", r.RemoteAddr, r.Referer(), sToken)

	// check that service mode allows this request
	if CONFIG.Mode == 1 {
		log.Printf("%s: this request is disabled by service mode\n", rMess)
		// send 404 response
		http.NotFound(w, r)
		return
	}

	// get the long URL
	longURL, err := tokenDB.Get(r.URL.Path[1:])
	if err != nil {
		log.Printf("%s: token is not found\n", rMess)
		// send 404 response
		http.NotFound(w, r)
		return
	}
	log.Printf("%s: redirected to %s\n", rMess, longURL)

	// make redirect response
	http.Redirect(w, r, longURL, http.StatusMovedPermanently)
}

/* test for test env:
curl -v POST -H "Content-Type: application/json" -d '{"url":"https://www.w3schools.com/html/html_forms.asp","exp":"10"}' http://localhost:8080/token
*/

// getNewToken handle the new token creation for passed url and sets expiration for it
func getNewToken(w http.ResponseWriter, r *http.Request) {
	// ????: check some authorisation???

	rMess := fmt.Sprintf("token request from %s (%s)", r.RemoteAddr, r.Referer())

	// Check that service mode allows this request
	if CONFIG.Mode == 2 {
		log.Printf("%s: this request is disabled by service mode\n", rMess)
		// request is not supported: send 404 response
		http.NotFound(w, r)
		return
	}

	// read the request body
	buf := make([]byte, r.ContentLength)
	_, err := r.Body.Read(buf)
	if err != nil && !errors.Is(err, io.EOF) {
		log.Printf("%s: request body reading error: %v", rMess, err)
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	// parse JSON to parameters structure

	// the requst parameters structure
	var params struct {
		URL string `json:"url"`                  // long URL
		Exp int    `json:"exp,string,omitempty"` // Expiration
	}

	err = json.Unmarshal(buf, &params)
	if err != nil || params.URL == "" {
		log.Printf("%s: bad request parameters:%s", rMess, buf)
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	// log received params
	rMess += fmt.Sprintf(" parameters: '%s', %d", params.URL, params.Exp)

	// set the default expiration if it is not passed
	if params.Exp == 0 {
		params.Exp = CONFIG.DefaultExp
	}

	// create new token
	sToken, err := tokenDB.New(params.URL, params.Exp)
	if err != nil {
		log.Printf("%s: token creation error: %v\n", rMess, err)
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
		log.Printf("%s: response body marshaling error: %v\n", rMess, err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	// log new token record
	log.Printf("%s: token saved: [ %s | %s | %d ]\n", rMess, sToken, params.URL, params.Exp)

	// send response
	w.Write(resp)
}

// myMUX selects the handler function according to request URL
func myMUX(w http.ResponseWriter, r *http.Request) {
	switch r.URL.Path {
	case "/":
		// request for health-check
		home(w, r)
	case "/token":
		// request for new short url/token
		getNewToken(w, r)
	case "/favicon.ico":
		// Chromium make such requests together with request for redirect to show the site icon on tab header
		// In this code it is used for health check
		return
	default:
		// all the rest are requests for redirect (probably)
		redirect(w, r)
	}
}

func main() {

	var err error

	// set logging format
	log.SetFlags(log.Ldate | log.Ltime | log.Lshortfile)

	// get the configuratin variables
	err = readConfig("cnf.json")
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
