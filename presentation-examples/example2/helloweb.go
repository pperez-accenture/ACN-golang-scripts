package main

import(
	"fmt"
	"log"
	"net/http"
)

//This is an advanced way to do a Hello world, don't you think?
func main() {
	//Configuring root path
    http.HandleFunc("/", func (w http.ResponseWriter, r *http.Request) {
		log.Println("Accesing to the site.")
        fmt.Fprintf(w, "Hello World!")
	})

	//Stating Server. You can change the port, if it's necessary.
	http.ListenAndServe(":8081", nil)
	log.Println("Running site in port 8081.")
}