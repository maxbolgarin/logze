package logze

import (
	"context"
	stdlog "log"
	"runtime/debug"
	"sync"
	"time"

	"github.com/rs/zerolog"
)

var (
	log   = NewConsoleJSON()
	logMu sync.RWMutex
)

// Default returns a copy of the global logger instance.
//
// The global logger is initialized with JSON output to stderr at info level.
// This function returns a copy, so modifications (like adding fields) won't
// affect the global instance.
//
// Use this when you want to add request-specific or component-specific fields
// to a logger without affecting other parts of your application.
//
// This function is thread-safe and uses a read lock for concurrent access.
//
// Example usage:
//
//	logger := logze.Default().With("component", "auth", "request_id", reqID)
//	logger.Info("User authenticated") // Includes component and request_id
func Default() Logger {
	logMu.RLock()
	defer logMu.RUnlock()
	return log
}

// D is a convenient shorthand for [Default].
//
// Example usage:
//
//	logger := logze.D().With("user_id", userID)
func D() Logger {
	return log
}

// DefaultPtr returns a pointer to the global logger instance.
//
// Unlike Default(), this returns a pointer to the actual global logger.
// Use this when you need to share the exact same logger instance or when
// working with APIs that expect a pointer.
//
// ⚠️ Warning: Modifying the returned logger affects the global instance.
// ⚠️ Thread Safety: This function acquires a read lock. Do not call Update()
// or other mutating operations on the global logger while holding the pointer,
// as this may cause deadlocks.
//
// Example usage:
//
//	ptr := logze.DefaultPtr()
//	// Be careful: changes affect global instance
func DefaultPtr() *Logger {
	logMu.RLock()
	defer logMu.RUnlock()
	return &log
}

// DP is a convenient shorthand for [DefaultPtr].
//
// Example usage:
//
//	ptr := logze.DP()
func DP() *Logger {
	return &log
}

// SetDefault replaces the global logger with the provided logger instance.
//
// This affects all subsequent calls to package-level logging functions
// (Info, Error, etc.) and Default()/D() functions. Use this to configure
// global logging behavior for your entire application.
//
// This function is thread-safe and uses a write lock for exclusive access.
//
// Example usage:
//
//	logger := logze.New(logze.C(fileWriter).WithLevel("warn").WithSimpleErrorCounter())
//	logze.SetDefault(logger)
//
//	// Now all package-level logging uses the new configuration
//	logze.Info("This uses the new global logger")
func SetDefault(l Logger) {
	logMu.Lock()
	defer logMu.Unlock()
	log = l
}

// Init initializes the global logger with the specified configuration and fields.
//
// This function creates a new logger using the provided config and replaces
// the global logger instance. It also configures the standard library log
// package to use this logger as its output destination.
//
// This is typically called once during application startup to establish
// global logging configuration.
//
// This function is thread-safe and uses a write lock for exclusive access.
//
// Example usage:
//
//	config := logze.C(logFile).WithLevel("info").WithSimpleErrorCounter()
//	logze.Init(config, "service", "api", "version", "2.1.0")
//
//	// Now all package functions and standard log calls use this configuration
//	logze.Info("Application started")
//	log.Println("This also goes through logze")
func Init(cfg Config, fields ...interface{}) {
	logMu.Lock()
	defer logMu.Unlock()
	log = New(cfg, fields...)
	SetStdLogger(log)
}

// Update reconfigures the global logger with new settings and fields.
//
// This function updates the global logger in-place using the provided config,
// replacing its configuration, output writers, level, and other settings.
// It also reconfigures the standard library log package.
//
// This function is thread-safe and uses a write lock for exclusive access.
// It properly closes the old diode writer to prevent goroutine leaks.
//
// Example usage:
//
//	// Switch from development to production configuration
//	prodConfig := logze.C(prodFile).WithLevel("warn").WithNoDiode()
//	logze.Update(prodConfig, "environment", "production")
func Update(cfg Config, fields ...interface{}) {
	logMu.Lock()
	defer logMu.Unlock()
	log.Update(cfg, fields...)
	SetStdLogger(log)
}

