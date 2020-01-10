package main

import (
	"io/ioutil"
	"os"
	"testing"
)

// test config reading from nonexistent file
func Test01Tools00WrongFile(t *testing.T) {
	saveDSN := os.Getenv("URLSHORTENER_DSN")
	saveCONFIG := CONFIG
	defer func() {
		CONFIG = saveCONFIG
		os.Setenv("URLSHORTENER_DSN", saveDSN)
	}()
	os.Unsetenv("URLSHORTENER_DSN")
	CONFIG = Config{}

	err := readConfig("wrong.wrong.wrong.file.json")

	if err == nil {
		t.Error("no error for wrong filename")
	}
	t.Logf("Receved: %v", err)
}

// test it with emty file with URLSHORTENER_DSN unset
func Test01Tools05EmptyFile(t *testing.T) {
	tmpfile, err := ioutil.TempFile(os.TempDir(), "testing*.json")
	if err != nil {
		t.Errorf("temp file creation error: %w", err)
	}
	defer os.Remove(tmpfile.Name())

	if err := tmpfile.Close(); err != nil {
		t.Errorf("temp file closing error: %w", err)
	}

	saveDSN := os.Getenv("URLSHORTENER_DSN")
	saveCONFIG := CONFIG
	defer func() {
		CONFIG = saveCONFIG
		os.Setenv("URLSHORTENER_DSN", saveDSN)
	}()
	os.Unsetenv("URLSHORTENER_DSN")
	CONFIG = Config{}

	err = readConfig(tmpfile.Name())

	if err == nil {
		t.Error("no error for empty file")
	}
	t.Logf("Receved: %v", err)
}

// test it with empty JSON with URLSHORTENER_DSN unset
func Test01Tools10EmptyJSON(t *testing.T) {
	tmpfile, err := ioutil.TempFile(os.TempDir(), "testing*.json")
	if err != nil {
		t.Errorf("temp file creation error: %w", err)
	}
	defer os.Remove(tmpfile.Name())

	err = ioutil.WriteFile(tmpfile.Name(), []byte(`{ }`), 0600)
	if err != nil {
		t.Errorf("temp file write error: %v", err)
	}
	saveDSN := os.Getenv("URLSHORTENER_DSN")
	saveCONFIG := CONFIG
	defer func() {
		CONFIG = saveCONFIG
		os.Setenv("URLSHORTENER_DSN", saveDSN)
	}()
	os.Unsetenv("URLSHORTENER_DSN")
	CONFIG = Config{}

	err = readConfig(tmpfile.Name())

	if err == nil {
		t.Error("no error for empty JSON with URLSHORTENER_DSN unset")
	}
	t.Logf("Receved: %v", err)
}

// test it with empty JSON but with set URLSHORTENER_DSN
func Test02Tools15EmptyJSON_(t *testing.T) {
	tmpfile, err := ioutil.TempFile(os.TempDir(), "testing*.json")
	if err != nil {
		t.Errorf("temp file creation error: %w", err)
	}
	defer os.Remove(tmpfile.Name())

	err = ioutil.WriteFile(tmpfile.Name(), []byte(`{ }`), 0600)
	if err != nil {
		t.Errorf("temp file write error: %v", err)
	}

	saveCONFIG := CONFIG
	saveDriver := os.Getenv("URLSHORTENER_DBdriver")
	saveDSN := os.Getenv("URLSHORTENER_DSN")
	defer func() {
		CONFIG = saveCONFIG
		os.Setenv("URLSHORTENER_DBdriver", saveDriver)
		os.Setenv("URLSHORTENER_DSN", saveDSN)
	}()

	CONFIG = Config{}
	os.Setenv("URLSHORTENER_DBdriver", "testDBdriver")
	os.Setenv("URLSHORTENER_DSN", "testDSNvalue")

	err = readConfig(tmpfile.Name())

	if err != nil {
		t.Error("error for empty JSON with set URLSHORTENER_DSN")
	}
	if CONFIG.DSN != "testDSNvalue" ||
		CONFIG.Timeout != DefaultTimeout ||
		CONFIG.ListenHostPort != DefaultListenHostPort ||
		CONFIG.DefaultExp != DefaultDefaultExp ||
		CONFIG.ShortDomain != DefaultShortDomain ||
		CONFIG.Mode != DefaultMode {
		t.Error("Wrong default values set")
	}
}

