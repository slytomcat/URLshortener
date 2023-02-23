package main

// URLshortener is a microservice to shorten long URLs
// and to handle the redirection by generated short URLs.
//
// See details in README.md
//
// This file contains the main routine

import (
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"
)

var (
	version string = "v.local_build"
)

func main() {
	if len(os.Args) == 2 && os.Args[1] == "-v" {
		fmt.Printf("URLshortener %s\n", version)
		os.Exit(0)
	}
	// log the version
	log.Printf("URLshortener %s", version)
	// set logging format
	log.SetFlags(log.Ldate | log.Ltime | log.Lmicroseconds | log.Lshortfile)

	// get exiting error
	err := doMain()
	if err != http.ErrServerClosed {
		panic(err)
	} else {
		log.Println(err)
	}
}

// doMain performs all preparation and starts server
func doMain() error {
	// get the configuratin variables
	config, err := readConfig()
	if err != nil {
		return fmt.Errorf("configuration read error: %w", err)
	}

	// initialize database connection
	tokenDB, err := NewTokenDB(config.RedisAddrs, config.RedisPassword)
	if err != nil {
		return fmt.Errorf("database interface creation error: %w", err)
	}
	defer func() {
		// close DB connection
		err = tokenDB.Close()
		if err != nil {
			log.Printf("DB connection close error: %v", err)
		}
	}()

	return stratService(config, tokenDB)
}

func stratService(config *Config, tokenDB TokenDB) error {

	// make service handler
	handler := NewHandler(config, tokenDB, NewShortToken(config.TokenLength))
	handlerErr := make(chan error, 1)
	// start service
	go func() {
		handlerErr <- handler.start()
	}()

	// wait for server start
	time.Sleep(300 * time.Millisecond)
	if err := handler.healthCheck(); err != nil {
		fmt.Printf("initial health-check failed: %v\nexiting...\n", err)
	} else {
		log.Println("initial health-check successfuly passed")
		c := make(chan os.Signal, 1)
		signal.Notify(c, syscall.SIGINT, syscall.SIGTERM)
		// sleep until a signal is received.
		<-c
	}
	// Close service
	handler.stop()

	return <-handlerErr
}
