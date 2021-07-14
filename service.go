package main

// URLshortener is a microservice to shorten long URLs
// and to handle the redirection by generated short URLs.
//
// See details in README.md
//
// This file contains service handler interface

import (
	"context"
	_ "embed"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log"
	"net/http"
	"strings"
	"sync/atomic"
	"time"

	_ "embed"
)

const (
	// homePage is a simple home page template to display on health check request
	homePage = `
<html>
	<head>
		<title>URL shortener</title>
	</head>
	<body>
		<h1>Home page of URLshortener</h1>
		<br>
		URLshortener %s
		<br><br>
		Service status: healthy, %d attempts per %d ms
		<br><br>
		<a href=/ui/generate>Create short URL manually</a>
		<br><br><br><br>
		See sources at <a href="https://github.com/slytomcat/URLshortener">https://github.com/slytomcat/URLshortener</a>
	</body>
</html>`
	// generatePage - is a template for short URL generation
	generatePage = `	 
<html>
	<head>
	<title>Short URL generator</title>
	</head>
	<body>
	   <br>
	   %s
	   <br>
	   <form action="/ui/generate" name=f method="GET">
		   <input maxLength=1024 size=70 name=s value="" title="URL to be shortened">
		   <input type=submit value="get short URL">
	   </form>   
	</body>
</html>`
	generatorPagePart = `
<br><br>
Short URL: %s
<br><br>
QR code for short URL:
<br>
<img src="http://chart.apis.google.com/chart?chs=300x300&cht=qr&choe=UTF-8&chl=%s" />
<br>
<br>
Short URL lifetime: %d days
<br>`
)

var (

	// favicon is binary image (PNG) that is a response on /favicon.ico request
	//go:embed favicon.png
	favicon []byte
)

// ServiceHandler interface
type ServiceHandler interface {
	ServeHTTP(http.ResponseWriter, *http.Request) // http server handler function
	healthCheck() error                           // Health-check function
	start() error                                 // Service start method
	stop()                                        // Service stop method
}

// serviceHandler is an istance of ServiceHandler interface
type serviceHandler struct {
	tokenDB    TokenDB      // Database interface
	shortToken ShortToken   // Short token generator
	config     *Config      // service configuration
	server     *http.Server // service server
	attempts   int32        // calculated number of attempts during time-out
}

