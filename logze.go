// Package logze implements a zerolog wrapper providing a convenient and short interface
// for structured logging with an slog-like key-value API.
package logze

import (
	"context"
	"errors"
	"fmt"
	"io"
	"os"
	"regexp"
	"runtime/debug"
	"strings"
	"time"

	"github.com/rs/zerolog"
	"github.com/rs/zerolog/diode"
)

// Logger represents a high-level structured logger with convenient methods.
// It wraps [zerolog.Logger] and provides additional features like error counting,
// message filtering, stack traces, and context integration.
//
// Logger supports:
//   - Multiple log levels (trace, debug, info, warn, error, fatal)
//   - Structured logging with key-value fields
//   - Error counting and stack trace collection
//   - Message filtering and ignoring specific messages
//   - Context integration for request-scoped logging
//   - Conditional logging with *If methods
//   - Formatted logging with *f methods
//
// The zero value of Logger behaves as [zerolog.Nop] (no-op logger).
// Use [New], [NewConsoleJSON], or other constructors to create a functional logger.
//
// Example usage:
//
//	logger := logze.New(logze.C().WithConsoleJSON().WithLevel("info"))
//	logger.Info("User logged in", "user_id", 123, "ip", "192.168.1.1")
//	logger.Err(err, "Failed to process request", "request_id", "abc123")
type Logger struct {
	l             zerolog.Logger
	errCounter    ErrorCounter
	toIgnore      []string
	ignoreMap     map[string]struct{} // Pre-compiled ignore map for O(1) lookup
	toIgnoreRegex []*regexp.Regexp
	stackTrace    bool
	inited        bool
	diodeWriter   *diode.Writer
}

// New creates a new [Logger] instance with the specified configuration and default fields.
//
// Parameters:
//   - cfg: Configuration defining output writers, log level, hooks, and other settings
//   - fields: Optional key-value pairs that will be included in every log message
//
// Default behaviors:
//   - Output: [io.Discard] if no writers specified (logs are discarded)
//   - Level: "info" if not specified
//   - Time format: [time.RFC3339] if not specified
//   - Diode: Enabled by default for non-blocking writes (see warning below)
//
// Fields should be passed as alternating key-value pairs and will be applied to all messages
// produced by this logger instance.
//
// Example usage:
//
//	logger := New(C().WithConsoleJSON().WithLevel("debug"), "service", "api", "version", "1.0")
//	logger.Info("User authenticated", "user_id", 123, "method", "oauth")
//	logger.Err(err, "Database connection failed", "host", "localhost", "port", 5432)
//
// Example output:
//
//	{"level":"info","time":"2023-11-20T18:48:14+03:00","message":"User authenticated","service":"api","version":"1.0","user_id":123,"method":"oauth"}
//	{"level":"error","time":"2023-11-20T18:48:14+03:00","error":"Database connection failed","message":"Database connection failed","service":"api","version":"1.0","host":"localhost","port":5432}
//
// ⚠️  IMPORTANT: Diode Writer Behavior
//
// By default, New() enables a diode writer for non-blocking log writes. This prevents
// logging from blocking your application but requires time to flush messages to the output.
// If your application exits immediately after logging, messages may be lost.
//
// Solutions:
//   - Call logger.Close() or logger.CloseDiode() before application exit
//   - Use cfg.WithNoDiode() to disable diode and ensure immediate writes
//   - Note: Disabling diode may cause blocking if writing to stderr with high log volume
//
// ⚠️  IMPORTANT: Global Time Format Side Effect
//
// Setting cfg.TimeFieldFormat modifies the global zerolog.TimeFieldFormat variable,
// which affects ALL zerolog loggers in the application, not just this logze instance.
// This is a limitation of the underlying zerolog library. If you have multiple logger
// configurations with different time formats, the last one created will take effect
// for all loggers.
func New(cfg Config, fields ...interface{}) Logger {
	if len(cfg.Writers) == 0 || cfg.Level == LevelDisabled {
		cfg.Writers = []io.Writer{io.Discard}
	}
	if cfg.Level == "" {
		cfg.Level = LevelInfo
	}
	if cfg.TimeFieldFormat == "" {
		cfg.TimeFieldFormat = time.RFC3339
	}
	// Only touch the zerolog global when the format actually changes: writing it
	// unconditionally races with every concurrent log call in the process.
	// zerolog's default is already time.RFC3339, so the default path never writes.
	if zerolog.TimeFieldFormat != cfg.TimeFieldFormat {
		zerolog.TimeFieldFormat = cfg.TimeFieldFormat
	}

	level, err := zerolog.ParseLevel(cfg.Level)
	if err != nil {
		panic("cannot parse level=" + cfg.Level)
	}

	output := cfg.Writers[0]
	if len(cfg.Writers) > 1 {
		output = zerolog.MultiLevelWriter(cfg.Writers...)
	}

	var diodeWriter *diode.Writer
	if !cfg.NoDiode {
		if cfg.DiodeSize == 0 {
			cfg.DiodeSize = DefaultDiodeSize
		}
		if cfg.DiodePollingInterval == 0 {
			cfg.DiodePollingInterval = DefaultDiodePollingInterval
		}
		if cfg.UseDiodeWaiter {
			cfg.DiodePollingInterval = 0
		}
		if cfg.DiodeAlertFunc == nil {
			cfg.DiodeAlertFunc = func(missed int) {
				fmt.Fprintf(os.Stderr, "WRN: logger dropped %d messages\n", missed)
			}
		}
		// To fix problem of blocking goroutine when writing in Stderr
		// https://github.com/cloudfoundry/go-diodes
		w := diode.NewWriter(output, cfg.DiodeSize, cfg.DiodePollingInterval, cfg.DiodeAlertFunc)
		output = w
		diodeWriter = &w
	}

	temp := zerolog.New(output).With().Timestamp().Fields(fields)
	if cfg.AddCaller {
		temp = temp.CallerWithSkipFrameCount(cfg.CallerSkipFrameCount)
	}
	l := temp.Logger().Level(level)

	if len(cfg.Hooks) > 0 {
		l = l.Hook(cfg.Hooks...)
	}
	if cfg.Hook != nil {
		l = l.Hook(cfg.Hook)
	}
	if cfg.Sampler != nil {
		l = l.Sample(cfg.Sampler)
	}

	// Pre-compile ignore map for O(1) lookups
	ignoreMap := make(map[string]struct{}, len(cfg.ToIgnore))
	for _, ignore := range cfg.ToIgnore {
		ignoreMap[ignore] = struct{}{}
	}

	return Logger{
		l:             l,
		toIgnore:      cfg.ToIgnore,
		ignoreMap:     ignoreMap,
		toIgnoreRegex: cfg.ToIgnoreRegex,
		errCounter:    cfg.ErrorCounter,
		stackTrace:    cfg.StackTrace,
		inited:        true,
		diodeWriter:   diodeWriter,
	}
}

// NewFromZerolog creates a new [Logger] wrapping an existing [zerolog.Logger].
//
// This function is useful when you already have a configured zerolog.Logger instance
// and want to use logze's additional features like error counting, message filtering,
// and enhanced context integration.
//
// Note: The returned Logger will not have diode writer, error counter, or message
// filtering features unless configured separately using With* methods.
//
// Example usage:
//
//	zlogger := zerolog.New(os.Stdout).With().Timestamp().Logger()
//	logger := NewFromZerolog(zlogger).WithSimpleErrorCounter().WithLevel("debug")
//	logger.Info("Migrated from zerolog", "component", "auth")
func NewFromZerolog(l zerolog.Logger) Logger {
	return Logger{
		l:      l,
		inited: true,
	}
}