// SetStdLogger configures the standard library log package to use the specified logger.
//
// This function redirects all standard library log output (log.Printf, log.Println, etc.)
// to go through the provided logze logger. It also disables standard log formatting
// flags since logze handles its own formatting.
//
// Optional fields can be provided to add context to all standard library log messages.
//
// Example usage:
//
//	logger := logze.NewConsoleJSON("component", "stdlib")
//	logze.SetStdLogger(logger, "source", "legacy_code")
//
//	// Now standard library log calls will include component and source fields
//	log.Println("This goes through logze with context")
func SetStdLogger(l Logger, fields ...interface{}) {
	stdlog.SetFlags(0)
	stdlog.SetOutput(l.WithFields(fields...))
	log = l
}

// WithFields creates a logger with additional fields based on the global logger.
//
// This function returns a new logger instance that includes the specified fields
// in all log messages. The global logger remains unchanged.
//
// Fields should be provided as alternating key-value pairs.
//
// This function is thread-safe and uses a read lock for concurrent access.
//
// Example usage:
//
//	requestLogger := logze.WithFields("request_id", "abc123", "user_id", 456)
//	requestLogger.Info("Processing request") // Includes request_id and user_id
func WithFields(fields ...interface{}) Logger {
	logMu.RLock()
	defer logMu.RUnlock()
	return log.WithFields(fields...)
}

// With is a convenient shorthand for [WithFields].
//
// Example usage:
//
//	logger := logze.With("component", "auth", "operation", "login")
func With(fields ...interface{}) Logger {
	return log.With(fields...)
}

// WithLevel creates a logger with a specific log level based on the global logger.
//
// This function returns a new logger instance with the specified minimum log level.
// The global logger remains unchanged.
//
// Example usage:
//
//	debugLogger := logze.WithLevel("debug")
//	debugLogger.Debug("This will be logged") // Even if global level is info
func WithLevel(level string) Logger {
	return log.WithLevel(level)
}

// WithErrorCounter returns [Logger] with the provided [ErrorCounter], based on a global logger.
func WithErrorCounter(ec ErrorCounter) Logger {
	return log.WithErrorCounter(ec)
}

// WithSimpleErrorCounter returns [Logger] with a simple [ErrorCounter],
// based on a global logger.
func WithSimpleErrorCounter() Logger {
	return log.WithSimpleErrorCounter()
}

// WithToIgnore returns [Logger] with the provided list of messages to ignore based on a global logger.
func WithToIgnore(toIgnore ...string) Logger {
	return log.WithToIgnore(toIgnore...)
}

// WithCaller returns [Logger] with the provided caller skip frame count based on a global logger.
func WithCaller(callerSkipFrameCount int) Logger {
	return log.WithCaller(callerSkipFrameCount)
}

// WithDefaultCaller returns [Logger] with the default caller skip frame count based on a global logger.
func WithDefaultCaller() Logger {
	return log.WithDefaultCaller()
}

// GetErrorCounter returns Logger's underlying [ErrorCounter] from global logger.
func GetErrorCounter() ErrorCounter {
	return log.GetErrorCounter()
}

// CloseDiode closes the underlying [diode.Writer] if it is used.
func CloseDiode() error {
	return log.CloseDiode()
}

// Close closes the underlying [diode.Writer] if it is used.
func Close() error {
	return log.Close()
}

// WithSampler returns [Logger] with the provided [zerolog.Sampler].
func WithSampler(sampler zerolog.Sampler) Logger {
	return log.WithSampler(sampler)
}

// WithPercentageSampler returns [Logger] with the provided percentage sampler.
func WithPercentageSampler(percentage float64, levels ...string) Logger {
	return log.WithPercentageSampler(percentage, levels...)
}

