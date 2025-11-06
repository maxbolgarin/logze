package logze

import (
	"io"
	"os"
	"regexp"
	"sync/atomic"
	"time"

	"github.com/rs/zerolog"
	"gopkg.in/natefinch/lumberjack.v2"
)

// DefaultDiodeSize is a default size of a diode writer. Logs will be lost if there will be more logs than that value
// in a small period of time (of time less that Config.DiodePollingInterval).
const (
	DefaultDiodeSize            = 1000
	DefaultDiodePollingInterval = 10 * time.Millisecond
	DefaultCallerSkipFrameCount = 5
)

// Enumerating string representations of all supported levels.
const (
	LevelTrace    = "trace"
	LevelDebug    = "debug"
	LevelInfo     = "info"
	LevelWarn     = "warn"
	LevelError    = "error"
	LevelFatal    = "fatal"
	LevelDisabled = "disabled"
)

// Levels is a list of all supported levels in string format.
var Levels = []string{
	LevelTrace, LevelDebug, LevelInfo, LevelWarn, LevelError, LevelFatal, LevelDisabled,
}

// LevelsAny is a list of all supported levels in interface{} format.
var LevelsAny = []interface{}{
	LevelTrace, LevelDebug, LevelInfo, LevelWarn, LevelError, LevelFatal, LevelDisabled,
}

// Config defines the configuration options for creating a [Logger] instance.
//
// Config uses the builder pattern - create an instance with [NewConfig] or [C], then chain
// With* methods to configure specific options. Direct struct initialization is not recommended
// as it bypasses default value handling and validation.
//
// The configuration covers all aspects of logging behavior including output destinations,
// formatting, sampling, error handling, performance optimizations, and more.
//
// Example usage:
//
//	config := logze.NewConfig().
//		WithConsoleJSON().
//		WithLevel("info").
//		WithSimpleErrorCounter().
//		WithAddCaller()
//	logger := logze.New(config, "service", "api")
//
// See individual With* methods for detailed configuration options.
type Config struct {
	// Writers is a list of writers where logger will log its data.
	// Default value is [io.Discard].
	Writers []io.Writer

	// Level is a log level in string format. Supported levels are:
	// trace, debug, info, warn, error, fatal, disabled.
	Level string

	// TimeFieldFormat is a format for time field. Default value is RFC3339.
	// You can use values from zerolog like [zerolog.TimeFormatUnix], [zerolog.TimeFormatUnixMs],
	// [zerolog.TimeFormatUnixMicro], [zerolog.TimeFormatUnixNano], [time.RFC3339], [time.RFC3339Nano] or custom.
	// UNIX Time is faster and smaller than most timestamps
	TimeFieldFormat string

	// ConsoleTimeFormat is the time format for console output (ConsoleWriter).
	// This only affects console output, not JSON output.
	// Default value is "2006-01-02 15:04:05".
	// Common formats: time.RFC3339, time.Kitchen, "15:04:05", etc.
	ConsoleTimeFormat string

	// Hook is a [zerolog.Hook] that will be used when creating [Logger].
	// Default value is nil.
	Hook zerolog.Hook

	// Hooks is a list of [zerolog.Hook] that will be used when creating [Logger].
	// Default value is nil.
	Hooks []zerolog.Hook

	// ToIgnore is a list of messages that will be ignored.
	// Default value is nil.
	ToIgnore []string

	// ToIgnoreRegex is a list of regular expression patterns for messages that will be ignored.
	// Messages matching any of these patterns will not be logged.
	// Default value is nil.
	ToIgnoreRegex []*regexp.Regexp

	// ErrorCounter is a counter of logged errors. Use WithSimpleErrorCounter method to use a simple counter.
	// Default value is nil.
	ErrorCounter ErrorCounter

	// DiodeSize is a size of a diode writer. Logs will be lost if there will be more logs than that value
	// in a small period of time (of time less that Config.DiodePollingInterval).
	// Default value is 1000.
	DiodeSize int

	// DiodePollingInterval is a time after which diode writer will flush its buffer.
	// Default value is 10ms.
	DiodePollingInterval time.Duration

	// DiodeAlertFunc is a function that will be called when diode writer will flush its buffer.
	// Default value is a function that writes a message in stderr.
	DiodeAlertFunc func(int)

	// UseDiodeWaiter if true, will enable diode waiter istead of poller.
	// Default value is false.
	UseDiodeWaiter bool

	// NoDiode if true, will disable diode writer.
	// Default value is false.
	NoDiode bool

	// StackTrace if true, will enable stack trace for Error and Errorf methods.
	// Default value is false.
	StackTrace bool

	// AddCaller if true, will add caller information to the log.
	// Default value is false.
	AddCaller bool

	// CallerSkipFrameCount is a number of frames to skip to get the caller information.
	// Default value is 5.
	CallerSkipFrameCount int

	// Sample is a [zerolog.LevelSampler] that will be used when creating [Logger].
	// Default value is nil.
	Sampler zerolog.Sampler
}

