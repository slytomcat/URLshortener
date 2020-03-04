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
	ConfigFile string
)

func init() {
	// prepare command line parameter and usage
	flag.StringVar(&ConfigFile, "config", "./cnfr.json", "`path` to the configuration file")
	flag.Usage = func() {
		fmt.Fprintf(os.Stderr, "Usage:\n\n\t\t"+filepath.Base(os.Args[0])+" [-config=<Path/to/config>]\n\n")
		flag.PrintDefaults()
	}
}

func main() {
	// set logging format
	log.SetFlags(log.Ldate | log.Ltime | log.Lmicroseconds | log.Lshortfile)
	// log exiting error
	log.Println(doMain())
}

// doMain performs all preparation and starts server
func doMain() error {

	// parse command line parameters
	flag.Parse()

	// get the configuratin variables
	err := readConfig(ConfigFile)
	if err != nil {
		return fmt.Errorf("configuration read error: %w", err)
	}

	// initialize database connection
	if err = NewTokenDB(); err != nil {
		return fmt.Errorf("error database interface creation: %w", err)
	}

	// run service
	return ServiceStart()

}
