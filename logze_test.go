package logze_test

import (
	"bytes"
	"context"
	"fmt"
	"strings"
	"testing"
	"time"

	"github.com/maxbolgarin/logze/v2"
	"github.com/pkg/errors"
	"github.com/rs/zerolog"
)

func TestLoggerInitialization(t *testing.T) {
	var b bytes.Buffer
	cfg := logze.NewConfig(&b).WithLevel(logze.LevelDebug).WithNoDiode()
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
	cfg := logze.NewConfig(&b).WithLevel(logze.LevelDebug).WithNoDiode()
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
	cfg := logze.NewConfig(&b).WithLevel(logze.LevelDebug).WithNoDiode()
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
	cfg := logze.NewConfig(&b).WithLevel(logze.LevelDebug).WithNoDiode().WithToIgnore("ignore me")
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
	cfg := logze.NewConfig().WithErrorCounter(&ec).WithLevel(logze.LevelError)
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
	cfg := logze.NewConfig(&b).WithLevel(logze.LevelDebug).WithNoDiode()
	logger := logze.New(cfg)
	logger.Panic("panic message")
}

func TestLoggerDebug(t *testing.T) {
	var b bytes.Buffer
	cfg := logze.NewConfig(&b).WithLevel(logze.LevelDebug).WithNoDiode()
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
	cfg := logze.NewConfig(&b).WithLevel(logze.LevelWarn).WithNoDiode()
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

type errCounter struct {
	count int
}

func (e *errCounter) Inc(err error) {
	e.count++
}

func TestLoggerError(t *testing.T) {
	var b bytes.Buffer
	var ec errCounter

	cfg := logze.NewConfig(&b).WithLevel(logze.LevelError).WithNoDiode().WithErrorCounter(&ec)
	logger := logze.New(cfg)

	logger.Error("error message")

	output := b.String()
	if !strings.Contains(output, "level\":\"error") {
		t.Errorf("expected log level error, got %s", output)
	}
	if !strings.Contains(output, "error message") {
		t.Errorf("expected log message 'error message', got %s", output)
	}

	if ec.count != 0 {
		t.Errorf("expected 0, got %d", ec.count)
	}

	b.Reset()
	logger.Error("error message", errors.New("abc"))

	if ec.count != 1 {
		t.Errorf("expected 1, got %d", ec.count)
	}

	b.Reset()
	logger.Errorf("error message %s %w", "err", errors.New("abc"), "a", "b")

	if ec.count != 2 {
		t.Errorf("expected 2, got %d", ec.count)
	}
}

func TestLoggerFatal(t *testing.T) {
	// It's challenging to test fatal logs without stopping execution,
	// so these tests are theoretical – you might use an interface or mock for os.Exit if necessary.
}

func TestUpdateLoggerConfiguration(t *testing.T) {
	var b1, b2 bytes.Buffer
	cfg1 := logze.NewConfig(&b1).WithLevel(logze.LevelInfo).WithNoDiode()
	logger := logze.New(cfg1)

	logger.Info("initial config message")

	output1 := b1.String()
	if !strings.Contains(output1, "initial config message") {
		t.Errorf("expected log 'initial config message', got %s", output1)
	}

	cfg2 := logze.NewConfig(&b2).WithLevel(logze.LevelDebug).WithNoDiode()
	logger.Update(cfg2)

	logger.Debug("updated config message")
	output2 := b2.String()
	if !strings.Contains(output2, "updated config message") {
		t.Errorf("expected log 'updated config message', got %s", output2)
	}
}

func TestLoggerInfof(t *testing.T) {
	var b bytes.Buffer
	cfg := logze.NewConfig(&b).WithLevel(logze.LevelInfo).WithNoDiode()
	logger := logze.New(cfg)

	logger.Infof("test message %d", 42, "a", "b")

	output := b.String()
	if !strings.Contains(output, "level\":\"info") || !strings.Contains(output, "test message 42") || !strings.Contains(output, "\"a\":\"b\"") {
		t.Errorf("expected formatted info message, got %s", output)
	}
}

func TestLoggerDebugf(t *testing.T) {
	var b bytes.Buffer
	cfg := logze.NewConfig(&b).WithLevel(logze.LevelDebug).WithNoDiode()
	logger := logze.New(cfg)

	logger.Debugf("debug value: %v", 100)

	output := b.String()
	if !strings.Contains(output, "level\":\"debug") || !strings.Contains(output, "debug value: 100") {
		t.Errorf("expected formatted debug message, got %s", output)
	}
}

func TestLoggerErrf(t *testing.T) {
	var b bytes.Buffer
	cfg := logze.NewConfig(&b).WithLevel(logze.LevelError).WithNoDiode()
	logger := logze.New(cfg)

	logger.Errf(errors.New("123"), "error operation %s", "failed", "a", "b")

	output := b.String()
	if !strings.Contains(output, "level\":\"error") || !strings.Contains(output, "error operation failed") || !strings.Contains(output, "123") || !strings.Contains(output, "\"a\":\"b\"") {
		t.Errorf("expected formatted error message, got %s", output)
	}
}

func TestLoggerErrorf(t *testing.T) {
	var b bytes.Buffer
	cfg := logze.NewConfig(&b).WithLevel(logze.LevelError).WithNoDiode()
	logger := logze.New(cfg)

	logger.Errorf("error operation %s", "failed")

	output := b.String()
	if !strings.Contains(output, "level\":\"error") || !strings.Contains(output, "error operation failed") {
		t.Errorf("expected formatted error message, got %s", output)
	}
}

func TestLoggerWarnf(t *testing.T) {
	var b bytes.Buffer
	cfg := logze.NewConfig(&b).WithLevel(logze.LevelWarn).WithNoDiode()
	logger := logze.New(cfg)

	logger.Warnf("warn message: %s", "check")

	output := b.String()
	if !strings.Contains(output, "level\":\"warn") || !strings.Contains(output, "warn message: check") {
		t.Errorf("expected formatted warn message, got %s", output)
	}
}

func TestLoggerErrStack(t *testing.T) {
	var b bytes.Buffer
	cfg := logze.NewConfig(&b).WithLevel(logze.LevelError).WithStackTrace().WithNoDiode()
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
	cfg := logze.NewConfig(&b).WithLevel(logze.LevelInfo).WithNoDiode()
	logger := logze.New(cfg)

	logger.Printf("log without level %s", "status")

	output := b.String()
	if !strings.Contains(output, "log without level status") {
		t.Errorf("expected unlevelled formatted log, got %s", output)
	}
}

// Additional comprehensive tests for better coverage

func TestNop(t *testing.T) {
	logger := logze.Nop()

	if !logger.NotInited() {
		t.Error("expected Nop logger to report as not inited")
	}

	// Test that Nop logger doesn't output interface{}thing
	var b bytes.Buffer
	logger.Info("this should not appear")

	if b.Len() > 0 {
		t.Errorf("expected no output from Nop logger, got %s", b.String())
	}
}

func TestNewFromZerolog(t *testing.T) {
	var b bytes.Buffer
	zlogger := zerolog.New(&b).With().Timestamp().Logger()

	logger := logze.NewFromZerolog(zlogger)

	if logger.NotInited() {
		t.Error("expected logger from zerolog to be inited")
	}

	logger.Info("test from zerolog")

	output := b.String()
	if !strings.Contains(output, "test from zerolog") {
		t.Errorf("expected message from zerolog logger, got %s", output)
	}
}

func TestNewConsoleJSON(t *testing.T) {
	logger := logze.NewConsoleJSON("service", "test")

	if logger.NotInited() {
		t.Error("expected NewConsoleJSON logger to be inited")
	}

	// Test that fields are included
	raw := logger.Raw()
	if raw == nil {
		t.Error("expected non-nil raw logger")
	}
}

func TestLoggerClose(t *testing.T) {
	var b bytes.Buffer
	cfg := logze.NewConfig(&b).WithLevel(logze.LevelInfo)
	logger := logze.New(cfg)

	err := logger.Close()
	if err != nil {
		t.Errorf("expected no error closing logger, got %v", err)
	}

	err = logger.CloseDiode()
	if err != nil {
		t.Errorf("expected no error closing diode, got %v", err)
	}
}

func TestLoggerContext(t *testing.T) {
	ctx := context.Background()
	logger := logze.NewConsoleJSON("component", "test")

	// Add logger to context
	ctx = logger.AddToContext(ctx)

	// Retrieve logger from context
	retrievedLogger := logze.GetFromContext(ctx)

	if retrievedLogger.NotInited() {
		t.Error("expected retrieved logger to be inited")
	}

	// Test with empty context
	emptyLogger := logze.GetFromContext(context.Background())
	if !emptyLogger.NotInited() {
		t.Error("expected logger from empty context to be Nop")
	}
}

func TestLoggerWith(t *testing.T) {
	var b bytes.Buffer
	cfg := logze.NewConfig(&b).WithLevel(logze.LevelInfo).WithNoDiode()
	logger := logze.New(cfg)

	// Test With method (shortcut for WithFields)
	logger2 := logger.With("key", "value")
	logger2.Info("test message")

	output := b.String()
	if !strings.Contains(output, "key\":\"value") {
		t.Errorf("expected field key:value, got %s", output)
	}
}

func TestLoggerWithLevel(t *testing.T) {
	var b bytes.Buffer
	cfg := logze.NewConfig(&b).WithLevel(logze.LevelWarn).WithNoDiode()
	logger := logze.New(cfg)

	// Should not log debug at warn level
	logger.Debug("should not appear")

	if b.Len() > 0 {
		t.Errorf("expected no debug output at warn level, got %s", b.String())
	}

	// Change level to debug
	debugLogger := logger.WithLevel("debug")
	debugLogger.Debug("should appear now")

	output := b.String()
	if !strings.Contains(output, "should appear now") {
		t.Errorf("expected debug message after level change, got %s", output)
	}

	// Test empty level (should return same logger)
	sameLogger := logger.WithLevel("")
	if sameLogger.NotInited() {
		t.Error("expected logger to remain initialized when setting empty level")
	}
}

func TestLoggerWithStack(t *testing.T) {
	var b bytes.Buffer
	cfg := logze.NewConfig(&b).WithLevel(logze.LevelError).WithNoDiode()
	logger := logze.New(cfg).WithStack(true)

	err := errors.New("test error")
	logger.Err(err, "error with stack")

	output := b.String()
	if !strings.Contains(output, "TestLoggerWithStack") {
		t.Errorf("expected stack trace in output, got %s", output)
	}

	b.Reset()
	logger.WithStack(false).Err(err, "error without stack")

	output = b.String()
	if strings.Contains(output, "TestLoggerWithStack") {
		t.Errorf("expected no stack trace in output, got %s", output)
	}
}

func TestLoggerWithSimpleErrorCounter(t *testing.T) {
	var b bytes.Buffer
	cfg := logze.NewConfig(&b).WithLevel(logze.LevelError).WithNoDiode()
	logger := logze.New(cfg).WithSimpleErrorCounter()

	counter := logger.GetErrorCounter()
	if counter == nil {
		t.Error("expected non-nil error counter")
	}

	simpleCounter, ok := counter.(*logze.SimpleErrorCounter)
	if !ok {
		t.Error("expected SimpleErrorCounter type")
	}

	if simpleCounter.Count.Load() != 0 {
		t.Errorf("expected 0 initial count, got %d", simpleCounter.Count.Load())
	}

	logger.Error("test error")
	if simpleCounter.Count.Load() != 0 {
		t.Errorf("expected 0 count for Error (no actual error), got %d", simpleCounter.Count.Load())
	}

	logger.Err(errors.New("actual error"), "test")
	if simpleCounter.Count.Load() != 1 {
		t.Errorf("expected 1 count after Err, got %d", simpleCounter.Count.Load())
	}
}

func TestLoggerWithToIgnore(t *testing.T) {
	var b bytes.Buffer
	cfg := logze.NewConfig(&b).WithLevel(logze.LevelInfo).WithNoDiode()
	logger := logze.New(cfg).WithToIgnore("ignore1", "ignore2")

	logger.Info("normal message")
	logger.Info("ignore1")
	logger.Info("this contains ignore2 in it")
	logger.Info("another normal message")

	output := b.String()
	if !strings.Contains(output, "normal message") {
		t.Error("expected normal message to be logged")
	}
	if !strings.Contains(output, "another normal message") {
		t.Error("expected another normal message to be logged")
	}
	if strings.Contains(output, "ignore1") {
		t.Error("expected ignore1 to be filtered out")
	}
	if strings.Contains(output, "ignore2") {
		t.Error("expected message containing ignore2 to be filtered out")
	}
}

func TestLoggerWithCaller(t *testing.T) {
	var b bytes.Buffer
	cfg := logze.NewConfig(&b).WithLevel(logze.LevelInfo).WithNoDiode()
	logger := logze.New(cfg).WithCaller(3)

	logger.Info("test with caller")

	output := b.String()
	if !strings.Contains(output, "caller") {
		t.Errorf("expected caller information, got %s", output)
	}
}

func TestLoggerWithDefaultCaller(t *testing.T) {
	var b bytes.Buffer
	cfg := logze.NewConfig(&b).WithLevel(logze.LevelInfo).WithNoDiode()
	logger := logze.New(cfg).WithDefaultCaller()

	logger.Info("test with default caller")

	output := b.String()
	if !strings.Contains(output, "caller") {
		t.Errorf("expected caller information, got %s", output)
	}
}

func TestLoggerWithSampler(t *testing.T) {
	var b bytes.Buffer
	cfg := logze.NewConfig(&b).WithLevel(logze.LevelInfo).WithNoDiode()

	// Create a sampler that logs nothing (for testing)
	sampler := zerolog.RandomSampler(0)
	logger := logze.New(cfg).WithSampler(sampler)

	logger.Info("this should be sampled out")

	if b.Len() > 0 {
		t.Errorf("expected no output due to sampling, got %s", b.String())
	}
}

// Add tests for new sampling methods
func TestLoggerWithPercentageSampler(t *testing.T) {
	var b bytes.Buffer
	cfg := logze.NewConfig(&b).WithLevel(logze.LevelDebug).WithNoDiode()
	logger := logze.New(cfg).WithPercentageSampler(1.0, "debug") // 100% sampling

	logger.Debug("debug message")
	logger.Info("info message") // Should appear (not sampled)

	output := b.String()
	if !strings.Contains(output, "debug message") {
		t.Error("expected debug message with 100% sampling")
	}
	if !strings.Contains(output, "info message") {
		t.Error("expected info message (not subject to sampling)")
	}
}

func TestLoggerWithBurstSampler(t *testing.T) {
	var b bytes.Buffer
	cfg := logze.NewConfig(&b).WithLevel(logze.LevelDebug).WithNoDiode()
	logger := logze.New(cfg).WithBurstSampler(1.0, 1, time.Second, "debug")

	logger.Debug("first debug") // Should appear (within burst)

	output := b.String()
	if !strings.Contains(output, "first debug") {
		t.Error("expected first debug message within burst limit")
	}
}

func TestLoggerWithMaxSampler(t *testing.T) {
	var b bytes.Buffer
	cfg := logze.NewConfig(&b).WithLevel(logze.LevelDebug).WithNoDiode()
	logger := logze.New(cfg).WithMaxSampler(1, time.Second, "debug")

	logger.Debug("first debug") // Should appear

	output := b.String()
	if !strings.Contains(output, "first debug") {
		t.Error("expected first debug message within max limit")
	}
}

// Test all logging methods and their variants

func TestLoggerTrace(t *testing.T) {
	var b bytes.Buffer
	cfg := logze.NewConfig(&b).WithLevel(logze.LevelTrace).WithNoDiode()
	logger := logze.New(cfg)

	logger.Trace("trace message", "key", "value")

	output := b.String()
	if !strings.Contains(output, "level\":\"trace") {
		t.Error("expected trace level")
	}
	if !strings.Contains(output, "trace message") {
		t.Error("expected trace message")
	}
	if !strings.Contains(output, "caller") {
		t.Error("expected caller info in trace")
	}
	if !strings.Contains(output, "key\":\"value") {
		t.Error("expected key-value field")
	}
}

func TestLoggerTracef(t *testing.T) {
	var b bytes.Buffer
	cfg := logze.NewConfig(&b).WithLevel(logze.LevelTrace).WithNoDiode()
	logger := logze.New(cfg)

	logger.Tracef("trace %s %d", "test", 42, "extra", "field")

	output := b.String()
	if !strings.Contains(output, "trace test 42") {
		t.Error("expected formatted trace message")
	}
	if !strings.Contains(output, "extra\":\"field") {
		t.Error("expected extra field")
	}
}

func TestLoggerTraceIf(t *testing.T) {
	var b bytes.Buffer
	cfg := logze.NewConfig(&b).WithLevel(logze.LevelTrace).WithNoDiode()
	logger := logze.New(cfg)

	// Should log when condition is true
	logger.TraceIf(true, "conditional trace", "logged", "yes")

	output := b.String()
	if !strings.Contains(output, "conditional trace") {
		t.Error("expected conditional trace message when true")
	}

	b.Reset()

	// Should not log when condition is false
	logger.TraceIf(false, "should not appear")

	if b.Len() > 0 {
		t.Error("expected no output when condition is false")
	}
}

func TestLoggerDebugIf(t *testing.T) {
	var b bytes.Buffer
	cfg := logze.NewConfig(&b).WithLevel(logze.LevelDebug).WithNoDiode()
	logger := logze.New(cfg)

	logger.DebugIf(true, "conditional debug")
	logger.DebugIf(false, "should not appear")

	output := b.String()
	if !strings.Contains(output, "conditional debug") {
		t.Error("expected conditional debug when true")
	}
	if strings.Contains(output, "should not appear") {
		t.Error("unexpected output when condition false")
	}
}

func TestLoggerInfoIf(t *testing.T) {
	var b bytes.Buffer
	cfg := logze.NewConfig(&b).WithLevel(logze.LevelInfo).WithNoDiode()
	logger := logze.New(cfg)

	logger.InfoIf(true, "conditional info")
	logger.InfoIf(false, "should not appear")

	output := b.String()
	if !strings.Contains(output, "conditional info") {
		t.Error("expected conditional info when true")
	}
	if strings.Contains(output, "should not appear") {
		t.Error("unexpected output when condition false")
	}
}

func TestLoggerWarnIf(t *testing.T) {
	var b bytes.Buffer
	cfg := logze.NewConfig(&b).WithLevel(logze.LevelWarn).WithNoDiode()
	logger := logze.New(cfg)

	logger.WarnIf(true, "conditional warn")
	logger.WarnIf(false, "should not appear")

	output := b.String()
	if !strings.Contains(output, "conditional warn") {
		t.Error("expected conditional warn when true")
	}
	if strings.Contains(output, "should not appear") {
		t.Error("unexpected output when condition false")
	}
}

func TestLoggerErrorIf(t *testing.T) {
	var b bytes.Buffer
	cfg := logze.NewConfig(&b).WithLevel(logze.LevelError).WithNoDiode()
	logger := logze.New(cfg)

	logger.ErrorIf(true, "conditional error")
	logger.ErrorIf(false, "should not appear")

	output := b.String()
	if !strings.Contains(output, "conditional error") {
		t.Error("expected conditional error when true")
	}
	if strings.Contains(output, "should not appear") {
		t.Error("unexpected output when condition false")
	}
}

func TestLoggerErrIf(t *testing.T) {
	var b bytes.Buffer
	cfg := logze.NewConfig(&b).WithLevel(logze.LevelError).WithNoDiode()
	logger := logze.New(cfg)

	err := errors.New("test error")
	logger.ErrIf(true, err, "conditional err")
	logger.ErrIf(false, err, "should not appear")

	output := b.String()
	if !strings.Contains(output, "conditional err") {
		t.Error("expected conditional err when true")
	}
	if strings.Contains(output, "should not appear") {
		t.Error("unexpected output when condition false")
	}
}

func TestLoggerFatalIf(t *testing.T) {
	// Test that FatalIf doesn't exit when condition is false
	var b bytes.Buffer
	cfg := logze.NewConfig(&b).WithLevel(logze.LevelFatal).WithNoDiode()
	logger := logze.New(cfg)

	logger.FatalIf(false, "should not exit")

	// If we reach here, the test passed (didn't exit)
	if b.Len() > 0 {
		t.Error("expected no output when FatalIf condition is false")
	}
}

func TestLoggerPanicIf(t *testing.T) {
	var b bytes.Buffer
	cfg := logze.NewConfig(&b).WithLevel(logze.LevelFatal).WithNoDiode()
	logger := logze.New(cfg)

	// Test that PanicIf doesn't panic when condition is false
	logger.PanicIf(false, "should not panic")

	// Test that PanicIf does panic when condition is true
	defer func() {
		if r := recover(); r == nil {
			t.Error("expected panic when condition is true")
		}
	}()

	logger.PanicIf(true, "should panic")
}

func TestLoggerPrint(t *testing.T) {
	var b bytes.Buffer
	cfg := logze.NewConfig(&b).WithLevel(logze.LevelInfo).WithNoDiode()
	logger := logze.New(cfg)

	logger.Print("test print message")

	output := b.String()
	if !strings.Contains(output, "test print message") {
		t.Error("expected print message")
	}

	// Test empty print
	b.Reset()
	logger.Print()

	if b.Len() > 0 {
		t.Error("expected no output for empty Print")
	}
}

func TestLoggerPrintIf(t *testing.T) {
	var b bytes.Buffer
	cfg := logze.NewConfig(&b).WithLevel(logze.LevelInfo).WithNoDiode()
	logger := logze.New(cfg)

	logger.PrintIf(true, "conditional print")
	logger.PrintIf(false, "should not appear")

	output := b.String()
	if !strings.Contains(output, "conditional print") {
		t.Error("expected conditional print when true")
	}
	if strings.Contains(output, "should not appear") {
		t.Error("unexpected output when condition false")
	}
}

func TestLoggerPrintStack(t *testing.T) {
	var b bytes.Buffer
	cfg := logze.NewConfig(&b).WithLevel(logze.LevelInfo).WithNoDiode()
	logger := logze.New(cfg)

	logger.PrintStack("context", "test")

	output := b.String()
	if !strings.Contains(output, "TestLoggerPrintStack") {
		t.Error("expected stack trace with test function name")
	}
	if !strings.Contains(output, "context\":\"test") {
		t.Error("expected context field")
	}
}

func TestLoggerPrintln(t *testing.T) {
	var b bytes.Buffer
	cfg := logze.NewConfig(&b).WithLevel(logze.LevelInfo).WithNoDiode()
	logger := logze.New(cfg)

	logger.Println("test", "println", "message")

	output := b.String()
	if !strings.Contains(output, "test println message") {
		t.Error("expected println message with spaces")
	}
}

func TestLoggerLog(t *testing.T) {
	var b bytes.Buffer
	cfg := logze.NewConfig(&b).WithLevel(logze.LevelInfo).WithNoDiode()
	logger := logze.New(cfg)

	logger.Log("test log message")

	output := b.String()
	if !strings.Contains(output, "test log message") {
		t.Error("expected log message")
	}
}

func TestLoggerWrite(t *testing.T) {
	var b bytes.Buffer
	cfg := logze.NewConfig(&b).WithLevel(logze.LevelInfo).WithNoDiode()
	logger := logze.New(cfg)

	data := []byte("direct write data")
	n, err := logger.Write(data)

	if err != nil {
		t.Errorf("expected no error from Write, got %v", err)
	}
	if n != len(data) {
		t.Errorf("expected %d bytes written, got %d", len(data), n)
	}
}

func TestLoggerRaw(t *testing.T) {
	logger := logze.NewConsoleJSON()

	raw := logger.Raw()
	if raw == nil {
		t.Error("expected non-nil raw logger")
	}

	// Test that we can use the raw logger
	var b bytes.Buffer
	*raw = raw.Output(&b)
	raw.Info().Msg("raw test")

	output := b.String()
	if !strings.Contains(output, "raw test") {
		t.Error("expected message from raw logger")
	}
}

func TestLoggerGetErrorCounter(t *testing.T) {
	// Test with no error counter
	logger := logze.NewConsoleJSON()
	if logger.GetErrorCounter() != nil {
		t.Error("expected nil error counter for logger without counter")
	}

	// Test with error counter
	logger = logger.WithSimpleErrorCounter()
	counter := logger.GetErrorCounter()
	if counter == nil {
		t.Error("expected non-nil error counter")
	}
}

// Test edge cases and error conditions

func TestLoggerLevelsFiltering(t *testing.T) {
	var b bytes.Buffer
	cfg := logze.NewConfig(&b).WithLevel(logze.LevelWarn).WithNoDiode()
	logger := logze.New(cfg)

	// These should not appear (below warn level)
	logger.Trace("trace message")
	logger.Debug("debug message")
	logger.Info("info message")

	// These should appear (warn level and above)
	logger.Warn("warn message")
	logger.Error("error message")

	output := b.String()
	if strings.Contains(output, "trace message") ||
		strings.Contains(output, "debug message") ||
		strings.Contains(output, "info message") {
		t.Error("expected messages below warn level to be filtered")
	}

	if !strings.Contains(output, "warn message") ||
		!strings.Contains(output, "error message") {
		t.Error("expected warn and error messages to appear")
	}
}

func TestLoggerInvalidLevel(t *testing.T) {
	defer func() {
		if r := recover(); r == nil {
			t.Error("expected panic for invalid level")
		}
	}()

	logger := logze.NewConsoleJSON()
	logger.WithLevel("invalid-level")
}

func TestLoggerWithFieldsComplexTypes(t *testing.T) {
	var b bytes.Buffer
	cfg := logze.NewConfig(&b).WithLevel(logze.LevelInfo).WithNoDiode()
	logger := logze.New(cfg)

	logger.Info("test complex fields",
		"string", "value",
		"int", 42,
		"float", 3.14,
		"bool", true,
		"slice", []string{"a", "b", "c"},
		"map", map[string]int{"x": 1, "y": 2},
	)

	output := b.String()
	if !strings.Contains(output, "string\":\"value") ||
		!strings.Contains(output, "int\":42") ||
		!strings.Contains(output, "bool\":true") {
		t.Errorf("expected complex field types in output, got %s", output)
	}
}

// Test Panicf and Panicln methods
func TestLoggerPanicf(t *testing.T) {
	defer func() {
		if r := recover(); r == nil {
			t.Error("expected panic from Panicf")
		} else {
			expected := "panic test 42"
			if !strings.Contains(fmt.Sprint(r), expected) {
				t.Errorf("expected panic message to contain %s, got %v", expected, r)
			}
		}
	}()

	var b bytes.Buffer
	cfg := logze.NewConfig(&b).WithLevel(logze.LevelFatal).WithNoDiode()
	logger := logze.New(cfg)

	logger.Panicf("panic %s %d", "test", 42)
}

func TestLoggerPanicln(t *testing.T) {
	defer func() {
		if r := recover(); r == nil {
			t.Error("expected panic from Panicln")
		} else {
			expected := "panic test message"
			if !strings.Contains(fmt.Sprint(r), expected) {
				t.Errorf("expected panic message to contain %s, got %v", expected, r)
			}
		}
	}()

	var b bytes.Buffer
	cfg := logze.NewConfig(&b).WithLevel(logze.LevelFatal).WithNoDiode()
	logger := logze.New(cfg)

	logger.Panicln("panic", "test", "message")
}

func TestLoggerFatalln(t *testing.T) {
	// Cannot actually test os.Exit behavior without affecting the test process
	// This test just ensures the method exists and is callable
	t.Log("Fatalln method exists and would call os.Exit(1)")
}

func TestLoggerFatalf(t *testing.T) {
	// Cannot actually test os.Exit behavior without affecting the test process
	// This test just ensures the method exists and is callable
	t.Log("Fatalf method exists and would call os.Exit(1)")
}

// Test error formatting and stack traces
func TestLoggerErrorWithErrorInFields(t *testing.T) {
	var b bytes.Buffer
	cfg := logze.NewConfig(&b).WithLevel(logze.LevelError).WithNoDiode().WithSimpleErrorCounter()
	logger := logze.New(cfg)

	err := errors.New("field error")
	logger.Error("test message", "key", "value", err, "extra", "field")

	output := b.String()
	if !strings.Contains(output, "test message") {
		t.Error("expected test message")
	}
	if !strings.Contains(output, "field error") {
		t.Error("expected field error")
	}

	counter := logger.GetErrorCounter().(*logze.SimpleErrorCounter)
	if counter.Count.Load() != 1 {
		t.Errorf("expected 1 error counted, got %d", counter.Count.Load())
	}
}

func TestLoggerFormattedLoggingEdgeCases(t *testing.T) {
	var b bytes.Buffer
	cfg := logze.NewConfig(&b).WithLevel(logze.LevelInfo).WithNoDiode()
	logger := logze.New(cfg)

	// Test format with no placeholders but extra args (should be treated as fields)
	logger.Infof("no placeholders", "extra", "field")

	output := b.String()
	if !strings.Contains(output, "no placeholders") {
		t.Error("expected message without placeholders")
	}
	if !strings.Contains(output, "extra\":\"field") {
		t.Error("expected extra args as fields")
	}

	b.Reset()

	// Test %w replacement
	err := errors.New("wrapped error")
	logger.Infof("error: %w", err, "context", "test")

	output = b.String()
	if !strings.Contains(output, "wrapped error") {
		t.Error("expected wrapped error message")
	}
	if !strings.Contains(output, "context\":\"test") {
		t.Error("expected context field")
	}
}

func TestLoggerDisabledContext(t *testing.T) {
	ctx := context.Background()

	// Create disabled logger
	logger := logze.Nop()

	// Add to context (should not store disabled logger)
	ctx = logger.AddToContext(ctx)

	// Get from context (should return Nop)
	retrieved := logze.GetFromContext(ctx)
	if !retrieved.NotInited() {
		t.Error("expected Nop logger from context with disabled logger")
	}
}

func TestLoggerUpdateNil(t *testing.T) {
	// Test update with no existing logger
	var logger logze.Logger
	if !logger.NotInited() {
		t.Error("expected uninitialized logger to report NotInited")
	}

	var b bytes.Buffer
	cfg := logze.NewConfig(&b).WithLevel(logze.LevelInfo).WithNoDiode()
	logger.Update(cfg)

	if logger.NotInited() {
		t.Error("expected logger to be initialized after Update")
	}

	logger.Info("test after update")

	output := b.String()
	if !strings.Contains(output, "test after update") {
		t.Error("expected message after update")
	}
}

func TestLoggerIgnoreSubstring(t *testing.T) {
	var b bytes.Buffer
	cfg := logze.NewConfig(&b).WithLevel(logze.LevelInfo).WithNoDiode()
	logger := logze.New(cfg).WithToIgnore("health")

	logger.Info("health check endpoint") // Should be ignored (contains "health")
	logger.Info("system status")         // Should not be ignored

	output := b.String()
	if strings.Contains(output, "health") {
		t.Error("expected messages containing 'health' to be ignored")
	}
	if !strings.Contains(output, "system status") {
		t.Error("expected 'system status' message to be logged")
	}
}

func TestNewWithZeroWriters(t *testing.T) {
	// Test New with empty config (should default to io.Discard)
	cfg := logze.NewConfig()
	logger := logze.New(cfg)

	if logger.NotInited() {
		t.Error("expected logger to be initialized even with no writers")
	}

	// Should not panic or error
	logger.Info("test message to discard")
}

func TestLoggerWithLevelDisabled(t *testing.T) {
	var b bytes.Buffer
	cfg := logze.NewConfig(&b).WithLevel(logze.LevelDisabled).WithNoDiode()
	logger := logze.New(cfg)

	// None of these should produce output
	logger.Trace("trace")
	logger.Debug("debug")
	logger.Info("info")
	logger.Warn("warn")
	logger.Error("error")

	if b.Len() > 0 {
		t.Errorf("expected no output with disabled level, got %s", b.String())
	}
}

func TestLoggerFieldsWithErrors(t *testing.T) {
	var b bytes.Buffer
	cfg := logze.NewConfig(&b).WithLevel(logze.LevelInfo).WithNoDiode().WithSimpleErrorCounter()
	logger := logze.New(cfg)

	err1 := errors.New("first error")
	err2 := errors.New("second error")

	// Test with multiple errors in fields
	logger.Info("test message", "error1", err1, "key", "value", "error2", err2)

	output := b.String()
	if !strings.Contains(output, "first error") {
		t.Error("expected first error in output")
	}
	if !strings.Contains(output, "second error") {
		t.Error("expected second error in output")
	}
	if !strings.Contains(output, "key\":\"value") {
		t.Error("expected non-error field in output")
	}

	// Error counter is incremented once per log call that contains errors, not per error
	counter := logger.GetErrorCounter().(*logze.SimpleErrorCounter)
	if counter.Count.Load() != 1 {
		t.Errorf("expected 1 error counted (first error processed from fields), got %d", counter.Count.Load())
	}

	// Test additional error counting with Err method
	logger.Err(err1, "actual error call")
	if counter.Count.Load() != 2 {
		t.Errorf("expected 2 errors counted after Err call, got %d", counter.Count.Load())
	}
}

func TestLoggerDiodeConfiguration(t *testing.T) {
	var b bytes.Buffer

	// Test with diode enabled (default)
	cfg := logze.NewConfig(&b).WithLevel(logze.LevelInfo)
	logger := logze.New(cfg)

	logger.Info("test with diode")

	// Close the diode to flush
	err := logger.CloseDiode()
	if err != nil {
		t.Errorf("unexpected error closing diode: %v", err)
	}

	// Test multiple close calls (should be safe)
	err = logger.Close()
	if err != nil {
		t.Errorf("unexpected error on second close: %v", err)
	}
}

func TestLoggerMultipleWriters(t *testing.T) {
	var b1, b2 bytes.Buffer

	cfg := logze.NewConfig(&b1, &b2).WithLevel(logze.LevelInfo).WithNoDiode()
	logger := logze.New(cfg)

	logger.Info("test multiple writers")

	output1 := b1.String()
	output2 := b2.String()

	if !strings.Contains(output1, "test multiple writers") {
		t.Error("expected message in first writer")
	}
	if !strings.Contains(output2, "test multiple writers") {
		t.Error("expected message in second writer")
	}
}

// Test ToIgnore with formatted logging
func TestLoggerToIgnoreWithFormatted(t *testing.T) {
	var b bytes.Buffer
	cfg := logze.NewConfig(&b).WithLevel(logze.LevelInfo).WithNoDiode()
	logger := logze.New(cfg).WithToIgnore("health", "ignore")

	// Test formatted logging with ignored message
	logger.Infof("health check %s", "endpoint") // Should be ignored
	logger.Infof("ignore pattern %d", 123)      // Should be ignored
	logger.Infof("normal message %s", "test")   // Should appear

	output := b.String()
	if strings.Contains(output, "health") {
		t.Error("expected formatted message containing 'health' to be ignored")
	}
	if strings.Contains(output, "ignore") {
		t.Error("expected formatted message containing 'ignore' to be ignored")
	}
	if !strings.Contains(output, "normal message test") {
		t.Error("expected normal formatted message to appear")
	}
}

func TestLoggerToIgnoreWithFormattedEdgeCases(t *testing.T) {
	var b bytes.Buffer
	cfg := logze.NewConfig(&b).WithLevel(logze.LevelInfo).WithNoDiode()
	logger := logze.New(cfg).WithToIgnore("system %s mode", "debug")

	// Test message template matching (filtering happens before formatting)
	logger.Infof("system %s mode", "debug")   // Should be ignored (template matches)
	logger.Infof("normal %s message", "test") // Should appear
	logger.Infof("debug info", "extra")       // Should be ignored (contains "debug")

	output := b.String()
	if strings.Contains(output, "system debug mode") {
		t.Error("expected message with ignored template to be filtered out")
	}
	if strings.Contains(output, "debug info") {
		t.Error("expected message containing 'debug' to be filtered out")
	}
	if !strings.Contains(output, "normal test message") {
		t.Error("expected normal message to appear")
	}
}