// NewConsoleJSON creates a new [Logger] with JSON output to stderr and info level.
//
// This is a convenience constructor for quick setup during development or simple applications.
// It's equivalent to: New(NewConfig().WithConsoleJSON().WithLevel("info"), fields...)
//
// The logger outputs structured JSON logs to stderr with timestamps and includes any
// provided fields in all log messages.
//
// Example usage:
//
//	logger := NewConsoleJSON("service", "web-api", "version", "2.1.0")
//	logger.Info("Server starting", "port", 8080)
//	// Output: {"level":"info","time":"2023-11-20T18:48:14+03:00","message":"Server starting","service":"web-api","version":"2.1.0","port":8080}
//
// For more configuration options, use [New] with a custom [Config].
func NewConsoleJSON(fields ...interface{}) Logger {
	return New(NewConfig().WithConsoleJSON(), fields...)
}

// Nop creates a no-operation logger that discards all log messages.
//
// This is useful for testing, disabling logging in certain conditions, or as a safe
// default when a logger is optional. All logging methods will execute without error
// but produce no output.
//
// Example usage:
//
//	var logger logze.Logger
//	if verbose {
//		logger = logze.NewConsoleJSON()
//	} else {
//		logger = logze.Nop()
//	}
//	logger.Info("This may or may not be logged")
//
// Nop loggers are safe for concurrent use and have minimal performance overhead.
func Nop() Logger {
	return Logger{l: zerolog.Nop()}
}

// CloseDiode gracefully shuts down the diode writer, ensuring all buffered log messages are flushed.
//
// This method should be called before application shutdown when using the default diode writer
// to prevent loss of buffered log messages. If no diode writer is configured, this method
// returns nil and has no effect.
//
// Example usage:
//
//	logger := logze.New(logze.C().WithConsoleJSON())
//	defer logger.CloseDiode() // Ensure logs are flushed before exit
//
//	logger.Info("Application shutting down")
//	// Without CloseDiode(), this message might be lost
func (l Logger) CloseDiode() error {
	if l.diodeWriter != nil {
		return l.diodeWriter.Close()
	}
	return nil
}

// Close is an alias for [Logger.CloseDiode] that implements [io.Closer].
//
// This allows Logger to be used with defer statements and other APIs that expect
// an io.Closer interface.
//
// Example usage:
//
//	logger := logze.New(logze.C().WithConsoleJSON())
//	defer logger.Close() // Implements io.Closer
func (l Logger) Close() error {
	return l.CloseDiode()
}

type ctxKey struct{}

// AddToContext embeds this logger into a context, making it available for request-scoped logging.
//
// This enables passing logger instances through context chains, which is particularly
// useful in web applications, gRPC services, and other request-oriented architectures.
//
// Disabled loggers are not stored in the context to avoid unnecessary memory usage.
//
// Example usage:
//
//	logger := logze.NewConsoleJSON("request_id", "abc123")
//	ctx = logger.AddToContext(ctx)
//
//	// Pass ctx to other functions
//	processRequest(ctx)
//
//	func processRequest(ctx context.Context) {
//		logger := logze.GetFromContext(ctx)
//		logger.Info("Processing request") // Includes request_id
//	}
func (l Logger) AddToContext(ctx context.Context) context.Context {
	if _, ok := ctx.Value(ctxKey{}).(*Logger); !ok && l.l.GetLevel() == zerolog.Disabled {
		// Do not store disabled logger.
		return ctx
	}
	return context.WithValue(ctx, ctxKey{}, &l)
}

// GetFromContext retrieves a logger instance from the context.
//
// If no logger is found in the context, returns a no-op logger that safely
// discards all log messages. This ensures that logging calls never panic
// even when no logger has been embedded in the context.
//
// Example usage:
//
//	func handleRequest(ctx context.Context) {
//		logger := logze.GetFromContext(ctx)
//		logger.Info("Handling request") // Safe even if no logger in context
//
//		// Add request-specific fields
//		logger = logger.With("user_id", getUserID(ctx))
//		logger.Debug("User authenticated")
//	}
//
// See [Logger.AddToContext] for embedding loggers into contexts.
func GetFromContext(ctx context.Context) Logger {
	l, ok := ctx.Value(ctxKey{}).(*Logger)
	if !ok || l == nil {
		return Nop()
	}
	return *l
}

// Update replaces the logger's configuration and fields with new values.
//
// This method modifies the logger in-place, replacing its underlying configuration,
// output writers, level, error counter, and other settings. Any existing fields
// are replaced with the new ones.
//
// ⚠️  THREAD SAFETY WARNING
//
// Update is NOT safe for concurrent use. Ensure no other goroutines are using
// this logger instance while calling Update. For concurrent scenarios, create
// a new logger instance instead.
//
// Example usage:
//
//	logger := logze.NewConsoleJSON()
//	logger.Info("Initial message")
//
//	// Reconfigure for production use
//	prodConfig := logze.C(file).WithLevel("warn").WithSimpleErrorCounter()
//	logger.Update(prodConfig, "environment", "production")
//	logger.Info("This won't be logged due to warn level")
func (l *Logger) Update(cfg Config, fields ...interface{}) {
	// Close the old diode writer to prevent goroutine leak
	if l.diodeWriter != nil {
		l.diodeWriter.Close()
	}

	*l = New(cfg, fields...)
}

// NotInited reports whether the logger has been properly initialized.
//
// Returns true if the logger is a zero-value struct that hasn't been created
// through any constructor ([New], [NewConsoleJSON], etc.). Zero-value loggers
// behave as no-op loggers but this method can be used to detect uninitialized state.
//
// Example usage:
//
//	var logger logze.Logger
//	if logger.NotInited() {
//		logger = logze.NewConsoleJSON()
//	}
//	logger.Info("Logger is now ready")
func (l Logger) NotInited() bool {
	return !l.inited
}

// WithFields creates a new logger with additional fields that will be included in every log message.
//
// Fields should be provided as alternating key-value pairs. These fields will be
// added to any fields already configured on the logger and will appear in all
// subsequent log messages from the returned logger.
//
// This method returns a new logger instance; the original logger is not modified.
//
// ⚠️ IMPORTANT: Resource Sharing
//
// The returned logger shares certain resources with the parent logger:
//   - diodeWriter: Both loggers use the same underlying diode writer. Closing either
//     the parent or derived logger will affect both. If you need independent lifecycle
//     management, create a new logger with New() instead of deriving.
//   - errCounter: The error counter is shared, so errors logged by either logger
//     increment the same counter.
//   - toIgnore/ignoreMap: Message filtering configuration is shared.
//
// Example usage:
//
//	baseLogger := logze.NewConsoleJSON("service", "api")
//	requestLogger := baseLogger.WithFields("request_id", "abc123", "user_id", 456)
//	requestLogger.Info("Processing request")
//	// Output includes: "service":"api","request_id":"abc123","user_id":456
//
//	// ⚠️ Be careful with lifecycle:
//	defer baseLogger.Close() // This also closes requestLogger's diode writer!
//
// Fields can be any JSON-serializable values: strings, numbers, booleans, slices, maps.
func (l Logger) WithFields(fields ...interface{}) Logger {
	return Logger{
		l:             l.l.With().Fields(fields).Logger(),
		errCounter:    l.errCounter,
		toIgnore:      l.toIgnore,
		ignoreMap:     l.ignoreMap,
		toIgnoreRegex: l.toIgnoreRegex,
		stackTrace:    l.stackTrace,
		inited:        l.inited,
		diodeWriter:   l.diodeWriter,
	}
}

