package main

import (
	"errors"
	"io"
	"log"
	"net/http"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/joho/godotenv"
	"github.com/stretchr/testify/assert"
)

var (
	servTestConfig  *Config
	servTestDB      TokenDB
	servTestHandler ServiceHandler
)

// try to start service with not working db
func Test10Serv03CheckHealthCheck(t *testing.T) {
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
	assert.Equal(t, http.StatusOK, resp.StatusCode)

	conf.Mode = disableExpire

	resp, err = http.Get("http://localhost:8080/")
	assert.NoError(t, err)
	defer resp.Body.Close()
	assert.Equal(t, http.StatusOK, resp.StatusCode)

	conf.Mode = 0
	errDb.expFunc = func(string, int) error { return errors.New("some error") }

	resp, err = http.Get("http://localhost:8080/")
	assert.NoError(t, err)
	defer resp.Body.Close()
	assert.Equal(t, http.StatusInternalServerError, resp.StatusCode)

	conf.Mode = disableExpire

	resp, err = http.Get("http://localhost:8080/")
	assert.NoError(t, err)
	defer resp.Body.Close()
	assert.Equal(t, http.StatusInternalServerError, resp.StatusCode)

	conf.Mode = disableShortener | disableRedirect
	errDb.getFunc = func(s string) (string, error) { return "wrongURL", nil }

	resp, err = http.Get("http://localhost:8080/")
	assert.NoError(t, err)
	defer resp.Body.Close()
	assert.Equal(t, http.StatusInternalServerError, resp.StatusCode)

	conf.Mode = 0
	errDb.getFunc = func(string) (string, error) { return "", errors.New("some error") }

	resp, err = http.Get("http://localhost:8080/")
	assert.NoError(t, err)
	defer resp.Body.Close()
	assert.Equal(t, http.StatusInternalServerError, resp.StatusCode)

	conf.Mode = disableRedirect

	resp, err = http.Get("http://localhost:8080/")
	assert.NoError(t, err)
	defer resp.Body.Close()
	assert.Equal(t, http.StatusInternalServerError, resp.StatusCode)

	conf.Mode = 0
	errDb.setFunc = func(string, string, int) (bool, error) { return false, nil }

	resp, err = http.Get("http://localhost:8080/")
	assert.NoError(t, err)
	defer resp.Body.Close()
	assert.Equal(t, http.StatusInternalServerError, resp.StatusCode)

	conf.Mode = disableShortener

	resp, err = http.Get("http://localhost:8080/")
	assert.NoError(t, err)
	defer resp.Body.Close()
	assert.Equal(t, http.StatusInternalServerError, resp.StatusCode)

	errDb.closeFunc = func() error { return errors.New("some error") }

	testHandler.stop()

}

