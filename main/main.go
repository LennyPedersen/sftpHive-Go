package main

import (
    "crypto/aes"
    "crypto/cipher"
    "encoding/base64"
    "errors"
    "flag"
    "fmt"
    "log"
    "os"
    "path/filepath"
    "strconv"
    "time"

    "sftphive/config"
    "sftphive/sftp"

    "github.com/robfig/cron/v3"
)

var jobStatuses = make(map[string]string)

func main() {
    // Define flags
    customerName := flag.String("customer", "", "Customer name to run the SFTP job for")
    skipScheduler := flag.Bool("skip-scheduler", false, "Run only for the specified customer and skip the scheduler")
    flag.Parse()

    if *skipScheduler && *customerName == "" {
        fmt.Println("Please specify a customer name when using the --skip-scheduler flag")
        return
    }

    if *skipScheduler {
        runSingleCustomer(*customerName)
    } else {
        runWithScheduler()
    }
}

func runSingleCustomer(customerName string) {
    config, err := config.LoadConfig(customerName)
    if err != nil {
        log.Fatalf("Error loading configuration: %v", err)
    }

    // Setup logging
    logFileName := fmt.Sprintf("%s.log", time.Now().Format("2006-01-02"))
    logFilePath := filepath.Join(config.LogFilePath, logFileName)
    logFile, err := os.OpenFile(logFilePath, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0666)
    if err != nil {
        log.Fatalf("Error opening log file: %v", err)
    }
    defer logFile.Close()

    logStream := log.New(logFile, "", log.LstdFlags)
    logStream.Printf("Starting SFTP operation for %s", customerName)

    updateJobStatus(customerName, "Running")
    runSFTPJob(customerName, config, logStream)
    updateJobStatus(customerName, "Completed")
}

func runWithScheduler() {
    // Load the entire configuration
    configs, err := config.LoadAllConfigs("appsettings.json")
    if err != nil {
        log.Fatalf("Error loading configurations: %v", err)
    }

    // Setup logging
    logFilePath := "service.log"
    logFile, err := os.OpenFile(logFilePath, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0666)
    if err != nil {
        log.Fatalf("Error opening log file: %v", err)
    }
    defer logFile.Close()

    logStream := log.New(logFile, "", log.LstdFlags)
    logStream.Println("Starting SFTP service")

    // Create a new cron scheduler
    c := cron.New()

    // Schedule jobs for each customer
    for customerName, customerConfig := range configs {
        schedule := customerConfig.Schedule
        if schedule == "" {
            schedule = "@daily" // Default to daily if no schedule is provided
        }

        config := customerConfig // Capture range variable
        _, err := c.AddFunc(schedule, func() {
            updateJobStatus(customerName, "Running")
            runSFTPJob(customerName, config, logStream)
            updateJobStatus(customerName, "Completed")
        })
        if err != nil {
            logStream.Printf("Error scheduling job for %s: %v", customerName, err)
        }
        updateJobStatus(customerName, "Scheduled")
    }

    // Start the cron scheduler
    c.Start()

    // Keep the service running
    select {}
}

func runSFTPJob(customerName string, config config.Configuration, logStream *log.Logger) {
    // Decrypt the SFTP password
    key := "mysecretencryptionkey" // Must be 32 bytes long for AES-256
    sftpPassword, err := decrypt(config.SftpPassword, key)
    if err != nil {
        logStream.Printf("Failed to decrypt SFTP password for %s: %v", customerName, err)
        return
    }

    // Convert SftpPort from string to int
    sftpPort, err := strconv.Atoi(config.SftpPort)
    if err != nil {
        logStream.Printf("Invalid SftpPort for %s: %v", customerName, err)
        return
    }

    client, err := sftp.NewSFTPClient(config.SftpUserName, sftpPassword, config.SftpServer, sftpPort)
    if err != nil {
        logStream.Printf("Failed to connect to SFTP server for %s: %v", customerName, err)
        return
    }
    defer client.Close()

    uploadedFiles := []string{}
    downloadedFiles := []string{}

    if config.DownloadEnabled {
        err := sftp.DownloadDirectory(client, config.DownloadLocalPath, config.DownloadRemotePath, config.DownloadFileExtensions, logStream, &downloadedFiles, config.DownloadRootOnly, config.DeleteRemoteFileAfterDownload)
        if err != nil {
            logStream.Printf("Error downloading files for %s: %v", customerName, err)
        }
    } else {
        err := sftp.UploadDirectory(client, config.LocalPath, config.RemotePath, config.TempRemotePath, config.FileExtensions, config.NewExtension, logStream, &uploadedFiles, config.UploadRootOnly, config.UseTempFolder)
        if err != nil {
            logStream.Printf("Error uploading files for %s: %v", customerName, err)
        }

        err = moveFilesToArchive(config.ArchivePath, logStream, uploadedFiles)
        if err != nil {
            logStream.Printf("Error moving files to archive for %s: %v", customerName, err)
        }

        if config.DeleteFoldersAfterArchive {
            err = deleteArchivedFiles(logStream, uploadedFiles)
            if err != nil {
                logStream.Printf("Error deleting archived files for %s: %v", customerName, err)
            }
        }
    }

    err = sftp.CleanUpArchive(config.ArchivePath, config.CleanupThresholdDays, logStream)
    if err != nil {
        logStream.Printf("Error cleaning up archive for %s: %v", customerName, err)
    }

    logStream.Printf("SFTP job for %s completed", customerName)
}

func moveFilesToArchive(archivePath string, logStream *log.Logger, uploadedFiles []string) error {
    archivePathWithDate := filepath.Join(archivePath, time.Now().Format("2006-01-02"))

    if err := os.MkdirAll(archivePathWithDate, 0755); err != nil {
        return err
    }

    for _, file := range uploadedFiles {
        destFile := filepath.Join(archivePathWithDate, filepath.Base(file))
        if err := os.Rename(file, destFile); err != nil {
            logStream.Printf("Error moving file %s to archive: %v", file, err)
            return err
        }
        logStream.Printf("Moved file %s to archive %s", file, destFile)
    }
    return nil
}

func deleteArchivedFiles(logStream *log.Logger, archivedFiles []string) error {
    for _, file := range archivedFiles {
        if err := os.Remove(file); err != nil {
            logStream.Printf("Error deleting file %s: %v", file, err)
            return err
        }
        logStream.Printf("Deleted file %s", file)
    }
    return nil
}

func updateJobStatus(customerName, status string) {
    jobStatuses[customerName] = status
}

// Decrypt function
func decrypt(cipherText, key string) (string, error) {
    block, err := aes.NewCipher([]byte(key))
    if err != nil {
        return "", err
    }

    decodedCipherText, err := base64.URLEncoding.DecodeString(cipherText)
    if err != nil {
        return "", err
    }

    if len(decodedCipherText) < aes.BlockSize {
        return "", errors.New("cipherText too short")
    }

    iv := decodedCipherText[:aes.BlockSize]
    decodedCipherText = decodedCipherText[aes.BlockSize:]

    stream := cipher.NewCFBDecrypter(block, iv)
    stream.XORKeyStream(decodedCipherText, decodedCipherText)

    return string(decodedCipherText), nil
}
