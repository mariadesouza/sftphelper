# sftphelper

This package provides an easy to use wrapper for the golang sftp package

## Exported Functions

### ConnectWithKeyFile

	Input:
		host string,
		username string,
		privkeyfile string
	Returns:
		SFTPConnection

### ConnectWithPassword

	Input:
		host string,
		username string,
		password string
	Returns:
		SFTPConnection

### type SFTPConnection 
	type SFTPConnection struct {
		SftpClient *sftp.Client
	IsVerbose  bool
	}

#### func (*SFTPConnection) DownloadFiles(source string, target string) error 
#### func (*SFTPConnection) UploadFile(sourcePath, targetPath string) error 
#### func (*SFTPConnection) RemoveFile(sourcePath string) error
#### func (*SFTPConnection) WalkDirectories(path string, callback CallbackFunc) error 
#### func (*SFTPConnection) WalkFiles(path string, callback CallbackFunc) error 
#### func (*SFTPConnection) Close() error 

## Example Usage

import ("sftphelper")

Connect With Key File: 
sftpConnection, err := sftphelper.ConnectWithKeyFile("test.sftpserver.com", "ec2-user", "/home/me/.ssh/sftpkeyfile.pem")

OR 

Connect With Password: 
sftpConnection, err := sftphelper.ConnectWithKeyFile("test.sftpserver.com", "ec2-user", "password1")

sftpConnection.DownloadFiles("/home/ftptest/sourcePath/", "/home/me/")

## Contributors

Maria DeSouza <maria.g.desouza@gmail.com>

