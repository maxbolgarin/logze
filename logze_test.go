package logze_test

import (
	"bytes"
	"fmt"
	"strings"
	"testing"

	"github.com/maxbolgarin/logze"
	"github.com/pkg/errors"
)

func TestLoggerInitialization(t *testing.T) {
	var b bytes.Buffer
	cfg := logze.NewConfig(&b).WithLevel(logze.DebugLevel).WithNoDiode()
	logger := logze.New(cfg)

	if logger.NotInited() {
		t.Errorf("expected logger to be inited")
	}
	if logger.Raw() == nil {
		t.Errorf("expected logger to be not nil")
	}
}

func TestLoggerInfo(t *testing.T) {
	var b bytes.Buffer
	cfg := logze.NewConfig(&b).WithLevel(logze.DebugLevel).WithNoDiode()
	logger := logze.New(cfg)

	logger.Info("test message")

	output := b.String()
	if !strings.Contains(output, "level\":\"info") {
		t.Errorf("expected %s, got %s", "level\":\"info", output)
	}
	if !strings.Contains(output, "test message") {
		t.Errorf("expected %s, got %s", "test message", output)
	}
}

func TestLoggerWithFields(t *testing.T) {
	var b bytes.Buffer
	cfg := logze.NewConfig(&b).WithLevel(logze.DebugLevel).WithNoDiode()
	logger := logze.New(cfg).WithFields("foo", "bar")

	logger.Info("test message")

	output := b.String()
	if !strings.Contains(output, "foo\":\"bar") {
		t.Errorf("expected %s, got %s", "foo\":\"bar", output)
	}
	if !strings.Contains(output, "test message") {
		t.Errorf("expected %s, got %s", "test message", output)
	}
}

func TestLoggerIgnoreMessages(t *testing.T) {
	var b bytes.Buffer
	cfg := logze.NewConfig(&b).WithLevel(logze.DebugLevel).WithNoDiode().WithToIgnore("ignore me")
	logger := logze.New(cfg)

	logger.Info("this should be logged")
	logger.Info("ignore me")

	output := b.String()
	if !strings.Contains(output, "this should be logged") {
		t.Errorf("expected %s, got %s", "this should be logged", output)
	}
	if strings.Contains(output, "ignore me") {
		t.Errorf("expected %s, got %s", "ignore me", output)
	}
}

func TestLoggerErrorCounter(t *testing.T) {
	var ec logze.SimpleErrorCounter
	cfg := logze.NewConfig().WithErrorCounter(&ec).WithLevel(logze.ErrorLevel)
	logger := logze.New(cfg)

	if ec.Count.Load() != 0 {
		t.Errorf("expected 0, got %d", ec.Count.Load())
	}

	logger.Err(errors.New("error occurred"), "error test")
	if ec.Count.Load() != 1 {
		t.Errorf("expected 1, got %d", ec.Count.Load())
	}
}

func TestLoggerPanic(t *testing.T) {
	defer func() {
		if r := recover(); r != nil {
			if !strings.Contains(fmt.Sprint(r), "panic message") {
				t.Errorf("expected panic message, got %s", r)
			}
		}
	}()
	var b bytes.Buffer
	cfg := logze.NewConfig(&b).WithLevel(logze.DebugLevel).WithNoDiode()
	logger := logze.New(cfg)
	logger.Panic("panic message")
}

func TestLoggerDebug(t *testing.T) {
	var b bytes.Buffer
	cfg := logze.NewConfig(&b).WithLevel(logze.DebugLevel).WithNoDiode()
	logger := logze.New(cfg)

	logger.Debug("debug message")

	output := b.String()
	if !strings.Contains(output, "level\":\"debug") {
		t.Errorf("expected log level debug, got %s", output)
	}
	if !strings.Contains(output, "debug message") {
		t.Errorf("expected log message 'debug message', got %s", output)
	}
}

func TestLoggerWarn(t *testing.T) {
	var b bytes.Buffer
	cfg := logze.NewConfig(&b).WithLevel(logze.WarnLevel).WithNoDiode()
	logger := logze.New(cfg)

	logger.Warn("warn message")

	output := b.String()
	if !strings.Contains(output, "level\":\"warn") {
		t.Errorf("expected log level warn, got %s", output)
	}
	if !strings.Contains(output, "warn message") {
		t.Errorf("expected log message 'warn message', got %s", output)
	}
}

func TestLoggerError(t *testing.T) {
	var b bytes.Buffer
	cfg := logze.NewConfig(&b).WithLevel(logze.ErrorLevel).WithNoDiode()
	logger := logze.New(cfg)

	logger.Error("error message")

	output := b.String()
	if !strings.Contains(output, "level\":\"error") {
		t.Errorf("expected log level error, got %s", output)
	}
	if !strings.Contains(output, "error message") {
		t.Errorf("expected log message 'error message', got %s", output)
	}
}

