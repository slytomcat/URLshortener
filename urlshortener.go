package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"os/signal"
	"path/filepath"
	"syscall"
	"time"
)

var (
	// ConfigFile - is the path to the configuration file
	configFile string
)

func init() {
	// prepare command line parameter and usage
	flag.StringVar(&configFile, "config", "./cnfr.json", "`path` to the configuration file")
	flag.Usage = func() {
		fmt.Fprintf(os.Stderr, "Usage:\n\n\t\t"+filepath.Base(os.Args[0])+" [-config=<Path/to/config>]\n\n")
		flag.PrintDefaults()
	}
}

func main() {
	// set logging format
	log.SetFlags(log.Ldate | log.Ltime | log.Lmicroseconds | log.Lshortfile)
	exit := make(chan bool)
	// log exiting error
	log.Println(doMain(configFile, exit))
	// wait for service exit
	<-exit
}

// doMain performs all preparation and starts server
func doMain(configPath string, exit chan bool) error {

	// parse command line parameters
	flag.Parse()

	// get the configuratin variables
	config, err := readConfig(configPath)
	if err != nil {
		return fmt.Errorf("configuration read error: %w", err)
	}

	// initialize database connection
	tokenDB, err := NewTokenDB(config.ConnectOptions)
	if err != nil {
		return fmt.Errorf("database interface creation error: %w", err)
	}

	// get service handler
	handler := NewHandler(config, tokenDB, NewShortToken(config.TokenLength), exit)

	// register the SIGINT and SIGTERM handler
	c := make(chan os.Signal, 1)
	signal.Notify(c, syscall.SIGINT, syscall.SIGTERM)
	// start signal handler
	go func() {
		// sleep until a signal is received.
		<-c
		// Close service
		handler.Stop()
	}()

	// start health checker
	go func() {
		// wait for server start
		<-time.After(300 * time.Millisecond)
		// and perform health-check
		if err := handler.HealthCheck(); err != nil {
			log.Printf("initial health-check failed: %v", err)
			// Close service
			handler.Stop()
			return
		}
		log.Println("initial health-check successfuly passed")
	}()

	// run server
	return handler.Start()

}
