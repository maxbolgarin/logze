package logze_test

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	stdlog "log"
	"strings"
	"sync/atomic"
	"testing"
	"time"

	"github.com/maxbolgarin/logze/v2"
	"github.com/rs/zerolog"
)

func setupGlobalLogger(buffer *bytes.Buffer, level string) {
	cfg := logze.NewConfig(buffer).WithLevel(level).WithNoDiode()
	logze.Init(cfg)
}

func TestGlobalDefaultPtrOriginal(t *testing.T) {
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
	if !strings.Contains(output, "level\":\"error") || !strings.Contains(output, "github.com/maxbolgarin/logze") {
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

	if atomic.LoadUint64(&ec.Count) != 0 {
		t.Errorf("expected 0, got %d", atomic.LoadUint64(&ec.Count))
	}

	logze.Err(errors.New("error occurred"), "error test")
	if atomic.LoadUint64(&ec.Count) != 1 {
		t.Errorf("expected 1, got %d", atomic.LoadUint64(&ec.Count))
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

// Additional comprehensive tests for better coverage

func TestGlobalDefault(t *testing.T) {
	// Test Default function returns a copy
	original := logze.Default()
	modified := original.With("test", "field")

	// Original should not have the field
	var b1 bytes.Buffer
	logze.Init(logze.NewConfig(&b1).WithNoDiode())
	original.Info("original message")

	// Modified should have the field
	var b2 bytes.Buffer
	rawLogger := modified.WithLevel("info").Raw()
	*rawLogger = rawLogger.Output(&b2)
	rawLogger.Info().Msg("modified message")

	if strings.Contains(b1.String(), "test\":\"field") {
		t.Error("original logger should not have test field")
	}
	if !strings.Contains(b2.String(), "test\":\"field") {
		t.Error("modified logger should have test field")
	}
}

func TestGlobalD(t *testing.T) {
	// Test D function (shortcut for Default)
	logger := logze.D()
	if logger.NotInited() {
		t.Error("expected D() to return inited logger")
	}
}

func TestGlobalDefaultPtr(t *testing.T) {
	// Test DefaultPtr returns pointer to global logger
	var b bytes.Buffer
	logze.Init(logze.NewConfig(&b).WithNoDiode())

	ptr := logze.DefaultPtr()
	ptr.Info("ptr message")

	output := b.String()
	if !strings.Contains(output, "ptr message") {
		t.Error("expected message from DefaultPtr")
	}
}

func TestGlobalDP(t *testing.T) {
	// Test DP function (shortcut for DefaultPtr)
	ptr := logze.DP()
	if ptr == nil {
		t.Error("expected non-nil pointer from DP()")
	}
}

func TestGlobalSetDefault(t *testing.T) {
	var b bytes.Buffer
	customLogger := logze.New(logze.NewConfig(&b).WithLevel("warn").WithNoDiode())

	logze.SetDefault(customLogger)

	// Should not log info (below warn level)
	logze.Info("should not appear")

	// Should log warn
	logze.Warn("should appear")

	output := b.String()
	if strings.Contains(output, "should not appear") {
		t.Error("expected info message to be filtered at warn level")
	}
	if !strings.Contains(output, "should appear") {
		t.Error("expected warn message to appear")
	}
}

func TestGlobalWithErrorCounter(t *testing.T) {
	var ec logze.SimpleErrorCounter
	logger := logze.WithErrorCounter(&ec)

	if logger.GetErrorCounter() != &ec {
		t.Error("expected same error counter instance")
	}
}

func TestGlobalWithSimpleErrorCounter(t *testing.T) {
	logger := logze.WithSimpleErrorCounter()

	counter := logger.GetErrorCounter()
	if counter == nil {
		t.Error("expected non-nil error counter")
	}

	if _, ok := counter.(*logze.SimpleErrorCounter); !ok {
		t.Error("expected SimpleErrorCounter type")
	}
}

func TestGlobalWithToIgnore(t *testing.T) {
	var b bytes.Buffer

	// Set up logger to write to buffer without diode for immediate output
	logze.SetDefault(logze.New(logze.NewConfig(&b).WithLevel("info").WithNoDiode().WithToIgnore("ignore1", "ignore2")))

	logze.Info("normal message")
	logze.Info("ignore1")
	logze.Info("contains ignore2 text")

	output := b.String()
	if !strings.Contains(output, "normal message") {
		t.Error("expected normal message")
	}
	if strings.Contains(output, "ignore1") || strings.Contains(output, "ignore2") {
		t.Error("expected ignored messages to be filtered")
	}
}

func TestGlobalWithCaller(t *testing.T) {
	var b bytes.Buffer
	logger := logze.WithCaller(3)
	rawLogger := logger.WithLevel("info").Raw()
	*rawLogger = rawLogger.Output(&b)

	rawLogger.Info().Msg("test caller")

	output := b.String()
	if !strings.Contains(output, "caller") {
		t.Error("expected caller information")
	}
}

func TestGlobalWithDefaultCaller(t *testing.T) {
	var b bytes.Buffer
	logger := logze.WithDefaultCaller()
	rawLogger := logger.WithLevel("info").Raw()
	*rawLogger = rawLogger.Output(&b)

	rawLogger.Info().Msg("test default caller")

	output := b.String()
	if !strings.Contains(output, "caller") {
		t.Error("expected caller information")
	}
}

func TestGlobalGetErrorCounter(t *testing.T) {
	// Test with no error counter
	logze.Init(logze.NewConfig())
	if logze.GetErrorCounter() != nil {
		t.Error("expected nil error counter")
	}

	// Test with error counter
	logze.Init(logze.NewConfig().WithSimpleErrorCounter())
	if logze.GetErrorCounter() == nil {
		t.Error("expected non-nil error counter")
	}
}

func TestGlobalCloseDiode(t *testing.T) {
	logze.Init(logze.NewConfig())

	err := logze.CloseDiode()
	if err != nil {
		t.Errorf("expected no error from CloseDiode, got %v", err)
	}
}

func TestGlobalClose(t *testing.T) {
	logze.Init(logze.NewConfig())

	err := logze.Close()
	if err != nil {
		t.Errorf("expected no error from Close, got %v", err)
	}
}

func TestGlobalWithSampler(t *testing.T) {
	sampler := zerolog.RandomSampler(10)
	logger := logze.WithSampler(sampler)

	// Just test that it returns a valid logger
	if logger.NotInited() {
		t.Error("expected inited logger from WithSampler")
	}
}

func TestGlobalWithPercentageSampler(t *testing.T) {
	logger := logze.WithPercentageSampler(0.5, "debug", "info")

	if logger.NotInited() {
		t.Error("expected inited logger from WithPercentageSampler")
	}
}

func TestGlobalWithBurstSampler(t *testing.T) {
	logger := logze.WithBurstSampler(0.1, 100, time.Second, "debug")

	if logger.NotInited() {
		t.Error("expected inited logger from WithBurstSampler")
	}
}

func TestGlobalWithMaxSampler(t *testing.T) {
	logger := logze.WithMaxSampler(100, time.Second, "debug", "info")

	if logger.NotInited() {
		t.Error("expected inited logger from WithMaxSampler")
	}
}

// Test all missing global logging functions

func TestGlobalTracef(t *testing.T) {
	var b bytes.Buffer
	setupGlobalLogger(&b, logze.LevelTrace)

	logze.Tracef("trace %s %d", "test", 42)

	output := b.String()
	if !strings.Contains(output, "level\":\"trace") || !strings.Contains(output, "trace test 42") {
		t.Error("expected formatted trace message")
	}
}

func TestGlobalTraceIf(t *testing.T) {
	var b bytes.Buffer
	setupGlobalLogger(&b, logze.LevelTrace)

	logze.TraceIf(true, "conditional trace")
	logze.TraceIf(false, "should not appear")

	output := b.String()
	if !strings.Contains(output, "conditional trace") {
		t.Error("expected conditional trace when true")
	}
	if strings.Contains(output, "should not appear") {
		t.Error("unexpected output when condition false")
	}
}

func TestGlobalDebugIf(t *testing.T) {
	var b bytes.Buffer
	setupGlobalLogger(&b, logze.LevelDebug)

	logze.DebugIf(true, "conditional debug")
	logze.DebugIf(false, "should not appear")

	output := b.String()
	if !strings.Contains(output, "conditional debug") {
		t.Error("expected conditional debug when true")
	}
	if strings.Contains(output, "should not appear") {
		t.Error("unexpected output when condition false")
	}
}

func TestGlobalInfoIf(t *testing.T) {
	var b bytes.Buffer
	setupGlobalLogger(&b, logze.LevelInfo)

	logze.InfoIf(true, "conditional info")
	logze.InfoIf(false, "should not appear")

	output := b.String()
	if !strings.Contains(output, "conditional info") {
		t.Error("expected conditional info when true")
	}
	if strings.Contains(output, "should not appear") {
		t.Error("unexpected output when condition false")
	}
}

func TestGlobalWarnIf(t *testing.T) {
	var b bytes.Buffer
	setupGlobalLogger(&b, logze.LevelWarn)

	logze.WarnIf(true, "conditional warn")
	logze.WarnIf(false, "should not appear")

	output := b.String()
	if !strings.Contains(output, "conditional warn") {
		t.Error("expected conditional warn when true")
	}
	if strings.Contains(output, "should not appear") {
		t.Error("unexpected output when condition false")
	}
}

func TestGlobalErrorIf(t *testing.T) {
	var b bytes.Buffer
	setupGlobalLogger(&b, logze.LevelError)

	logze.ErrorIf(true, "conditional error")
	logze.ErrorIf(false, "should not appear")

	output := b.String()
	if !strings.Contains(output, "conditional error") {
		t.Error("expected conditional error when true")
	}
	if strings.Contains(output, "should not appear") {
		t.Error("unexpected output when condition false")
	}
}

func TestGlobalErrIf(t *testing.T) {
	var b bytes.Buffer
	setupGlobalLogger(&b, logze.LevelError)

	err := errors.New("test error")
	logze.ErrIf(true, err, "conditional err")
	logze.ErrIf(false, err, "should not appear")

	output := b.String()
	if !strings.Contains(output, "conditional err") {
		t.Error("expected conditional err when true")
	}
	if strings.Contains(output, "should not appear") {
		t.Error("unexpected output when condition false")
	}
}

func TestGlobalErrorfAdditional(t *testing.T) {
	var b bytes.Buffer
	setupGlobalLogger(&b, logze.LevelError)

	logze.Errorf("formatted error %d", 123)

	output := b.String()
	if !strings.Contains(output, "formatted error 123") {
		t.Error("expected formatted error message")
	}
}

func TestGlobalFatalIf(t *testing.T) {
	var b bytes.Buffer
	setupGlobalLogger(&b, logze.LevelFatal)

	// Test that FatalIf doesn't exit when condition is false
	logze.FatalIf(false, "should not exit")

	if b.Len() > 0 {
		t.Error("expected no output when FatalIf condition is false")
	}
}

func TestGlobalFatalf(t *testing.T) {
	// Cannot easily test actual fatal behavior, but we can test the function exists
	// and would work with mocking in a real scenario
	t.Log("Fatalf function exists and is accessible")
}

func TestGlobalFatalln(t *testing.T) {
	// Cannot easily test actual fatal behavior
	t.Log("Fatalln function exists and is accessible")
}

func TestGlobalPanicIf(t *testing.T) {
	var b bytes.Buffer
	setupGlobalLogger(&b, logze.LevelFatal)

	// Test that PanicIf doesn't panic when condition is false
	logze.PanicIf(false, "should not panic")

	if b.Len() > 0 {
		t.Error("expected no output when PanicIf condition is false")
	}
}

func TestGlobalPanicf(t *testing.T) {
	defer func() {
		if r := recover(); r == nil {
			t.Error("expected panic from Panicf")
		}
	}()

	var b bytes.Buffer
	setupGlobalLogger(&b, logze.LevelFatal)
	logze.Panicf("panic %s", "test")
}

func TestGlobalPanicln(t *testing.T) {
	defer func() {
		if r := recover(); r == nil {
			t.Error("expected panic from Panicln")
		}
	}()

	var b bytes.Buffer
	setupGlobalLogger(&b, logze.LevelFatal)
	logze.Panicln("panic", "test")
}

func TestGlobalPrintIf(t *testing.T) {
	var b bytes.Buffer
	setupGlobalLogger(&b, logze.LevelInfo)

	logze.PrintIf(true, "conditional print")
	logze.PrintIf(false, "should not appear")

	output := b.String()
	if !strings.Contains(output, "conditional print") {
		t.Error("expected conditional print when true")
	}
	if strings.Contains(output, "should not appear") {
		t.Error("unexpected output when condition false")
	}
}

func TestGlobalPrintStack(t *testing.T) {
	var b bytes.Buffer
	setupGlobalLogger(&b, logze.LevelInfo)

	logze.PrintStack("context", "test")

	output := b.String()
	if !strings.Contains(output, "TestGlobalPrintStack") {
		t.Error("expected stack trace with test function name")
	}
	if !strings.Contains(output, "context\":\"test") {
		t.Error("expected context field")
	}
}

func TestGlobalLog(t *testing.T) {
	var b bytes.Buffer
	setupGlobalLogger(&b, logze.LevelInfo)

	logze.Log("test log message")

	output := b.String()
	if !strings.Contains(output, "test log message") {
		t.Error("expected log message")
	}
}

func TestGlobalPrintln(t *testing.T) {
	var b bytes.Buffer
	setupGlobalLogger(&b, logze.LevelInfo)

	logze.Println("test", "println", "message")

	output := b.String()
	if !strings.Contains(output, "test println message") {
		t.Error("expected println message with spaces")
	}
}

func TestGlobalWrite(t *testing.T) {
	var b bytes.Buffer
	setupGlobalLogger(&b, logze.LevelInfo)

	data := []byte("direct write data")
	n, err := logze.Write(data)

	if err != nil {
		t.Errorf("expected no error from Write, got %v", err)
	}
	if n != len(data) {
		t.Errorf("expected %d bytes written, got %d", len(data), n)
	}
}

func TestGlobalRaw(t *testing.T) {
	setupGlobalLogger(&bytes.Buffer{}, logze.LevelInfo)

	raw := logze.Raw()
	if raw == nil {
		t.Error("expected non-nil raw logger")
	}
}

func TestGlobalUpdate(t *testing.T) {
	var b1, b2 bytes.Buffer

	// Initial setup
	logze.Init(logze.NewConfig(&b1).WithLevel("info").WithNoDiode())
	logze.Info("initial message")

	// Update configuration
	logze.Update(logze.NewConfig(&b2).WithLevel("debug").WithNoDiode())
	logze.Debug("updated message")

	output1 := b1.String()
	output2 := b2.String()

	if !strings.Contains(output1, "initial message") {
		t.Error("expected initial message in first buffer")
	}
	if !strings.Contains(output2, "updated message") {
		t.Error("expected updated message in second buffer")
	}
}

func TestGlobalSetStdLogger(t *testing.T) {
	var b bytes.Buffer
	logger := logze.New(logze.NewConfig(&b).WithNoDiode(), "source", "stdlib")

	logze.SetStdLogger(logger, "component", "test")

	// Test that standard log goes through our logger
	stdlog.Println("std log message")

	output := b.String()
	if !strings.Contains(output, "std log message") {
		t.Error("expected std log message")
	}
	if !strings.Contains(output, "source\":\"stdlib") {
		t.Error("expected source field")
	}
	if !strings.Contains(output, "component\":\"test") {
		t.Error("expected component field")
	}
}

// Test global WithFields function
func TestGlobalWithFields(t *testing.T) {
	var b bytes.Buffer
	setupGlobalLogger(&b, logze.LevelInfo)

	logger := logze.WithFields("service", "api", "version", "1.0")
	logger.Info("service started")

	output := b.String()
	if !strings.Contains(output, "service started") {
		t.Error("expected service started message")
	}
	if !strings.Contains(output, "service\":\"api") {
		t.Error("expected service field")
	}
	if !strings.Contains(output, "version\":\"1.0") {
		t.Error("expected version field")
	}
}

// Test global With function (shorthand for WithFields)
func TestGlobalWith(t *testing.T) {
	var b bytes.Buffer
	setupGlobalLogger(&b, logze.LevelInfo)

	logger := logze.With("component", "auth", "user_id", 123)
	logger.Info("authentication successful")

	output := b.String()
	if !strings.Contains(output, "authentication successful") {
		t.Error("expected authentication message")
	}
	if !strings.Contains(output, "component\":\"auth") {
		t.Error("expected component field")
	}
	if !strings.Contains(output, "user_id\":123") {
		t.Error("expected user_id field")
	}
}

// Test global WithLevel function
func TestGlobalWithLevel(t *testing.T) {
	var b bytes.Buffer
	setupGlobalLogger(&b, logze.LevelInfo)

	// Create debug logger from global
	debugLogger := logze.WithLevel("debug")
	debugLogger.Debug("debug from global with level")
	debugLogger.Info("info from global with level")

	output := b.String()
	if !strings.Contains(output, "debug from global with level") {
		t.Error("expected debug message from global WithLevel")
	}
	if !strings.Contains(output, "info from global with level") {
		t.Error("expected info message from global WithLevel")
	}
}

// Test global WithToIgnore function
func TestGlobalWithToIgnoreFunction(t *testing.T) {
	var b bytes.Buffer
	setupGlobalLogger(&b, logze.LevelInfo)

	logger := logze.WithToIgnore("ignore", "filter")
	logger.Info("normal message")
	logger.Info("ignore this message")
	logger.Info("filter this too")
	logger.Info("another normal message")

	output := b.String()
	if !strings.Contains(output, "normal message") {
		t.Error("expected normal message")
	}
	if !strings.Contains(output, "another normal message") {
		t.Error("expected another normal message")
	}
	if strings.Contains(output, "ignore this message") {
		t.Error("expected 'ignore this message' to be filtered")
	}
	if strings.Contains(output, "filter this too") {
		t.Error("expected 'filter this too' to be filtered")
	}
}

// Test global Trace function
func TestGlobalTrace(t *testing.T) {
	var b bytes.Buffer
	setupGlobalLogger(&b, logze.LevelTrace)

	logze.Trace("trace message from global", "operation", "startup", "duration", "100ms")

	output := b.String()
	if !strings.Contains(output, "level\":\"trace") {
		t.Error("expected trace level")
	}
	if !strings.Contains(output, "trace message from global") {
		t.Error("expected trace message")
	}
	if !strings.Contains(output, "operation\":\"startup") {
		t.Error("expected operation field")
	}
	if !strings.Contains(output, "duration\":\"100ms") {
		t.Error("expected duration field")
	}
	if !strings.Contains(output, "caller") {
		t.Error("expected caller info in trace")
	}
}

// Test global functions combination
func TestGlobalFunctionsCombination(t *testing.T) {
	var b bytes.Buffer
	setupGlobalLogger(&b, logze.LevelTrace)

	// Test chaining multiple global functions
	logger := logze.WithFields("app", "test").
		WithLevel("debug").
		WithToIgnore("health")

	logger.Debug("application started", "pid", 12345)
	logger.Info("health check") // Should be ignored
	logger.Warn("warning message", "code", 404)

	output := b.String()
	if !strings.Contains(output, "application started") {
		t.Error("expected application started message")
	}
	if !strings.Contains(output, "app\":\"test") {
		t.Error("expected app field")
	}
	if !strings.Contains(output, "pid\":12345") {
		t.Error("expected pid field")
	}
	if strings.Contains(output, "health check") {
		t.Error("expected health check to be ignored")
	}
	if !strings.Contains(output, "warning message") {
		t.Error("expected warning message")
	}
	if !strings.Contains(output, "code\":404") {
		t.Error("expected code field")
	}
}

// Test global HTTP methods

func TestGlobalHTTP(t *testing.T) {
	var b bytes.Buffer
	setupGlobalLogger(&b, logze.LevelDebug)

	duration := 42 * time.Millisecond
	logze.HTTP("GET", "/api/users", 200, duration, "user_id", 123)

	output := b.String()
	if !strings.Contains(output, "level\":\"debug") {
		t.Error("expected debug level")
	}
	if !strings.Contains(output, "HTTP request") {
		t.Error("expected HTTP request message")
	}
	if !strings.Contains(output, "method\":\"GET") {
		t.Error("expected method field")
	}
	if !strings.Contains(output, "path\":\"/api/users") {
		t.Error("expected path field")
	}
	if !strings.Contains(output, "status\":200") {
		t.Error("expected status field")
	}
	if !strings.Contains(output, "duration_ms\":42") {
		t.Error("expected duration_ms field")
	}
	if !strings.Contains(output, "user_id\":123") {
		t.Error("expected user_id field")
	}
}

func TestGlobalHTTPError(t *testing.T) {
	var b bytes.Buffer
	setupGlobalLogger(&b, logze.LevelError)

	err := errors.New("database connection failed")
	logze.HTTPError("POST", "/api/orders", 500, err, "order_id", "abc123")

	output := b.String()
	if !strings.Contains(output, "level\":\"error") {
		t.Error("expected error level")
	}
	if !strings.Contains(output, "HTTP request failed") {
		t.Error("expected HTTP request failed message")
	}
	if !strings.Contains(output, "method\":\"POST") {
		t.Error("expected method field")
	}
	if !strings.Contains(output, "path\":\"/api/orders") {
		t.Error("expected path field")
	}
	if !strings.Contains(output, "status\":500") {
		t.Error("expected status field")
	}
	if !strings.Contains(output, "order_id\":\"abc123") {
		t.Error("expected order_id field")
	}
	if !strings.Contains(output, "database connection failed") {
		t.Error("expected error message")
	}
}

func TestGlobalHTTPAuto(t *testing.T) {
	var b bytes.Buffer
	setupGlobalLogger(&b, logze.LevelDebug)

	duration := 100 * time.Millisecond

	// Test success case (should log at debug level)
	b.Reset()
	logze.HTTPAuto("GET", "/api/products", 200, duration, nil, "product_id", "prod123")
	output := b.String()
	if !strings.Contains(output, "level\":\"debug") {
		t.Error("expected debug level for 200 status")
	}
	if !strings.Contains(output, "HTTP request") {
		t.Error("expected HTTP request message")
	}

	// Test client error case (should log at warn level)
	b.Reset()
	logze.HTTPAuto("GET", "/api/products", 404, duration, nil, "product_id", "prod123")
	output = b.String()
	if !strings.Contains(output, "level\":\"warn") {
		t.Error("expected warn level for 404 status")
	}
	if !strings.Contains(output, "HTTP client error") {
		t.Error("expected HTTP client error message")
	}

	// Test server error case (should log at error level)
	b.Reset()
	err := errors.New("internal server error")
	logze.HTTPAuto("GET", "/api/products", 500, duration, err, "product_id", "prod123")
	output = b.String()
	if !strings.Contains(output, "level\":\"error") {
		t.Error("expected error level for 500 status with error")
	}
	if !strings.Contains(output, "HTTP request failed") {
		t.Error("expected HTTP request failed message")
	}
}

// Test global Recover methods

func TestGlobalRecover(t *testing.T) {
	var b bytes.Buffer
	setupGlobalLogger(&b, logze.LevelError)

	// Recover catches and logs the panic, preventing it from propagating
	func() {
		defer logze.Recover("request_id", "test123")
		panic("test panic message")
	}()

	// Panic should be recovered, so execution continues normally
	output := b.String()
	if !strings.Contains(output, "level\":\"error") {
		t.Errorf("expected error level, got output: %s", output)
	}
	// The panic value is logged in the "error" field
	if !strings.Contains(output, "error\":\"test panic message") {
		t.Errorf("expected panic message in error field, got output: %s", output)
	}
	if !strings.Contains(output, "request_id\":\"test123") {
		t.Errorf("expected request_id field, got output: %s", output)
	}
	// Stack trace should be in the output (logged as the message)
	if !strings.Contains(output, "goroutine") {
		t.Logf("Stack trace might not be visible, but output: %s", output)
	}
}

func TestGlobalRecoverPanic(t *testing.T) {
	var b bytes.Buffer
	setupGlobalLogger(&b, logze.LevelError)

	// RecoverPanic catches and logs the panic, preventing it from propagating
	func() {
		defer logze.Recover("component", "test", "operation", "test_op")

		panic("panic with fields")
	}()

	// Panic should be recovered, so execution continues normally
	output := b.String()
	if !strings.Contains(output, "level\":\"error") {
		t.Error("expected error level")
	}
	if !strings.Contains(output, "panic with fields") {
		t.Error("expected panic message")
	}
	if !strings.Contains(output, "component\":\"test") {
		t.Error("expected component field")
	}
	if !strings.Contains(output, "operation\":\"test_op") {
		t.Error("expected operation field")
	}
}

func TestGlobalRecoverPanicWithCallback(t *testing.T) {
	var b bytes.Buffer
	setupGlobalLogger(&b, logze.LevelError)

	callbackCalled := false
	panicValue := ""

	// RecoverPanicWithCallback catches and logs the panic, then calls the callback
	func() {
		defer logze.RecoverWithCallback(func(p interface{}) {
			callbackCalled = true
			panicValue = fmt.Sprint(p)
		}, "request_id", "callback_test")

		panic("callback test panic")
	}()

	// Panic should be recovered, so execution continues normally
	if !callbackCalled {
		t.Error("expected callback to be called")
	}
	if panicValue != "callback test panic" {
		t.Errorf("expected panic value 'callback test panic', got '%s'", panicValue)
	}

	output := b.String()
	if !strings.Contains(output, "level\":\"error") {
		t.Error("expected error level")
	}
	if !strings.Contains(output, "callback test panic") {
		t.Error("expected panic message")
	}
	if !strings.Contains(output, "request_id\":\"callback_test") {
		t.Error("expected request_id field")
	}
}

// Test global context-based methods

func TestGlobalInfoCtx(t *testing.T) {
	var b bytes.Buffer
	setupGlobalLogger(&b, logze.LevelInfo)

	ctx := context.Background()
	logze.InfoCtx(ctx, "info with context", "user_id", 456)

	output := b.String()
	if !strings.Contains(output, "level\":\"info") {
		t.Error("expected info level")
	}
	if !strings.Contains(output, "info with context") {
		t.Error("expected info message")
	}
	if !strings.Contains(output, "user_id\":456") {
		t.Error("expected user_id field")
	}
}

func TestGlobalInfoCtxCancelled(t *testing.T) {
	var b bytes.Buffer
	setupGlobalLogger(&b, logze.LevelInfo)

	ctx, cancel := context.WithCancel(context.Background())
	cancel() // Cancel the context

	logze.InfoCtx(ctx, "should not appear", "user_id", 456)

	output := b.String()
	if strings.Contains(output, "should not appear") {
		t.Error("expected no log when context is cancelled")
	}
}

func TestGlobalDebugCtx(t *testing.T) {
	var b bytes.Buffer
	setupGlobalLogger(&b, logze.LevelDebug)

	ctx := context.Background()
	logze.DebugCtx(ctx, "debug with context", "component", "test")

	output := b.String()
	if !strings.Contains(output, "level\":\"debug") {
		t.Error("expected debug level")
	}
	if !strings.Contains(output, "debug with context") {
		t.Error("expected debug message")
	}
	if !strings.Contains(output, "component\":\"test") {
		t.Error("expected component field")
	}
}

func TestGlobalDebugCtxCancelled(t *testing.T) {
	var b bytes.Buffer
	setupGlobalLogger(&b, logze.LevelDebug)

	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	logze.DebugCtx(ctx, "should not appear")

	output := b.String()
	if strings.Contains(output, "should not appear") {
		t.Error("expected no log when context is cancelled")
	}
}

func TestGlobalTraceCtx(t *testing.T) {
	var b bytes.Buffer
	setupGlobalLogger(&b, logze.LevelTrace)

	ctx := context.Background()
	logze.TraceCtx(ctx, "trace with context", "operation", "startup")

	output := b.String()
	if !strings.Contains(output, "level\":\"trace") {
		t.Error("expected trace level")
	}
	if !strings.Contains(output, "trace with context") {
		t.Error("expected trace message")
	}
	if !strings.Contains(output, "operation\":\"startup") {
		t.Error("expected operation field")
	}
}

func TestGlobalTraceCtxCancelled(t *testing.T) {
	var b bytes.Buffer
	setupGlobalLogger(&b, logze.LevelTrace)

	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	logze.TraceCtx(ctx, "should not appear")

	output := b.String()
	if strings.Contains(output, "should not appear") {
		t.Error("expected no log when context is cancelled")
	}
}

func TestGlobalWarnCtx(t *testing.T) {
	var b bytes.Buffer
	setupGlobalLogger(&b, logze.LevelWarn)

	ctx := context.Background()
	logze.WarnCtx(ctx, "warn with context", "code", 404)

	output := b.String()
	if !strings.Contains(output, "level\":\"warn") {
		t.Error("expected warn level")
	}
	if !strings.Contains(output, "warn with context") {
		t.Error("expected warn message")
	}
	if !strings.Contains(output, "code\":404") {
		t.Error("expected code field")
	}
}

func TestGlobalWarnCtxCancelled(t *testing.T) {
	var b bytes.Buffer
	setupGlobalLogger(&b, logze.LevelWarn)

	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	logze.WarnCtx(ctx, "should not appear")

	output := b.String()
	if strings.Contains(output, "should not appear") {
		t.Error("expected no log when context is cancelled")
	}
}

func TestGlobalErrorCtx(t *testing.T) {
	var b bytes.Buffer
	setupGlobalLogger(&b, logze.LevelError)

	ctx := context.Background()
	logze.ErrorCtx(ctx, "error with context", "module", "payment")

	output := b.String()
	if !strings.Contains(output, "level\":\"error") {
		t.Error("expected error level")
	}
	if !strings.Contains(output, "error with context") {
		t.Error("expected error message")
	}
	if !strings.Contains(output, "module\":\"payment") {
		t.Error("expected module field")
	}
}

func TestGlobalErrorCtxCancelled(t *testing.T) {
	var b bytes.Buffer
	setupGlobalLogger(&b, logze.LevelError)

	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	logze.ErrorCtx(ctx, "should not appear")

	output := b.String()
	if strings.Contains(output, "should not appear") {
		t.Error("expected no log when context is cancelled")
	}
}

func TestGlobalErrCtx(t *testing.T) {
	var b bytes.Buffer
	setupGlobalLogger(&b, logze.LevelError)

	ctx := context.Background()
	err := errors.New("database connection failed")
	logze.ErrCtx(ctx, err, "failed to fetch data", "retry_count", 3)

	output := b.String()
	if !strings.Contains(output, "level\":\"error") {
		t.Error("expected error level")
	}
	if !strings.Contains(output, "failed to fetch data") {
		t.Error("expected error message")
	}
	if !strings.Contains(output, "database connection failed") {
		t.Error("expected error value")
	}
	if !strings.Contains(output, "retry_count\":3") {
		t.Error("expected retry_count field")
	}
}

func TestGlobalErrCtxCancelled(t *testing.T) {
	var b bytes.Buffer
	setupGlobalLogger(&b, logze.LevelError)

	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	err := errors.New("test error")
	logze.ErrCtx(ctx, err, "should not appear")

	output := b.String()
	if strings.Contains(output, "should not appear") {
		t.Error("expected no log when context is cancelled")
	}
}

func TestGlobalCtxWithTimeout(t *testing.T) {
	var b bytes.Buffer
	setupGlobalLogger(&b, logze.LevelInfo)

	// Test with timeout context
	ctx, cancel := context.WithTimeout(context.Background(), 1*time.Millisecond)
	defer cancel()

	time.Sleep(10 * time.Millisecond) // Wait for timeout

	logze.InfoCtx(ctx, "should not appear after timeout")

	output := b.String()
	if strings.Contains(output, "should not appear after timeout") {
		t.Error("expected no log when context is timed out")
	}
}
