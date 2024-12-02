package logze

import (
	"io"
	"os"
	"sync/atomic"
	"time"

	"github.com/rs/zerolog"
)

// DefaultDiodeSize is a default size of a diode writer. Logs will be lost if there will be more logs than that value
// in a small period of time (of time less that Config.DiodePollingInterval).
const DefaultDiodeSize = 1000

// Enumerating string representations of all supported levels.
const (
	TraceLevel    = "trace"
	DebugLevel    = "debug"
	InfoLevel     = "info"
	WarnLevel     = "warn"
	ErrorLevel    = "error"
	FatalLevel    = "fatal"
	DisabledLevel = "disabled"
)

// Levels is a list of all supported levels in string format.
var Levels = []string{
	TraceLevel, DebugLevel, InfoLevel, WarnLevel, ErrorLevel, FatalLevel, DisabledLevel,
}

// LevelsAny is a list of all supported levels in any format.
var LevelsAny = []any{
	TraceLevel, DebugLevel, InfoLevel, WarnLevel, ErrorLevel, FatalLevel, DisabledLevel,
}

// Config is using for initializing [Logger]. You should use [NewConfig] and With* methods instead of creating
// a [Config] struct directly.
type Config struct {
	// Writers is a list of writers where logger will log its data.
	// Default value is [io.Discard].
	Writers []io.Writer

	// Level is a log level in string format. Supported levels are:
	// trace, debug, info, warn, error, fatal, disabled.
	Level string

	// Hook is a zerolog.Hook that will be used when creating logger.
	// Default value is nil.
	Hook zerolog.Hook

	// ToIgnore is a list of messages that will be ignored.
	// Default value is nil.
	ToIgnore []string

	// ErrorCounter is a counter of logged errors. Use WithSimpleErrorCounter method to use a simple counter.
	// Default value is nil.
	ErrorCounter ErrorCounter

	// DiodeSize is a size of a diode writer. Logs will be lost if there will be more logs than that value
	// in a small period of time (of time less that Config.DiodePollingInterval).
	// Default value is 1000.
	DiodeSize int

	// DiodePollingInterval is a time after which diode writer will flush its buffer.
	// Default value is 100ms.
	DiodePollingInterval time.Duration

	// DiodeAlertFunc is a function that will be called when diode writer will flush its buffer.
	// Default value is a function that logs a message in warn level.
	DiodeAlertFunc func(int)

	// DisableDiode if true, will disable diode writer.
	// Default value is false.
	DisableDiode bool

	// StackTrace if true, will enable stack trace for Error and Errorf methods.
	// Default value is false.
	StackTrace bool
}

// NewConfig returns [Config] with provided list of [io.Writer], where [Logger] should logs its data.
func NewConfig(writers ...io.Writer) Config {
	return Config{
		Writers: writers,
	}
}

// WithLevel returns [Config] with initialized level (in string format) provided as argument.
func (c Config) WithLevel(level string) Config {
	c.Level = level
	return c
}

// WithHook returns [Config] with initialized [zerolog.Hook] provided as argument.
func (c Config) WithHook(hook zerolog.Hook) Config {
	c.Hook = hook
	return c
}

// WithWriter returns [Config] with added provided [io.Writer] to a list of writers.
func (c Config) WithWriter(w io.Writer) Config {
	c.Writers = append(c.Writers, w)
	return c
}

// WithConsole returns [Config] with a configurated output to stderr in a pretty console format with colors.
// This format may significantly slow down logging in an application compared to a default JSON format.
func (c Config) WithConsole() Config {
	return c.WithWriter(getConsoleWriter(os.Stderr, true))
}

// WithConsoleNoColor returns [Config] a with configurated output to stderr in a pretty console format without colors.
// This format may significantly slow down logging in an application compared to a default JSON format.
func (c Config) WithConsoleNoColor() Config {
	return c.WithWriter(getConsoleWriter(os.Stderr, false))
}

// WithConsoleJSON returns [Config] with a configurated output to stderr in a JSON format.
func (c Config) WithConsoleJSON() Config {
	return c.WithWriter(os.Stderr)
}

// WithToIgnore returns [Config] with a list of messages that will be ignored.
func (c Config) WithToIgnore(toIgnore ...string) Config {
	c.ToIgnore = toIgnore
	return c
}

// WithDiodeSize returns [Config] with a new size of diode writer.
func (c Config) WithDiodeSize(size int) Config {
	c.DiodeSize = size
	return c
}

// WithDiodePollingInterval returns [Config] with enabled diode polling with provided interval.
func (c Config) WithDiodePollingInterval(interval time.Duration) Config {
	c.DiodePollingInterval = interval
	return c
}

// WithDiodeAlert returns [Config] with provided diode alert func.
func (c Config) WithDiodeAlert(foo func(int)) Config {
	c.DiodeAlertFunc = foo
	return c
}

// WithDisabledDiode returns [Config] with disabled diode writer.
func (c Config) WithDisabledDiode() Config {
	c.DisableDiode = true
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

// WithErrorCounter returns [Config] with a simple [ErrorCounter] inited with the provided name.
func (c Config) WithSimpleErrorCounter(name string) Config {
	c.ErrorCounter = newErrorCounter(name)
	return c
}

func getConsoleWriter(w io.Writer, color bool) zerolog.ConsoleWriter {
	return zerolog.ConsoleWriter{
		Out:        w,
		NoColor:    !color,
		TimeFormat: time.DateTime,
	}
}

// ErrorCounter provides an interface to count logged errors. Use WithSimpleErrorCounter in [Config] creation
// to use a simple error counter and GetErrorCounter method in [Logger] to retrieve it.
type ErrorCounter interface {
	Inc(err ...error)
}

type simpleErrorCounter struct {
	name  string
	count atomic.Int64
}

func newErrorCounter(name string) *simpleErrorCounter {
	return &simpleErrorCounter{
		name: name,
	}
}

func (c *simpleErrorCounter) Name() string {
	return c.name
}

func (c *simpleErrorCounter) Inc(...error) {
	c.count.Add(1)
}

func (c *simpleErrorCounter) Get() int64 {
	return c.count.Load()
}

type noopErrorCounter struct{}

func (noopErrorCounter) Name() string { return "" }
func (noopErrorCounter) Inc(...error) {}
func (noopErrorCounter) Get() int64   { return 0 }
