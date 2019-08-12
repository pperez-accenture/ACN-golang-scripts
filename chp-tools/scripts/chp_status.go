/*
CHP Status Fetcher
Author: Patricio Pérez - p.perez.bustos@accenture.com
v0.1 | 2018.03.02
Requires: Go 1.9 or higher
*/

package main

import (
	"os"
	"io"
	"fmt"
	"time"
	"flag"
	"bytes"
	"regexp"
	"strings"
	"strconv"
	"net/url"
	"net/http"
	"io/ioutil"
	"html/template"
	 b64 "encoding/base64"
	"github.com/vadimi/go-http-ntlm"	//Library to connect to the server with NTLM authentication.
	"github.com/PuerkitoBio/goquery"	//Library to read the HTML and recover data like jQuery.	
)

//I. Structs

//Login Object
type loginData struct {
	base, uri, post, user, pass string
}

//NTLN Object
type ntlmData struct {
	domain, user, pass, credential string
}

//Session Object
type loginSessionData struct{
	user, id string
}

//Request Object
type request struct{
	action, uri, data string
}

//CHP Attachment Object
type chpAttachment struct{
	EncodedFile, Name string
}

//CHP Object
type chpData struct{
	ChpNumber, Creator, ChpType, ChpResource, ChpApp, ChpName, ChpDetails, ChpRzn, ChpRes, SchdlDate, SchdlStart, SchdlEnd string
	Attachments []chpAttachment
}

//HTML Report Object
type htmlReport struct{
	Chp chpData
	GenDate, ExecTime string
}

//II. Setters
//1. Generic methods
func newRequestObj(action, uri, data string)(request){
	return request{action, uri, data}
}

func setUrlRequest(action, uri, query string)(request){
	path,_ := url.Parse(fmt.Sprintf("http://%s/%s", uri, query))
	//fmt.Printf("%s\n", path)
	return newRequestObj(action, path.String(), "")
}

//2. Specific methods
func setLogin(uri string, user string, pass string) (loginData) {
	base_uri := fmt.Sprintf("http://%s", uri)
	login_uri := fmt.Sprintf("%s/fsLogin.asp?msps=0&LangId=S&CodeId=ESP", base_uri)
	p1, p2 := user, pass
	p3, p8 := "0", "00000000-0000-0000-0000-000000000000"
	login_req := fmt.Sprintf("<root><type>SQL</type><p1>%s</p1><p2>%s</p2><p3>%s</p3><p5></p5><p6></p6><p7></p7><p8>%s</p8></root>",p1,p2,p3,p8)
	return loginData{base_uri, login_uri, login_req, user, pass}
}

func setCHPQueryRQ(uri, user, id, chp string)(request){
	path := fmt.Sprintf("http://%s/Core/UI/vwSearchResults.asp?rid=%s&sno=%s&ui=W&home=Yes&srchType=SearchRequests&srchFor=%s", uri, user, id, chp)
	//fmt.Printf("%s\n", path)
	return newRequestObj("GET", path, "")
}

func setCHPFetchRQ(server, query string)(request){
	return setUrlRequest("GET", server, query)
}
func setPostCHPFetchRQ(server, query string)(request){
	return setUrlRequest("POST", server, query)
}

func setCHPDetailsRQ(uri, user, id, detail_id, entity, entity_id, section, section_id string)(request){
	path := fmt.Sprintf("http://%s/UDF/EntityProfiles.aspx?rid=%s&sno=%s&ui=W&entityID=%s&entity=%s&hideDiv=%s&section=%s&layoutstyle=0&sectionId=%s", uri, user, id, entity_id, entity, detail_id, section, section_id)
	//fmt.Printf("%s\n", path)
	return newRequestObj("POST", path, "")
}

