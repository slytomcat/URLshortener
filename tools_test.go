package main

import "testing"

import "io/ioutil"

import "os"

func Test01Tools00WrongFile(t *testing.T) {

	err := readConfig("wrong.wrong.wrong.file.json")

	if err == nil {
		t.Error("no error for wrong filename")
	}
	t.Logf("Receved: %v", err)
}

func Test01Tools05EmptyFile(t *testing.T) {
	tmpfile, err := ioutil.TempFile(os.TempDir(), "testing*.json")
	if err != nil {
		t.Errorf("temp file creation error: %w", err)
	}
	defer os.Remove(tmpfile.Name())

	if err := tmpfile.Close(); err != nil {
		t.Errorf("temp file closing error: %w", err)
	}
	err = readConfig(tmpfile.Name())

	if err == nil {
		t.Error("no error for empty file")
	}
	t.Logf("Receved: %v", err)
}

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
	CONFIG = Config{
		DSN:            "",
		MaxOpenConns:   0,
		ListenHostPort: "",
		DefaultExp:     0,
		ShortDomain:    "",
	}

	err = readConfig(tmpfile.Name())

	if err == nil {
		t.Error("no error for empty JSON with with URLSHORTENER_DSN unset")
	}
	t.Logf("Receved: %v", err)
}

func Test01Tools15EmptyJSON_(t *testing.T) {
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
	saveDSN := os.Getenv("URLSHORTENER_DSN")
	defer func() {
		CONFIG = saveCONFIG
		os.Setenv("URLSHORTENER_DSN", saveDSN)
	}()

	CONFIG = Config{
		DSN:            "",
		MaxOpenConns:   0,
		ListenHostPort: "",
		DefaultExp:     0,
		ShortDomain:    "",
	}

	os.Setenv("URLSHORTENER_DSN", "testDSNvalue")

	err = readConfig(tmpfile.Name())

	if err != nil {
		t.Errorf("error for empty JSON with URLSHORTENER_DSN set ")
	}
	if CONFIG.DSN != "testDSNvalue" ||
		CONFIG.MaxOpenConns != 10 ||
		CONFIG.ListenHostPort != "localhost:8080" ||
		CONFIG.DefaultExp != 1 ||
		CONFIG.ShortDomain != "localhost:8080" {
		t.Error("Wrong default values set")
	}
}

func Test01Tools20FullJSON(t *testing.T) {
	saveCONFIG := CONFIG
	saveDSN := os.Getenv("URLSHORTENER_DSN")
	defer func() {
		CONFIG = saveCONFIG
		os.Setenv("URLSHORTENER_DSN", saveDSN)
	}()

	CONFIG = Config{
		DSN:            "",
		MaxOpenConns:   0,
		ListenHostPort: "",
		DefaultExp:     0,
		ShortDomain:    "",
	}
	os.Unsetenv("URLSHORTENER_DSN")

	err := readConfig("example.cnf.json")

	if err != nil {
		t.Errorf("error reading of example.cnf.json: %v", err)
	}
	if CONFIG.DSN != "shortener:<password>@<protocol>(<host>:<port>)/shortener_DB" ||
		CONFIG.MaxOpenConns != 33 ||
		CONFIG.ListenHostPort != "0.0.0.0:80" ||
		CONFIG.DefaultExp != 30 ||
		CONFIG.ShortDomain != "<shortDomain>" {
		t.Error("Wrong values set")
	}
}
