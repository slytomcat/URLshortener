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
	"testing"
	"time"

	"github.com/go-redis/redis/v7"
)

var (
	servTestConfig  *Config
	servTestDB      TokenDB
	servTestHandler ServiceHandler
	servTestexit    chan bool = make(chan bool)
)

// try to start service with not working db
func Test10Serv03Start(t *testing.T) {
	logger := log.Writer()
	r, w, _ := os.Pipe()
	log.SetOutput(w)

	conf := Config{
		ListenHostPort: "localhost:8080",
		ShortDomain:    "localhost:8080",
	}

	errDb, _ := testDBNewTokenDB(redis.UniversalOptions{})
	testHandler := NewHandler(&conf, errDb, NewShortToken(5), servTestexit)

	go func() {
		log.Println(testHandler.Start())
	}()

	select {
	case <-servTestexit:
		t.Error("servise starting error")
	case <-time.After(time.Second * 3):
		break
	}

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

	resp, err := http.Get("http://localhost:8080/")
	if err != nil {
		t.Errorf("health-check request error: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusInternalServerError {
		t.Errorf("unexpected response status: %d", resp.StatusCode)
	}

	conf.Mode = disableShortener

	resp, err = http.Get("http://localhost:8080/")
	if err != nil {
		t.Errorf("health-check request error: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusInternalServerError {
		t.Errorf("unexpected response status: %d", resp.StatusCode)
	}

	go testHandler.Stop()

	<-servTestexit

}

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
	servTestDB, err = NewTokenDB(servTestConfig.ConnectOptions)
	if err != nil {
		t.Fatalf("error database interface creation: %v", err)
	}

	// create short token interface
	sToken := NewShortToken(servTestConfig.TokenLength)

	// create service handler
	servTestHandler = NewHandler(servTestConfig, servTestDB, sToken, servTestexit)
	if err != nil {
		t.Errorf("servece creation error: %v", err)
	}

	go func() {
		log.Println(servTestHandler.Start())
	}()

	select {
	case <-servTestexit:
		t.Error("servise starting error")
	case <-time.After(time.Second * 3):
		break
	}

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
func Test10Serv10Home(t *testing.T) {
	resp, err := http.Get("http://" + servTestConfig.ListenHostPort)
	if err != nil {
		t.Errorf("health check request error: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		t.Errorf("unexpected response status: %d", resp.StatusCode)
	}

	buf, err := ioutil.ReadAll(resp.Body)
	if err != nil && !errors.Is(err, io.EOF) {
		t.Errorf("response body reading error: %v", err)
	}

	if !bytes.Contains(buf, []byte("Home page of URLshortener")) {
		t.Errorf("wrong response on health check request: %s", buf)
	}

	if bytes.Contains(buf, []byte("Service status: healthy, 0 attempts")) {
		t.Errorf("zero attempts while success healtcheck: %s", buf)
	}
}

// test request for short URL with empty request body
func Test10Serv15BadTokenRequest(t *testing.T) {
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
func Test10Serv20BadTokenRequest2(t *testing.T) {
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
func Test10Serv25GetTokenWOexp(t *testing.T) {

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
func Test10Serv30ExpireTokenWObody(t *testing.T) {

	resp, err := http.Post("http://"+servTestConfig.ListenHostPort+"/api/v1/expire", "application/json",
		strings.NewReader(``))
	if err != nil {
		t.Errorf("expire request error: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusBadRequest {
		t.Errorf("unexpected response status: %d", resp.StatusCode)
	}

}

// request expire without parameters
func Test10Serv35ExpireTokenWOparams(t *testing.T) {

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
func Test10Serv40ExpireNotExistingToken(t *testing.T) {

	resp, err := http.Post("http://"+servTestConfig.ListenHostPort+"/api/v1/expire", "application/json",
		strings.NewReader(`{"token":"`+strings.Repeat("(", servTestConfig.TokenLength)+`"}`)) // use non Base64 symbols
	if err != nil {
		t.Errorf("expire request error: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusNotModified {
		t.Errorf("unexpected response status: %d", resp.StatusCode)
	}

}

// test redirect with wrong token
func Test10Serv45RedirectTo404(t *testing.T) {
	resp, err := http.Get("http://" + servTestConfig.ListenHostPort + "/not_existing_token")
	if err != nil {
		t.Errorf("not-existing token request error: %v", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusNotFound {
		t.Errorf("unexpected response status: %d", resp.StatusCode)
	}
}

// test redirect with wrong token
func Test10Serv50RedirectTo404_(t *testing.T) {
	resp, err := http.Get("http://" + servTestConfig.ListenHostPort + "/" + strings.Repeat("(", servTestConfig.TokenLength))
	if err != nil {
		t.Errorf("not-existing token request error: %v", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusNotFound {
		t.Errorf("unexpected response status: %d", resp.StatusCode)
	}
}

// try unsupported request in mode = disableRedirect
func Test10Serv55ServiceModeDisableRedirect(t *testing.T) {

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
func Test10Serv60ServiceModeDisableShortener(t *testing.T) {

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
func Test10Serv65ServiceModeDisableExpire(t *testing.T) {

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
func Test10Serv70HealthCheckModeDisableRedirect(t *testing.T) {

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
func Test10Serv75HealthCheckModeDisableShortener(t *testing.T) {

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
func Test10Serv80HealthCheckModeDisableExpire(t *testing.T) {

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
func Test10Serv85InteruptService(t *testing.T) {
	logger := log.Writer()
	r, w, _ := os.Pipe()
	log.SetOutput(w)

	go servTestHandler.Stop()

	select {
	case <-time.After(time.Second * 2):
		t.Error("no exit reported")
	case <-servTestexit:
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

// try tokens' duplicate
func Test10Serv90Duble(t *testing.T) {

	servTestConfig, err := readConfig("./cnfr.json")
	if err != nil {
		t.Fatalf("configuration read error: %v", err)
	}

	servTestDB, err := NewTokenDB(servTestConfig.ConnectOptions)
	if err != nil {
		t.Fatalf("error database interface creation: %v", err)
	}

	// create short token interface
	sToken := NewShortTokenD(servTestConfig.TokenLength)

	// make exit chanel
	servTestexit = make(chan bool)

	// create service handler
	servTestHandler = NewHandler(servTestConfig, servTestDB, sToken, servTestexit)

	token, _ := sToken.Get()
	servTestDB.Delete(token)

	go func() {
		log.Println(servTestHandler.Start())
	}()

	time.Sleep(time.Second * 2)

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

	token, _ = sToken.Get()
	servTestDB.Delete(token)

	go servTestHandler.Stop()

	<-servTestexit
}

func Test10Serv91BadToken(t *testing.T) {

	servTestConfig, err := readConfig("./cnfr.json")
	if err != nil {
		t.Fatalf("configuration read error: %v", err)
	}

	servTestDB, err := NewTokenDB(servTestConfig.ConnectOptions)
	if err != nil {
		t.Fatalf("error database interface creation: %v", err)
	}

	// create short token interface
	sToken := NewShortTokenE(servTestConfig.TokenLength)

	// make exit chanel
	servTestexit = make(chan bool)

	// create service handler
	servTestHandler = NewHandler(servTestConfig, servTestDB, sToken, servTestexit)
	if err != nil {
		t.Errorf("servece creation error: %v", err)
	}

	go func() {
		log.Println(servTestHandler.Start())
	}()

	time.Sleep(time.Second * 2)

	resp, err := http.Post("http://"+servTestConfig.ListenHostPort+"/api/v1/token", "application/json",
		strings.NewReader(`{"url": "http://`+servTestConfig.ShortDomain+`"}`))
	if err != nil {
		t.Errorf("token request error: %v", err)
	}

	if resp.StatusCode != http.StatusInternalServerError {
		t.Errorf("unexpected response status: %d", resp.StatusCode)
	}
	resp.Body.Close()

	go servTestHandler.Stop()

	<-servTestexit
}

func Test10Serv92BadDB(t *testing.T) {

	servTestConfig, err := readConfig("./cnfr.json")
	if err != nil {
		t.Fatalf("configuration read error: %v", err)
	}

	servTestDB, err := testDBNewTokenDB(redis.UniversalOptions{})
	if err != nil {
		t.Fatalf("error database interface creation: %v", err)
	}

	// create short token interface
	sToken := NewShortToken(5)

	// make exit chanel
	servTestexit = make(chan bool)

	// create service handler
	servTestHandler = NewHandler(servTestConfig, servTestDB, sToken, servTestexit)

	go func() {
		log.Println(servTestHandler.Start())
	}()

	time.Sleep(time.Second * 2)

	resp, err := http.Post("http://"+servTestConfig.ListenHostPort+"/api/v1/token", "application/json",
		strings.NewReader(`{"url": "http://`+servTestConfig.ShortDomain+`"}`))
	if err != nil {
		t.Errorf("token request error: %v", err)
	}

	if resp.StatusCode != http.StatusInternalServerError {
		t.Errorf("unexpected response status: %d", resp.StatusCode)
	}
	resp.Body.Close()

	go servTestHandler.Stop()

	<-servTestexit
}
