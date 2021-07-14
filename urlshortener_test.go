package main

import (
	"errors"
	"io"
	"log"
	"os"
	"syscall"
	"testing"
	"time"

	"github.com/joho/godotenv"
	"github.com/stretchr/testify/assert"
)

// try to start with wrong path to configuration file
func Test20Main00WrongConfig(t *testing.T) {
	// use saveEnv from tools_test
	defer saveEnv()()
	os.Unsetenv("URLSHORTENER_REDISADDRS")

	err := doMain()

	assert.Error(t, err)
	assert.Equal(t, "configuration read error: config error: required key URLSHORTENER_REDISADDRS missing value", err.Error())
}

// try to pass wrong path to config
func Test20Main05WrongDB(t *testing.T) {
	// use saveEnv from tools_test to save/restore the environment
	defer saveEnv()()
	os.Setenv("URLSHORTENER_REDISADDRS", "wrong.host:1234")
	// defer the panic recovery and error handling
	defer func() {
		if err := recover(); err != nil {
			err := err.(error)
			assert.Error(t, err)
			assert.Equal(t, "database interface creation error: dial tcp: lookup wrong.host: no such host", err.Error())
		}
	}()
	// run service
	main()
	t.Error("No panic when expected")
	// we shouldn't get here as main() have to panic with wrong DB connection address
	// handle this in defer function
}

func Test20Main07WrongDB2(t *testing.T) {
	conf := Config{
		ListenHostPort: "localhost:8080",
		ShortDomain:    "localhost:8080",
		Timeout:        500,
		TokenLength:    6,
	}
	errDb := newMockDB()
	errDb.setFunc = func(string, string, int) (bool, error) { return false, errors.New("some error") }
	err := stratService(&conf, errDb)
	assert.Error(t, err)
	assert.Equal(t, "http: Server closed", err.Error())
}

// try to start service correctly
func Test20Main20SuccessAndKill(t *testing.T) {
	logger := log.Writer()
	r, w, _ := os.Pipe()
	log.SetOutput(w)

	godotenv.Load()
	// run service
	go main()

	time.Sleep(time.Second * 2)

	w.Close()
	log.SetOutput(logger)
	buf, err := io.ReadAll(r)
	assert.NoError(t, err)

	assert.Contains(t, string(buf), "starting server at")
	assert.Contains(t, string(buf), "URLshortener "+version)

	logger = log.Writer()
	r, w, _ = os.Pipe()
	log.SetOutput(w)

	syscall.Kill(syscall.Getpid(), syscall.SIGINT)

	time.Sleep(time.Second * 2)

	w.Close()
	log.SetOutput(logger)
	buf, err = io.ReadAll(r)
	assert.NoError(t, err)

	assert.Contains(t, string(buf), "http: Server closed")
}
