package main

import (
	"encoding/json"
	"errors"
	"io/ioutil"
	"log"
	"os"
	"strconv"
)

// Config - configuration structure
type Config struct {
	DBdriver       string // database driver
	DSN            string // MySQL or Redis connection string
	MaxOpenConns   int    `json:",string"` // DB connections pool size for MySQL
	ListenHostPort string // host and port to listen on
	DefaultExp     int    `json:",string"` // Default expiration of token (days)
	ShortDomain    string // Short domain name for short URL creation
	Mode           int    `json:",string"` // Service mode (see README.md)
}

const (
	// DefaultMaxOpenConns - default pool size of DB connections for MySQL
	DefaultMaxOpenConns = 10
	// DefaultListenHostPort - default host and port to listen on
	DefaultListenHostPort = "localhost:8080"
	// DefaultDefaultExp - default token expiration
	DefaultDefaultExp = 1
	// DefaultShortDomain - default short domain
	DefaultShortDomain = "localhost:8080"
	// DefaultMode - default service mode
	DefaultMode = 0
)

// CONFIG - structure with the configuration variables
var CONFIG Config

// readConfig reads config and also tries to get the DB connection string from environment variable
func readConfig(cfgFile string) error {
	var err error
	// try to read config data from evirinment
	CONFIG.DBdriver = os.Getenv("URLSHORTENER_DBdriver")
	CONFIG.DSN = os.Getenv("URLSHORTENER_DSN")
	if cons := os.Getenv("URLSHORTENER_MaxOpenConns"); cons != "" {
		CONFIG.MaxOpenConns, err = strconv.Atoi(cons)
		if err != nil {
			log.Printf("Warning: environments variable URLSHORTENER_MaxOpenConns conversion error: %v\n", err)
		}
	}
	CONFIG.ListenHostPort = os.Getenv("URLSHORTENER_ListenHostPort")
	if exp := os.Getenv("URLSHORTENER_DefaultExp"); exp != "" {
		CONFIG.DefaultExp, err = strconv.Atoi(exp)
		if err != nil {
			log.Printf("Warning: environments variable URLSHORTENER_DefaultExp conversion error: %v\n", err)
		}
	}
	CONFIG.ShortDomain = os.Getenv("URLSHORTENER_ShortDomain")
	if mode := os.Getenv("URLSHORTENER_Mode"); mode != "" {
		CONFIG.Mode, err = strconv.Atoi(os.Getenv("URLSHORTENER_Mode"))
		if err != nil {
			log.Printf("Warning: environments variable URLSHORTENER_Mode conversion error: %v\n", err)
		}
	}

	// read config file into buffer
	buf, err := ioutil.ReadFile(cfgFile)
	if err == nil {
		// parse config file
		err = json.Unmarshal(buf, &CONFIG)
	}

	// log config reading/parsing error
	if err != nil {
		log.Printf("Warning: configuration file '%s' reading/parsing error: %s\n", cfgFile, err)
	}

	// check mandatory config variable DSN
	if CONFIG.DSN == "" || CONFIG.DBdriver == "" {
		return errors.New("Mandatory config values DSN or DBdriver are not set")
	}

	// set default values for optional config variables
	if CONFIG.MaxOpenConns == 0 {
		CONFIG.MaxOpenConns = DefaultMaxOpenConns
	}
	if CONFIG.ListenHostPort == "" {
		CONFIG.ListenHostPort = DefaultListenHostPort
	}
	if CONFIG.DefaultExp == 0 {
		CONFIG.DefaultExp = DefaultDefaultExp
	}
	if CONFIG.ShortDomain == "" {
		CONFIG.ShortDomain = DefaultShortDomain
	}

	// do not set CONFIG.Mode as default value is 0

	return nil
}
