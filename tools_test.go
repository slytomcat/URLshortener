package main

import (
	"io/ioutil"
	"os"
	"testing"
)

// test config reading from nonexistent file
func Test01Tools00WrongFile(t *testing.T) {
	saveConnectOptions := os.Getenv("URLSHORTENER_ConnectOptions")
	saveCONFIG := CONFIG
	defer func() {
		CONFIG = saveCONFIG
		os.Setenv("URLSHORTENER_ConnectOptions", saveConnectOptions)
	}()
	os.Unsetenv("URLSHORTENER_ConnectOptions")
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

	saveConnectOptions := os.Getenv("URLSHORTENER_ConnectOptions")
	saveCONFIG := CONFIG
	defer func() {
		CONFIG = saveCONFIG
		os.Setenv("URLSHORTENER_ConnectOptions", saveConnectOptions)
	}()
	os.Unsetenv("URLSHORTENER_ConnectOptions")
	CONFIG = Config{}

	err = readConfig(tmpfile.Name())

	if err == nil {
		t.Error("no error for empty file")
	}
	t.Logf("Receved: %v", err)
}

// test it with empty JSON with URLSHORTENER_ConnectOptions unset
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

	saveConnectOptions := os.Getenv("URLSHORTENER_ConnectOptions")
	saveCONFIG := CONFIG
	defer func() {
		CONFIG = saveCONFIG
		os.Setenv("URLSHORTENER_ConnectOptions", saveConnectOptions)
	}()
	os.Unsetenv("URLSHORTENER_ConnectOptions")
	CONFIG = Config{}

	err = readConfig(tmpfile.Name())

	if err == nil {
		t.Error("no error for empty JSON with URLSHORTENER_ConnectOptions unset")
	}
	t.Logf("Receved: %v", err)
}

// test it with empty JSON but with set URLSHORTENER_ConnectOptions
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

	saveConnectOptions := os.Getenv("URLSHORTENER_ConnectOptions")
	saveCONFIG := CONFIG
	defer func() {
		CONFIG = saveCONFIG
		os.Setenv("URLSHORTENER_ConnectOptions", saveConnectOptions)
	}()
	CONFIG = Config{}
	os.Setenv("URLSHORTENER_ConnectOptions", `{"Addrs":["testhost:6379"]}`)

	err = readConfig(tmpfile.Name())

	if err != nil {
		t.Errorf("error for empty JSON with set URLSHORTENER_ConnectOptions: %w", err)
	}
	if !(len(CONFIG.ConnectOptions.Addrs) == 1 && CONFIG.ConnectOptions.Addrs[0] == "testhost:6379") ||
		CONFIG.TokenLength != DefaultTokenLength ||
		CONFIG.Timeout != DefaultTimeout ||
		CONFIG.ListenHostPort != DefaultListenHostPort ||
		CONFIG.DefaultExp != DefaultDefaultExp ||
		CONFIG.ShortDomain != DefaultShortDomain ||
		CONFIG.Mode != DefaultMode {
		t.Error("Wrong default values set")
	}
}

// test full success from example.cnfr.json
func Test01Tools20FullJSON(t *testing.T) {
	saveConnectOptions := os.Getenv("URLSHORTENER_ConnectOptions")
	saveCONFIG := CONFIG
	defer func() {
		CONFIG = saveCONFIG
		os.Setenv("URLSHORTENER_ConnectOptions", saveConnectOptions)
	}()
	CONFIG = Config{}
	os.Unsetenv("URLSHORTENER_ConnectOptions")

	err := readConfig("example.cnfr.json")

	if err != nil {
		t.Errorf("error reading of example.cnfr.json: %v", err)
	}

	if !(len(CONFIG.ConnectOptions.Addrs) == 1 &&
		CONFIG.ConnectOptions.Addrs[0] == "<RedisHost>:6379" &&
		CONFIG.ConnectOptions.DB == 7 &&
		CONFIG.ConnectOptions.Password == "Long long password that is configured for Redis authorization") ||
		CONFIG.TokenLength != 5 ||
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
	saveConnectOptions := os.Getenv("URLSHORTENER_ConnectOptions")
	saveTokenLength := os.Getenv("URLSHORTENER_TokenLength")
	saveTimeout := os.Getenv("URLSHORTENER_Timeout")
	saveHost := os.Getenv("URLSHORTENER_ListenHostPort")
	saveExp := os.Getenv("URLSHORTENER_DefaultExp")
	saveDomain := os.Getenv("URLSHORTENER_ShortDomain")
	saveMode := os.Getenv("URLSHORTENER_Mode")

	return func() {
		CONFIG = saveCONFIG
		os.Setenv("URLSHORTENER_ConnectOptions", saveConnectOptions)
		os.Setenv("URLSHORTENER_TokenLength", saveTokenLength)
		os.Setenv("URLSHORTENER_Timeout", saveTimeout)
		os.Setenv("URLSHORTENER_ListenHostPort", saveHost)
		os.Setenv("URLSHORTENER_DefaultExp", saveExp)
		os.Setenv("URLSHORTENER_ShortDomain", saveDomain)
		os.Setenv("URLSHORTENER_Mode", saveMode)
	}

}

