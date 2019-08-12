package xlsReader

/////////////////////////
// LDAP Parser
/////////////////////////
// BY Patricio Pérez
// p.perez.bustos@accenture.com
/////////////////////////
// 2.1 Common functions
/////////////////////////

import (
	"os"
)

//This function writes into the file the message and a new line.
func writeLine(f *os.File, msg string){
	f.WriteString(msg)
	f.WriteString("\n")
}