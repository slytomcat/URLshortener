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
	TokenLength    int                    // token length
	Timeout        int                    // New token creation timeout in ms
	ListenHostPort string                 // host and port to listen on
	DefaultExp     int                    // Default expiration of token (days)
	ShortDomain    string                 // Short domain name for short URL creation
	Mode           int                    // Service mode (see README.md)
}

const (
	defaultTokenLength    = 6                // default length of token
	defaultTimeout        = 500              // default timeout of new token creation
	defaultListenHostPort = "localhost:8080" // default host and port to listen on
	defaultDefaultExp     = 1                // default token expiration
	defaultShortDomain    = "localhost:8080" // default short domain
	defaultMode           = 0                // default service mode

	// Service modes
	disableRedirect  = 1 << iota // disable redirect request
	disableShortener             // disable request for short URL
	disableExpire                // disable expire request
)

func parseConOpt(s string) (redis.UniversalOptions, error) {
	conOpt := redis.UniversalOptions{}
	return conOpt, json.Unmarshal([]byte(s), &conOpt)
}

// readConfig reads configuration file and also tries to get data from environment variables
func readConfig(cfgFile string) (*Config, error) {
	var err error
	value := ""
	config := Config{}
	// try to read config data from evirinment

	if value = os.Getenv("URLSHORTENER_ConnectOptions"); value != "" {
		// parse JSON value of ConnectOptions
		config.ConnectOptions, err = parseConOpt(value)
		if err != nil {
			log.Printf("Warning: environments variable URLSHORTENER_ConnectOptions parsing error: %v\n", err)
		}
	}

	if value = os.Getenv("URLSHORTENER_TokenLength"); value != "" {
		config.TokenLength, err = strconv.Atoi(value)
		if err != nil {
			log.Printf("Warning: environments variable URLSHORTENER_Timeout conversion error: %v\n", err)
		}
	}

	if value = os.Getenv("URLSHORTENER_Timeout"); value != "" {
		config.Timeout, err = strconv.Atoi(value)
		if err != nil {
			log.Printf("Warning: environments variable URLSHORTENER_Timeout conversion error: %v\n", err)
		}
	}
	config.ListenHostPort = os.Getenv("URLSHORTENER_ListenHostPort")
	if value = os.Getenv("URLSHORTENER_DefaultExp"); value != "" {
		config.DefaultExp, err = strconv.Atoi(value)
		if err != nil {
			log.Printf("Warning: environments variable URLSHORTENER_DefaultExp conversion error: %v\n", err)
		}
	}
	config.ShortDomain = os.Getenv("URLSHORTENER_ShortDomain")
	if value = os.Getenv("URLSHORTENER_Mode"); value != "" {
		config.Mode, err = strconv.Atoi(value)
		if err != nil {
			log.Printf("Warning: environments variable URLSHORTENER_Mode conversion error: %v\n", err)
		}
	}

	// read config file into buffer
	buf, err := ioutil.ReadFile(cfgFile)
	if err == nil {
		// parse config file
		err = json.Unmarshal(buf, &config)
	}

	// log config reading/parsing error
	if err != nil {
		log.Printf("Warning: configuration file '%s' reading/parsing error: %s\n", cfgFile, err)
	}

	// check mandatory config variable
	if len(config.ConnectOptions.Addrs) == 0 {
		return nil, errors.New("mandatory configuration value ConnectOptions is not set")
	}

	// set default values for optional config variables
	if config.TokenLength == 0 {
		config.TokenLength = defaultTokenLength
	}
	if config.Timeout == 0 {
		config.Timeout = defaultTimeout
	}
	if config.ListenHostPort == "" {
		config.ListenHostPort = defaultListenHostPort
	}
	if config.DefaultExp == 0 {
		config.DefaultExp = defaultDefaultExp
	}
	if config.ShortDomain == "" {
		config.ShortDomain = defaultShortDomain
	}

	// do not set config.Mode as default value is 0

	return &config, nil
}
