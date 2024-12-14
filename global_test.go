package logze_test

import (
	"bytes"
	"fmt"
	"strings"
	"testing"

	"github.com/maxbolgarin/logze/v2"
	"github.com/pkg/errors"
)

func setupGlobalLogger(buffer *bytes.Buffer, level string) {
	cfg := logze.NewConfig(buffer).WithLevel(level).WithNoDiode()
	logze.Init(cfg)
}

func TestGlobalDefaultPtr(t *testing.T) {
	var b bytes.Buffer
	logze.Init(logze.NewConfig(&b).WithNoDiode())

	log1 := logze.DefaultPtr()
	log1.Info("test message")

	output := b.String()
	if !strings.Contains(output, "test message") {
		t.Errorf("expected %s, got %s", "test message", output)
	}

	var b2 bytes.Buffer
	logze.Init(logze.NewConfig(&b2).WithNoDiode())

	log1.Info("test message")

	output = b2.String()
	if !strings.Contains(output, "test message") {
		t.Errorf("expected %s, got %s", "test message", output)
	}
}

func TestGlobalInfo(t *testing.T) {
	var b bytes.Buffer
	setupGlobalLogger(&b, logze.LevelDebug)

	logze.Info("test message")

	output := b.String()
	if !strings.Contains(output, "level\":\"info") {
		t.Errorf("expected %s, got %s", "level\":\"info", output)
	}
	if !strings.Contains(output, "test message") {
		t.Errorf("expected %s, got %s", "test message", output)
	}
}

func TestGlobalInfof(t *testing.T) {
	var b bytes.Buffer
	setupGlobalLogger(&b, logze.LevelInfo)

	logze.Infof("test message %d", 42)

	output := b.String()
	if !strings.Contains(output, "level\":\"info") || !strings.Contains(output, "test message 42") {
		t.Errorf("expected formatted info message, got %s", output)
	}
}

func TestGlobalDebug(t *testing.T) {
	var b bytes.Buffer
	setupGlobalLogger(&b, logze.LevelDebug)

	logze.Debug("debug message")

	output := b.String()
	if !strings.Contains(output, "level\":\"debug") {
		t.Errorf("expected log level debug, got %s", output)
	}
	if !strings.Contains(output, "debug message") {
		t.Errorf("expected log message 'debug message', got %s", output)
	}
}

func TestGlobalDebugf(t *testing.T) {
	var b bytes.Buffer
	setupGlobalLogger(&b, logze.LevelDebug)

	logze.Debugf("debug value: %v", 100)

	output := b.String()
	if !strings.Contains(output, "level\":\"debug") || !strings.Contains(output, "debug value: 100") {
		t.Errorf("expected formatted debug message, got %s", output)
	}
}

func TestGlobalWarn(t *testing.T) {
	var b bytes.Buffer
	setupGlobalLogger(&b, logze.LevelWarn)

	logze.Warn("warn message")

	output := b.String()
	if !strings.Contains(output, "level\":\"warn") {
		t.Errorf("expected log level warn, got %s", output)
	}
	if !strings.Contains(output, "warn message") {
		t.Errorf("expected log message 'warn message', got %s", output)
	}
}

func TestGlobalWarnf(t *testing.T) {
	var b bytes.Buffer
	setupGlobalLogger(&b, logze.LevelWarn)

	logze.Warnf("warn message: %s", "check")

	output := b.String()
	if !strings.Contains(output, "level\":\"warn") || !strings.Contains(output, "warn message: check") {
		t.Errorf("expected formatted warn message, got %s", output)
	}
}

func TestGlobalError(t *testing.T) {
	var b bytes.Buffer
	setupGlobalLogger(&b, logze.LevelError)

	logze.Error("error message")

	output := b.String()
	if !strings.Contains(output, "level\":\"error") {
		t.Errorf("expected log level error, got %s", output)
	}
	if !strings.Contains(output, "error message") {
		t.Errorf("expected log message 'error message', got %s", output)
	}
}

func TestGlobalErrorf(t *testing.T) {
	var b bytes.Buffer
	setupGlobalLogger(&b, logze.LevelError)

	logze.Errorf("error operation %s", "failed")

	output := b.String()
	if !strings.Contains(output, "level\":\"error") || !strings.Contains(output, "error operation failed") {
		t.Errorf("expected formatted error message, got %s", output)
	}
}

func TestGlobalErrStack(t *testing.T) {
	var b bytes.Buffer
	setupGlobalLogger(&b, logze.LevelError)

	err := errors.New("stack trace test error")
	logze.ErrStack(err, "additional", "info")

	output := b.String()
	if !strings.Contains(output, "level\":\"error") || !strings.Contains(output, "TestGlobalErrStack") || !strings.Contains(output, "additional\":\"info") {
		t.Errorf("expected error message with stack trace, got %s", output)
	}
}

func TestGlobalPrint(t *testing.T) {
	var b bytes.Buffer
	setupGlobalLogger(&b, logze.LevelInfo)

	logze.Print("log without level")

	output := b.String()
	if !strings.Contains(output, "log without level") {
		t.Errorf("expected unlevelled log, got %s", output)
	}
}

func TestGlobalPrintf(t *testing.T) {
	var b bytes.Buffer
	setupGlobalLogger(&b, logze.LevelInfo)

	logze.Printf("log without level %s", "status")

	output := b.String()
	if !strings.Contains(output, "log without level status") {
		t.Errorf("expected unlevelled formatted log, got %s", output)
	}
}

func TestGlobalErrorCounter(t *testing.T) {
	var ec logze.SimpleErrorCounter
	cfg := logze.NewConfig().WithErrorCounter(&ec).WithLevel(logze.LevelError)
	logze.Init(cfg)

	if ec.Count.Load() != 0 {
		t.Errorf("expected 0, got %d", ec.Count.Load())
	}

	logze.Err(errors.New("error occurred"), "error test")
	if ec.Count.Load() != 1 {
		t.Errorf("expected 1, got %d", ec.Count.Load())
	}
}

func TestGlobalPanic(t *testing.T) {
	defer func() {
		if r := recover(); r != nil {
			if !strings.Contains(fmt.Sprint(r), "panic message") {
				t.Errorf("expected panic message, got %s", r)
			}
		}
	}()
	var b bytes.Buffer
	setupGlobalLogger(&b, logze.LevelDebug)
	logze.Panic("panic message")
}

func TestGlobalIgnoreMessages(t *testing.T) {
	var b bytes.Buffer
	cfg := logze.NewConfig(&b).WithLevel(logze.LevelDebug).WithNoDiode().WithToIgnore("ignore me")
	logze.Init(cfg)

	logze.Info("this should be logged")
	logze.Info("ignore me")

	output := b.String()
	if !strings.Contains(output, "this should be logged") {
		t.Errorf("expected %s, got %s", "this should be logged", output)
	}
	if strings.Contains(output, "ignore me") {
		t.Errorf("expected %s, got %s", "ignore me", output)
	}
}
