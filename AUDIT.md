# Logze v2 Code Audit Report

**Date:** 2025-11-06
**Auditor:** Claude (AI Assistant)
**Version:** v2 branch
**Commit:** 5c797b9

---

## Executive Summary

This comprehensive audit of the logze v2 codebase identified **2 critical bugs**, **4 medium-priority issues**, and **12 UX improvement opportunities**. Additionally, a major enhancement opportunity for **OpenTelemetry trace integration** was identified that would significantly improve the library's observability capabilities.

All critical bugs have been **fixed** as part of this audit.

### Key Findings

- ✅ **FIXED:** Missing `StackTrace()` method implementation causing interface compliance issues
- ✅ **FIXED:** Diode writer goroutine leak in `Update()` methods
- ✅ **DOCUMENTED:** Global `TimeFieldFormat` side effect
- 🎯 **Major Opportunity:** OpenTelemetry trace integration for modern distributed systems
- 💡 **12 UX improvements** identified for enhanced developer experience

---

## Table of Contents

1. [Critical Bugs (FIXED)](#critical-bugs-fixed)
2. [Medium Priority Issues](#medium-priority-issues)
3. [Low Priority Issues](#low-priority-issues)
4. [UX Improvements](#ux-improvements)
5. [Trace Integration Proposal](#trace-integration-proposal)
6. [Implementation Priorities](#implementation-priorities)
7. [Testing Recommendations](#testing-recommendations)

---

## Critical Bugs (FIXED)

### 1. Missing `StackTrace()` Method Implementation ✅ FIXED

**Severity:** HIGH
**File:** `stack.go:86-127`
**Status:** ✅ Fixed

#### Problem

The `genericStackTraceError` interface defines a `StackTrace() []uintptr` method, but the `stackError` struct didn't implement it. This caused errors wrapped with `WithStack()` to not properly implement the interface.

```go
// Interface definition (line 86-88)
type genericStackTraceError interface {
    StackTrace() []uintptr
}

// Implementation was missing!
type stackError struct {
    err   error
    stack stackTrace
}
```

#### Impact

- Errors wrapped with `WithStack()` didn't satisfy the `genericStackTraceError` interface
- Type assertions would fail
- Integration with error handling libraries expecting this interface would break

#### Fix Applied

```go
// StackTrace implements the genericStackTraceError interface
func (s *stackError) StackTrace() []uintptr {
    return s.stack
}
```

**Location:** `stack.go:104-107`

---

### 2. Diode Writer Goroutine Leak in `Update()` ✅ FIXED

**Severity:** MEDIUM-HIGH
**Files:** `logze.go:333-347`, `global.go:118-120`
**Status:** ✅ Fixed

#### Problem

When `Update()` replaces a logger configuration, the old `diodeWriter` was never closed. Each diode writer spawns a background goroutine (poller or waiter) that continues running indefinitely.

```go
// Before fix
func (l *Logger) Update(cfg Config, fields ...interface{}) {
    newLogger := New(cfg, fields...)
    l.l = newLogger.l
    // ... other assignments
    l.diodeWriter = newLogger.diodeWriter  // Old writer leaked!
}
```

#### Impact

- **Goroutine leak:** Each `Update()` call leaves an orphaned goroutine running
- **Resource leak:** Memory consumed by abandoned diode buffers
- **Severity increases** with frequent updates (e.g., dynamic log level changes)
- **Production risk:** Long-running applications calling `Update()` repeatedly would accumulate goroutines

#### Fix Applied

```go
func (l *Logger) Update(cfg Config, fields ...interface{}) {
    // Close the old diode writer to prevent goroutine leak
    if l.diodeWriter != nil {
        l.diodeWriter.Close()
    }

    newLogger := New(cfg, fields...)
    // ... rest of implementation
}
```

**Locations:**
- `logze.go:334-337`
- `global.go:118-120` (improved documentation)

---

### 3. Global `TimeFieldFormat` Side Effect ✅ DOCUMENTED

**Severity:** MEDIUM
**File:** `logze.go:95`
**Status:** ✅ Documented (Limitation from zerolog)

#### Problem

Setting `cfg.TimeFieldFormat` modifies the **global** `zerolog.TimeFieldFormat` variable, affecting ALL zerolog loggers in the application, not just the logze instance being created.

```go
func New(cfg Config, fields ...interface{}) Logger {
    // ...
    zerolog.TimeFieldFormat = cfg.TimeFieldFormat  // Global side effect!
    // ...
}
```

#### Impact

- **Unexpected behavior:** Creating a logze logger changes time format for unrelated zerolog loggers
- **Race conditions:** Multiple goroutines creating loggers with different time formats
- **Non-local effects:** Changes in one module affect loggers in other modules
- **Violation of principle of least surprise**

#### Fix Applied

Added comprehensive documentation warning users about this limitation:

```go
// ⚠️  IMPORTANT: Global Time Format Side Effect
//
// Setting cfg.TimeFieldFormat modifies the global zerolog.TimeFieldFormat variable,
// which affects ALL zerolog loggers in the application, not just this logze instance.
// This is a limitation of the underlying zerolog library. If you have multiple logger
// configurations with different time formats, the last one created will take effect
// for all loggers.
```

**Location:** `logze.go:86-92`

#### Recommendations

1. Document this behavior in README.md
2. Consider using a single consistent time format per application
3. If zerolog adds per-logger time format support, migrate to it
4. As a workaround, applications could set `zerolog.TimeFieldFormat` once at startup

---

## Medium Priority Issues

### 4. Race Condition in Global Logger Operations

**Severity:** MEDIUM
**Files:** `global.go:76-78, 118-120`
**Status:** Documented, not fixed

#### Problem

The global logger operations `SetDefault()` and `Update()` are not thread-safe. Concurrent calls can cause data races.

```go
var log = NewConsoleJSON()  // Global variable

func SetDefault(l Logger) {
    log = l  // Not atomic, not protected
}
```

#### Impact

- Race detector warnings when updating global logger concurrently
- Potential for seeing partial logger state during updates
- Unsafe for dynamic configuration changes in multi-threaded applications

#### Current Mitigation

Enhanced documentation warning users:

```go
// ⚠️ THREAD SAFETY WARNING: This function is NOT safe for concurrent use.
// Ensure no other goroutines are using the global logger while calling Update.
// Consider using a mutex or other synchronization mechanism if you need to
// update the global logger from multiple goroutines.
```

#### Recommended Fix

Add mutex protection:

```go
var (
    log   Logger
    logMu sync.RWMutex
)

func SetDefault(l Logger) {
    logMu.Lock()
    defer logMu.Unlock()
    log = l
}

func Default() Logger {
    logMu.RLock()
    defer logMu.RUnlock()
    return log
}
```

**Trade-off:** Adds small performance overhead to all global logger operations.

---

### 5. Inconsistent Error Field Handling in `setErrorWithStack`

**Severity:** MEDIUM
**File:** `logze.go:1301-1335`
**Status:** Needs review

#### Problem

The logic for removing errors from the fields array in `setErrorWithStack` has edge cases that may not be handled correctly:

```go
func (l Logger) setErrorWithStack(ev *zerolog.Event, inFormat bool, args ...interface{}) (*zerolog.Event, []interface{}) {
    newFields := args
    for i, a := range args {
        err, ok := a.(error)
        if !ok {
            continue
        }
        // ... stack trace handling

        if !inFormat {
            // Complex logic for removing error from fields
            if i > 0 && i%2 == 1 {
                // Error is at odd position (value position), remove key-value pair
                newFields = make([]interface{}, 0, len(args)-2)
                newFields = append(newFields, args[:i-1]...)
                newFields = append(newFields, args[i+1:]...)
            } else if i == 0 {
                // Error is at the beginning
                newFields = args[1:]
            } else {
                // Error is at even position (key position), only remove the error
                // ⚠️ This would leave a key without a value!
                newFields = make([]interface{}, 0, len(args)-1)
                newFields = append(newFields, args[:i]...)
                newFields = append(newFields, args[i+1:]...)
            }
        }
        return ev.Err(err), newFields
    }
    return ev, newFields
}
```

#### Issues

1. **Comment mismatch:** The comment says "only remove the error" but this removes the error leaving the previous key orphaned
2. **Edge case:** If error is at an even position > 0, it's being treated as a key, but the removal would break the key-value pairing
3. **Single pass:** Only processes the first error found, ignoring subsequent errors in args

#### Impact

- Malformed log output when errors are positioned unexpectedly in fields
- Orphaned keys without values in structured output
- Inconsistent behavior based on error position

#### Recommended Fix

Revise the logic to properly handle all cases:

```go
if !inFormat {
    // Only remove error if it's a value (odd position) in key-value pairs
    if i%2 == 1 && i > 0 {
        // Error is a value, remove the key-value pair
        newFields = make([]interface{}, 0, len(args)-2)
        newFields = append(newFields, args[:i-1]...)
        newFields = append(newFields, args[i+1:]...)
    } else if i == 0 {
        // Error is at the beginning, treat as standalone
        newFields = args[1:]
    }
    // If error is at even position > 0, it's likely a mistake
    // Keep it as-is to maintain key-value pairing
}
```

---

### 6. Shared `diodeWriter` in Derived Loggers

**Severity:** LOW-MEDIUM
**File:** `logze.go:377-387`
**Status:** Needs documentation

#### Problem

When creating derived loggers with `WithFields()` or `With()`, the new logger shares the same `diodeWriter` pointer as the parent.

```go
func (l Logger) WithFields(fields ...interface{}) Logger {
    return Logger{
        l:           l.l.With().Fields(fields).Logger(),
        errCounter:  l.errCounter,
        toIgnore:    l.toIgnore,
        ignoreMap:   l.ignoreMap,
        stackTrace:  l.stackTrace,
        inited:      l.inited,
        diodeWriter: l.diodeWriter,  // Shared pointer!
    }
}
```

#### Impact

- Closing one logger affects all derived loggers
- Unexpected behavior when managing logger lifecycle
- Could cause panics if one logger is closed while another is still writing

#### Recommended Fix

Document this sharing behavior clearly:

```go
// WithFields creates a new logger with additional fields that will be included in every log message.
//
// IMPORTANT: The returned logger shares the same diode writer as the parent logger.
// Closing either the parent or derived logger will affect both. To avoid this, create
// a new independent logger with New() instead of deriving from an existing one.
```

**Alternative:** Consider whether derived loggers should get independent diode writers, though this adds complexity and resource usage.

---

## Low Priority Issues

### 7. Error Counter Only Counts First Error

**Severity:** LOW
**File:** `logze.go:1301-1335`

When multiple errors are present in fields, only the first one increments the error counter.

```go
// Only increments counter for first error found
for i, a := range args {
    err, ok := a.(error)
    if !ok {
        continue
    }
    l.incErrorCounter(err)
    return ev.Err(err), newFields  // Returns immediately
}
```

**Impact:** Error counts may be underreported in edge cases.

**Fix:** Either document this behavior or process all errors.

---

### 8. No Validation for Fields Length

**Severity:** LOW
**Files:** Multiple

Fields should be key-value pairs (even number of arguments), but there's no validation:

```go
logger.Info("message", "key1", "value1", "key2")  // Missing value for key2
```

**Impact:** Silent errors, malformed JSON output.

**Recommendation:** Add validation in debug mode or document requirement clearly.

---

## UX Improvements

### 9. HTTP Logging Convenience Methods

**Priority:** HIGH
**Value:** Simplifies common use case

Add specialized methods for HTTP request/response logging:

```go
// HTTP request logging
func (l Logger) HTTP(method, path string, status int, duration time.Duration, fields ...interface{}) {
    l.Info("HTTP request",
        "method", method,
        "path", path,
        "status", status,
        "duration_ms", duration.Milliseconds(),
        fields...)
}

// HTTP error logging
func (l Logger) HTTPError(method, path string, status int, err error, fields ...interface{}) {
    l.Err(err, "HTTP request failed",
        "method", method,
        "path", path,
        "status", status,
        fields...)
}

// HTTP with automatic status level detection
func (l Logger) HTTPAuto(method, path string, status int, duration time.Duration, err error, fields ...interface{}) {
    if err != nil || status >= 500 {
        l.HTTPError(method, path, status, err, fields...)
    } else if status >= 400 {
        l.Warn("HTTP client error", "method", method, "path", path, "status", status, fields...)
    } else {
        l.HTTP(method, path, status, duration, fields...)
    }
}
```

**Example usage:**

```go
logger.HTTP("GET", "/api/users", 200, time.Since(start), "user_id", userID)
logger.HTTPError("POST", "/api/orders", 500, err, "order_id", orderID)
logger.HTTPAuto("GET", "/api/products", status, time.Since(start), err)
```

---

### 10. Duration/Latency Helpers

**Priority:** MEDIUM
**Value:** Cleaner code for performance logging

```go
// Auto-add duration field
func (l Logger) WithDuration(start time.Time) Logger {
    return l.With("duration_ms", time.Since(start).Milliseconds())
}

// Auto-timed execution
func (l Logger) Timed(msg string, fn func()) {
    start := time.Now()
    fn()
    l.Info(msg, "duration_ms", time.Since(start).Milliseconds())
}

// Async timed execution with result
func (l Logger) TimedFunc(msg string, fn func() error) error {
    start := time.Now()
    err := fn()
    if err != nil {
        l.Err(err, msg, "duration_ms", time.Since(start).Milliseconds())
    } else {
        l.Info(msg, "duration_ms", time.Since(start).Milliseconds())
    }
    return err
}
```

**Example usage:**

```go
// Simple duration tracking
defer logger.WithDuration(time.Now()).Info("Processing completed")

// Auto-timed execution
logger.Timed("Database query", func() {
    db.Query(...)
})

// With error handling
err := logger.TimedFunc("API call", func() error {
    return client.Call()
})
```

---

### 11. Panic Recovery Helper

**Priority:** MEDIUM
**Value:** Simplifies panic handling patterns

```go
// For use in defer to recover and log panics
func (l Logger) RecoverPanic(fields ...interface{}) {
    if r := recover(); r != nil {
        stack := debug.Stack()
        l.Error("Panic recovered",
            append(fields,
                "panic", r,
                "stack", string(stack))...)
    }
}

// With callback for custom handling
func (l Logger) RecoverPanicWithCallback(callback func(interface{}), fields ...interface{}) {
    if r := recover(); r != nil {
        stack := debug.Stack()
        l.Error("Panic recovered",
            append(fields,
                "panic", r,
                "stack", string(stack))...)
        if callback != nil {
            callback(r)
        }
    }
}
```

**Example usage:**

```go
func handleRequest(logger logze.Logger) {
    defer logger.RecoverPanic("request_id", reqID)

    // ... code that might panic
}

func criticalOperation(logger logze.Logger) {
    defer logger.RecoverPanicWithCallback(func(p interface{}) {
        metrics.IncrementPanicCounter()
        alerts.SendPanicAlert(p)
    }, "operation", "critical")

    // ... code that might panic
}
```

---

### 12. Regex Support for Message Filtering

**Priority:** MEDIUM
**Value:** More flexible log filtering

Add regex-based message filtering:

```go
// In Config
type Config struct {
    // ... existing fields
    ToIgnoreRegex []*regexp.Regexp
}

func (c Config) WithToIgnoreRegex(patterns ...string) Config {
    c.ToIgnoreRegex = make([]*regexp.Regexp, len(patterns))
    for i, pattern := range patterns {
        c.ToIgnoreRegex[i] = regexp.MustCompile(pattern)
    }
    return c
}

// In Logger.log()
func (l Logger) log(ev *zerolog.Event, msg string, fields []interface{}) {
    // ... existing ignore check

    // Check regex patterns
    for _, re := range l.toIgnoreRegex {
        if re.MatchString(msg) {
            return
        }
    }

    // ... rest of implementation
}
```

**Example usage:**

```go
logger := logze.New(
    logze.C().
        WithConsoleJSON().
        WithToIgnore("health", "ping").           // Exact/substring match
        WithToIgnoreRegex(`^GET /metrics.*`),     // Regex match
)
```

---

### 13. Configurable Console Time Format

**Priority:** LOW
**Value:** Better readability in different contexts

Currently hardcoded in `config.go:601`:

```go
func getConsoleWriter(w io.Writer, color bool) zerolog.ConsoleWriter {
    return zerolog.ConsoleWriter{
        Out:        w,
        NoColor:    !color,
        TimeFormat: "2006-01-02 15:04:05",  // Hardcoded!
    }
}
```

**Improvement:**

```go
type Config struct {
    // ... existing fields
    ConsoleTimeFormat string
}

func (c Config) WithConsoleTimeFormat(format string) Config {
    c.ConsoleTimeFormat = format
    return c
}

func getConsoleWriter(w io.Writer, color bool, timeFormat string) zerolog.ConsoleWriter {
    if timeFormat == "" {
        timeFormat = "2006-01-02 15:04:05"
    }
    return zerolog.ConsoleWriter{
        Out:        w,
        NoColor:    !color,
        TimeFormat: timeFormat,
    }
}
```

---

### 14. Structured Error Counter

**Priority:** MEDIUM
**Value:** Better error categorization

Extend error counter to track error types:

```go
type DetailedErrorCounter interface {
    Inc(err error)
    IncType(errType string)
    GetCounts() map[string]uint64
    GetTotal() uint64
}

type SimpleDetailedErrorCounter struct {
    mu     sync.RWMutex
    counts map[string]uint64
    total  uint64
}

func (c *SimpleDetailedErrorCounter) Inc(err error) {
    c.mu.Lock()
    defer c.mu.Unlock()

    errType := fmt.Sprintf("%T", err)
    c.counts[errType]++
    c.total++
}

func (c *SimpleDetailedErrorCounter) GetCounts() map[string]uint64 {
    c.mu.RLock()
    defer c.mu.RUnlock()

    result := make(map[string]uint64, len(c.counts))
    for k, v := range c.counts {
        result[k] = v
    }
    return result
}
```

---

### 15. Log Rotation Support

**Priority:** LOW
**Value:** Complete file logging solution

```go
import "gopkg.in/natefinch/lumberjack.v2"

func (c Config) WithRotatingFile(filename string, maxSizeMB int, maxAge int, maxBackups int) Config {
    logger := &lumberjack.Logger{
        Filename:   filename,
        MaxSize:    maxSizeMB,
        MaxAge:     maxAge,
        MaxBackups: maxBackups,
        Compress:   true,
    }
    c.Writers = append(c.Writers, logger)
    return c
}
```

**Example usage:**

```go
logger := logze.New(
    logze.C().
        WithRotatingFile("app.log", 100, 30, 5).  // 100MB, 30 days, 5 backups
        WithLevel("info"),
)
```

---

### 16. Logger Inspection Methods

**Priority:** LOW
**Value:** Runtime debugging and configuration validation

```go
func (l Logger) GetLevel() string {
    level := l.l.GetLevel()
    return level.String()
}

func (l Logger) IsEnabled(level string) bool {
    lvl, err := zerolog.ParseLevel(level)
    if err != nil {
        return false
    }
    return l.l.GetLevel() <= lvl
}

func (l Logger) HasErrorCounter() bool {
    return l.errCounter != nil
}

func (l Logger) HasDiode() bool {
    return l.diodeWriter != nil
}
```

**Example usage:**

```go
if logger.IsEnabled("debug") {
    // Perform expensive debug logging
    logger.Debug("Expensive operation", "data", heavyComputation())
}

if !logger.HasErrorCounter() {
    logger = logger.WithSimpleErrorCounter()
}
```

---

### 17. Context-Aware Logging

**Priority:** HIGH
**Value:** Better integration with context-based workflows

```go
// Automatically extract trace info and check context cancellation
func (l Logger) InfoCtx(ctx context.Context, msg string, fields ...interface{}) {
    // Check if context is cancelled
    select {
    case <-ctx.Done():
        return  // Don't log if context is cancelled
    default:
    }

    // Add trace info if available (see Trace Integration section)
    if traceFields := extractTraceContext(ctx); len(traceFields) > 0 {
        fields = append(fields, traceFields...)
    }

    l.Info(msg, fields...)
}

// Similar for other levels
func (l Logger) ErrCtx(ctx context.Context, err error, msg string, fields ...interface{})
func (l Logger) DebugCtx(ctx context.Context, msg string, fields ...interface{})
// etc.
```

---

### 18. Batch Logging

**Priority:** LOW
**Value:** Extreme high-throughput scenarios

```go
type BatchLogger struct {
    parent Logger
    buffer []batchEntry
    mu     sync.Mutex
    size   int
}

type batchEntry struct {
    level  string
    msg    string
    fields []interface{}
}

func (l Logger) Batch(size int) *BatchLogger {
    return &BatchLogger{
        parent: l,
        buffer: make([]batchEntry, 0, size),
        size:   size,
    }
}

func (b *BatchLogger) Info(msg string, fields ...interface{}) {
    b.add("info", msg, fields)
}

func (b *BatchLogger) add(level, msg string, fields []interface{}) {
    b.mu.Lock()
    defer b.mu.Unlock()

    b.buffer = append(b.buffer, batchEntry{level, msg, fields})
    if len(b.buffer) >= b.size {
        b.flushLocked()
    }
}

func (b *BatchLogger) Flush() {
    b.mu.Lock()
    defer b.mu.Unlock()
    b.flushLocked()
}
```

---

### 19. Diode Statistics

**Priority:** LOW
**Value:** Operational visibility

```go
type DiodeStats struct {
    DroppedMessages uint64
    BufferSize      int
    CurrentUsage    int
}

func (l Logger) GetDiodeStats() *DiodeStats {
    if l.diodeWriter == nil {
        return nil
    }

    // Would need to extend diode.Writer to expose these stats
    return &DiodeStats{
        // ... stats from diode writer
    }
}
```

---

### 20. Better Error Method Naming

**Priority:** LOW
**Value:** Consistency and clarity

Current state has some confusion:
- `Err(err, msg, fields)` - logs error with message
- `Error(msg, fields)` - logs error-level message
- `Erro(err, msg, fields)` - integrates with erro package

**Recommendation:** Document the distinction more clearly in README and add examples showing when to use each.

---

## Trace Integration Proposal

### Overview

Modern distributed systems require tight integration between **logging** and **tracing** for complete observability. This proposal outlines a comprehensive OpenTelemetry integration for logze.

### Benefits

1. **Automatic log-trace correlation** - Every log automatically includes trace ID and span ID
2. **Standard format** - Uses W3C Trace Context specification
3. **Universal compatibility** - Works with Jaeger, Zipkin, Datadog, Honeycomb, etc.
4. **Bidirectional integration** - Logs appear as span events, traces referenced in logs
5. **Sampling coordination** - Log sampling can follow trace sampling decisions
6. **Better debugging** - Jump from logs to traces and vice versa

### Architecture

```
┌─────────────────────────────────────────────────────────┐
│  Application Code                                       │
│  logger.InfoCtx(ctx, "User logged in", "user_id", 123) │
└──────────────────────┬──────────────────────────────────┘
                       │
┌──────────────────────┴──────────────────────────────────┐
│  Logze with OTel Integration                            │
│  1. Extract trace context from ctx                      │
│  2. Add trace_id, span_id, trace_flags to fields       │
│  3. Optionally add log as span event                    │
│  4. Log with standard logze flow                        │
└──────────────────────┬──────────────────────────────────┘
                       │
           ┌───────────┴───────────┐
           │                       │
┌──────────▼─────────┐  ┌─────────▼──────────┐
│  Log Output        │  │  Trace Backend     │
│  (JSON/Console)    │  │  (Jaeger, etc.)    │
│  {                 │  │  Span {            │
│    "trace_id":     │  │    events: [       │
│    "abc123",       │  │      "User login"  │
│    "message":      │  │    ]               │
│    "User login"    │  │  }                 │
│  }                 │  │                    │
└────────────────────┘  └────────────────────┘
```

### Implementation Plan

#### Phase 1: Core Trace Context Extraction

**New file:** `trace.go`

```go
package logze

import (
    "context"
    "go.opentelemetry.io/otel/trace"
)

// TraceExtractor defines how to extract trace information from context
type TraceExtractor interface {
    Extract(ctx context.Context) []interface{}
}

// OtelTraceExtractor extracts OpenTelemetry trace context
type OtelTraceExtractor struct{}

func (o OtelTraceExtractor) Extract(ctx context.Context) []interface{} {
    span := trace.SpanFromContext(ctx)
    if !span.SpanContext().IsValid() {
        return nil
    }

    sc := span.SpanContext()
    fields := []interface{}{
        "trace_id", sc.TraceID().String(),
        "span_id", sc.SpanID().String(),
    }

    if sc.IsSampled() {
        fields = append(fields, "trace_flags", "sampled")
    }

    return fields
}

// extractTraceContext is a helper that uses the configured extractor
func extractTraceContext(ctx context.Context, extractor TraceExtractor) []interface{} {
    if ctx == nil || extractor == nil {
        return nil
    }
    return extractor.Extract(ctx)
}
```

#### Phase 2: Configuration

**Update:** `config.go`

```go
type Config struct {
    // ... existing fields

    // TraceExtractor extracts trace context from context.Context
    TraceExtractor TraceExtractor

    // AddSpanEvents if true, adds logs as span events
    AddSpanEvents bool

    // TraceSampledOnly if true, only logs when trace is sampled
    TraceSampledOnly bool

    // AlwaysLogSampled if true, ignores sampling for sampled traces
    AlwaysLogSampled bool
}

func (c Config) WithOpenTelemetry() Config {
    c.TraceExtractor = OtelTraceExtractor{}
    return c
}

func (c Config) WithTraceExtractor(extractor TraceExtractor) Config {
    c.TraceExtractor = extractor
    return c
}

func (c Config) WithSpanEvents() Config {
    c.AddSpanEvents = true
    return c
}

func (c Config) WithTraceSampledLogging() Config {
    c.TraceSampledOnly = true
    return c
}

func (c Config) WithAlwaysLogSampledTraces() Config {
    c.AlwaysLogSampled = true
    return c
}
```

#### Phase 3: Context-Aware Logging Methods

**Update:** `logze.go`

```go
type Logger struct {
    // ... existing fields
    traceExtractor TraceExtractor
    addSpanEvents  bool
}

// InfoCtx logs with automatic trace context extraction
func (l Logger) InfoCtx(ctx context.Context, msg string, fields ...interface{}) {
    // Check context cancellation
    select {
    case <-ctx.Done():
        return
    default:
    }

    // Extract and add trace fields
    if l.traceExtractor != nil {
        if traceFields := l.traceExtractor.Extract(ctx); len(traceFields) > 0 {
            fields = append(fields, traceFields...)
        }
    }

    // Add as span event if enabled
    if l.addSpanEvents {
        if span := trace.SpanFromContext(ctx); span.IsRecording() {
            span.AddEvent(msg)
        }
    }

    l.Info(msg, fields...)
}

// ErrCtx logs error with trace context and marks span as error
func (l Logger) ErrCtx(ctx context.Context, err error, msg string, fields ...interface{}) {
    // Check context cancellation
    select {
    case <-ctx.Done():
        return
    default:
    }

    // Extract and add trace fields
    if l.traceExtractor != nil {
        if traceFields := l.traceExtractor.Extract(ctx); len(traceFields) > 0 {
            fields = append(fields, traceFields...)
        }
    }

    // Mark span as error and add event
    if span := trace.SpanFromContext(ctx); span.IsRecording() {
        span.RecordError(err)
        span.AddEvent(msg)
    }

    l.Err(err, msg, fields...)
}

// Similar for DebugCtx, WarnCtx, etc.
func (l Logger) DebugCtx(ctx context.Context, msg string, fields ...interface{})
func (l Logger) WarnCtx(ctx context.Context, msg string, fields ...interface{})
func (l Logger) TraceCtx(ctx context.Context, msg string, fields ...interface{})
```

#### Phase 4: Sampling Coordination

```go
// In Logger.log() method
func (l Logger) log(ev *zerolog.Event, msg string, fields []interface{}) {
    // Check trace-based sampling if enabled
    if l.traceSampledOnly && l.traceExtractor != nil {
        // Extract from thread-local context (would need to be stored)
        // Only log if trace is sampled
        // Implementation depends on how to access current context
    }

    // ... rest of existing implementation
}
```

### Example Usage

```go
package main

import (
    "context"
    "github.com/maxbolgarin/logze/v2"
    "go.opentelemetry.io/otel"
    "go.opentelemetry.io/otel/trace"
)

func main() {
    // Setup logger with OTel integration
    logger := logze.New(
        logze.C().
            WithConsoleJSON().
            WithLevel("info").
            WithOpenTelemetry().        // Enable OTel integration
            WithSpanEvents().            // Add logs as span events
            WithTraceSampledLogging(),   // Only log sampled traces
        "service", "api",
    )

    // Create a tracer
    tracer := otel.Tracer("example")

    // Start a span
    ctx, span := tracer.Start(context.Background(), "handleRequest")
    defer span.End()

    // Log with automatic trace correlation
    logger.InfoCtx(ctx, "Processing request", "user_id", 123)
    // Output: {"level":"info","trace_id":"abc123...","span_id":"def456...","message":"Processing request","user_id":123}

    // Error logging automatically marks span
    if err := processUser(ctx, 123); err != nil {
        logger.ErrCtx(ctx, err, "Failed to process user")
        // - Logs with trace context
        // - Marks span as error
        // - Adds error event to span
    }
}

func processUser(ctx context.Context, userID int) error {
    logger := logze.GetFromContext(ctx)

    // All logs automatically include trace context
    logger.InfoCtx(ctx, "Loading user", "user_id", userID)

    // ... business logic

    return nil
}
```

### Log Output with Traces

**Without trace integration:**
```json
{"level":"info","message":"Processing request","user_id":123,"time":"2024-01-15T10:30:00Z"}
```

**With trace integration:**
```json
{
  "level": "info",
  "trace_id": "4bf92f3577b34da6a3ce929d0e0e4736",
  "span_id": "00f067aa0ba902b7",
  "trace_flags": "sampled",
  "message": "Processing request",
  "user_id": 123,
  "time": "2024-01-15T10:30:00Z"
}
```

**Corresponding trace in Jaeger:**
```
Span: handleRequest (4bf92f3577b34da6a3ce929d0e0e4736)
  └─ Events:
     └─ Processing request (00f067aa0ba902b7)
```

### Dependencies

```go
// go.mod additions
require (
    go.opentelemetry.io/otel v1.21.0
    go.opentelemetry.io/otel/trace v1.21.0
)
```

**Note:** These should be optional dependencies. The trace features should only import OTel packages when explicitly enabled.

### Backward Compatibility

1. **No breaking changes** - All trace features are opt-in
2. **Zero overhead** - No performance impact if trace integration is not enabled
3. **Graceful degradation** - Works fine if trace context is missing
4. **Optional dependency** - OTel is only needed if trace features are used

### Testing Strategy

**New file:** `trace_test.go`

```go
func TestOtelTraceExtractor(t *testing.T) {
    // Test with valid trace context
    // Test with invalid trace context
    // Test with nil context
    // Test field format
}

func TestInfoCtx(t *testing.T) {
    // Test trace field injection
    // Test context cancellation
    // Test span event creation
    // Test without trace extractor
}

func TestErrCtx(t *testing.T) {
    // Test span error marking
    // Test trace field injection
    // Test error counter increment
}

func TestTraceSampling(t *testing.T) {
    // Test sampled trace logging
    // Test unsampled trace filtering
    // Test AlwaysLogSampled behavior
}
```

### Documentation

Update README.md with new section:

```markdown
## 🔗 Distributed Tracing Integration

Logze provides seamless integration with OpenTelemetry for distributed tracing:

### Quick Start

```go
logger := logze.New(
    logze.C().
        WithConsoleJSON().
        WithOpenTelemetry().  // Enable trace integration
)

// Use context-aware logging
logger.InfoCtx(ctx, "User logged in", "user_id", 123)
```

### Features

- ✅ Automatic trace ID and span ID injection
- ✅ Logs appear as span events in traces
- ✅ Errors automatically mark spans
- ✅ Sampling coordination
- ✅ W3C Trace Context standard

### Advanced Configuration

```go
logger := logze.New(
    logze.C().
        WithConsoleJSON().
        WithOpenTelemetry().
        WithSpanEvents().            // Add logs as span events
        WithTraceSampledLogging().   // Only log sampled traces
        WithAlwaysLogSampledTraces() // Override sampling for sampled traces
)
```
```

---

## Implementation Priorities

### Immediate (Week 1)

1. ✅ **Fix critical bugs** - COMPLETED
   - Missing StackTrace() method
   - Diode writer leaks
   - Documentation improvements

### Short-term (Weeks 2-4)

2. 🎯 **OpenTelemetry Integration** - HIGH VALUE
   - Phase 1: Core trace extraction
   - Phase 2: Configuration
   - Phase 3: Context-aware methods
   - Phase 4: Sampling coordination

3. 🔧 **HTTP Logging Methods** - HIGH VALUE
   - Add HTTP(), HTTPError(), HTTPAuto()
   - Documentation and examples

4. 🔧 **Context-Aware Logging** - HIGH VALUE
   - InfoCtx, ErrCtx, etc. (may be part of OTel integration)

### Medium-term (Months 2-3)

5. 💡 **Duration Helpers** - MEDIUM VALUE
   - WithDuration(), Timed(), TimedFunc()

6. 💡 **Panic Recovery** - MEDIUM VALUE
   - RecoverPanic(), RecoverPanicWithCallback()

7. 💡 **Regex Filtering** - MEDIUM VALUE
   - WithToIgnoreRegex()

8. 💡 **Structured Error Counter** - MEDIUM VALUE
   - DetailedErrorCounter interface

### Long-term (Months 3-6)

9. 🌟 **Logger Inspection** - LOW VALUE
   - GetLevel(), IsEnabled(), HasErrorCounter()

10. 🌟 **Configurable Console Format** - LOW VALUE
    - WithConsoleTimeFormat()

11. 🌟 **Log Rotation** - LOW VALUE
    - WithRotatingFile()

12. 🌟 **Batch Logging** - LOW VALUE
    - Batch(), Flush()

### Ongoing

- 📝 Documentation improvements
- 🧪 Test coverage expansion
- 🐛 Bug fixes as discovered
- 💬 Community feedback integration

---

## Testing Recommendations

### Unit Tests

1. **Stack trace tests**
   ```go
   func TestStackErrorImplementsInterface(t *testing.T)
   func TestStackTraceCapture(t *testing.T)
   func TestStackTraceFormat(t *testing.T)
   ```

2. **Diode leak tests**
   ```go
   func TestUpdateClosesOldDiode(t *testing.T)
   func TestNoLeakOnUpdate(t *testing.T)  // Use runtime.NumGoroutine()
   ```

3. **Trace integration tests**
   ```go
   func TestTraceContextExtraction(t *testing.T)
   func TestSpanEventCreation(t *testing.T)
   func TestErrorMarksSpan(t *testing.T)
   func TestSamplingCoordination(t *testing.T)
   ```

### Integration Tests

1. **End-to-end trace integration**
   - Create span, log with context, verify trace ID in output
   - Verify span events appear in trace
   - Verify error marking

2. **Concurrent logging**
   - Multiple goroutines logging simultaneously
   - Verify no race conditions
   - Verify all logs have correct trace context

3. **Performance tests**
   - Benchmark trace extraction overhead
   - Benchmark context-aware logging
   - Compare with and without trace integration

### Regression Tests

1. **Diode leak regression**
   - Repeatedly call Update(), verify no goroutine leak

2. **Stack trace regression**
   - Verify errors wrapped with WithStack() work correctly
   - Verify stack traces appear in logs

### Load Tests

1. **High-throughput logging**
   - 10k+ logs/second with trace integration
   - Verify no dropped logs
   - Verify performance acceptable

2. **Diode stress test**
   - Log faster than diode can flush
   - Verify alert function is called
   - Verify graceful degradation

---

## Conclusion

This audit has identified and **fixed critical bugs** that could cause goroutine leaks and interface compliance issues. The codebase is now more robust and better documented.

The **OpenTelemetry trace integration** represents the most significant enhancement opportunity, enabling logze to provide complete observability for modern distributed systems. Combined with the proposed UX improvements, logze can become a best-in-class logging solution for Go applications.

### Summary Statistics

- ✅ **2 critical bugs fixed**
- ✅ **3 issues documented**
- 💡 **12 UX improvements proposed**
- 🎯 **1 major enhancement opportunity (Trace integration)**
- 📊 **~2000 lines of code reviewed**
- 🧪 **15+ test scenarios recommended**

### Next Steps

1. Review and approve fixes
2. Prioritize UX improvements based on user feedback
3. Begin OpenTelemetry integration implementation
4. Expand test coverage
5. Update documentation with new features
6. Consider community feedback and feature requests

---

**Audit completed:** 2025-11-06
**Status:** Critical bugs fixed, ready for next phase of enhancements
