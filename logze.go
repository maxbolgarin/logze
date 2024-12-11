// Package logze implements a zerolog wrapper providing a convenient and short interface for structural logging based on slog package.
package logze

import (
	"fmt"
	"io"
	"os"
	"runtime/debug"
	"strings"
	"time"

	"github.com/pkg/errors"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/diode"
	"github.com/rs/zerolog/pkgerrors"
)

// Logger represents an initialized logger.
// Default value behaves as default [zerolog.Logger].
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
//	lg := New(C().WithConsoleJSON(), "foo", "bar")
//	lg.Info("some message", "key", "value")
//	lg.Err(errors.New("some error"), "cannot handle")
//
// You will have output:
//
//	{"level":"info","time":"2023-11-20T18:48:14+03:00","message":"some message","foo":"bar","key":"value"}
//	{"level":"error","time":"2023-11-20T18:48:14+03:00","error":"some error","message":"cannot handle","foo":"bar"}
//
// Warning! If you use diode (default behaviour), logger need some time to flush messages.
// Thats why you won't see any logs if you shoutdown your app right after logging.
// Use [Config.WithNoDiode] to disable it,
// but you will need to fix problem of blocking goroutine when writing may loge in Stderr if you have it.
func New(cfg Config, fields ...any) Logger {
	if len(cfg.Writers) == 0 || cfg.Level == DisabledLevel {
		cfg.Writers = []io.Writer{io.Discard}
	}
	if cfg.Level == "" {
		cfg.Level = InfoLevel
	}
	if cfg.TimeFieldFormat == "" {
		cfg.TimeFieldFormat = time.RFC3339
	}
	zerolog.TimeFieldFormat = cfg.TimeFieldFormat

	level, err := zerolog.ParseLevel(cfg.Level)
	if err != nil {
		panic("cannot parse level=" + cfg.Level)
	}

	output := cfg.Writers[0]
	if len(cfg.Writers) > 1 {
		output = zerolog.MultiLevelWriter(cfg.Writers...)
	}
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
		output = diode.NewWriter(output, cfg.DiodeSize, cfg.DiodePollingInterval, cfg.DiodeAlertFunc)
	}

	l := zerolog.New(output).With().Timestamp().Fields(fields).Logger().Level(level)

	if cfg.Hook != nil {
		l = l.Hook(cfg.Hook)
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
		l:      l,
		inited: true,
	}
}

// NewConsoleJSON returns a new [Logger] with JSON logging to stderr.
func NewConsoleJSON(fields ...any) Logger {
	return New(NewConfig().WithConsoleJSON(), fields...)
}

// Nop returns a new [Logger] with no logging.
func Nop() Logger {
	return Logger{l: zerolog.Nop()}
}

// Update replaces underlying logger with a new one created using provided config and fields.
// It is NOT safe for concurrent use.
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

// WithSimpleErrorCounter returns [Logger] with a simple [ErrorCounter].
func (l Logger) WithSimpleErrorCounter() Logger {
	l.errCounter = newSimpleErrorCounter()
	return l
}

