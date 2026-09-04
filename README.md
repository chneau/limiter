# limiter

A lightweight, high-performance, channel-based goroutine concurrency limiter in Go.

[![Go Reference](https://pkg.go.dev/badge/github.com/chneau/limiter.svg)](https://pkg.go.dev/github.com/chneau/limiter)
[![Go Report Card](https://goreportcard.com/badge/github.com/chneau/limiter)](https://goreportcard.com/report/github.com/chneau/limiter)
[![License: MIT](https://img.shields.io/badge/License-MIT-yellow.svg)](https://opensource.org/licenses/MIT)

---

## ⚡ Features

- **Lightweight & Idiomatic**: Built directly on Go's channel synchronization primitive (`type Limiter chan struct{}`).
- **Panic Safety**: Automatic slot recovery using `defer` ensures tickets are recycled even if a job panics.
- **Context Cancellation**: Cancel or timeout waiting tasks with `ExecuteContext(ctx, job)`.
- **Non-blocking Execution**: Conditionally attempt execution without blocking using `TryExecute(job)`.
- **Introspection**: Inspect active state with `Cap()`, `Available()`, and `Running()`.
- **100% Backwards Compatible**: Preserves existing `New(limit)`, `Execute(job)`, and `Wait()` APIs.

---

## 🚀 Quick Start

### Installation

```bash
go get github.com/chneau/limiter
```

### Usage Example

```go
package main

import (
	"context"
	"fmt"
	"time"

	"github.com/chneau/limiter"
)

func main() {
	// Limit concurrency to 4 simultaneous goroutines
	limit := limiter.New(4)

	for i := range 10 {
		limit.Execute(func() {
			fmt.Printf("Job %d executing (running: %d)\n", i, limit.Running())
			time.Sleep(50 * time.Millisecond)
		})
	}

	// Non-blocking try
	if !limit.TryExecute(func() { fmt.Println("Opportunistic execution") }) {
		fmt.Println("Slots currently full, skipped")
	}

	// Context with timeout
	ctx, cancel := context.WithTimeout(context.Background(), 100*time.Millisecond)
	defer cancel()

	_ = limit.ExecuteContext(ctx, func() {
		fmt.Println("Running within context deadline")
	})

	// Wait for all scheduled jobs to finish
	limit.Wait()
}
```

---

## 🛠️ Development Commands

```bash
# Run tests with race detector and coverage
go test -v -race -cover ./...

# Run benchmarks
go test -benchmem -bench=. ./...

# Modernize codebase
modernize ./...

# Run linter
golangci-lint run ./...
```

---

## 📄 License

[MIT](LICENSE)
