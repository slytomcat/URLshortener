package main

import (
	"bytes"
	"encoding/json"
	"errors"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"strings"
	"testing"
	"time"
)

// try to start service
func Test50mainStart(t *testing.T) {
	logger := log.Writer()
	r, w, _ := os.Pipe()
	log.SetOutput(w)

	go main()
	time.Sleep(time.Second * 3)

	w.Close()
	log.SetOutput(logger)
	buf, err := ioutil.ReadAll(r)
	if err != nil {
		t.Error(err)
	}
	if !bytes.Contains(buf, []byte("starting server at")) {
		t.Errorf("received not expected output: %s", buf)
	}
	log.Printf("%s", buf)

}

// Full success test: get short URL and make redirect by it
func Test55MainFullSuccess(t *testing.T) {
	// use health check as long url
	resp, err := http.Post("http://"+CONFIG.ListenHostPort+"/api/v1/token", "application/json",
		strings.NewReader(`{"url": "http://`+CONFIG.ShortDomain+`/favicon.ico", "exp": "3"}`))
	if err != nil {
		t.Errorf("token request error: %v", err)
	}
	defer resp.Body.Close()
	buf := make([]byte, resp.ContentLength)
	_, err = resp.Body.Read(buf)
	if err != nil && !errors.Is(err, io.EOF) {
		t.Errorf("response body reading error: %v", err)
	}
	var repl struct {
		URL   string `json:"url"`
		Token string `json:"token"`
	}
	err = json.Unmarshal(buf, &repl)
	if err != nil {
		t.Errorf("response body parsing error: %v", err)
	}

	resp2, err := http.Get("http://" + repl.URL)
	if err != nil {
		t.Errorf("redirect request error: %v", err)
	}
	defer resp2.Body.Close()

	if resp2.ContentLength != 0 {
		t.Error("response body not empty")
	}

	if resp2.StatusCode != http.StatusOK {
		t.Errorf("wrong status : %d", resp.StatusCode)
	}

	resp3, err := http.Post("http://"+CONFIG.ListenHostPort+"/api/v1/expire", "application/json",
		strings.NewReader(`{"token": "`+repl.Token+`"}`))
	if err != nil {
		t.Errorf("expire request error: %v", err)
	}
	defer resp3.Body.Close()

	if resp3.StatusCode != http.StatusOK {
		t.Errorf("wrong status : %d", resp.StatusCode)
	}

}

// test health check
func Test57MainHome(t *testing.T) {
	resp, err := http.Get("http://" + CONFIG.ListenHostPort)
	if err != nil {
		t.Errorf("health check request error: %v", err)
	}
	defer resp.Body.Close()
	buf := make([]byte, resp.ContentLength)
	_, err = resp.Body.Read(buf)
	if err != nil && !errors.Is(err, io.EOF) {
		t.Errorf("response body reading error: %v", err)
	}

	if !bytes.Contains(buf, []byte("Home page of URLshortener")) {
		t.Error("wrong response on health check request")
	}

	if resp.StatusCode != http.StatusOK {
		t.Errorf("wrong status : %d", resp.StatusCode)
	}

}

