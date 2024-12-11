package logze_test

import (
	"bytes"
	"fmt"
	"log/slog"
	"runtime/debug"
	"testing"

	"github.com/maxbolgarin/logze/v2"
	"github.com/pkg/errors"
	"github.com/rs/zerolog"
)

func setupLogzeLogger(buffer *bytes.Buffer) logze.Logger {
	cfg := logze.NewConfig(buffer).WithLevel(logze.LevelDebug).WithNoDiode()
	return logze.New(cfg)
}

func setupZerologLogger(buffer *bytes.Buffer) zerolog.Logger {
	return zerolog.New(buffer).With().Timestamp().Logger().Level(zerolog.DebugLevel)
}

func setupSLogger(buffer *bytes.Buffer) *slog.Logger {
	h := slog.NewJSONHandler(buffer, &slog.HandlerOptions{
		Level: slog.LevelDebug,
	})
	return slog.New(h)
}

// Info

func BenchmarkZerologInfo(b *testing.B) {
	var buffer bytes.Buffer
	logger := setupZerologLogger(&buffer)

	for i := 0; i < b.N; i++ {
		buffer.Reset()
		logger.Info().Str("key", "value").Int("number", 123).Msg("error message")
	}
}

func BenchmarkLogzeInfo(b *testing.B) {
	var buffer bytes.Buffer
	logger := setupLogzeLogger(&buffer)

	for i := 0; i < b.N; i++ {
		buffer.Reset()
		logger.Info("error message", "key", "value", "number", 123)
	}
}

func BenchmarkSLogInfo(b *testing.B) {
	var buffer bytes.Buffer
	logger := setupSLogger(&buffer)

	for i := 0; i < b.N; i++ {
		buffer.Reset()
		logger.Info("error message", "key", "value", "number", 123)
	}
}

// Info format

func BenchmarkZerologInfoFormat(b *testing.B) {
	var buffer bytes.Buffer
	logger := setupZerologLogger(&buffer)

	for i := 0; i < b.N; i++ {
		buffer.Reset()
		logger.Info().Str("key", "value").Int("number", 123).Msgf("error message %s", "formatted")
	}
}

func BenchmarkLogzeInfoFormat(b *testing.B) {
	var buffer bytes.Buffer
	logger := setupLogzeLogger(&buffer)

	for i := 0; i < b.N; i++ {
		buffer.Reset()
		logger.Infof("error message %s", "formatted", "key", "value", "number", 123)
	}
}

func BenchmarkSLogInfoFormat(b *testing.B) {
	var buffer bytes.Buffer
	logger := setupSLogger(&buffer)

	for i := 0; i < b.N; i++ {
		buffer.Reset()
		logger.Info(fmt.Sprintf("error message %s", "formatted"), "key", "value", "number", 123)
	}
}

// Error

func BenchmarkZerologError(b *testing.B) {
	var buffer bytes.Buffer
	logger := setupZerologLogger(&buffer)
	err := errors.New("an error occurred")

	for i := 0; i < b.N; i++ {
		buffer.Reset()
		logger.Error().Err(err).Str("key", "value").Int("number", 123).Msg("error message")
	}
}

func BenchmarkLogzeError(b *testing.B) {
	var buffer bytes.Buffer
	logger := setupLogzeLogger(&buffer)
	err := errors.New("an error occurred")

	for i := 0; i < b.N; i++ {
		buffer.Reset()
		logger.Error("error message", "error", err, "key", "value", "number", 123)
	}
}

func BenchmarkSLogError(b *testing.B) {
	var buffer bytes.Buffer
	logger := setupSLogger(&buffer)
	err := errors.New("an error occurred")

	for i := 0; i < b.N; i++ {
		buffer.Reset()
		logger.Error("error message", "error", err, "key", "value", "number", 123)
	}
}

// Error with stack

func BenchmarkZerologErrorWithStack(b *testing.B) {
	var buffer bytes.Buffer
	logger := setupZerologLogger(&buffer)
	err := errors.New("an error occurred")

	for i := 0; i < b.N; i++ {
		buffer.Reset()
		logger.Error().Stack().Err(errors.WithStack(err)).Str("key", "value").Int("number", 123).Msg("error message")
	}
}

func BenchmarkLogzeErrorWithStack(b *testing.B) {
	var buffer bytes.Buffer
	logger := setupLogzeLogger(&buffer).WithStack(true)
	err := errors.New("an error occurred")

	for i := 0; i < b.N; i++ {
		buffer.Reset()
		logger.Error("error message", "error", err, "key", "value", "number", 123)
	}
}

func BenchmarkSLogErrorWithStack(b *testing.B) {
	var buffer bytes.Buffer
	logger := setupSLogger(&buffer)
	err := errors.New("an error occurred")

	for i := 0; i < b.N; i++ {
		buffer.Reset()
		stack := debug.Stack()
		logger.Error("error message", "error", err, "key", "value", "number", 123, "stack", string(stack))
	}
}

// Info console

func BenchmarkZerologInfoConsole(b *testing.B) {
	var buffer bytes.Buffer
	logger := zerolog.New(zerolog.ConsoleWriter{Out: &buffer}).With().Timestamp().Logger().Level(zerolog.DebugLevel)

	for i := 0; i < b.N; i++ {
		buffer.Reset()
		logger.Info().Str("key", "value").Int("number", 123).Msg("error message")
	}
}

func BenchmarkLogzeInfoConsole(b *testing.B) {
	var buffer bytes.Buffer
	logger := logze.New(logze.C(zerolog.ConsoleWriter{Out: &buffer}).WithLevel(logze.LevelDebug).WithNoDiode())

	for i := 0; i < b.N; i++ {
		buffer.Reset()
		logger.Info("error message", "key", "value", "number", 123)
	}
}

func BenchmarkSLogInfoConsole(b *testing.B) {
	var buffer bytes.Buffer
	h := slog.NewTextHandler(&buffer, &slog.HandlerOptions{
		Level: slog.LevelDebug,
	})
	logger := slog.New(h)

	for i := 0; i < b.N; i++ {
		buffer.Reset()
		logger.Info("error message", "key", "value", "number", 123)
	}
}

// Additional logze features

func BenchmarkLogzeErr(b *testing.B) {
	var buffer bytes.Buffer
	logger := setupLogzeLogger(&buffer)
	err := errors.New("an error occurred")

	for i := 0; i < b.N; i++ {
		buffer.Reset()
		logger.Err(err, "error message", "key", "value", "number", 123)
	}
}

func BenchmarkLogzeToIgnore5(b *testing.B) {
	var buffer bytes.Buffer
	logger := setupLogzeLogger(&buffer).WithToIgnore(
		"ignore me",
		"ignore me too",
		"ignore me three",
		"some error in http module",
		"GOAWAY received",
	)
	err := errors.New("an error occurred")

	for i := 0; i < b.N; i++ {
		buffer.Reset()
		logger.Error("error message", "error", err, "key", "value", "number", 123)
	}
}
