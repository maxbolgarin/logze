// Package logze implements a zerolog wrapper providing a convenient and short interface for structural logging.
package logze

import (
	"fmt"
	"io"
	"os"
	"strings"

	"github.com/maxbolgarin/errm"
	"github.com/pkg/errors"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/diode"
	"github.com/rs/zerolog/pkgerrors"
)

// Logger represents an initialized logger. Default value behaves as default [zerolog.Logger].
type Logger struct {
	l          zerolog.Logger
	errCounter ErrorCounter
	toIgnore   []string
	stackTrace bool
	inited     bool
}

// New returns a new [Logger] with provided config and fields.
//   - Default output is [io.Discard], so you should provide at least one [io.Writer] in [Config] when creating a logger.
//   - Default level is info.
//   - Fields should be passed as (key, value) pairs, its will be applied to all messages.
//
// For example, if you use [Logger] like that:
//
//	lg := New(NewConfig().WithConsoleJSON(), "foo", "bar")
//	lg.Info("some message", "key", "value")
//	lg.Error(errors.New("some error"), "cannot handle")
//
// You will have output:
//
//	{"level":"info","time":"2023-11-20T18:48:14+03:00","message":"some message","foo":"bar","key":"value"}
//	{"level":"error","time":"2023-11-20T18:48:14+03:00","error":"some error","message":"cannot handle","foo":"bar"}
func New(cfg Config, fields ...any) Logger {
	if len(cfg.Writers) == 0 || cfg.Level == DisabledLevel {
		cfg.Writers = []io.Writer{io.Discard}
	}
	if cfg.Level == "" {
		cfg.Level = InfoLevel
	}
	level, err := zerolog.ParseLevel(cfg.Level)
	if err != nil {
		panic("cannot parse level=" + cfg.Level)
	}

	output := cfg.Writers[0]
	if len(cfg.Writers) > 1 {
		output = zerolog.MultiLevelWriter(cfg.Writers...)
	}
	if !cfg.DisableDiode {
		if cfg.DiodeSize == 0 {
			cfg.DiodeSize = DefaultDiodeSize
		}
		if cfg.DiodeAlertFunc == nil {
			cfg.DiodeAlertFunc = func(missed int) {
				fmt.Fprintf(os.Stderr, "WRN: logger dropped %d messages\n", missed)
			}
		}
		// To fix problem of blocking goroutine when writing in Stderr
		// https://github.com/cloudfoundry/go-diodes
		output = diode.NewWriter(output, cfg.DiodeSize, cfg.DiodePollingInterval, cfg.DiodeAlertFunc)
	}

	l := zerolog.New(output).With().Timestamp().Fields(fields).Logger().Level(level)

	if cfg.Hook != nil {
		l = l.Hook(cfg.Hook)
	}

	if cfg.ErrorCounter == nil {
		cfg.ErrorCounter = noopErrorCounter{}
	}

	zerolog.ErrorStackMarshaler = pkgerrors.MarshalStack

	return Logger{
		l:          l,
		toIgnore:   cfg.ToIgnore,
		errCounter: cfg.ErrorCounter,
		stackTrace: cfg.StackTrace,
		inited:     true,
	}
}

// NewFromZerolog returns a new [Logger] based on provided [zerolog.Logger].
func NewFromZerolog(l zerolog.Logger) Logger {
	return Logger{
		l:          l,
		inited:     true,
		errCounter: noopErrorCounter{},
	}
}

// NewDefault returns a new [Logger] with logging to stderr.
func NewDefault(fields ...any) Logger {
	return New(NewConfig().WithConsoleJSON(), fields...)
}

// Nop returns a new [Logger] with no logging.
func Nop() Logger {
	return Logger{l: zerolog.Nop()}
}

// Update replaces underlying logger with a new one created using provided config and fields.
// It is not safe for concurrent use!
func (l *Logger) Update(cfg Config, fields ...any) {
	newLogger := New(cfg, fields...)
	l.l = newLogger.l
	l.inited = newLogger.inited
	l.errCounter = newLogger.errCounter
	l.stackTrace = newLogger.stackTrace
	l.toIgnore = newLogger.toIgnore
}

// NotInited returns true if [Logger] is not inited (struct with default values).
func (l Logger) NotInited() bool {
	return !l.inited
}

