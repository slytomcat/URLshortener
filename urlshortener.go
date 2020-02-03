package main

import (
	"fmt"
	"log"
)

func main() {
	// set logging format
	log.SetFlags(log.Ldate | log.Ltime | log.Lshortfile)
	// log exiting error
	log.Println(doMain())
}

// doMain performs all preparation and starts server
func doMain() error {

	// get the configuratin variables
	err := readConfig("cnfr.json")
	if err != nil {
		return fmt.Errorf("configuration read error: %v", err)
	}

	// initialize database connection
	if err = NewTokenDB(); err != nil {
		return fmt.Errorf("error database interface creation: %v", err)
	}

	// start initial attempts measurement
	go attepmptsMeasurement()

	// run service
	return ServiceStart()

}
