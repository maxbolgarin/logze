package logze

import (
	"errors"
	"fmt"
	"runtime"
	"strconv"
)

// CaptureStackTraceJSON directly captures and formats the current stack trace for logging.
// It returns a valid JSON array of {"function":...,"file":...,"line":...} objects.
// This is more efficient than WithStack + MarshalStack for logging purposes.
func CaptureStackTraceJSON() []byte {
	const depth = 32
	var pcs [depth]uintptr
	n := runtime.Callers(2, pcs[:]) // Skip runtime.Callers and CaptureStackTraceJSON

	if n == 0 {
		return nil
	}

	frames := runtime.CallersFrames(pcs[:n])

	out := make([]byte, 0, n*150)
	out = append(out, '[')
	first := true
	for {
		frame, more := frames.Next()
		if frame.Function != "" || frame.File != "" {
			if !first {
				out = append(out, ',')
			}
			first = false
			out = append(out, `{"function":`...)
			out = appendQuotedString(out, frame.Function)
			out = append(out, `,"file":`...)
			out = appendQuotedString(out, frame.File)
			out = append(out, `,"line":`...)
			out = strconv.AppendInt(out, int64(frame.Line), 10)
			out = append(out, '}')
		}
		if !more {
			break
		}
	}
	out = append(out, ']')

	return out
}

// appendQuotedString appends s as a quoted JSON string. Typical function names
// and file paths need no escaping, so it takes a cheap fast path and falls back
// to strconv.AppendQuote only when escaping is required.
func appendQuotedString(out []byte, s string) []byte {
	for i := 0; i < len(s); i++ {
		if c := s[i]; c < 0x20 || c == '"' || c == '\\' || c >= 0x80 {
			return strconv.AppendQuote(out, s)
		}
	}
	out = append(out, '"')
	out = append(out, s...)
	return append(out, '"')
}

// WithStack wraps an error with stack trace information (replaces errors.WithStack).
// It returns the error unchanged if it is nil or if a stack trace is already attached
// anywhere in its chain. The returned error supports errors.Is, errors.As and
// errors.Unwrap against the wrapped error.
//
// Example:
//
//	err := errors.New("an error occurred")
//	logze.C().WithStackTrace().WithConsole().WithNoDiode().New().Err(err, "ABC")
//
//	log := setupZerologLogger(os.Stdout)
//	log.Error().Stack().Err(logze.WithStack(err)).Str("key", "value").Int("number", 123).Msg("error message")
func WithStack(err error) error {
	if err == nil {
		return nil
	}

	// Skip if the error chain already carries a stack trace
	var st genericStackTraceError
	if errors.As(err, &st) {
		return err
	}

	// Capture current stack trace
	const depth = 32
	var pcs [depth]uintptr
	n := runtime.Callers(2, pcs[:]) // Skip runtime.Callers and WithStack

	stack := make(stackTrace, n)
	copy(stack, pcs[:n])

	return &stackError{
		err:   err,
		stack: stack,
	}
}

// Custom stack trace implementation to replace pkg/errors functionality
type genericStackTraceError interface {
	StackTrace() []uintptr
}

// stackTrace represents a stack trace captured at runtime
type stackTrace []uintptr

// stackError wraps an error with stack trace information
type stackError struct {
	err   error
	stack stackTrace
}

// Error implements the error interface
func (s *stackError) Error() string {
	return s.err.Error()
}

// Unwrap returns the wrapped error, making the wrapper transparent
// to errors.Is and errors.As.
func (s *stackError) Unwrap() error {
	return s.err
}

// StackTrace implements the genericStackTraceError interface
func (s *stackError) StackTrace() []uintptr {
	return s.stack
}

// Format implements the fmt.Formatter interface for detailed error formatting
func (s *stackError) Format(st fmt.State, verb rune) {
	switch verb {
	case 'v':
		if st.Flag('+') {
			fmt.Fprintf(st, "%+v", s.err)
			st.Write([]byte("\n"))
			frames := runtime.CallersFrames(s.stack)
			for {
				frame, more := frames.Next()
				if frame.Function != "" || frame.File != "" {
					fmt.Fprintf(st, "%s\n\t%s:%d\n", frame.Function, frame.File, frame.Line)
				}
				if !more {
					break
				}
			}
			return
		}
		fallthrough
	case 's':
		st.Write([]byte(s.Error()))
	case 'q':
		fmt.Fprintf(st, "%q", s.Error())
	}
}
