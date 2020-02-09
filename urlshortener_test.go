package main

import (
	"bytes"
	"log"
	"os"
	"os/exec"
	"strings"
	"testing"
)

// try to start with wrong path to configuration file
func Test20Main00WrongConfig(t *testing.T) {
	// use saveEnv from tools_test
	defer saveEnv()()
	os.Unsetenv("URLSHORTENER_ConnectOptions")
	os.Args = []string{"prog", "-config=/bad/path/to/config/file"}
	CONFIG = Config{}

	err := doMain()

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
	os.Args = []string{"prog", "-config=/bad/path"}
	CONFIG = Config{}

	err := doMain()

	if err == nil {
		t.Error("no error when expected")
	}
	if !strings.HasPrefix(err.Error(), "error database interface creation") {
		t.Errorf("wrong error: %v", err)
	}
}

// try to get usage message
func Test20Main10Usage(t *testing.T) {

	err := exec.Command("go", "build").Run()
	if err != nil {
		t.Errorf("building error: %v", err)
	}

	buf, err := exec.Command("./URLshortener", "-wrongOption").CombinedOutput()

	if err == nil {
		t.Error("no error when expected")
	}
	t.Logf("received expected error: %v", err)

	if !bytes.Contains(buf, []byte("Usage:")) {
		t.Errorf("received unexpected output: %s", buf)
	}
	log.Printf("%s", buf)
}
