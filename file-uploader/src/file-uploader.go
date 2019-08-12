package main

import (
	"flag"
	"log"
	"net"
	"os"
	"path"
	"path/filepath"
	"strconv"
	"sync"
	"fmt"

	"github.com/pkg/sftp"
	"golang.org/x/crypto/ssh"
)

//Data contains the values of the parameters.
type Data struct {
	SslFlag    bool
	SslKeyPath string
	Host       string
	Port       string
	User       string
	Pass       string
	DirOutput  string
	Files      []string
}

//Global variable to use Data.
var data Data

//The init function will check the parameters.
func init() {
	data = Data{}

	//Setting the flags
	sslFlag := flag.Bool("s", false, "Check if you need to work with SSL")
	sslKey := flag.String("k", "/path/to/rsa/key", "Set the key path.")
	host := flag.String("H", "localhost", "Set the remote server to connect.")
	port := flag.Int("P", 22, "Set the remote port to connect.")
	user := flag.String("U", "username", "Set the username of the remote server.")
	pass := flag.String("W", "password", "Set the password of the remote server.")
	output := flag.String("O", "/", "Set the path to put the files.")

	//Reading flag and arguments
	flag.Parse()
	args := flag.Args()

	data.SslFlag = *sslFlag
	data.SslKeyPath = *sslKey
	if len(*sslKey) < 1 {
		log.Panic("Set the path for the Key")
	}

	data.Host = *host
	data.Port = strconv.Itoa(*port)
	data.User = *user
	data.Pass = *pass
	data.DirOutput = *output

	if len(args) < 1 {
		log.Panic("Please set the file(s) to transfer.")
	}

	data.Files = args
}

func uploadFile(srcPath, dstPath string, c *sftp.Client, wg *sync.WaitGroup) {
	defer wg.Done()

	log.Printf("[%s] Copying into %s", srcPath, dstPath)

	var dstFile *sftp.File
	var err error

	//Get the name of the file
	fileName := filepath.Base(srcPath)

	//Before the next step, check if the src file doesn't exist
	srcFile, err := os.Open(srcPath)
	if err != nil {
		log.Printf("[%s] Error in Source Path.\n", srcPath)
		log.Println(err)
	}
	defer srcFile.Close()

	//Create the path, if it doesn't exist.
	c.MkdirAll(dstPath)

	//Create the new File (or use the existing one)
	copyTo := path.Join(dstPath, fileName)
	dstFile, err = c.Create(copyTo)

	if err != nil {
		log.Printf("[%s] Error in Destination Path.\n", copyTo)
		log.Println(err)
	}

	defer dstFile.Close()

	//Copy the file
	dstFile.ReadFrom(srcFile)

	// Check if there was an error.
	_, err = c.Lstat(copyTo)
	if err != nil {
		log.Printf("[%s] Error validating the upload.\n", copyTo)
		log.Fatal(err)
	} else {
		log.Printf("[%s] Successfully uploaded to %s.", srcPath, dstPath)
	}
}

func (d *Data) getPasswordKey() (ssh.ClientConfig, error) {
	var sshConfig *ssh.ClientConfig

	sshConfig = &ssh.ClientConfig{
		User: d.User,
		Auth: []ssh.AuthMethod{ssh.Password(d.Pass)},
		HostKeyCallback: func(hostname string, remote net.Addr, key ssh.PublicKey) error {
			return nil
		},
	}

	return *sshConfig, nil
}

func main() {
	var clientConfig ssh.ClientConfig

	//TODO: Implement the SSL way.
	clientConfig, _ = data.getPasswordKey()

	//Now, try to connect to the remote machine.
	url := fmt.Sprintf("%s:%s", data.Host, data.Port)
	sshConn, err := ssh.Dial("tcp", url, &clientConfig)
	if err != nil {
		panic("Failed to dial: " + err.Error())
	}
	defer sshConn.Close()

	log.Printf("Connected to %s", url)

	//Start a new SFTP Session over the existing ssh connection.
	client, err := sftp.NewClient(sshConn)
	if err != nil {
		panic(err)
	}
	defer client.Close()

	var readerWg sync.WaitGroup //Async file reading from zip file.

	for _, file := range data.Files {
		readerWg.Add(1)

		log.Printf("[%s] Transfering file.", file)
		go uploadFile(file, data.DirOutput, client, &readerWg)
	}

	readerWg.Wait()
}
