// Package crawler implements a concurrent web crawler that scans API documentation pages
// for configurable regex patterns and follows same-host links up to a maximum depth.
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

// maxPreviewLineLength truncates very long (e.g. minified) lines in the context preview.
const maxPreviewLineLength = 200

// Config holds the setup variables for the crawler.
type Config struct {
	StartURL     string
	MaxDepth     int
	Concurrency  int
	ContextLines int
	Client       *http.Client
	IssueRegex   *regexp.Regexp
	LinkRegex    *regexp.Regexp
}

// Crawler tracks visited pages and handles sync routines.
type Crawler struct {
	cfg     Config
	visited map[string]bool
	mu      sync.Mutex
	wg      sync.WaitGroup
	sem     chan struct{}
}

// New initializes a Crawler instance.
func New(cfg Config) *Crawler {
	return &Crawler{
		cfg:     cfg,
		visited: make(map[string]bool),
		sem:     make(chan struct{}, cfg.Concurrency),
	}
}

// Start initiates the recursive concurrent crawling process.
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
	defer func() { _ = resp.Body.Close() }()

	if resp.StatusCode != http.StatusOK {
		return
	}

	bodyBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		return
	}
	bodyStr := string(bodyBytes)

	if matchIndices := c.cfg.IssueRegex.FindAllStringIndex(bodyStr, -1); len(matchIndices) > 0 {
		matches := make([]string, 0, len(matchIndices))
		for _, idx := range matchIndices {
			matches = append(matches, bodyStr[idx[0]:idx[1]])
		}
		fmt.Printf("[ISSUE FOUND] URL: %s -> Matches: %v\n", targetURL, uniqueSlice(matches))
		fmt.Print(contextPreview(bodyStr, matchIndices, c.cfg.ContextLines))
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
		// Fragments don't change what the server returns; avoid re-crawling the same page per-anchor
		resolved.Fragment = ""
		resolvedURL := resolved.String()

		if parsedFound.Host == "" || strings.EqualFold(parsedFound.Host, baseURL.Host) {
			c.wg.Add(1)
			go c.crawl(resolvedURL, depth+1)
		}
	}
}

// contextPreview renders a few lines of context around each match in matchIndices so the
// output shows what was actually found on the page, not just the URL it was found on.
func contextPreview(body string, matchIndices [][]int, contextLines int) string {
	lines := strings.Split(body, "\n")
	lineStarts := make([]int, len(lines))
	offset := 0
	for i, line := range lines {
		lineStarts[i] = offset
		offset += len(line) + 1
	}

	var b strings.Builder
	for _, idx := range matchIndices {
		matchLine := lineForOffset(lineStarts, idx[0])
		start := matchLine - contextLines
		if start < 0 {
			start = 0
		}
		end := matchLine + contextLines
		if end >= len(lines) {
			end = len(lines) - 1
		}

		fmt.Fprintln(&b, "    ---")
		for i := start; i <= end; i++ {
			marker := "  "
			if i == matchLine {
				marker = "->"
			}
			fmt.Fprintf(&b, "    %s %4d: %s\n", marker, i+1, truncateLine(lines[i]))
		}
	}
	return b.String()
}

// lineForOffset returns the index of the line containing the given byte offset into the
// original (unsplit) body, given the byte offset each line in lineStarts starts at.
func lineForOffset(lineStarts []int, offset int) int {
	line := 0
	for i, start := range lineStarts {
		if start > offset {
			break
		}
		line = i
	}
	return line
}

func truncateLine(line string) string {
	if len(line) <= maxPreviewLineLength {
		return line
	}
	return line[:maxPreviewLineLength] + "…"
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
