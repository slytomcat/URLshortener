package main

import (
	"bytes"
	"errors"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"strings"
	"syscall"
	"testing"
	"time"
)

var (
	servTestConfig *Config
	servTestDB     Token
	exit           chan bool
)

// try to start service
func Test10Serv05Start(t *testing.T) {
	var err error
	logger := log.Writer()
	r, w, _ := os.Pipe()
	log.SetOutput(w)

	servTestConfig, err = readConfig("./cnfr.json")
	if err != nil {
		t.Fatalf("configuration read error: %v", err)
	}

	// initialize database connection
	servTestDB, err = NewTokenDB(servTestConfig.ConnectOptions, servTestConfig.Timeout, servTestConfig.TokenLength)
	if err != nil {
		t.Fatalf("error database interface creation: %v", err)
	}
	exit = make(chan bool)
	// run service
	go func() {
		log.Println(ServiceStart(servTestConfig, exit))
	}()

	time.Sleep(time.Second * 3)

	w.Close()
	log.SetOutput(logger)
	buf, err := ioutil.ReadAll(r)
	if err != nil {
		t.Error(err)
	}
	if !bytes.Contains(buf, []byte("starting server at")) {
		t.Errorf("received unexpected output: %s", buf)
	}
	log.Printf("%s", buf)

}

// test health check
func Test10Serv15Home(t *testing.T) {
	resp, err := http.Get("http://" + servTestConfig.ListenHostPort)
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
		t.Errorf("wrong response on health check request: %v", buf)
	}

	if resp.StatusCode != http.StatusOK {
		t.Errorf("unexpected response status: %d", resp.StatusCode)
	}

}