// NewConfig creates a new configuration instance with the specified output writers.
//
// This is the preferred way to create a Config instance. The writers specify where
// log output will be sent - you can provide multiple writers to send logs to
// multiple destinations simultaneously.
//
// If no writers are provided, logging will be disabled (outputs to io.Discard).
// Use the With* methods to configure additional options like log level, formatting,
// hooks, sampling, and performance settings.
//
// Example usage:
//
//	// Single writer (console JSON)
//	config := NewConfig(os.Stderr)
//
//	// Multiple writers (file + console)
//	logFile, _ := os.OpenFile("app.log", os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0644)
//	config := NewConfig(logFile, os.Stderr)
//
//	// No writers (disabled logging)
//	config := NewConfig()
func NewConfig(writers ...io.Writer) Config {
	return Config{
		Writers: writers,
	}
}

// C is a convenient shorthand for [NewConfig].
//
// This function provides the same functionality as NewConfig but with a shorter name
// for more concise configuration chains.
//
// Example usage:
//
//	logger := logze.New(logze.C(os.Stderr).WithLevel("debug").WithAddCaller())
func C(writers ...io.Writer) Config {
	return NewConfig(writers...)
}

// New creates a logger instance using this configuration with optional default fields.
//
// This is a convenience method equivalent to calling logze.New(config, fields...).
// It allows for more fluent configuration chains when you want to immediately
// create a logger from a config.
//
// Fields should be provided as alternating key-value pairs and will be included
// in every log message produced by the created logger.
//
// Example usage:
//
//	logger := logze.C(os.Stderr).
//		WithLevel("info").
//		WithAddCaller().
//		New("service", "api", "version", "2.1.0")
func (c Config) New(fields ...interface{}) Logger {
	return New(c, fields...)
}

// Logger is an alias for [Config.New] that creates a logger with optional fields.
//
// This method provides the same functionality as New but with a more explicit name
// that clearly indicates a Logger instance will be created.
//
// Example usage:
//
//	logger := config.Logger("component", "database", "driver", "postgres")
func (c Config) Logger(fields ...interface{}) Logger {
	return c.New(fields...)
}

// WithLevel configures the minimum log level for the logger.
//
// Valid levels are: "trace", "debug", "info", "warn", "error", "fatal", "disabled".
// Messages below the specified level will be discarded for performance.
// Level hierarchy: trace < debug < info < warn < error < fatal
//
// Example usage:
//
//	config := logze.C(os.Stderr).WithLevel("warn")
//	// Only warn, error, and fatal messages will be logged
func (c Config) WithLevel(level string) Config {
	c.Level = level
	return c
}

// WithTrace configures the logger to log all messages (trace level and above).
//
// Trace is the most verbose level, typically used for detailed debugging
// information. Use sparingly in production due to performance impact.
//
// Example usage:
//
//	config := logze.C(os.Stderr).WithTrace() // Logs everything
func (c Config) WithTrace() Config {
	c.Level = LevelTrace
	return c
}

// WithDebug configures the logger to log debug level and above messages.
//
// Debug level is useful for development and troubleshooting but typically
// disabled in production for performance reasons.
//
// Example usage:
//
//	config := logze.C(os.Stderr).WithDebug() // Logs debug, info, warn, error, fatal
func (c Config) WithDebug() Config {
	c.Level = LevelDebug
	return c
}

// WithInfo configures the logger to log info level and above messages.
//
// Info is the default and most common production log level, capturing
// important application events without excessive verbosity.
//
// Example usage:
//
//	config := logze.C(os.Stderr).WithInfo() // Logs info, warn, error, fatal
func (c Config) WithInfo() Config {
	c.Level = LevelInfo
	return c
}

