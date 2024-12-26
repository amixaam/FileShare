package main

import (
	"embed"
	"fmt"
	"log"
	"net/http"
	"os"

	"fileshare/config"
	"fileshare/handlers"
	"fileshare/utils"
)

//go:embed templates/* static/*
var contentFS embed.FS

func main() {
    cfg, err := config.LoadConfig()
    if err != nil {
        log.Fatal("Error loading config:", err)
    }

    if err := os.MkdirAll(cfg.Directory, 0755); err != nil {
        log.Fatal("Error creating directory:", err)
    }

    if _, err := os.Stat(cfg.Directory); os.IsNotExist(err) {
        log.Fatalf("Directory does not exist: %s", cfg.Directory)
    }

    server := handlers.NewServer(contentFS, cfg.Directory, cfg.ShowDotfiles, cfg.Domain)
    
    // Set up handlers
    http.HandleFunc("/static/", server.HandleStatic)
    http.HandleFunc("/zip/", server.HandleZip)
    http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
        log.Printf("%s %s %s", r.RemoteAddr, r.Method, r.URL)
        server.HandleFileServer(w, r)
    })

    // Get local IP addresses
    ips, err := utils.GetLocalIPs()
    if err != nil {
        log.Printf("Warning: Could not determine local IP addresses: %v", err)
    }

    log.Printf("Serving files from: %s", cfg.Directory)
    log.Printf("Local addresses:")
    log.Printf("  http://localhost:%s", cfg.Port)
    for _, ip := range ips {
        log.Printf("  http://%s:%s", ip, cfg.Port)
    }

    log.Fatal(http.ListenAndServe(fmt.Sprintf(":%s", cfg.Port), nil))
}