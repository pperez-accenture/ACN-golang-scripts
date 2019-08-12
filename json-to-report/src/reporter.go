package main

import (
	//"fmt"
	"os"
	"log"
	"time"
	"regexp"
	"io/ioutil"
	"encoding/json"
	"html/template"
)

type WLServerStatus struct{
	Domain	 		string `json:"domain"`
    Servers			[]WLServer `json:"servers"`
}

type WLServer struct{
	ListenAddress	string `json:"sslListenAddress"`
    ServerPath		string `json:"currentDir"`
	Name	 		string `json:"name"`
    State			string `json:"state"`
	AdminURL 		string `json:"adminUrl"`
    DefaultURL		string `json:"defaultUrl"`
	HomePath		string `json:"home"`
}

type WLServerReport struct{
	WLDomainStatus 	[]WLServerStatus
	GenDate			string
}

func readJsonFile(fileName string)(WLServerStatus){
	//Read JSON
	raw, err := ioutil.ReadFile(fileName)
	if err != nil { log.Fatal(err) }
	
	//Transforms into object.
	var report WLServerStatus
	
	err = json.Unmarshal(raw, &report)
	if err != nil { log.Fatal(err) }
	
	return report
}

func outputHtml(fileName string, templatePath string, domains []WLServerStatus) (error){
	//Opening the Template File
	dat, err := ioutil.ReadFile(templatePath)
	if err != nil {
		return err
	}
	//Transform the file into a string
	str := string(dat)
	
	//Transforms the HTML data to template object (Used for HTML Tags)
	t := template.Must(template.New("main").Funcs(template.FuncMap{
		"safeHTML": func(s string) template.HTML {
			b := []byte(s)
			return template.HTML(b)
		},
	}).Parse(str))
	
	//Now we create the file
	f, err := os.Create(fileName)
	if err != nil {
		return err
	}
	//Fill the template with the data and save it into the file
	now := time.Now()
	date := now.Format("2006/01/02 15:04 MST")
	report := WLServerReport{domains, date}
	err = t.Execute(f, report) 
	if err != nil {
		return err
	}
	
	//Finally, close the file
	f.Close() 
	return nil
}

func main(){
	files, err := ioutil.ReadDir("./")
	
    if err != nil {
        log.Fatal(err)
    }

	//Read only the JSON files.
	var re = regexp.MustCompile(`(?m).*\.json`)
	
	var domains []WLServerStatus
    for _, f := range files {
		fileName := f.Name()
		if re.MatchString(fileName){
			report := readJsonFile(fileName)
			domains = append(domains, report)
		}
    }
	err = outputHtml("../out/report.html", "../template/report.html", domains);
	if err != nil {
		log.Fatal(err)
	}
} 