package logze

import (
	"fmt"
	"runtime"
	"strconv"
)

const (
	functionKey = `{"function":"`
	fileKey     = `","file":"`
	lineKey     = `","line":`
)

// CaptureStackTraceJSON directly captures and formats the current stack trace for logging
// This is more efficient than WithStack + MarshalStack for logging purposes
func CaptureStackTraceJSON() []byte {
	const depth = 32
	var pcs [depth]uintptr
	n := runtime.Callers(2, pcs[:]) // Skip runtime.Callers and CaptureStackTrace

	if n == 0 {
		return nil
	}

	out := make([]byte, 0, n*150)
	out = append(out, '[')
	for i := 0; i < n; i++ {
		pc := pcs[i]
		f := runtime.FuncForPC(pc)
		if f == nil {
			continue
		}

		file, line := f.FileLine(pc)
		out = append(out, functionKey...)
		out = append(out, f.Name()...)
		out = append(out, fileKey...)
		out = append(out, file...)
		out = append(out, lineKey...)
		out = append(out, strconv.Itoa(line)...)
		out = append(out, '}')
		if i < n-1 {
			out = append(out, ',')
		}
	}

	out = append(out, ']')
	return out
}

// WithStack wraps an error with stack trace information (replaces errors.WithStack)
//
// Example:
//
//	err := errors.New("an error occurred")
//	logze.C().WithStackTrace().WithConsole().WithNoDiode().New().Err(err, "ABC")
//
//	log := setupZerologLogger(os.Stdout)
//	log.Error().Stack().Err(errors.WithStack(err)).Str("key", "value").Int("number", 123).Msg("error message")
func WithStack(err error) error {
	if err == nil {
		return nil
	}

	// Skip if error already has stack trace
	if _, ok := err.(genericStackTraceError); ok {
		return err
	}

	// Capture current stack trace
	const depth = 32
	var pcs [depth]uintptr
	n := runtime.Callers(2, pcs[:]) // Skip runtime.Callers and withStack

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

// Format implements the fmt.Formatter interface for detailed error formatting
func (s *stackError) Format(st fmt.State, verb rune) {
	switch verb {
	case 'v':
		if st.Flag('+') {
			fmt.Fprintf(st, "%+v", s.err)
			st.Write([]byte("\n"))
			for _, pc := range s.stack {
				f := runtime.FuncForPC(pc)
				if f != nil {
					file, line := f.FileLine(pc)
					fmt.Fprintf(st, "%s\n\t%s:%d\n", f.Name(), file, line)
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
