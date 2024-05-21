package config

import (
    "encoding/json"
    "io"
    "os"
)

type Configuration struct {
    LocalPath                    string
    RemotePath                   string
    SftpServer                   string
    SftpPort                     string
    SftpUserName                 string
    SftpPassword                 string
    LogFilePath                  string
    ArchivePath                  string
    DeleteFoldersAfterArchive    bool
    FileExtensions               string
    CleanupThresholdDays         int
    UploadRootOnly               bool
    TempRemotePath               string
    UseTempFolder                bool
    NewExtension                 string
    DownloadEnabled              bool
    DownloadRemotePath           string
    DownloadLocalPath            string
    DownloadFileExtensions       string
    DownloadRootOnly             bool
    DeleteRemoteFileAfterDownload bool
    Schedule                     string // Add a Schedule field for cron jobs
}

func LoadAllConfigs(filePath string) (map[string]Configuration, error) {
    file, err := os.Open(filePath)
    if err != nil {
        return nil, err
    }
    defer file.Close()

    var configs map[string]Configuration
    bytes, err := io.ReadAll(file)
    if err != nil {
        return nil, err
    }

    err = json.Unmarshal(bytes, &configs)
    if err != nil {
        return nil, err
    }

    return configs, nil
}

func LoadConfig(customerName string) (Configuration, error) {
    configs, err := LoadAllConfigs("configs.json")
    if err != nil {
        return Configuration{}, err
    }

    config, ok := configs[customerName]
    if !ok {
        return Configuration{}, os.ErrNotExist
    }
    return config, nil
}
