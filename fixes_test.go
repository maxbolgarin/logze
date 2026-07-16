package logze_test

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	stdlog "log"
	"os"
	"strings"
	"sync"
	"testing"

	"github.com/maxbolgarin/logze/v2"
)

// TestMain verifies properties of the pristine global logger before any test
// mutates it, then runs the suite.
func TestMain(m *testing.M) {
	// Importing logze must not spawn a diode goroutine: the implicit global
	// logger is synchronous.
	if logze.Default().HasDiode() {
		fmt.Fprintln(os.Stderr, "FAIL: initial global logger must not have a diode writer")
		os.Exit(1)
	}
	os.Exit(m.Run())
}

func TestUpdatePreservesToIgnoreRegex(t *testing.T) {
	var b bytes.Buffer
	cfg := logze.NewConfig(&b).WithNoDiode().WithToIgnoreRegex(`^skip`)
	logger := logze.New(cfg)

	logger.Update(cfg)

	logger.Info("skip this message")
	logger.Info("keep this message")

	output := b.String()
	if strings.Contains(output, "skip this message") {
		t.Error("expected regex-ignored message to be dropped after Update")
	}
	if !strings.Contains(output, "keep this message") {
		t.Error("expected non-ignored message to be logged after Update")
	}
}

func TestFormatVerbCounting(t *testing.T) {
	tests := []struct {
		name       string
		format     string
		args       []interface{}
		wantMsg    string
		wantFields []string
	}{
		{
			name:       "escaped percent with fields",
			format:     "done 100%%",
			args:       []interface{}{"key", "value"},
			wantMsg:    `"message":"done 100%"`,
			wantFields: []string{`"key":"value"`},
		},
		{
			name:    "verb followed by escaped percent",
			format:  "value: %d%%",
			args:    []interface{}{42},
			wantMsg: `"message":"value: 42%"`,
		},
		{
			name:    "wrap verb rendered as string",
			format:  "cause: %w",
			args:    []interface{}{errors.New("boom")},
			wantMsg: `"message":"cause: boom"`,
		},
		{
			name:       "verbs with extra field args",
			format:     "user %s did %s",
			args:       []interface{}{"alice", "login", "request_id", "r1"},
			wantMsg:    `"message":"user alice did login"`,
			wantFields: []string{`"request_id":"r1"`},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			var b bytes.Buffer
			logger := logze.New(logze.NewConfig(&b).WithNoDiode())

			logger.Infof(tc.format, tc.args...)

			output := b.String()
			if !strings.Contains(output, tc.wantMsg) {
				t.Errorf("expected %s in output, got %s", tc.wantMsg, output)
			}
			for _, f := range tc.wantFields {
				if !strings.Contains(output, f) {
					t.Errorf("expected field %s in output, got %s", f, output)
				}
			}
		})
	}
}

func TestWithStackErrorChain(t *testing.T) {
	base := errors.New("base error")
	wrapped := fmt.Errorf("context: %w", base)

	withStack := logze.WithStack(wrapped)

	if !errors.Is(withStack, base) {
		t.Error("expected errors.Is to see through WithStack wrapper")
	}
	if errors.Unwrap(withStack) != wrapped {
		t.Error("expected Unwrap to return the wrapped error")
	}
	if withStack.Error() != wrapped.Error() {
		t.Errorf("expected error message %q, got %q", wrapped.Error(), withStack.Error())
	}

	var st interface{ StackTrace() []uintptr }
	if !errors.As(withStack, &st) {
		t.Fatal("expected stack trace interface via errors.As")
	}
	if len(st.StackTrace()) == 0 {
		t.Error("expected non-empty stack trace")
	}

	// Wrapping again must not add a second stack, even through error chains
	doubleWrapped := logze.WithStack(fmt.Errorf("more context: %w", withStack))
	if doubleWrapped.Error() != "more context: "+wrapped.Error() {
		t.Errorf("unexpected double-wrapped error message: %q", doubleWrapped.Error())
	}
	if _, ok := doubleWrapped.(interface{ StackTrace() []uintptr }); ok {
		t.Error("expected no new stack wrapper when the chain already has one")
	}
}

func TestCaptureStackTraceJSONIsValid(t *testing.T) {
	stack := logze.CaptureStackTraceJSON()
	if len(stack) == 0 {
		t.Fatal("expected non-empty stack trace")
	}
	if !json.Valid(stack) {
		t.Fatalf("expected valid JSON, got %s", stack)
	}
	if !strings.Contains(string(stack), "TestCaptureStackTraceJSONIsValid") {
		t.Errorf("expected calling function in stack trace, got %s", stack)
	}
}

