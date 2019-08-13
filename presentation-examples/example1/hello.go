//A golang file is composed of three parts:
//1) Package name
//2) Package imports
//3) File content.

//1) Package name: This is the name of the location of the script(s).
package main

//2) Package imports. These are the collection(s) of scripts that you are
//going to use in this file.
import (
	"fmt"
)

//3) File content.
//As default, the main() function is the default one to execute.
func main(){
	//Basic Hello World example
	fmt.Println("Hello World")
}