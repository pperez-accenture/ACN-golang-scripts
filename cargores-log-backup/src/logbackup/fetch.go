package logbackup

import (
	"fmt"
	"log"
	"net"
	"os"
	"path"
	"regexp"
	"runtime"
	"strings"
	"sync"
	"time"

	"unicode/utf8"

	"github.com/pkg/sftp"
	"golang.org/x/crypto/ssh"
)

//copyFile will run asynchronously
//The task will copy from the src file into the dest folder.
func copyFile(rootSrc, rootDst, srcPath string, c *sftp.Client, wg *sync.WaitGroup, count *int) {
	defer wg.Done()

	//Recover the path excluding the one setted as root.
	re := regexp.MustCompile(fmt.Sprintf(`(?m)%s(?P<A>.*)`, rootSrc))
	found := re.FindAllStringSubmatch(srcPath, -1)[0]
	dest := found[1]

	if dest != "" {
		fixUtf := func(r rune) rune {
			if r == utf8.RuneError {
				return -1
			}
			return r
		}

		//If the system is Windows, the filename must not contain some restricted characters
		if runtime.GOOS == "windows" {
			rex := regexp.MustCompile(`(<|>|:|"|\/|\\|\||\?|\*)`)
			dest = rex.ReplaceAllString(dest, "")
		}

		dstBase := path.Base(rootSrc)
		dstPath := path.Clean(fmt.Sprintf("%s/%s/%s",
			strings.Map(fixUtf, rootDst),
			strings.Map(fixUtf, dstBase),
			strings.Map(fixUtf, dest),
		)) //Destination Path and file.

		log.Printf("Copying %s into %s", srcPath, dstPath)

		//Try to open the source file
		tries := Config.MaxAttempts    //Setting the max attempts
		timeout := Config.IntervalTime //TTW
		ok := false                    //Flag to control a success

		var srcFile *sftp.File
		var err error

		//Check if it's possible to use the file
		for i := 0; i < tries && !ok; i++ {
			srcFile, err = c.Open(strings.Map(fixUtf, srcPath))

			//If it's not possible, try again after a while.
			if err != nil {
				d, _ := time.ParseDuration(fmt.Sprintf("%fs", timeout))
				time.Sleep(d)
				tries++
			} else {
				ok = true
			}
		}

		//Cannot use the file, sorry.
		if !ok {
			log.Printf("Error in Source Path: %s\n", srcPath)
			log.Printf("Unable to copy file %s after %d attempts. Ignoring file.\n", srcPath, tries)
			return
		}

		defer srcFile.Close()

		//Before the next step, create the path if it doesn't exist
		newPath := path.Dir(dstPath)
		os.MkdirAll(newPath, os.ModePerm)

		//Create the destination file
		dstFile, err := os.Create(dstPath)
		if err != nil {
			log.Println("Error in Destination Path")
			log.Fatal(err)
		}
		defer dstFile.Close()

		//Copy the file
		srcFile.WriteTo(dstFile)

		// flush in-memory copy
		err = dstFile.Sync()
		if err != nil {
			log.Fatal(err)
		}

		*count--
		log.Printf("Successfully copied %s into %s. Files left: %d", srcPath, dstPath, *count)
	} else {
		log.Printf("There was an error trying to copy file %s. Filename validation error.", srcPath)
	}
}

//backup will run asynchronously
//The task will connect to remote computer
//and backup the logs from the Path list.
func backup(server LogServer, wg *sync.WaitGroup, count *int) {
	var copyWg sync.WaitGroup //Create an async copy job

	defer wg.Done()

	log.Printf("Connecting to %s", server.Host)

	//First, let's config the connetion.
	config := &ssh.ClientConfig{
		User: server.User,
		HostKeyCallback: func(hostname string, remote net.Addr, key ssh.PublicKey) error {
			return nil
		},
		Auth: []ssh.AuthMethod{
			ssh.Password(server.Pass),
		},
	}
	config.SetDefaults()

	//Now, try to connect to the remote machine.
	sshConn, err := ssh.Dial("tcp", server.Host, config)
	if err != nil {
		panic("Failed to dial: " + err.Error())
	}
	defer sshConn.Close()

	log.Printf("Connected to %s", server.Host)

	//Start a new SFTP Session over the existing ssh connection.
	client, err := sftp.NewClient(sshConn)
	if err != nil {
		panic(err)
	}
	defer client.Close()

	//Check all the files to copy
	for _, sPath := range server.Paths {
		wlk := client.Walk(sPath)

		left := 0
		for wlk.Step() {
			if wlk.Err() != nil {
				continue
			}
			//If the process is reading a file, it should create a copy to local
			if !wlk.Stat().IsDir() {
				*count++
				left++

				copyWg.Add(1)                                                                //Add a new Task
				go copyFile(sPath, server.GetSavePath(), wlk.Path(), client, &copyWg, &left) //Run the Task
			}
		}
	}

	copyWg.Wait()
}

//BackupFromServer will try to get all the logs from the remote machine and process everything.
func (f *LogBackup) BackupFromServer() {
	var serverWg sync.WaitGroup //Async donwload

	lstServers := f.GetServerList() //Get the list of servers.

	//Metrics
	start := time.Now()
	count := 0

	//Generates the Async Fetching.
	for _, server := range lstServers {
		serverWg.Add(1)                     //Add a new Task
		go backup(server, &serverWg, &count) //Run the Task
	}

	serverWg.Wait() //Wait until the task is done.

	//Metrics
	finish := time.Since(start)
	log.Println("=============================================================================")
	log.Printf("Files copied successfuly. Total: %d files and took %s seconds to process.\n", count, finish)
	log.Println("=============================================================================")
}