// With is a convenient shorthand for [Logger.WithFields].
//
// It creates a new logger with additional fields, identical to calling WithFields.
// This shorter name makes chaining more readable in complex logging setups.
//
// Example usage:
//
//	logger := logze.NewConsoleJSON().With("component", "auth").With("version", "2.1")
//	logger.Info("Authentication module initialized")
func (l Logger) With(fields ...interface{}) Logger {
	return l.WithFields(fields...)
}

// WithLevel creates a new logger with the specified log level.
//
// Valid levels are: "trace", "debug", "info", "warn", "error", "fatal", "disabled".
// Messages below the specified level will be discarded. If an empty string is
// provided, the logger is returned unchanged.
//
// This method returns a new logger instance; the original logger is not modified.
//
// Example usage:
//
//	logger := logze.NewConsoleJSON().WithLevel("warn")
//	logger.Debug("This won't be logged")  // Below warn level
//	logger.Warn("This will be logged")    // At warn level
//	logger.Error("This will be logged")   // Above warn level
//
// Level hierarchy: trace < debug < info < warn < error < fatal
func (l Logger) WithLevel(level string) Logger {
	if level == "" {
		return l
	}
	lvl, err := zerolog.ParseLevel(level)
	if err != nil {
		panic("cannot parse level=" + level)
	}
	l.l = l.l.Level(lvl)
	return l
}

// WithStack creates a new logger with stack trace collection enabled.
//
// When enabled, error logging methods will automatically collect and include
// stack traces in the log output (for Error logs). This is useful for debugging but adds
// performance overhead.
//
// If optional stackTrace argument is false, the stack trace will not be collected.
//
// Example usage:
//
//	logger := logze.NewConsoleJSON().WithStack(true)
//	logger.Err(err, "Database connection failed")
//	// Output will include stack trace information
func (l Logger) WithStack(stackTrace ...bool) Logger {
	l.stackTrace = true
	if len(stackTrace) > 0 && !stackTrace[0] {
		l.stackTrace = false
	}
	return l
}

// WithErrorCounter creates a new logger with the specified error counter.
//
// Error counters track the number of errors logged and can be used for
// monitoring, alerting, or debugging purposes. The counter is incremented
// whenever Err, Error, Fatal, or Panic methods are called.
//
// Example usage:
//
//	counter := &MyCustomErrorCounter{}
//	logger := logze.NewConsoleJSON().WithErrorCounter(counter)
//	logger.Err(err, "Something failed")
//	// counter.Inc(err) was called automatically
func (l Logger) WithErrorCounter(ec ErrorCounter) Logger {
	l.errCounter = ec
	return l
}

// WithSimpleErrorCounter creates a new logger with a built-in atomic error counter.
//
// This is a convenience method that adds a simple thread-safe error counter
// to track the total number of errors logged. Use GetErrorCounter() to access
// the counter and retrieve the count.
//
// Example usage:
//
//	logger := logze.NewConsoleJSON().WithSimpleErrorCounter()
//	logger.Err(err1, "First error")
//	logger.Err(err2, "Second error")
//
//	counter := logger.GetErrorCounter().(*logze.SimpleErrorCounter)
//	fmt.Printf("Total errors: %d", counter.Load()) // Prints: Total errors: 2
func (l Logger) WithSimpleErrorCounter() Logger {
	l.errCounter = newSimpleErrorCounter()
	return l
}

// WithToIgnore creates a new logger that filters out specified messages.
//
// Messages that exactly match any of the provided strings, or contain them as
// substrings, will be discarded and not logged. This is useful for reducing
// noise from repeated or unimportant messages.
//
// Example usage:
//
//	logger := logze.NewConsoleJSON().WithToIgnore("health check", "ping")
//	logger.Info("health check")     // This won't be logged
//	logger.Info("ping from server") // This won't be logged (contains "ping")
//	logger.Info("user login")       // This will be logged
//
// Note: Message filtering happens before field processing for performance.
func (l Logger) WithToIgnore(toIgnore ...string) Logger {
	l.toIgnore = toIgnore
	// Rebuild ignore map for O(1) lookups
	l.ignoreMap = make(map[string]struct{}, len(toIgnore))
	for _, ignore := range toIgnore {
		l.ignoreMap[ignore] = struct{}{}
	}
	return l
}

// WithCaller creates a new logger that includes caller information in log messages.
//
// The callerSkipFrameCount parameter determines how many stack frames to skip
// when determining the caller. This is useful when wrapping the logger in
// other functions and you want to report the actual caller, not the wrapper.
//
// Example usage:
//
//	logger := logze.NewConsoleJSON().WithCaller(2)
//	logger.Info("Message with caller info")
//	// Output includes: "caller":"/path/to/file.go:123"
//
// Use [WithDefaultCaller] for the standard skip count.
func (l Logger) WithCaller(callerSkipFrameCount int) Logger {
	l.l = l.l.With().CallerWithSkipFrameCount(callerSkipFrameCount).Logger()
	return l
}

// WithDefaultCaller creates a new logger with caller information using the default skip frame count.
//
// This is equivalent to WithCaller(DefaultCallerSkipFrameCount) and is the
// recommended way to enable caller information for most use cases.
//
// Example usage:
//
//	logger := logze.NewConsoleJSON().WithDefaultCaller()
//	logger.Info("Message with caller info")
//	// Output includes: "caller":"/path/to/file.go:123"
func (l Logger) WithDefaultCaller() Logger {
	l.l = l.l.With().CallerWithSkipFrameCount(DefaultCallerSkipFrameCount).Logger()
	return l
}

// WithSampler creates a new logger with the specified sampling configuration.
//
// Sampling allows you to log only a subset of messages to reduce log volume
// and improve performance. The sampler determines which messages are logged
// and which are discarded.
//
// Example usage:
//
//	// Log only 10% of debug messages
//	sampler := zerolog.RandomSampler(10)
//	logger := logze.NewConsoleJSON().WithSampler(sampler)
func (l Logger) WithSampler(sampler zerolog.Sampler) Logger {
	l.l = l.l.Sample(sampler)
	return l
}

// WithPercentageSampler creates a new logger with a percentage-based sampler.
//
// This method creates a sampler that logs only a specified percentage of messages
// at the specified levels. The percentage is calculated based on the number of
// messages logged at each level.
//
// Example usage:
//
//	logger := logze.NewConsoleJSON().WithPercentageSampler(0.1, "debug", "info")
//	logger.Debug("This will be logged 10% of the time")
//	logger.Info("This will be logged 10% of the time")
func (l Logger) WithPercentageSampler(percentage float64, levels ...string) Logger {
	sampler := percentageSampler(percentage)
	return l.WithSampler(getLevelSampler(sampler, levels...))
}

// WithBurstSampler creates a new logger with a burst-based sampler.
//
// This method creates a sampler that logs only a specified number of messages
// at the specified levels within a given time period. The burst is the maximum
// number of messages that can be logged within the period.
//
// Example usage:
//
//	logger := logze.NewConsoleJSON().WithBurstSampler(0.1, 100, 1*time.Second, "debug", "info")
//	logger.Debug("Logged up to 100 times per second, then 10% of the time")
//	logger.Info("Logged up to 100 times per second, then 10% of the time")
func (l Logger) WithBurstSampler(percentage float64, burst int, period time.Duration, levels ...string) Logger {
	sampler := burstSampler(percentage, burst, period)
	return l.WithSampler(getLevelSampler(sampler, levels...))
}