// WithFields returns [Logger] with applied fields to all messages, provided as (key, value) pairs.
func (l Logger) WithFields(fields ...any) Logger {
	l.l = l.l.With().Fields(fields).Logger()
	return l
}

// With is a shortcut for [Logger.WithFields].
func (l Logger) With(fields ...any) Logger {
	return l.WithFields(fields...)
}

// WithLevel returns [Logger] with an applied log level.
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

// WithStack returns [Logger] with an applied stackTrace.
func (l Logger) WithStack(stackTrace bool) Logger {
	l.stackTrace = stackTrace
	return l
}

// WithErrorCounter returns [Logger] with the provided [ErrorCounter].
func (l Logger) WithErrorCounter(ec ErrorCounter) Logger {
	l.errCounter = ec
	return l
}

// WithSimpleErrorCounter returns [Logger] with a simple [ErrorCounter] inited with the provided name.
func (l Logger) WithSimpleErrorCounter(name string) Logger {
	l.errCounter = newErrorCounter(name)
	return l
}

// Trace logs a message in trace level adding provided fields and information about method caller.
func (l Logger) Trace(msg string, fields ...any) {
	l.log(l.l.Trace().Caller(1), msg, fields)
}

// Tracef logs a formatted message in trace level adding provided fields after formatting args
// and information about method caller.
func (l Logger) Tracef(msg string, args ...any) {
	l.logf(l.l.Trace().Caller(1), msg, args)
}

// Debug logs a message in debug level adding provided fields.
func (l Logger) Debug(msg string, fields ...any) {
	l.log(l.l.Debug(), msg, fields)
}

// Debugf logs a formatted message in debug level adding provided fields after formatting args.
func (l Logger) Debugf(msg string, args ...any) {
	l.logf(l.l.Debug(), msg, args)
}

// Info logs a message in info level adding provided fields.
func (l Logger) Info(msg string, fields ...any) {
	l.log(l.l.Info(), msg, fields)
}

// Infof logs a formatted message in info level adding provided fields after formatting args.
func (l Logger) Infof(msg string, args ...any) {
	l.logf(l.l.Info(), msg, args)
}

// Warn logs a message in warning level adding provided fields.
func (l Logger) Warn(msg string, fields ...any) {
	l.log(l.l.Warn(), msg, fields)
}

// Warnf logs a formatted message in warn level adding provided fields after formatting args.
func (l Logger) Warnf(msg string, args ...any) {
	l.logf(l.l.Warn(), msg, args)
}

// Err logs a provided error in error level adding provided fields.
func (l Logger) Err(err error, fields ...any) {
	l.Error(err, "", fields...)
}

// Error logs a provided error and message in error level adding provided fields.
func (l Logger) Error(err error, msg string, fields ...any) {
	ev := l.setErrorWithStack(err, l.l.Error())
	l.log(ev, msg, fields)
}

// Errorf logs a provided error and formatted message in error level adding provided fields after formatting args.
func (l Logger) Errorf(err error, msg string, args ...any) {
	ev := l.setErrorWithStack(err, l.l.Error())
	l.logf(ev, msg, args)
}

// Fatal logs a message in fatal level using fmt.Sprint to interpret args, then calls os.Exit(1).
func (l Logger) Fatal(v ...any) {
	s := fmt.Sprint(v...)
	l.log(l.l.WithLevel(zerolog.FatalLevel), s, nil)
	l.incErrorConter(errors.New(s))
	os.Exit(1)
}

// Fatalf logs a formatted message in fatal level, then calls os.Exit(1).
func (l Logger) Fatalf(format string, args ...any) {
	l.log(l.l.WithLevel(zerolog.FatalLevel), format, args)
	l.incErrorConter(fmt.Errorf(format, args...))
	os.Exit(1)
}

// Fatalln logs a message in fatal level using fmt.Sprintln to interpret args, then calls os.Exit(1).
func (l Logger) Fatalln(v ...any) {
	s := fmt.Sprintln(v...)
	l.log(l.l.WithLevel(zerolog.FatalLevel), s, nil)
	l.incErrorConter(errors.New(s))
	os.Exit(1)
}

