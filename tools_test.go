package main

import (
	"io/ioutil"
	"os"
	"testing"
)

// saveEnv stores environment variables, os.Args and CONFIG and returns
// function that restores saved items to original values
func saveEnv() func() {
	saveArgs := os.Args
	saveConnectOptions := os.Getenv("URLSHORTENER_ConnectOptions")
	saveTokenLength := os.Getenv("URLSHORTENER_TokenLength")
	saveTimeout := os.Getenv("URLSHORTENER_Timeout")
	saveHost := os.Getenv("URLSHORTENER_ListenHostPort")
	saveExp := os.Getenv("URLSHORTENER_DefaultExp")
	saveDomain := os.Getenv("URLSHORTENER_ShortDomain")
	saveMode := os.Getenv("URLSHORTENER_Mode")

	// returned func restores all stored items to original values
	return func() {
		os.Args = saveArgs
		os.Setenv("URLSHORTENER_ConnectOptions", saveConnectOptions)
		os.Setenv("URLSHORTENER_TokenLength", saveTokenLength)
		os.Setenv("URLSHORTENER_Timeout", saveTimeout)
		os.Setenv("URLSHORTENER_ListenHostPort", saveHost)
		os.Setenv("URLSHORTENER_DefaultExp", saveExp)
		os.Setenv("URLSHORTENER_ShortDomain", saveDomain)
		os.Setenv("URLSHORTENER_Mode", saveMode)
	}
}

// test config reading from not existing file
func Test01Tools00WrongFile(t *testing.T) {

	defer saveEnv()()

	os.Unsetenv("URLSHORTENER_ConnectOptions")
	// &&& CONFIG = Config{}

	_, err := readConfig("wrong.wrong.wrong.file.json")

	if err == nil {
		t.Error("no error for wrong filename")
	}
	t.Logf("Receved: %v", err)
}

// test with empty file and unset URLSHORTENER_DSN
func Test01Tools05EmptyFile(t *testing.T) {
	tmpfile, err := ioutil.TempFile(os.TempDir(), "testing*.json")
	if err != nil {
		t.Errorf("temp file creation error: %w", err)
	}
	defer os.Remove(tmpfile.Name())

	if err := tmpfile.Close(); err != nil {
		t.Errorf("temp file closing error: %w", err)
	}

	defer saveEnv()()

	os.Unsetenv("URLSHORTENER_ConnectOptions")
	// &&& CONFIG = Config{}

	_, err = readConfig(tmpfile.Name())

	if err == nil {
		t.Error("no error for empty file")
	}
	t.Logf("Receved: %v", err)
}

// test with empty JSON file and unset URLSHORTENER_ConnectOptions
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

	defer saveEnv()()

	os.Unsetenv("URLSHORTENER_ConnectOptions")
	// &&& CONFIG = Config{}

	_, err = readConfig(tmpfile.Name())

	if err == nil {
		t.Error("no error for empty JSON with URLSHORTENER_ConnectOptions unset")
	}
	t.Logf("Receved: %v", err)
}

// test with empty JSON file but with set URLSHORTENER_ConnectOptions
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

	defer saveEnv()()

	// &&& CONFIG = Config{}
	os.Setenv("URLSHORTENER_ConnectOptions", `{"Addrs":["testhost:6379"]}`)

	config, err := readConfig(tmpfile.Name())

	if err != nil {
		t.Errorf("error for empty JSON with set URLSHORTENER_ConnectOptions: %w", err)
	}
	if !(len(config.ConnectOptions.Addrs) == 1 && config.ConnectOptions.Addrs[0] == "testhost:6379") ||
		config.TokenLength != defaultTokenLength ||
		config.Timeout != defaultTimeout ||
		config.ListenHostPort != defaultListenHostPort ||
		config.DefaultExp != defaultDefaultExp ||
		config.ShortDomain != defaultShortDomain ||
		config.Mode != defaultMode {
		t.Error("Wrong default values set")
	}
}