// WithMaxSampler creates a new logger with a max-based sampler.
//
// This method creates a sampler that logs only a specified number of messages
// at the specified levels within a given time period. The max is the maximum
// number of messages that can be logged within the period.
//
// Example usage:
//
//	logger := logze.NewConsoleJSON().WithMaxSampler(100, 1*time.Second, "debug", "info")
//	logger.Debug("Logged at most 100 times per second")
//	logger.Info("Logged at most 100 times per second")
func (l Logger) WithMaxSampler(max int, period time.Duration, levels ...string) Logger {
	sampler := burstSampler(0, max, period)
	return l.WithSampler(getLevelSampler(sampler, levels...))
}

// Trace logs a message at trace level with optional fields and caller information.
//
// Trace level is the most verbose logging level, typically used for detailed
// debugging information that you normally wouldn't want in production.
// This method automatically includes caller information.
//
// Fields should be provided as alternating key-value pairs.
//
// Example usage:
//
//	logger.Trace("Function entry", "function", "processUser", "user_id", 123)
//	// Output: {"level":"trace","caller":"main.go:45","message":"Function entry","function":"processUser","user_id":123}
func (l Logger) Trace(msg string, fields ...interface{}) {
	l.log(l.l.Trace().Caller(1), msg, fields)
}

// Tracef logs a formatted message at trace level with caller information.
//
// This is the formatted version of Trace. Format verbs are processed using fmt.Sprintf.
// Any additional arguments beyond the format placeholders are treated as structured fields.
//
// Example usage:
//
//	logger.Tracef("Processing %d items for user %s", count, username, "request_id", "abc123")
//	// Formats the message, then adds request_id as a structured field
func (l Logger) Tracef(msg string, args ...interface{}) {
	l.logf(l.l.Trace().Caller(1), msg, args)
}

// TraceIf conditionally logs a message at trace level with caller information.
//
// This is a convenience method that only logs if the condition is true.
// Useful for avoiding expensive field preparation when logging might be skipped.
//
// Example usage:
//
//	logger.TraceIf(debugMode, "Debug info", "state", expensiveStateCapture())
//	// Only evaluates expensiveStateCapture() if debugMode is true
func (l Logger) TraceIf(condition bool, msg string, fields ...interface{}) {
	if condition {
		l.Trace(msg, fields...)
	}
}

// Debug logs a message at debug level with optional structured fields.
//
// Debug level is used for diagnostic information that's useful during development
// and troubleshooting but typically disabled in production for performance.
//
// Fields should be provided as alternating key-value pairs.
//
// Example usage:
//
//	logger.Debug("Cache hit", "key", "user:123", "ttl", "5m", "size", 1024)
//	// Output: {"level":"debug","message":"Cache hit","key":"user:123","ttl":"5m","size":1024}
func (l Logger) Debug(msg string, fields ...interface{}) {
	l.log(l.l.Debug(), msg, fields)
}

// Debugf logs a formatted message at debug level.
//
// This is the formatted version of Debug. Format verbs are processed using fmt.Sprintf.
// Any additional arguments beyond the format placeholders are treated as structured fields.
//
// Example usage:
//
//	logger.Debugf("Query executed in %dms", duration, "query", sqlQuery, "rows", rowCount)
//	// Formats the duration into the message, then adds query and rows as fields
func (l Logger) Debugf(msg string, args ...interface{}) {
	l.logf(l.l.Debug(), msg, args)
}

// DebugIf conditionally logs a message at debug level.
//
// This is a convenience method that only logs if the condition is true.
// Useful for conditional debugging without cluttering code with if statements.
//
// Example usage:
//
//	logger.DebugIf(cfg.VerboseSQL, "SQL executed", "query", query, "duration", elapsed)
//	// Only logs SQL information when verbose SQL logging is enabled
func (l Logger) DebugIf(condition bool, msg string, fields ...interface{}) {
	if condition {
		l.Debug(msg, fields...)
	}
}

// Info logs a message at info level with optional structured fields.
//
// Info level is the standard logging level for general application information
// such as startup messages, important business events, and successful operations.
// This level is typically enabled in production.
//
// Fields should be provided as alternating key-value pairs.
//
// Example usage:
//
//	logger.Info("User authenticated", "user_id", 123, "method", "oauth", "ip", "192.168.1.1")
//	// Output: {"level":"info","message":"User authenticated","user_id":123,"method":"oauth","ip":"192.168.1.1"}
func (l Logger) Info(msg string, fields ...interface{}) {
	l.log(l.l.Info(), msg, fields)
}

// Infof logs a formatted message at info level.
//
// This is the formatted version of Info. Format verbs are processed using fmt.Sprintf.
// Any additional arguments beyond the format placeholders are treated as structured fields.
//
// Example usage:
//
//	logger.Infof("Server started on port %d", port, "environment", env, "version", version)
//	// Formats the port into the message, then adds environment and version as fields
func (l Logger) Infof(msg string, args ...interface{}) {
	l.logf(l.l.Info(), msg, args)
}

// InfoIf conditionally logs a message at info level.
//
// This is a convenience method that only logs if the condition is true.
// Useful for conditional informational logging based on configuration or state.
//
// Example usage:
//
//	logger.InfoIf(cfg.LogSuccessfulRequests, "Request completed", "duration", elapsed, "status", 200)
//	// Only logs successful requests when the feature is enabled
func (l Logger) InfoIf(condition bool, msg string, fields ...interface{}) {
	if condition {
		l.Info(msg, fields...)
	}
}

// Warn logs a message at warning level with optional structured fields.
//
// Warning level indicates potentially problematic situations that don't prevent
// the application from continuing but should be investigated. Examples include
// deprecated API usage, fallback mechanisms, or performance issues.
//
// Fields should be provided as alternating key-value pairs.
//
// Example usage:
//
//	logger.Warn("API rate limit approaching", "current", 950, "limit", 1000, "user_id", 123)
//	// Output: {"level":"warn","message":"API rate limit approaching","current":950,"limit":1000,"user_id":123}
func (l Logger) Warn(msg string, fields ...interface{}) {
	l.log(l.l.Warn(), msg, fields)
}

// Warnf logs a formatted message at warning level.
//
// This is the formatted version of Warn. Format verbs are processed using fmt.Sprintf.
// Any additional arguments beyond the format placeholders are treated as structured fields.
//
// Example usage:
//
//	logger.Warnf("Cache miss for key %s", key, "cache_type", "redis", "fallback", "database")
//	// Formats the key into the message, then adds cache_type and fallback as fields
func (l Logger) Warnf(msg string, args ...interface{}) {
	l.logf(l.l.Warn(), msg, args)
}

// WarnIf conditionally logs a message at warning level.
//
// This is a convenience method that only logs if the condition is true.
// Useful for conditional warnings based on thresholds or configuration.
//
// Example usage:
//
//	logger.WarnIf(responseTime > threshold, "Slow response", "duration", responseTime, "threshold", threshold)
//	// Only warns when response time exceeds the configured threshold
func (l Logger) WarnIf(condition bool, msg string, fields ...interface{}) {
	if condition {
		l.Warn(msg, fields...)
	}
}

