package main

import (
	"errors"
	"io"
	"log"
	"os"
	"syscall"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// try to start with wrong path to configuration file
func Test20Main00WrongConfig(t *testing.T) {
	t.Setenv("URLSHORTENER_REDISADDRS", "")
	err := doMain()
	require.Error(t, err)
	require.Equal(t, "configuration read error: config error: wrong or missed value of URLSHORTENER_REDISADDRS", err.Error())
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
	err := startService(&conf, errDb)
	require.Error(t, err)
	require.Equal(t, "http: Server closed", err.Error())
}

func catchOutput() func() string {
	out := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w
	return func() string {
		w.Close()
		os.Stdout = out
		buf, err := io.ReadAll(r)
		if err != nil {
			panic(err)
		}
		return string(buf)
	}
}

func catchLog() func() string {
	out := log.Writer()
	r, w, _ := os.Pipe()
	log.SetOutput(w)
	return func() string {
		w.Close()
		log.SetOutput(out)
		buf, err := io.ReadAll(r)
		if err != nil {
			panic(err)
		}
		return string(buf)
	}
}

// try to start service correctly
func Test20Main20SuccessAndKill(t *testing.T) {
	outF := catchLog()
	envSet(t)
	// run service
	go main()
	time.Sleep(time.Second * 2)
	out := outF()
	assert.Contains(t, out, "starting server at")
	assert.Contains(t, out, "URLshortener "+version)
	outF = catchLog()
	syscall.Kill(syscall.Getpid(), syscall.SIGINT)
	time.Sleep(time.Second * 2)
	require.Contains(t, outF(), "http: Server closed")
}

func TestMainVersion(t *testing.T) {
	args := os.Args
	os.Args = []string{"app", "-v"}
	defer func() {
		os.Args = args
	}()
	outF := catchOutput()
	main()
	require.Contains(t, outF(), version)
}