// WithToIgnore returns [Logger] with the provided list of messages to ignore.
func (l Logger) WithToIgnore(toIgnore ...string) Logger {
	l.toIgnore = toIgnore
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
func (l Logger) Err(err error, msg string, fields ...any) {
	l.log(l.setErrorWithStack(l.l.Error(), err), msg, fields)
}

// Error logs a message in error level adding provided fields.
func (l Logger) Error(msg string, fields ...any) {
	l.log(l.l.Error(), msg, fields)
}

// Errorf logs a formatted message in error level adding provided fields after formatting args.
func (l Logger) Errorf(msg string, args ...any) {
	l.logf(l.l.Error(), msg, args)
}

// ErrStack logs a stack trace of provided error as message in error level adding fields.
func (l Logger) ErrStack(err error, fields ...any) {
	_, ok := err.(interface {
		StackForLogger() []any
	})
	if !ok {
		err = errors.WithStack(err)
	}
	l.log(l.l.Error(), fmt.Sprintf("%+v", err), fields)
}

// Fatal logs a message in fatal level using fmt.Sprint to interpret args, then calls os.Exit(1).
func (l Logger) Fatal(v ...any) {
	s := fmt.Sprint(v...)
	l.incErrorConter(errors.New(s))
	l.log(l.l.WithLevel(zerolog.FatalLevel), s, nil)
	os.Exit(1)
}

// Fatalf logs a formatted message in fatal level, then calls os.Exit(1).
func (l Logger) Fatalf(format string, args ...any) {
	l.incErrorConter(fmt.Errorf(format, args...))
	l.log(l.l.WithLevel(zerolog.FatalLevel), format, args)
	os.Exit(1)
}

// Fatalln logs a message in fatal level using fmt.Sprintln to interpret args, then calls os.Exit(1).
func (l Logger) Fatalln(v ...any) {
	s := fmt.Sprintln(v...)
	l.incErrorConter(errors.New(s))
	l.log(l.l.WithLevel(zerolog.FatalLevel), s, nil)
	os.Exit(1)
}

// Panic logs a message in fatal level using fmt.Sprint to interpret args, then calls panic().
func (l Logger) Panic(v ...any) {
	s := fmt.Sprint(v...)
	l.incErrorConter(errors.New(s))
	l.log(l.l.WithLevel(zerolog.FatalLevel), s, nil)
	panic(s)
}

// Panicf logs a formatted message in fatal level, then calls panic().
func (l Logger) Panicf(format string, args ...any) {
	l.incErrorConter(fmt.Errorf(format, args...))
	l.log(l.l.WithLevel(zerolog.FatalLevel), format, args)
	panic(fmt.Sprintf(format, args...))
}

// Panicln logs a message in fatal level using fmt.Sprintln to interpret args, then calls panic().
func (l Logger) Panicln(v ...any) {
	s := fmt.Sprintln(v...)
	l.incErrorConter(errors.New(s))
	l.log(l.l.WithLevel(zerolog.FatalLevel), s, nil)
	panic(s)
}

// Print logs a message without level using [fmt.Sprint] to interpret args.
func (l Logger) Print(v ...any) {
	if len(v) == 0 {
		return
	}
	l.log(l.l.Log(), fmt.Sprint(v...), nil)
}

// PrintStack logs a current stack trace.
func (l Logger) PrintStack(v ...any) {
	stack := debug.Stack()
	l.log(l.l.Log(), string(stack), v)
}

// Log logs a message without level using [fmt.Sprint] to interpret args.
// It is an alias for [Logger.Print].
func (l Logger) Log(v ...any) {
	l.Print(v...)
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
		ev = l.setErrorWithStack(ev, fields...)
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
		ev = l.setErrorWithStack(ev, args...)
		ev = ev.Fields(args[numberOfFormats:])
		args = args[:numberOfFormats]
	}
	if numberOfFormats == 0 && len(args) > 0 {
		ev = l.setErrorWithStack(ev, args...)
		ev = ev.Fields(args)
		args = nil
	}
	if len(args) == 0 {
		ev.Msg(msg)
		return
	}
	ev.Msgf(msg, args...)
}

func (l Logger) setErrorWithStack(ev *zerolog.Event, args ...any) *zerolog.Event {
	for i, a := range args {
		if err, ok := a.(error); ok {
			if l.stackTrace {
				// Hack to use github.com/maxbolgarin/errm without importing it
				errmErr, ok := err.(interface {
					StackForLogger() []any
				})
				if ok {
					ev = ev.Fields(errmErr.StackForLogger())
				} else {
					ev = ev.Stack()
					err = errors.WithStack(err)
				}
			}
			l.incErrorConter(err)
			if i-1 >= 0 {
				// we update underlying array so args updated in place
				_ = append(args[:i-1], args[i+1:]...)
			}
			return ev.Err(err)
		}
	}
	return ev
}

func (l Logger) incErrorConter(err error) {
	if l.errCounter != nil {
		l.errCounter.Inc(err)
	}
}
