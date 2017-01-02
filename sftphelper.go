package sftphelper

import (
	"fmt"
	"github.com/pkg/sftp"
	"golang.org/x/crypto/ssh"
	"io/ioutil"
	"log"
	"os"
	"os/user"
	"path/filepath"
	"strings"
)

// External packages to be downloaded
// go get golang.org/x/crypto/ssh
// go get github.com/pkg/sftp

// SFTPConnection : Main connection
type SFTPConnection struct {
	SftpClient *sftp.Client
	IsVerbose  bool
}

// ConnectWithKeyFile : This function will get an sftp conenction using the private key file
// Returns client ssh connection handler and error object if fails
// Host - ftp server name with or without port
// Username - ftp user (e.g ec2-user)
// privkeyfile - Fully qualified Private key file name
func ConnectWithKeyFile(host, username, privkeyfile string) (*SFTPConnection, error) {

	//Check Port if not passed with host
	if strings.Contains(host, ":") == false {
		host = host + ":22"
	}

	// Get the key parsed
	buf, err := ioutil.ReadFile(privkeyfile)
	if err != nil {
		return nil, fmt.Errorf("Failed to read keyfile: %v", err)
	}
	signer, err := ssh.ParsePrivateKey(buf)
	if err != nil {
		return nil, fmt.Errorf("Failed to parse private key: %v", err)
	}

	// Define the Client Config
	config := &ssh.ClientConfig{
		User: username,
		Auth: []ssh.AuthMethod{
			ssh.PublicKeys(signer),
		},
	}

	// Dial to the FTP server
	client, err := ssh.Dial("tcp", host, config)
	if err != nil {
		return nil, fmt.Errorf("Failed to dial: %v", err)
	}

	// Establish SFTP Connection
	sftp, err := sftp.NewClient(client)
	if err != nil {
		return nil, fmt.Errorf("Failed to create ftp session: %v", err)
	}

	sftpConnection := &SFTPConnection{}
	sftpConnection.SftpClient = sftp
	sftpConnection.IsVerbose = false

	return sftpConnection, nil
}

// ConnectWithPassword : This function will get an sftp conenction using a password
// Returns client ssh connection handler and error object if fails
// Host - sftp server name with or without port
// Username - sftp user (e.g ec2-user)
// password - password for sftp user
func ConnectWithPassword(host, username, password string) (*SFTPConnection, error) {

	//Check Port if not passed with host
	if strings.Contains(host, ":") == false {
		host = host + ":22"
	}

	config := &ssh.ClientConfig{
		User:            username,
		HostKeyCallback: nil,
		Auth: []ssh.AuthMethod{
			ssh.Password(password),
		},
	}

	// Dial to the FTP server
	client, err := ssh.Dial("tcp", host, config)
	if err != nil {
		return nil, fmt.Errorf("Failed to dial: %v", err)
	}

	// Establish SFTP Connection
	sftp, err := sftp.NewClient(client)
	if err != nil {
		return nil, fmt.Errorf("Failed to create ftp session: %v", err)
	}

	sftpConnection := &SFTPConnection{}
	sftpConnection.SftpClient = sftp
	sftpConnection.IsVerbose = false

	return sftpConnection, nil
}

// Close : Close the SFTP Session
func (sftpConnection *SFTPConnection) Close() error {
	err := sftpConnection.SftpClient.Close()
	if err != nil {
		fmt.Println(err)
		return err
	}
	sftpConnection.logMsg("Closed SFTP Connection")
	return nil
}

// DownloadFiles : This function will download all files from ftp_server folder to the folder including sub-folders
// Return: error if fails
// source - remote Source directory
// target - local TargetDir
func (sftpConnection *SFTPConnection) DownloadFiles(source string, target string) error {

	sftpClient := sftpConnection.SftpClient
	if sftpClient == nil {
		return fmt.Errorf("No valid SFTP Connection to DownloadFiles")
	}

	// Find if sourcePath is file or Directory
	remoteSrc, err := sftpClient.Stat(source)
	if err != nil {
		return err
	}

	if remoteSrc.IsDir() {

		walker := sftpClient.Walk(source)
		for walker.Step() {
			if walker.Err() != nil {
				log.Println("walker.Err", walker.Err())
				continue
			} // Skip if there is an error with the file

			item := walker.Stat()

			itemPath := walker.Path()

			if item.IsDir() {

				subPath := strings.Replace(itemPath, source, "", 1)
				dest := filepath.Join(target, subPath)

				// Check  if destination has a directory or create
				if _, err := os.Stat(dest); os.IsNotExist(err) {
					sftpConnection.logMsg("Creating Directory:" + dest)
					err = os.MkdirAll(dest, item.Mode())
					if err != nil {
						return err
					}
				}

			} else {
				itemName := item.Name()
				subPath := filepath.Dir(itemPath)

				destPath := strings.Replace(subPath, source, "", 1)
				dest := filepath.Join(target, destPath)

				err = writeFile(sftpClient, itemPath, dest, itemName)
				if err != nil {
					fmt.Println("Error writing  file :" + itemName)
				}
				sftpConnection.logMsg("Copied %s to %s", itemPath, dest)
			}
		}

	} else {
		err = writeFile(sftpClient, source, target, remoteSrc.Name())
		if err != nil {
			return err
		}
		sftpConnection.logMsg("Copied %s to %s", source, target)
	}

	return nil
}