// test request for short URL with empty request body
func Test60MainBadRequest(t *testing.T) {
	resp, err := http.Post("http://"+CONFIG.ListenHostPort+"/api/v1/token", "application/json",
		strings.NewReader(``))
	if err != nil {
		t.Errorf("token request error: %v", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusBadRequest {
		t.Errorf("wrong status : %d", resp.StatusCode)
	}
}

// test request for short URL with empty JSON
func Test61MainBadRequest2(t *testing.T) {
	resp, err := http.Post("http://"+CONFIG.ListenHostPort+"/api/v1/token", "application/json",
		strings.NewReader(`{}`))
	if err != nil {
		t.Errorf("token request error: %v", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusBadRequest {
		t.Errorf("wrong status : %d", resp.StatusCode)
	}
}

//test request for short URL without expiration in request
func Test62MainGetTokenWOexp(t *testing.T) {
	DEBUG = true
	defer func() { DEBUG = false }()

	tx, _ := tokenDB.DB.Begin()
	_, err := tx.Exec("DELETE FROM urls WHERE token=?", DEBUGToken)
	if err != nil {
		t.Errorf("Can't clear table: %v", err)
	}
	tx.Commit()

	resp, err := http.Post("http://"+CONFIG.ListenHostPort+"/api/v1/token", "application/json",
		strings.NewReader(`{"url": "http://`+CONFIG.ShortDomain+`"}`))
	if err != nil {
		t.Errorf("token request error: %v", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		t.Errorf("wrong status : %d", resp.StatusCode)
	}
}

// request expire without parameters
func Test64MainExpireTokenWOparams(t *testing.T) {

	resp, err := http.Post("http://"+CONFIG.ListenHostPort+"/api/v1/expire", "application/json",
		strings.NewReader(`{}`))
	if err != nil {
		t.Errorf("expire request error: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusBadRequest {
		t.Errorf("wrong status : %d", resp.StatusCode)
	}

}

// request expire for not existing token
func Test65MainExpireNotExistingToken(t *testing.T) {

	resp, err := http.Post("http://"+CONFIG.ListenHostPort+"/api/v1/expire", "application/json",
		strings.NewReader(`{"token":"$%#@*"}`)) // use non Base64 symbols
	if err != nil {
		t.Errorf("expire request error: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusNotModified {
		t.Errorf("wrong status : %d", resp.StatusCode)
	}

}

// try to get the same (debugging) token twice
func Test68MainGetTokenTwice(t *testing.T) {
	DEBUG = true
	defer func() { DEBUG = false }()
	resp, err := http.Post("http://"+CONFIG.ListenHostPort+"/api/v1/token", "application/json",
		strings.NewReader(`{"url": "http://`+CONFIG.ShortDomain+`"}`))
	if err != nil {
		t.Errorf("token request error: %v", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusGatewayTimeout {
		t.Errorf("wrong status : %d", resp.StatusCode)
	}
}

// test redirect with wrong token
func Test70Main404(t *testing.T) {
	resp, err := http.Get("http://" + CONFIG.ListenHostPort + "/not_existing_token")
	if err != nil {
		t.Errorf("not-existing token request error: %v", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusNotFound {
		t.Errorf("wrong status : %d", resp.StatusCode)
	}
}

// try unsupported request in mode = disableRedirect
func Test73MainServiceModeDisableRedirect(t *testing.T) {
	CONFIG.Mode = disableRedirect

	resp, err := http.Get("http://" + CONFIG.ListenHostPort + "/" + DEBUGToken)
	if err != nil {
		t.Errorf("redirect request error: %v", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusNotFound {
		t.Errorf("wrong status : %d", resp.StatusCode)
	}
}

// try unsupported request in mode = disableShortener
func Test73MainServiceModeDisableShortener(t *testing.T) {
	CONFIG.Mode = disableShortener

	resp, err := http.Post("http://"+CONFIG.ListenHostPort+"/api/v1/token", "application/json",
		strings.NewReader(`{"url": "http://someother.url"}`))
	if err != nil {
		t.Errorf("token request error: %v", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusNotFound {
		t.Errorf("wrong status : %d", resp.StatusCode)
	}
}

// try unsupported request in mode = disableExpire
func Test74MainServiceModeDisableExpire(t *testing.T) {
	CONFIG.Mode = disableExpire

	resp, err := http.Post("http://"+CONFIG.ListenHostPort+"/api/v1/expire", "application/json",
		strings.NewReader(`{"token": "`+DEBUGToken+`"}`))
	if err != nil {
		t.Errorf("expire request error: %v", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusNotFound {
		t.Errorf("wrong status : %d", resp.StatusCode)
	}
}

// try health check in service mode disableRedirect
func Test75MainHealthCheckModeDisableRedirect(t *testing.T) {
	CONFIG.Mode = disableRedirect

	resp, err := http.Get("http://" + CONFIG.ListenHostPort)
	if err != nil {
		t.Errorf("health check in disableRedirect mode request error: %v", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		t.Errorf("wrong status : %d", resp.StatusCode)
	}
}

// try health check in service mode disableShortener
func Test77MainHealthCheckModeDisableShortener(t *testing.T) {
	CONFIG.Mode = disableShortener

	resp, err := http.Get("http://" + CONFIG.ListenHostPort)
	if err != nil {
		t.Errorf("health check in disableShortener mode request error: %v", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		t.Errorf("wrong status : %d", resp.StatusCode)
	}
}

// try health check in service mode disableExpire
func Test78MainHealthCheckModeDisableExpire(t *testing.T) {
	CONFIG.Mode = disableExpire

	resp, err := http.Get("http://" + CONFIG.ListenHostPort)
	if err != nil {
		t.Errorf("health check in disableExpire mode request error: %v", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		t.Errorf("wrong status : %d", resp.StatusCode)
	}
}

// try to stop service
func Test99MainKill(t *testing.T) {
	logger := log.Writer()
	r, w, _ := os.Pipe()
	log.SetOutput(w)

	Server.Close()
	time.Sleep(time.Second * 3)

	w.Close()
	log.SetOutput(logger)
	buf, err := ioutil.ReadAll(r)
	if err != nil {
		t.Error(err)
	}
	if !bytes.Contains(buf, []byte("http: Server closed")) {
		t.Errorf("received not expected output: %s", buf)
	}
	log.Printf("%s", buf)
}
