package main

import (
	"crypto/tls"
	"flag"
	"fmt"
	"net/http"
	"os"
	"regexp"
	"time"

	"apicrawler/crawler"
)

const (
	MaxDepth    = 3 // How deep to follow links
	Concurrency = 5 // Number of parallel workers
)

func main() {
	rootURL := flag.String("root", "", "Target API documentation root URL to crawl (required)")
	flag.Parse()

	if *rootURL == "" {
		fmt.Fprintln(os.Stderr, "Error: --root is required")
		flag.Usage()
		os.Exit(1)
	}

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
		StartURL:    *rootURL,
		MaxDepth:    MaxDepth,
		Concurrency: Concurrency,
		Client:      client,
		IssueRegex:  issueRegex,
		LinkRegex:   linkRegex,
	}

	c := crawler.New(cfg)

	fmt.Printf("Starting crawl at: %s\n\n", *rootURL)
	c.Start()
}
