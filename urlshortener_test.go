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

	"github.com/go-redis/redis/v7"
)

// try to start with wrong path to configuration file
func Test20Main00WrongConfig(t *testing.T) {
	// use saveEnv from tools_test
	defer saveEnv()()
	os.Unsetenv("URLSHORTENER_ConnectOptions")

	err := doMain("/bad/path/to/config/file", make(chan bool, 1))

	if err == nil {
		t.Error("no error when expected")
	}
	if !strings.HasPrefix(err.Error(), "configuration read error") {
		t.Errorf("wrong error: %v", err)
	}
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
		err := recover()
		if err == nil {
			t.Error("no error when expected")
		}
		if !strings.HasPrefix(err.(error).Error(), "database interface creation error") {
			t.Errorf("wrong error received: %v", err)
		}
	}()
	// run service
	main()
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
	errDb, _ := testDBNewTokenDB(redis.UniversalOptions{})
	exit := make(chan bool, 1)
	go stratService(&conf, errDb, exit)

	select {
	case <-time.After(time.Second * 2):
		t.Error("no exit for 2 secs")
	case <-exit:
		break
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
}

// try to start service correctly
func Test20Main20Success(t *testing.T) {
	logger := log.Writer()
	r, w, _ := os.Pipe()
	log.SetOutput(w)

	defer func() {
		if err := recover(); err != nil {
			t.Errorf("server starting error:%v", err)
		}
	}()
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
