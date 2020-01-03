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
	if CONFIG.DBdriver == "MySQL" {
		tokenDB, err := TokenDBNewM()
		if err != nil {
			log.Fatalln(err)
		}
		log.Println(ServiceStart(tokenDB))
		// } else {
		// 	if 	if CONFIG.DBdriver == "Redis" {
		// 		tokenDB, err := TokenDBNewR()
		// 		if err != nil {
		// 			log.Fatalln(err)
		// 		}
		// 		log.Println(ServiceStart(tokenDB))
		// 	}
	}

}
