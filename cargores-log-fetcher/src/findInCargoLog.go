/*
***********************************
* findInCargoLog.go
***********************************
* Author: Patricio Pérez
* Mail: p.perez.bustos@accenture.com
* Date: 2018-06-20
* Desc: The script will navigate through the logs, check the AWB contained and extract
* the XML. Then, the script generates a chronological report with all the XML's found.
*/


package main

import (
	"os"
	//"log"
	"fmt"
	"flag"
	"sort"
	"time"
	//"bytes"
	"regexp"
	"strconv"
	"strings"
	"io/ioutil"
	"encoding/xml"
	"unicode/utf8"
	fp "path/filepath"
)

type AWB struct{
	Prefix string
	Number string
}

type Field struct{
	Name string `xml:"name,attr"`
	Type string `xml:"type,attr"`
	Value string `xml:"String"`
}

type InnerRQTag struct{
	AwbId AWB `xml:"AwbId"`
	Inner []InnerRQTag `xml:",any"`
}

type InnerRSTag struct{
	Inner []InnerRSTag `xml:",any"`
	ResField []Field `xml:"Field"`
}

type Message struct{
	//XML Metadata
	AWBResponse []InnerRSTag `xml:"ResponseContent"`
	AWBContent []InnerRQTag `xml:"RequestContent"`
	Date string `xml:"MessageOrigin>GMTDateTime"`
	MessageType string `xml:"messageType,attr"`
	ServiceType string `xml:"serviceType,attr"`
	RequestType string `xml:"crsRequestType,attr"`
	
	//Added parameters
	DateParsed time.Time
	FileName string
	XML string
}

type Log struct {
	Date string
	Message string
}

var emptyField []Field
var genTimeLayout = "2006-01-02 15:04:05"

//Recursive method to get the AWB.
func (t InnerRQTag) getAWBs() []AWB{
	var lstAwb []AWB
	
	//Verify if the node has Awb Defined.
	if t.AwbId != (AWB{}) {
		lstAwb = append(lstAwb, t.AwbId)
	}
	
	//If it's not, go one level lower to verify.
	for _,tag := range t.Inner {
		awb := tag.getAWBs()
		for _,a := range awb {
			if(a != (AWB{})) {
				lstAwb = append(lstAwb, a)
			}
		}
	}
	
	//Otherwise, return an empty AWB{}
	return lstAwb
}

//Recursive method to get the Fields.
func (t InnerRSTag) getFields() []Field{
	var lstFields []Field
	
	//Verify if the node has Field Defined.
	//if t.ResField != nil {
	//	lstFields = append(lstFields, t.ResField)
	//}
	
	for _,f := range t.ResField {
		lstFields = append(lstFields, f)
	}
	
	//If it's not, go one level lower to verify.
	for _,tag := range t.Inner {
		field := tag.getFields()
		for _,f := range field {
			if(f != (Field{})) {
				lstFields = append(lstFields, f)
			}
		}
	}
	
	//Otherwise, return an empty ResField{}
	return lstFields
}

//Get the AWB's
func (m *Message) GetAWBsFromRequest() []AWB{
	var lstAwb []AWB
	
	for _,tag := range m.AWBContent {
		awb := tag.getAWBs()
		for _,a := range awb {
			if(a != (AWB{})) {
				lstAwb = append(lstAwb, a)
			}
		}
	}
	return lstAwb
}

//Get the Fields Tags
func (m *Message) GetFieldsFromResponse() []Field{
	var lstFields []Field
	
	for _,tag := range m.AWBResponse {
		fields := tag.getFields()
		for _,f := range fields {
			if(f != (Field{})) {
				lstFields = append(lstFields, f)
			}
		}
	}
	return lstFields
}

//Get Parsed AWB's.
func (m *Message) GetParsedAWBsFromRequest() string{
	var lstAwb []AWB
	var str []string
	
	lstAwb = m.GetAWBsFromRequest()
	
	if len(lstAwb) == 0 {
		return ""
	}
	
	for _,awb := range lstAwb {
		str = append(str, fmt.Sprintf("%s-%s", awb.Prefix, awb.Number))
	}
	
	return strings.Join(str, ", ")
}

