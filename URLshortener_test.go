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
func Test55mainGetToken(t *testing.T) {
	resp, err := http.Post("http://"+CONFIG.ListenHostPort+"/token", "application/json",
		strings.NewReader(`{"url": "http://`+CONFIG.ShortDomain+`", "exp": "3"}`))
	if err != nil {
		t.Errorf("token request error: %v", err)
	}
	defer resp.Body.Close()
	buf := make([]byte, 256)
	n, err := resp.Body.Read(buf)
	if err != nil && !errors.Is(err, io.EOF) {
		t.Errorf("response body reading error: %v", err)
	}
	var rep struct {
		URL   string `json:"url"`
		Token string `json:"token"`
	}
	err = json.Unmarshal(buf[:n], &rep)
	if err != nil {
		t.Errorf("response body parsing error: %v", err)
	}
	t.Logf("Received: %v", rep.URL)

	resp2, err := http.Get("http://" + rep.URL)
	if err != nil {
		t.Errorf("redirect request error: %v", err)
	}
	defer resp2.Body.Close()
	n, err = resp2.Body.Read(buf)
	if err != nil && !errors.Is(err, io.EOF) {
		t.Errorf("response body reading error: %v", err)
	}

	if !bytes.Contains(buf[:n], []byte("Home page of URLshortener")) {
		t.Error("wrong response on helth check request")
	}

	if resp2.StatusCode != http.StatusOK {
		t.Errorf("wrong status : %d", resp.StatusCode)
	}

}

// test health check
func Test57mainHome(t *testing.T) {
	resp, err := http.Get("http://" + CONFIG.ListenHostPort)
	if err != nil {
		t.Errorf("health check request error: %v", err)
	}
	defer resp.Body.Close()
	buf := make([]byte, 256)
	n, err := resp.Body.Read(buf)
	if err != nil && !errors.Is(err, io.EOF) {
		t.Errorf("response body reading error: %v", err)
	}

	if !bytes.Contains(buf[:n], []byte("Home page of URLshortener")) {
		t.Error("wrong response on health check request")
	}

	if resp.StatusCode != http.StatusOK {
		t.Errorf("wrong status : %d", resp.StatusCode)
	}

}

// test request for short URL with empty request body
func Test60mainBadRequest(t *testing.T) {
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
func Test61mainBadRequest2(t *testing.T) {
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
func Test62mainGetTokenWOexp(t *testing.T) {
	DEBUG = true
	defer func() { DEBUG = false }()

	tx, _ := tokenDB.DB.Begin()
	_, err := tx.Exec("DELETE FROM urls WHERE token='______'")
	if err != nil {
		t.Errorf("Can't clear table: %v", err)
	}
	tx.Commit()

	resp, err := http.Post("http://"+CONFIG.ListenHostPort+"/token", "application/json",
		strings.NewReader(`{"url": "http://someother.url"}`))
	if err != nil {
		t.Errorf("token request error: %v", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		t.Errorf("wrong status : %d", resp.StatusCode)
	}
}

// try to get the same (debugging) token twice
func Test62mainGetTokenTwice(t *testing.T) {
	DEBUG = true
	defer func() { DEBUG = false }()
	resp, err := http.Post("http://"+CONFIG.ListenHostPort+"/token", "application/json",
		strings.NewReader(`{"url": "http://someother.url"}`))
	if err != nil {
		t.Errorf("token request error: %v", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusGatewayTimeout {
		t.Errorf("wrong status : %d", resp.StatusCode)
	}
}

// test redirect with wrong token
func Test70main404(t *testing.T) {
	resp, err := http.Get("http://" + CONFIG.ListenHostPort + "/not_existing_token")
	if err != nil {
		t.Errorf("not-existing token request error: %v", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusNotFound {
		t.Errorf("wrong status : %d", resp.StatusCode)
	}

}

// try to stop service
func Test99mainKill(t *testing.T) {
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
