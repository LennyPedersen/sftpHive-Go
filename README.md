# sftpHive

## Overview

sftpHive is a Go-based application for managing SFTP file transfers for multiple customers. It supports both scheduled and on-demand transfers, and includes a web server for monitoring job statuses and viewing logs.

## Directory Structure

```
sftphive/
│
├── config/
│ └── config.go
│
├── sftp/
│ ├── upload.go
│ ├── download.go
│ └── cleanup.go
│
├── static/
│ └── index.html
│
├── main/
│ └── main.go
├── server/
│ └── server.go
├── go.mod
└── configs.json
```


## Configuration

The `configs.json` file contains configurations for each customer. Below is an example configuration:

```json
{
    "customer1": {
        "LocalPath": "/local/path",
        "RemotePath": "/remote/path",
        "SftpServer": "sftp.example.com",
        "SftpPort": "22",
        "SftpUserName": "username",
        "SftpPassword": "password",
        "LogFilePath": "/path/to/logs",
        "ArchivePath": "/path/to/archive",
        "DeleteFoldersAfterArchive": true,
        "FileExtensions": "txt,csv",
        "CleanupThresholdDays": 90,
        "UploadRootOnly": false,
        "TempRemotePath": "/temp/path",
        "UseTempFolder": true,
        "NewExtension": ".bak",
        "DownloadEnabled": true,
        "DownloadRemotePath": "/download/path",
        "DownloadLocalPath": "/download/local/path",
        "DownloadFileExtensions": "txt,csv",
        "DownloadRootOnly": false,
        "DeleteRemoteFileAfterDownload": true,
        "Schedule": "@daily"
    }
}
```

## Running the Main Application

The main application handles SFTP job execution and scheduling. You can run it with or without the scheduler.

### Running with Scheduler
```bash
cd main
go run main.go
```

### Running for a Single Customer

```bash
cd main
go run main.go --customer <customerName> --skip-scheduler
```

Replace <customerName> with the name of the customer as specified in configs.json.

## Running the Web Server

The web server provides a UI for monitoring job statuses and viewing logs.

```bash
cd server
go run server.go
```

### Accessing the Web Interface

Open a web browser and navigate to http://localhost:8080 to view the status of the scheduled jobs and logs.

### Web Interface
The web interface uses Bootstrap for styling and jQuery for dynamic content updates. It provides a simple table to display job statuses and a section to view logs.

## Dependencies

Ensure you have the following Go packages installed:
```
github.com/robfig/cron/v3
github.com/pkg/sftp
```

## Building the Project

To build the project, you can use the Go build command:
```bash
go build -o sftphive main/main.go
```
```bash
go build -o sftphive-server server/server.go
```

This will create two executable files, sftphive and sftphive-server, which you can run to start the main application and the web server, respectively.