//Check if the AWB Request List contains the AWB number.
func (m *Message) ContainsAwb(prefix, number int) bool{
	var lstAwb []AWB
	
	lstAwb = m.GetAWBsFromRequest()
	
	for _, awb := range lstAwb {
		if 	awb.Prefix == strconv.Itoa(prefix) && 
			awb.Number == strconv.Itoa(number) {
			return true
		}
	}
	
	return false
}

func (m *Message) GetRequestSeviceType() string{
	switch {
		case len(m.ServiceType) > 0:
			return m.ServiceType
		case len(m.RequestType) > 0:
			return m.RequestType
	}
	return ""
}

func (m *Message) GetResponseTagValue(tag string) string{
	var lstFields []Field
	
	lstFields = m.GetFieldsFromResponse()
	
	for _, f := range lstFields{
		if f.Name == tag {
			return f.Value
		}
	}
	return ""
}

func Pop(a *[]string) string{
	var x string
	x, *a = (*a)[0], (*a)[1:]
	return x
}

func (m *Message) GetParsedPromSLAEtaValue() string{
	var sla string
	var slaSeparator string
	slaSeparator = "/"
	
	sla = m.GetResponseTagValue("PROM_SLA_ETA")
	
	if len(sla) == 0 {
		return ""
	} else {
		slaRes := strings.Split(sla, slaSeparator)
		slaDate := Pop(&slaRes)
		slaTime := Pop(&slaRes)
		minAfter := Pop(&slaRes)
		minBefore := Pop(&slaRes)
		etc := &slaRes
		
		dateLayout := "02Jan2006"	//Standard GO Layout for SLA Date.
		timeLayot := "1504"			//Standard GO Layout for SLA Time.
		
		var dt, t time.Time
		var err error
		
		//Parsing Date & Time
		dt, err = time.Parse(dateLayout, slaDate)
		if err != nil {
			fmt.Println(err)
		}
		
		t, err = time.Parse(timeLayot, slaTime)
		if err != nil {
			fmt.Println(err)
		}
		
		//Adding time to Date
		dt = dt.Add(time.Hour * time.Duration(t.Hour()))
		dt = dt.Add(time.Minute * time.Duration(t.Minute()))
		
		before, err := strconv.Atoi(minBefore)
		if err != nil {
			fmt.Println(err)
		}
		
		after, err := strconv.Atoi(minAfter)
		if err != nil {
			fmt.Println(err)
		}
		
		//Getting SLA Start & End.
		dtBefore := dt.Add(time.Minute * time.Duration(before * -1)) //Return time
		dtAfter :=  dt.Add(time.Minute * time.Duration(after))
		
		_ = etc //For now, we don't use the rest of the values.
		
		outLayout := genTimeLayout
		return fmt.Sprintf("SLA Start: %s, SLA End %s", dtBefore.Format(outLayout), dtAfter.Format(outLayout))
	}
}

func readDir(dir, fileMatch string, debug bool) ([]string, error){
	var fnd, nfnd int
	var files []string
	//This function will check all the files in the dir.
	
	if(debug) { fmt.Println("------------------------------------") }
	
	fmt.Println("Checking files in dir ", dir)
	
	if(debug) { fmt.Println("------------------------------------") }
	
	err := fp.Walk(dir, func(path string, info os.FileInfo, err error) error{
		//If the document is a folder, it will ignore it.
		if !info.IsDir() {
			//Here the function will verify if the file matches the expression entered.
			fn := fp.Base(path) //Extract only the filename
			
			match, er := fp.Match(fileMatch, fn) //Match the filename with the format entered as parameter.
			
			if(er != nil){
				fmt.Printf("Error matching file: [%v]\n", er)
				return er
			}
			
			if(match){
				fnd++
				if(debug){fmt.Printf("Match Found with expression: %v\n", path)}
				files = append(files, path)
			} else {
				nfnd++
				if(debug){fmt.Printf("Match not found with expression: %v\n", path)}
			}
		}
		
		return nil
	})
	
	if(debug){
		fmt.Println("------------------------------------")
		fmt.Println("File search results")
		fmt.Println("------------------------------------")
		fmt.Println("Total files:\t", fnd+nfnd)
		fmt.Println("Totals Matched:\t", fnd)
		fmt.Printf("Files Matched: \n%v\n", files)
		fmt.Println("------------------------------------")
	}
	
	if err != nil {
		fmt.Printf("Error reading folder: [%v]\n", err)
		return nil, err
	}
	
	return files, nil
}

