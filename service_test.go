package main

import (
	"io"
	"log"
	"net/http"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

var (
	servTestConfig  *Config
	servTestDB      TokenDB
	servTestHandler ServiceHandler
)

// try to start service with not working db
func Test10Serv03Start(t *testing.T) {
	conf := Config{
		ListenHostPort: "localhost:8080",
		ShortDomain:    "localhost:8080",
		Timeout:        500,
		TokenLength:    6,
	}

	errDb := newMockDB()

	testHandler := NewHandler(&conf, errDb, NewShortToken(5))

	go func() {
		assert.Error(t, testHandler.start())
	}()

	time.Sleep(time.Second)

	resp, err := http.Get("http://localhost:8080/")
	assert.NoError(t, err)
	defer resp.Body.Close()

	assert.Equal(t, http.StatusInternalServerError, resp.StatusCode)

	conf.Mode = disableShortener

	resp, err = http.Get("http://localhost:8080/")
	assert.NoError(t, err)
	defer resp.Body.Close()

	assert.Equal(t, http.StatusInternalServerError, resp.StatusCode)

	go testHandler.stop()

}

// try to start service
func Test10Serv05Start(t *testing.T) {
	var err error
	logger := log.Writer()
	r, w, _ := os.Pipe()
	log.SetOutput(w)

	servTestConfig, err = readConfig("./cnfr.json")
	assert.NoError(t, err)

	// initialize database connection
	servTestDB, err = NewTokenDB(servTestConfig.ConnectOptions)
	assert.NoError(t, err)

	// create short token interface
	sToken := NewShortToken(servTestConfig.TokenLength)

	// create service handler
	servTestHandler = NewHandler(servTestConfig, servTestDB, sToken)
	assert.NoError(t, err)

	go func() {
		log.Println(servTestHandler.start())
	}()

	w.Close()
	log.SetOutput(logger)
	buf, err := io.ReadAll(r)
	assert.NoError(t, err)

	assert.Contains(t, "starting server at", string(buf))
}

// test health check
func Test10Serv10Home(t *testing.T) {
	resp, err := http.Get("http://" + servTestConfig.ListenHostPort)
	assert.NoError(t, err)
	defer resp.Body.Close()

	assert.Equal(t, http.StatusOK, resp.StatusCode)

	buf, err := io.ReadAll(resp.Body)
	assert.NoError(t, err)

	assert.Contains(t, string(buf), "Home page of URLshortener")
}

// test bad method
func Test10Serv13BadMethod(t *testing.T) {
	resp, err := http.Head("http://" + servTestConfig.ListenHostPort)
	assert.NoError(t, err)
	defer resp.Body.Close()
	assert.Equal(t, http.StatusBadRequest, resp.StatusCode)
}

// test request for short URL with empty request body
func Test10Serv15BadTokenRequest(t *testing.T) {
	resp, err := http.Post("http://"+servTestConfig.ListenHostPort+"/api/v1/token", "application/json",
		strings.NewReader(``))
	assert.NoError(t, err)
	defer resp.Body.Close()

	assert.Equal(t, http.StatusBadRequest, resp.StatusCode)
}

// test request for short URL with empty JSON
func Test10Serv20BadTokenRequest2(t *testing.T) {
	resp, err := http.Post("http://"+servTestConfig.ListenHostPort+"/api/v1/token", "application/json",
		strings.NewReader(`{}`))
	assert.NoError(t, err)
	defer resp.Body.Close()

	assert.Equal(t, http.StatusBadRequest, resp.StatusCode)
}

//test request for short URL without expiration in request
func Test10Serv25GetTokenWOexp(t *testing.T) {

	resp, err := http.Post("http://"+servTestConfig.ListenHostPort+"/api/v1/token", "application/json",
		strings.NewReader(`{"url": "http://`+servTestConfig.ShortDomain+`"}`))
	assert.NoError(t, err)
	defer resp.Body.Close()
	assert.Equal(t, http.StatusOK, resp.StatusCode)
}

// request expire without parameters
func Test10Serv30ExpireTokenWObody(t *testing.T) {

	resp, err := http.Post("http://"+servTestConfig.ListenHostPort+"/api/v1/expire", "application/json",
		strings.NewReader(``))
	assert.NoError(t, err)
	defer resp.Body.Close()

	assert.Equal(t, http.StatusBadRequest, resp.StatusCode)
}

// request expire with empty JSON
func Test10Serv35ExpireTokenWOparams(t *testing.T) {

	resp, err := http.Post("http://"+servTestConfig.ListenHostPort+"/api/v1/expire", "application/json",
		strings.NewReader(`{}`))
	assert.NoError(t, err)
	defer resp.Body.Close()

	assert.Equal(t, http.StatusBadRequest, resp.StatusCode)
}

// request expire for not existing token
func Test10Serv40ExpireNotExistingToken(t *testing.T) {

	resp, err := http.Post("http://"+servTestConfig.ListenHostPort+"/api/v1/expire", "application/json",
		strings.NewReader(`{"token":"`+strings.Repeat("(", servTestConfig.TokenLength)+`"}`)) // use non Base64 symbols
	assert.NoError(t, err)
	defer resp.Body.Close()

	assert.Equal(t, http.StatusNotModified, resp.StatusCode)
}

// test redirect with wrong token (wrong lenght and wrong symbols)
func Test10Serv45RedirectTo404(t *testing.T) {
	resp, err := http.Get("http://" + servTestConfig.ListenHostPort + "/not+existing+token")
	assert.NoError(t, err)
	defer resp.Body.Close()

	assert.Equal(t, http.StatusNotFound, resp.StatusCode)
}

// test redirect with wrong token (correct lenght and wrong symbols)
func Test10Serv50RedirectTo404(t *testing.T) {
	resp, err := http.Get("http://" + servTestConfig.ListenHostPort + "/" + strings.Repeat("(", servTestConfig.TokenLength))
	assert.NoError(t, err)
	defer resp.Body.Close()

	assert.Equal(t, http.StatusNotFound, resp.StatusCode)
}

// test redirect with wrong token (correct lenght and correct symbols)
func Test10Serv53RedirectTo404(t *testing.T) {
	resp, err := http.Get("http://" + servTestConfig.ListenHostPort + "/" + strings.Repeat("A", servTestConfig.TokenLength))
	assert.NoError(t, err)
	defer resp.Body.Close()

	assert.Equal(t, http.StatusNotFound, resp.StatusCode)
}

// try unsupported request in mode = disableRedirect
func Test10Serv55ServiceModeDisableRedirect(t *testing.T) {

	servTestConfig.Mode = disableRedirect

	resp, err := http.Get("http://" + servTestConfig.ListenHostPort + "/" + strings.Repeat("_", servTestConfig.TokenLength))
	assert.NoError(t, err)
	defer resp.Body.Close()

	assert.Equal(t, http.StatusNotFound, resp.StatusCode)
}

// try unsupported request in mode = disableShortener
func Test10Serv60ServiceModeDisableShortener(t *testing.T) {

	servTestConfig.Mode = disableShortener

	resp, err := http.Post("http://"+servTestConfig.ListenHostPort+"/api/v1/token", "application/json",
		strings.NewReader(`{"url": "http://someother.url"}`))
	assert.NoError(t, err)
	defer resp.Body.Close()

	assert.Equal(t, http.StatusNotFound, resp.StatusCode)
}

// try unsupported request in mode = disableExpire
func Test10Serv65ServiceModeDisableExpire(t *testing.T) {

	servTestConfig.Mode = disableExpire

	resp, err := http.Post("http://"+servTestConfig.ListenHostPort+"/api/v1/expire", "application/json",
		strings.NewReader(`{"token": "`+strings.Repeat("_", servTestConfig.TokenLength)+`","exp":-1}`))
	assert.NoError(t, err)
	defer resp.Body.Close()

	assert.Equal(t, http.StatusNotFound, resp.StatusCode)
}

// try health check in service mode disableRedirect
func Test10Serv70HealthCheckModeDisableRedirect(t *testing.T) {

	servTestConfig.Mode = disableRedirect

	resp, err := http.Get("http://" + servTestConfig.ListenHostPort)
	assert.NoError(t, err)
	defer resp.Body.Close()

	assert.Equal(t, http.StatusOK, resp.StatusCode)
	time.Sleep(time.Second)
}

// try health check in service mode disableShortener
func Test10Serv75HealthCheckModeDisableShortener(t *testing.T) {

	servTestConfig.Mode = disableShortener

	resp, err := http.Get("http://" + servTestConfig.ListenHostPort)
	assert.NoError(t, err)
	defer resp.Body.Close()

	assert.Equal(t, http.StatusOK, resp.StatusCode)
	time.Sleep(time.Second)
}

// try health check in service mode disableExpire
func Test10Serv77HealthCheckModeDisableExpire(t *testing.T) {

	servTestConfig.Mode = disableExpire

	resp, err := http.Get("http://" + servTestConfig.ListenHostPort)
	assert.NoError(t, err)
	defer resp.Body.Close()

	assert.Equal(t, http.StatusOK, resp.StatusCode)
	time.Sleep(time.Second)
}

// test generate UI interface
func Test10Serv80genUI(t *testing.T) {
	resp, err := http.Get("http://" + servTestConfig.ListenHostPort + "/ui/generate")
	assert.NoError(t, err)
	defer resp.Body.Close()

	assert.Equal(t, http.StatusOK, resp.StatusCode)

	buf, err := io.ReadAll(resp.Body)
	assert.NoError(t, err)

	assert.Contains(t, string(buf), "URL to be shortened")
}

// test generate UI interface
func Test10Serv83genUIwp(t *testing.T) {
	resp, err := http.Get("http://" + servTestConfig.ListenHostPort + "/ui/generate?s=http:/some.url")
	assert.NoError(t, err)
	defer resp.Body.Close()

	assert.Equal(t, http.StatusOK, resp.StatusCode)

	buf, err := io.ReadAll(resp.Body)
	assert.NoError(t, err)
	assert.Contains(t, string(buf), "Short URL:")
}

func Test10Serv85genUIdisabled(t *testing.T) {
	servTestConfig.Mode = disableUI
	resp, err := http.Get("http://" + servTestConfig.ListenHostPort + "/ui/generate?s=http:/some.url")
	assert.NoError(t, err)
	defer resp.Body.Close()

	assert.Equal(t, http.StatusNotFound, resp.StatusCode)
}

// try to stop service
func Test10Serv89InteruptService(t *testing.T) {
	// logger := log.Writer()
	// r, w, _ := os.Pipe()
	// log.SetOutput(w)

	servTestHandler.stop()

	// w.Close()
	// log.SetOutput(logger)
	// buf, err := io.ReadAll(r)
	// assert.NoError(t, err)

	//assert.Contains(t, string(buf), "http: Server closed")
}

// try tokens' duplicate
func Test10Serv90Duble(t *testing.T) {

	servTestConfig, err := readConfig("./cnfr.json")
	assert.NoError(t, err)

	servTestDB, err := NewTokenDB(servTestConfig.ConnectOptions)
	assert.NoError(t, err)

	// create short token interface
	sToken := NewShortTokenD(servTestConfig.TokenLength)

	// create service handler
	servTestHandler = NewHandler(servTestConfig, servTestDB, sToken)

	token := sToken.Get()
	servTestDB.Delete(token)

	go func() {
		log.Println(servTestHandler.start())
	}()

	time.Sleep(time.Second * 2)

	resp, err := http.Post("http://"+servTestConfig.ListenHostPort+"/api/v1/token", "application/json",
		strings.NewReader(`{"url": "http://`+servTestConfig.ShortDomain+`"}`))
	assert.NoError(t, err)
	resp.Body.Close()

	assert.Equal(t, http.StatusOK, resp.StatusCode)
	// second request

	resp2, err := http.Post("http://"+servTestConfig.ListenHostPort+"/api/v1/token", "application/json",
		strings.NewReader(`{"url": "http://`+servTestConfig.ShortDomain+`"}`))
	assert.NoError(t, err)
	resp.Body.Close()
	assert.Equal(t, http.StatusRequestTimeout, resp2.StatusCode)

	token = sToken.Get()
	servTestDB.Delete(token)

	servTestHandler.stop()
}

func Test10Serv92BadDB(t *testing.T) {

	servTestConfig, err := readConfig("./cnfr.json")
	assert.NoError(t, err)

	servTestDB := newMockDB()

	// create short token interface
	sToken := NewShortToken(5)

	// create service handler
	servTestHandler = NewHandler(servTestConfig, servTestDB, sToken)

	go func() {
		log.Println(servTestHandler.start())
	}()

	time.Sleep(time.Second * 2)

	resp, err := http.Post("http://"+servTestConfig.ListenHostPort+"/api/v1/token", "application/json",
		strings.NewReader(`{"url": "http://`+servTestConfig.ShortDomain+`"}`))
	assert.NoError(t, err)
	resp.Body.Close()

	assert.Equal(t, http.StatusInternalServerError, resp.StatusCode)
	servTestHandler.stop()
}
