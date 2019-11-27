package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
)

// Config - configuration structure
type Config struct {
	DSN            string
	ListenHostPort string
	DefaultExp     int `json:",string"`
	ShortDomain    string
}

// CONFIG - structure with the configuration variables
var CONFIG Config

func readConfig() error {
	cfgFile := ".cnf.json"

	// open config file
	cf, err := os.Open(cfgFile)
	if err != nil {
		return fmt.Errorf("configuration file '%s' can't be read: %w", cfgFile, err)
	}
	defer cf.Close()

	// read config file into buffer
	buf := make([]byte, 1024)
	n, err := cf.Read(buf)
	if err != nil {
		return fmt.Errorf("configuration file '%s' reading error: %w", cfgFile, err)
	}

	// parse config file
	err = json.Unmarshal(buf[:n], &CONFIG)
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
