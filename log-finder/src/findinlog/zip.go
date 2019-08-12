package findinlog

import (
	"archive/zip"
	"bytes"
	"fmt"
	"log"
	"path/filepath"
	"reflect"
	"sort"
	"strings"
	"sync"
	"time"

	"github.com/mholt/archiver"
)

//FindConfig configures the data to fetch
type FindConfig struct {
	Path, Format string
	Prefix, Awb  int
	MaxWorkers   int
	Debug        bool
}

//MaxWorkers set the Maximium of Threads to work with.
var MaxWorkers = 10

func readAndCheck(fc FindConfig, f *zip.File, wg *sync.WaitGroup, m *[]Message, count *int, workCout *int) {
	fn := filepath.Base(f.Name)
	log.Printf("[%s] Reading content.\n", f.Name)

	rc, err := f.Open()
	if err != nil {
		log.Printf("[%s] There was an error trying to open the file.\n", f.Name)
		log.Println(err)
	} else {
		defer rc.Close()
		//Read the file and obtain the content.
		buf := new(bytes.Buffer)
		buf.ReadFrom(rc)

		//Get the data
		err = FetchMessage(fc.Debug, fn, buf.String(), fc.Prefix, fc.Awb, m)

		if err != nil {
			log.Printf("[%s] There was an error trying to process the data.\n", f.Name)
			log.Println(err)
		}
	}

	//Finally, notify that the job is done.
	*count--
	log.Printf("[%s] Finished processing file. Pending: %d\n", f.Name, *count)
	*workCout--
	wg.Done()
}

//SearchInZip Search files in Zip.
func SearchInZip(fc FindConfig) ([]Message, error) {
	var readerWg sync.WaitGroup //Async file reading from zip file.
	var msgs []Message

	//Setting the max of workers
	if fc.MaxWorkers > 0 {
		MaxWorkers = fc.MaxWorkers
	}

	log.Printf("Verifying file %s\n", fc.Path)

	//Check if file is a zip
	z, err := archiver.ByExtension(fc.Path)

	//If it's not a zipped file, exit
	if err != nil {
		log.Printf("File %s is not a valid compressed file. \n", fc.Path)
		return msgs, err
	} else if !reflect.DeepEqual(z, archiver.NewZip()) {
		log.Printf("File %s is not a zip file. \n", fc.Path)
		return msgs, err
	}

	//Open a zip archive for reading.
	r, err := zip.OpenReader(fc.Path)
	if err != nil {
		log.Printf("There was an error trying to read file %s. \n", fc.Path)
		return msgs, err
	}
	defer r.Close()

	log.Printf("Checking files in %s\n", fc.Path)

	//Iterate through the files in the archive.
	//For speeding purposes, the function is asyncronous.
	var fileCount int
	var pending = len(r.File)
	var working int

	log.Println("Starting...")

	for _, f := range r.File {
		name := f.FileHeader.Name

		if fc.Debug {
			log.Printf("[%s] Checking if is a directory. \n", name)
		} else {
			log.Printf("[%s] Checking file. \n", name)
		}
		if !f.FileInfo().IsDir() {
			if fc.Debug {
				log.Printf("[%s] It's a file. Checking if matches the filter expression.\n", name)
			}
			fileCount++

			//Check if the file consider the F

			fn := filepath.Base(name) //Extract only the filename

			//Match the filename with the format entered as parameter.
			match, er := filepath.Match(strings.ToLower(fc.Format), strings.ToLower(fn)) //Validate as case insensitive.

			if er != nil {
				log.Printf("[%s] There was an error trying to match the file.\n", name)
				log.Println(er)
			} else if match {
				working++
				if fc.Debug {
					log.Printf("[%s] Yes, it matches. Waiting to read.\n", name)
				} else {
					log.Printf("[%s] Check working capacity to read file.\n", name)
				}
				readerWg.Add(1)
				//Wait until there is more space in memory
				for working > MaxWorkers {
					time.Sleep(1000)
				}
				go readAndCheck(fc, f, &readerWg, &msgs, &pending, &working)
			} else {
				pending--
				if fc.Debug {
					log.Printf("[%s] No, it doesn't match. Continue. Pending: %d\n", name, pending)
				} else {
					log.Printf("[%s] Doesn't match the format entered. Skipping. Pending: %d\n", name, pending)
				}
			}
		} else {
			pending--
			if fc.Debug {
				log.Printf("[%s] Yes, is a dir. Skipping. Pending: %d\n", name, pending)
			} else {
				log.Printf("[%s] Skipping because is a directory. Pending: %d\n", name, pending)
			}
		}
	}

	//Wait to finish the process.
	readerWg.Wait()

	sort.Slice(msgs, func(i, j int) bool {
		if msgs[i].DateParsed.Sub(msgs[j].DateParsed) < 0 {
			return true
		}
		if msgs[i].DateParsed.Sub(msgs[j].DateParsed) > 0 {
			return false
		}
		if msgs[i].ServiceType < msgs[j].ServiceType {
			return true
		}
		if msgs[i].ServiceType > msgs[j].ServiceType {
			return false
		}

		return msgs[i].MessageType < msgs[j].MessageType
	})

	if fc.Debug {
		fmt.Println("------------------------------------")
		fmt.Println("Reading results")
		fmt.Println("------------------------------------")
		fmt.Println("Total files: ", fileCount)
		fmt.Println("Messages matched: ", len(msgs))
	}

	return msgs, nil
}