func setDLRQ(uri, base, attachId string)(request){
	path,_ := url.Parse(fmt.Sprintf("http://%s/helpdesk/dlOppAttach.aspx?%s&tAttachId=%s&sCustId=&AppPath=", uri, base, attachId))
	//fmt.Printf("%s\n", path)
	return newRequestObj("GET", path.String(), "")
}

func setNtlmData(domain string, user string, pass string) (ntlmData) {
	credential := fmt.Sprintf("%s\\%s", domain, user)
	return ntlmData{domain, user, pass, credential}
}

func setSession(p1, p2 string) (loginSessionData){
	return loginSessionData{p1, p2}
}

func newAttachment(name, file string)(chpAttachment){
	return chpAttachment{file, name}
}

func setCHPData(chp_number, creator, chp_type, chp_resource, chp_app, chp_name, chp_details, chp_rzn, chp_res, schdl_date, schdl_start, schdl_end string, attachment []chpAttachment)(chpData){
	return chpData{chp_number, creator, chp_type, chp_resource, chp_app, chp_name, chp_details, chp_rzn, chp_res, schdl_date, schdl_start, schdl_end, attachment}
}

func SetHtmlReport(chp chpData, date, execTime string)(htmlReport){
	return htmlReport{chp, date, execTime}
}

//III. Functions
func StreamToByte(stream io.Reader) []byte {
	buf := new(bytes.Buffer)
	buf.ReadFrom(stream)
	return buf.Bytes()
}

func formatTime(dt, oldFormat, newFormat string)(string, error){
	t, err := time.Parse(oldFormat, dt)

	if err != nil {
		return "", err
	}
	return t.Format(newFormat), nil
}

func ntlm_request_raw(rq request, ntlm ntlmData)(io.Reader, error){
	//Setting new NTLM client
	client := http.Client{
        Transport: &httpntlm.NtlmTransport{
            Domain:   ntlm.domain,
            User:     ntlm.user,
            Password: ntlm.pass,
        },
    }
	
	//Reading the data
	body := strings.NewReader(rq.data)	
	//Creating request
	req, err := http.NewRequest(rq.action, rq.uri, body)
	
	if err != nil {
		return nil, err
	}
	
	//Adding credentials
	req.SetBasicAuth(ntlm.user, ntlm.pass)
	
	//Calling to page
	resp, err := client.Do(req)
	
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close() //Closing Connection
	
	//Fetching result
	response, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}
	
	//Transforming to Reader Object
	str := strings.NewReader(string(response))
	return str, nil
}

func ntlm_request(rq request, ntlm ntlmData)(*goquery.Document, error){
	//First, we make the request
	str, err := ntlm_request_raw(rq, ntlm)
	if err != nil {
		return nil, err
	}
	
	//Now, we need to read the result and transforms it into a goQuery object 
	doc, err := goquery.NewDocumentFromReader(str)
	if err != nil {
		return nil, err
	}
	return doc, nil
}

func login_request(login loginData, ntlm ntlmData)(loginSessionData, error){
	var ses loginSessionData

	//Making the request
	rq := newRequestObj("POST", login.uri, login.post)
	doc, err := ntlm_request(rq, ntlm)
	if err != nil {
		return ses, err
	}
	
	doc.Find("root").Each(func(i int, s *goquery.Selection) {
		// For each item found, fetch the data
		p1 := s.Find("p1").Text()
		p2 := s.Find("p2").Text()
		ses = setSession(p1, p2)
	})
	return ses, nil
} 

func search_chp(server string, ses loginSessionData, ntlm ntlmData, chp string)(string, error){
	var uri string
	//Making the request
	rq := setCHPQueryRQ(server, ses.user, ses.id,chp)
	doc, err := ntlm_request(rq, ntlm)
	if err != nil {
		return "", err
	}
	
	//fmt.Printf("%s\n", doc.Html())
	
	doc.Find("body").Each(func(i int, s *goquery.Selection) {
		//Fetching the URL that we are going to use.
		re := regexp.MustCompile("\".*\"")
		uri = s.Find("script").Text()
		pUrl := re.FindAllString(uri, -1)
		uri = strings.Trim(pUrl[0], "\x22")
	})
	
	return uri, nil
}

