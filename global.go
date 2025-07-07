package logze

import (
	stdlog "log"

	"github.com/rs/zerolog"
)

var log = NewConsoleJSON()

// Default returns a copy on a global logger.
func Default() Logger {
	return log
}

// D is a shortcut for [Default].
func D() Logger {
	return log
}

// DefaultPtr returns a pointer to a global logger.
func DefaultPtr() *Logger {
	return &log
}

// DP is a shortcut for [DefaultPtr].
func DP() *Logger {
	return &log
}

// SetDefault sets provided [Logger] as a global logger.
func SetDefault(l Logger) {
	log = l
}

// Init calls [New] function and assigns the result to global [log] variable.
// It also calls [SetLoggerForDefault] with this new logger.
func Init(cfg Config, fields ...any) {
	log = New(cfg, fields...)
	SetStdLogger(log)
}

// Update calls [Logger.Update] method for global [log].
// It also calls [SetLoggerForDefault] with this new logger.
// It is not safe for concurrent use!
func Update(cfg Config, fields ...any) {
	log.Update(cfg, fields...)
	SetStdLogger(log)
}

// SetLoggerForDefault sets priovded [Logger] with (key, value) pairs as writer for default Go logger and also
// calls stdlog.SetFlags(0).
func SetStdLogger(l Logger, fields ...any) {
	stdlog.SetFlags(0)
	stdlog.SetOutput(l.WithFields(fields...))
	log = l
}

// WithFields returns [Logger] with applied fields, provided as (key, value) pairs, based on a global logger.
func WithFields(fields ...any) Logger {
	return log.WithFields(fields...)
}

// With is a shortcut for [WithFields].
func With(fields ...any) Logger {
	return log.With(fields...)
}

// WithLevel returns [Logger] with applied log level, based on a global logger.
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
	log.toIgnore = toIgnore
	return log
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

// Trace logs a message in trace level adding provided fields and information about method caller
// using a global logger.
func Trace(msg string, fields ...any) {
	log.log(log.l.Trace().Caller(1), msg, fields)
}

// Tracef logs a formatted message in trace level adding provided fields after formatting args
// and information about method caller using a global logger.
func Tracef(msg string, args ...any) {
	log.logf(log.l.Trace().Caller(1), msg, args)
}

// TraceIf logs a message in trace level adding provided fields and information about method caller if condition is true.
func TraceIf(condition bool, msg string, fields ...any) {
	log.TraceIf(condition, msg, fields...)
}

// Debug logs a message in debug level adding provided fields using a global logger.
func Debug(msg string, fields ...any) {
	log.Debug(msg, fields...)
}

// Debugf logs a formatted message in debug level adding provided fields after formatting args using a global logger.
func Debugf(msg string, args ...any) {
	log.Debugf(msg, args...)
}

// DebugIf logs a message in debug level adding provided fields if condition is true.
func DebugIf(condition bool, msg string, fields ...any) {
	log.DebugIf(condition, msg, fields...)
}

// Info logs a message in info level adding provided fields using a global logger.
func Info(msg string, fields ...any) {
	log.Info(msg, fields...)
}

// Infof logs a formatted message in info level adding provided fields after formatting args using a global logger.
func Infof(msg string, args ...any) {
	log.Infof(msg, args...)
}

// InfoIf logs a message in info level adding provided fields if condition is true.
func InfoIf(condition bool, msg string, fields ...any) {
	log.InfoIf(condition, msg, fields...)
}

// Warn logs a message in warning level adding provided fields using a global logger.
func Warn(msg string, fields ...any) {
	log.Warn(msg, fields...)
}

// Warnf logs a formatted message in warn level adding provided fields after formatting args using a global logger.
func Warnf(msg string, args ...any) {
	log.Warnf(msg, args...)
}

// WarnIf logs a message in warning level adding provided fields if condition is true.
func WarnIf(condition bool, msg string, fields ...any) {
	log.WarnIf(condition, msg, fields...)
}

// Err logs a provided error in error level adding provided fields using a global logger.
func Err(err error, msg string, fields ...any) {
	log.Err(err, msg, fields...)
}

// ErrIf logs a provided error in error level adding provided fields if condition is true.
func ErrIf(condition bool, err error, msg string, fields ...any) {
	log.ErrIf(condition, err, msg, fields...)
}

// Error logs a message in error level adding provided fields using a global logger.
func Error(msg string, fields ...any) {
	log.Error(msg, fields...)
}

// Errorf logs a formatted message in error level adding provided fields after formatting args using a global logger.
func Errorf(msg string, args ...any) {
	log.Errorf(msg, args...)
}

// ErrorIf logs a message in error level adding provided fields if condition is true.
func ErrorIf(condition bool, msg string, fields ...any) {
	log.ErrorIf(condition, msg, fields...)
}

// ErrStack logs a stack trace of provided error as message in error level adding fields.
func ErrStack(err error, fields ...any) {
	log.ErrStack(err, fields...)
}

// FatalIf logs a message in fatal level adding provided fields if condition is true, then calls os.Exit(1).
func FatalIf(condition bool, v ...any) {
	log.FatalIf(condition, v...)
}

// Fatal logs a message in fatal level using fmt.Sprint to interpret args sing a global logger, then calls os.Exit(1).
func Fatal(v ...any) {
	log.Fatal(v...)
}

// Fatalf logs a formatted message in fatal level using a global logger, then calls os.Exit(1).
func Fatalf(format string, args ...any) {
	log.Fatalf(format, args...)
}

// Fatalln logs a message in fatal level using fmt.Sprintln to interpret args using a global logger, then calls os.Exit(1).
func Fatalln(v ...any) {
	log.Fatalln(v...)
}

// PanicIf logs a message in fatal level adding provided fields if condition is true, then calls panic().
func PanicIf(condition bool, v ...any) {
	log.PanicIf(condition, v...)
}

// Panic logs a message in fatal level using fmt.Sprint to interpret args using a global logger, then calls panic().
func Panic(v ...any) {
	log.Panic(v...)
}

// Panicf logs a formatted message in fatal level using a global logger, then calls panic().
func Panicf(format string, args ...any) {
	log.Panicf(format, args...)
}

// Panicln logs a message in fatal level using fmt.Sprintln to interpret args using a global logger, then calls panic().
func Panicln(v ...any) {
	log.Panicln(v...)
}

// Print logs a message without level using [fmt.Sprint] to interpret args using a global logger.
func Print(v ...any) {
	log.Print(v...)
}

// PrintIf logs a message without level using [fmt.Sprint] to interpret args if condition is true.
func PrintIf(condition bool, v ...any) {
	log.PrintIf(condition, v...)
}

// PrintStack logs a current stack trace.
func PrintStack(v ...any) {
	log.PrintStack(v...)
}

// Log logs a message without level using [fmt.Sprint] to interpret args using a global logger.
// It is an alias for [Print].
func Log(v ...any) {
	log.Log(v...)
}

// Printf logs a formatted message without level using a global logger.
func Printf(format string, args ...any) {
	log.Printf(format, args...)
}

// Println writes a message without level using fmt.Sprintln to interpret args using a global logger.
func Println(v ...any) {
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
