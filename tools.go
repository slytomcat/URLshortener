package main

import (
	"encoding/json"
	"errors"
	"io/ioutil"
	"log"
	"os"
)

// Config - configuration structure
type Config struct {
	DSN            string // MySQL connection string
	MaxOpenConns   int    `json:",string"` // DB connections pool size
	ListenHostPort string // host and port to listen on
	DefaultExp     int    `json:",string"` // Default expiration of token (days)
	ShortDomain    string // Short domain name for short URL creation
	Mode           int    `json:",string"` // Service mode (see README.md)
}

// CONFIG - structure with the configuration variables
var CONFIG Config

// readConfig reads config and also tries to get the DB connection string from environment variable
func readConfig(cfgFile string) error {

	// try to read DSN from evirinment
	CONFIG.DSN = os.Getenv("URLSHORTENER_DSN")

	// read config file into buffer
	buf, err := ioutil.ReadFile(cfgFile)
	if err == nil {
		// parse config file
		err = json.Unmarshal(buf, &CONFIG)
	}

	// log config readin/parsing error
	if err != nil {
		log.Printf("Warning: configuration file '%s' reading/parsing error: %s\n", cfgFile, err)
	}

	// check mandatory config variable DSN
	if CONFIG.DSN == "" {
		return errors.New("DSN is not set")
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

	// do not set CONFIG.Mode as default value is 0

	return nil
}
