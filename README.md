# logze — structured logging with zerolog efficiency

[![Go Version][version-img]][doc] [![GoDoc][doc-img]][doc] [![Build][ci-img]][ci] [![Coverage][coverage-img]][coverage] [![GoReport][report-img]][report]

A high-performance structured logging library for Go that combines the efficiency of [zerolog](https://github.com/rs/zerolog) with the simplicity of [slog](https://pkg.go.dev/golang.org/x/exp/slog). Write clean, readable logging code that performs exceptionally well.


## ✨ Why Choose logze?

**Simple, Clean Interface:**

The following examples show the same logging operation in both zerolog and logze:

```go
// Zerolog example:
log.Error().Err(err).Str("address", "127.0.0.1").Int("retry", n).Msg("cannot start server")

// Logze example:
logze.Err(err, "cannot start server", "address", "127.0.0.1", "retry", n)
```

**Why this matters:** Logze eliminates the need to remember multiple method names (`.Str()`, `.Int()`, `.Msg()`) and chain method calls. Instead, you use one simple method call with alternating key-value pairs, making the code more readable and less error-prone.

## 📦 Installation

```bash
go get -u github.com/maxbolgarin/logze/v2
```

**High Performance:**
- 🚀 **3x faster** than standard `slog`
- ⚡ Only **15% slower** than raw `zerolog`
- 🎯 Zero allocations for most operations (thanks to `zerolog`)
- 📈 High-throughput with optional non-blocking I/O (thanks to `diode`)


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

This example demonstrates the core features of logze in just a few lines:

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

**What this does:**
1. **Creates a JSON logger** that outputs structured logs to stderr (ideal for production)
2. **Logs structured information** with key-value pairs that are easy to parse and search
3. **Handles error logging** efficiently with both the error object and contextual information
4. **Demonstrates global logging** for cases where you don't want to pass logger instances around

**Expected output:**
```json
{"level":"info","message":"Application started","version":"1.0.0","port":8080,"time":"2024-01-15T10:30:00Z"}
{"level":"error","message":"Failed to connect","error":"database connection failed","host":"localhost","retry":3,"time":"2024-01-15T10:30:01Z"}
{"level":"info","message":"Processing request","user_id":12345,"action":"create_post","time":"2024-01-15T10:30:02Z"}
```

## 📝 Usage Examples

### Basic Logging

This example shows how to log at different levels with various data types:

```go
logger := logze.New(logze.C().WithConsole().WithTrace())

// Different log levels
logger.Trace("Detailed execution trace", "function", "processData")
logger.Debug("Debug information", "variable", someVar)
logger.Info("General information", "statuses", []string{"healthy", "ready"})
logger.Warn("Warning message", "disk_usage", 85.5)
logger.Error("Error occurred", "component", "database")
```

**Explanation:**
- **Trace**: Most verbose level, typically used for detailed debugging (includes caller information)
- **Debug**: Detailed information for diagnosing problems
- **Info**: General operational messages about what the application is doing
- **Warn**: Potentially harmful situations that should be investigated
- **Error**: Error events that still allow the application to continue


### Structured Fields

This example demonstrates different ways to add structured data to your logs:

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

**What each approach does:**

1. **Prepared logger with `.With()`**: Creates a new logger instance that automatically includes specified fields in every log message. This is perfect for request-scoped logging where you want to include request ID, user ID, etc. in all related logs.

2. **Complex data types**: Shows how logze automatically handles different Go types (supports all types that `zerolog` supports):
   - `time.Duration` → formatted as readable duration
   - `map[string]string` → serialized as JSON object
   - `int` → numeric value
   - `bool` → boolean value

**Expected output:**
```json
{"level":"info","message":"User action","user_id":123,"action":"login","time":"2024-01-15T10:30:00Z"}
{"level":"info","message":"Request processed","duration":"150ms","headers":{"Content-Type":"application/json"},"response_size":1024,"success":true,"time":"2024-01-15T10:30:01Z"}
```

### Error Logging

This example shows different patterns for logging errors:

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

**Different error logging patterns:**

1. **`logger.Err(err, message, fields...)`**: Best practice for error logging. Automatically includes the error in the log entry and increments error counters if configured.

2. **`logger.Error(message, error, fields...)`**: Alternative syntax where you can mix the error with other fields. The error is automatically detected and properly formatted.

3. **`logger.WithStack().Error(message)`**: Captures and includes stack trace information, useful for debugging but has performance impact.

4. **Global stack trace configuration**: When you configure the logger with `WithStackTrace()`, all error logs from this logger instance automatically include stack traces.


### Formatted Logging

This example demonstrates printf-style formatting combined with structured fields:

```go
// Printf-style formatting with structured fields
// Returns: {"level":"info","message":"Processing 100 items in 10s","batch_id":"1234567890"}
logger.Infof("Processing %d items in %s", count, duration, "batch_id", batchID)

// All format verbs are supported
logger.Debugf("User %s (ID: %d) performed action: %v", username, userID, action, 
	"timestamp", time.Now(), "ip", clientIP)
```

**How formatted logging works:**

1. **Format string processing**: The first part of the arguments is used for printf-style formatting
2. **Structured fields**: interface{} remaining arguments after the format placeholders become structured key-value pairs
3. **Best of both worlds**: You get readable formatted messages plus searchable structured data

**Example breakdown:**
- `"Processing %d items in %s"` with `count=100, duration="10s"` becomes `"Processing 100 items in 10s"`
- `"batch_id", batchID` becomes a structured field in the JSON output
- This gives you both human-readable messages and machine-parseable data

### Conditional Logging

This example shows how to avoid expensive operations when logging is disabled:

```go
// Log only when condition is true
logger.InfoIf(debugMode, "Debug mode enabled", "level", "verbose")
logger.ErrorIf(err != nil, "Operation failed", "error", err)
```

### Sampling

This example demonstrates how to reduce log volume in high-throughput applications:

```go
// Sample 10% of debug logs
logger := logze.New(logze.C().WithConsoleJSON().WithPercentageSampler(0.1, logze.LevelDebug))

// Allow only 100 debug logs per second
logger := logze.New(logze.C().WithConsoleJSON().WithMaxSampler(100, time.Second, logze.LevelDebug))

// Pass 100 logs per second and then allow 10% of all logs
logger := logze.New(logze.C().WithConsoleJSON().WithBurstSampler(0.1, 100, time.Second))
```

**Sampling strategies explained:**

1. **Percentage Sampling**: Randomly samples X% of logs at specified levels
   - Use when you want a representative sample of all logs
   - Good for getting a statistical overview without overwhelming log storage

2. **Max Sampling**: Allows up to X logs per time period, then drops the rest
   - Use to prevent log flooding from busy code paths
   - Guarantees you won't exceed storage/bandwidth limits

3. **Burst Sampling**: Allows X logs immediately, then switches to percentage sampling
   - Use for bursty applications that need immediate logs but should throttle under sustained load
   - Combines the benefits of both approaches


## ⚙️ Configuration

### Output Configuration

These examples show different ways to configure where your logs are written:

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

**Output format explanations:**

1. **`WithConsole()`**: Human-readable colored output, perfect for development
   - Example: `2:04PM INF User logged in user_id=123 action=login`
   - **Pros**: Easy to read during development
   - **Cons**: Slow performance, not machine-parseable

2. **`WithConsoleJSON()`**: JSON output to stderr, ideal for production
   - Example: `{"level":"info","message":"User logged in","user_id":123,"time":"2024-01-15T14:04:00Z"}`
   - **Pros**: Fast, machine-parseable, works with log aggregators
   - **Cons**: Not human-readable

3. **File output**: Writes logs to files with automatic rotation support
   - **Use case**: When you need persistent logs or can't use stderr
   - **Remember**: Always defer `closer.Close()` to ensure logs are flushed

4. **Multiple outputs**: Send logs to multiple destinations simultaneously
   - **Use case**: Log to both files and console, or send to multiple log aggregators

### Level Configuration

These examples show how to control which log levels are output:

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

This example shows advanced configuration options:

```go
logger := logze.NewConfig().
	WithLevel(logze.LevelInfo).                     // Set log level
	WithAddCaller().                               // Include caller info
	WithStackTrace().                               // Enable stack traces
	WithSimpleErrorCounter().                      // Count errors
	WithToIgnore("health", "ping").               // Filter messages
	WithTimeFieldFormat(time.RFC3339).           // Custom time format
	WithNoDiode().                                // Disable buffering
   New("service", "api", "version", "2.1.0")    // Create logger based on config

logger.Info("health check") // Won't be printed
```

**Advanced features explained:**

1. **`WithAddCaller()`**: Adds file name and line number to logs
   - **Use case**: Debugging when you need to know exactly where logs come from
   - **Performance impact**: Slight overhead due to runtime reflection

2. **`WithStackTrace()`**: Automatically includes stack traces in error logs
   - **Use case**: Debugging complex error conditions
   - **Performance impact**: Significant overhead, use sparingly

3. **`WithSimpleErrorCounter()`**: Tracks how minterface{} errors have been logged
   - **Use case**: Monitoring and alerting based on error rates
   - **Access**: Use `logger.GetErrorCounter()` to get current count

4. **`WithToIgnore()`**: Filters out log messages containing specified strings
   - **Use case**: Reduce noise from health checks, monitoring pings, etc.
   - **Example**: Logs containing "health" or "ping" won't be output

5. **`WithNoDiode()`**: Disables async buffering for synchronous logging
   - **Use case**: When you need guaranteed log delivery (e.g., before program exit)
   - **Trade-off**: Better reliability but worse performance

6. **`New()`**: Allows to create a logger from config, fields `"service", "api", "version", "2.1.0"` are added to every log

## 🌍 Global Logger

These examples show how to use the global logger for convenience:

```go
// Initialize once at application start
logze.Init(logze.C().WithConsoleJSON().WithLevel("info"), "app", "myservice")

// Use interface{}where in your codebase
logze.Info("Server starting", "port", 8080)
logze.Err(err, "Failed to process request", "request_id", reqID)

// Create contextual loggers
requestLogger := logze.With("request_id", reqID, "user_id", userID)
requestLogger.Info("Processing request")

// Update global configuration
logze.Update(logze.C().WithLevel("debug")) // Enable debug logging
```

**Global logger patterns:**

1. **`logze.Init()`**: Sets up the global logger with configuration and default fields
   - **Best practice**: Call once during application startup
   - **Use case**: When you don't want to pass logger instances everywhere

2. **Direct global calls**: `logze.Info()`, `logze.Err()`, etc.
   - **Advantage**: Simple, no need to pass logger instances
   - **Disadvantage**: Less flexible than instance loggers

3. **`logze.With()`**: Creates a new logger with additional fields
   - **Use case**: Request-scoped logging where you want consistent fields
   - **Pattern**: Create once per request, use throughout request handling

4. **`logze.Update()`**: Changes global logger configuration at runtime
   - **Use case**: Enabling debug logging for troubleshooting without restart
   - **Warning**: Affects all subsequent logging calls globally
   
  
## 🔄 Migration Guide

### From `slog`

The migration from slog is nearly seamless:

```go
// Before (slog)
slog.Info("User created", "user_id", 123, "email", "user@example.com")
slog.Error("Database error", "error", err, "query", query)

// After (logze) - it's identical!
logze.Info("User created", "user_id", 123, "email", "user@example.com")
logze.Error("Database error", "error", err, "query", query) // Enhanced error handling
```

**Migration benefits:**
- **No API changes**: The basic logging calls are identical
- **Enhanced error handling**: Better error object processing and counting
- **Performance improvement**: 3x faster than slog with same interface
- **Additional features**: Caller info, stack traces, sampling, etc.

### From `zerolog`

Logze simplifies `zerolog` usage:

```go
// Before (zerolog)
log.Info().Str("user_id", "123").Str("email", "user@example.com").Msg("User created")
log.Error().Err(err).Str("query", query).Msg("Database error")

// After (logze) - much cleaner!
logze.Info("User created", "user_id", "123", "email", "user@example.com")
logze.Err(err, "Database error", "query", query)
```

**Migration advantages:**
- **Simpler syntax**: No method chaining or type-specific methods
- **Same performance**: Only 15% overhead compared to raw zerolog
- **Fewer mistakes**: No forgotten `.Msg()` calls or wrong type methods
- **Better readability**: Clear intent with single method calls

### From Standard Library

Transform printf-style logging into structured logging:

```go
// Before (standard log)
log.Printf("User %s created with ID %d", email, userID)

// After (logze) - structured and faster!
logze.Infof("User %s created with ID %d", email, userID)
// Or better yet:
logze.Info("User created", "email", email, "user_id", userID)
```

**Why structured logging is better:**
- **Searchable**: Query logs by specific field values
- **Aggregatable**: Calculate metrics from log data
- **Consistent**: Standardized format across different log sources
- **Future-proof**: Easy to parse and process programmatically

**Migration strategy:**
1. **Phase 1**: Replace `log.Printf` with `logze.Infof` (minimal changes)
2. **Phase 2**: Convert to structured logging with key-value pairs
3. **Phase 3**: Add appropriate log levels and error handling

## ✅ Pros and Cons

### Pros

- **🚀 High Performance**: 3x faster than `slog`, leveraging `zerolog`'s efficient engine
- **📝 Clean Interface**: Simple, readable logging calls with structured fields
- **🔧 Flexible Configuration**: Extensive configuration options for interface{} use case
- **🎯 Zero Allocations**: Most operations don't allocate memory
- **🔄 Easy Migration**: Compatible interface with `slog` patterns

### Cons

- **📈 Slight Overhead**: ~15% slower than raw `zerolog` due to field abstraction
- **💾 Message Loss Risk**: Default diode buffering can drop messages under extreme load
- **🎨 Console Performance**: Text/console output is significantly slower than JSON


## 📊 Benchmarks
Thoughts from benchmarks:
* `logze` is about 3 times faster than `slog` and for 15% slower than `zerolog`
* Format methods like `logze.Infof` or `Msgf` don't add big overhead (only 30% slower and +1 alloc)
* Stack trace is slow in all loggers, but `zerolog` is the fastest, `logze` slight slower and makes more allocations (price for easy of use).
* Console writer is slow in `logze` and `zerolog` and should be used only in development (that the case when `slog` wins - if you want to use text writer in production)


Here is result of `go test -bench=. -benchmem -benchtime=2s`:

```
goos: darwin
goarch: arm64
cpu: Apple M1 Pro
```

Logging `Info` and two fields:
```text
BenchmarkZerologInfo-8                   7689862               146.3 ns/op         0 B/op           0 allocs/op
BenchmarkLogzeInfo-8                     6985168               184.1 ns/op         0 B/op           0 allocs/op
BenchmarkSLogInfo-8                      2293354               523.7 ns/op         0 B/op           0 allocs/op
```

Logging `Infof` (formatted) and two fields:
```text
BenchmarkZerologInfoFormat-8            11758495               207.0 ns/op            24 B/op          1 allocs/op
BenchmarkLogzeInfoFormat-8              10047972               239.8 ns/op            24 B/op          1 allocs/op
BenchmarkSLogInfoFormat-8                4026775               601.0 ns/op            24 B/op          1 allocs/op
```


Logging `Error` with error and two fields:
```text
BenchmarkZerologError-8                  7172144               166.6 ns/op         0 B/op           0 allocs/op
BenchmarkLogzeError-8                    6193178               189.4 ns/op         0 B/op           0 allocs/op
BenchmarkSLogError-8                     1930131               651.0 ns/op         0 B/op           0 allocs/op
```


Logging `Error` with error, stack trace and two fields:
```text
BenchmarkZerologErrorWithStack-8          560763              2057 ns/op         824 B/op           7 allocs/op
BenchmarkLogzeErrorWithStack-8            485652              2406 ns/op        1307 B/op          10 allocs/op
BenchmarkSLogErrorWithStack-8             311107              3786 ns/op        1617 B/op           3 allocs/op
```


Logging `Info` and two fields using text handler for console / development:
```text
BenchmarkZerologInfoConsole-8             682558               3306 ns/op            1922 B/op         51 allocs/op
BenchmarkLogzeInfoConsole-8               704247               3366 ns/op            1922 B/op         51 allocs/op
BenchmarkSLogInfoConsole-8               4058850               588.1 ns/op             0 B/op          0 allocs/op
```


Additional `logze` features
```text
BenchmarkLogzeErr-8                      6057480               206.1 ns/op         0 B/op            0 allocs/op
BenchmarkLogzeToIgnore5-8                4703277               248.3 ns/op         64 B/op           1 allocs/op
BenchmarkLogzeErrStack-8                  438679               2644 ns/op          768 B/op          18 allocs/op
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

[version-img]: https://img.shields.io/badge/Go-%3E%3D%201.15-%23007d9c
[doc-img]: https://pkg.go.dev/badge/github.com/maxbolgarin/logze
[doc]: https://pkg.go.dev/github.com/maxbolgarin/logze
[ci-img]: https://github.com/maxbolgarin/logze/actions/workflows/go.yml/badge.svg
[ci]: https://github.com/maxbolgarin/logze/actions
[report-img]: https://goreportcard.com/badge/github.com/maxbolgarin/logze
[report]: https://goreportcard.com/report/github.com/maxbolgarin/logze
[coverage-img]: https://codecov.io/gh/maxbolgarin/logze/branch/main/graph/badge.svg
[coverage]: https://codecov.io/gh/maxbolgarin/logze