// ServeHTTP implement simpme mux that selects the handler function according to request URL
func (s *serviceHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	log.Println("access from:", r.RemoteAddr, r.Method, r.RequestURI, r.Header)

	// read the request body
	body, err := io.ReadAll(r.Body)
	if err != nil {
		log.Printf("request body reading error: %v", err)
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	switch r.Method + r.URL.Path {
	case "GET/":
		// request for health-check
		s.home(w, r)
	case "POST/api/v1/token":
		// request for new short url/token
		s.new(w, r, body)
	case "POST/api/v1/expire":
		// request for new short url/token
		s.expire(w, r, body)
	case "GET/ui/generate":
		// UI short URL generation page
		s.generate(w, r)
	case "GET/favicon.ico":
		// WEB-browsers make such requests together with the main request in order to show the site icon on tab header
		// In this code it is used for health check (as point to redirect from short url)
		w.Write(favicon)
	default:
		// all the rest GET requests are requests for redirect (probably)
		if r.Method == "GET" {
			s.redirect(w, r)
		} else {
			log.Printf("bad method/path: %s %s", r.Method, r.URL.Path)
			w.WriteHeader(http.StatusBadRequest)
		}
	}
}

// generate is UI short URL|QR generator
func (s *serviceHandler) generate(w http.ResponseWriter, r *http.Request) {
	rMess := fmt.Sprintf("UI generate request from %s (%s)", r.RemoteAddr, r.Referer())
	// check that service mode allows this request
	if s.config.Mode&disableUI != 0 {
		log.Printf("%s: this request is disabled by current service mode\n", rMess)
		// send 404 response
		http.NotFound(w, r)
		return
	}
	url := r.FormValue("s")
	part := ""

	if url != "" {
		// TO DO: make more sophisticated check for URL
		// if URL provided then make short URL for it
		sToken, err := s.generateToken(url, s.config.DefaultExp)

		if err != nil {
			log.Printf("%s: token generation error: %v", rMess, err)
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
		sURL := s.config.ShortDomain + "/" + sToken
		part = fmt.Sprintf(generatorPagePart, sURL, sURL, s.config.DefaultExp)
		rMess = fmt.Sprintf("%s: new token generated: %s", rMess, sToken)
	}

	// display results
	w.Write([]byte(fmt.Sprintf(generatePage, part)))
	log.Printf("%s: ui interface displaed", rMess)

}

/* test for test env:
curl -i -v http://localhost:8080/
*/

// Home shows simple home page if self-check succesfuly passed
func (s *serviceHandler) home(w http.ResponseWriter, r *http.Request) {
	rMess := fmt.Sprintf("health-check request from %s (%s)", r.RemoteAddr, r.Referer())
	// Perform self-test
	if err := s.healthCheck(); err != nil {
		// report error
		log.Printf("%s: error: %v\n", rMess, err)
		w.WriteHeader(http.StatusInternalServerError)
	} else {
		// log self-test results
		log.Printf("%s: success\n", rMess)
		// show the home page if self-test was successfully passed
		w.Write([]byte(fmt.Sprintf(
			homePage,
			version,
			atomic.LoadInt32(&s.attempts),
			s.config.Timeout)))
	}
}

// healthCheck performs full self-test of service in all service modes
func (s *serviceHandler) healthCheck() error {
	// self-test makes three requests:
	// 1. request for short URL
	// 2. request for redirect from short to long URL
	// 3. request to expire the token (received in the first request)

	// long URL for sef-check redirect
	url := s.config.ShortDomain + "/favicon.ico"

	var (
		// short URL request's replay parameters
		repl struct {
			URL   string `json:"url"`
			Token string `json:"token"`
		}
		err error
	)

	// self-test part 1: get short URL
	if s.config.Mode&disableShortener != 0 {
		// short token for this scenario
		sToken := "Debug.Token"
		_ = s.tokenDB.Delete(sToken)

		// use tokenDB inteface as web-interface is locked in this service mode
		if ok, err := s.tokenDB.Set(sToken, "http://"+url, 1); err != nil || !ok {
			return fmt.Errorf("new token creation error: %w, ok: %v", err, ok)
		}
		// store results
		repl.Token = sToken
		repl.URL = s.config.ShortDomain + "/" + repl.Token
	} else {
		// make the HTTP request for new token
		resp, err := http.Post("http://"+s.config.ListenHostPort+"/api/v1/token", "application/json",
			strings.NewReader(`{"url": "`+url+`","exp": 1}`))
		if err != nil {
			return fmt.Errorf("new token request error: %w", err)
		}
		defer resp.Body.Close()

		// check response status code
		if resp.StatusCode != http.StatusOK {
			return fmt.Errorf("new token request: unexpected responce status: %v", resp.StatusCode)
		}

		// read response body
		buf, err := io.ReadAll(resp.Body)
		if err != nil {
			return fmt.Errorf("new token response body reading error : %w", err)
		}

		// parse response body
		if err = json.Unmarshal(buf, &repl); err != nil {
			return fmt.Errorf("new token response body parsing error: %w", err)
		}

		// check receved token
		if repl.Token == "" {
			return errors.New("empty token returned")
		}
	}

	// self-test part 2: check redirect
	rURL := "" // vaiable to store redirect URL
	if s.config.Mode&disableRedirect != 0 {
		// use tokenDB interface as web-interface is locked in this service mode
		rURL, err = s.tokenDB.Get(repl.Token)
		if err != nil {
			return fmt.Errorf("URL receiving error: %w", err)
		}

	} else {
		// try to make the HTTP request for redirect by short URL
		resp2, err := http.Get("http://" + repl.URL)
		if err != nil {
			return fmt.Errorf("redirect request error: %w", err)
		}
		defer resp2.Body.Close()

		// check redirect response status
		if resp2.StatusCode != http.StatusOK {
			return fmt.Errorf("redirect request: unexpected responce status: %v", resp2.StatusCode)
		}

		// get redirection URL
		rURL = resp2.Request.URL.String()
	}
	// check redirection URL
	if rURL != "http://"+url {
		return fmt.Errorf("wrong redirection URL: expected %s, receved %v", url, rURL)
	}

	// self-test part 3: make received token as expired
	if s.config.Mode&disableExpire != 0 {
		// use tokenDB interface as web-interface is locked in this service mode
		if err := s.tokenDB.Expire(repl.Token, -1); err != nil {
			return fmt.Errorf("expire request error: %w", err)
		}
	} else {
		// make the HTTP request to expire token
		resp3, err := http.Post("http://"+s.config.ListenHostPort+"/api/v1/expire", "application/json",
			strings.NewReader(`{"token": "`+repl.Token+`","exp":-1}`))
		if err != nil {
			return fmt.Errorf("expire request error: %w", err)
		}
		defer resp3.Body.Close()

		// check response status
		if resp3.StatusCode != http.StatusOK {
			return fmt.Errorf("expire request: unexpected response status: %v", resp3.StatusCode)
		}
	}

	return nil
}

/* test for test env:
curl -i -v http://localhost:8080/<token>
*/

// Redirect handles redirection to URL that was stored for the specified token
func (s *serviceHandler) redirect(w http.ResponseWriter, r *http.Request) {

	sToken := r.URL.Path[1:] // GET and POST always contain at least "/" in URL
	rMess := fmt.Sprintf("redirect request from %s (%s), token: %s", r.RemoteAddr, r.Referer(), sToken)

	// check that service mode allows this request
	if s.config.Mode&disableRedirect != 0 {
		log.Printf("%s: this request is disabled by current service mode\n", rMess)
		// send 404 response
		http.NotFound(w, r)
		return
	}

	// // check the token
	// err := s.shortToken.Check(sToken)
	// if err != nil {
	// 	log.Printf("%s: token check failed: %v\n", rMess, err)
	// 	http.NotFound(w, r)
	// 	return
	// }

	// get the long URL
	longURL, err := s.tokenDB.Get(sToken)
	if err != nil {
		log.Printf("%s: token was not found\n", rMess)
		// send 404 response
		http.NotFound(w, r)
		return
	}

	// log the request results
	log.Printf("%s: redirected to %s\n", rMess, longURL)

	// respond by redirect
	http.Redirect(w, r, longURL, http.StatusFound)
}

/* test for test env:
curl -v POST -H "Content-Type: application/json" -d '{"url":"<long url>","exp":10}' http://localhost:8080/api/v1/token
*/

// new handle the new token creation for passed url and sets expiration for it
func (s *serviceHandler) new(w http.ResponseWriter, r *http.Request, body []byte) {
	// TODO: check some authorisation ???

	rMess := fmt.Sprintf("token request from %s (%s)", r.RemoteAddr, r.Referer())

	// Check that service mode allows this request
	if s.config.Mode&disableShortener != 0 {
		log.Printf("%s: this request is disabled by current service mode\n", rMess)
		// request is not supported: send 404 response
		http.NotFound(w, r)
		return
	}

	// the requst parameters structure
	var params struct {
		URL string `json:"url"`           // long URL
		Exp int    `json:"exp,omitempty"` // Expiration
	}

	// parse body to parameters structure
	err := json.Unmarshal(body, &params)
	if err != nil || params.URL == "" {
		log.Printf("%s: bad request parameters:%s", rMess, body)
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	// log received params
	rMess += fmt.Sprintf(" parameters: '%s', %d", params.URL, params.Exp)

	sToken, err := s.generateToken(params.URL, params.Exp)
	// handle token generation error
	if err != nil {
		log.Printf("%s: token generation error:%s", rMess, body)
		if strings.Contains(err.Error(), "creation error") {
			w.WriteHeader(http.StatusRequestTimeout)
		} else {
			w.WriteHeader(http.StatusInternalServerError)
		}
		return
	}

	// make response body
	resp, _ := json.Marshal(
		struct {
			Token string `json:"token"` // token
			URL   string `json:"url"`   // short URL
		}{
			Token: sToken,
			URL:   s.config.ShortDomain + "/" + sToken,
		})

	// log new token request information
	log.Printf("%s: URL saved, token: %s , exp: %d\n", rMess, sToken, params.Exp)

	// send response
	w.Write(resp)
}

// generateToken generates token or writes the error in w
func (s *serviceHandler) generateToken(url string, exp int) (string, error) {
	// Using many attempts to store the new random token dramatically increases maximum amount of
	// used tokens since:
	// probability of the failure of n attempts = (probability of failure of single attempt)^n.

	// Limit number of attempts by time not by count

	// Count attempts and time for reports
	var attempt int64
	var startTime time.Time
	var err error

	// add referenece type if it is missing
	if !strings.HasPrefix(strings.ToLower(url), "http") {
		url = "http://" + url
	}

	// set the default expiration if it is not passed
	if exp == 0 {
		exp = s.config.DefaultExp
	}

	// Calculate statistics and report if some dangerous situation appears
	defer func() {
		elapsedTime := time.Since(startTime)
		// perform statistical calculation and reporting in another go-routine
		go func() {
			if attempt > 0 {
				MaxAtt := attempt * int64(s.config.Timeout) * 1000000 / elapsedTime.Nanoseconds()
				// use atomic to avoid race conditions
				atomic.StoreInt32(&s.attempts, int32(MaxAtt))
				// report warnings of some not good measurements
				if MaxAtt*3/4 < attempt {
					log.Printf("Warning: Measured %d attempts for %d ns. Calculated %d max attempts per %d ms\n", attempt, elapsedTime, MaxAtt, s.config.Timeout)
				}
				if MaxAtt > 0 && MaxAtt < 10 {
					log.Printf("Warning: Too low number of attempts: %d per timeout (%d ms)\n", MaxAtt, s.config.Timeout)
				}
			}
		}()
	}()

	sToken := ""

	// make time-out chanel
	stop := time.After(time.Millisecond * time.Duration(s.config.Timeout))

	// Remember starting time
	startTime = time.Now()

	// start trying to store new token
	for ok := false; !ok; {
		select {
		case <-stop:
			// timeout exceeded
			return "", fmt.Errorf("token creation error: %v, ok: %v", err, ok)
		default:
			// get short token
			sToken = s.shortToken.Get()
			// count attempts
			attempt++
			// store token in DB
			ok, err = s.tokenDB.Set(sToken, url, exp)
			if err != nil {
				return "", fmt.Errorf("token storing error: %v", err)
			}
		}
	}

	return sToken, nil
}

/* test for test env:
curl -v POST -H "Content-Type: application/json" -d '{"token":"<token>","exp":<exp>}' http://localhost:8080/api/v1/expire
*/

// expire makes token-longURL record as expired
func (s *serviceHandler) expire(w http.ResponseWriter, r *http.Request, body []byte) {

	rMess := fmt.Sprintf("expire request from %s (%s)", r.RemoteAddr, r.Referer())

	// Check that service mode allows this request
	if s.config.Mode&disableExpire != 0 {
		log.Printf("%s: this request is disabled by current service mode\n", rMess)
		// request is not supported: send 404 response
		http.NotFound(w, r)
		return
	}

	// make the requst parameters structure
	var params struct {
		Token string `json:"token"`         // Token of short URL token
		Exp   int    `json:"exp,omitempty"` // Expiration
	}

	// parse JSON from body to parameters structure
	err := json.Unmarshal(body, &params)
	if err != nil || params.Token == "" {
		log.Printf("%s: bad request parameters:%s", rMess, body)
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	// update token expiration
	err = s.tokenDB.Expire(params.Token, params.Exp)
	if err != nil {
		log.Printf("%s: updating token expiration error: %s", rMess, err)
		w.WriteHeader(http.StatusNotModified)
		return
	}

	// log request results
	log.Printf("%s: token expiration of %s has set to %d\n", rMess, params.Token, params.Exp)

	// send response
	w.WriteHeader(http.StatusOK)
}

// Start returns started server
func (s *serviceHandler) start() error {

	log.Println("starting server at", s.config.ListenHostPort)

	return s.server.ListenAndServe()
}

// Stop performs graceful shutdown of server and database interfaces
// It reports success shutdown via serviceHandler.exit chanel
func (s *serviceHandler) stop() {
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()
	err := s.server.Shutdown(ctx)
	if err != nil {
		log.Printf("HTTP server shutdown error: %v", err)
	}
}

// NewHandler returns new service handler
func NewHandler(config *Config, tokenDB TokenDB, shortToken ShortToken) ServiceHandler {

	// make handler
	handler := &serviceHandler{
		tokenDB:    tokenDB,
		shortToken: shortToken,
		config:     config,
		server:     nil,
		attempts:   0,
	}

	// create server
	handler.server = &http.Server{
		Addr:    config.ListenHostPort,
		Handler: handler,
	}

	return handler
}