// WithWarn configures the logger to log warning level and above messages.
//
// Warning level captures potentially problematic situations that don't
// prevent operation but should be investigated.
//
// Example usage:
//
//	config := logze.C(os.Stderr).WithWarn() // Logs warn, error, fatal only
func (c Config) WithWarn() Config {
	c.Level = LevelWarn
	return c
}

// WithError configures the logger to log error level and above messages.
//
// Error level captures actual errors and fatal conditions only.
// Use this for very quiet logging focused on problems.
//
// Example usage:
//
//	config := logze.C(os.Stderr).WithError() // Logs error, fatal only
func (c Config) WithError() Config {
	c.Level = LevelError
	return c
}

// WithFatal configures the logger to log only fatal level messages.
//
// Fatal level captures only the most critical errors that cause
// application termination. This creates very minimal logging.
//
// Example usage:
//
//	config := logze.C(os.Stderr).WithFatal() // Logs fatal only
func (c Config) WithFatal() Config {
	c.Level = LevelFatal
	return c
}

// WithDisabled completely disables logging.
//
// No log messages will be processed or output regardless of level.
// Useful for testing or when logging needs to be completely turned off.
//
// Example usage:
//
//	config := logze.C().WithDisabled() // No logging at all
func (c Config) WithDisabled() Config {
	c.Level = LevelDisabled
	return c
}

// WithHook returns [Config] with initialized [zerolog.Hook] provided as argument.
func (c Config) WithHook(hook zerolog.Hook) Config {
	c.Hook = hook
	return c
}

// WithHooks returns [Config] with initialized list of [zerolog.Hook] provided as argument.
func (c Config) WithHooks(hooks ...zerolog.Hook) Config {
	c.Hooks = hooks
	return c
}

// WithWriter adds an additional output writer to the logger configuration.
//
// This method appends the writer to the existing list of writers, allowing
// logs to be sent to multiple destinations simultaneously. Each log message
// will be written to all configured writers.
//
// Example usage:
//
//	var buffer bytes.Buffer
//	config := logze.C(os.Stderr).WithWriter(&buffer)
//	// Logs will go to both stderr and the buffer
func (c Config) WithWriter(w io.Writer) Config {
	c.Writers = append(c.Writers, w)
	return c
}

// WithFile configures logging output to a file with optional permissions.
//
// This method opens the specified file for append operations, creating it if
// it doesn't exist. The file is added to the list of output writers.
//
// Returns the updated config, a closer for the file (you should call Close()
// when done), and interface{} error that occurred during file opening.
//
// Default file permissions are 0644 if not specified.
//
// Example usage:
//
//	config, closer, err := logze.C().WithFile("app.log", 0644)
//	if err != nil {
//		log.Fatal(err)
//	}
//	defer closer.Close()
//
//	logger := config.New("service", "api")
func (c Config) WithFile(filename string, perm ...os.FileMode) (Config, io.Closer, error) {
	if len(perm) == 0 {
		perm = []os.FileMode{0644}
	}
	f, err := os.OpenFile(filename, os.O_CREATE|os.O_WRONLY|os.O_APPEND, perm[0])
	if err != nil {
		return c, nil, err
	}
	c.Writers = append(c.Writers, f)
	return c, f, nil
}

// WithRotatingFile configures logging output to a file with automatic rotation.
//
// This method uses the lumberjack library to provide log rotation based on file size,
// age, and number of backups. Rotated files are automatically compressed.
//
// Parameters:
//   - filename: Path to the log file
//   - maxSizeMB: Maximum size in megabytes before rotating (default: 100MB)
//   - maxAgeDays: Maximum age in days to retain old log files (0 = keep all)
//   - maxBackups: Maximum number of old log files to retain (0 = keep all)
//
// Features:
//   - Automatic rotation when file reaches maxSize
//   - Old files are compressed (.gz)
//   - Old files follow naming: filename.YYYY-MM-DD.HH-MM-SS.gz
//   - Thread-safe for concurrent writes
//
// Example usage:
//
//	// Rotate at 100MB, keep 30 days, max 5 backups
//	logger := logze.New(
//	    logze.C().
//	        WithRotatingFile("app.log", 100, 30, 5).
//	        WithLevel("info"),
//	    "service", "api",
//	)
//
//	// Rotate at 50MB, keep all files (no age limit), max 10 backups
//	logger := logze.New(
//	    logze.C().
//	        WithRotatingFile("app.log", 50, 0, 10).
//	        WithConsoleJSON(), // Can combine with console output
//	    "service", "api",
//	)
//
// Note: Requires the gopkg.in/natefinch/lumberjack.v2 package.
// The rotating writer implements io.WriteCloser and should be closed when done.
func (c Config) WithRotatingFile(filename string, maxSizeMB, maxAgeDays, maxBackups int) (Config, io.WriteCloser) {
	logger := &lumberjack.Logger{
		Filename:   filename,
		MaxSize:    maxSizeMB,
		MaxAge:     maxAgeDays,
		MaxBackups: maxBackups,
		Compress:   true,
	}
	c.Writers = append(c.Writers, logger)
	return c, logger
}

