package findinlog

import (
	"encoding/xml"
	"fmt"
	"log"
	fp "path/filepath"
	"regexp"
	"strconv"
	"strings"
	"time"
	"unicode/utf8"
)

//AWB Contains the Prefix & Number to Search.
type AWB struct {
	Prefix string
	Number string
}

//Field is a representation of a tag.
type Field struct {
	Name  string `xml:"name,attr"`
	Type  string `xml:"type,attr"`
	Value string `xml:"String"`
}

//InnerRQTag contains the value and the AwbID Tag (if it's possible.)
type InnerRQTag struct {
	AwbID AWB          `xml:"AwbId"`
	Inner []InnerRQTag `xml:",any"`
}

//InnerRSTag is a pseudo interface.
type InnerRSTag struct {
	Inner    []InnerRSTag `xml:",any"`
	ResField []Field      `xml:"Field"`
}

//Message is the root tag to search.
type Message struct {
	//XML Metadata
	AWBResponse []InnerRSTag `xml:"ResponseContent"`
	AWBContent  []InnerRQTag `xml:"RequestContent"`
	Date        string       `xml:"MessageOrigin>GMTDateTime"`
	MessageType string       `xml:"messageType,attr"`
	ServiceType string       `xml:"serviceType,attr"`
	RequestType string       `xml:"crsRequestType,attr"`

	//Added parameters
	DateParsed time.Time
	FileName   string
	XML        string
}

//Log is the output formar
type Log struct {
	Date    string
	Message string
}

var emptyField []Field
var genTimeLayout = "2006-01-02 15:04:05"

//Recursive method to get the AWB.
func (t InnerRQTag) getAWBs() []AWB {
	var lstAwb []AWB

	//Verify if the node has Awb Defined.
	if t.AwbID != (AWB{}) {
		lstAwb = append(lstAwb, t.AwbID)
	}

	//If it's not, go one level lower to verify.
	for _, tag := range t.Inner {
		awb := tag.getAWBs()
		for _, a := range awb {
			if a != (AWB{}) {
				lstAwb = append(lstAwb, a)
			}
		}
	}

	//Otherwise, return an empty AWB{}
	return lstAwb
}

//Recursive method to get the Fields.
func (t InnerRSTag) getFields() []Field {
	var lstFields []Field

	//Verify if the node has Field Defined.
	//if t.ResField != nil {
	//	lstFields = append(lstFields, t.ResField)
	//}

	for _, f := range t.ResField {
		lstFields = append(lstFields, f)
	}

	//If it's not, go one level lower to verify.
	for _, tag := range t.Inner {
		field := tag.getFields()
		for _, f := range field {
			if f != (Field{}) {
				lstFields = append(lstFields, f)
			}
		}
	}

	//Otherwise, return an empty ResField{}
	return lstFields
}

//GetAWBsFromRequest Get the AWB's
func (m *Message) GetAWBsFromRequest() []AWB {
	var lstAwb []AWB

	for _, tag := range m.AWBContent {
		awb := tag.getAWBs()
		for _, a := range awb {
			if a != (AWB{}) {
				lstAwb = append(lstAwb, a)
			}
		}
	}
	return lstAwb
}

//GetFieldsFromResponse Get the Fields Tags
func (m *Message) GetFieldsFromResponse() []Field {
	var lstFields []Field

	for _, tag := range m.AWBResponse {
		fields := tag.getFields()
		for _, f := range fields {
			if f != (Field{}) {
				lstFields = append(lstFields, f)
			}
		}
	}
	return lstFields
}

//GetParsedAWBsFromRequest Get Parsed AWB's.
func (m *Message) GetParsedAWBsFromRequest() string {
	var lstAwb []AWB
	var str []string

	lstAwb = m.GetAWBsFromRequest()

	if len(lstAwb) == 0 {
		return ""
	}

	for _, awb := range lstAwb {
		str = append(str, fmt.Sprintf("%s-%s", awb.Prefix, awb.Number))
	}

	return strings.Join(str, ", ")
}

//ContainsAwb Check if the AWB Request List contains the AWB number.
func (m *Message) ContainsAwb(prefix, number int) bool {
	var lstAwb []AWB

	lstAwb = m.GetAWBsFromRequest()

	for _, awb := range lstAwb {
		if awb.Prefix == strconv.Itoa(prefix) &&
			awb.Number == strconv.Itoa(number) {
			return true
		}
	}

	return false
}

//GetRequestSeviceType validates the Service Type for the Message.
func (m *Message) GetRequestSeviceType() string {
	switch {
	case len(m.ServiceType) > 0:
		return m.ServiceType
	case len(m.RequestType) > 0:
		return m.RequestType
	}
	return ""
}

//GetResponseTagValue get the value from tag.
func (m *Message) GetResponseTagValue(tag string) string {
	var lstFields []Field

	lstFields = m.GetFieldsFromResponse()

	for _, f := range lstFields {
		if f.Name == tag {
			return f.Value
		}
	}
	return ""
}

//Pop extracts the element from the first index
func Pop(a *[]string) string {
	var x string
	x, *a = (*a)[0], (*a)[1:]
	return x
}

