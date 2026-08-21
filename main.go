package main

import (
	"apicrawler/crawler"
	"crypto/tls"
	"flag"
	"fmt"
	"net/http"
	"os"
	"regexp"
	"time"
)

const (
	MaxDepth    = 3 // How deep to follow links
	Concurrency = 5 // Number of parallel workers
)

func main() {
	rootURL := flag.String("root", "", "Target API documentation root URL to crawl (required)")
	pattern := flag.String("pattern", "", "Custom grep-like regex to search page content for "+
		"(default: built-in issue keywords). Case-insensitive.")
	flag.Parse()

	if *rootURL == "" {
		fmt.Fprintln(os.Stderr, "Error: --root is required")
		flag.Usage()
		os.Exit(1)
	}

	// Define patterns: Target issues (e.g., deprecated, legacy, known bug, 4xx/5xx)
	// or, if --pattern is supplied, whatever the caller wants to search for.
	searchPattern := `(?i)(deprecated|known issue|security vulnerability|legacy endpoint|todo)`
	if *pattern != "" {
		searchPattern = "(?i)(" + *pattern + ")"
	}
	issueRegex, err := regexp.Compile(searchPattern)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: invalid --pattern regex: %v\n", err)
		os.Exit(1)
	}
	// Extractor for absolute and relative href links
	linkRegex := regexp.MustCompile(`href="([^"]+)"`)

	transport := &http.Transport{
		//nolint:gosec // Docs crawler intentionally accepts servers with invalid TLS certificates.
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