// WithConsole configures colored console output for human-readable logging.
//
// This outputs logs to stderr in a pretty-printed format with colors and
// human-readable timestamps. While great for development, this format is
// significantly slower than JSON and should generally be avoided in production.
//
// Example output: "2:04PM INF User authenticated user_id=123"
//
// Example usage:
//
//	config := logze.C().WithConsole() // For development/debugging
func (c Config) WithConsole() Config {
	return c.WithWriter(getConsoleWriter(os.Stderr, true, c.ConsoleTimeFormat))
}

// WithConsoleNoColor configures uncolored console output for human-readable logging.
//
// Similar to WithConsole but without ANSI color codes, suitable for environments
// that don't support colored output or when colors are undesirable.
//
// ⚠️ Performance Warning: Console format is significantly slower than JSON.
//
// Example usage:
//
//	config := logze.C().WithConsoleNoColor() // For CI/CD or simple terminals
func (c Config) WithConsoleNoColor() Config {
	return c.WithWriter(getConsoleWriter(os.Stderr, false, c.ConsoleTimeFormat))
}

// WithConsoleJSON configures structured JSON output to stderr.
//
// This is the recommended output format for production as it's fast, parseable,
// and integrates well with log aggregation systems. Each log entry is a single
// line of JSON.
//
// Example output: {"level":"info","time":"2023-11-20T18:48:14+03:00","message":"User authenticated","user_id":123}
//
// Example usage:
//
//	config := logze.C().WithConsoleJSON() // Recommended for production
func (c Config) WithConsoleJSON() Config {
	return c.WithWriter(os.Stderr)
}

// WithToIgnore returns [Config] with a list of messages that will be ignored.
func (c Config) WithToIgnore(toIgnore ...string) Config {
	c.ToIgnore = toIgnore
	return c
}

// WithToIgnoreRegex returns [Config] with a list of regular expression patterns for filtering messages.
// Messages matching any of these patterns will not be logged. This provides more flexible filtering
// than WithToIgnore, supporting pattern matching instead of exact/substring matches.
//
// The patterns are compiled into [*regexp.Regexp] objects. If a pattern fails to compile,
// this method will panic - ensure your patterns are valid regular expressions.
//
// Parameters:
//   - patterns: Regular expression patterns in string format
//
// Example usage:
//
//	config := logze.C().
//		WithConsoleJSON().
//		WithToIgnore("health", "ping").           // Exact/substring match
//		WithToIgnoreRegex(`^GET /metrics.*`,      // Regex: starts with "GET /metrics"
//		                  `(?i)debug`,            // Regex: case-insensitive "debug"
//		                  `\[test-\d+\]`)         // Regex: [test-123] format
//
// Note: Regex matching is more powerful but slightly slower than exact string matching.
// Use WithToIgnore for simple cases and WithToIgnoreRegex when pattern matching is needed.
func (c Config) WithToIgnoreRegex(patterns ...string) Config {
	c.ToIgnoreRegex = make([]*regexp.Regexp, len(patterns))
	for i, pattern := range patterns {
		c.ToIgnoreRegex[i] = regexp.MustCompile(pattern)
	}
	return c
}

// WithTimeFieldFormat returns [Config] with a new format for time field.
// TimeFieldFormat is a format for time field. Default value is RFC3339.
// You can use values from zerolog like [zerolog.TimeFormatUnix], [zerolog.TimeFormatUnixMs],
// [zerolog.TimeFormatUnixMicro], [zerolog.TimeFormatUnixNano], [time.RFC3339], [time.RFC3339Nano] or custom.
// UNIX Time is faster and smaller than most timestamps
func (c Config) WithTimeFieldFormat(format string) Config {
	c.TimeFieldFormat = format
	return c
}