// WithBurstSampler returns [Logger] with the provided burst sampler.
func WithBurstSampler(percentage float64, burst int, period time.Duration, levels ...string) Logger {
	return log.WithBurstSampler(percentage, burst, period, levels...)
}

// WithMaxSampler returns [Logger] with the provided max sampler.
func WithMaxSampler(max int, period time.Duration, levels ...string) Logger {
	return log.WithMaxSampler(max, period, levels...)
}

// Trace logs a message in trace level adding provided fields and information about method caller
// using a global logger. This function is thread-safe.
func Trace(msg string, fields ...interface{}) {
	logMu.RLock()
	l := log
	logMu.RUnlock()
	l.log(l.l.Trace().Caller(1), msg, fields)
}

// Tracef logs a formatted message in trace level adding provided fields after formatting args
// and information about method caller using a global logger. This function is thread-safe.
func Tracef(msg string, args ...interface{}) {
	logMu.RLock()
	l := log
	logMu.RUnlock()
	l.logf(l.l.Trace().Caller(1), msg, args)
}

// TraceIf logs a message in trace level adding provided fields and information about method caller if condition is true.
// This function is thread-safe.
func TraceIf(condition bool, msg string, fields ...interface{}) {
	if condition {
		Trace(msg, fields...)
	}
}

// Debug logs a message in debug level adding provided fields using a global logger. This function is thread-safe.
func Debug(msg string, fields ...interface{}) {
	logMu.RLock()
	l := log
	logMu.RUnlock()
	l.Debug(msg, fields...)
}

// Debugf logs a formatted message in debug level adding provided fields after formatting args using a global logger.
// This function is thread-safe.
func Debugf(msg string, args ...interface{}) {
	logMu.RLock()
	l := log
	logMu.RUnlock()
	l.Debugf(msg, args...)
}

// DebugIf logs a message in debug level adding provided fields if condition is true. This function is thread-safe.
func DebugIf(condition bool, msg string, fields ...interface{}) {
	if condition {
		Debug(msg, fields...)
	}
}

// Info logs a message in info level adding provided fields using a global logger. This function is thread-safe.
func Info(msg string, fields ...interface{}) {
	logMu.RLock()
	l := log
	logMu.RUnlock()
	l.Info(msg, fields...)
}

// Infof logs a formatted message in info level adding provided fields after formatting args using a global logger.
// This function is thread-safe.
func Infof(msg string, args ...interface{}) {
	logMu.RLock()
	l := log
	logMu.RUnlock()
	l.Infof(msg, args...)
}

// InfoIf logs a message in info level adding provided fields if condition is true. This function is thread-safe.
func InfoIf(condition bool, msg string, fields ...interface{}) {
	if condition {
		Info(msg, fields...)
	}
}

// Warn logs a message in warning level adding provided fields using a global logger. This function is thread-safe.
func Warn(msg string, fields ...interface{}) {
	logMu.RLock()
	l := log
	logMu.RUnlock()
	l.Warn(msg, fields...)
}

// Warnf logs a formatted message in warn level adding provided fields after formatting args using a global logger.
// This function is thread-safe.
func Warnf(msg string, args ...interface{}) {
	logMu.RLock()
	l := log
	logMu.RUnlock()
	l.Warnf(msg, args...)
}

// WarnIf logs a message in warning level adding provided fields if condition is true. This function is thread-safe.
func WarnIf(condition bool, msg string, fields ...interface{}) {
	if condition {
		Warn(msg, fields...)
	}
}

// Err logs a provided error in error level adding provided fields using a global logger. This function is thread-safe.
func Err(err error, msg string, fields ...interface{}) {
	logMu.RLock()
	l := log
	logMu.RUnlock()
	l.Err(err, msg, fields...)
}

// ErrIf logs a provided error in error level adding provided fields if condition is true. This function is thread-safe.
func ErrIf(condition bool, err error, msg string, fields ...interface{}) {
	if condition {
		Err(err, msg, fields...)
	}
}

