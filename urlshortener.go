package main

import "log"

func main() {
	// set logging format
	log.SetFlags(log.Ldate | log.Ltime | log.Lshortfile)

	// get the configuratin variables
	err := readConfig("cnf.json")
	if err != nil {
		log.Fatalln(err)
	}

	if err = NewTokenDB(); err != nil {
		log.Fatalf("error database iteface creation: %v\n", err)
	}

	log.Println(ServiceStart())

}
