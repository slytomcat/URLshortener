package main

// URLshortener is a microservice to shorten long URLs
// and to handle the redirection by generated short URLs.
//
// See details in README.md
//
// This file contains the configuration reading tools

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
	TokenLength    int                    // token length
	Timeout        int                    // New token creation timeout in ms
	ListenHostPort string                 // host and port to listen on
	DefaultExp     int                    // Default expiration of token (days)
	ShortDomain    string                 // Short domain name for short URL creation
	Mode           int                    // Service mode (see README.md)
}

const (
	// Service modes
	disableRedirect  = 1 << iota // = 1 disable redirect request
	disableShortener             // = 2 disable request for short URL
	disableExpire                // = 4 disable expire request
	disableUI                    // = 8 disable UI generation page

	defaultTokenLength    = 6                // default length of token
	defaultTimeout        = 500              // default timeout of new token creation (ms)
	defaultListenHostPort = "localhost:8080" // default host and port to listen on
	defaultDefaultExp     = 1                // default token expiration
	defaultShortDomain    = "localhost:8080" // default short domain
	defaultMode           = 0                // default service mode

)

// readConfig reads configuration file and also tries to get data from environment variables
func readConfig(cfgFile string) (*Config, error) {
	var err error
	value := ""

	// make the config with default values
	config := Config{
		ConnectOptions: redis.UniversalOptions{}, // read it from file or rfom URLSHORTENER_ConnectOptions env var
		TokenLength:    defaultTokenLength,       // default length of token
		Timeout:        defaultTimeout,           // default timeout of new token creation (ms)
		ListenHostPort: defaultListenHostPort,    // default host and port to listen on
		DefaultExp:     defaultDefaultExp,        // default token expiration
		ShortDomain:    defaultShortDomain,       // default short domain
		Mode:           defaultMode,              // default service mode
	}

	// read config file into buffer
	buf, err := ioutil.ReadFile(cfgFile)
	if err == nil && len(buf) > 0 {
		// parse config file
		err = json.Unmarshal(buf, &config)
	}

	// log config readin/parsing error
	if err != nil {
		log.Printf("Warning: configuration file '%s' reading/parsing error: %s\n", cfgFile, err)
	}

	// try to read config data from evirinment
	if value = os.Getenv("URLSHORTENER_ConnectOptions"); value != "" {
		// parse JSON value of ConnectOptions
		connectOptions := redis.UniversalOptions{}
		err := json.Unmarshal([]byte(value), &connectOptions)
		if err != nil {
			log.Printf("Warning: environments variable URLSHORTENER_ConnectOptions parsing error: %v\n", err)
		} else {
			config.ConnectOptions = connectOptions
		}
	}

	if value = os.Getenv("URLSHORTENER_TokenLength"); value != "" {
		tokenLength, err := strconv.Atoi(value)
		if err != nil {
			log.Printf("Warning: environments variable URLSHORTENER_Timeout conversion error: %v\n", err)
		} else {
			config.TokenLength = tokenLength
		}
	}

	if value = os.Getenv("URLSHORTENER_Timeout"); value != "" {
		timeout, err := strconv.Atoi(value)
		if err != nil {
			log.Printf("Warning: environments variable URLSHORTENER_Timeout conversion error: %v\n", err)
		} else {
			config.Timeout = timeout
		}
	}

	if listenHostPort := os.Getenv("URLSHORTENER_ListenHostPort"); listenHostPort != "" {
		config.ListenHostPort = listenHostPort
	}

	if value = os.Getenv("URLSHORTENER_DefaultExp"); value != "" {
		defaultExp, err := strconv.Atoi(value)
		if err != nil {
			log.Printf("Warning: environments variable URLSHORTENER_DefaultExp conversion error: %v\n", err)
		} else {
			config.DefaultExp = defaultExp
		}
	}

	if shortDomain := os.Getenv("URLSHORTENER_ShortDomain"); shortDomain != "" {
		config.ShortDomain = shortDomain
	}

	if value = os.Getenv("URLSHORTENER_Mode"); value != "" {
		mode, err := strconv.Atoi(value)
		if err != nil {
			log.Printf("Warning: environments variable URLSHORTENER_Mode conversion error: %v\n", err)
		} else {
			config.Mode = mode
		}
	}

	// check mandatory config variable
	if len(config.ConnectOptions.Addrs) == 0 {
		return nil, errors.New("mandatory configuration value ConnectOptions is not set")
	}

	return &config, nil
}
