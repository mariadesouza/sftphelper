package main

import (
	"sftphelper"
	"fmt"
)


func main() {
	

	sftpConnection, err := sftphelper.ConnectWithKeyFile("test.sftp.com", "ec2-user", "/Users/me/.ssh/id_rsa_sftp.pem")
	if (err != nil){
		panic(err)
	}
	defer sftpConnection.Close()

	sftpConnection.IsVerbose = true
	
	sftpConnection.DownloadFiles("/home/ftptest/", "/Users/me/test")
	if (err != nil){
		fmt.Println(err)
	}

	sftpConnection.DownloadFiles("/home/ftptest/inner/scrapedfeed.json", "/Users/me/test/")
	if (err != nil){
		fmt.Println(err)
	}
	sftpConnection.DownloadFiles("/home/ftptest/inner/scrapedfeed.json", "/Users/me/test/inner/scrapedfeed.json")
	if (err != nil){
		fmt.Println(err)
	}
	sftpConnection.UploadFile( "/Users/me/test/inner/scrapedfeed.json","/home/ftptest/outbound/")
	if (err != nil){
		fmt.Println(err)
	}
	sftpConnection.RemoveFile( "/home/ftptest/outbound/scrapedfeed.json")
	if (err != nil){
		fmt.Println(err)
	}

	sftpConnection.WalkDirectories("/home/ftptest/", processDirectory)
	sftpConnection.WalkFiles("/home/ftptest/", processFiles)

}


func processDirectory(sftpConnection *sftphelper.SFTPConnection, itemPath, itemName string) {
	fmt.Println(itemName)
	fmt.Println(itemPath)
}

func processFiles(sftpConnection *sftphelper.SFTPConnection, itemPath, itemName string) {
	fmt.Println(itemName)
	fmt.Println(itemPath)
}