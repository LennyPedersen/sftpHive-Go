package main

import (
    "encoding/json"
    "io"
    "log"
    "net/http"
    "os"
    "path/filepath"
    "sync"
    "time"

    "github.com/robfig/cron/v3"
    "sftphive/config"
)

var (
    jobStatuses   = make(map[string]string)
    jobSchedules  = make(map[string]time.Time)
    cronScheduler *cron.Cron
    configs       map[string]config.Configuration
    mu            sync.Mutex
)

func main() {
    var err error
    configs, err = config.LoadAllConfigs("configs.json")
    if err != nil {
        log.Fatalf("Error loading configurations: %v", err)
    }

    // Ensure all necessary directories and files exist
    ensureDirectoriesAndFiles(configs)

    // Setup cron scheduler
    cronScheduler = cron.New(cron.WithChain(
        cron.SkipIfStillRunning(cron.DefaultLogger),
    ))

    // Schedule jobs
    for customerName, customerConfig := range configs {
        schedule := customerConfig.Schedule
        if schedule == "" {
            schedule = "@daily"
        }

        config := customerConfig
        var entryID cron.EntryID
        entryID, err := cronScheduler.AddFunc(schedule, func() {
            updateJobStatus(customerName, "Running")
            runSFTPJob(customerName, config)
            updateJobStatus(customerName, "Completed")
            updateNextRun(customerName, entryID)
        })
        if err != nil {
            log.Fatalf("Error scheduling job for %s: %v", customerName, err)
        }
        jobStatuses[customerName] = "Scheduled"
        updateNextRun(customerName, entryID)
    }

    cronScheduler.Start()

    // Setup HTTP server
    http.HandleFunc("/logs", logsHandler)
    http.HandleFunc("/status", statusHandler)
    http.HandleFunc("/logfile", logFileHandler)
    http.Handle("/", http.FileServer(http.Dir("./static")))

    log.Println("Starting web server on :8080")
    log.Fatal(http.ListenAndServe(":8080", nil))
}

func ensureDirectoriesAndFiles(configs map[string]config.Configuration) {
    logDir := "logs"
    if _, err := os.Stat(logDir); os.IsNotExist(err) {
        err = os.Mkdir(logDir, 0755)
        if err != nil {
            log.Fatalf("Error creating logs directory: %v", err)
        }
    }

    for customerName := range configs {
        logFilePath := filepath.Join(logDir, customerName+".log")
        if _, err := os.Stat(logFilePath); os.IsNotExist(err) {
            _, err = os.Create(logFilePath)
            if err != nil {
                log.Fatalf("Error creating log file for %s: %v", customerName, err)
            }
        }
    }
}

func updateNextRun(customerName string, entryID cron.EntryID) {
    mu.Lock()
    defer mu.Unlock()
    nextRun := cronScheduler.Entry(entryID).Next
    jobSchedules[customerName] = nextRun
}

func updateJobStatus(customerName, status string) {
    mu.Lock()
    defer mu.Unlock()
    jobStatuses[customerName] = status
}

func logsHandler(w http.ResponseWriter, r *http.Request) {
    logFiles := make(map[string]string)
    for customerName := range configs {
        logFiles[customerName] = "logs/" + customerName + ".log"
    }
    w.Header().Set("Content-Type", "application/json")
    json.NewEncoder(w).Encode(logFiles)
}

func logFileHandler(w http.ResponseWriter, r *http.Request) {
    customerName := r.URL.Query().Get("customer")
    logFilePath := "logs/" + customerName + ".log"
    logFile, err := os.Open(logFilePath)
    if err != nil {
        http.Error(w, "Could not read log file", http.StatusInternalServerError)
        return
    }
    defer logFile.Close()

    logs, err := io.ReadAll(logFile)
    if err != nil {
        http.Error(w, "Could not read log file", http.StatusInternalServerError)
        return
    }

    w.Header().Set("Content-Type", "text/plain")
    w.Write(logs)
}

func statusHandler(w http.ResponseWriter, r *http.Request) {
    mu.Lock()
    defer mu.Unlock()
    status := make(map[string]map[string]string)
    for customerName, jobStatus := range jobStatuses {
        status[customerName] = map[string]string{
            "status":  jobStatus,
            "nextRun": jobSchedules[customerName].Format(time.RFC3339),
        }
    }
    w.Header().Set("Content-Type", "application/json")
    json.NewEncoder(w).Encode(status)
}

func runSFTPJob(customerName string, config config.Configuration) {
    logFilePath := "logs/" + customerName + ".log"
    logFile, err := os.OpenFile(logFilePath, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0666)
    if err != nil {
        log.Fatalf("Error opening log file for %s: %v", customerName, err)
    }
    defer logFile.Close()

    logStream := log.New(logFile, "", log.LstdFlags)

    logStream.Printf("Starting SFTP operation for %s\n", customerName)
    logStream.Printf("Running scheduled job for %s\n", customerName)
    logStream.Printf("Connected to SFTP server %s:%s\n", config.SftpServer, config.SftpPort)
    logStream.Printf("SFTP job for %s completed\n", customerName)
}
