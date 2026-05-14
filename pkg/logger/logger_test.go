package logger

import (
	"bytes"
	"testing"

	"github.com/LarsArtmann/art-dupl/internal/testutil"
)

func TestDefaultConfig(t *testing.T) {
	cfg := DefaultConfig()

	if cfg.Level != "info" {
		t.Errorf("Expected level 'info', got '%s'", cfg.Level)
	}

	if cfg.Output == nil {
		t.Error("Expected output to be non-nil")
	}

	if cfg.ReportCaller {
		t.Error("Expected ReportCaller to be false")
	}

	if cfg.Prefix != "" {
		t.Errorf("Expected empty prefix, got '%s'", cfg.Prefix)
	}
}

func TestNewLogger_NilConfig(t *testing.T) {
	logger := NewLogger(nil)

	if logger == nil {
		t.Error("Expected logger to be non-nil")
	}

	if _, ok := logger.(*charmLogger); !ok {
		t.Error("Expected *charmLogger type")
	}
}

func TestNewLogger_ValidConfig(t *testing.T) {
	cfg := &Config{
		Level:        "debug",
		Output:       &bytes.Buffer{},
		ReportCaller: false,
		Prefix:       "test",
	}

	logger := NewLogger(cfg)

	if logger == nil {
		t.Error("Expected logger to be non-nil")
	}
}

func TestNewLogger_InvalidLevel(t *testing.T) {
	cfg := &Config{
		Level:  "invalid",
		Output: &bytes.Buffer{},
	}

	logger := NewLogger(cfg)

	if logger == nil {
		t.Error("Expected logger to be non-nil with invalid level")
	}
}

func TestCharmLogger_Methods(t *testing.T) {
	tests := []struct {
		name   string
		method func(logger Logger, msg string, args ...any)
		msg    string
	}{
		{
			name:   "Debug",
			method: func(l Logger, msg string, args ...any) { l.Debug(msg, args...) },
			msg:    "debug message",
		},
		{
			name:   "Info",
			method: func(l Logger, msg string, args ...any) { l.Info(msg, args...) },
			msg:    "info message",
		},
		{
			name:   "Warn",
			method: func(l Logger, msg string, args ...any) { l.Warn(msg, args...) },
			msg:    "warn message",
		},
		{
			name:   "Error",
			method: func(l Logger, msg string, args ...any) { l.Error(msg, args...) },
			msg:    "error message",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			buf := &bytes.Buffer{}
			cfg := &Config{
				Level:  "debug",
				Output: buf,
			}
			logger := NewLogger(cfg)

			tt.method(logger, tt.msg)

			output := buf.String()
			testutil.AssertStringContains(
				t,
				output,
				tt.msg,
				"Expected output to contain '"+tt.msg+"'",
			)
		})
	}
}

func TestCharmLogger_WithArgs(t *testing.T) {
	buf := &bytes.Buffer{}
	cfg := &Config{
		Level:  "debug",
		Output: buf,
	}
	logger := NewLogger(cfg)

	logger.Info("test message", "key", "value")

	output := buf.String()
	testutil.AssertStringContains(
		t,
		output,
		"test message",
		"Expected output to contain 'test message'",
	)

	testutil.AssertStringContains(t, output, "key", "Expected output to contain key")
	testutil.AssertStringContains(t, output, "value", "Expected output to contain value")
}

func TestCharmLogger_WithoutArgs(t *testing.T) {
	buf := &bytes.Buffer{}
	cfg := &Config{
		Level:  "debug",
		Output: buf,
	}
	logger := NewLogger(cfg)

	logger.Info("test message")

	output := buf.String()
	testutil.AssertStringContains(
		t,
		output,
		"test message",
		"Expected output to contain 'test message'",
	)
}

func TestNoOpLogger(t *testing.T) {
	logger := &NoOpLogger{}

	// These should not panic
	logger.Debug("debug")
	logger.Info("info")
	logger.Warn("warn")
	logger.Error("error")
}

func TestNoOpLogger_ImplementsInterface(t *testing.T) {
	var _ Logger = &NoOpLogger{}
}

func TestCharmLogger_ImplementsInterface(t *testing.T) {
	var _ Logger = &charmLogger{}
}

func TestDefaultLogger(t *testing.T) {
	if Default == nil {
		t.Error("Expected Default to be non-nil")
	}

	// Default is NoOpLogger, these should not panic
	Default.Debug("debug")
	Default.Info("info")
	Default.Warn("warn")
	Default.Error("error")
}

func TestNewLogger_AllLevels(t *testing.T) {
	levels := []string{"debug", "info", "warn", "error"}

	for _, level := range levels {
		t.Run(level, func(t *testing.T) {
			cfg := &Config{
				Level:  level,
				Output: &bytes.Buffer{},
			}

			logger := NewLogger(cfg)
			if logger == nil {
				t.Errorf("Expected logger to be non-nil for level '%s'", level)
			}
		})
	}
}

func TestNewLogger_WithPrefix(t *testing.T) {
	buf := &bytes.Buffer{}
	cfg := &Config{
		Level:  "info",
		Output: buf,
		Prefix: "MYPREFIX",
	}
	logger := NewLogger(cfg)

	logger.Info("test message")

	output := buf.String()
	testutil.AssertStringContains(
		t,
		output,
		"MYPREFIX",
		"Expected output to contain prefix 'MYPREFIX'",
	)
}

func TestNewLogger_WithReportCaller(t *testing.T) {
	buf := &bytes.Buffer{}
	cfg := &Config{
		Level:        "info",
		Output:       buf,
		ReportCaller: true,
	}
	logger := NewLogger(cfg)

	logger.Info("test message")

	output := buf.String()
	// When ReportCaller is true, output should contain file/line info
	// The exact format depends on charmbracelet/log implementation
	if output == "" {
		t.Error("Expected non-empty output with ReportCaller enabled")
	}
}