//GetParsedPromSLAEtaValue extracts the value of PROM_SLA_ETA
func (m *Message) GetParsedPromSLAEtaValue() string {
	var sla string
	var slaSeparator string
	slaSeparator = "/"

	sla = m.GetResponseTagValue("PROM_SLA_ETA")

	if len(sla) == 0 {
		return ""
	}

	slaRes := strings.Split(sla, slaSeparator)
	slaDate := Pop(&slaRes)
	slaTime := Pop(&slaRes)
	minAfter := Pop(&slaRes)
	minBefore := Pop(&slaRes)
	etc := &slaRes

	dateLayout := "02Jan2006" //Standard GO Layout for SLA Date.
	timeLayot := "1504"       //Standard GO Layout for SLA Time.

	var dt, t time.Time
	var err error

	//Parsing Date & Time
	dt, err = time.Parse(dateLayout, slaDate)
	if err != nil {
		log.Println(err)
	}

	t, err = time.Parse(timeLayot, slaTime)
	if err != nil {
		log.Println(err)
	}

	//Adding time to Date
	dt = dt.Add(time.Hour * time.Duration(t.Hour()))
	dt = dt.Add(time.Minute * time.Duration(t.Minute()))

	before, err := strconv.Atoi(minBefore)
	if err != nil {
		log.Println(err)
	}

	after, err := strconv.Atoi(minAfter)
	if err != nil {
		log.Println(err)
	}

	//Getting SLA Start & End.
	dtBefore := dt.Add(time.Minute * time.Duration(before*-1)) //Return time
	dtAfter := dt.Add(time.Minute * time.Duration(after))

	_ = etc //For now, we don't use the rest of the values.

	outLayout := genTimeLayout
	return fmt.Sprintf("SLA Start: %s, SLA End %s", dtBefore.Format(outLayout), dtAfter.Format(outLayout))

}

//CleanUTF8 parses the string as a valid UTF-8 data.
func CleanUTF8(s string) string {
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

//IsDateOk checks the date from the tag.
func (l *Log) IsDateOk() bool {
	//For now, there are few defined formats.
	var formats []string
	generic := genTimeLayout
	formats = append(formats, "Mon 2006-01-02 15:04:05") //Matchs: Thu 2018-08-02 11:47:24
	formats = append(formats, "02 Jan 2006 15:04:05")    //Matchs: 02 Aug 2018 12:08:10
	formats = append(formats, generic)                   //Matchs: 2018-07-21 14:12:17

	for _, format := range formats {
		_, err := time.Parse(format, l.Date)
		if err == nil {
			return true
		}
	}
	return false
}

//GetParsedDate transforms the date to the one expected.
func (l *Log) GetParsedDate() time.Time {
	//For now, there are few defined formats.
	var formats []string
	var t time.Time
	generic := genTimeLayout
	formats = append(formats, "Mon 2006-01-02 15:04:05") //Matchs: Thu 2018-08-02 11:47:24
	formats = append(formats, "02 Jan 2006 15:04:05")    //Matchs: 02 Aug 2018 12:08:10
	formats = append(formats, generic)                   //Matchs: 2018-07-21 14:12:17

	for _, format := range formats {
		parsed, err := time.Parse(format, l.Date)
		if err == nil {
			return parsed
		}
	}
	return t
}

//SetLogData parses the Data as a Log Struct.
func (l *Log) SetLogData(match, rexGroups []string) {
	md := map[string]string{}
	for i, m := range match {
		md[rexGroups[i]] = m
	}
	l.Date = md["DateTime"]
	l.Message = md["Message"]

	return
}

//FetchMessage reads the file and checks all the messages that contains the Prefix and AWB.
//Uses a reference to a Message Array Object to have it working Asyncronously.
func FetchMessage(debug bool, filename, data string, prefix, awb int, msgs *[]Message) error {
	//Remove special characters before continue.
	data = CleanUTF8(data)

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
			log.Printf("[%s] Error reading XML in file: [%v]\n", filename, er)
			if debug {
				log.Printf("[%s] XML Formatted:\n%s\n", filename, xmlInput)
			}
			return er
		}

		//Verify if the found date is OK.
		//If it's not, the assigned date will be the one found in the XML.
		if debug {
			log.Printf("[%s] logObj.IsDateOk(): %v\n", filename, logObj.IsDateOk())
			log.Printf("[%s] logObj.IsDateOk(): %v\n", filename, logObj.GetParsedDate())
		}
		if !logObj.IsDateOk() {
			logObj.Date = msgContext.Date
		}
		msgContext.DateParsed = logObj.GetParsedDate()
		msgContext.XML = logObj.Message
		msgContext.FileName = fp.Base(filename)

		if debug {
			log.Printf("[%s]\n------------------------------------\nMessage Metadata\n------------------------------------\n", filename)
			//log.Printf("Unmarshal%#v\n", msgContext.AWBContent) //Only needed to debug specific cases.
			log.Println("[", filename, "]", "Awbs Found\t\t", msgContext.GetParsedAWBsFromRequest())
			log.Println("[", filename, "]", "MessageType\t\t", msgContext.MessageType)
			log.Println("[", filename, "]", "ServiceReqType\t\t", msgContext.GetRequestSeviceType())
			log.Println("[", filename, "]", "XML Date\t\t", msgContext.Date)
			log.Println("[", filename, "]", "File\t\t\t", msgContext.FileName)
			log.Println("[", filename, "]", "Log Date\t\t", msgContext.DateParsed)
			//log.Printf("XML:\n%s\n", msgContext.XML) //Only needed to debug specific cases.
		}

		//The XML has metadata to work & needs to match the required AWB.
		if msgContext.ContainsAwb(prefix, awb) {
			if debug {
				log.Printf("[%s] AWB Found\n", filename)
				//log.Printf("Unmarshal: %#v\n", msgContext.GetFieldsFromResponse()) //Only needed to debug specific cases.
				//log.Printf("SLA Flight Response: %#v\n", msgContext.GetResponseTagValue("PROM_SLA_ETA")) //Only needed to debug specific cases.
			}

			*msgs = append(*msgs, msgContext)
		}
	}

	return nil
}