// test full success from environment variables
func Test01Tools30FullEnv(t *testing.T) {
	defer saveEnv()()

	CONFIG = Config{}
	os.Setenv("URLSHORTENER_ConnectOptions", `{"Addrs":["TestHost:6379"]}`)
	os.Setenv("URLSHORTENER_TokenLength", "9")
	os.Setenv("URLSHORTENER_Timeout", "787")
	os.Setenv("URLSHORTENER_ListenHostPort", "testHost:testPort")
	os.Setenv("URLSHORTENER_DefaultExp", "42")
	os.Setenv("URLSHORTENER_ShortDomain", "test.domain")
	os.Setenv("URLSHORTENER_Mode", "66")

	err := readConfig("wrong.wrong.wrong.file.json")
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}

	if !(len(CONFIG.ConnectOptions.Addrs) == 1 && CONFIG.ConnectOptions.Addrs[0] == "TestHost:6379") ||
		CONFIG.TokenLength != 9 ||
		CONFIG.Timeout != 787 ||
		CONFIG.ListenHostPort != "testHost:testPort" ||
		CONFIG.DefaultExp != 42 ||
		CONFIG.ShortDomain != "test.domain" ||
		CONFIG.Mode != 66 {
		t.Error("Wrong values set")
	}
}

func Test01Tools31WrongConnectionOptions(t *testing.T) {
	defer saveEnv()()

	CONFIG = Config{}
	os.Setenv("URLSHORTENER_ConnectOptions", `{"Addrs":6379}`)

	err := readConfig("wrong.wrong.wrong.file.json")
	if err == nil {
		t.Error("no error when expected:")
	}
}

// test wromg enviroment variable URLSHORTENER_TokenLength
func Test01Tools33WrongTokenLength(t *testing.T) {
	defer saveEnv()()

	CONFIG = Config{}
	os.Setenv("URLSHORTENER_ConnectOptions", `{"Addrs":["TestHost:6379"]}`)
	os.Setenv("URLSHORTENER_TokenLength", "%$")

	err := readConfig("wrong.wrong.wrong.file.json")
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}

	if CONFIG.TokenLength != DefaultTokenLength {
		t.Error("Wrong values set")
	}
}

// test wromg enviroment variable URLSHORTENER_Timeout
func Test01Tools35WrongTimeout(t *testing.T) {
	defer saveEnv()()

	CONFIG = Config{}
	os.Setenv("URLSHORTENER_ConnectOptions", `{"Addrs":["TestHost:6379"]}`)
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
func Test01Tools40WrongEnvDefaultExp(t *testing.T) {
	defer saveEnv()()

	CONFIG = Config{}
	os.Setenv("URLSHORTENER_ConnectOptions", `{"Addrs":["TestHost:6379"]}`)
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
func Test01Tools45WrongEnvMode(t *testing.T) {
	defer saveEnv()()

	CONFIG = Config{}
	os.Setenv("URLSHORTENER_ConnectOptions", `{"Addrs":["TestHost:6379"]}`)
	os.Setenv("URLSHORTENER_Mode", "@#4$")

	err := readConfig("wrong.wrong.wrong.file.json")
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}

	if CONFIG.Mode != DefaultMode {
		t.Error("Wrong values set")
	}
}