func CleanUTF8(s string) string{
	if !utf8.ValidString(s) {
        v := make([]rune, 0, len(s))
        for i, r := range s {
            if r == utf8.RuneError {
                _, size := utf8.DecodeRuneInString(s[i:])
                if size == 1 {
                    continue
                }
            }
            v = append(v, r)
        }
        s = string(v)
    }
	return s
}

func (l *Log) IsDateOk() bool {
	//For now, there are few defined formats.
	var formats []string
	generic := genTimeLayout
	formats = append(formats, "Mon 2006-01-02 15:04:05") //Matchs: Thu 2018-08-02 11:47:24
	formats = append(formats, "02 Jan 2006 15:04:05") //Matchs: 02 Aug 2018 12:08:10
	formats = append(formats, generic) //Matchs: 2018-07-21 14:12:17
	
	for _,format := range formats {
		_, err := time.Parse(format, l.Date)
		if err == nil {
			return true
		}
	}
	return false
}

func (l *Log) GetParsedDate() time.Time {
	//For now, there are few defined formats.
	var formats []string
	var t time.Time
	generic := genTimeLayout
	formats = append(formats, "Mon 2006-01-02 15:04:05") //Matchs: Thu 2018-08-02 11:47:24
	formats = append(formats, "02 Jan 2006 15:04:05") //Matchs: 02 Aug 2018 12:08:10
	formats = append(formats, generic) //Matchs: 2018-07-21 14:12:17
	
	for _,format := range formats {
		parsed, err := time.Parse(format, l.Date)
		if err == nil {
			return parsed
		}
	}
	return t
}

func (l *Log) SetLogData(match, rexGroups []string) {
	md := map[string]string{}
	for i, m := range match {
		md[rexGroups[i]] = m
	}
	l.Date = md["DateTime"]
	l.Message = md["Message"]
	
	return
}

func readLogXMLFile(prefix, awb int, files []string, debug bool) ([]Message, error){
	var msgs []Message
	if(debug) { fmt.Println("------------------------------------") }
	fmt.Println("Reading files")
	if(debug) { fmt.Println("------------------------------------") }
	
	//Reading the files collection.
	for _, file := range files {
		if(debug) { fmt.Println("------------------------------------") }
		fmt.Println("Reading ", file)
		if(debug) { fmt.Println("------------------------------------") }
		
		//Reading file
		dat, err := ioutil.ReadFile(file)
		
		if err != nil {
			fmt.Printf("Error reading file: [%v]\n", err)
			return msgs, err
		}
		
		//Remove special characters before continue.
		data := CleanUTF8(string(dat))
		
		//This expression will find all <Message></Message> Tags in the log.
		//The new expression also considers the Date & Time of the Log Registry.
		regExpression := `(?mUs)(?:(?P<DateTime>^.*\d{2}:\d{2}:\d{2}).*)+?\s+?(?P<Message><Message\s.*</Message>)+?`
		rex := regexp.MustCompile(regExpression)
		rexGroups := rex.SubexpNames()
		
		//Searching for all the <Message /> tags.
	
		for _, match := range rex.FindAllStringSubmatch(data, -1) {
			//With the new Regex, it will neeed a new Struct.
			var logObj Log
			logObj.SetLogData(match, rexGroups)
			
			//Now that the XML was extracted, it's time to process it.
			xmlInput := []byte(fmt.Sprintf("%s%s", `<?xml version="1.0" encoding="UTF-8"?>`, logObj.Message))
			
			var msgContext Message
			er := xml.Unmarshal(xmlInput, &msgContext)
			
			if er != nil {
				fmt.Printf("Error reading XML in file %s: [%v]\n", file, er)
				if(debug){
					fmt.Println("------------------------------------")
					fmt.Printf("XML Formatted:\n%s\n", xmlInput)
					fmt.Println("------------------------------------")
				}
				return msgs, er
			} else {
				//Verify if the found date is OK.
				//If it's not, the assigned date will be the one found in the XML.
				if(debug){
					fmt.Println("------------------------------------")
					fmt.Printf("logObj.IsDateOk(): %v\n", logObj.IsDateOk())
					fmt.Printf("logObj.IsDateOk(): %v\n", logObj.GetParsedDate())
				}
				if !logObj.IsDateOk() { logObj.Date = msgContext.Date }
				msgContext.DateParsed = logObj.GetParsedDate()
				msgContext.XML = logObj.Message
				msgContext.FileName = fp.Base(file)
			}
			
			if(debug){
				fmt.Println("------------------------------------")
				fmt.Println("Message Metadata")
				fmt.Println("------------------------------------")
				//fmt.Printf("Unmarshal%#v\n", msgContext.AWBContent) //Only needed to debug specific cases.
				fmt.Println("Awbs Found\t\t", msgContext.GetParsedAWBsFromRequest())
				fmt.Println("MessageType\t\t", msgContext.MessageType)
				fmt.Println("ServiceReqType\t\t", msgContext.GetRequestSeviceType())
				fmt.Println("XML Date\t\t", msgContext.Date)
				fmt.Println("File\t\t\t", msgContext.FileName)
				fmt.Println("Log Date\t\t", msgContext.DateParsed) 
				//fmt.Printf("XML:\n%s\n", msgContext.XML) //Only needed to debug specific cases.
			}
			
			//The XML has metadata to work & needs to match the required AWB.
			if(msgContext.ContainsAwb(prefix, awb)){
				if(debug){
					fmt.Println("------------------------------------")
					fmt.Println("AWB Found")
					//fmt.Printf("Unmarshal: %#v\n", msgContext.GetFieldsFromResponse()) //Only needed to debug specific cases.
					//fmt.Printf("SLA Flight Response: %#v\n", msgContext.GetResponseTagValue("PROM_SLA_ETA")) //Only needed to debug specific cases.
					fmt.Println("------------------------------------")
				}
				
				msgs = append(msgs, msgContext)
			}
		}
	}
	
	sort.Slice(msgs, func(i, j int) bool {
		if(msgs[i].DateParsed.Sub(msgs[j].DateParsed) < 0){
			return true
		}
		if(msgs[i].DateParsed.Sub(msgs[j].DateParsed) > 0){
			return false
		}
		if(msgs[i].ServiceType < msgs[j].ServiceType) {
			return true
		}
		if(msgs[i].ServiceType > msgs[j].ServiceType) {
			return false
		}
		
		return msgs[i].MessageType < msgs[j].MessageType
	})
	
	if(debug){
		fmt.Println("------------------------------------")
		fmt.Println("Reading results")
		fmt.Println("------------------------------------")
		fmt.Println("Total files: ", len(files))
		fmt.Println("Messages matched: ", len(msgs))
	}
	
	return msgs, nil
}