// Err logs an error with additional context at error level.
//
// This method logs both the error and a descriptive message with optional structured fields.
// The error is automatically added to the log entry and, if error counting is enabled,
// increments the error counter. Stack traces are included if enabled via WithStack.
//
// Fields should be provided as alternating key-value pairs.
//
// Example usage:
//
//	logger.Err(err, "Database connection failed", "host", "localhost", "database", "users", "retry", 3)
//	// Output: {"level":"error","error":"connection refused","message":"Database connection failed","host":"localhost","database":"users","retry":3}
func (l Logger) Err(err error, msg string, fields ...interface{}) {
	ev := l.l.Error()
	if ev == nil {
		return
	}
	ev, _ = l.setErrorWithStack(ev, false, err)
	l.log(ev, msg, fields)
}

// Errf logs an error with a formatted message at error level.
//
// This is the formatted version of Err. Format verbs are processed using fmt.Sprintf.
// The error is automatically handled and any additional arguments beyond format placeholders
// are treated as structured fields.
//
// Example usage:
//
//	logger.Errf(err, "Failed to process %d items", count, "batch_id", batchID, "remaining", remaining)
//	// Formats the count into the message, then adds batch_id and remaining as fields
func (l Logger) Errf(err error, msg string, args ...interface{}) {
	ev := l.l.Error()
	if ev == nil {
		return
	}
	ev, _ = l.setErrorWithStack(ev, false, err)
	l.logf(ev, msg, args)
}

// ErrIf conditionally logs an error with additional context.
//
// This is a convenience method that only logs if the condition is true.
// Useful for conditional error logging based on severity or configuration.
//
// Example usage:
//
//	logger.ErrIf(err != nil, err, "Operation failed", "operation", "user_update", "user_id", userID)
//	// Only logs if an error actually occurred
func (l Logger) ErrIf(condition bool, err error, msg string, fields ...interface{}) {
	if condition {
		l.Err(err, msg, fields...)
	}
}

// Error logs an error message at error level without an associated error object.
//
// Use this method when you have an error condition but no specific error object,
// or when logging error-level information that isn't necessarily about a failure.
//
// Fields should be provided as alternating key-value pairs.
//
// Example usage:
//
//	logger.Error("Validation failed", "field", "email", "value", userEmail, "reason", "invalid format")
//	// Output: {"level":"error","message":"Validation failed","field":"email","value":"user@domain","reason":"invalid format"}
func (l Logger) Error(msg string, fields ...interface{}) {
	l.log(l.l.Error(), msg, fields)
}

// Errorf logs a formatted error message at error level.
//
// This is the formatted version of Error. Format verbs are processed using fmt.Sprintf.
// Any additional arguments beyond the format placeholders are treated as structured fields.
//
// Example usage:
//
//	logger.Errorf("Rate limit exceeded: %d requests in %s", count, duration, "client_ip", clientIP)
//	// Formats the count and duration into the message, then adds client_ip as a field
func (l Logger) Errorf(msg string, args ...interface{}) {
	l.logf(l.l.Error(), msg, args)
}

// ErrorIf conditionally logs an error message at error level.
//
// This is a convenience method that only logs if the condition is true.
// Useful for conditional error-level logging without associated error objects.
//
// Example usage:
//
//	logger.ErrorIf(attempts > maxRetries, "Max retries exceeded", "attempts", attempts, "max", maxRetries)
//	// Only logs when retry limit is actually exceeded
func (l Logger) ErrorIf(condition bool, msg string, fields ...interface{}) {
	if condition {
		l.Error(msg, fields...)
	}
}

// ErrStack logs an error with its full stack trace as the message.
//
// This method formats the error with its complete stack trace and logs it as the
// main message content. This is useful for debugging when you need to see the
// exact call path that led to an error.
//
// If the error doesn't already contain stack trace information, it will be wrapped
// to include stack trace details. Errors from github.com/maxbolgarin/errm are
// automatically supported for enhanced stack trace formatting.
//
// Example usage:
//
//	logger.ErrStack(err, "component", "payment", "transaction_id", txID)
//	// Logs the full stack trace as the message, with component and transaction_id as fields
//
// Note: The stack trace output can be quite verbose. Consider using Err() with WithStack()
// for more structured error logging in production.
func (l Logger) ErrStack(err error, fields ...interface{}) {
	l.log(l.l.Error(), fmt.Sprintf("%+v", WithStack(err)), fields)
}

// ErroError is an interface that represents an error that can be logged with a message and fields.
// This in an Error interface from github.com/maxbolgarin/erro.
type ErroError interface {
	error
	Message() string
	AllFields() []interface{}
}

// Erro logs an error with a message and fields.
//
// This method logs the error message and fields.
//
// Example usage:
//
//	err := erro.New("error occurred", "user_id", userID)
//	logger.Erro(err, "")
//	// Logs the error message with user_id as a field
//
// Note: This method is useful when you have an error that implements the ErroError interface.
// If you have a regular error, use the Err() method instead.
func (l Logger) Erro(errRaw error, msg string, fields ...interface{}) {
	err, ok := errRaw.(ErroError)
	if !ok {
		l.Err(errRaw, msg, fields...)
		return
	}
	ev := l.l.Error()
	if ev == nil {
		return
	}
	erroFields := err.AllFields()
	if l.errCounter != nil {
		l.errCounter.Inc(err)
	}
	if msg == "" {
		l.log(ev, err.Message(), append(fields, erroFields...))
		return
	}

	newFields := make([]interface{}, 0, len(fields)+len(erroFields)+2)
	newFields = append(newFields, "error", err.Message())
	newFields = append(newFields, fields...)
	newFields = append(newFields, erroFields...)
	l.log(ev, msg, newFields)
}

// Fatal logs a fatal error message and immediately terminates the program with exit code 1.
//
// ⚠️  WARNING: This method calls os.Exit(1) after logging, terminating the program.
// Use this only for unrecoverable errors where the application cannot continue.
// No deferred functions will run after this call.
//
// Arguments are concatenated using fmt.Sprint. If error counting is enabled,
// the error counter is incremented before termination.
//
// Example usage:
//
//	logger.Fatal("Database initialization failed - cannot continue")
//	// Logs the message and immediately exits with code 1
//
// For recoverable errors, use Error() or Err() instead.
func (l Logger) Fatal(v ...interface{}) {
	s := fmt.Sprint(v...)
	l.incErrorCounter(errors.New(s))
	l.log(l.l.WithLevel(zerolog.FatalLevel), s, nil)
	l.CloseDiode() //nolint:errcheck // flush buffered logs before exit
	osExit(1)
}

// osExit is an indirection over [os.Exit] so tests can intercept Fatal* methods.
var osExit = os.Exit

// Fatalf logs a formatted fatal error message and immediately terminates the program with exit code 1.
//
// ⚠️  WARNING: This method calls os.Exit(1) after logging, terminating the program.
// Use this only for unrecoverable errors where the application cannot continue.
// No deferred functions will run after this call.
//
// Arguments are formatted using fmt.Sprintf. If error counting is enabled,
// the error counter is incremented before termination.
//
// Example usage:
//
//	logger.Fatalf("Failed to bind to port %d: %v", port, err)
//	// Logs the formatted message and immediately exits with code 1
func (l Logger) Fatalf(format string, args ...interface{}) {
	l.incErrorCounter(fmt.Errorf(format, args...))
	l.logf(l.l.WithLevel(zerolog.FatalLevel), format, args)
	l.CloseDiode() //nolint:errcheck // flush buffered logs before exit
	osExit(1)
}

// FatalIf conditionally logs a fatal error and terminates the program.
//
// ⚠️  WARNING: This method calls os.Exit(1) if the condition is true.
// Use this only for unrecoverable errors where the application cannot continue.
// No deferred functions will run after this call.
//
// Example usage:
//
//	logger.FatalIf(config == nil, "Configuration file is required")
//	// Only exits if config is actually nil
func (l Logger) FatalIf(condition bool, v ...interface{}) {
	if condition {
		l.Fatal(v...)
	}
}

