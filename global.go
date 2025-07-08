package logze

import (
	stdlog "log"
	"time"

	"github.com/rs/zerolog"
)

var log = NewConsoleJSON()

// Default returns a copy of the global logger instance.
//
// The global logger is initialized with JSON output to stderr at info level.
// This function returns a copy, so modifications (like adding fields) won't
// affect the global instance.
//
// Use this when you want to add request-specific or component-specific fields
// to a logger without affecting other parts of your application.
//
// Example usage:
//
//	logger := logze.Default().With("component", "auth", "request_id", reqID)
//	logger.Info("User authenticated") // Includes component and request_id
func Default() Logger {
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
//
// Example usage:
//
//	ptr := logze.DefaultPtr()
//	// Be careful: changes affect global instance
func DefaultPtr() *Logger {
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
// Example usage:
//
//	logger := logze.New(logze.C(fileWriter).WithLevel("warn").WithSimpleErrorCounter())
//	logze.SetDefault(logger)
//
//	// Now all package-level logging uses the new configuration
//	logze.Info("This uses the new global logger")
func SetDefault(l Logger) {
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
// Example usage:
//
//	config := logze.C(logFile).WithLevel("info").WithSimpleErrorCounter()
//	logze.Init(config, "service", "api", "version", "2.1.0")
//
//	// Now all package functions and standard log calls use this configuration
//	logze.Info("Application started")
//	log.Println("This also goes through logze")
func Init(cfg Config, fields ...interface{}) {
	log = New(cfg, fields...)
	SetStdLogger(log)
}

// Update reconfigures the global logger with new settings and fields.
//
// This function updates the global logger in-place using the provided config,
// replacing its configuration, output writers, level, and other settings.
// It also reconfigures the standard library log package.
//
// ⚠️ THREAD SAFETY WARNING: This function is NOT safe for concurrent use.
// Ensure no other goroutines are using the global logger while calling Update.
//
// Example usage:
//
//	// Switch from development to production configuration
//	prodConfig := logze.C(prodFile).WithLevel("warn").WithNoDiode()
//	logze.Update(prodConfig, "environment", "production")
func Update(cfg Config, fields ...interface{}) {
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
// Example usage:
//
//	requestLogger := logze.WithFields("request_id", "abc123", "user_id", 456)
//	requestLogger.Info("Processing request") // Includes request_id and user_id
func WithFields(fields ...interface{}) Logger {
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
// using a global logger.
func Trace(msg string, fields ...interface{}) {
	log.log(log.l.Trace().Caller(1), msg, fields)
}

// Tracef logs a formatted message in trace level adding provided fields after formatting args
// and information about method caller using a global logger.
func Tracef(msg string, args ...interface{}) {
	log.logf(log.l.Trace().Caller(1), msg, args)
}

// TraceIf logs a message in trace level adding provided fields and information about method caller if condition is true.
func TraceIf(condition bool, msg string, fields ...interface{}) {
	log.TraceIf(condition, msg, fields...)
}

// Debug logs a message in debug level adding provided fields using a global logger.
func Debug(msg string, fields ...interface{}) {
	log.Debug(msg, fields...)
}

// Debugf logs a formatted message in debug level adding provided fields after formatting args using a global logger.
func Debugf(msg string, args ...interface{}) {
	log.Debugf(msg, args...)
}

// DebugIf logs a message in debug level adding provided fields if condition is true.
func DebugIf(condition bool, msg string, fields ...interface{}) {
	log.DebugIf(condition, msg, fields...)
}

// Info logs a message in info level adding provided fields using a global logger.
func Info(msg string, fields ...interface{}) {
	log.Info(msg, fields...)
}

// Infof logs a formatted message in info level adding provided fields after formatting args using a global logger.
func Infof(msg string, args ...interface{}) {
	log.Infof(msg, args...)
}

// InfoIf logs a message in info level adding provided fields if condition is true.
func InfoIf(condition bool, msg string, fields ...interface{}) {
	log.InfoIf(condition, msg, fields...)
}

// Warn logs a message in warning level adding provided fields using a global logger.
func Warn(msg string, fields ...interface{}) {
	log.Warn(msg, fields...)
}

// Warnf logs a formatted message in warn level adding provided fields after formatting args using a global logger.
func Warnf(msg string, args ...interface{}) {
	log.Warnf(msg, args...)
}

// WarnIf logs a message in warning level adding provided fields if condition is true.
func WarnIf(condition bool, msg string, fields ...interface{}) {
	log.WarnIf(condition, msg, fields...)
}

// Err logs a provided error in error level adding provided fields using a global logger.
func Err(err error, msg string, fields ...interface{}) {
	log.Err(err, msg, fields...)
}

// ErrIf logs a provided error in error level adding provided fields if condition is true.
func ErrIf(condition bool, err error, msg string, fields ...interface{}) {
	log.ErrIf(condition, err, msg, fields...)
}

// Error logs a message in error level adding provided fields using a global logger.
func Error(msg string, fields ...interface{}) {
	log.Error(msg, fields...)
}

// Errorf logs a formatted message in error level adding provided fields after formatting args using a global logger.
func Errorf(msg string, args ...interface{}) {
	log.Errorf(msg, args...)
}

// ErrorIf logs a message in error level adding provided fields if condition is true.
func ErrorIf(condition bool, msg string, fields ...interface{}) {
	log.ErrorIf(condition, msg, fields...)
}

// ErrStack logs a stack trace of provided error as message in error level adding fields.
func ErrStack(err error, fields ...interface{}) {
	log.ErrStack(err, fields...)
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