func AddSeparator(f *os.File, length ...int){
	l := 150
	if len(length) > 0 {
		l = length[0]
	}
	
	for i:=0;i<l;i++ {
		f.WriteString("-")
	}
	f.WriteString("\n")
}

func CreateReport(name string, messages []Message, debug bool) error{
	if(debug) { fmt.Println("------------------------------------") }
	fmt.Println("Creating report")
	if(debug) { fmt.Println("------------------------------------") }
	
	f, err := os.Create(name)
	if err != nil {
		fmt.Printf("Error creating report file: [%v]\n", err)
		return err
	}
	defer f.Close()
	
	//Creating the report file.
	f.WriteString(fmt.Sprintf("Report created at: %v\n", time.Now()))
	AddSeparator(f)
	AddSeparator(f, 0)
	AddSeparator(f, 120)
	
	if(len(messages) <= 0){
		if(debug) { fmt.Println("------------------------------------") }
		fmt.Println("There are no results to report")
		if(debug) { fmt.Println("------------------------------------") }
		f.WriteString("No results found to report\n")
	}
	
	for _, msg := range messages {
		f.WriteString(fmt.Sprintf("Time: %s, Message type: %s, Service type: %s, Found in: %s\n", 
			msg.Date, msg.MessageType, msg.GetRequestSeviceType(), msg.FileName))
		eta := msg.GetResponseTagValue("PROM_SLA_ETA")
		if(eta != ""){
			f.WriteString(fmt.Sprintf("PROM_SLA_ETA: %s, %s\n", eta, msg.GetParsedPromSLAEtaValue()))
		}
		AddSeparator(f, 120)
		f.WriteString(fmt.Sprintf("%s\n",msg.XML))
		AddSeparator(f, 120)
	}
	
	return nil
}

