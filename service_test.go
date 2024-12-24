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

	"github.com/stretchr/testify/require"
)

var (
	serviceTestConfig  *Config
	serviceTestDB      TokenDB
	serviceTestHandler ServiceHandler
)

// try to start service with not working db
func Test10Service03CheckHealthCheck(t *testing.T) {
	conf := Config{
		ListenHostPort: "localhost:8080",
		ShortDomain:    "localhost:8080",
		Timeout:        500,
		TokenLength:    6,
	}

	errDb := newMockDB()

	testHandler := NewHandler(&conf, errDb, NewShortToken(5))

	go func() {
		require.Error(t, testHandler.start())
	}()
	require.Eventually(t, checkStart("http://localhost:8080/"), time.Second, 10*time.Millisecond)

	conf.Mode = disableExpire

	resp, err := http.Get("http://localhost:8080/healthcheck")
	require.NoError(t, err)
	defer resp.Body.Close()
	require.Equal(t, http.StatusOK, resp.StatusCode)

	conf.Mode = 0
	errDb.expFunc = func(string, int) error { return errors.New("some error") }

	resp, err = http.Get("http://localhost:8080/healthcheck")
	require.NoError(t, err)
	defer resp.Body.Close()
	require.Equal(t, http.StatusInternalServerError, resp.StatusCode)

	conf.Mode = disableExpire

	resp, err = http.Get("http://localhost:8080/healthcheck")
	require.NoError(t, err)
	defer resp.Body.Close()
	require.Equal(t, http.StatusInternalServerError, resp.StatusCode)

	conf.Mode = disableShortener | disableRedirect
	errDb.getFunc = func(s string) (string, error) { return "wrongURL", nil }

	resp, err = http.Get("http://localhost:8080/healthcheck")
	require.NoError(t, err)
	defer resp.Body.Close()
	require.Equal(t, http.StatusInternalServerError, resp.StatusCode)

	conf.Mode = 0
	errDb.getFunc = func(string) (string, error) { return "", errors.New("some error") }

	resp, err = http.Get("http://localhost:8080/healthcheck")
	require.NoError(t, err)
	defer resp.Body.Close()
	require.Equal(t, http.StatusInternalServerError, resp.StatusCode)

	conf.Mode = disableRedirect

	resp, err = http.Get("http://localhost:8080/healthcheck")
	require.NoError(t, err)
	defer resp.Body.Close()
	require.Equal(t, http.StatusInternalServerError, resp.StatusCode)

	conf.Mode = 0
	errDb.setFunc = func(string, string, int) (bool, error) { return false, nil }

	resp, err = http.Get("http://localhost:8080/healthcheck")
	require.NoError(t, err)
	defer resp.Body.Close()
	require.Equal(t, http.StatusInternalServerError, resp.StatusCode)

	conf.Mode = disableShortener

	resp, err = http.Get("http://localhost:8080/healthcheck")
	require.NoError(t, err)
	defer resp.Body.Close()
	require.Equal(t, http.StatusInternalServerError, resp.StatusCode)

	errDb.closeFunc = func() error { return errors.New("some error") }

	testHandler.stop()

}

func checkStart(url string) func() bool {
	return func() bool {
		resp, err := http.Get(url)
		if err == nil {
			defer resp.Body.Close()
		}
		if err != nil || resp.StatusCode != http.StatusOK {
			return false
		}
		return true
	}
}

