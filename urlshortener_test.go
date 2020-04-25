package main

import (
	"bytes"
	"flag"
	"io/ioutil"
	"log"
	"os"
	"strings"
	"syscall"
	"testing"
	"time"
)

// try to start with wrong path to configuration file
func Test20Main00WrongConfig(t *testing.T) {
	// use saveEnv from tools_test
	defer saveEnv()()
	os.Unsetenv("URLSHORTENER_ConnectOptions")

	err := doMain("/bad/path/to/config/file", nil)

	if err == nil {
		t.Error("no error when expected")
	}
	if !strings.HasPrefix(err.Error(), "configuration read error") {
		t.Errorf("wrong error: %v", err)
	}
}

// try to pass wrong path to config
func Test20Main05WrongDB(t *testing.T) {
	// use saveEnv from tools_test
	defer saveEnv()()
	os.Setenv("URLSHORTENER_ConnectOptions", `{"Addrs":["wrong.host:6379"]}`)

	err := doMain("/bad/path", nil)

	if err == nil {
		t.Error("no error when expected")
	}
	if !strings.HasPrefix(err.Error(), "database interface creation error") {
		t.Errorf("wrong error: %v", err)
	}
}

// try to get usage message
func Test20Main10Usage(t *testing.T) {
	logger := os.Stderr
	r, w, _ := os.Pipe()
	os.Stderr = w

	flag.Usage()

	w.Close()
	os.Stderr = logger
	buf, err := ioutil.ReadAll(r)
	if err != nil {
		t.Error(err)
	}

	if !bytes.Contains(buf, []byte("[-config=<Path/to/config>]")) {
		t.Errorf("received unexpected output: %s", buf)
	}
	log.Printf("%s", buf)
}

// try to start correctly
func Test20Main20Success(t *testing.T) {
	logger := log.Writer()
	r, w, _ := os.Pipe()
	log.SetOutput(w)

	// run service
	go main()

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
	if !bytes.Contains(buf, []byte("URLshortener "+version)) {
		t.Errorf("no version shown on start: %s", buf)
	}
	log.Printf("%s", buf)
}

func Test20Main25Kill(t *testing.T) {

	logger := log.Writer()
	r, w, _ := os.Pipe()
	log.SetOutput(w)

	syscall.Kill(syscall.Getpid(), syscall.SIGINT)

	time.Sleep(time.Second * 2)

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
