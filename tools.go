package main

import (
	"encoding/json"
	"errors"
	"fmt"
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

	// read config file into buffer
	buf, err := ioutil.ReadFile(cfgFile)
	if err == nil {
		// parse config file
		err = json.Unmarshal(buf, &CONFIG)
		if err == nil {
			// check mandatory config variable
			if CONFIG.DSN == "" {
				// try to read it from evirinment
				CONFIG.DSN = os.Getenv("URLSHORTENER_DSN")
				if CONFIG.DSN == "" {
					return errors.New("DSN is not set")
				}
			}
		}
	}
	if err != nil {
		err = fmt.Errorf("configuration file '%s' reading/parsing error: %w", cfgFile, err)
		if CONFIG.DSN == "" {
			return err
		}
		log.Println(err)
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