// try to start service
func Test10Service05All(t *testing.T) {
	envSet(t)

	testConfig, err := readConfig()
	require.NoError(t, err)

	// initialize database connection
	serviceTestDB, err = NewTokenDB(testConfig.RedisAddrs, testConfig.RedisPassword)
	require.NoError(t, err)

	// create service handler
	serviceTestHandler = NewHandler(testConfig, serviceTestDB, NewShortToken(testConfig.TokenLength))

	require.NoError(t, err)
	logger := log.Writer()
	r, w, _ := os.Pipe()
	log.SetOutput(w)

	go func() {
		log.Println(serviceTestHandler.start())
	}()

	require.Eventually(t, checkStart("http://"+testConfig.ListenHostPort), time.Second, 10*time.Millisecond)
	w.Close()
	log.SetOutput(logger)
	buf, err := io.ReadAll(r)
	require.NoError(t, err)
	require.Contains(t, string(buf), "starting server at")

	t.Run("do health check", func(t *testing.T) {
		resp, err := http.Get("http://" + testConfig.ListenHostPort + "/healthcheck")
		require.NoError(t, err)
		defer resp.Body.Close()

		require.Equal(t, http.StatusOK, resp.StatusCode)
		buf, err := io.ReadAll(resp.Body)
		require.NoError(t, err)
		require.Contains(t, string(buf), "Home page of URLshortener")
	})

	t.Run("bad method", func(t *testing.T) {
		resp, err := http.Head("http://" + testConfig.ListenHostPort)
		require.NoError(t, err)
		defer resp.Body.Close()
		require.Equal(t, http.StatusBadRequest, resp.StatusCode)
	})

	t.Run("short URL query with empty request body", func(t *testing.T) {
		resp, err := http.Post("http://"+testConfig.ListenHostPort+"/api/v1/token", "application/json",
			strings.NewReader(``))
		require.NoError(t, err)
		defer resp.Body.Close()

		require.Equal(t, http.StatusBadRequest, resp.StatusCode)
	})

	t.Run("short URL request with empty JSON", func(t *testing.T) {
		resp, err := http.Post("http://"+testConfig.ListenHostPort+"/api/v1/token", "application/json",
			strings.NewReader(`{}`))
		require.NoError(t, err)
		defer resp.Body.Close()

		require.Equal(t, http.StatusBadRequest, resp.StatusCode)
	})

	t.Run("short URL request without expiration", func(t *testing.T) {
		resp, err := http.Post("http://"+testConfig.ListenHostPort+"/api/v1/token", "application/json",
			strings.NewReader(`{"url": "http://`+testConfig.ShortDomain+`"}`))
		require.NoError(t, err)
		defer resp.Body.Close()
		require.Equal(t, http.StatusOK, resp.StatusCode)
	})

	t.Run("expire request without parameters", func(t *testing.T) {
		resp, err := http.Post("http://"+testConfig.ListenHostPort+"/api/v1/expire", "application/json",
			strings.NewReader(``))
		require.NoError(t, err)
		defer resp.Body.Close()

		require.Equal(t, http.StatusBadRequest, resp.StatusCode)
	})

	t.Run("expire request with empty JSON", func(t *testing.T) {
		resp, err := http.Post("http://"+testConfig.ListenHostPort+"/api/v1/expire", "application/json",
			strings.NewReader(`{}`))
		require.NoError(t, err)
		defer resp.Body.Close()

		require.Equal(t, http.StatusBadRequest, resp.StatusCode)
	})

	t.Run("expire request for not existing token", func(t *testing.T) {
		resp, err := http.Post("http://"+testConfig.ListenHostPort+"/api/v1/expire", "application/json",
			strings.NewReader(`{"token":"`+strings.Repeat("(", testConfig.TokenLength)+`"}`)) // use non Base64 symbols
		require.NoError(t, err)
		defer resp.Body.Close()

		require.Equal(t, http.StatusNotModified, resp.StatusCode)
	})

	t.Run("redirect request with wrong token (wrong length)", func(t *testing.T) {
		resp, err := http.Get("http://" + testConfig.ListenHostPort + "/not+existing+token")
		require.NoError(t, err)
		defer resp.Body.Close()

		require.Equal(t, http.StatusNotFound, resp.StatusCode)
	})

	t.Run("redirect request with wrong token (wrong symbols)", func(t *testing.T) {
		resp, err := http.Get("http://" + testConfig.ListenHostPort + "/" + strings.Repeat("(", testConfig.TokenLength))
		require.NoError(t, err)
		defer resp.Body.Close()

		require.Equal(t, http.StatusNotFound, resp.StatusCode)
	})

	t.Run("redirect request with wrong token (correct length&symbols)", func(t *testing.T) {
		resp, err := http.Get("http://" + testConfig.ListenHostPort + "/" + strings.Repeat("A", testConfig.TokenLength))
		require.NoError(t, err)
		defer resp.Body.Close()

		require.Equal(t, http.StatusNotFound, resp.StatusCode)
	})

	t.Run("unsupported request in mode = disableRedirect", func(t *testing.T) {
		testConfig.Mode = disableRedirect

		resp, err := http.Get("http://" + testConfig.ListenHostPort + "/" + strings.Repeat("_", testConfig.TokenLength))
		require.NoError(t, err)
		defer resp.Body.Close()

		require.Equal(t, http.StatusNotFound, resp.StatusCode)
	})

	t.Run("unsupported request in mode = disableShortener", func(t *testing.T) {
		testConfig.Mode = disableShortener

		resp, err := http.Post("http://"+testConfig.ListenHostPort+"/api/v1/token", "application/json",
			strings.NewReader(`{"url": "http://someother.url"}`))
		require.NoError(t, err)
		defer resp.Body.Close()

		require.Equal(t, http.StatusNotFound, resp.StatusCode)
	})

	t.Run("unsupported request in mode = disableExpire", func(t *testing.T) {
		testConfig.Mode = disableExpire

		resp, err := http.Post("http://"+testConfig.ListenHostPort+"/api/v1/expire", "application/json",
			strings.NewReader(`{"token": "`+strings.Repeat("_", testConfig.TokenLength)+`","exp":-1}`))
		require.NoError(t, err)
		defer resp.Body.Close()

		require.Equal(t, http.StatusNotFound, resp.StatusCode)
	})

	t.Run("health check in service mode disableRedirect", func(t *testing.T) {
		testConfig.Mode = disableRedirect

		resp, err := http.Get("http://" + testConfig.ListenHostPort + "/healthcheck")
		require.NoError(t, err)
		defer resp.Body.Close()

		require.Equal(t, http.StatusOK, resp.StatusCode)
		time.Sleep(time.Second)
	})

	t.Run("health check in service mode disableShortener", func(t *testing.T) {
		testConfig.Mode = disableShortener

		resp, err := http.Get("http://" + testConfig.ListenHostPort + "/healthcheck")
		require.NoError(t, err)
		defer resp.Body.Close()

		require.Equal(t, http.StatusOK, resp.StatusCode)
		time.Sleep(time.Second)
	})

	t.Run("health check in service mode disableExpire", func(t *testing.T) {
		testConfig.Mode = disableExpire

		resp, err := http.Get("http://" + testConfig.ListenHostPort + "/healthcheck")
		require.NoError(t, err)
		defer resp.Body.Close()

		require.Equal(t, http.StatusOK, resp.StatusCode)
		time.Sleep(time.Second)
	})

	t.Run("generate UI interface", func(t *testing.T) {
		resp, err := http.Get("http://" + testConfig.ListenHostPort + "/ui/generate")
		require.NoError(t, err)
		defer resp.Body.Close()

		require.Equal(t, http.StatusOK, resp.StatusCode)

		buf, err := io.ReadAll(resp.Body)
		require.NoError(t, err)

		require.Contains(t, string(buf), "URL to be shortened")
	})

	t.Run("generate UI interface with QR", func(t *testing.T) {
		resp, err := http.Get("http://" + testConfig.ListenHostPort + "/ui/generate?s=http:/some.url")
		require.NoError(t, err)
		defer resp.Body.Close()

		require.Equal(t, http.StatusOK, resp.StatusCode)

		buf, err := io.ReadAll(resp.Body)
		require.NoError(t, err)
		require.Contains(t, string(buf), "Short URL:")
	})

	t.Run("generate UI interface disabled", func(t *testing.T) {
		testConfig.Mode = disableUI
		resp, err := http.Get("http://" + testConfig.ListenHostPort + "/ui/generate?s=http:/some.url")
		require.NoError(t, err)
		defer resp.Body.Close()

		require.Equal(t, http.StatusNotFound, resp.StatusCode)
	})

	t.Run("stop service", func(t *testing.T) {
		logger := log.Writer()
		r, w, _ := os.Pipe()
		log.SetOutput(w)

		serviceTestHandler.stop()
		time.Sleep(time.Second)

		w.Close()
		log.SetOutput(logger)
		buf, err := io.ReadAll(r)
		require.NoError(t, err)

		require.Contains(t, string(buf), "http: Server closed")
	})
}

