package findinlog

import (
	"fmt"
	"os"
	"time"

	"strings"
	//"github.com/go-xmlfmt/xmlfmt"
)

//Report set the Report Data to return.
type Report struct {
	name   string
	Format string
	Debug  bool
}

func addSeparator(f *os.File, length ...int) {
	l := 150
	if len(length) > 0 {
		l = length[0]
	}

	for i := 0; i < l; i++ {
		f.WriteString("-")
	}
	f.WriteString("\n")
}

//CreateReport Create the report to generate.
func (r *Report) CreateReport(messages []Message) error {
	if r.Debug {
		fmt.Println("------------------------------------")
	}
	fmt.Println("Creating report")
	if r.Debug {
		fmt.Println("------------------------------------")
	}

	f, err := os.Create(r.name)
	if err != nil {
		fmt.Printf("Error creating report file: [%v]\n", err)
		return err
	}
	defer f.Close()

	//Creating the report file.
	f.WriteString(fmt.Sprintf("Report created at: %v\n", time.Now()))
	addSeparator(f)
	addSeparator(f, 0)
	addSeparator(f, 120)

	if len(messages) <= 0 {
		if r.Debug {
			fmt.Println("------------------------------------")
		}
		fmt.Println("There are no results to report")
		if r.Debug {
			fmt.Println("------------------------------------")
		}
		f.WriteString("No results found to report\n")
	}

	for _, msg := range messages {
		f.WriteString(fmt.Sprintf("Time: %s, Message type: %s, Service type: %s, Found in: %s\n",
			msg.Date, msg.MessageType, msg.GetRequestSeviceType(), msg.FileName))
		eta := msg.GetResponseTagValue("PROM_SLA_ETA")
		if eta != "" {
			f.WriteString(fmt.Sprintf("PROM_SLA_ETA: %s, %s\n", eta, msg.GetParsedPromSLAEtaValue()))
		}
		addSeparator(f, 120)
		//f.WriteString(fmt.Sprintf("%s\n", xmlfmt.FormatXML(strings.Replace(msg.XML, "\n", "", -1), "\t", "  ")))
		f.WriteString(fmt.Sprintf("%s\n", msg.XML))
		addSeparator(f, 120)
	}

	return nil
}

//SetReportFileName Sets the name of the report, based on the parameter defined.
func (r *Report) SetReportFileName(prefix, awb int) {
	//Set the Time.
	t := time.Now()

	r.name = r.Format

	//Replacements
	r.name = strings.Replace(r.name, "%y", fmt.Sprintf("%d", t.Year()), -1)
	r.name = strings.Replace(r.name, "%M", fmt.Sprintf("%02d", t.Month()), -1)
	r.name = strings.Replace(r.name, "%d", fmt.Sprintf("%02d", t.Day()), -1)
	r.name = strings.Replace(r.name, "%h", fmt.Sprintf("%02d", t.Hour()), -1)
	r.name = strings.Replace(r.name, "%m", fmt.Sprintf("%02d", t.Minute()), -1)
	r.name = strings.Replace(r.name, "%s", fmt.Sprintf("%02d", t.Second()), -1)
	r.name = strings.Replace(r.name, "%n", fmt.Sprintf("%03d-%08d", prefix, awb), -1)
	r.name = strings.Replace(r.name, "%p", fmt.Sprintf("%03d", prefix), -1)
	r.name = strings.Replace(r.name, "%c", fmt.Sprintf("%08d", awb), -1)
}
