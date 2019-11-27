package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"os"
)

// Config - configuration structure
type Config struct {
	DSN            string
	MaxOpenConns   int `json:",string"`
	ListenHostPort string
	DefaultExp     int `json:",string"`
	ShortDomain    string
}

// CONFIG - structure with the configuration variables
var CONFIG Config

func readConfig(cfgFile string) error {

	// read config file into buffer
	buf, err := ioutil.ReadFile(cfgFile)
	if err != nil {
		return fmt.Errorf("configuration file '%s' reading error: %w", cfgFile, err)
	}

	// parse config file
	err = json.Unmarshal(buf, &CONFIG)
	if err != nil {
		return fmt.Errorf("configuration file '%s' parsing error: %w", cfgFile, err)
	}

	// check mandatory config variable
	if CONFIG.DSN == "" {
		// try to read it from evirinment
		CONFIG.DSN = os.Getenv("URLSHORTENER_DSN")
		if CONFIG.DSN == "" {
			return errors.New("DSN is not set")
		}
	}

	// set default values for optional config variables
	if CONFIG.MaxOpenConns == 0 {
		CONFIG.MaxOpenConns = 10
	}
	if CONFIG.ListenHostPort == "" {
		CONFIG.ListenHostPort = "localhost:8080"
	}
	if CONFIG.DefaultExp == 0 {
		CONFIG.DefaultExp = 1
	}
	if CONFIG.ShortDomain == "" {
		CONFIG.ShortDomain = "localhost:8080"
	}
	return nil
}