// try tokens' duplicate
func Test10Service90Double(t *testing.T) {
	envSet(t)

	servTestConfig, err := readConfig()
	require.NoError(t, err)

	servTestDB, err := NewTokenDB(servTestConfig.RedisAddrs, servTestConfig.RedisPassword)
	require.NoError(t, err)

	// create short token interface
	sToken := mockShortToken(servTestConfig.TokenLength)

	// create service handler
	serviceTestHandler = NewHandler(servTestConfig, servTestDB, sToken)

	servTestDB.Delete(sToken.Get())

	go func() {
		log.Println(serviceTestHandler.start())
	}()

	require.Eventually(t, checkStart("http://"+servTestConfig.ListenHostPort), time.Second, 10*time.Millisecond)

	resp, err := http.Post("http://"+servTestConfig.ListenHostPort+"/api/v1/token", "application/json",
		strings.NewReader(`{"url": "http://`+servTestConfig.ShortDomain+`"}`))
	require.NoError(t, err)
	resp.Body.Close()

	require.Equal(t, http.StatusOK, resp.StatusCode)
	// second request

	resp2, err := http.Post("http://"+servTestConfig.ListenHostPort+"/api/v1/token", "application/json",
		strings.NewReader(`{"url": "http://`+servTestConfig.ShortDomain+`"}`))
	require.NoError(t, err)
	resp.Body.Close()
	require.Equal(t, http.StatusRequestTimeout, resp2.StatusCode)

	servTestDB.Delete(sToken.Get())

	serviceTestHandler.stop()
}

func Test10Service92BadDB(t *testing.T) {
	envSet(t)

	servTestConfig, err := readConfig()
	require.NoError(t, err)

	servTestDB := newMockDB()
	servTestDB.setFunc = func(string, string, int) (bool, error) { return false, errors.New("some error") }

	// create short token interface
	sToken := NewShortToken(5)

	// create service handler
	serviceTestHandler = NewHandler(servTestConfig, servTestDB, sToken)

	go func() {
		log.Println(serviceTestHandler.start())
	}()

	require.Eventually(t, checkStart("http://"+servTestConfig.ListenHostPort), time.Second, 10*time.Millisecond)

	resp, err := http.Post("http://"+servTestConfig.ListenHostPort+"/api/v1/token", "application/json",
		strings.NewReader(`{"url": "http://`+servTestConfig.ShortDomain+`"}`))
	require.NoError(t, err)
	resp.Body.Close()

	require.Equal(t, http.StatusInternalServerError, resp.StatusCode)
	serviceTestHandler.stop()
}
