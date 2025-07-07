# logze — Structural logging with zerolog efficiency and slog interface

[![Go Version][version-img]][doc] [![GoDoc][doc-img]][doc] [![Build][ci-img]][ci] [![Coverage][coverage-img]][coverage] [![GoReport][report-img]][report]

A high-performance structured logging library for Go that combines the efficiency of [zerolog](https://github.com/rs/zerolog) with the simplicity of [slog](https://pkg.go.dev/golang.org/x/exp/slog). Write clean, readable logging code that performs exceptionally well.


## ✨ Why Choose logze?

**Simple, Clean Interface:**
```go
// Zerolog example:
log.Error().Err(err).Str("address", "127.0.0.1").Int("retry", n).Msg("cannot start server")

// Logze example:
logze.Err(err, "cannot start server", "address", "127.0.0.1", "retry", n)
```

## 📦 Installation

```bash
go get -u github.com/maxbolgarin/logze/v2
```

**High Performance:**
- 🚀 **3x faster** than standard `slog`
- ⚡ Only **15% slower** than raw `zerolog`
- 🎯 Zero allocations for most operations
- 📈 High-throughput with optional non-blocking I/O


## 📖 Table of Contents

- [Quick Start](#quick-start)
- [Installation](#installation)
- [Core Concepts](#core-concepts)
- [Usage Examples](#usage-examples)
  - [Basic Logging](#basic-logging)
  - [Structured Fields](#structured-fields)
  - [Error Logging](#error-logging)
  - [Formatted Logging](#formatted-logging)
  - [Conditional Logging](#conditional-logging)
  - [Sampling](#sampling)
- [Configuration](#configuration)
  - [Output Configuration](#output-configuration)
  - [Level Configuration](#level-configuration)
  - [Advanced Features](#advanced-features)
- [Global Logger](#global-logger)
- [Pros and Cons](#pros-and-cons)
- [Benchmarks](#benchmarks)
- [Migration Guide](#migration-guide)
- [Contributing](#contributing)


## ⚡ Quick Start

```go
package main

import (
	"errors"
	"github.com/maxbolgarin/logze/v2"
)

func main() {
	// 1. Create a logger with JSON output
	logger := logze.NewConsoleJSON()
	
	// 2. Log with structured fields
	logger.Info("Application started", "version", "1.0.0", "port", 8080)
	
	// 3. Log errors efficiently
	err := errors.New("database connection failed")
	logger.Err(err, "Failed to connect", "host", "localhost", "retry", 3)
	
	// 4. Use global logger for convenience
	logze.Info("Processing request", "user_id", 12345, "action", "create_post")
}
```

## 📝 Usage Examples

### Basic Logging

```go
logger := logze.New(logze.C().WithConsole().WithTrace())

// Different log levels
logger.Trace("Detailed execution trace", "function", "processData")
logger.Debug("Debug information", "variable", someVar)
logger.Info("General information", "statuses", []string{"healthy", "ready"})
logger.Warn("Warning message", "disk_usage", 85.5)
logger.Error("Error occurred", "component", "database")
```

### Structured Fields

```go
// Prepare a new logger with additional fields
logger := logger.With("user_id", 123, "action", "login")

// Log with the prepared logger
logger.Info("User action")

// Complex data types
logger.Info("Request processed",
	"duration", time.Since(start),
	"headers", map[string]string{"Content-Type": "application/json"},
	"response_size", 1024,
	"success", true,
)
```

### Error Logging

```go
// Basic error logging
err := errors.New("connection timeout")
logger.Err(err, "Database connection failed", "host", "db.example.com")

// Error with error object
// Returns: {"level":"error","message":"Database connection failed","error":"connection timeout","host":"db.example.com"}
logger.Error("Database connection failed", err, "host", "db.example.com")

// Error with stack trace and no error object
logger.WithStack().Error("Validation failed")

// Logging with stack trace
logger := logze.New(logze.C().WithConsoleJSON().WithStackTrace())
logger.Err(err, "Critical failure", "operation", "save_user")

```

### Formatted Logging

```go
// Printf-style formatting with structured fields
// Returns: {"level":"info","message":"Processing 100 items in 10s","batch_id":"1234567890"}
logger.Infof("Processing %d items in %s", count, duration, "batch_id", batchID)

// All format verbs are supported
logger.Debugf("User %s (ID: %d) performed action: %v", username, userID, action, 
	"timestamp", time.Now(), "ip", clientIP)
```

### Conditional Logging

```go
// Log only when condition is true
logger.InfoIf(debugMode, "Debug mode enabled", "level", "verbose")
logger.ErrorIf(err != nil, "Operation failed", "error", err)

// Useful for performance-sensitive code
logger.DebugIf(isVerbose, "Detailed state", "state", expensiveStateCalculation())
```

### Sampling

```go
// Sample 10% of debug logs
logger := logze.New(logze.C().WithConsoleJSON().WithPercentageSampler(0.1, logze.LevelDebug))

// Allow only 100 debug logs per second
logger := logze.New(logze.C().WithConsoleJSON().WithMaxSampler(100, time.Second, logze.LevelDebug))

// Pass 100 logs per second and then allow 10% of all logs
logger := logze.New(logze.C().WithConsoleJSON().WithBurstSampler(0.1, 100, time.Second))
```


## ⚙️ Configuration

### Output Configuration

```go
// Console output (development)
logger := logze.New(logze.C().WithConsole()) // Colored output
logger := logze.New(logze.C().WithConsoleNoColor()) // Plain text
logger := logze.New(logze.C().WithConsoleJSON()) // JSON to stderr

// File output
config, closer, err := logze.C().WithFile("app.log", 0644)
if err != nil {
	log.Fatal(err)
}
defer closer.Close()
logger := logze.New(config)

// Multiple outputs
var fileWriter io.Writer // your file writer
logger := logze.New(logze.C(fileWriter).WithConsole())

// Custom writer
logger := logze.New(logze.C(customWriter))
```

### Level Configuration

```go
// Set minimum log level
logger := logze.New(logze.C().WithLevel(logze.LevelWarn)) // Only warn, error, fatal

// Level-specific configuration
logger := logze.New(logze.C().WithDebug()) // Debug and above
logger := logze.New(logze.C().WithInfo())  // Info and above (default)
logger := logze.New(logze.C().WithError()) // Error and above only

// Disable logging completely
logger := logze.New(logze.C().WithDisabled())
```

### Advanced Features

```go
config := logze.NewConfig().
    WithLevel(logze.LevelInfo).                     // Set log level
    WithAddCaller().                               // Include caller info
    WithStackTrace().                               // Enable stack traces
    WithSimpleErrorCounter().                      // Count errors
    WithToIgnore("health", "ping").               // Filter messages
    WithTimeFieldFormat(time.RFC3339).           // Custom time format
    WithNoDiode()                                // Disable buffering

logger := logze.New(config, "service", "api", "version", "2.1.0")
```

## 🌍 Global Logger

For convenience, use the global logger throughout your application:

```go
// Initialize once at application start
logze.Init(logze.C().WithConsoleJSON().WithLevel("info"), "app", "myservice")

// Use anywhere in your codebase
logze.Info("Server starting", "port", 8080)
logze.Err(err, "Failed to process request", "request_id", reqID)

// Create contextual loggers
requestLogger := logze.With("request_id", reqID, "user_id", userID)
requestLogger.Info("Processing request")

// Update global configuration
logze.Update(logze.C().WithLevel("debug")) // Enable debug logging
```


## ✅ Pros and Cons

### Pros

- **🚀 High Performance**: 3x faster than `slog`, leveraging `zerolog`'s efficient engine
- **📝 Clean Interface**: Simple, readable logging calls with structured fields
- **🔧 Flexible Configuration**: Extensive configuration options for any use case
- **🎯 Zero Allocations**: Most operations don't allocate memory
- **🔄 Easy Migration**: Compatible interface with `slog` patterns

### Cons

- **📈 Slight Overhead**: ~15% slower than raw `zerolog` due to field abstraction
- **🧠 Learning Curve**: Advanced features (diode, sampling) may be complex for beginners
- **💾 Message Loss Risk**: Default diode buffering can drop messages under extreme load
- **🎨 Console Performance**: Text/console output is significantly slower than JSON


## 📊 Benchmarks

Performance comparison on Apple M1 Pro:

### Basic Info Logging (message + 2 fields)
```
BenchmarkZerologInfo-8     14,365,629    151.3 ns/op    0 B/op    0 allocs/op
BenchmarkLogzeInfo-8       13,851,320    171.5 ns/op    0 B/op    0 allocs/op  (+13%)
BenchmarkSLogInfo-8         4,491,897    533.2 ns/op    0 B/op    0 allocs/op  (+252%)
```

### Formatted Logging (printf-style + 2 fields)
```
BenchmarkZerologInfoFormat-8  11,758,495  207.0 ns/op   24 B/op   1 allocs/op
BenchmarkLogzeInfoFormat-8    10,047,972  239.8 ns/op   24 B/op   1 allocs/op  (+16%)
BenchmarkSLogInfoFormat-8      4,026,775  601.0 ns/op   24 B/op   1 allocs/op  (+190%)
```

### Error Logging (error + 2 fields)
```
BenchmarkZerologError-8    13,900,646    173.6 ns/op    0 B/op    0 allocs/op
BenchmarkLogzeError-8      10,827,030    221.7 ns/op    0 B/op    0 allocs/op  (+28%)
BenchmarkSLogError-8        3,757,587    635.7 ns/op    0 B/op    0 allocs/op  (+266%)
```

### Console Output (development)
```
BenchmarkZerologInfoConsole-8   682,558   3,306 ns/op   1,922 B/op   51 allocs/op
BenchmarkLogzeInfoConsole-8     704,247   3,366 ns/op   1,922 B/op   51 allocs/op  (+2%)
BenchmarkSLogInfoConsole-8    4,058,850     588 ns/op       0 B/op    0 allocs/op  (-82%)
```

## 🔄 Migration Guide

### From `slog`
```go
// Before (slog)
slog.Info("User created", "user_id", 123, "email", "user@example.com")
slog.Error("Database error", "error", err, "query", query)

// After (logze) - it's identical!
logze.Info("User created", "user_id", 123, "email", "user@example.com")
logze.Error("Database error", "error", err, "query", query) // Enhanced error handling
```

### From `zerolog`
```go
// Before (zerolog)
log.Info().Str("user_id", "123").Str("email", "user@example.com").Msg("User created")
log.Error().Err(err).Str("query", query).Msg("Database error")

// After (logze) - much cleaner!
logze.Info("User created", "user_id", "123", "email", "user@example.com")
logze.Err(err, "Database error", "query", query)
```

### From Standard Library
```go
// Before (standard log)
log.Printf("User %s created with ID %d", email, userID)

// After (logze) - structured and faster!
logze.Infof("User %s created with ID %d", email, userID)
// Or better yet:
logze.Info("User created", "email", email, "user_id", userID)
```

## 🤝 Contributing

We welcome contributions! Please feel free to:

- 🐛 Report bugs by opening issues
- 💡 Suggest features or improvements
- 🔧 Submit pull requests
- 📖 Improve documentation
- ⚡ Add benchmarks or tests

## 📄 License

This project is licensed under the MIT License. See the [LICENSE](LICENSE) file for details.

---

[version-img]: https://img.shields.io/badge/Go-%3E%3D%201.19-%23007d9c
[doc-img]: https://pkg.go.dev/badge/github.com/maxbolgarin/logze
[doc]: https://pkg.go.dev/github.com/maxbolgarin/logze
[ci-img]: https://github.com/maxbolgarin/logze/actions/workflows/go.yml/badge.svg
[ci]: https://github.com/maxbolgarin/logze/actions
[report-img]: https://goreportcard.com/badge/github.com/maxbolgarin/logze
[report]: https://goreportcard.com/report/github.com/maxbolgarin/logze
[coverage-img]: https://codecov.io/gh/maxbolgarin/logze/branch/main/graph/badge.svg
[coverage]: https://codecov.io/gh/maxbolgarin/logze
