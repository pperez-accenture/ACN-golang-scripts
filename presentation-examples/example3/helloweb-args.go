package main

import(
	"fmt"
	"log"
	"net/http"

	"os"
	"strconv"
	"math"
)

//This is a personalized function. Note the syntax for the variables if like "var type"
func startServer(port int){
	//The same initialization of the previous example
    http.HandleFunc("/", func (w http.ResponseWriter, r *http.Request) {
		log.Println("Accesing to the site.")
        fmt.Fprintf(w, "Welcome again to my website!")
	})

	//Sprintf is the equivalent of format.
	log.Println("Running site in port", port, ".")
	http.ListenAndServe(fmt.Sprintf(":%d", port), nil)
}

func main() {
	//Initialization examples.
	//This is the default initialization
	var lstArgs []string
	lstArgs = os.Args

	//Arg0 is the executable path
	log.Printf("List of arguments: %v ", lstArgs)

	//There are more arguments?
	//len is a basic function of golang, so there is no need of a package.
	if (len(lstArgs) < 2) {
		panic("No arguments found")
	}
	
	//But you can also initilize the variable without knowing the type.
	arg1 := os.Args[1]

	//Since the argument is a port, you need a verification.
	//Atoi is great for that, since it will check if the string is a valid number or not.
	//Also, note that the function return 2 arguments, the result and possible error.
	port, err := strconv.Atoi(arg1)
	//The err is an error type
	//The way to initialize is: var err error
	
	// There in no need of forced parenthesis.
	if err != nil {
		panic(err)
	} else if !(port > 0 && port < (int) (math.Pow(2, 16) - 1)) { //Check the conversion from float to int.
		//Golang is a little grammar nazi.
		//Only when you open a bracket or comma, Golang admits newline.
		//If you make if var {} else 
		//{}, golang is going to throw a compile error
		panic("Invalid port")
	}

	startServer(port)
} 