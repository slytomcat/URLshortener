package main

import (
	"errors"
	"io"
	"log"
	"os"
	"syscall"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

// try to start with wrong path to configuration file
func Test20Main00WrongConfig(t *testing.T) {
	t.Setenv("URLSHORTENER_REDISADDRS", "")
	err := doMain()

	require.Error(t, err)
	require.Equal(t, "configuration read error: config error: required key URLSHORTENER_REDISADDRS missing value", err.Error())
}

// try to pass wrong addr of redis server
func Test20Main05WrongDB(t *testing.T) {
	t.Setenv("URLSHORTENER_REDISADDRS", "wrong.host:1234")
	require.PanicsWithError(t, "database interface creation error: dial tcp: lookup wrong.host on 127.0.0.53:53: no such host", main)
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
	require.Error(t, err)
	require.Equal(t, "http: Server closed", err.Error())
}

// try to start service correctly
func Test20Main20SuccessAndKill(t *testing.T) {
	logger := log.Writer()
	r, w, _ := os.Pipe()
	log.SetOutput(w)

	envSet(t)

	// run service
	go main()

	time.Sleep(time.Second * 2)

	w.Close()
	log.SetOutput(logger)
	buf, err := io.ReadAll(r)
	require.NoError(t, err)

	require.Contains(t, string(buf), "starting server at")
	require.Contains(t, string(buf), "URLshortener "+version)

	logger = log.Writer()
	r, w, _ = os.Pipe()
	log.SetOutput(w)

	syscall.Kill(syscall.Getpid(), syscall.SIGINT)

	time.Sleep(time.Second * 2)

	w.Close()
	log.SetOutput(logger)
	buf, err = io.ReadAll(r)
	require.NoError(t, err)

	require.Contains(t, string(buf), "http: Server closed")
}