// Fatalln logs a fatal error message with a newline and immediately terminates the program with exit code 1.
//
// ⚠️  WARNING: This method calls os.Exit(1) after logging, terminating the program.
// Use this only for unrecoverable errors where the application cannot continue.
// No deferred functions will run after this call.
//
// Arguments are concatenated using fmt.Sprintln (which adds spaces between
// arguments and a newline at the end). If error counting is enabled,
// the error counter is incremented before termination.
//
// Example usage:
//
//	logger.Fatalln("Critical error:", err)
//	// Logs "Critical error: <error message>\n" and immediately exits with code 1
func (l Logger) Fatalln(v ...interface{}) {
	s := fmt.Sprintln(v...)
	l.incErrorCounter(errors.New(s))
	l.log(l.l.WithLevel(zerolog.FatalLevel), s, nil)
	l.CloseDiode() //nolint:errcheck // flush buffered logs before exit
	osExit(1)
}

// Panic logs a message at fatal level and immediately panics with the message.
//
// ⚠️  WARNING: This method calls panic() after logging, which will unwind the stack
// and terminate the current goroutine unless recovered. Use this only for truly
// exceptional conditions that represent programming errors or unrecoverable states.
//
// Arguments are concatenated using fmt.Sprint. If error counting is enabled,
// the error counter is incremented before panicking.
//
// Note: Unlike Fatal, Panic does not flush the diode writer (the panic may be
// recovered and the logger reused). With the default diode enabled, the panic
// log entry may be lost if the process terminates before the next flush.
//
// Example usage:
//
//	logger.Panic("Invariant violated: user cannot be nil at this point")
//	// Logs the message and immediately panics with the message string
//
// For recoverable errors, use Error() or Err() instead.
func (l Logger) Panic(v ...interface{}) {
	s := fmt.Sprint(v...)
	l.incErrorCounter(errors.New(s))
	l.log(l.l.WithLevel(zerolog.FatalLevel), s, nil)
	panic(s)
}

// Panicf logs a formatted message at fatal level and immediately panics with the formatted message.
//
// ⚠️  WARNING: This method calls panic() after logging, which will unwind the stack
// and terminate the current goroutine unless recovered. Use this only for truly
// exceptional conditions that represent programming errors or unrecoverable states.
//
// Arguments are formatted using fmt.Sprintf. If error counting is enabled,
// the error counter is incremented before panicking.
//
// Note: Unlike Fatal, Panic does not flush the diode writer (the panic may be
// recovered and the logger reused). With the default diode enabled, the panic
// log entry may be lost if the process terminates before the next flush.
//
// Example usage:
//
//	logger.Panicf("Buffer overflow: tried to write %d bytes to %d byte buffer", writeSize, bufSize)
//	// Logs and panics with the formatted message
func (l Logger) Panicf(format string, args ...interface{}) {
	l.incErrorCounter(fmt.Errorf(format, args...))
	l.logf(l.l.WithLevel(zerolog.FatalLevel), format, args)
	panic(fmt.Sprintf(format, args...))
}

// PanicIf conditionally logs a message and panics if the condition is true.
//
// ⚠️  WARNING: This method calls panic() if the condition is true, which will unwind
// the stack and terminate the current goroutine unless recovered.
//
// Example usage:
//
//	logger.PanicIf(len(items) == 0, "Items slice cannot be empty")
//	// Only panics if the slice is actually empty
func (l Logger) PanicIf(condition bool, v ...interface{}) {
	if condition {
		l.Panic(v...)
	}
}

// Panicln logs a message with a newline at fatal level and immediately panics with the message.
//
// ⚠️  WARNING: This method calls panic() after logging, which will unwind the stack
// and terminate the current goroutine unless recovered. Use this only for truly
// exceptional conditions that represent programming errors or unrecoverable states.
//
// Arguments are concatenated using fmt.Sprintln (which adds spaces between
// arguments and a newline at the end). If error counting is enabled,
// the error counter is incremented before panicking.
//
// Note: Unlike Fatal, Panic does not flush the diode writer (the panic may be
// recovered and the logger reused). With the default diode enabled, the panic
// log entry may be lost if the process terminates before the next flush.
//
// Example usage:
//
//	logger.Panicln("Critical invariant failed:", details)
//	// Logs "Critical invariant failed: <details>\n" and panics with that string
func (l Logger) Panicln(v ...interface{}) {
	s := fmt.Sprintln(v...)
	l.incErrorCounter(errors.New(s))
	l.log(l.l.WithLevel(zerolog.FatalLevel), s, nil)
	panic(s)
}

// Print logs a message without any level designation using fmt.Sprint to format arguments.
//
// This method outputs messages that don't fit into standard log levels or when you
// want unstructured output. The message appears without level, timestamp, or other
// metadata depending on your output configuration.
//
// Returns immediately if no arguments are provided.
//
// Example usage:
//
//	logger.Print("System starting up...")
//	logger.Print("Current time:", time.Now())
//	// Output depends on your logger configuration but typically just the message content
func (l Logger) Print(v ...interface{}) {
	if len(v) == 0 {
		return
	}
	l.log(l.l.Log(), fmt.Sprint(v...), nil)
}

// PrintIf conditionally logs a message without level designation.
//
// This is a convenience method that only logs if the condition is true.
// Useful for conditional output without cluttering code with if statements.
//
// Example usage:
//
//	logger.PrintIf(verbose, "Detailed startup information:", details)
//	// Only prints when verbose mode is enabled
func (l Logger) PrintIf(condition bool, v ...interface{}) {
	if condition {
		l.Print(v...)
	}
}

// PrintStack logs the current goroutine's stack trace with optional additional fields.
//
// This method captures and logs the call stack from the point where it's called,
// which is useful for debugging unexpected code paths or understanding call flow.
// Any additional arguments are treated as structured fields.
//
// Example usage:
//
//	logger.PrintStack("component", "router", "unexpected_path", requestPath)
//	// Logs the full stack trace with component and unexpected_path as structured fields
//
// Note: Stack traces can be quite verbose. Use sparingly in production code.
func (l Logger) PrintStack(v ...interface{}) {
	stack := debug.Stack()
	l.log(l.l.Log(), string(stack), v)
}

// Printf logs a formatted message without level designation.
//
// This method formats the message using fmt.Sprintf and outputs it without
// standard log level metadata. Any arguments beyond the format placeholders
// are treated as structured fields.
//
// Example usage:
//
//	logger.Printf("Processing %d items", count)
//	logger.Printf("User %s logged in", username, "session_id", sessionID)
//	// Second example formats username into message, adds session_id as field
func (l Logger) Printf(format string, args ...interface{}) {
	l.logf(l.l.Log(), format, args)
}

// Println logs a message with a newline without level designation.
//
// Arguments are formatted using fmt.Sprintln, which adds spaces between arguments
// and a newline at the end. The output appears without standard log level metadata.
//
// Example usage:
//
//	logger.Println("Starting application...")
//	logger.Println("Version:", version, "Build:", buildID)
//	// Output: "Version: 1.0.0 Build: abc123\n" (exact format depends on configuration)
func (l Logger) Println(v ...interface{}) {
	l.log(l.l.Log(), fmt.Sprintln(v...), nil)
}