// try to start service
func Test10Serv05All(t *testing.T) {

	godotenv.Load()

	servTestConfig, err := readConfig()
	assert.NoError(t, err)

	// initialize database connection
	servTestDB, err = NewTokenDB(servTestConfig.RedisAddrs, servTestConfig.RedisPassword)
	assert.NoError(t, err)

	// create service handler
	servTestHandler = NewHandler(servTestConfig, servTestDB, NewShortToken(servTestConfig.TokenLength))

	assert.NoError(t, err)
	logger := log.Writer()
	r, w, _ := os.Pipe()
	log.SetOutput(w)

	go func() {
		log.Println(servTestHandler.start())
	}()

	time.Sleep(3 * time.Second)
	w.Close()
	log.SetOutput(logger)
	buf, err := io.ReadAll(r)
	assert.NoError(t, err)
	assert.Contains(t, string(buf), "starting server at")

	t.Run("do health check", func(t *testing.T) {
		resp, err := http.Get("http://" + servTestConfig.ListenHostPort)
		assert.NoError(t, err)
		defer resp.Body.Close()

		assert.Equal(t, http.StatusOK, resp.StatusCode)
		buf, err := io.ReadAll(resp.Body)
		assert.NoError(t, err)
		assert.Contains(t, string(buf), "Home page of URLshortener")
	})

	t.Run("bad method", func(t *testing.T) {
		resp, err := http.Head("http://" + servTestConfig.ListenHostPort)
		assert.NoError(t, err)
		defer resp.Body.Close()
		assert.Equal(t, http.StatusBadRequest, resp.StatusCode)
	})

	t.Run("short URL query with empty request body", func(t *testing.T) {
		resp, err := http.Post("http://"+servTestConfig.ListenHostPort+"/api/v1/token", "application/json",
			strings.NewReader(``))
		assert.NoError(t, err)
		defer resp.Body.Close()

		assert.Equal(t, http.StatusBadRequest, resp.StatusCode)
	})

	t.Run("short URL request with empty JSON", func(t *testing.T) {
		resp, err := http.Post("http://"+servTestConfig.ListenHostPort+"/api/v1/token", "application/json",
			strings.NewReader(`{}`))
		assert.NoError(t, err)
		defer resp.Body.Close()

		assert.Equal(t, http.StatusBadRequest, resp.StatusCode)
	})

	t.Run("short URL request without expiration", func(t *testing.T) {
		resp, err := http.Post("http://"+servTestConfig.ListenHostPort+"/api/v1/token", "application/json",
			strings.NewReader(`{"url": "http://`+servTestConfig.ShortDomain+`"}`))
		assert.NoError(t, err)
		defer resp.Body.Close()
		assert.Equal(t, http.StatusOK, resp.StatusCode)
	})

	t.Run("expire request without parameters", func(t *testing.T) {
		resp, err := http.Post("http://"+servTestConfig.ListenHostPort+"/api/v1/expire", "application/json",
			strings.NewReader(``))
		assert.NoError(t, err)
		defer resp.Body.Close()

		assert.Equal(t, http.StatusBadRequest, resp.StatusCode)
	})

	t.Run("expire request with empty JSON", func(t *testing.T) {
		resp, err := http.Post("http://"+servTestConfig.ListenHostPort+"/api/v1/expire", "application/json",
			strings.NewReader(`{}`))
		assert.NoError(t, err)
		defer resp.Body.Close()

		assert.Equal(t, http.StatusBadRequest, resp.StatusCode)
	})

	t.Run("expire request for not existing token", func(t *testing.T) {
		resp, err := http.Post("http://"+servTestConfig.ListenHostPort+"/api/v1/expire", "application/json",
			strings.NewReader(`{"token":"`+strings.Repeat("(", servTestConfig.TokenLength)+`"}`)) // use non Base64 symbols
		assert.NoError(t, err)
		defer resp.Body.Close()

		assert.Equal(t, http.StatusNotModified, resp.StatusCode)
	})

	t.Run("redirect request with wrong token (wrong length)", func(t *testing.T) {
		resp, err := http.Get("http://" + servTestConfig.ListenHostPort + "/not+existing+token")
		assert.NoError(t, err)
		defer resp.Body.Close()

		assert.Equal(t, http.StatusNotFound, resp.StatusCode)
	})

	t.Run("redirect request with wrong token (wrong symbols)", func(t *testing.T) {
		resp, err := http.Get("http://" + servTestConfig.ListenHostPort + "/" + strings.Repeat("(", servTestConfig.TokenLength))
		assert.NoError(t, err)
		defer resp.Body.Close()

		assert.Equal(t, http.StatusNotFound, resp.StatusCode)
	})

	t.Run("redirect request with wrong token (correct lenght&symbols)", func(t *testing.T) {
		resp, err := http.Get("http://" + servTestConfig.ListenHostPort + "/" + strings.Repeat("A", servTestConfig.TokenLength))
		assert.NoError(t, err)
		defer resp.Body.Close()

		assert.Equal(t, http.StatusNotFound, resp.StatusCode)
	})

	t.Run("unsupported request in mode = disableRedirect", func(t *testing.T) {
		servTestConfig.Mode = disableRedirect

		resp, err := http.Get("http://" + servTestConfig.ListenHostPort + "/" + strings.Repeat("_", servTestConfig.TokenLength))
		assert.NoError(t, err)
		defer resp.Body.Close()

		assert.Equal(t, http.StatusNotFound, resp.StatusCode)
	})

	t.Run("unsupported request in mode = disableShortener", func(t *testing.T) {
		servTestConfig.Mode = disableShortener

		resp, err := http.Post("http://"+servTestConfig.ListenHostPort+"/api/v1/token", "application/json",
			strings.NewReader(`{"url": "http://someother.url"}`))
		assert.NoError(t, err)
		defer resp.Body.Close()

		assert.Equal(t, http.StatusNotFound, resp.StatusCode)
	})

	t.Run("unsupported request in mode = disableExpire", func(t *testing.T) {
		servTestConfig.Mode = disableExpire

		resp, err := http.Post("http://"+servTestConfig.ListenHostPort+"/api/v1/expire", "application/json",
			strings.NewReader(`{"token": "`+strings.Repeat("_", servTestConfig.TokenLength)+`","exp":-1}`))
		assert.NoError(t, err)
		defer resp.Body.Close()

		assert.Equal(t, http.StatusNotFound, resp.StatusCode)
	})

	t.Run("health check in service mode disableRedirect", func(t *testing.T) {
		servTestConfig.Mode = disableRedirect

		resp, err := http.Get("http://" + servTestConfig.ListenHostPort)
		assert.NoError(t, err)
		defer resp.Body.Close()

		assert.Equal(t, http.StatusOK, resp.StatusCode)
		time.Sleep(time.Second)
	})

	t.Run("health check in service mode disableShortener", func(t *testing.T) {
		servTestConfig.Mode = disableShortener

		resp, err := http.Get("http://" + servTestConfig.ListenHostPort)
		assert.NoError(t, err)
		defer resp.Body.Close()

		assert.Equal(t, http.StatusOK, resp.StatusCode)
		time.Sleep(time.Second)
	})

	t.Run("health check in service mode disableExpire", func(t *testing.T) {
		servTestConfig.Mode = disableExpire

		resp, err := http.Get("http://" + servTestConfig.ListenHostPort)
		assert.NoError(t, err)
		defer resp.Body.Close()

		assert.Equal(t, http.StatusOK, resp.StatusCode)
		time.Sleep(time.Second)
	})

	t.Run("generate UI interface", func(t *testing.T) {
		resp, err := http.Get("http://" + servTestConfig.ListenHostPort + "/ui/generate")
		assert.NoError(t, err)
		defer resp.Body.Close()

		assert.Equal(t, http.StatusOK, resp.StatusCode)

		buf, err := io.ReadAll(resp.Body)
		assert.NoError(t, err)

		assert.Contains(t, string(buf), "URL to be shortened")
	})

	t.Run("generate UI interface with QR", func(t *testing.T) {
		resp, err := http.Get("http://" + servTestConfig.ListenHostPort + "/ui/generate?s=http:/some.url")
		assert.NoError(t, err)
		defer resp.Body.Close()

		assert.Equal(t, http.StatusOK, resp.StatusCode)

		buf, err := io.ReadAll(resp.Body)
		assert.NoError(t, err)
		assert.Contains(t, string(buf), "Short URL:")
	})

	t.Run("generate UI interface disabled", func(t *testing.T) {
		servTestConfig.Mode = disableUI
		resp, err := http.Get("http://" + servTestConfig.ListenHostPort + "/ui/generate?s=http:/some.url")
		assert.NoError(t, err)
		defer resp.Body.Close()

		assert.Equal(t, http.StatusNotFound, resp.StatusCode)
	})

	t.Run("stop service", func(t *testing.T) {
		logger := log.Writer()
		r, w, _ := os.Pipe()
		log.SetOutput(w)

		servTestHandler.stop()
		time.Sleep(time.Second)

		w.Close()
		log.SetOutput(logger)
		buf, err := io.ReadAll(r)
		assert.NoError(t, err)

		assert.Contains(t, string(buf), "http: Server closed")
	})
}

// try tokens' duplicate
func Test10Serv90Duble(t *testing.T) {
	godotenv.Load()

	servTestConfig, err := readConfig()
	assert.NoError(t, err)

	servTestDB, err := NewTokenDB(servTestConfig.RedisAddrs, servTestConfig.RedisPassword)
	assert.NoError(t, err)

	// create short token interface
	sToken := mockShortToken(servTestConfig.TokenLength)

	// create service handler
	servTestHandler = NewHandler(servTestConfig, servTestDB, sToken)

	servTestDB.Delete(sToken.Get())

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

	servTestDB.Delete(sToken.Get())

	servTestHandler.stop()
}

func Test10Serv92BadDB(t *testing.T) {

	godotenv.Load()
	servTestConfig, err := readConfig()
	assert.NoError(t, err)

	servTestDB := newMockDB()
	servTestDB.setFunc = func(string, string, int) (bool, error) { return false, errors.New("some error") }

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
