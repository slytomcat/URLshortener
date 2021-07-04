package main

import (
	"flag"
	"io"
	"log"
	"os"
	"syscall"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

// try to start with wrong path to configuration file
func Test20Main00WrongConfig(t *testing.T) {
	// use saveEnv from tools_test
	defer saveEnv()()
	os.Unsetenv("URLSHORTENER_ConnectOptions")

	err := doMain("/bad/path/to/config/file")

	assert.Error(t, err)
	assert.Equal(t, "configuration read error: mandatory configuration value ConnectOptions is not set", err.Error())
}

// try to pass wrong path to config
func Test20Main05WrongDB(t *testing.T) {
	// use saveEnv from tools_test to save/restore the environment
	defer saveEnv()()
	os.Setenv("URLSHORTENER_ConnectOptions", `{"Addrs":["wrong.host:6379"]}`)
	// save configFile and restore it in defer func
	SaveConfigFile := configFile
	configFile = "/bad/path/to/config/file"
	defer func() { configFile = SaveConfigFile }()
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
	err := stratService(&conf, errDb)
	assert.Error(t, err)
	assert.Equal(t, "http: Server closed", err.Error())
}

// try to get usage message
func Test20Main10Usage(t *testing.T) {
	logger := os.Stderr
	r, w, _ := os.Pipe()
	os.Stderr = w

	flag.Usage()

	w.Close()
	os.Stderr = logger
	buf, err := io.ReadAll(r)
	assert.NoError(t, err)
	assert.Contains(t, string(buf), "[-config=<Path/to/config>]")
}

// try to start service correctly
func Test20Main20SuccessAndKill(t *testing.T) {
	logger := log.Writer()
	r, w, _ := os.Pipe()
	log.SetOutput(w)

	defer assert.Nil(t, recover())
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