// Log is an alias for [Logger.Print] that logs a message without level designation.
//
// This method exists for compatibility with standard library logging interfaces
// and provides the same functionality as Print.
//
// Example usage:
//
//	logger.Log("Application event occurred")
//	// Identical to logger.Print("Application event occurred")
func (l Logger) Log(v ...interface{}) {
	l.Print(v...)
}

// Write implements [io.Writer] interface, allowing the logger to be used anywhere an io.Writer is expected.
//
// This method writes the provided bytes directly to the underlying zerolog writer,
// bypassing normal log formatting and structure. The data is written as-is.
//
// Example usage:
//
//	var w io.Writer = logger
//	fmt.Fprintf(w, "Direct write: %s\n", data)
//
//	// Or with standard library log package:
//	log.SetOutput(logger)
//	log.Println("This goes through the logger")
func (l Logger) Write(p []byte) (n int, err error) {
	return l.l.Write(p)
}

// Raw returns direct access to the underlying [zerolog.Logger] for advanced usage.
//
// This method provides access to the native zerolog API when you need functionality
// not exposed by the logze wrapper. Use sparingly, as it bypasses logze's additional
// features like error counting and message filtering.
//
// Example usage:
//
//	rawLogger := logger.Raw()
//	event := rawLogger.Info().
//		Str("custom_field", "value").
//		Dur("elapsed", duration)
//	event.Msg("Custom structured log")
//
// Consider using logger methods instead when possible for consistency.
func (l Logger) Raw() *zerolog.Logger {
	return &l.l
}

// GetErrorCounter returns the error counter associated with this logger, if any.
//
// Returns nil if no error counter was configured via WithErrorCounter or
// WithSimpleErrorCounter. The returned counter can be used to retrieve
// error statistics or reset counts.
//
// Example usage:
//
//	if counter := logger.GetErrorCounter(); counter != nil {
//		if simple, ok := counter.(*logze.SimpleErrorCounter); ok {
//			errorCount := simple.Load()
//			fmt.Printf("Total errors logged: %d\n", errorCount)
//		}
//	}
func (l Logger) GetErrorCounter() ErrorCounter {
	return l.errCounter
}

// GetLevel returns the current log level as a string.
//
// This method inspects the underlying zerolog logger and returns the configured
// log level. Useful for runtime debugging and configuration validation.
func (l Logger) GetLevel() string {
	level := l.l.GetLevel()
	return level.String()
}

// IsEnabled checks if a specific log level is enabled.
//
// This method is useful for conditionally performing expensive operations only
// when the log level would actually be output.
func (l Logger) IsEnabled(level string) bool {
	lvl, err := zerolog.ParseLevel(level)
	if err != nil {
		return false
	}
	return l.l.GetLevel() <= lvl
}

// HasErrorCounter returns true if the logger has an error counter configured.
//
// This method checks whether error counting is enabled for this logger instance.
// Useful for conditional logic that depends on error tracking being available.
func (l Logger) HasErrorCounter() bool {
	return l.errCounter != nil
}

// HasDiode returns true if the logger has a diode writer configured.
//
// This method checks whether non-blocking asynchronous logging is enabled via
// a diode writer. Useful for understanding the logger's performance characteristics.
func (l Logger) HasDiode() bool {
	return l.diodeWriter != nil
}

func (l Logger) log(ev *zerolog.Event, msg string, fields []interface{}) {
	if ev == nil {
		// Level is disabled or the message was sampled out — skip all work.
		return
	}
	if l.shouldIgnore(msg) {
		return
	}

	if len(fields) > 0 {
		ev, fields = l.setErrorWithStack(ev, false, fields...)
		ev = ev.Fields(fields)
	}
	ev.Msg(msg)
}

func (l Logger) logf(ev *zerolog.Event, msg string, args []interface{}) {
	if ev == nil {
		// Level is disabled or the message was sampled out — skip all work.
		return
	}
	if l.shouldIgnore(msg) {
		return
	}

	numberOfFormats, hasW, hasEscape := countFormatVerbs(msg)
	if numberOfFormats > 0 && numberOfFormats <= len(args) {
		ev, args = l.setErrorWithStack(ev, true, args...)
		ev = ev.Fields(args[numberOfFormats:])
		args = args[:numberOfFormats]
		if hasW {
			// fmt supports %w only in fmt.Errorf, so render it as %s
			msg = strings.ReplaceAll(msg, "%w", "%s")
		}
		ev.Msgf(msg, args...)
	} else if numberOfFormats == 0 && len(args) > 0 {
		ev, args = l.setErrorWithStack(ev, false, args...)
		ev = ev.Fields(args)
		if hasEscape {
			// This branch bypasses fmt, so unescape literal percents manually
			msg = strings.ReplaceAll(msg, "%%", "%")
		}
		ev.Msg(msg)
	} else {
		if hasEscape {
			// This branch bypasses fmt, so unescape literal percents manually
			msg = strings.ReplaceAll(msg, "%%", "%")
		}
		ev.Msg(msg)
	}
}

// shouldIgnore reports whether msg matches any configured ignore rule.
func (l Logger) shouldIgnore(msg string) bool {
	// Fast path: exact matches (O(1))
	if len(l.ignoreMap) > 0 {
		if _, exists := l.ignoreMap[msg]; exists {
			return true
		}
	}
	// Slower path: substring matches
	for _, ignore := range l.toIgnore {
		if strings.Contains(msg, ignore) {
			return true
		}
	}
	// Regex patterns
	for _, re := range l.toIgnoreRegex {
		if re.MatchString(msg) {
			return true
		}
	}
	return false
}

// countFormatVerbs counts fmt verbs in format, treating "%%" as a literal
// percent sign. It also reports whether the format contains a "%w" verb and
// whether it contains any escaped percents.
func countFormatVerbs(format string) (n int, hasW, hasEscape bool) {
	for i := 0; i < len(format); i++ {
		if format[i] != '%' {
			continue
		}
		if i+1 < len(format) && format[i+1] == '%' {
			i++ // skip the escaped percent
			hasEscape = true
			continue
		}
		n++
		if i+1 < len(format) && format[i+1] == 'w' {
			hasW = true
		}
	}
	return n, hasW, hasEscape
}

const stackKey = "stack"

func (l Logger) setErrorWithStack(ev *zerolog.Event, inFormat bool, args ...interface{}) (*zerolog.Event, []interface{}) {
	newFields := args
	for i, a := range args {
		err, ok := a.(error)
		if !ok {
			continue
		}
		if l.stackTrace {
			stack := CaptureStackTraceJSON()
			ev = ev.RawJSON(stackKey, stack)
		}

		l.incErrorCounter(err)
		if !inFormat {
			// Remove the error from fields to avoid duplicate logging
			// The error is logged via ev.Err(err) below, so we remove it from the fields
			if i == 0 {
				// Error is at the beginning, remove it
				newFields = args[1:]
			} else if i%2 == 1 {
				// Error is at odd position (value in key-value pair)
				// Remove both the key (at i-1) and the error (at i)
				newFields = make([]interface{}, 0, len(args)-2)
				newFields = append(newFields, args[:i-1]...)
				newFields = append(newFields, args[i+1:]...)
			} else {
				// Error is at even position > 0 (used as a key, unusual but possible)
				// Remove the error and the following value to maintain key-value pairing
				if i+1 < len(args) {
					// Remove error and next value
					newFields = make([]interface{}, 0, len(args)-2)
					newFields = append(newFields, args[:i]...)
					newFields = append(newFields, args[i+2:]...)
				} else {
					// Error is the last element, just remove it
					newFields = args[:i]
				}
			}
		}
		return ev.Err(err), newFields
	}
	return ev, newFields
}

