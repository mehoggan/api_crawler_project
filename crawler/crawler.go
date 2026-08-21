package crawler

import (
	"fmt"
	"io"
	"net/http"
	"net/url"
	"regexp"
	"strings"
	"sync"
)

// Config holds the setup variables for the crawler
type Config struct {
	StartURL    string
	MaxDepth    int
	Concurrency int
	Client      *http.Client
	IssueRegex  *regexp.Regexp
	LinkRegex   *regexp.Regexp
}

// Crawler tracks visited pages and handles sync routines
type Crawler struct {
	cfg     Config
	visited map[string]bool
	mu      sync.Mutex
	wg      sync.WaitGroup
	sem     chan struct{}
}

// New initializes a Crawler instance
func New(cfg Config) *Crawler {
	return &Crawler{
		cfg:     cfg,
		visited: make(map[string]bool),
		sem:     make(chan struct{}, cfg.Concurrency),
	}
}

// Start initiates the recursive concurrent crawling process
func (c *Crawler) Start() {
	c.wg.Add(1)
	go c.crawl(c.cfg.StartURL, 0)
	c.wg.Wait()
}

func (c *Crawler) crawl(targetURL string, depth int) {
	defer c.wg.Done()

	if depth > c.cfg.MaxDepth {
		return
	}

	c.mu.Lock()
	if c.visited[targetURL] {
		c.mu.Unlock()
		return
	}
	c.visited[targetURL] = true
	c.mu.Unlock()

	c.sem <- struct{}{}
	defer func() { <-c.sem }()

	resp, err := c.cfg.Client.Get(targetURL)
	if err != nil {
		return
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return
	}

	bodyBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		return
	}
	bodyStr := string(bodyBytes)

	if matches := c.cfg.IssueRegex.FindAllString(bodyStr, -1); len(matches) > 0 {
		fmt.Printf("[ISSUE FOUND] URL: %s -> Matches: %v\n", targetURL, uniqueSlice(matches))
	}

	foundLinks := c.cfg.LinkRegex.FindAllStringSubmatch(bodyStr, -1)
	baseURL, _ := url.Parse(c.cfg.StartURL)

	for _, match := range foundLinks {
		foundURL := match[1]
		parsedFound, err := url.Parse(foundURL)
		if err != nil {
			continue
		}

		resolved := baseURL.ResolveReference(parsedFound)
		resolved.Fragment = "" // fragments don't change what the server returns; avoid re-crawling the same page per-anchor
		resolvedURL := resolved.String()

		if parsedFound.Host == "" || strings.EqualFold(parsedFound.Host, baseURL.Host) {
			c.wg.Add(1)
			go c.crawl(resolvedURL, depth+1)
		}
	}
}

func uniqueSlice(slice []string) []string {
	keys := make(map[string]bool)
	list := []string{}
	for _, entry := range slice {
		lower := strings.ToLower(entry)
		if _, value := keys[lower]; !value {
			keys[lower] = true
			list = append(list, lower)
		}
	}
	return list
}
