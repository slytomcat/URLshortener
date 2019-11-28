package main

import (
	"bytes"
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"strings"
	"testing"
	"time"
)

func Test50mainStart(t *testing.T) {
	go main()
	time.Sleep(time.Second * 3)
}

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

func Test57mainHome(t *testing.T) {
	resp, err := http.Get("http://" + CONFIG.ListenHostPort)
	if err != nil {
		t.Errorf("helth check request error: %v", err)
	}
	defer resp.Body.Close()
	buf := make([]byte, 256)
	n, err := resp.Body.Read(buf)
	if err != nil && !errors.Is(err, io.EOF) {
		t.Errorf("response body reading error: %v", err)
	}

	if !bytes.Contains(buf[:n], []byte("Home page of URLshortener")) {
		t.Error("wrong response on helth check request")
	}

	if resp.StatusCode != http.StatusOK {
		t.Errorf("wrong status : %d", resp.StatusCode)
	}

}

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

func Test99mainKill(t *testing.T) {
	shutDown <- true
	time.Sleep(time.Second * 3)
}