// Error logs a message in error level adding provided fields using a global logger. This function is thread-safe.
func Error(msg string, fields ...interface{}) {
	logMu.RLock()
	l := log
	logMu.RUnlock()
	l.Error(msg, fields...)
}

// Errorf logs a formatted message in error level adding provided fields after formatting args using a global logger.
// This function is thread-safe.
func Errorf(msg string, args ...interface{}) {
	logMu.RLock()
	l := log
	logMu.RUnlock()
	l.Errorf(msg, args...)
}

// ErrorIf logs a message in error level adding provided fields if condition is true. This function is thread-safe.
func ErrorIf(condition bool, msg string, fields ...interface{}) {
	if condition {
		Error(msg, fields...)
	}
}

// ErrStack logs a stack trace of provided error as message in error level adding fields. This function is thread-safe.
func ErrStack(err error, fields ...interface{}) {
	logMu.RLock()
	l := log
	logMu.RUnlock()
	l.ErrStack(err, fields...)
}

// FatalIf logs a message in fatal level adding provided fields if condition is true, then calls os.Exit(1).
func FatalIf(condition bool, v ...interface{}) {
	log.FatalIf(condition, v...)
}

// Fatal logs a message in fatal level using fmt.Sprint to interpret args sing a global logger, then calls os.Exit(1).
func Fatal(v ...interface{}) {
	log.Fatal(v...)
}

// Fatalf logs a formatted message in fatal level using a global logger, then calls os.Exit(1).
func Fatalf(format string, args ...interface{}) {
	log.Fatalf(format, args...)
}

// Fatalln logs a message in fatal level using fmt.Sprintln to interpret args using a global logger, then calls os.Exit(1).
func Fatalln(v ...interface{}) {
	log.Fatalln(v...)
}

// PanicIf logs a message in fatal level adding provided fields if condition is true, then calls panic().
func PanicIf(condition bool, v ...interface{}) {
	log.PanicIf(condition, v...)
}

// Panic logs a message in fatal level using fmt.Sprint to interpret args using a global logger, then calls panic().
func Panic(v ...interface{}) {
	log.Panic(v...)
}

// Panicf logs a formatted message in fatal level using a global logger, then calls panic().
func Panicf(format string, args ...interface{}) {
	log.Panicf(format, args...)
}

// Panicln logs a message in fatal level using fmt.Sprintln to interpret args using a global logger, then calls panic().
func Panicln(v ...interface{}) {
	log.Panicln(v...)
}

// Print logs a message without level using [fmt.Sprint] to interpret args using a global logger.
func Print(v ...interface{}) {
	log.Print(v...)
}

// PrintIf logs a message without level using [fmt.Sprint] to interpret args if condition is true.
func PrintIf(condition bool, v ...interface{}) {
	log.PrintIf(condition, v...)
}

// PrintStack logs a current stack trace.
func PrintStack(v ...interface{}) {
	log.PrintStack(v...)
}

// Log logs a message without level using [fmt.Sprint] to interpret args using a global logger.
// It is an alias for [Print].
func Log(v ...interface{}) {
	log.Log(v...)
}

// Printf logs a formatted message without level using a global logger.
func Printf(format string, args ...interface{}) {
	log.Printf(format, args...)
}

// Println writes a message without level using fmt.Sprintln to interpret args using a global logger.
func Println(v ...interface{}) {
	log.Println(v...)
}

// Write writes bytes to underlying [io.Writer] using a global logger.
func Write(p []byte) (n int, err error) {
	return log.Write(p)
}

// Raw returns Logger's underlying [zerolog.Logger] from global logger.
func Raw() *zerolog.Logger {
	return log.Raw()
}