// WithConsoleTimeFormat returns [Config] with a custom time format for console output.
// This only affects the ConsoleWriter output format, not JSON output.
//
// The format string uses Go's time formatting layout (e.g., "2006-01-02 15:04:05").
// Common formats:
//   - "2006-01-02 15:04:05" (default)
//   - time.RFC3339 ("2006-01-02T15:04:05Z07:00")
//   - time.Kitchen ("3:04PM")
//   - "15:04:05" (time only)
//   - "Jan 02 15:04:05" (syslog style)
//
// Example usage:
//
//	config := logze.C().
//		WithConsole().
//		WithConsoleTimeFormat(time.Kitchen)  // Shows "3:04PM"
//
//	config := logze.C().
//		WithConsole().
//		WithConsoleTimeFormat("15:04:05")    // Shows "14:30:45"
func (c Config) WithConsoleTimeFormat(format string) Config {
	c.ConsoleTimeFormat = format
	return c
}

// WithDiodeSize returns [Config] with a new size of diode writer.
// If there will be more logs than [Config.DiodeSize] in a period of time less that [Config.DiodePollingInterval],
// then diode writer won't accept new logs.
func (c Config) WithDiodeSize(size int) Config {
	c.DiodeSize = size
	return c
}

// WithDiodePollingInterval returns [Config] with enabled diode polling with provided interval.
// Logs will be flushed to a writer every [Config.DiodePollingInterval].
// Default value is 10ms.
func (c Config) WithDiodePollingInterval(interval time.Duration) Config {
	c.DiodePollingInterval = interval
	return c
}

// WithDiodeAlert returns [Config] with provided diode alert func.
func (c Config) WithDiodeAlert(foo func(int)) Config {
	c.DiodeAlertFunc = foo
	return c
}

// WithNoDiode returns [Config] with disabled diode writer.
func (c Config) WithNoDiode() Config {
	c.NoDiode = true
	return c
}

// WithDiodeWaiter returns [Config] with enabled diode waiter.
func (c Config) WithDiodeWaiter() Config {
	c.UseDiodeWaiter = true
	return c
}

// WithStackTrace returns [Config] with an enabled stack trace for Error and Errorf methods.
func (c Config) WithStackTrace() Config {
	c.StackTrace = true
	return c
}

// WithErrorCounter returns [Config] with the provided [ErrorCounter].
func (c Config) WithErrorCounter(ec ErrorCounter) Config {
	c.ErrorCounter = ec
	return c
}

// WithErrorCounter returns [Config] with a simple [ErrorCounter].
func (c Config) WithSimpleErrorCounter() Config {
	c.ErrorCounter = newSimpleErrorCounter()
	return c
}

// WithAddCaller returns [Config] with an enabled caller information.
func (c Config) WithAddCaller() Config {
	c.AddCaller = true
	c.CallerSkipFrameCount = DefaultCallerSkipFrameCount
	return c
}

// WithCallerSkipFrameCount returns [Config] with a new caller skip frame count.
func (c Config) WithCallerSkipFrameCount(count int) Config {
	c.AddCaller = true
	c.CallerSkipFrameCount = count
	return c
}

// WithSampler returns [Config] with a new [zerolog.Sampler].
func (c Config) WithSampler(sampler zerolog.Sampler) Config {
	c.Sampler = sampler
	return c
}

// WithPercentageSampler returns [Config] with a new percentage sampler.
// Percentage is a float64 percentage of logs that will be sampled from 0 to 1.
// Levels is an optional list of levels that will be sampled. If no levels are provided,
// the sampler will be used for all levels.
//
// Example usage:
//
//	config := logze.C().WithPercentageSampler(0.1, "debug", "info")
//	logger := config.New()
//	logger.Debug("This will be logged 10% of the time")
//	logger.Info("This will be logged 10% of the time")
func (c Config) WithPercentageSampler(percentage float64, levels ...string) Config {
	sampler := percentageSampler(percentage)
	c.Sampler = getLevelSampler(sampler, levels...)
	return c
}

// WithBurstSampler returns [Config] with a new burst sampler.
// Percentage is a float64 percentage of logs that will be sampled from 0 to 1.
// Burst is the maximum number of event per period allowed before percentage sampling.
// Period is a time interval after which the percentage sampler will be called again.
// Levels is an optional list of levels that will be sampled. If no levels are provided,
// the sampler will be used for all levels.
//
// Example usage:
//
//	config := logze.C().WithBurstSampler(0.1, 100, 1*time.Second, "debug", "info")
//	logger := config.New()
//	logger.Debug("This will be logged 10% of the time")
//	logger.Info("This will be logged 10% of the time")
func (c Config) WithBurstSampler(percentage float64, burst int, period time.Duration, levels ...string) Config {
	sampler := burstSampler(percentage, burst, period)
	c.Sampler = getLevelSampler(sampler, levels...)
	return c
}