func (l Logger) incErrorCounter(err error) {
	if l.errCounter != nil {
		l.errCounter.Inc(err)
	}
}

// HTTP logs an HTTP request at debug level with standardized fields.
// This is a convenience method for consistent HTTP request logging.
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
//	logger.HTTP("GET", "/api/users", 200, time.Since(start), "user_id", userID)
//
// Example output:
//
//	{"level":"debug","message":"HTTP request","method":"GET","path":"/api/users","status":200,"duration_ms":42}
func (l Logger) HTTP(method, path string, status int, duration time.Duration, fields ...interface{}) {
	l.Debug("HTTP request",
		append([]interface{}{
			"method", method,
			"path", path,
			"status", status,
			"duration_ms", duration.Milliseconds(),
		}, fields...)...)
}

// HTTPError logs an HTTP request error at error level with standardized fields.
// This is a convenience method for consistent HTTP error logging.
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
//	    logger.HTTPError("POST", "/api/orders", 500, err, "order_id", orderID)
//	}
//
// Example output:
//
//	{"level":"error","error":"database connection failed","message":"HTTP request failed","method":"POST","path":"/api/orders","status":500,"order_id":"abc123"}
func (l Logger) HTTPError(method, path string, status int, err error, fields ...interface{}) {
	l.Err(err, "HTTP request failed",
		append([]interface{}{
			"method", method,
			"path", path,
			"status", status,
		}, fields...)...)
}

// HTTPAuto automatically selects the appropriate log level based on HTTP status code and error.
// This is a convenience method for automatic log level detection.
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
//	logger.HTTPAuto("GET", "/api/products", resp.StatusCode, time.Since(start), err, "product_id", prodID)
//
// This method will automatically choose:
//   - Error log for 500+ status or non-nil error
//   - Warn log for 400-499 status
//   - Debug log for 200-399 status
func (l Logger) HTTPAuto(method, path string, status int, duration time.Duration, err error, fields ...interface{}) {
	if err != nil || status >= 500 {
		l.HTTPError(method, path, status, err, fields...)
	} else if status >= 400 {
		l.Warn("HTTP client error",
			append([]interface{}{
				"method", method,
				"path", path,
				"status", status,
				"duration_ms", duration.Milliseconds(),
			}, fields...)...)
	} else {
		l.HTTP(method, path, status, duration, fields...)
	}
}

// Recover recovers from panics and logs them at error level with stack trace.
// This method should be called with defer to catch and log panics.
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
//	func handleRequest(logger logze.Logger, reqID string) {
//	    defer logger.Recover("request_id", reqID)
//
//	    // ... code that might panic
//	    processData()
//	}
//
//	func criticalOperation(logger logze.Logger) {
//	    defer logger.Recover("operation", "critical", "module", "payment")
//
//	    // ... code that might panic
//	}
//
// Example output:
//
//	{"level":"error","message":"... stack trace ...","error":"runtime error: index out of range","request_id":"abc123"}
func (l Logger) Recover(fields ...interface{}) {
	if r := recover(); r != nil {
		stack := debug.Stack()
		f := make([]interface{}, 0, len(fields)+2)
		f = append(f, "error", r)
		f = append(f, fields...)
		l.Error(string(stack), f...)
	}
}

// RecoverWithCallback recovers from panics, logs them, and executes a callback function.
// This method should be called with defer to catch and log panics with custom handling.
//
// If a panic occurs, it:
//   - Recovers from the panic
//   - Logs the panic value at error level with stack trace
//   - Executes the provided callback function with the panic value
//   - Includes any additional fields provided
//
// The callback can be used for:
//   - Incrementing panic metrics/counters
//   - Sending alerts
//   - Performing cleanup
//   - Re-throwing the panic if needed
//
// Parameters:
//   - callback: Function to call if a panic occurs (receives the panic value)
//   - fields: Optional key-value pairs to include in the panic log
//
// Example usage:
//
//	func handleRequest(logger logze.Logger, metrics *Metrics) {
//	    defer logger.RecoverWithCallback(func(p interface{}) {
//	        metrics.IncrementPanicCounter()
//	        alerts.SendPanicAlert(p)
//	    }, "request_id", reqID)
//
//	    // ... code that might panic
//	}
//
//	func mustNotPanic(logger logze.Logger) {
//	    defer logger.RecoverWithCallback(func(p interface{}) {
//	        // Re-throw after logging
//	        panic(p)
//	    }, "operation", "must-not-panic")
//
//	    // ... code that should not panic
//	}
//
// Example output:
//
//	{"level":"error","message":"... stack trace ...","error":"runtime error: nil pointer dereference","request_id":"abc123"}
func (l Logger) RecoverWithCallback(callback func(interface{}), fields ...interface{}) {
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

// InfoCtx logs a message at info level, checking context cancellation first.
//
// If the context is cancelled (Done channel is closed), this method returns
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
//	func handleRequest(ctx context.Context, logger logze.Logger) {
//	    logger.InfoCtx(ctx, "Processing request", "user_id", 123)
//	    // If request is cancelled, log won't be written
//	}
//
//	func processWithTimeout(ctx context.Context, logger logze.Logger) {
//	    ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
//	    defer cancel()
//
//	    time.Sleep(10 * time.Second) // Simulated long operation
//	    // This won't log because context timed out
//	    logger.InfoCtx(ctx, "Operation completed")
//	}
func (l Logger) InfoCtx(ctx context.Context, msg string, fields ...interface{}) {
	select {
	case <-ctx.Done():
		return // Context cancelled, skip logging
	default:
		l.Info(msg, fields...)
	}
}

// DebugCtx logs a message at debug level, checking context cancellation first.
// See InfoCtx for details about context checking behavior.
func (l Logger) DebugCtx(ctx context.Context, msg string, fields ...interface{}) {
	select {
	case <-ctx.Done():
		return
	default:
		l.Debug(msg, fields...)
	}
}

// TraceCtx logs a message at trace level, checking context cancellation first.
// See InfoCtx for details about context checking behavior.
func (l Logger) TraceCtx(ctx context.Context, msg string, fields ...interface{}) {
	select {
	case <-ctx.Done():
		return
	default:
		l.Trace(msg, fields...)
	}
}

// WarnCtx logs a message at warn level, checking context cancellation first.
// See InfoCtx for details about context checking behavior.
func (l Logger) WarnCtx(ctx context.Context, msg string, fields ...interface{}) {
	select {
	case <-ctx.Done():
		return
	default:
		l.Warn(msg, fields...)
	}
}

// ErrorCtx logs a message at error level, checking context cancellation first.
// See InfoCtx for details about context checking behavior.
func (l Logger) ErrorCtx(ctx context.Context, msg string, fields ...interface{}) {
	select {
	case <-ctx.Done():
		return
	default:
		l.Error(msg, fields...)
	}
}

// ErrCtx logs an error with message at error level, checking context cancellation first.
//
// This method combines error logging with context cancellation checking.
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
//	func processData(ctx context.Context, logger logze.Logger) error {
//	    result, err := fetchData()
//	    if err != nil {
//	        logger.ErrCtx(ctx, err, "Failed to fetch data", "retry_count", 3)
//	        return err
//	    }
//	    return nil
//	}
func (l Logger) ErrCtx(ctx context.Context, err error, msg string, fields ...interface{}) {
	select {
	case <-ctx.Done():
		return
	default:
		l.Err(err, msg, fields...)
	}
}