func splitDetails(s string) (string, string, string, string) {
    x := strings.Split(s, ",")
    return strings.TrimSpace(x[0]), strings.TrimSpace(x[1]), strings.TrimSpace(x[4]), strings.TrimSpace(x[2])
}

func completeDate(dt, separator string)(string){
	//Since the date will possible come as m/d/yyyy, we need to transform to mm/dd/yyyy
	if(len(dt) < 1){
		return dt
	}
	splt := strings.Split(dt, separator)
	m,_ := strconv.Atoi(splt[0])
	d,_ := strconv.Atoi(splt[1])
	y,_ := strconv.Atoi(splt[2])
	
	return fmt.Sprintf("%02d%s%02d%s%d",m,separator,d,separator,y)
}

func fetch_chp(server string, uri string, ses loginSessionData, ntlm ntlmData)(chpData, error){
	//Variables needed for the CHP:
	var chp chpData
	var chp_number, creator, chp_type, chp_resource, chp_app, chp_name, chp_details, chp_rzn, chp_res string
	var schdl_date, schdl_start, schdl_end string
	var lstAttachment []chpAttachment 
	
	//Making the request
	rq := setCHPFetchRQ(server, uri)
	
	//Calling to obtain the info about 
	out, rq_err := ntlm_request(rq, ntlm)
	if rq_err != nil {
		return chp, rq_err
	}
	
	//We need to read the fields that interest us
	out.Find("body").Each(func(i int, s *goquery.Selection) {
		fmt.Printf("Filling CHP Info\n")
		//Main data from the CHP
		chp_PostXML := s.Find("script:contains('sXMLPostPath')").First().Text()					//Used to recover the attachments & Other Data
		
		chp_number = s.Find("#Master_IdPageHeaderTitle").Text() 		//N° CHP
		creator =  s.Find("#trRequestInitiator a").Text() 			//Autor CHP
		chp_type = s.Find("#tdRequestTypeContent").Text()			//Tipo CHP
		chp_resource = s.Find("#tdRequestAssigneeContent").Text()	//Recurso Asignado
		chp_app = s.Find("#tdRequestProductContent").Text()			//Aplicativo
		chp_name = s.Find("#tdRequestDescriptionContent").Text()		//Título CHP
		chp_details, _ = s.Find("#tdRequestDetailsContent").Html()	//Detalle CHP
		chp_rzn = s.Find("#tdRequestCauseContent").Text()			//ID Resolución
		chp_res,_ = s.Find("#tdRequestSolutionContent").Html()			//Resolución
		
		//If it is scheduled, we need to recover that async data.
		var detail_id, tag_id string
		detail_id = "UDFLayoutREQPCON0"
		tag_id = fmt.Sprintf("#img%s", detail_id)
		
		chp_reqpcon0, exists := s.Find(tag_id).Attr("onclick")
		
		fmt.Printf("Checking CHP Schedule\n")
		if !exists {
			//If the tag does not exists, that means that the CHP is still unscheduled.
			fmt.Printf("Schedule not found\n")
		} else {
			fmt.Printf("Reading CHP Schedule\n")
			//Otherwise, first we need to recover the data, then assign it and fetch the details.
			var entity, entity_id, section, section_id  string
			//First, we need to find the code inside the javascript
			re := regexp.MustCompile("\\(.*\\)")
			pData := re.FindAllString(chp_reqpcon0, -1)
			//Now, we need to identify the data
			data := strings.Replace(strings.Trim(pData[0], "()"), "'", "", -1) //Cleaning the parameters
			entity, entity_id, section, section_id = splitDetails(data) //Splitting the result
			section_id = strings.Replace(section_id, "REQ_", "", 1) //Replacing the code to get the one used on the URL
			
			//Now we need to do the async call and fetch the result
			async_rq := setCHPDetailsRQ(server, ses.user, ses.id, detail_id, entity, entity_id, section, section_id)
			async_rs, err := ntlm_request(async_rq, ntlm)
			
			//If all is OK, it is time to recover the missing data
			if(err == nil){
				async_rs.Find("table tr").Each(func(j int, asnc *goquery.Selection) {
					td1 := strings.TrimSpace(asnc.Find("td").First().Text())
					if strings.HasPrefix(td1, "Fecha Servicio") {
						var terr error
						schdl_date = strings.TrimSpace(asnc.Find("td").Last().Text())
						//Formatting Date to UTC Version
						schdl_date = completeDate(schdl_date, "/")
						
						schdl_date, terr = formatTime(schdl_date, "01/02/2006", "2006/01/02")
						if terr != nil {
							fmt.Print("%v\n", terr)
						}
					}
					if strings.HasPrefix(td1, "Hora Inicio") {
						schdl_start = strings.TrimSpace(asnc.Find("td").Last().Text())
					}
					if strings.HasPrefix(td1, "Hora Fin") {
						schdl_end = strings.TrimSpace(asnc.Find("td").Last().Text())
					}
				})
				//fmt.Printf("%v,%v,%v\n", schdl_date, schdl_start, schdl_end)
			} else{
				fmt.Printf("%v\n", err)
			}
		}
		
		//Finally, we check if there are any attachments to download.
		//First, it is necessary to get the PostURL.
		fmt.Printf("Checking CHP Attachments\n")
		
		re_xml := regexp.MustCompile("sXMLPostPath = \"(?P<xml_uri>.*)\"")
		pUrl := re_xml.FindAllStringSubmatch(chp_PostXML, -1)[0]
		post_rq := setPostCHPFetchRQ(server, "HelpDesk/" + pUrl[1] + "&Action=5")
		
		re_base := regexp.MustCompile("BASEQRYSTRING = \"(?P<xml_uri>.*)\"")
		pBase := re_base.FindAllStringSubmatch(chp_PostXML, -1)[0]

		//Calling to obtain the other info
		post_out, rq_err := ntlm_request(post_rq, ntlm)
		
		if rq_err == nil {
			//Reading the XML to get all the attachments
			post_out.Find("#RequestAttachments_tblMain tbody tr").Each(func(J int, ps *goquery.Selection) {
				atch := ps.Find("td").Has("a[href*=downloadfile]").First()
				atch_name := atch.Text()
				atch_html, _ := atch.Html()
				
				//With the HTML, we are going to download the file.
				re_dl := regexp.MustCompile("downloadfile\\(\\&\\#39;(?P<dl_id>.*)\\&\\#39;\\)")
				pDl := re_dl.FindAllStringSubmatch(atch_html, -1)[0]
				
				fmt.Printf("Downloading %s\n", atch_name)
				dl_rq := setDLRQ(server, pBase[1], pDl[1])
				dl_data, dl_err := ntlm_request_raw(dl_rq, ntlm)
				if dl_err == nil {
					sEnc := b64.StdEncoding.EncodeToString([]byte(StreamToByte(dl_data)))
					lstAttachment = append(lstAttachment, newAttachment(atch_name, sEnc))
				} else {
					fmt.Printf("%v\n", dl_err)
				}
			})
		} else {
			fmt.Printf("%v\n", rq_err)
		}
		
	})
	
	//Finally, we save the data into the CHP Object to return
	chp = setCHPData(
		chp_number, creator, chp_type, chp_resource,
		chp_app, chp_name, chp_details, chp_rzn, chp_res,
		schdl_date, schdl_start, schdl_end, lstAttachment)
	
	return chp, nil
}

