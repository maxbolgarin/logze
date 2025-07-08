package logze_test

import (
	"bytes"
	"errors"
	"io"
	"os"
	"strings"
	"sync/atomic"
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

func TestWithLevelMethods(t *testing.T) {
	tests := []struct {
		name     string
		method   func(logze.Config) logze.Config
		expected string
	}{
		{
			name:     "WithTrace",
			method:   logze.Config.WithTrace,
			expected: logze.LevelTrace,
		},
		{
			name:     "WithDebug",
			method:   logze.Config.WithDebug,
			expected: logze.LevelDebug,
		},
		{
			name:     "WithInfo",
			method:   logze.Config.WithInfo,
			expected: logze.LevelInfo,
		},
		{
			name:     "WithWarn",
			method:   logze.Config.WithWarn,
			expected: logze.LevelWarn,
		},
		{
			name:     "WithError",
			method:   logze.Config.WithError,
			expected: logze.LevelError,
		},
		{
			name:     "WithFatal",
			method:   logze.Config.WithFatal,
			expected: logze.LevelFatal,
		},
		{
			name:     "WithDisabled",
			method:   logze.Config.WithDisabled,
			expected: logze.LevelDisabled,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cfg := logze.NewConfig()
			cfg = tt.method(cfg)
			if cfg.Level != tt.expected {
				t.Errorf("expected level %s, got %s", tt.expected, cfg.Level)
			}
		})
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

func TestWithFile(t *testing.T) {
	// Create a temporary file for testing
	tempFile, err := os.CreateTemp("", "logze_test_*.log")
	if err != nil {
		t.Fatalf("Failed to create temp file: %v", err)
	}
	tempFileName := tempFile.Name()
	tempFile.Close()

	// Clean up after the test
	defer os.Remove(tempFileName)

	// Test with default permissions
	cfg := logze.NewConfig()
	cfg, closer, err := cfg.WithFile(tempFileName)
	if err != nil {
		t.Errorf("expected no error with default permissions, got %v", err)
	}
	if closer == nil {
		t.Error("expected closer to not be nil")
	}
	defer closer.Close()

	if len(cfg.Writers) != 1 {
		t.Errorf("expected 1 writer, got %d", len(cfg.Writers))
	}

	// Test with custom permissions
	cfg = logze.NewConfig()
	cfg, closer, err = cfg.WithFile(tempFileName, 0600)
	if err != nil {
		t.Errorf("expected no error with custom permissions, got %v", err)
	}
	if closer == nil {
		t.Error("expected closer to not be nil")
	}
	defer closer.Close()

	if len(cfg.Writers) != 1 {
		t.Errorf("expected 1 writer, got %d", len(cfg.Writers))
	}

	info, err := os.Stat(tempFileName)
	if err != nil {
		t.Errorf("expected no error with custom permissions, got %v", err)
	}

	if info.Mode() != 0600 {
		t.Errorf("expected file mode 0600, got %v", info.Mode())
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

func TestWithAddCaller(t *testing.T) {
	cfg := logze.NewConfig().WithAddCaller()

	if !cfg.AddCaller {
		t.Errorf("expected AddCaller to be true, got false")
	}
}

func TestWithCallerSkipFrameCount(t *testing.T) {
	count := 10
	cfg := logze.NewConfig().WithCallerSkipFrameCount(count)

	if cfg.CallerSkipFrameCount != count {
		t.Errorf("expected CallerSkipFrameCount to be %d, got %d", count, cfg.CallerSkipFrameCount)
	}
}

// Additional comprehensive tests for better coverage

func TestConfigNew(t *testing.T) {
	var b bytes.Buffer
	cfg := logze.NewConfig(&b).WithLevel(logze.LevelInfo).WithNoDiode()

	logger := cfg.New("service", "test")
	logger.Info("test message")

	output := b.String()
	if !strings.Contains(output, "test message") {
		t.Error("expected message from config.New()")
	}
	if !strings.Contains(output, "service\":\"test") {
		t.Error("expected service field from config.New()")
	}
}

func TestConfigLogger(t *testing.T) {
	var b bytes.Buffer
	cfg := logze.NewConfig(&b).WithLevel(logze.LevelInfo).WithNoDiode()

	logger := cfg.Logger("component", "auth")
	logger.Info("auth message")

	output := b.String()
	if !strings.Contains(output, "auth message") {
		t.Error("expected message from config.Logger()")
	}
	if !strings.Contains(output, "component\":\"auth") {
		t.Error("expected component field from config.Logger()")
	}
}

func TestWithHooks(t *testing.T) {
	hook1 := zerolog.HookFunc(func(e *zerolog.Event, level zerolog.Level, msg string) {
		e.Str("hook1", "called")
	})
	hook2 := zerolog.HookFunc(func(e *zerolog.Event, level zerolog.Level, msg string) {
		e.Str("hook2", "called")
	})

	cfg := logze.NewConfig().WithHooks(hook1, hook2)

	if len(cfg.Hooks) != 2 {
		t.Errorf("expected 2 hooks, got %d", len(cfg.Hooks))
	}
}

func TestWithSampler(t *testing.T) {
	sampler := zerolog.RandomSampler(10)
	cfg := logze.NewConfig().WithSampler(sampler)

	if cfg.Sampler == nil {
		t.Error("expected non-nil sampler")
	}
}

func TestWithPercentageSampler(t *testing.T) {
	cfg := logze.NewConfig().WithPercentageSampler(0.5, "debug", "info")

	if cfg.Sampler == nil {
		t.Error("expected non-nil sampler from WithPercentageSampler")
	}

	// Test with 100% sampling
	cfg = logze.NewConfig().WithPercentageSampler(1.0)
	if cfg.Sampler == nil {
		t.Error("expected non-nil sampler for 100% sampling")
	}

	// Test with 0% sampling
	cfg = logze.NewConfig().WithPercentageSampler(0.0)
	if cfg.Sampler == nil {
		t.Error("expected non-nil sampler for 0% sampling")
	}
}

func TestWithBurstSampler(t *testing.T) {
	cfg := logze.NewConfig().WithBurstSampler(0.1, 100, time.Second, "debug")

	if cfg.Sampler == nil {
		t.Error("expected non-nil sampler from WithBurstSampler")
	}

	// Test with zero burst
	cfg = logze.NewConfig().WithBurstSampler(0.5, 0, time.Second)
	if cfg.Sampler == nil {
		t.Error("expected non-nil sampler with zero burst")
	}

	// Test with negative burst (should be corrected to 0)
	cfg = logze.NewConfig().WithBurstSampler(0.5, -10, time.Second)
	if cfg.Sampler == nil {
		t.Error("expected non-nil sampler with negative burst")
	}
}

func TestWithMaxSampler(t *testing.T) {
	cfg := logze.NewConfig().WithMaxSampler(100, time.Second, "debug", "info")

	if cfg.Sampler == nil {
		t.Error("expected non-nil sampler from WithMaxSampler")
	}

	// Test with zero max
	cfg = logze.NewConfig().WithMaxSampler(0, time.Second)
	if cfg.Sampler == nil {
		t.Error("expected non-nil sampler with zero max")
	}
}

func TestSamplerEdgeCases(t *testing.T) {
	// Test percentage sampler with edge cases through public API

	// Test negative percentage (should be handled gracefully)
	cfg := logze.NewConfig().WithPercentageSampler(-0.1, "debug")
	if cfg.Sampler == nil {
		t.Error("expected non-nil sampler for negative percentage")
	}

	// Test over 100% percentage (should be handled gracefully)
	cfg = logze.NewConfig().WithPercentageSampler(1.5, "debug")
	if cfg.Sampler == nil {
		t.Error("expected non-nil sampler for over 100% percentage")
	}

	// Test burst sampler with edge cases

	// Test with negative burst (should be corrected)
	cfg = logze.NewConfig().WithBurstSampler(0.5, -10, time.Second, "debug")
	if cfg.Sampler == nil {
		t.Error("expected non-nil burst sampler with negative burst")
	}

	// Test with zero/negative period (should be corrected)
	cfg = logze.NewConfig().WithBurstSampler(0.5, 100, 0, "debug")
	if cfg.Sampler == nil {
		t.Error("expected non-nil burst sampler with zero period")
	}

	cfg = logze.NewConfig().WithBurstSampler(0.5, 100, -time.Second, "debug")
	if cfg.Sampler == nil {
		t.Error("expected non-nil burst sampler with negative period")
	}

	// Test level sampler behavior

	// Test with no levels (should work)
	cfg = logze.NewConfig().WithPercentageSampler(0.5)
	if cfg.Sampler == nil {
		t.Error("expected non-nil sampler with no levels")
	}

	// Test with invalid level mixed with valid levels
	cfg = logze.NewConfig().WithPercentageSampler(0.5, "invalid", "debug", "info")
	if cfg.Sampler == nil {
		t.Error("expected non-nil sampler ignoring invalid level")
	}
}

func TestSimpleErrorCounter(t *testing.T) {
	counter := &logze.SimpleErrorCounter{}

	// Test initial count
	if atomic.LoadUint64(&counter.Count) != 0 {
		t.Error("expected initial count to be 0")
	}

	// Test increment
	err := errors.New("test error")
	counter.Inc(err)

	if atomic.LoadUint64(&counter.Count) != 1 {
		t.Errorf("expected count to be 1 after increment, got %d", atomic.LoadUint64(&counter.Count))
	}

	// Test multiple increments
	counter.Inc(err)
	counter.Inc(err)

	if atomic.LoadUint64(&counter.Count) != 3 {
		t.Errorf("expected count to be 3 after multiple increments, got %d", atomic.LoadUint64(&counter.Count))
	}

	// Test increment with nil error (should still increment)
	counter.Inc(nil)

	if atomic.LoadUint64(&counter.Count) != 3 {
		t.Errorf("expected count to be 3 after nil error increment, got %d", atomic.LoadUint64(&counter.Count))
	}
}

func TestGetConsoleWriter(t *testing.T) {
	// Test colored console writer
	var b bytes.Buffer
	cfg := logze.NewConfig().WithWriter(&b).WithConsole()

	if len(cfg.Writers) != 2 {
		t.Errorf("expected 2 writers after WithConsole, got %d", len(cfg.Writers))
	}

	// Test non-colored console writer
	b.Reset()
	cfg = logze.NewConfig().WithWriter(&b).WithConsoleNoColor()

	if len(cfg.Writers) != 2 {
		t.Errorf("expected 2 writers after WithConsoleNoColor, got %d", len(cfg.Writers))
	}
}

func TestWithFileEdgeCases(t *testing.T) {
	// Test with non-existent directory
	cfg := logze.NewConfig()
	_, _, err := cfg.WithFile("/non/existent/directory/test.log")

	if err == nil {
		t.Error("expected error when creating file in non-existent directory")
	}

	// Test with multiple permission values (should use first)
	tempDir := t.TempDir()
	tempFile := tempDir + "/test.log"

	cfg, closer, err := cfg.WithFile(tempFile, 0600, 0644)
	if err != nil {
		t.Errorf("expected no error with multiple permissions, got %v", err)
	}
	defer closer.Close()

	info, err := os.Stat(tempFile)
	if err != nil {
		t.Errorf("expected file to exist, got %v", err)
	}

	if info.Mode()&0777 != 0600 {
		t.Errorf("expected file mode 0600, got %v", info.Mode()&0777)
	}
}

func TestConfigChaining(t *testing.T) {
	// Test method chaining
	cfg := logze.NewConfig().
		WithLevel("debug").
		WithTrace().
		WithConsoleJSON().
		WithAddCaller().
		WithStackTrace().
		WithNoDiode().
		WithSimpleErrorCounter().
		WithToIgnore("ignore").
		WithPercentageSampler(0.5, "debug")

	if cfg.Level != logze.LevelTrace {
		t.Errorf("expected trace level, got %s", cfg.Level)
	}
	if !cfg.AddCaller {
		t.Error("expected AddCaller to be true")
	}
	if !cfg.StackTrace {
		t.Error("expected StackTrace to be true")
	}
	if !cfg.NoDiode {
		t.Error("expected NoDiode to be true")
	}
	if cfg.ErrorCounter == nil {
		t.Error("expected non-nil ErrorCounter")
	}
	if len(cfg.ToIgnore) != 1 {
		t.Errorf("expected 1 ignore rule, got %d", len(cfg.ToIgnore))
	}
	if cfg.Sampler == nil {
		t.Error("expected non-nil sampler")
	}
}

func TestNewConfigEdgeCases(t *testing.T) {
	// Test with no writers
	cfg := logze.NewConfig()
	if len(cfg.Writers) != 0 {
		t.Errorf("expected 0 writers, got %d", len(cfg.Writers))
	}

	// Test with multiple writers
	cfg = logze.NewConfig(os.Stdout, os.Stderr, io.Discard)
	if len(cfg.Writers) != 3 {
		t.Errorf("expected 3 writers, got %d", len(cfg.Writers))
	}
}

// Test sampling functionality with actual logger
func TestSamplingIntegration(t *testing.T) {
	var b bytes.Buffer

	// Test percentage sampler with 0% (should log nothing)
	cfg := logze.NewConfig(&b).WithLevel("debug").WithNoDiode().WithPercentageSampler(0, "debug")
	logger := cfg.New()

	// Try to log minterface{} messages
	for i := 0; i < 100; i++ {
		logger.Debug("debug message")
	}

	if b.Len() > 0 {
		t.Error("expected no output with 0% sampling")
	}

	// Test max sampler
	b.Reset()
	cfg = logze.NewConfig(&b).WithLevel("debug").WithNoDiode().WithMaxSampler(1, time.Second, "debug")
	logger = cfg.New()

	logger.Debug("first debug")  // Should appear
	logger.Debug("second debug") // Should be filtered

	output := b.String()
	lines := strings.Count(output, "\n")
	if lines > 1 {
		t.Errorf("expected at most 1 log line with max sampler, got %d", lines)
	}
}

// Test all remaining configuration options

func TestTimeFieldFormatConfiguration(t *testing.T) {
	var b bytes.Buffer

	// Test with custom time format
	cfg := logze.NewConfig(&b).WithTimeFieldFormat(time.RFC822).WithLevel("info").WithNoDiode()
	logger := cfg.New()

	logger.Info("test time format")

	output := b.String()
	if !strings.Contains(output, "test time format") {
		t.Error("expected log message")
	}
	// Note: Checking exact time format would be brittle due to timing
}

func TestDiodeConfiguration(t *testing.T) {
	var b bytes.Buffer

	// Test custom diode size
	cfg := logze.NewConfig(&b).WithDiodeSize(500).WithLevel("info")
	if cfg.DiodeSize != 500 {
		t.Errorf("expected diode size 500, got %d", cfg.DiodeSize)
	}

	// Test custom polling interval
	cfg = cfg.WithDiodePollingInterval(50 * time.Millisecond)
	if cfg.DiodePollingInterval != 50*time.Millisecond {
		t.Errorf("expected polling interval 50ms, got %v", cfg.DiodePollingInterval)
	}

	// Test custom alert function
	alertFunc := func(missed int) {
		// Alert function would be called when diode overflows
	}
	cfg = cfg.WithDiodeAlert(alertFunc)
	if cfg.DiodeAlertFunc == nil {
		t.Error("expected non-nil alert function")
	}

	// Test diode waiter
	cfg = cfg.WithDiodeWaiter()
	if !cfg.UseDiodeWaiter {
		t.Error("expected UseDiodeWaiter to be true")
	}
}

func TestHooksIntegration(t *testing.T) {
	var b bytes.Buffer
	hookCalled := false

	hook := zerolog.HookFunc(func(e *zerolog.Event, level zerolog.Level, msg string) {
		hookCalled = true
		e.Str("hook_added", "true")
	})

	cfg := logze.NewConfig(&b).WithLevel("info").WithNoDiode().WithHook(hook)
	logger := cfg.New()

	logger.Info("test hook")

	output := b.String()
	if !hookCalled {
		t.Error("expected hook to be called")
	}
	if !strings.Contains(output, "hook_added\":\"true") {
		t.Error("expected hook to add field")
	}
}

func TestMultipleHooks(t *testing.T) {
	var b bytes.Buffer
	hook1Called := false
	hook2Called := false

	hook1 := zerolog.HookFunc(func(e *zerolog.Event, level zerolog.Level, msg string) {
		hook1Called = true
		e.Str("hook1", "called")
	})

	hook2 := zerolog.HookFunc(func(e *zerolog.Event, level zerolog.Level, msg string) {
		hook2Called = true
		e.Str("hook2", "called")
	})

	cfg := logze.NewConfig(&b).WithLevel("info").WithNoDiode().WithHooks(hook1, hook2)
	logger := cfg.New()

	logger.Info("test multiple hooks")

	output := b.String()
	if !hook1Called || !hook2Called {
		t.Error("expected both hooks to be called")
	}
	if !strings.Contains(output, "hook1\":\"called") || !strings.Contains(output, "hook2\":\"called") {
		t.Error("expected both hooks to add fields")
	}
}

func TestCallerConfiguration(t *testing.T) {
	var b bytes.Buffer

	// Test WithAddCaller sets defaults
	cfg := logze.NewConfig(&b).WithAddCaller()
	if !cfg.AddCaller {
		t.Error("expected AddCaller to be true")
	}
	if cfg.CallerSkipFrameCount != logze.DefaultCallerSkipFrameCount {
		t.Errorf("expected default skip count %d, got %d", logze.DefaultCallerSkipFrameCount, cfg.CallerSkipFrameCount)
	}

	// Test WithCallerSkipFrameCount also enables caller
	cfg = logze.NewConfig().WithCallerSkipFrameCount(7)
	if !cfg.AddCaller {
		t.Error("expected AddCaller to be enabled when setting skip count")
	}
	if cfg.CallerSkipFrameCount != 7 {
		t.Errorf("expected skip count 7, got %d", cfg.CallerSkipFrameCount)
	}
}

func TestStackTraceWithActualError(t *testing.T) {
	var b bytes.Buffer
	cfg := logze.NewConfig(&b).WithLevel("error").WithNoDiode().WithStackTrace()
	logger := cfg.New()

	err := errors.New("test error")
	logger.Err(err, "error with stack trace")

	output := b.String()
	if !strings.Contains(output, "test error") {
		t.Error("expected error message")
	}
	// Stack trace should be included automatically due to WithStackTrace
	if !strings.Contains(output, "stack") {
		t.Error("expected stack trace in output")
	}
}

func TestComplexConfigurationChain(t *testing.T) {
	var b bytes.Buffer
	alertFunc := func(int) {
		// Alert function for complex configuration test
	}

	hook := zerolog.HookFunc(func(e *zerolog.Event, level zerolog.Level, msg string) {
		e.Str("chained", "config")
	})

	// Chain all configuration methods
	cfg := logze.NewConfig(&b).
		WithTrace().                                  // Set level to trace
		WithTimeFieldFormat(time.RFC822).             // Custom time format
		WithHook(hook).                               // Add hook
		WithAddCaller().                              // Enable caller
		WithStackTrace().                             // Enable stack traces
		WithToIgnore("ignore").                       // Add ignore list
		WithSimpleErrorCounter().                     // Add error counter
		WithDiodeSize(200).                           // Custom diode size
		WithDiodePollingInterval(5*time.Millisecond). // Custom interval
		WithDiodeAlert(alertFunc).                    // Custom alert
		WithPercentageSampler(1.0, "trace")           // 100% sampling for trace

	// Verify all configurations
	if cfg.Level != logze.LevelTrace {
		t.Errorf("expected trace level, got %s", cfg.Level)
	}
	if cfg.TimeFieldFormat != time.RFC822 {
		t.Errorf("expected RFC822 format, got %s", cfg.TimeFieldFormat)
	}
	if cfg.Hook == nil {
		t.Error("expected non-nil hook")
	}
	if !cfg.AddCaller {
		t.Error("expected AddCaller true")
	}
	if !cfg.StackTrace {
		t.Error("expected StackTrace true")
	}
	if len(cfg.ToIgnore) != 1 || cfg.ToIgnore[0] != "ignore" {
		t.Errorf("expected ignore list [ignore], got %v", cfg.ToIgnore)
	}
	if cfg.ErrorCounter == nil {
		t.Error("expected non-nil error counter")
	}
	if cfg.DiodeSize != 200 {
		t.Errorf("expected diode size 200, got %d", cfg.DiodeSize)
	}
	if cfg.DiodePollingInterval != 5*time.Millisecond {
		t.Errorf("expected 5ms interval, got %v", cfg.DiodePollingInterval)
	}
	if cfg.DiodeAlertFunc == nil {
		t.Error("expected non-nil alert function")
	}
	if cfg.Sampler == nil {
		t.Error("expected non-nil sampler")
	}

	// Test the configured logger works - need to disable diode for immediate output
	cfg = cfg.WithNoDiode()
	logger := cfg.New("service", "test")
	logger.Trace("trace message")

	output := b.String()
	if !strings.Contains(output, "trace message") {
		t.Errorf("expected trace message, got output: %s", output)
	}
	if !strings.Contains(output, "chained\":\"config") {
		t.Errorf("expected hook field, got output: %s", output)
	}
	if !strings.Contains(output, "service\":\"test") {
		t.Errorf("expected service field, got output: %s", output)
	}
}

func TestConsoleWriterVariants(t *testing.T) {
	// Test WithConsole adds a console writer
	cfg := logze.NewConfig().WithConsole()
	if len(cfg.Writers) != 1 {
		t.Errorf("expected 1 writer after WithConsole, got %d", len(cfg.Writers))
	}

	// Test WithConsoleNoColor adds a console writer
	cfg = logze.NewConfig().WithConsoleNoColor()
	if len(cfg.Writers) != 1 {
		t.Errorf("expected 1 writer after WithConsoleNoColor, got %d", len(cfg.Writers))
	}

	// Test WithConsoleJSON adds stderr writer
	cfg = logze.NewConfig().WithConsoleJSON()
	if len(cfg.Writers) != 1 {
		t.Errorf("expected 1 writer after WithConsoleJSON, got %d", len(cfg.Writers))
	}
	if cfg.Writers[0] != os.Stderr {
		t.Error("expected stderr writer from WithConsoleJSON")
	}
}

func TestConfigCreationMethods(t *testing.T) {
	// Test that Config.New and Config.Logger work identically
	var b1, b2 bytes.Buffer

	cfg1 := logze.NewConfig(&b1).WithLevel("info").WithNoDiode()
	cfg2 := logze.NewConfig(&b2).WithLevel("info").WithNoDiode()

	logger1 := cfg1.New("method", "New")
	logger2 := cfg2.Logger("method", "Logger")

	logger1.Info("test message")
	logger2.Info("test message")

	output1 := b1.String()
	output2 := b2.String()

	// Should have similar structure (different timestamps make exact comparison difficult)
	if !strings.Contains(output1, "test message") || !strings.Contains(output1, "method\":\"New") {
		t.Error("expected message and field from Config.New")
	}
	if !strings.Contains(output2, "test message") || !strings.Contains(output2, "method\":\"Logger") {
		t.Error("expected message and field from Config.Logger")
	}
}

func TestErrorCounterInterface(t *testing.T) {
	// Test SimpleErrorCounter implementation
	cfg := logze.NewConfig().WithSimpleErrorCounter()

	if cfg.ErrorCounter == nil {
		t.Error("expected error counter to be set")
	}

	// Test Inc method works
	testErr := errors.New("test error")
	cfg.ErrorCounter.Inc(testErr)

	simple := cfg.ErrorCounter.(*logze.SimpleErrorCounter)
	if atomic.LoadUint64(&simple.Count) != 1 {
		t.Errorf("expected count 1, got %d", atomic.LoadUint64(&simple.Count))
	}

	// Test multiple increments
	cfg.ErrorCounter.Inc(testErr)
	cfg.ErrorCounter.Inc(testErr)

	if atomic.LoadUint64(&simple.Count) != 3 {
		t.Errorf("expected count 3, got %d", atomic.LoadUint64(&simple.Count))
	}
}

func TestLevelsConstants(t *testing.T) {
	// Test that all level constants are available
	expectedLevels := []string{
		logze.LevelTrace, logze.LevelDebug, logze.LevelInfo,
		logze.LevelWarn, logze.LevelError, logze.LevelFatal, logze.LevelDisabled,
	}

	if len(logze.Levels) != len(expectedLevels) {
		t.Errorf("expected %d levels, got %d", len(expectedLevels), len(logze.Levels))
	}

	for i, expected := range expectedLevels {
		if i >= len(logze.Levels) || logze.Levels[i] != expected {
			t.Errorf("expected level %s at index %d, got %v", expected, i, logze.Levels)
		}
	}

	// Test Levelsinterface{} has same length
	if len(logze.LevelsAny) != len(logze.Levels) {
		t.Errorf("expected Levelsinterface{} to have same length as Levels")
	}
}

func TestDefaultConstants(t *testing.T) {
	// Test that default constants have reasonable values
	if logze.DefaultDiodeSize <= 0 {
		t.Errorf("expected positive diode size, got %d", logze.DefaultDiodeSize)
	}

	if logze.DefaultDiodePollingInterval <= 0 {
		t.Errorf("expected positive polling interval, got %v", logze.DefaultDiodePollingInterval)
	}

	if logze.DefaultCallerSkipFrameCount <= 0 {
		t.Errorf("expected positive caller skip count, got %d", logze.DefaultCallerSkipFrameCount)
	}
}

// Test sampler creation with warn and error levels
func TestSamplerWithWarnAndErrorLevels(t *testing.T) {
	var b bytes.Buffer

	// Test percentage sampler with warn and error levels
	cfg := logze.NewConfig(&b).WithLevel("trace").WithNoDiode().WithPercentageSampler(1.0, "warn", "error")
	logger := cfg.New()

	logger.Trace("trace message") // Should appear (not sampled)
	logger.Debug("debug message") // Should appear (not sampled)
	logger.Info("info message")   // Should appear (not sampled)
	logger.Warn("warn message")   // Should appear (100% sampling)
	logger.Error("error message") // Should appear (100% sampling)

	output := b.String()
	if !strings.Contains(output, "trace message") {
		t.Error("expected trace message (not subject to sampling)")
	}
	if !strings.Contains(output, "debug message") {
		t.Error("expected debug message (not subject to sampling)")
	}
	if !strings.Contains(output, "info message") {
		t.Error("expected info message (not subject to sampling)")
	}
	if !strings.Contains(output, "warn message") {
		t.Error("expected warn message (100% sampling)")
	}
	if !strings.Contains(output, "error message") {
		t.Error("expected error message (100% sampling)")
	}
}

func TestBurstSamplerWithWarnLevel(t *testing.T) {
	var b bytes.Buffer

	// Test burst sampler with warn level
	cfg := logze.NewConfig(&b).WithLevel("warn").WithNoDiode().WithBurstSampler(1.0, 1, time.Second, "warn")
	logger := cfg.New()

	logger.Warn("first warn")  // Should appear (within burst)
	logger.Warn("second warn") // May be sampled out depending on burst behavior

	output := b.String()
	if !strings.Contains(output, "first warn") {
		t.Error("expected first warn message within burst limit")
	}
	// Don't check second message as burst behavior may filter it
}

func TestMaxSamplerWithErrorLevel(t *testing.T) {
	var b bytes.Buffer

	// Test max sampler with error level
	cfg := logze.NewConfig(&b).WithLevel("error").WithNoDiode().WithMaxSampler(1, time.Second, "error")
	logger := cfg.New()

	logger.Error("first error") // Should appear (within max limit)

	output := b.String()
	if !strings.Contains(output, "first error") {
		t.Error("expected first error message within max limit")
	}
}

func TestSamplerWithAllLevels(t *testing.T) {
	var b bytes.Buffer

	// Test sampler applied to all levels (no specific levels provided)
	cfg := logze.NewConfig(&b).WithLevel("trace").WithNoDiode().WithPercentageSampler(1.0) // 100% for all levels
	logger := cfg.New()

	logger.Trace("trace message")
	logger.Debug("debug message")
	logger.Info("info message")
	logger.Warn("warn message")
	logger.Error("error message")

	output := b.String()
	if !strings.Contains(output, "trace message") {
		t.Error("expected trace message with global sampling")
	}
	if !strings.Contains(output, "debug message") {
		t.Error("expected debug message with global sampling")
	}
	if !strings.Contains(output, "info message") {
		t.Error("expected info message with global sampling")
	}
	if !strings.Contains(output, "warn message") {
		t.Error("expected warn message with global sampling")
	}
	if !strings.Contains(output, "error message") {
		t.Error("expected error message with global sampling")
	}
}