// test request for short URL with empty request body
func Test10Serv20BadTokenRequest(t *testing.T) {
	resp, err := http.Post("http://"+servTestConfig.ListenHostPort+"/api/v1/token", "application/json",
		strings.NewReader(``))
	if err != nil {
		t.Errorf("token request error: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusBadRequest {
		t.Errorf("unexpected response status: %d", resp.StatusCode)
	}
}

// test request for short URL with empty JSON
func Test10Serv30BadTokenRequest2(t *testing.T) {
	resp, err := http.Post("http://"+servTestConfig.ListenHostPort+"/api/v1/token", "application/json",
		strings.NewReader(`{}`))
	if err != nil {
		t.Errorf("token request error: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusBadRequest {
		t.Errorf("unexpected response status: %d", resp.StatusCode)
	}
}

//test request for short URL without expiration in request
func Test10Serv35GetTokenWOexp(t *testing.T) {

	// clear debug token
	servTestDB.Delete(strings.Repeat("_", servTestConfig.TokenLength))

	resp, err := http.Post("http://"+servTestConfig.ListenHostPort+"/api/v1/token", "application/json",
		strings.NewReader(`{"url": "http://`+servTestConfig.ShortDomain+`"}`))
	if err != nil {
		t.Errorf("token request error: %v", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		t.Errorf("unexpected response status: %d", resp.StatusCode)
	}
}

// request expire without parameters
func Test10Serv40ExpireTokenWOparams(t *testing.T) {

	resp, err := http.Post("http://"+servTestConfig.ListenHostPort+"/api/v1/expire", "application/json",
		strings.NewReader(`{}`))
	if err != nil {
		t.Errorf("expire request error: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusBadRequest {
		t.Errorf("unexpected response status: %d", resp.StatusCode)
	}

}

// request expire for not existing token
func Test10Serv45ExpireNotExistingToken(t *testing.T) {

	resp, err := http.Post("http://"+servTestConfig.ListenHostPort+"/api/v1/expire", "application/json",
		strings.NewReader(`{"token":"$%#@*"}`)) // use non Base64 symbols
	if err != nil {
		t.Errorf("expire request error: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusNotModified {
		t.Errorf("unexpected response status: %d", resp.StatusCode)
	}

}

// try to get the same (debugging) token twice
func Test10Serv50GetTokenTwice(t *testing.T) {

	defer SetDebug(1)()
	// first request
	resp, err := http.Post("http://"+servTestConfig.ListenHostPort+"/api/v1/token", "application/json",
		strings.NewReader(`{"url": "http://`+servTestConfig.ShortDomain+`"}`))
	if err != nil {
		t.Errorf("token request error: %v", err)
	}
	if resp.StatusCode != http.StatusOK {
		t.Errorf("unexpected response status: %d", resp.StatusCode)
	}
	resp.Body.Close()
	// second request

	resp2, err := http.Post("http://"+servTestConfig.ListenHostPort+"/api/v1/token", "application/json",
		strings.NewReader(`{"url": "http://`+servTestConfig.ShortDomain+`"}`))
	if err != nil {
		t.Errorf("token request error: %v", err)
	}
	if resp2.StatusCode != http.StatusRequestTimeout {
		t.Errorf("unexpected response status: %d", resp.StatusCode)
	}
	resp.Body.Close()

}

// test redirect with wrong token
func Test10Serv60RedirectTo404(t *testing.T) {
	resp, err := http.Get("http://" + servTestConfig.ListenHostPort + "/not_existing_token")
	if err != nil {
		t.Errorf("not-existing token request error: %v", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusNotFound {
		t.Errorf("unexpected response status: %d", resp.StatusCode)
	}
}

// try unsupported request in mode = disableRedirect
func Test10Serv65ServiceModeDisableRedirect(t *testing.T) {

	servTestConfig.Mode = disableRedirect

	resp, err := http.Get("http://" + servTestConfig.ListenHostPort + "/" + strings.Repeat("_", servTestConfig.TokenLength))
	if err != nil {
		t.Errorf("redirect request error: %v", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusNotFound {
		t.Errorf("unexpected response status: %d", resp.StatusCode)
	}
}

// try unsupported request in mode = disableShortener
func Test10Serv70ServiceModeDisableShortener(t *testing.T) {

	servTestConfig.Mode = disableShortener

	resp, err := http.Post("http://"+servTestConfig.ListenHostPort+"/api/v1/token", "application/json",
		strings.NewReader(`{"url": "http://someother.url"}`))
	if err != nil {
		t.Errorf("token request error: %v", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusNotFound {
		t.Errorf("unexpected response status: %d", resp.StatusCode)
	}
}

// try unsupported request in mode = disableExpire
func Test10Serv75ServiceModeDisableExpire(t *testing.T) {

	servTestConfig.Mode = disableExpire

	resp, err := http.Post("http://"+servTestConfig.ListenHostPort+"/api/v1/expire", "application/json",
		strings.NewReader(`{"token": "`+strings.Repeat("_", servTestConfig.TokenLength)+`","exp":-1}`))
	if err != nil {
		t.Errorf("expire request error: %v", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusNotFound {
		t.Errorf("wrong status: %d", resp.StatusCode)
	}
}

// try health check in service mode disableRedirect
func Test10Serv80HealthCheckModeDisableRedirect(t *testing.T) {

	servTestConfig.Mode = disableRedirect

	resp, err := http.Get("http://" + servTestConfig.ListenHostPort)
	if err != nil {
		t.Errorf("health check in disableRedirect mode request error: %v", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		t.Errorf("unexpected response status: %d", resp.StatusCode)
	}
	time.Sleep(time.Second)
}

// try health check in service mode disableShortener
func Test10Serv85HealthCheckModeDisableShortener(t *testing.T) {

	servTestConfig.Mode = disableShortener

	resp, err := http.Get("http://" + servTestConfig.ListenHostPort)
	if err != nil {
		t.Errorf("health check in disableShortener mode request error: %v", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		t.Errorf("unexpected response status: %d", resp.StatusCode)
	}
	time.Sleep(time.Second)
}

// try health check in service mode disableExpire
func Test10Serv90HealthCheckModeDisableExpire(t *testing.T) {

	servTestConfig.Mode = disableExpire

	resp, err := http.Get("http://" + servTestConfig.ListenHostPort)
	if err != nil {
		t.Errorf("health check in disableExpire mode request error: %v", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		t.Errorf("unexpected response status: %d", resp.StatusCode)
	}
	time.Sleep(time.Second)
}

// try to stop service
func Test10Serv95InteruptService(t *testing.T) {
	logger := log.Writer()
	r, w, _ := os.Pipe()
	log.SetOutput(w)

	syscall.Kill(syscall.Getpid(), syscall.SIGINT)

	select {
	case <-time.After(time.Second * 2):
		t.Error("no exit reported")
	case <-exit:
		t.Log("exit reported")
	}

	w.Close()
	log.SetOutput(logger)
	buf, err := ioutil.ReadAll(r)
	if err != nil {
		t.Error(err)
	}
	if !bytes.Contains(buf, []byte("http: Server closed")) {
		t.Errorf("received unexpected output: %s", buf)
	}
	log.Printf("%s", buf)
}
