package main

import (
	"flag"
	"log"
	"strconv"
	"strings"

	find "./findinlog"
)

func setInit(fc *find.FindConfig, r *find.Report) bool {
	//Setting flags
	report := `Report file name. Parameters: %y year, %M Month (01-12), %d day (01-31), %h hour, %s seconds, %n AWB Number (123-12345678), %p AWB Prefix, %c AWB Code.`
	threadDesc := "Set the quantity of parallel workers. More workers means the process will run faster. WARNING: More Workers also means more memmory consumption. Use it with care."

	zipPath := flag.String("z", "./example.zip", "File Path")
	fileRegEx := flag.String("f", "booking*.log*", "File name in regular expression")
	debugFlag := flag.Bool("v", false, "Set debug")
	threadFlag := flag.Int("t", 10, threadDesc)
	reportName := flag.String("r", "%y-%M-%d_%h-%m-%s_%n-report.txt", report)

	//Reading flag and arguments
	flag.Parse()
	args := flag.Args()

	//Passing parameters
	debug := *debugFlag
	zip := *zipPath
	fileName := *fileRegEx
	threads := *threadFlag

	if debug {
		log.Println("------------------------------------")
		log.Println("Parameters Entered")
		log.Println("------------------------------------")
		log.Println("Input parameters: ")
		log.Println("Zip: ", zip)
		log.Println("File regex: ", fileName)
		log.Println("Number of threads: ", threads)
		log.Println("AWB: ", args)
	}

	if len(args) > 2 {
		log.Println("You have more arguments than the required. You need to enter the AWB like 00 00000000 or 00-00000000. Exiting.")
		return false
	}

	if len(args) < 1 {
		log.Println("You didn't add the AWB. Please enter the AWB Prefix and Code like 00 00000000 or 00-00000000. Exiting.")
		return false
	}

	//Declaring variables
	var prefix, awb int
	var pref, code string
	var err error

	//Argument validation
	switch len(args) {
	case 1:
		sp := strings.Split(args[0], "-")
		if len(sp) != 2 {
			log.Println("There is a problem with the AWB code. Verifiy if follows the format 00-00000000. Example: 45-12345678. Exiting.")
			return false
		}
		pref = sp[0]
		code = sp[1]
	case 2:
		pref = args[0]
		code = args[1]
	}

	prefix, err = strconv.Atoi(pref)

	if err != nil {
		log.Println("There is a problem with the AWB code. Verify if only contain numbers and/or hyphen (-). Exiting.")
		return false
	}

	awb, err = strconv.Atoi(code)

	if err != nil {
		log.Println("There is a problem with the AWB code. Verify if only contain numbers and/or hyphen (-). Exiting.")
		return false
	}

	if debug {
		log.Println("------------------------------------")
		log.Println("Parsing OK")
		log.Println("------------------------------------")
		log.Println("AWB Prefix: ", prefix)
		log.Println("AWB Code: ", awb)
	}

	if debug {
		log.Println("------------------------------------")
	}

	fc.Path = zip
	fc.Format = fileName
	fc.Prefix, fc.Awb = prefix, awb
	fc.Debug = debug
	fc.MaxWorkers = threads

	r.Format = *reportName
	r.Debug = false

	return true
}

func main() {
	var fc find.FindConfig
	var r find.Report

	ok := setInit(&fc, &r)

	if !ok {
		return
	}

	msgs, err := find.SearchInZip(fc)

	if err != nil {
		log.Panic(err)
	}

	r.SetReportFileName(fc.Prefix, fc.Awb)
	r.CreateReport(msgs)
}
