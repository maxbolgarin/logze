package logze

import (
	stdlog "log"

	"github.com/rs/zerolog"
)

// Log is a global logger.
var Log = NewDefault()

// Init calls [New] function and assigns the result to global [Log] variable.
// It also calls [SetLoggerForDefault] with this new logger.
func Init(cfg Config, fields ...any) {
	Log = New(cfg, fields...)
	SetLoggerForDefault(Log)
}

// Update calls [Logger.Update] method for global [Log].
// It also calls [SetLoggerForDefault] with this new logger.
// It is not safe for concurrent use!
func Update(cfg Config, fields ...any) {
	Log.Update(cfg, fields...)
	SetLoggerForDefault(Log)
}

// SetLoggerForDefault sets priovded [Logger] with (key, value) pairs as writer for default Go logger and also
// calls stdlog.SetFlags(0).
func SetLoggerForDefault(l Logger, fields ...any) {
	stdlog.SetFlags(0)
	stdlog.SetOutput(l.WithFields(fields...))
}

// WithFields returns [Logger] with applied fields, provided as (key, value) pairs, based on a global logger.
func WithFields(fields ...any) Logger {
	return Log.WithFields(fields...)
}

// With is a shortcut for [WithFields].
func With(fields ...any) Logger {
	return Log.With(fields...)
}

// WithLevel returns [Logger] with applied log level, based on a global logger.
func WithLevel(level string) Logger {
	return Log.WithLevel(level)
}

// WithErrorCounter returns [Logger] with the provided [ErrorCounter], based on a global logger.
func WithErrorCounter(ec ErrorCounter) Logger {
	return Log.WithErrorCounter(ec)
}

// WithSimpleErrorCounter returns [Logger] with a simple [ErrorCounter] inited with the provided name,
// based on a global logger.
func WithSimpleErrorCounter(name string) Logger {
	return Log.WithSimpleErrorCounter(name)
}

// Trace logs a message in trace level adding provided fields and information about method caller
// using a global logger.
func Trace(msg string, fields ...any) {
	Log.log(Log.l.Trace().Caller(1), msg, fields)
}

// Tracef logs a formatted message in trace level adding provided fields after formatting args
// and information about method caller using a global logger.
func Tracef(msg string, args ...any) {
	Log.logf(Log.l.Trace().Caller(1), msg, args)
}

// Debug logs a message in debug level adding provided fields using a global logger.
func Debug(msg string, fields ...any) {
	Log.Debug(msg, fields...)
}

// Debugf logs a formatted message in debug level adding provided fields after formatting args using a global logger.
func Debugf(msg string, args ...any) {
	Log.Debugf(msg, args...)
}

// Info logs a message in info level adding provided fields using a global logger.
func Info(msg string, fields ...any) {
	Log.Info(msg, fields...)
}

// Infof logs a formatted message in info level adding provided fields after formatting args using a global logger.
func Infof(msg string, args ...any) {
	Log.Infof(msg, args...)
}

// Warn logs a message in warning level adding provided fields using a global logger.
func Warn(msg string, fields ...any) {
	Log.Warn(msg, fields...)
}

// Warnf logs a formatted message in warn level adding provided fields after formatting args using a global logger.
func Warnf(msg string, args ...any) {
	Log.Warnf(msg, args...)
}

// Err logs a provided error in error level adding provided fields using a global logger.
func Err(err error, fields ...any) {
	Log.Err(err, fields...)
}

// Error logs a provided error and message in error level adding provided fields using a global logger.
func Error(err error, msg string, fields ...any) {
	Log.Error(err, msg, fields...)
}

// Errorf logs a provided error and formatted message in error level adding provided fields after formatting args
// using a global logger.
func Errorf(err error, msg string, args ...any) {
	Log.Errorf(err, msg, args...)
}

// Fatal logs a message in fatal level using fmt.Sprint to interpret args sing a global logger, then calls os.Exit(1).
func Fatal(v ...any) {
	Log.Fatal(v...)
}

// Fatalf logs a formatted message in fatal level using a global logger, then calls os.Exit(1).
func Fatalf(format string, args ...any) {
	Log.Fatalf(format, args...)
}

// Fatalln logs a message in fatal level using fmt.Sprintln to interpret args using a global logger, then calls os.Exit(1).
func Fatalln(v ...any) {
	Log.Fatalln(v...)
}

// Panic logs a message in fatal level using fmt.Sprint to interpret args using a global logger, then calls panic().
func Panic(v ...any) {
	Log.Panic(v...)
}

// Panicf logs a formatted message in fatal level using a global logger, then calls panic().
func Panicf(format string, args ...any) {
	Log.Panicf(format, args...)
}

// Panicln logs a message in fatal level using fmt.Sprintln to interpret args using a global logger, then calls panic().
func Panicln(v ...any) {
	Log.Panicln(v...)
}

// Print logs a message without level using fmt.Sprint to interpret args using a global logger.
func Print(v ...any) {
	Log.Print(v...)
}

// Printf logs a formatted message without level using a global logger.
func Printf(format string, args ...any) {
	Log.Printf(format, args...)
}

// Println writes a message without level using fmt.Sprintln to interpret args using a global logger.
func Println(v ...any) {
	Log.Println(v...)
}

// Write writes bytes to underlying [io.Writer] using a global logger.
func Write(p []byte) (n int, err error) {
	return Log.Write(p)
}

// Raw returns Logger's underlying [zerolog.Logger] from global logger.
func Raw() *zerolog.Logger {
	return Log.Raw()
}

// GetErrorCounter returns Logger's underlying [ErrorCounter] from global logger.
func GetErrorCounter() ErrorCounter {
	return Log.GetErrorCounter()
}
