package fetch

import (
	"fmt"
	"io"
	"log"
	"strings"
	"bytes"
	"regexp"

	"io/ioutil"
	"net/http"

	// Here the system will use an external package
	// To use it, in the console you need to run go get and the url
	// example: go get "github.com/PuerkitoBio/goquery"
	"github.com/PuerkitoBio/goquery"
)

//Fetcher have the implemented process to transform the HTML to Data.
type Fetcher struct {
	URLPath string
}

//GameData defines the struct for game info fetching.
type GameData struct {
	//Game Title
	Title string
	//Release Year
	Year  string
	//Places Launched
	Launched string
	//PC/Console release type.
	Platform string
	//(Optional) Edition
	Edition string
}

func readPage(url string) io.Reader{
	//Request the page
	resp, err := http.Get(url)

	if err != nil {
		panic(err)
	}

	//Defer is a great feature of go.
	//It's like finally in a try/catch. When the process leaves the function, the system will execute the sentence.
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		log.Fatalf("status code error: %d %s", resp.StatusCode, resp.Status)
	}

	//Read the result.
	body, err := ioutil.ReadAll(resp.Body)

	if err != nil {
		panic(err)
	}

	return bytes.NewReader(body)
}

//You can initialize and assign the variable to return
func getParams(regEx, data string) (paramsMap map[string]string) {

	//This function will check the expression and return all the ocurrences.
    var compRegEx = regexp.MustCompile(regEx)
    match := compRegEx.FindStringSubmatch(data)

    paramsMap = make(map[string]string)
    for i, name := range compRegEx.SubexpNames() {
        if i > 0 && i <= len(match) {
            paramsMap[name] = match[i]
        }
    }
    return
}

//Process reads the HTML and get the important data.
func (f *Fetcher) Process() (lst []GameData){
	//Buffer for reader and writer. Used also for file.
	var body io.Reader

	body = readPage(f.URLPath)

	// Load the HTML document
	doc, err := goquery.NewDocumentFromReader(body)

	if err != nil {
		log.Fatal(err)
	}

	// Find the review items
	doc.Find("form").Closest("div").Find("h2").Each(func(i int, s *goquery.Selection) {

		// For each item found, get the year and the data
		gameList := s.Next()

		gameList.Find("li").Each(func(j int, s1 *goquery.Selection) {
			var gd GameData

			gd.Year = s.Text()

			//The Game info have the following expression: Title {(Optional) - Subtitles} {(Optional) - Edition release ) - Platform(s) released. (Country Released)
			sections := strings.Split(s1.Text(), " - ")

			var title []string

			for _, str := range sections {
				if strings.Contains(str, "released.") {
					rex := getParams(`(?m)(?P<Version>.*) version released\. \((?P<Launched>.*)\)$`, str)

					//Remove any space
					gd.Launched = strings.TrimSpace(rex["Launched"])
					gd.Platform = strings.TrimSpace(rex["Version"])

				} else if strings.Contains(str, "release") {
					str = strings.Replace(str, "release", "", -1)
					gd.Edition = strings.TrimSpace(str)
				} else {
					title = append(title, str)
				}
			}

			gd.Title = strings.Join(title, " - ")

			fmt.Printf("%#v\n", gd)

			lst = append(lst, gd)
		})

	})

	return
}