func TestPercentageSamplerBounds(t *testing.T) {
	t.Run("zero percentage never logs", func(t *testing.T) {
		var b bytes.Buffer
		logger := logze.New(logze.NewConfig(&b).WithNoDiode().WithPercentageSampler(0))
		for i := 0; i < 100; i++ {
			logger.Info("sampled message")
		}
		if b.Len() > 0 {
			t.Errorf("expected no output with zero percentage, got %s", b.String())
		}
	})

	t.Run("negative percentage never logs", func(t *testing.T) {
		var b bytes.Buffer
		logger := logze.New(logze.NewConfig(&b).WithNoDiode().WithPercentageSampler(-1))
		for i := 0; i < 100; i++ {
			logger.Info("sampled message")
		}
		if b.Len() > 0 {
			t.Errorf("expected no output with negative percentage, got %s", b.String())
		}
	})

	t.Run("percentage above one always logs", func(t *testing.T) {
		var b bytes.Buffer
		logger := logze.New(logze.NewConfig(&b).WithNoDiode().WithPercentageSampler(2))
		for i := 0; i < 10; i++ {
			logger.Info("sampled message")
		}
		if got := strings.Count(b.String(), "sampled message"); got != 10 {
			t.Errorf("expected 10 logged messages, got %d", got)
		}
	})
}

func TestSimpleErrorCounterLoad(t *testing.T) {
	counter := &logze.SimpleErrorCounter{}
	if counter.Load() != 0 {
		t.Errorf("expected 0, got %d", counter.Load())
	}
	counter.Inc(errors.New("boom"))
	counter.Inc(errors.New("boom again"))
	counter.Inc(nil) // nil errors are not counted
	if counter.Load() != 2 {
		t.Errorf("expected 2, got %d", counter.Load())
	}
}

func TestDisabledLevelSkipsErrorWork(t *testing.T) {
	var b bytes.Buffer
	counter := &logze.SimpleErrorCounter{}
	logger := logze.New(logze.NewConfig(&b).
		WithLevel(logze.LevelFatal).
		WithNoDiode().
		WithErrorCounter(counter)).
		WithStack()

	logger.Err(errors.New("suppressed"), "below level")
	logger.Errf(errors.New("suppressed"), "below level %d", 1)
	logger.Error("below level", "error", errors.New("suppressed"))

	if b.Len() > 0 {
		t.Errorf("expected no output below fatal level, got %s", b.String())
	}
	if counter.Load() != 0 {
		t.Errorf("expected no counted errors for suppressed logs, got %d", counter.Load())
	}
}

func TestGlobalConcurrentAccess(t *testing.T) {
	defer logze.SetDefault(logze.New(logze.NewConfig(os.Stderr).WithNoDiode()))

	stop := make(chan struct{})
	var wg sync.WaitGroup
	for i := 0; i < 4; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for {
				select {
				case <-stop:
					return
				default:
				}
				logze.Info("message", "key", 1)
				logze.Debugf("formatted %d", 2)
				logze.With("a", "b").Warn("warn")
				logze.Err(errors.New("boom"), "error")
				_ = logze.D()
				_ = logze.Default()
				_ = logze.Raw()
				_ = logze.GetErrorCounter()
				logze.Printf("plain %d", 3)
			}
		}()
	}

	for i := 0; i < 50; i++ {
		logze.SetDefault(logze.New(logze.NewConfig(io.Discard).WithNoDiode()))
		logze.Init(logze.NewConfig(io.Discard).WithNoDiode())
	}
	close(stop)
	wg.Wait()
}

func TestSetStdLoggerKeepsGlobal(t *testing.T) {
	var globalBuf, stdBuf bytes.Buffer
	logze.SetDefault(logze.New(logze.NewConfig(&globalBuf).WithNoDiode()))
	defer logze.SetDefault(logze.New(logze.NewConfig(os.Stderr).WithNoDiode()))

	logze.SetStdLogger(logze.New(logze.NewConfig(&stdBuf).WithNoDiode()))

	logze.Info("global message")
	stdlog.Println("std message")

	if !strings.Contains(globalBuf.String(), "global message") {
		t.Error("expected global logger to remain unchanged after SetStdLogger")
	}
	if strings.Contains(globalBuf.String(), "std message") {
		t.Error("expected std message to not reach the global logger writer")
	}
	if !strings.Contains(stdBuf.String(), "std message") {
		t.Error("expected std message in the std logger writer")
	}
}
