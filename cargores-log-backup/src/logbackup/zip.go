package logbackup

import (
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"github.com/mholt/archiver"
)

func searchAndUnzip(server LogServer, wg *sync.WaitGroup, count *int) {
	log.Printf("Checking in folder %s\n", server.GetSavePath())

	//Check all the files inside the folder.
	filepath.Walk(server.GetSavePath(), func(path string, info os.FileInfo, err error) error {
		if err != nil {
			log.Printf("prevent panic by handling failure accessing a path %q: %v\n", path, err)
			return err
		}

		//Check if the file is a valid compressed file.
		if !info.IsDir() {
			_, err := archiver.ByExtension(path)

			//If is compressed, the file must be uncompressed.
			if err == nil {
				//Get the name of the file
				name := strings.TrimSuffix(path, filepath.Ext(path))            //Delete the extension and keep the name.
				basedir := filepath.Base(name)                                  //Rescue the name of the file.
				name = strings.Replace(name, ".", "_", -1)                      //With this, the process skips any error that could happen in windows systems.
				os.MkdirAll(name, os.ModePerm)                                  //Create the new folder.
				name = fmt.Sprintf("%s%c%s", name, filepath.Separator, basedir) //Adding a slash at the end to set the name as folder.

				//Uncompress
				log.Printf("Uncompressing file %s into %s\n", path, name)
				err = archiver.DecompressFile(path, name)

				//If there is an error, try to apply the other method
				if err != nil {
					log.Printf("There was an error when tried to unzip the file %s. Cannot uncompress it.\n", name)
					log.Println(err)
					log.Println("Trying again with other method.")
					err = archiver.DecompressFile(path, name)
				}

				if err != nil {
					log.Printf("There was an error when tried to unzip the file %s. Cannot uncompress it.\n", name)
					log.Println(err)
				} else {
					//Since it was unzipped, just increase the counter
					*count++
					log.Printf("File %s unzipped successfully.\n", name)
				}

				//Finally, remove the zipped files.
				e := os.Remove(path)

				if e != nil {
					log.Printf("There was an error while trying to delete the file %s.", path)
					log.Println(e)
				}
			}
		}
		return nil
	})

	wg.Done()
}

//Unzip uncompress all the compressed files, then remove them.
func (f *LogBackup) Unzip() {
	var serverWg sync.WaitGroup //Async unzipping

	lstServers := f.GetServerList()

	//Metrics
	start := time.Now()
	count := 0

	//Generates the Async Fetching.
	for _, server := range lstServers {
		serverWg.Add(1)                              //Add a new Task
		go searchAndUnzip(server, &serverWg, &count) //Run the Task
	}

	serverWg.Wait() //Wait until the task is done.

	//Metrics
	finish := time.Since(start)
	log.Println("=============================================================================")
	log.Printf("Files unzipped successfuly. Total: %d files and took %s seconds to process.\n", count, finish)
	log.Println("=============================================================================")
}

func getAllFilePaths(server LogServer, wg *sync.WaitGroup, lstFiles *[]string) {
	filepath.Walk(server.GetSavePath(), func(path string, info os.FileInfo, err error) error {
		if err != nil {
			log.Printf("prevent panic by handling failure accessing a path %q: %v", path, err)
			return err
		}

		//Check if the file is a valid file.
		if !info.IsDir() {
			//Add the path into the list.
			*lstFiles = append(*lstFiles, path)
		}

		return nil
	})

	wg.Done()
}

//ZipEverything will check all the paths and generate a new Zipped file.
//After that, it will remove all the files.
func (f *LogBackup) ZipEverything() {
	//Metrics
	start := time.Now()

	//Set the name and path of the zipped file.
	name := fmt.Sprintf("%s.zip", start.Format("20060102150405"))
	filePath := f.ZipSavePath(name)

	log.Printf("Creating zipped file %s", filePath)

	//Zip the LogFolder completely to keep the structure.
	err := archiver.Archive([]string{Config.LogFolder}, filePath)

	//Check if something happened.
	if err != nil {
		log.Println("The zipped procces found a problem.")
		log.Panic(err)
	} else {
		finish := time.Since(start)

		log.Println("=============================================================================")
		log.Printf("Zip created successfuly. Took %s seconds to finish.\n", finish)
		log.Println("=============================================================================")
	}
}

//Clear remove all the files from the log folder.
func clear() {
	log.Println("Cleaning space...")
	os.RemoveAll(Config.LogFolder)
	log.Println("Done")
}

//ZipEverythingAndClear call to ZipEverything and clear functions.
func (f *LogBackup) ZipEverythingAndClear() {
	f.ZipEverything()
	clear()
}
