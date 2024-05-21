package sftp

import (
    "log"
    "os"
    "path/filepath"
    "strings"

    "github.com/pkg/sftp"
)

func UploadDirectory(client *sftp.Client, localPath, remotePath, tempRemotePath, fileExtensions, newExtension string, logStream *log.Logger, uploadedFiles *[]string, uploadRootOnly, useTempFolder bool) error {
    allowedExtensions := strings.Split(fileExtensions, ",")

    files, err := os.ReadDir(localPath)
    if err != nil {
        return err
    }

    for _, file := range files {
        if file.IsDir() && !uploadRootOnly {
            subDir := filepath.Join(localPath, file.Name())
            remoteSubDir := filepath.Join(remotePath, file.Name())
            err = UploadDirectory(client, subDir, remoteSubDir, tempRemotePath, fileExtensions, newExtension, logStream, uploadedFiles, uploadRootOnly, useTempFolder)
            if err != nil {
                return err
            }
        } else {
            ext := strings.TrimPrefix(filepath.Ext(file.Name()), ".")
            if fileExtensions == "" || contains(allowedExtensions, ext) {
                localFile := filepath.Join(localPath, file.Name())
                remoteFile := filepath.Join(remotePath, file.Name())

                err := uploadFile(client, localFile, remoteFile, tempRemotePath, newExtension, logStream, uploadedFiles, useTempFolder)
                if err != nil {
                    logStream.Printf("Error uploading file %s: %s", localFile, err)
                }
            }
        }
    }
    return nil
}

func uploadFile(client *sftp.Client, localFile, remoteFile, tempRemotePath, newExtension string, logStream *log.Logger, uploadedFiles *[]string, useTempFolder bool) error {
    srcFile, err := os.Open(localFile)
    if err != nil {
        return err
    }
    defer srcFile.Close()

    if useTempFolder {
        if err := client.MkdirAll(tempRemotePath); err != nil {
            return err
        }
        remoteFile = filepath.Join(tempRemotePath, filepath.Base(localFile))
    }

    dstFile, err := client.Create(remoteFile)
    if err != nil {
        return err
    }
    defer dstFile.Close()

    if _, err := dstFile.ReadFrom(srcFile); err != nil {
        return err
    }

    logStream.Printf("Uploaded %s to %s", localFile, remoteFile)
    *uploadedFiles = append(*uploadedFiles, localFile)

    if useTempFolder {
        // Rename file to change extension
        remoteFileWithNewExtension := strings.TrimSuffix(remoteFile, filepath.Ext(remoteFile)) + newExtension
        err := client.Rename(remoteFile, remoteFileWithNewExtension)
        if err != nil {
            return err
        }
        finalRemoteFile := filepath.Join(filepath.Dir(remoteFileWithNewExtension), filepath.Base(remoteFileWithNewExtension))
        err = client.Rename(remoteFileWithNewExtension, finalRemoteFile)
        if err != nil {
            return err
        }
        logStream.Printf("Moved %s to %s", remoteFileWithNewExtension, finalRemoteFile)
    }
    return nil
}

func contains(slice []string, item string) bool {
    for _, a := range slice {
        if a == item {
            return true
        }
    }
    return false
}