func SetReportFileName(reportName *string, prefix, awb int){
	//Set the Time.
	t := time.Now()
	
	//Replacements
	*reportName = strings.Replace(*reportName, "%y", fmt.Sprintf("%d", t.Year()), -1)
	*reportName = strings.Replace(*reportName, "%M", fmt.Sprintf("%02d", t.Month()), -1)
	*reportName = strings.Replace(*reportName, "%d", fmt.Sprintf("%02d", t.Day()), -1)
	*reportName = strings.Replace(*reportName, "%h", fmt.Sprintf("%02d", t.Hour()), -1)
	*reportName = strings.Replace(*reportName, "%m", fmt.Sprintf("%02d", t.Minute()), -1)
	*reportName = strings.Replace(*reportName, "%s", fmt.Sprintf("%02d", t.Second()), -1)
	*reportName = strings.Replace(*reportName, "%n", fmt.Sprintf("%03d-%08d", prefix, awb), -1)
	*reportName = strings.Replace(*reportName, "%p", fmt.Sprintf("%03d", prefix, awb), -1)
	*reportName = strings.Replace(*reportName, "%c", fmt.Sprintf("%08d", prefix, awb), -1)
}

func main() {
	//Setting flags
	report := `Report file name. Parameters: %y year, %M Month (01-12), %d day (01-31), %h hour, %s seconds, %n AWB Number (123-12345678), %p AWB Prefix, %c AWB Code.`

	dirPath := flag.String("d", "./", "File directory")
	fileRegEx := flag.String("f", "booking*.log*", "File name in regular expression")
	debugFlag := flag.Bool("v", false, "Set debug")
	reportName := flag.String("r", "%y-%M-%d_%h-%m-%s_%n-report.txt", report)
	
	//Reading flag and arguments
	flag.Parse()
	args := flag.Args()
	
	//Passing parameters
	debug := *debugFlag
	dir := *dirPath
	fileName := *fileRegEx
	
	if(debug){
		fmt.Println("------------------------------------")
		fmt.Println("Parameters Entered")
		fmt.Println("------------------------------------")
		fmt.Println("Input parameters: ")
		fmt.Println("dir: ", dir)
		fmt.Println("file regex: ", fileName)
		fmt.Println("AWB: ", args)
	}
	
	if(len(args) > 2){
		fmt.Println("You have more arguments than the required. You need to enter the AWB like 00 00000000 or 00-00000000. Exiting.")
		return
	}
	
	if(len(args) < 1){
		fmt.Println("You didn't add the AWB. Please enter the AWB Prefix and Code like 00 00000000 or 00-00000000. Exiting.")
		return
	}
	
	//Declaring variables
	var prefix, awb int
	var pref, code string
	var err error
	
	//Argument validation
	switch(len(args)){
		case 1:
			sp := strings.Split(args[0], "-")
			if(len(sp) != 2){
				fmt.Println("There is a problem with the AWB code. Verifiy if follows the format 00-00000000. Example: 45-12345678. Exiting.")
				return
			}
			pref = sp[0]
			code = sp[1]
		case 2:
			pref = args[0]
			code = args[1]
	}
	
	prefix, err = strconv.Atoi(pref)
	
	if(err != nil){
		fmt.Println("There is a problem with the AWB code. Verify if only contain numbers and/or hyphen (-). Exiting.")
		return
	}
	
	awb, err = strconv.Atoi(code)
	
	if(err != nil){
		fmt.Println("There is a problem with the AWB code. Verify if only contain numbers and/or hyphen (-). Exiting.")
		return
	}
	
	if(debug){
		fmt.Println("------------------------------------")
		fmt.Println("Parsing OK")
		fmt.Println("------------------------------------")
		fmt.Println("AWB Prefix: ", prefix)
		fmt.Println("AWB Code: ", awb)
	}
	
	if(debug){
		fmt.Println("------------------------------------")
	}
	
	//Reading dir
	var files []string
	files, err = readDir(dir, fileName, debug)
	
	if(err != nil){
		fmt.Println("There was an error reading the folder. Exiting.")
		return
	}
	
	//Read files and check logs.
	var messages []Message
	messages, err = readLogXMLFile(prefix, awb, files, debug)
	
	if(err != nil){
		fmt.Println("There was an error reading the files. Exiting.")
		return
	}
	
	//Creating report.
	SetReportFileName(reportName, prefix, awb)
	CreateReport(*reportName, messages, debug)
}

