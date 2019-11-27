package main

import (
	"fmt"
	"log"
	"net/http"
)

var homePage = string(`
<html>
	<body>
	   Home page of URLshortener
	</body>
</html>
`)

var tokenDB *TokenDB

// Home shows home page
func Home(w http.ResponseWriter) {
	fmt.Fprint(w, homePage)
}

func redirect(w http.ResponseWriter, r *http.Request, sToken string) {
	longURL, err := tokenDB.Get(sToken)
	if err != nil {
		w.WriteHeader(404)
		fmt.Printf("URL fror token '%s' was not found\n", sToken)
		return
	}
	fmt.Println("Redirest to ", longURL)
	w.Header()["Location"] = []string{longURL}
	w.WriteHeader(301)
}

func myMUX(w http.ResponseWriter, r *http.Request) {
	path := r.URL.Path
	fmt.Println("Path:", path)
	switch path {
	case "/":
		fmt.Println("Home")
		Home(w)
	case "/token":
		fmt.Println("Request for token")
		// todo
	case "/favicon.ico":
		return
	default:
		fmt.Println("Request for redirect")
		redirect(w, r, path[1:])
	}
}

func main() {
	var err error
	tokenDB, err = TokenDBNew()
	if err != nil {
		panic(err)
	}

	http.HandleFunc("/", myMUX)
	fmt.Println("starting server at :8080")
	log.Fatal(http.ListenAndServe(":8080", nil))
}