func outputHtml(fileName string, templatePath string, chp chpData, start, end time.Time)(error){
	//Opening the Template File
	dat, err := ioutil.ReadFile(templatePath)
	if err != nil {
		return err
	}
	//Transform the file into a string
	str := string(dat)
	
	//Transforms the HTML to template object
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
	execTime := end.Sub(start)
	date := end.Format("2006/01/02 15:04 MST")
	report := SetHtmlReport(chp, date, execTime.String())
	err = t.Execute(f, report) 
	if err != nil {
		return err
	}
	
	//Finally, close the file
	f.Close() 
	return nil
}

/// IV. Main
func main()  {
	//1. Variables
	var chp, server, ntlm_domain, ntlm_user, ntlm_pass, login_user, login_pass, outFile, templateFile  string
	
	//Starting time
	start := time.Now()

	//1.1 Checking flags and arguments
	
	//Setting arguments as flags:
	serverFlag := flag.String("chpServer", "57.228.131.171", "Defines the Server of the Changepoint (http://{chpServer})")
	ntlmDomainFlag := flag.String("ntlmDomain", "DOMAIN", "NTLM Domain")
	ntlmUserFlag := flag.String("ntlmUser", "user", "NTLM User")
	ntlmPassFlag := flag.String("ntlmPass", "pass", "NTLM Password")
	loginUserFlag := flag.String("chpUser", "foo@bar.com", "Defines the CHP User")
	loginPassFlag := flag.String("chpPass","loremIpsum", "Defines the CHP User")
	outFileFlag := flag.String("outputPath", "../output/chp-status.html", "Sets the path and name of the Output File. \nBy Default, it will save on the same path where this script is located.")
	templateFileFlag := flag.String("templatePath", "../templates/chp-status.html", "Sets the path and name of the Template File used to save. \nBy Default, the template should be located on the same path where this script is located.")
	
	//Reading flags and arguments
	flag.Parse()
	
	//Cheching the main argument
	args := flag.Args()
	
	if(len(args) > 0){ 
		chp = args[0] //The first argument must be the CHP.
	} else {
		chp = "TCK-2018-0000001" //We set a default chp just for a demostration.
	}
	
	//Passing the argument values (if they are not set, it will read the default value for each.x
	server, ntlm_domain, ntlm_user, ntlm_pass = *serverFlag, *ntlmDomainFlag, *ntlmUserFlag, *ntlmPassFlag
	login_user, login_pass, outFile, templateFile = *loginUserFlag, *loginPassFlag, *outFileFlag, *templateFileFlag
	
	fmt.Printf("Reading Data From the Changepoint: %v\n",chp)
	
	//2. System Variables
	var login loginData
	var ntlm ntlmData

	//3. Function
	//3.1 Setting objects
	login = setLogin(server, login_user, login_pass)
	ntlm = setNtlmData(ntlm_domain, ntlm_user, ntlm_pass)
	
	//3.2 Login
	fetch_start := time.Now()
	fmt.Printf("Logging into the CHP\n")
	ses, err:=login_request(login, ntlm)
	if err != nil {
		fmt.Printf("%s\n", err)
		return
	}
	
	//3.3 Search
	fmt.Printf("Searching CHP\n")
	uri, s_err := search_chp(server, ses, ntlm, chp);
	if s_err != nil {
		fmt.Printf("%s\n", s_err)
		return
	}
	
	//3.4 Fetch
	fmt.Printf("Obtaining Results\n")
	chp_res, c_err := fetch_chp(server, uri, ses, ntlm)
	if c_err != nil {
		fmt.Printf("%s\n", c_err)
		return
	}
	
	//3.5 Generate Output file
	fetch_end := time.Now()
	ferr := outputHtml(outFile, templateFile, chp_res, fetch_start, fetch_end)
	if ferr != nil {
		fmt.Printf("%s\n", ferr)
		return
	}
	
	//Printing execution time
	end := time.Now()
	fmt.Printf("Execution time: %v.\n", end.Sub(start))
}