// HTTP logs an HTTP request at debug level with standardized fields using a global logger.
// This function is thread-safe.
//
// Parameters:
//   - method: HTTP method (GET, POST, etc.)
//   - path: Request path
//   - status: HTTP status code
//   - duration: Request processing duration
//   - fields: Additional key-value pairs to include
//
// Example usage:
//
//	start := time.Now()
//	// ... handle request
//	logze.HTTP("GET", "/api/users", 200, time.Since(start), "user_id", userID)
func HTTP(method, path string, status int, duration time.Duration, fields ...interface{}) {
	logMu.RLock()
	l := log
	logMu.RUnlock()
	l.HTTP(method, path, status, duration, fields...)
}

// HTTPError logs an HTTP request error at error level with standardized fields using a global logger.
// This function is thread-safe.
//
// Parameters:
//   - method: HTTP method (GET, POST, etc.)
//   - path: Request path
//   - status: HTTP status code
//   - err: The error that occurred
//   - fields: Additional key-value pairs to include
//
// Example usage:
//
//	if err != nil {
//	    logze.HTTPError("POST", "/api/orders", 500, err, "order_id", orderID)
//	}
func HTTPError(method, path string, status int, err error, fields ...interface{}) {
	logMu.RLock()
	l := log
	logMu.RUnlock()
	l.HTTPError(method, path, status, err, fields...)
}

// HTTPAuto automatically selects the appropriate log level based on HTTP status code and error using a global logger.
// This function is thread-safe.
//
// Log levels:
//   - Error level: If err is not nil OR status >= 500 (server errors)
//   - Warn level: If status >= 400 (client errors)
//   - Debug level: Otherwise (success)
//
// Parameters:
//   - method: HTTP method (GET, POST, etc.)
//   - path: Request path
//   - status: HTTP status code
//   - duration: Request processing duration
//   - err: Optional error (can be nil)
//   - fields: Additional key-value pairs to include
//
// Example usage:
//
//	start := time.Now()
//	resp, err := client.Call()
//	logze.HTTPAuto("GET", "/api/products", resp.StatusCode, time.Since(start), err, "product_id", prodID)
func HTTPAuto(method, path string, status int, duration time.Duration, err error, fields ...interface{}) {
	logMu.RLock()
	l := log
	logMu.RUnlock()
	l.HTTPAuto(method, path, status, duration, err, fields...)
}

// Recover recovers from panics and logs them at error level with stack trace using a global logger.
// This function should be called with defer to catch and log panics. This function is thread-safe.
//
// If a panic occurs, it:
//   - Recovers from the panic
//   - Logs the panic value at error level
//   - Includes the full stack trace in the log
//   - Includes any additional fields provided
//
// The panic is caught and logged, but not re-thrown. If you need to re-throw
// the panic after logging, use RecoverPanicWithCallback and call panic() in the callback.
//
// Parameters:
//   - fields: Optional key-value pairs to include in the panic log
//
// Example usage:
//
//	func handleRequest(reqID string) {
//	    defer logze.Recover("request_id", reqID)
//
//	    // ... code that might panic
//	    processData()
//	}
func Recover(fields ...interface{}) {
	logMu.RLock()
	l := log
	logMu.RUnlock()
	if r := recover(); r != nil {
		stack := debug.Stack()
		f := make([]interface{}, 0, len(fields)+2)
		f = append(f, "error", r)
		f = append(f, fields...)
		l.Error(string(stack), f...)
	}
}

