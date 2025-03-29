package main

// URLshortener is a microservice to shorten long URLs
// and to handle the redirection by generated short URLs.
//
// See details in README.md
//
// This file contains the configuration reading tools

import (
	"cmp"
	"fmt"
	"os"
	"strconv"
	"strings"
)

// Config - configuration structure
type Config struct {
	RedisAddrs     []string `required:"true"`          // Redis connection addresses
	RedisPassword  string   `default:""`               // Redis connection password
	TokenLength    int      `default:"6"`              // token length
	Timeout        int      `default:"500"`            // New token creation timeout in ms
	ListenHostPort string   `default:"localhost:8080"` // host and port to listen on
	DefaultExp     int      `default:"1"`              // Default expiration of token (days)
	ShortDomain    string   `default:"localhost:8080"` // Short domain name for short URL creation
	Mode           uint     `default:"0"`              // Service mode (see README.md)
}

const (
	// Service modes
	disableRedirect    uint = 1 << iota // = 1 disable redirect request
	disableShortener                    // = 2 disable request for short URL
	disableExpire                       // = 4 disable expire request
	disableUI                           // = 8 disable UI generation page
	disableLengthCheck                  // = 16 disable token length check (during redirect)
	incorrectOption
	TokenLength
	envRedisAddrs         = "URLSHORTENER_REDISADDRS"
	envRedisPassword      = "URLSHORTENER_REDISPASSWORD"
	envTokenLength        = "URLSHORTENER_TOKENLENGTH"
	envTimeout            = "URLSHORTENER_TIMEOUT"
	envListenHostPort     = "URLSHORTENER_LISTENHOSTPORT"
	envDefaultExp         = "URLSHORTENER_DEFAULTEXP"
	envShortDomain        = "URLSHORTENER_SHORTDOMAIN"
	envMode               = "URLSHORTENER_MODE"
	defaultTokenLength    = "6"
	defaultTimeout        = "500"
	defaultListenHostPort = "localhost:8080"
	defaultDefaultExp     = "1"
	defaultShortDomain    = "localhost:8080"
	defaultMode           = "0"
)

// readConfig reads configuration from environment variables
func readConfig() (*Config, error) {
	addrs := []string{}
	for s := range strings.SplitSeq(os.Getenv(envRedisAddrs), ",") {
		s = strings.Trim(s, " \t")
		if len(s) > 0 {
			addrs = append(addrs, s)
		}
	}
	if len(addrs) == 0 {
		return nil, fmt.Errorf("config error: wrong or missed value of %s", envRedisAddrs)
	}
	length, err := strconv.ParseUint(cmp.Or(os.Getenv(envTokenLength), defaultTokenLength), 10, 64)
	if err != nil {
		return nil, fmt.Errorf("config error: wrong value of %s: %w", envTokenLength, err)
	}
	timeout, err := strconv.ParseUint(cmp.Or(os.Getenv(envTimeout), defaultTimeout), 10, 64)
	if err != nil {
		return nil, fmt.Errorf("config error: wrong value of %s: %w", envTimeout, err)
	}
	exp, err := strconv.ParseUint(cmp.Or(os.Getenv(envDefaultExp), defaultDefaultExp), 10, 64)
	if err != nil {
		return nil, fmt.Errorf("config error: wrong value of %s: %w", envDefaultExp, err)
	}
	mode, err := strconv.ParseUint(cmp.Or(os.Getenv(envMode), defaultMode), 10, 64)
	if err != nil {
		return nil, fmt.Errorf("config error: wrong value of %s: %w", envMode, err)
	}
	if uint(mode) >= incorrectOption {
		return nil, fmt.Errorf("config error: wrong value of %s: %xH (%d)", envMode, mode, mode)
	}

	return &Config{
		RedisAddrs:     addrs,
		RedisPassword:  os.Getenv(envRedisPassword),
		TokenLength:    int(length),
		Timeout:        int(timeout),
		ListenHostPort: cmp.Or(os.Getenv(envListenHostPort), defaultListenHostPort),
		DefaultExp:     int(exp),
		ShortDomain:    cmp.Or(os.Getenv(envShortDomain), defaultShortDomain),
		Mode:           uint(mode),
	}, nil
}
