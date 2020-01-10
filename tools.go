package main

import (
	"encoding/json"
	"errors"
	"io/ioutil"
	"log"
	"os"
	"strconv"

	"github.com/go-redis/redis/v7"
)

// Config - configuration structure
type Config struct {
	ConnectOptions redis.UniversalOptions // Redis connection options
	TokenLength    int                    `json:",int"` // token length
	Timeout        int                    `json:",int"` // New token creation timeout in ms
	ListenHostPort string                 // host and port to listen on
	DefaultExp     int                    `json:",int"` // Default expiration of token (days)
	ShortDomain    string                 // Short domain name for short URL creation
	Mode           int                    `json:",int"` // Service mode (see README.md)
}

const (
	// DefaultTokenLength - default length of token
	DefaultTokenLength = 6
	// DefaultTimeout - default timeout of new token creation
	DefaultTimeout = 500
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

func parseConOpt(s string) redis.UniversalOptions {
	conOpt := redis.UniversalOptions{}
	json.Unmarshal([]byte(s), &conOpt)
	return conOpt
}

// readConfig reads configuration file and also tries to get data from environment variables
func readConfig(cfgFile string) error {
	var err error
	// try to read config data from evirinment

	if value := os.Getenv("URLSHORTENER_ConnectOptions"); value != "" {
		// parse JSON value of ConnectOptions
		CONFIG.ConnectOptions = parseConOpt(value)
	}

	if value := os.Getenv("URLSHORTENER_TokenLength"); value != "" {
		CONFIG.TokenLength, err = strconv.Atoi(value)
		if err != nil {
			log.Printf("Warning: environments variable URLSHORTENER_Timeout conversion error: %v\n", err)
		}
	}

	if value := os.Getenv("URLSHORTENER_Timeout"); value != "" {
		CONFIG.Timeout, err = strconv.Atoi(value)
		if err != nil {
			log.Printf("Warning: environments variable URLSHORTENER_Timeout conversion error: %v\n", err)
		}
	}
	CONFIG.ListenHostPort = os.Getenv("URLSHORTENER_ListenHostPort")
	if value := os.Getenv("URLSHORTENER_DefaultExp"); value != "" {
		CONFIG.DefaultExp, err = strconv.Atoi(value)
		if err != nil {
			log.Printf("Warning: environments variable URLSHORTENER_DefaultExp conversion error: %v\n", err)
		}
	}
	CONFIG.ShortDomain = os.Getenv("URLSHORTENER_ShortDomain")
	if value := os.Getenv("URLSHORTENER_Mode"); value != "" {
		CONFIG.Mode, err = strconv.Atoi(value)
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
	if len(CONFIG.ConnectOptions.Addrs) == 0 {
		return errors.New("mandatory configuration value ConnectOptions is not set")
	}

	// set default values for optional config variables
	if CONFIG.TokenLength == 0 {
		CONFIG.TokenLength = DefaultTokenLength
	}
	if CONFIG.Timeout == 0 {
		CONFIG.Timeout = DefaultTimeout
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
