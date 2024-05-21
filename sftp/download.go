package sftp

import (
    "log"
    "os"
    "path/filepath"
    "strings"

    "github.com/pkg/sftp"
)

func DownloadDirectory(client *sftp.Client, localPath, remotePath, fileExtensions string, logStream *log.Logger, downloadedFiles *[]string, downloadRootOnly, deleteAfterDownload bool) error {
    extensions := strings.Split(fileExtensions, ",")

    files, err := client.ReadDir(remotePath)
    if err != nil {
        return err
    }

    for _, file := range files {
        if file.IsDir() && !downloadRootOnly {
            subDir := filepath.Join(localPath, file.Name())
            remoteSubDir := filepath.Join(remotePath, file.Name())
            err = DownloadDirectory(client, subDir, remoteSubDir, fileExtensions, logStream, downloadedFiles, downloadRootOnly, deleteAfterDownload)
            if err != nil {
                return err
            }
        } else {
            ext := strings.TrimPrefix(filepath.Ext(file.Name()), ".")
            if fileExtensions == "" || contains(extensions, ext) {
                localFile := filepath.Join(localPath, file.Name())
                remoteFile := filepath.Join(remotePath, file.Name())

                err := downloadFile(client, localFile, remoteFile, logStream, downloadedFiles, deleteAfterDownload)
                if err != nil {
                    logStream.Printf("Error downloading file %s: %s", remoteFile, err)
                }
            }
        }
    }
    return nil
}

func downloadFile(client *sftp.Client, localFile, remoteFile string, logStream *log.Logger, downloadedFiles *[]string, deleteAfterDownload bool) error {
    dstFile, err := os.Create(localFile)
    if err != nil {
        return err
    }
    defer dstFile.Close()

    srcFile, err := client.Open(remoteFile)
    if err != nil {
        return err
    }
    defer srcFile.Close()

    if _, err := srcFile.WriteTo(dstFile); err != nil {
        return err
    }

    logStream.Printf("Downloaded %s to %s", remoteFile, localFile)
    *downloadedFiles = append(*downloadedFiles, remoteFile)

    if deleteAfterDownload {
        err := client.Remove(remoteFile)
        if err != nil {
            logStream.Printf("Failed to delete %s from remote server: %s", remoteFile, err)
            return err
        }
        logStream.Printf("Deleted %s from remote server", remoteFile)
    }
    return nil
}
