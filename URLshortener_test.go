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
	log.Printf("%s", buf[20:])

}

// Full success test: get short URL and make redirect by it
func Test55MainFullSuccess(t *testing.T) {
	// use health check as long url
	resp, err := http.Post("http://"+CONFIG.ListenHostPort+"/token", "application/json",
		strings.NewReader(`{"url": "http://`+CONFIG.ShortDomain+`", "exp": "3"}`))
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
	buf = make([]byte, resp2.ContentLength)
	_, err = resp2.Body.Read(buf)
	if err != nil && !errors.Is(err, io.EOF) {
		t.Errorf("response body reading error: %v", err)
	}

	if !bytes.Contains(buf, []byte("Home page of URLshortener")) {
		t.Error("wrong response body on helth check request")
	}

	if resp2.StatusCode != http.StatusOK {
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
	resp, err := http.Post("http://"+CONFIG.ListenHostPort+"/token", "application/json",
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
	resp, err := http.Post("http://"+CONFIG.ListenHostPort+"/token", "application/json",
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

	resp, err := http.Post("http://"+CONFIG.ListenHostPort+"/token", "application/json",
		strings.NewReader(`{"url": "http://`+CONFIG.ShortDomain+`"}`))
	if err != nil {
		t.Errorf("token request error: %v", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		t.Errorf("wrong status : %d", resp.StatusCode)
	}
}

// try to get the same (debugging) token twice
func Test62MainGetTokenTwice(t *testing.T) {
	DEBUG = true
	defer func() { DEBUG = false }()
	resp, err := http.Post("http://"+CONFIG.ListenHostPort+"/token", "application/json",
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

// try unsupported request in mode = 1
func Test73MainServiceMode1(t *testing.T) {
	CONFIG.Mode = 1

	resp, err := http.Get("http://" + CONFIG.ListenHostPort + "/" + DEBUGToken)
	if err != nil {
		t.Errorf("redirect request error: %v", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusNotFound {
		t.Errorf("wrong status : %d", resp.StatusCode)
	}
}

// try unsupported request in mode = 2
func Test73MainServiceMode2(t *testing.T) {
	CONFIG.Mode = 2

	resp, err := http.Post("http://"+CONFIG.ListenHostPort+"/token", "application/json",
		strings.NewReader(`{"url": "http://someother.url"}`))
	if err != nil {
		t.Errorf("token request error: %v", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusNotFound {
		t.Errorf("wrong status : %d", resp.StatusCode)
	}
}

// try health check in service mode 1
func Test75MainHealthCheckMode1(t *testing.T) {
	CONFIG.Mode = 1

	resp, err := http.Get("http://" + CONFIG.ListenHostPort)
	if err != nil {
		t.Errorf("health check mode1 request error: %v", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		t.Errorf("wrong status : %d", resp.StatusCode)
	}
}

// try health check in service mode 1
func Test77MainHealthCheckMode2(t *testing.T) {
	CONFIG.Mode = 2

	resp, err := http.Get("http://" + CONFIG.ListenHostPort)
	if err != nil {
		t.Errorf("health check mode1 request error: %v", err)
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
	log.Printf("%s", buf[20:])
}