//CallbackFunc : Callback function template
type CallbackFunc func(sftpConnection *SFTPConnection, filePath string, fileName string)

// WalkDirectories : Iterates over all the inner directories and calls a processing function
func (sftpConnection *SFTPConnection) WalkDirectories(path string, callback CallbackFunc) error {

	sftpClient := sftpConnection.SftpClient
	walker := sftpClient.Walk(path)
	for walker.Step() {
		if walker.Err() != nil {
			fmt.Println("walker.Err", walker.Err())
			continue
		} // Skip if there is an error with the file

		item := walker.Stat()
		if item.IsDir() {
			itemName := item.Name()
			itemPath := walker.Path()
			callback(sftpConnection, itemPath, itemName)
		}
	}
	return nil
}

// WalkFiles : Iterates over all the inner files in the remote directory and calls a processing function
func (sftpConnection *SFTPConnection) WalkFiles(path string, callback CallbackFunc) error {
	sftpClient := sftpConnection.SftpClient
	walker := sftpClient.Walk(path)
	for walker.Step() {
		if walker.Err() != nil {
			fmt.Println("walker.Err", walker.Err())
			continue
		} // Skip if there is an error with the file

		item := walker.Stat()
		if item.IsDir() {
			continue
		} // Skip Directories

		fileName := item.Name()
		filePath := walker.Path()

		callback(sftpConnection, filePath, fileName)
	}
	return nil
}

func (sftpConnection *SFTPConnection) logMsg(format string, v ...interface{}) {
	if sftpConnection.IsVerbose {
		log.Printf(format+"\n", v...)
	}
}

//UploadFile : Copy source file to target
func (sftpConnection *SFTPConnection) UploadFile(sourcePath, targetPath string) error {

	sftpClient := sftpConnection.SftpClient

	targetFile, err := sftpClient.Create(targetPath)
	if err != nil {
		return err
	}

	// Open the source file
	srcFile, err := sftpClient.Open(sourcePath)
	if err != nil {
		return err
	}
	defer srcFile.Close()

	srcFile.WriteTo(targetFile)

	return nil
}

// RemoveFile : Removes a file from FTP Server
func (sftpConnection *SFTPConnection) RemoveFile(sourcePath string) error {

	sftpClient := sftpConnection.SftpClient

	if sftpClient != nil {
		err := sftpClient.Remove(sourcePath)
		return err
	}
	return fmt.Errorf("No  valid SFTP Connection to delete file")
}

// Write the source file into the target folder with the fileName
func writeFile(sftpClient *sftp.Client, source string, target string, fileName string) error {

	// Open the Source file
	srcFile, err := sftpClient.Open(source)
	if err != nil {
		return err
	}
	defer srcFile.Close()

	var destFileName string

	dest, _ := os.Stat(target)
	if dest.IsDir() {
		destFileName = filepath.Join(target, fileName)
	} else {
		destFileName = target
	}

	// Create the destination file
	dstFile, err := os.Create(destFileName)
	if err != nil {
		return err
	}
	defer dstFile.Close()
	// Copy the file
	srcFile.WriteTo(dstFile)

	return nil
}

// UserHomeDir : Gets Users Home Directory
func UserHomeDir() (homedir string) {
	usr, _ := user.Current()
	homedir = usr.HomeDir
	return
}

// KeyFilePath : Gets Key File Path - User Home directory + .ssh + key file name
func KeyFilePath(filename string) (path string) {
	path = filepath.Join(UserHomeDir(), ".ssh", filename)
	_, err := os.Stat(path)
	if os.IsNotExist(err) {
		return ""
	}
	return
}
