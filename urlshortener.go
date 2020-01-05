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
	// create new data base interface
	switch CONFIG.DBdriver {
	case "MySQL":
		tokenDB, err := TokenDBNewM()
		if err != nil {
			log.Fatalln(err)
		}
		log.Println(ServiceStart(tokenDB))
	case "Redis":
		tokenDB, err := TokenDBNewM()
		if err != nil {
			log.Fatalln(err)
		}
		log.Println(ServiceStart(tokenDB))
	default:
		log.Fatalln("error: wrong walue of DBdriver configuration value.")
	}
}