// test full success from example.cnfr.json
func Test01Tools20FullJSON(t *testing.T) {

	defer saveEnv()()

	os.Unsetenv("URLSHORTENER_ConnectOptions")

	config, err := readConfig("example.cnfr.json")

	if err != nil {
		t.Errorf("error reading of example.cnfr.json: %v", err)
	}

	if !(len(config.ConnectOptions.Addrs) == 1 &&
		config.ConnectOptions.Addrs[0] == "<RedisHost>:6379" &&
		config.ConnectOptions.DB == 7 &&
		config.ConnectOptions.Password == "Long long password that is configured for Redis authorization") ||
		config.TokenLength != 5 ||
		config.Timeout != 777 ||
		config.ListenHostPort != "0.0.0.0:80" ||
		config.DefaultExp != 30 ||
		config.ShortDomain != "<shortDomain>" ||
		config.Mode != disableExpire {
		t.Error("Wrong values set")
	}
}

// test full success from environment variables
func Test01Tools30FullEnv(t *testing.T) {

	defer saveEnv()()

	// ???CONFIG = Config{}

	os.Setenv("URLSHORTENER_ConnectOptions", `{"Addrs":["TestHost:6379"]}`)
	os.Setenv("URLSHORTENER_TokenLength", "9")
	os.Setenv("URLSHORTENER_Timeout", "787")
	os.Setenv("URLSHORTENER_ListenHostPort", "testHost:testPort")
	os.Setenv("URLSHORTENER_DefaultExp", "42")
	os.Setenv("URLSHORTENER_ShortDomain", "test.domain")
	os.Setenv("URLSHORTENER_Mode", "7")

	config, err := readConfig("wrong.wrong.wrong.file.json")
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}

	if !(len(config.ConnectOptions.Addrs) == 1 && config.ConnectOptions.Addrs[0] == "TestHost:6379") ||
		config.TokenLength != 9 ||
		config.Timeout != 787 ||
		config.ListenHostPort != "testHost:testPort" ||
		config.DefaultExp != 42 ||
		config.ShortDomain != "test.domain" ||
		config.Mode != disableRedirect|disableShortener|disableExpire {
		t.Error("Wrong values set")
	}
}

// test wrong connection options
func Test01Tools31WrongConnectionOptions(t *testing.T) {

	defer saveEnv()()

	os.Setenv("URLSHORTENER_ConnectOptions", `{"Addrs":6379}`)

	_, err := readConfig("wrong.wrong.wrong.file.json")
	if err == nil {
		t.Error("no error when expected:")
	}
}

// test wromg enviroment variable URLSHORTENER_TokenLength
func Test01Tools33WrongTokenLength(t *testing.T) {

	defer saveEnv()()

	os.Setenv("URLSHORTENER_ConnectOptions", `{"Addrs":["TestHost:6379"]}`)
	os.Setenv("URLSHORTENER_TokenLength", "%$")

	config, err := readConfig("wrong.wrong.wrong.file.json")
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}

	if config.TokenLength != defaultTokenLength {
		t.Error("Wrong values set")
	}
}

// test wromg enviroment variable URLSHORTENER_Timeout
func Test01Tools35WrongTimeout(t *testing.T) {
	defer saveEnv()()

	os.Setenv("URLSHORTENER_ConnectOptions", `{"Addrs":["TestHost:6379"]}`)
	os.Setenv("URLSHORTENER_Timeout", "@#2$")
	config, err := readConfig("wrong.wrong.wrong.file.json")
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}

	if config.Timeout != defaultTimeout {
		t.Error("Wrong values set")
	}
}

// test wromg enviroment variable URLSHORTENER_DefaultExp
func Test01Tools40WrongEnvDefaultExp(t *testing.T) {

	defer saveEnv()()

	os.Setenv("URLSHORTENER_ConnectOptions", `{"Addrs":["TestHost:6379"]}`)
	os.Setenv("URLSHORTENER_DefaultExp", "@#2$")

	config, err := readConfig("wrong.wrong.wrong.file.json")
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}

	if config.DefaultExp != defaultDefaultExp {
		t.Error("Wrong values set")
	}
}

// test wromg enviroment variable URLSHORTENER_Mode
func Test01Tools45WrongEnvMode(t *testing.T) {

	defer saveEnv()()

	os.Setenv("URLSHORTENER_ConnectOptions", `{"Addrs":["TestHost:6379"]}`)
	os.Setenv("URLSHORTENER_Mode", "@#4$")

	config, err := readConfig("wrong.wrong.wrong.file.json")
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}

	if config.Mode != defaultMode {
		t.Error("Wrong values set")
	}
}
