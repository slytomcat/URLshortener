package main

import (
	"bytes"
	"log"
	"os"
	"os/exec"
	"strings"
	"testing"
)

// saveCfgArgs saves CONFIG, os.Args and URLSHORTENER_ConnectOptions eviroment variable
// and returns the function that restores them to original values
func saveCfgArgs() func() {
	saveConnectOptions := os.Getenv("URLSHORTENER_ConnectOptions")
	saveArgs := os.Args
	saveCONFIG := CONFIG

	return func() {
		os.Setenv("URLSHORTENER_ConnectOptions", saveConnectOptions)
		os.Args = saveArgs
		CONFIG = saveCONFIG
	}
}

// try to start with wrong path to configuration file
func Test20Main00WrongConfig(t *testing.T) {
	defer saveCfgArgs()()
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
	defer saveCfgArgs()()
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
