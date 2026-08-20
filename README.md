# Concurrent API Issue Crawler

A concurrency-safe API documentation crawler written in Go. It scans web pages recursively up to a configurable depth and uses regular expressions to find keywords indicating API issues, bugs, or deprecations.

## Project Structure

```text
apicrawler/
├── go.mod          # Go module file
├── main.go         # Project entry point and configuration
└── crawler/
    └── crawler.go  # Core concurrent crawling logic
```

## Features

- **Goroutine Concurrency**: Processes multiple URLs simultaneously.
- **Rate Throttling**: Restricts maximum active connections using channel signals.
- **Race Condition Prevention**: Employs `sync.Mutex` for map checking and `sync.WaitGroup` for thread synchronization.
- **Relative URL Resolution**: Automatically transforms relative paths into fully qualified target domains.

## Requirements

- [Go](https://go.dev/doc/install) 1.21 or higher

## Getting Started

1. **Configure Settings**: Open `main.go` and update the constants:
   - `StartURL`: Your target API documentation page.
   - `MaxDepth`: Maximum link depth to hop.
   - `Concurrency`: Active parallel network routines.

2. **Run the Project**:
   ```bash
   go run main.go
   ```
