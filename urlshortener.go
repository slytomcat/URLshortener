package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"path/filepath"
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
	// log exiting error
	log.Println(doMain(configFile))
}

// doMain performs all preparation and starts server
func doMain(configPath string) error {

	// parse command line parameters
	flag.Parse()

	// get the configuratin variables
	err := readConfig(configPath)
	if err != nil {
		return fmt.Errorf("configuration read error: %w", err)
	}

	// initialize database connection
	tokenDB, err := NewTokenDB(CONFIG.ConnectOptions, CONFIG.Timeout, CONFIG.TokenLength)
	if err != nil {
		return fmt.Errorf("error database interface creation: %w", err)
	}

	// run service
	return ServiceStart(tokenDB)

}