// RecoverWithCallback recovers from panics, logs them, and executes a callback function using a global logger.
// This function should be called with defer to catch and log panics with custom handling. This function is thread-safe.
//
// If a panic occurs, it:
//   - Recovers from the panic
//   - Logs the panic value at error level with stack trace
//   - Executes the provided callback function with the panic value
//   - Includes any additional fields provided
//
// Parameters:
//   - callback: Function to call if a panic occurs (receives the panic value)
//   - fields: Optional key-value pairs to include in the panic log
//
// Example usage:
//
//	func handleRequest(metrics *Metrics) {
//	    defer logze.RecoverWithCallback(func(p interface{}) {
//	        metrics.IncrementPanicCounter()
//	        alerts.SendPanicAlert(p)
//	    }, "request_id", reqID)
//
//	    // ... code that might panic
//	}
func RecoverWithCallback(callback func(interface{}), fields ...interface{}) {
	logMu.RLock()
	l := log
	logMu.RUnlock()
	// We need to call the method directly to ensure recover() works in the defer chain
	if r := recover(); r != nil {
		stack := debug.Stack()
		f := make([]interface{}, 0, len(fields)+2)
		f = append(f, "error", r)
		f = append(f, fields...)
		l.Error(string(stack), f...)
		if callback != nil {
			callback(r)
		}
	}
}

// InfoCtx logs a message at info level, checking context cancellation first using a global logger.
// This function is thread-safe.
//
// If the context is cancelled (Done channel is closed), this function returns
// immediately without logging. This prevents unnecessary logging operations
// when the request/operation has been cancelled.
//
// Parameters:
//   - ctx: Context to check for cancellation
//   - msg: Log message
//   - fields: Optional key-value pairs
//
// Example usage:
//
//	func handleRequest(ctx context.Context) {
//	    logze.InfoCtx(ctx, "Processing request", "user_id", 123)
//	    // If request is cancelled, log won't be written
//	}
func InfoCtx(ctx context.Context, msg string, fields ...interface{}) {
	logMu.RLock()
	l := log
	logMu.RUnlock()
	l.InfoCtx(ctx, msg, fields...)
}

// DebugCtx logs a message at debug level, checking context cancellation first using a global logger.
// This function is thread-safe.
// See InfoCtx for details about context checking behavior.
func DebugCtx(ctx context.Context, msg string, fields ...interface{}) {
	logMu.RLock()
	l := log
	logMu.RUnlock()
	l.DebugCtx(ctx, msg, fields...)
}

// TraceCtx logs a message at trace level, checking context cancellation first using a global logger.
// This function is thread-safe.
// See InfoCtx for details about context checking behavior.
func TraceCtx(ctx context.Context, msg string, fields ...interface{}) {
	logMu.RLock()
	l := log
	logMu.RUnlock()
	l.TraceCtx(ctx, msg, fields...)
}

// WarnCtx logs a message at warn level, checking context cancellation first using a global logger.
// This function is thread-safe.
// See InfoCtx for details about context checking behavior.
func WarnCtx(ctx context.Context, msg string, fields ...interface{}) {
	logMu.RLock()
	l := log
	logMu.RUnlock()
	l.WarnCtx(ctx, msg, fields...)
}

// ErrorCtx logs a message at error level, checking context cancellation first using a global logger.
// This function is thread-safe.
// See InfoCtx for details about context checking behavior.
func ErrorCtx(ctx context.Context, msg string, fields ...interface{}) {
	logMu.RLock()
	l := log
	logMu.RUnlock()
	l.ErrorCtx(ctx, msg, fields...)
}

// ErrCtx logs an error with message at error level, checking context cancellation first using a global logger.
// This function is thread-safe.
//
// This function combines error logging with context cancellation checking.
// Unlike ErrorCtx which logs a message at error level, ErrCtx specifically
// logs an error object along with a message.
//
// Parameters:
//   - ctx: Context to check for cancellation
//   - err: Error to log
//   - msg: Log message
//   - fields: Optional key-value pairs
//
// Example usage:
//
//	func processData(ctx context.Context) error {
//	    result, err := fetchData()
//	    if err != nil {
//	        logze.ErrCtx(ctx, err, "Failed to fetch data", "retry_count", 3)
//	        return err
//	    }
//	    return nil
//	}
func ErrCtx(ctx context.Context, err error, msg string, fields ...interface{}) {
	logMu.RLock()
	l := log
	logMu.RUnlock()
	l.ErrCtx(ctx, err, msg, fields...)
}