// test full success from example.cnf.json
func Test03Tools20FullJSON(t *testing.T) {
	saveCONFIG := CONFIG
	saveDriver := os.Getenv("URLSHORTENER_DBdriver")
	saveDSN := os.Getenv("URLSHORTENER_DSN")
	defer func() {
		CONFIG = saveCONFIG
		os.Setenv("URLSHORTENER_DBdriver", saveDriver)
		os.Setenv("URLSHORTENER_DSN", saveDSN)
	}()

	CONFIG = Config{}
	os.Unsetenv("URLSHORTENER_DBdriver")
	os.Unsetenv("URLSHORTENER_DSN")

	err := readConfig("example.cnf.json")

	if err != nil {
		t.Errorf("error reading of example.cnf.json: %v", err)
	}
	if CONFIG.DSN != "shortener:<password>@<protocol>(<host>:<port>)/shortener_DB" ||
		CONFIG.Timeout != 777 ||
		CONFIG.ListenHostPort != "0.0.0.0:80" ||
		CONFIG.DefaultExp != 30 ||
		CONFIG.ShortDomain != "<shortDomain>" ||
		CONFIG.Mode != 1 {
		t.Error("Wrong values set")
	}
}

func saveEnv() func() {
	saveCONFIG := CONFIG
	saveDSN := os.Getenv("URLSHORTENER_DSN")
	saveTimeout := os.Getenv("URLSHORTENER_Timeout")
	saveHost := os.Getenv("URLSHORTENER_ListenHostPort")
	saveExp := os.Getenv("URLSHORTENER_DefaultExp")
	saveDomain := os.Getenv("URLSHORTENER_ShortDomain")
	saveMode := os.Getenv("URLSHORTENER_Mode")

	return func() {
		CONFIG = saveCONFIG
		os.Setenv("URLSHORTENER_DSN", saveDSN)
		os.Setenv("URLSHORTENER_Timeout", saveTimeout)
		os.Setenv("URLSHORTENER_ListenHostPort", saveHost)
		os.Setenv("URLSHORTENER_DefaultExp", saveExp)
		os.Setenv("URLSHORTENER_ShortDomain", saveDomain)
		os.Setenv("URLSHORTENER_Mode", saveMode)
	}

}

// test full success from environment variables
func Test03Tools30FullEnv(t *testing.T) {
	defer saveEnv()()

	CONFIG = Config{}
	os.Setenv("URLSHORTENER_DSN", "TestDSN:DSN")
	os.Setenv("URLSHORTENER_Timeout", "787")
	os.Setenv("URLSHORTENER_ListenHostPort", "testHost:testPort")
	os.Setenv("URLSHORTENER_DefaultExp", "42")
	os.Setenv("URLSHORTENER_ShortDomain", "test.domain")
	os.Setenv("URLSHORTENER_Mode", "66")

	err := readConfig("wrong.wrong.wrong.file.json")
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}

	if CONFIG.DSN != "TestDSN:DSN" ||
		CONFIG.Timeout != 787 ||
		CONFIG.ListenHostPort != "testHost:testPort" ||
		CONFIG.DefaultExp != 42 ||
		CONFIG.ShortDomain != "test.domain" ||
		CONFIG.Mode != 66 {
		t.Error("Wrong values set")
	}
}

// test wromg enviroment variable URLSHORTENER_MaxOpenConns
func Test03Tools35WrongTimeout(t *testing.T) {
	defer saveEnv()()

	CONFIG = Config{}
	os.Setenv("URLSHORTENER_DBdriver", "TestDriver")
	os.Setenv("URLSHORTENER_DSN", "TestDSN:DSN")
	os.Setenv("URLSHORTENER_Timeout", "@#2$")

	err := readConfig("wrong.wrong.wrong.file.json")
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}

	if CONFIG.Timeout != DefaultTimeout {
		t.Error("Wrong values set")
	}
}

// test wromg enviroment variable URLSHORTENER_DefaultExp
func Test03Tools40WrongEnvDefaultExp(t *testing.T) {
	defer saveEnv()()

	CONFIG = Config{}
	os.Setenv("URLSHORTENER_DBdriver", "TestDriver")
	os.Setenv("URLSHORTENER_DSN", "TestDSN:DSN")
	os.Setenv("URLSHORTENER_DefaultExp", "@#2$")

	err := readConfig("wrong.wrong.wrong.file.json")
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}

	if CONFIG.DefaultExp != DefaultDefaultExp {
		t.Error("Wrong values set")
	}
}

// test wromg enviroment variable URLSHORTENER_Mode
func Test03Tools45WrongEnvMode(t *testing.T) {
	defer saveEnv()()

	CONFIG = Config{}
	os.Setenv("URLSHORTENER_DBdriver", "TestDriver")
	os.Setenv("URLSHORTENER_DSN", "TestDSN:DSN")
	os.Setenv("URLSHORTENER_Mode", "@#4$")

	err := readConfig("wrong.wrong.wrong.file.json")
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}

	if CONFIG.Mode != DefaultMode {
		t.Error("Wrong values set")
	}
}
