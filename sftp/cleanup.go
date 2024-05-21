package sftp

import (
    "log"
    "os"
    "path/filepath"
    "time"
)

func CleanUpArchive(archivePath string, cleanupDays int, logStream *log.Logger) error {
    cutoffDate := time.Now().AddDate(0, 0, -cleanupDays)
    err := filepath.Walk(archivePath, func(path string, info os.FileInfo, err error) error {
        if err != nil {
            return err
        }
        if info.ModTime().Before(cutoffDate) {
            if info.IsDir() {
                if err := os.RemoveAll(path); err != nil {
                    logStream.Printf("Error removing directory %s: %v", path, err)
                    return err
                }
                logStream.Printf("Removed directory %s", path)
            } else {
                if err := os.Remove(path); err != nil {
                    logStream.Printf("Error removing file %s: %v", path, err)
                    return err
                }
                logStream.Printf("Removed file %s", path)
            }
        }
        return nil
    })
    return err
}