// WithMaxSampler returns [Config] with a new max sampler.
// Max is the maximum number of requests allowed per period.
// Period is a time interval after which the max count resets.
// interface{} requests beyond the max limit will be dropped.
// Levels is an optional list of levels that will be sampled. If no levels are provided,
// the sampler will be used for all levels.
//
// Example usage:
//
//	config := logze.C().WithMaxSampler(100, 1*time.Second, "debug", "info")
//	logger := config.New()
//	logger.Debug("This will be logged 10% of the time")
//	logger.Info("This will be logged 10% of the time")
func (c Config) WithMaxSampler(max int, period time.Duration, levels ...string) Config {
	sampler := burstSampler(0, max, period)
	c.Sampler = getLevelSampler(sampler, levels...)
	return c
}

func percentageSampler(percentage float64) zerolog.Sampler {
	if percentage < 0 {
		return zerolog.RandomSampler(0)
	}
	if percentage > 1 {
		return zerolog.RandomSampler(1)
	}
	return zerolog.RandomSampler(float64(1) / percentage)
}

func burstSampler(percentage float64, burst int, period time.Duration) zerolog.Sampler {
	if burst < 0 {
		burst = 0
	}
	if period <= 0 {
		period = 1 * time.Second
	}
	return &zerolog.BurstSampler{
		Burst:       uint32(burst),
		Period:      period,
		NextSampler: percentageSampler(percentage),
	}
}

func getLevelSampler(sampler zerolog.Sampler, levels ...string) zerolog.Sampler {
	if len(levels) == 0 {
		return sampler
	}
	resultSampler := &zerolog.LevelSampler{}
	for _, level := range levels {
		switch level {
		case LevelTrace:
			resultSampler.TraceSampler = sampler
		case LevelDebug:
			resultSampler.DebugSampler = sampler
		case LevelInfo:
			resultSampler.InfoSampler = sampler
		case LevelWarn:
			resultSampler.WarnSampler = sampler
		case LevelError:
			resultSampler.ErrorSampler = sampler
		}
	}
	return resultSampler
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

// ErrorCounter provides an interface for tracking the number of errors logged.
//
// Implementations of ErrorCounter are called automatically whenever error-level
// logging occurs (Err, Error, Fatal, Panic methods). This enables error tracking
// for monitoring, alerting, or debugging purposes.
//
// The interface is intentionally simple to allow for various implementations:
// simple counters, metrics systems, alerting systems, etc.
//
// Example custom implementation:
//
//	type MetricsErrorCounter struct {
//		metric prometheus.Counter
//	}
//
//	func (m *MetricsErrorCounter) Inc(err error) {
//		m.metric.Inc()
//		// Could also categorize by error type, etc.
//	}
//
// Use [Config.WithSimpleErrorCounter] for a basic atomic counter or
// [Config.WithErrorCounter] for custom implementations.
type ErrorCounter interface {
	Inc(err error)
}

// SimpleErrorCounter is a thread-safe error counter using atomic operations.
//
// This implementation provides a basic error counting mechanism suitable for
// most use cases. The Count field can be read directly to get the current
// error count, and all operations are atomic for safe concurrent use.
//
// Example usage:
//
//	logger := logze.NewConsoleJSON().WithSimpleErrorCounter()
//	// ... application code that logs errors ...
//
//	if counter := logger.GetErrorCounter(); counter != nil {
//		simple := counter.(*logze.SimpleErrorCounter)
//		errorCount := simple.Count.Load()
//		if errorCount > threshold {
//			// Take action based on error count
//		}
//	}
type SimpleErrorCounter struct {
	Count uint64
}

// Inc increments the error counter by 1.
//
// This method is called automatically by the logger whenever error-level
// logging occurs. The increment operation is atomic and safe for concurrent use.
// The error parameter is currently unused but provided for interface compatibility
// and potential future enhancements.
func (c *SimpleErrorCounter) Inc(err error) {
	if err == nil {
		return
	}
	atomic.AddUint64(&c.Count, 1)
}

func newSimpleErrorCounter() *SimpleErrorCounter {
	return &SimpleErrorCounter{}
}
