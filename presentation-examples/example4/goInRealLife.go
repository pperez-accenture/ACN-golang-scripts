package main

import (
	"log"
	"flag" //Flag, in my opinion are better than only arguments.

	"encoding/json"
	"io/ioutil"

	//Here, the code is using a local package.
	//Note that the package can be renamed.
	ft "./fetch"
)

// Config is a struct. Consider struct as an equivalent of a class
// (although Golang is procedural and not an object oriented envionment).
type Config struct {
	FileName string
}

// SetFileName set the value of the FileName.
// You can also define a function for the struct
// The parenthesis behind the function indicates the owner of the function.
// Like C, you can use the variable as reference (pointers). 
// In this case, you can pass the struct variable as reference of self.
// As a convention, all the public functions needs to start with a Uppercase.
func (cf *Config) SetFileName(fnm string) {
	cf.FileName = fnm
}

// Conf is a Global Variable.
// Everyone in the package can use this variable.
// Also, as a convention, all the public variables needs to start with Uppercase.
var Conf Config

//You can also initialize the variables. Note that Golang is Case Sensitive.
var (
	FileName = "output.json"
	URLTest = "https://www.mobygames.com/stats/this-day/0621"
)

//The functionality of flag is very simple. First, initialize the identificators,
//Then, parse the data.
//Also, you can set one o more return types
func readFlags(fnm *string) (bool, error){
	//Initiaize
	flag.StringVar(fnm, "o", FileName, "set the name of the file")
	//Read
	flag.Parse()
	//Error can be returned as null.
	return true, nil
}

//The init function is a great way to verify everything before start the execution.
//The order of execution is init() -> main().
func init() {
	var fnm string

	//This is the way to initialize a struct.
	Conf = Config{}

	//Pass the variable as a pointer.
	readFlags(&fnm)
	
	//Call the function defined in config.
	Conf.SetFileName(fnm)
}

func main() {
	//After the init, you can check if the value is set.
	//With %v you can show the value of any variable. But with %#v you can inspect the structure of the variable.
	log.Printf("Config: %#v", Conf)

	//Now, here the fetch is called

	//This is the natural way to assign values to a struct
	f := ft.Fetcher{URLPath: URLTest}

	data := f.Process()

	//Write the data
	//As you can see, if you don't need the data for a variable,
	//You can use _ (underscore). This is like dropping the value to null.
	file, _ := json.MarshalIndent(data, "", " ")
	_ = ioutil.WriteFile(Conf.FileName, file, 0644)

}
