package logze_test

import (
	"io"
	"os"
	"testing"
	"time"

	"github.com/maxbolgarin/logze/v2"
	"github.com/rs/zerolog"
)

func TestNewConfig(t *testing.T) {
	writer := io.Discard
	cfg := logze.NewConfig(writer)

	if len(cfg.Writers) != 1 {
		t.Errorf("expected 1 writer, got %d", len(cfg.Writers))
	}

	if cfg.Writers[0] != writer {
		t.Errorf("expected writer to be io.Discard, got %v", cfg.Writers[0])
	}
}

func TestCShortcut(t *testing.T) {
	writer := io.Discard
	cfg := logze.C(writer)

	if len(cfg.Writers) != 1 {
		t.Errorf("expected 1 writer, got %d", len(cfg.Writers))
	}

	if cfg.Writers[0] != writer {
		t.Errorf("expected writer to be io.Discard, got %v", cfg.Writers[0])
	}
}

func TestWithLevel(t *testing.T) {
	cfg := logze.NewConfig()
	if cfg.Level != "" {
		t.Errorf("expected empty, got %s", cfg.Level)
	}

	cfg = cfg.WithLevel(logze.LevelDebug)
	if cfg.Level != logze.LevelDebug {
		t.Errorf("expected %s, got %s", logze.LevelDebug, cfg.Level)
	}
}

func TestWithHook(t *testing.T) {
	var testHook zerolog.Hook
	cfg := logze.NewConfig().WithHook(testHook)

	if cfg.Hook != testHook {
		t.Errorf("expected hook to be %#v, got %#v", testHook, cfg.Hook)
	}
}

func TestWithWriter(t *testing.T) {
	writer1 := io.Discard
	writer2 := os.Stderr
	cfg := logze.NewConfig(writer1).WithWriter(writer2)

	if len(cfg.Writers) != 2 {
		t.Errorf("expected 2 writers, got %d", len(cfg.Writers))
	}

	if cfg.Writers[1] != writer2 {
		t.Errorf("expected second writer to be os.Stderr, got %v", cfg.Writers[1])
	}
}

func TestWithConsole(t *testing.T) {
	cfg := logze.NewConfig().WithConsole()

	// Assuming getConsoleWriter outputs a particular format,
	// we will not validate that as it depends on the zerolog integration
	if len(cfg.Writers) == 0 {
		t.Errorf("expected at least 1 writer, got %d", len(cfg.Writers))
	}
}

func TestWithConsoleNoColor(t *testing.T) {
	cfg := logze.NewConfig().WithConsoleNoColor()

	if len(cfg.Writers) == 0 {
		t.Errorf("expected at least 1 writer, got %d", len(cfg.Writers))
	}
}

func TestWithConsoleJSON(t *testing.T) {
	cfg := logze.NewConfig().WithConsoleJSON()

	if len(cfg.Writers) == 0 || cfg.Writers[0] != os.Stderr {
		t.Errorf("expected os.Stderr writer, got %v", cfg.Writers)
	}
}

func TestWithToIgnore(t *testing.T) {
	ignoreList := []string{"ignore_this", "and_this"}
	cfg := logze.NewConfig().WithToIgnore(ignoreList...)

	if len(cfg.ToIgnore) != 2 {
		t.Errorf("expected 2 items in ToIgnore, got %d", len(cfg.ToIgnore))
	}

	if cfg.ToIgnore[0] != "ignore_this" || cfg.ToIgnore[1] != "and_this" {
		t.Errorf("unexpected entries in ToIgnore: %v", cfg.ToIgnore)
	}
}

func TestWithTimeFieldFormat(t *testing.T) {
	format := time.RFC1123
	cfg := logze.NewConfig().WithTimeFieldFormat(format)

	if cfg.TimeFieldFormat != format {
		t.Errorf("expected format %s, got %s", format, cfg.TimeFieldFormat)
	}
}

func TestWithDiodeSize(t *testing.T) {
	size := 500
	cfg := logze.NewConfig().WithDiodeSize(size)

	if cfg.DiodeSize != size {
		t.Errorf("expected diode size %d, got %d", size, cfg.DiodeSize)
	}
}

func TestWithDiodePollingInterval(t *testing.T) {
	interval := 20 * time.Millisecond
	cfg := logze.NewConfig().WithDiodePollingInterval(interval)

	if cfg.DiodePollingInterval != interval {
		t.Errorf("expected diode polling interval %v, got %v", interval, cfg.DiodePollingInterval)
	}
}

func TestWithDiodeAlert(t *testing.T) {
	alertFunc := func(size int) {}
	cfg := logze.NewConfig().WithDiodeAlert(alertFunc)

	if cfg.DiodeAlertFunc == nil {
		t.Errorf("expected a diode alert function, got nil")
	}
}

func TestWithNoDiode(t *testing.T) {
	cfg := logze.NewConfig().WithNoDiode()

	if !cfg.NoDiode {
		t.Errorf("expected NoDiode to be true, got false")
	}
}

func TestWithDiodeWaiter(t *testing.T) {
	cfg := logze.NewConfig().WithDiodeWaiter()

	if !cfg.UseDiodeWaiter {
		t.Errorf("expected UseDiodeWaiter to be true, got false")
	}
}

func TestWithStackTrace(t *testing.T) {
	cfg := logze.NewConfig().WithStackTrace()

	if !cfg.StackTrace {
		t.Errorf("expected StackTrace to be true, got false")
	}
}

func TestWithErrorCounter(t *testing.T) {
	// Custom error counter setup
	var customCounter logze.ErrorCounter = &logze.SimpleErrorCounter{}
	cfg := logze.NewConfig().WithErrorCounter(customCounter)

	if cfg.ErrorCounter != customCounter {
		t.Errorf("expected ErrorCounter to be custom, got another instance")
	}
}

func TestWithSimpleErrorCounter(t *testing.T) {
	cfg := logze.NewConfig().WithSimpleErrorCounter()

	if cfg.ErrorCounter == nil {
		t.Error("expected a non-nil ErrorCounter")
	}
}
