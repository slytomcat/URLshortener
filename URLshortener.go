package main

// URLshortener is a microservice to shorten long URLs
// and to handle the redirection by generated short URLs.
//
// Request for short URL:
// URL: <host>[:<port>]/api/v1/token
// Method: POST
// Body: JSON with following parameters:
//   url - URL to shorten, mandatory
//   exp - short URL expiration in days, optional
// Success response: JSON with following parameters:
//   token - token for short URL
//   url - short URL
//
// Request to expire token
// URL: <host>[:<port>]/api/v1/expire
// Method: POST
// Body: JSON with following parameters:
//   token - token for short URL
// Success response: HTTP 200 OK
//
// Redirect to long URL:
// URL: <host>[:<port>]/<token> - URL from response on request for short URL
// Method: GET
// No parameters
// Success response contain the redirection to long URL
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

const (
	// Service modes
	disableRedirect  = 1 << iota // disable redirect request
	disableShortener             // disable request for short URL
	disableExpire                // disable expire request
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

// healthCheck performs full self-test of service in all service modes
func healthCheck() error {

	// url for sef-check redirect
	url := "http://" + CONFIG.ShortDomain + "/favicon.ico"

	// replay parameters
	var repl struct {
		URL   string `json:"url"`
		Token string `json:"token"`
	}
	var err error

	// self-test part 1: get short URL
	if CONFIG.Mode&disableShortener != 0 {
		// use tokenDB inteface as web-interface is locked in this mode
		if repl.Token, err = tokenDB.New(url, 1); err != nil {
			return err
		}
		repl.URL = CONFIG.ShortDomain + "/" + repl.Token
	} else {
		// make the HTTP request for new token
		resp, err := http.Post("http://"+CONFIG.ListenHostPort+"/api/v1/token", "application/json",
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
	if CONFIG.Mode&disableRedirect != 0 {
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
	if CONFIG.Mode&disableExpire != 0 {
		if err := tokenDB.Expire(repl.Token, -1); err != nil {
			return err
		}
	} else {
		// make the HTTP request to expire token
		resp3, err := http.Post("http://"+CONFIG.ListenHostPort+"/api/v1/expire", "application/json",
			strings.NewReader(`{"token": "`+repl.Token+`"}`))
		if err != nil {
			return err
		}
		defer resp3.Body.Close()

		// check response status
		if resp3.StatusCode != http.StatusOK {
			return err
		}
	}
	return nil
}

/* test for test env:
curl -i -v http://localhost:8080/
*/

// Home shows simple home page if self-check succesfuly passed
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
	if CONFIG.Mode&disableRedirect != 0 {
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
curl -v POST -H "Content-Type: application/json" -d '{"url":"https://www.w3schools.com/html/html_forms.asp","exp":"10"}' http://localhost:8080/api/v1/token
*/

// getNewToken handle the new token creation for passed url and sets expiration for it
func getNewToken(w http.ResponseWriter, r *http.Request) {
	// ????: check some authorisation???

	rMess := fmt.Sprintf("token request from %s (%s)", r.RemoteAddr, r.Referer())

	// Check that service mode allows this request
	if CONFIG.Mode&disableShortener != 0 {
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
	log.Printf("%s: URL saved, token: %s , exp: %d\n", rMess, sToken, params.Exp)
	// send response
	w.Write(resp)
}

/* test for test env:
curl -v POST -H "Content-Type: application/json" -d '{"token":"<token>"}' http://localhost:8080/api/v1/expire
*/

// expireToken makes token-longURL record as expired
func expireToken(w http.ResponseWriter, r *http.Request) {

	rMess := fmt.Sprintf("expire request from %s (%s)", r.RemoteAddr, r.Referer())

	// Check that service mode allows this request
	if CONFIG.Mode&disableExpire != 0 {
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
		Token string `json:"token"`                // Token of short URL token
		Exp   int    `json:"exp,string,omitempty"` // Expiration
	}

	err = json.Unmarshal(buf, &params)
	if err != nil || params.Token == "" {
		log.Printf("%s: bad request parameters:%s", rMess, buf)
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	// update token expiration
	err = tokenDB.Expire(params.Token, params.Exp)
	if err != nil {
		log.Printf("%s: updating token expiration error: %s", rMess, err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	// log result
	log.Printf("%s: token expiration of %s has set to %d\n", rMess, params.Token, params.Exp)
	// send response
	w.WriteHeader(http.StatusOK)
}

// myMUX selects the handler function according to request URL
func myMUX(w http.ResponseWriter, r *http.Request) {
	switch r.URL.Path {
	case "/":
		// request for health-check
		home(w, r)
	case "/api/v1/token":
		// request for new short url/token
		getNewToken(w, r)
	case "/api/v1/expire":
		// request for new short url/token
		expireToken(w, r)
	case "/favicon.ico":
		// WEB-brousers make such requests together with main request to show the site icon on tab header
		// In this code it is used for health check
		return
	default:
		// all the rest are requests for redirect (probably)
		redirect(w, r)
	}
}

func main() {

	// set logging format
	log.SetFlags(log.Ldate | log.Ltime | log.Lshortfile)

	// get the configuratin variables
	err := readConfig("cnf.json")
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

	// create and start server
	log.Println("starting server at", CONFIG.ListenHostPort)
	Server = &http.Server{
		Addr:    CONFIG.ListenHostPort,
		Handler: nil}
	log.Println(Server.ListenAndServe())
}
