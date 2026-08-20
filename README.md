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

1. **Set the crawl target**: Pass the API documentation root via the `--root` flag (required — the program exits with an error if it's omitted).
2. **Configure other settings** *(optional)*: Open `main.go` and update the constants:
   - `MaxDepth`: Maximum link depth to hop.
   - `Concurrency`: Active parallel network routines.

3. **Run the Project**:
   ```bash
   go run main.go --root https://example.com/api/docs
   ```

## Build and Run

### Run directly (no binary)

```bash
go run main.go --root https://example.com/api/docs
```

### Build a binary

```bash
go build -o apicrawler .
```

This produces an `apicrawler` executable in the project root. Run it with:

```bash
./apicrawler --root https://example.com/api/docs
```

### Cross-compiling

Go's toolchain supports building for other platforms by setting `GOOS`/`GOARCH`, e.g. for Linux from any host:

```bash
GOOS=linux GOARCH=amd64 go build -o apicrawler-linux-amd64 .
```

### Verifying the build

```bash
go vet ./...
go build ./...
```

### CLI flags

| Flag     | Required | Description                                    |
|----------|----------|-------------------------------------------------|
| `--root` | Yes      | Target API documentation root URL to crawl.     |

`MaxDepth` and `Concurrency` remain hardcoded constants in `main.go` (see [Getting Started](#getting-started)) — rebuild (or `go run`) after changing them, as there's no flag or environment variable for those yet.
