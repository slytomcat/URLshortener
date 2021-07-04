package main

// URLshortener is a microservice to shorten long URLs
// and to handle the redirection by generated short URLs.
//
// See details in README.md
//
// This file contains the configuration reading tools

import (
	"fmt"

	"github.com/kelseyhightower/envconfig"
)

// Config - configuration structure
type Config struct {
	RedisAddrs     []string `required:"true"`          // Redis connection adresses
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
	disableRedirect  uint = 1 << iota // = 1 disable redirect request
	disableShortener                  // = 2 disable request for short URL
	disableExpire                     // = 4 disable expire request
	disableUI                         // = 8 disable UI generation page
)

// readConfig reads configuration file and also tries to get data from environment variables
func readConfig() (*Config, error) {

	config := Config{}
	err := envconfig.Process("URLSHORTENER", &config)
	if err != nil {
		return nil, fmt.Errorf("config error: %w", err)
	}

	if config.Mode > disableUI {
		return nil, fmt.Errorf("config error: wrong mode value: %x", config.Mode)
	}
	return &config, nil
}