func TestLoggerFatal(t *testing.T) {
	// It's challenging to test fatal logs without stopping execution,
	// so these tests are theoretical – you might use an interface or mock for os.Exit if necessary.
}

func TestUpdateLoggerConfiguration(t *testing.T) {
	var b1, b2 bytes.Buffer
	cfg1 := logze.NewConfig(&b1).WithLevel(logze.InfoLevel).WithNoDiode()
	logger := logze.New(cfg1)

	logger.Info("initial config message")

	output1 := b1.String()
	if !strings.Contains(output1, "initial config message") {
		t.Errorf("expected log 'initial config message', got %s", output1)
	}

	cfg2 := logze.NewConfig(&b2).WithLevel(logze.DebugLevel).WithNoDiode()
	logger.Update(cfg2)

	logger.Debug("updated config message")
	output2 := b2.String()
	if !strings.Contains(output2, "updated config message") {
		t.Errorf("expected log 'updated config message', got %s", output2)
	}
}

func TestLoggerInfof(t *testing.T) {
	var b bytes.Buffer
	cfg := logze.NewConfig(&b).WithLevel(logze.InfoLevel).WithNoDiode()
	logger := logze.New(cfg)

	logger.Infof("test message %d", 42, "a", "b")

	output := b.String()
	if !strings.Contains(output, "level\":\"info") || !strings.Contains(output, "test message 42") || !strings.Contains(output, "\"a\":\"b\"") {
		t.Errorf("expected formatted info message, got %s", output)
	}
}

func TestLoggerDebugf(t *testing.T) {
	var b bytes.Buffer
	cfg := logze.NewConfig(&b).WithLevel(logze.DebugLevel).WithNoDiode()
	logger := logze.New(cfg)

	logger.Debugf("debug value: %v", 100)

	output := b.String()
	if !strings.Contains(output, "level\":\"debug") || !strings.Contains(output, "debug value: 100") {
		t.Errorf("expected formatted debug message, got %s", output)
	}
}

func TestLoggerErrf(t *testing.T) {
	var b bytes.Buffer
	cfg := logze.NewConfig(&b).WithLevel(logze.ErrorLevel).WithNoDiode()
	logger := logze.New(cfg)

	logger.Errf(errors.New("123"), "error operation %s", "failed", "a", "b")

	output := b.String()
	if !strings.Contains(output, "level\":\"error") || !strings.Contains(output, "error operation failed") || !strings.Contains(output, "123") || !strings.Contains(output, "\"a\":\"b\"") {
		t.Errorf("expected formatted error message, got %s", output)
	}
}

func TestLoggerErrorf(t *testing.T) {
	var b bytes.Buffer
	cfg := logze.NewConfig(&b).WithLevel(logze.ErrorLevel).WithNoDiode()
	logger := logze.New(cfg)

	logger.Errorf("error operation %s", "failed")

	output := b.String()
	if !strings.Contains(output, "level\":\"error") || !strings.Contains(output, "error operation failed") {
		t.Errorf("expected formatted error message, got %s", output)
	}
}

func TestLoggerWarnf(t *testing.T) {
	var b bytes.Buffer
	cfg := logze.NewConfig(&b).WithLevel(logze.WarnLevel).WithNoDiode()
	logger := logze.New(cfg)

	logger.Warnf("warn message: %s", "check")

	output := b.String()
	if !strings.Contains(output, "level\":\"warn") || !strings.Contains(output, "warn message: check") {
		t.Errorf("expected formatted warn message, got %s", output)
	}
}

func TestLoggerErrStack(t *testing.T) {
	var b bytes.Buffer
	cfg := logze.NewConfig(&b).WithLevel(logze.ErrorLevel).WithStackTrace().WithNoDiode()
	logger := logze.New(cfg)

	err := errors.New("stack trace test error")
	logger.ErrStack(err, "additional", "info")

	output := b.String()
	// Since stack traces are long and complex, we verify presence of basic parts
	if !strings.Contains(output, "level\":\"error") || !strings.Contains(output, "TestLoggerErrStack") || !strings.Contains(output, "additional\":\"info") {
		t.Errorf("expected error message with stack trace, got %s", output)
	}

	b.Reset()

	logger.Err(err, "additional")

	output = b.String()
	// Since stack traces are long and complex, we verify presence of basic parts
	if !strings.Contains(output, "level\":\"error") || !strings.Contains(output, "TestLoggerErrStack") || !strings.Contains(output, "message\":\"additional") {
		t.Errorf("expected error message with stack trace, got %s", output)
	}
}

func TestLoggerPrintf(t *testing.T) {
	var b bytes.Buffer
	cfg := logze.NewConfig(&b).WithLevel(logze.InfoLevel).WithNoDiode()
	logger := logze.New(cfg)

	logger.Printf("log without level %s", "status")

	output := b.String()
	if !strings.Contains(output, "log without level status") {
		t.Errorf("expected unlevelled formatted log, got %s", output)
	}
}
