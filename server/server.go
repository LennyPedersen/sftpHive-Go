package main

import (
    "encoding/json"
    "io"
    "log"
    "net/http"
    "os"
    "sync"

    "github.com/robfig/cron/v3"
    "sftphive/config"
)

var (
    jobStatuses   = make(map[string]string)
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

    // Setup cron scheduler
    cronScheduler = cron.New()

    // Schedule jobs
    for customerName, customerConfig := range configs {
        schedule := customerConfig.Schedule
        if schedule == "" {
            schedule = "@daily"
        }

        jobStatuses[customerName] = "Scheduled"
    }

    cronScheduler.Start()

    // Setup HTTP server
    http.HandleFunc("/logs", logsHandler)
    http.HandleFunc("/status", statusHandler)
    http.Handle("/", http.FileServer(http.Dir("./static")))

    log.Println("Starting web server on :8080")
    log.Fatal(http.ListenAndServe(":8080", nil))
}

func logsHandler(w http.ResponseWriter, r *http.Request) {
    logFilePath := "service.log"
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
    w.Header().Set("Content-Type", "application/json")
    json.NewEncoder(w).Encode(jobStatuses)
}
