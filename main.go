package main

import (
	"crypto/tls"
	"fmt"
	"net/http"
	"regexp"
	"time"

	"apicrawler/crawler"
)

const (
	StartURL    = "https://example.com/api/docs" // Target API documentation root
	MaxDepth    = 3                               // How deep to follow links
	Concurrency = 5                               // Number of parallel workers
)

func main() {
	// Define patterns: Target issues (e.g., deprecated, legacy, known bug, 4xx/5xx)
	issueRegex := regexp.MustCompile(`(?i)(deprecated|known issue|security vulnerability|legacy endpoint|todo)`)
	// Extractor for absolute and relative href links
	linkRegex := regexp.MustCompile(`href="([^"]+)"`)

	transport := &http.Transport{
		TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
	}
	client := &http.Client{
		Timeout:   10 * time.Second,
		Transport: transport,
	}

	cfg := crawler.Config{
		StartURL:    StartURL,
		MaxDepth:    MaxDepth,
		Concurrency: Concurrency,
		Client:      client,
		IssueRegex:  issueRegex,
		LinkRegex:   linkRegex,
	}

	c := crawler.New(cfg)

	fmt.Printf("Starting crawl at: %s\n\n", StartURL)
	c.Start()
}
