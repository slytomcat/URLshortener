package main

import "log"

func main() {
	// set logging format
	log.SetFlags(log.Ldate | log.Ltime | log.Lshortfile)

	// get the configuratin variables
	err := readConfig("cnfr.json")
	if err != nil {
		log.Fatalln(err)
	}

	// initialize database connection
	if err = NewTokenDB(); err != nil {
		log.Fatalf("error database interface creation: %v\n", err)
	}

	log.Println(ServiceStart())

}
