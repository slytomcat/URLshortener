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

	// run service
	return ServiceStart(config, exit)

}
