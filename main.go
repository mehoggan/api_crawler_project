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

func main() {
	rootURL := flag.String("root", "", "Target API documentation root URL to crawl (required)")
	pattern := flag.String("pattern", "", "Custom grep-like regex to search page content for "+
		"(default: built-in issue keywords). Case-insensitive.")
	maxDepth := flag.Int("max-depth", 3, "Maximum link-follow depth (default: 3)")
	concurrency := flag.Int("concurrency", 5, "Number of parallel workers (default: 5)")
	flag.Parse()

	if *rootURL == "" {
		fmt.Fprintln(os.Stderr, "Error: --root is required")
		flag.Usage()
		os.Exit(1)
	}

	if *maxDepth < 1 {
		fmt.Fprintf(os.Stderr, "Error: --max-depth must be >= 1 (got %d)\n", *maxDepth)
		os.Exit(1)
	}
	if *concurrency < 1 {
		fmt.Fprintf(os.Stderr, "Error: --concurrency must be >= 1 (got %d)\n", *concurrency)
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
		MaxDepth:    *maxDepth,
		Concurrency: *concurrency,
		Client:      client,
		IssueRegex:  issueRegex,
		LinkRegex:   linkRegex,
	}

	c := crawler.New(cfg)

	fmt.Printf("Starting crawl at: %s\n\n", *rootURL)
	c.Start()
}