// Panic logs a message in fatal level using fmt.Sprint to interpret args, then calls panic().
func (l Logger) Panic(v ...any) {
	s := fmt.Sprint(v...)
	l.log(l.l.WithLevel(zerolog.FatalLevel), s, nil)
	l.incErrorConter(errors.New(s))
	panic(s)
}

// Panicf logs a formatted message in fatal level, then calls panic().
func (l Logger) Panicf(format string, args ...any) {
	l.log(l.l.WithLevel(zerolog.FatalLevel), format, args)
	l.incErrorConter(fmt.Errorf(format, args...))
	panic(fmt.Sprintf(format, args...))
}

// Panicln logs a message in fatal level using fmt.Sprintln to interpret args, then calls panic().
func (l Logger) Panicln(v ...any) {
	s := fmt.Sprintln(v...)
	l.log(l.l.WithLevel(zerolog.FatalLevel), s, nil)
	l.incErrorConter(errors.New(s))
	panic(s)
}

// Print logs a message without level using fmt.Sprint to interpret args.
func (l Logger) Print(v ...any) {
	if len(v) == 0 {
		return
	}
	l.log(l.l.Log(), fmt.Sprint(v...), nil)
}

// Printf logs a formatted message without level.
func (l Logger) Printf(format string, args ...any) {
	l.logf(l.l.Log(), format, args)
}

// Println writes a message without level using fmt.Sprintln to interpret args.
func (l Logger) Println(v ...any) {
	l.log(l.l.Log(), fmt.Sprintln(v...), nil)
}

// Write writes bytes to underlying [io.Writer].
func (l Logger) Write(p []byte) (n int, err error) {
	return l.l.Write(p)
}

// Raw returns Logger's underlying [zerolog.Logger].
func (l Logger) Raw() *zerolog.Logger {
	return &l.l
}

// GetErrorCounter returns Logger's underlying [ErrorCounter].
func (l Logger) GetErrorCounter() ErrorCounter {
	return l.errCounter
}

func (l Logger) log(ev *zerolog.Event, msg string, fields []any) {
	for _, ignore := range l.toIgnore {
		if strings.Contains(msg, ignore) {
			return
		}
	}
	if len(fields) > 1 {
		ev = ev.Fields(fields)
	}
	ev.Msg(msg)
}

func (l Logger) logf(ev *zerolog.Event, msg string, args []any) {
	for _, ignore := range l.toIgnore {
		if strings.Contains(msg, ignore) {
			return
		}
	}
	numberOfFormats := strings.Count(msg, "%")
	if numberOfFormats > 0 && numberOfFormats <= len(args) {
		ev = ev.Fields(args[numberOfFormats:])
		args = args[:numberOfFormats]
	}
	if numberOfFormats == 0 && len(args) > 0 {
		ev = ev.Fields(args)
		args = nil
	}
	if len(args) == 0 {
		ev.Msg(msg)
		return
	}
	ev.Msgf(msg, args...)
}

func (l Logger) setErrorWithStack(err error, ev *zerolog.Event) *zerolog.Event {
	if l.stackTrace {
		if errm.Check(err) {
			ev = ev.Fields(errm.StackForLogger(err))
		} else {
			ev = ev.Stack()
			err = errors.WithStack(err)
		}
	}
	l.incErrorConter(err)
	return ev.Err(err)
}

func (l Logger) incErrorConter(err error) {
	if l.errCounter != nil {
		l.errCounter.Inc(err)
	}
}

// SLogger is a wrapper for [Logger].
// It provides the same methods as [Logger] but with another Error interface (slog style).
type SLogger struct {
	Logger
}

// ConvertToS converts [Logger] to [SLogger].
func ConvertToS(l Logger) SLogger {
	return SLogger{Logger: l}
}

// S is a shortcut to [ConvertToS] that converts [Logger] to [SLogger].
func S(l Logger) SLogger {
	return SLogger{Logger: l}
}

// Error logs a message in error level adding provided fields.
func (l SLogger) Error(msg string, fields ...any) {
	l.Logger.Error(nil, msg, fields...)
}

// Errorf logs a formatted message in error level adding provided fields after formatting args.
func (l SLogger) Errorf(msg string, fields ...any) {
	l.Logger.Errorf(nil, msg, fields...)
}
