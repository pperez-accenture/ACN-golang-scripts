package main

import (
	"log"

	"./logbackup"
)

func main() {
	//Init the test data
	data, err := logbackup.ReadConfig("config.json")

	if err != nil {
		log.Printf("There is an error trying to load the file config.json")
		log.Panic(err)
	} else {
		//First, fetch the logs from the server
		data.BackupFromServer()
		//Then, unzip the data
		data.Unzip()
		//Finally, leave everything in a zip and remove the rest of the data.
		data.ZipEverythingAndClear()
	}
}
