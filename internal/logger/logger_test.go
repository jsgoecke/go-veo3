package logger

import (
	"bytes"
	"strings"
	"sync"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestLevel_String(t *testing.T) {
	tests := []struct {
		level Level
		want  string
	}{
		{DebugLevel, "DEBUG"},
		{InfoLevel, "INFO"},
		{WarnLevel, "WARN"},
		{ErrorLevel, "ERROR"},
		{Level(999), "UNKNOWN"},
	}

	for _, tt := range tests {
		t.Run(tt.want, func(t *testing.T) {
			result := tt.level.String()
			assert.Equal(t, tt.want, result)
		})
	}
}

func TestNewLogger(t *testing.T) {
	buf := &bytes.Buffer{}
	logger := NewLogger(InfoLevel, buf, false, false)

	assert.NotNil(t, logger)
	assert.Equal(t, InfoLevel, logger.level)
	assert.Equal(t, buf, logger.output)
	assert.False(t, logger.verbose)
	assert.False(t, logger.quiet)
}

func TestLogger_WithPrefix(t *testing.T) {
	buf := &bytes.Buffer{}
	logger := NewLogger(InfoLevel, buf, false, false)
	prefixedLogger := logger.WithPrefix("TEST")

	assert.NotNil(t, prefixedLogger)
	assert.Equal(t, "TEST", prefixedLogger.prefix)
	assert.Equal(t, logger.level, prefixedLogger.level)
	assert.Equal(t, logger.output, prefixedLogger.output)
}

func TestLogger_Debug(t *testing.T) {
	buf := &bytes.Buffer{}
	logger := NewLogger(DebugLevel, buf, true, false)

	logger.Debug("test debug message: %s", "value")
	output := buf.String()

	assert.Contains(t, output, "[DEBUG]")
	assert.Contains(t, output, "test debug message: value")
}

func TestLogger_Debug_NotShownWhenLevelTooHigh(t *testing.T) {
	buf := &bytes.Buffer{}
	logger := NewLogger(InfoLevel, buf, false, false)

	logger.Debug("this should not appear")
	output := buf.String()

	assert.Empty(t, output)
}

func TestLogger_Info(t *testing.T) {
	buf := &bytes.Buffer{}
	logger := NewLogger(InfoLevel, buf, false, false)

	logger.Info("test info message: %d", 42)
	output := buf.String()

	assert.Contains(t, output, "[INFO]")
	assert.Contains(t, output, "test info message: 42")
}

func TestLogger_Info_QuietMode(t *testing.T) {
	buf := &bytes.Buffer{}
	logger := NewLogger(InfoLevel, buf, false, true)

	logger.Info("this should not appear in quiet mode")
	output := buf.String()

	assert.Empty(t, output)
}

func TestLogger_Warn(t *testing.T) {
	buf := &bytes.Buffer{}
	logger := NewLogger(InfoLevel, buf, false, false)

	logger.Warn("test warning: %s", "something")
	output := buf.String()

	assert.Contains(t, output, "[WARN]")
	assert.Contains(t, output, "test warning: something")
}

func TestLogger_Error(t *testing.T) {
	buf := &bytes.Buffer{}
	logger := NewLogger(InfoLevel, buf, false, false)

	logger.Error("test error: %v", "failed")
	output := buf.String()

	assert.Contains(t, output, "[ERROR]")
	assert.Contains(t, output, "test error: failed")
}

func TestLogger_WithPrefix_Output(t *testing.T) {
	buf := &bytes.Buffer{}
	logger := NewLogger(InfoLevel, buf, false, false)
	prefixedLogger := logger.WithPrefix("MODULE")

	prefixedLogger.Info("test message")
	output := buf.String()

	assert.Contains(t, output, "[MODULE]")
	assert.Contains(t, output, "test message")
}

func TestLogger_SetLevel(t *testing.T) {
	buf := &bytes.Buffer{}
	logger := NewLogger(InfoLevel, buf, false, false)

	// Initially info level, debug should not show
	logger.Debug("should not appear")
	assert.Empty(t, buf.String())

	// Change to debug level
	logger.SetLevel(DebugLevel)
	logger.Debug("should appear")
	assert.Contains(t, buf.String(), "[DEBUG]")
}

func TestLogger_SetOutput(t *testing.T) {
	buf1 := &bytes.Buffer{}
	buf2 := &bytes.Buffer{}
	logger := NewLogger(InfoLevel, buf1, false, false)

	logger.Info("to buf1")
	assert.Contains(t, buf1.String(), "to buf1")
	assert.Empty(t, buf2.String())

	logger.SetOutput(buf2)
	logger.Info("to buf2")
	assert.Contains(t, buf2.String(), "to buf2")
}

func TestLogger_IsDebug(t *testing.T) {
	logger := NewLogger(DebugLevel, &bytes.Buffer{}, true, false)
	assert.True(t, logger.IsDebug())

	logger.SetLevel(InfoLevel)
	assert.False(t, logger.IsDebug())
}

func TestLogger_IsVerbose(t *testing.T) {
	logger := NewLogger(InfoLevel, &bytes.Buffer{}, true, false)
	assert.True(t, logger.IsVerbose())

	logger = NewLogger(InfoLevel, &bytes.Buffer{}, false, false)
	assert.False(t, logger.IsVerbose())
}

func TestLogger_IsQuiet(t *testing.T) {
	logger := NewLogger(InfoLevel, &bytes.Buffer{}, false, true)
	assert.True(t, logger.IsQuiet())

	logger = NewLogger(InfoLevel, &bytes.Buffer{}, false, false)
	assert.False(t, logger.IsQuiet())
}

func TestLogger_Nil(t *testing.T) {
	var logger *Logger

	// Should not panic with nil logger
	assert.NotPanics(t, func() {
		logger.Debug("test")
		logger.Info("test")
		logger.Warn("test")
		logger.Error("test")
	})

	assert.False(t, logger.IsDebug())
	assert.False(t, logger.IsVerbose())
	assert.False(t, logger.IsQuiet())
}

func TestInit(t *testing.T) {
	// Reset Default to test Init
	Default = nil
	once = sync.Once{}

	Init(true, false)
	assert.NotNil(t, Default)
	assert.Equal(t, DebugLevel, Default.level)
	assert.True(t, Default.verbose)
	assert.False(t, Default.quiet)

	// Reset for other tests
	Default = nil
	once = sync.Once{}
}

func TestGlobalFunctions(t *testing.T) {
	buf := &bytes.Buffer{}

	// Reset and initialize Default
	Default = nil
	once = sync.Once{}
	Init(false, false)
	Default.SetOutput(buf)

	Debug("debug message")
	Info("info message")
	Warn("warn message")
	Error("error message")

	output := buf.String()
	assert.Contains(t, output, "info message")
	assert.Contains(t, output, "warn message")
	assert.Contains(t, output, "error message")
	// Debug should not appear since verbose is false
	assert.NotContains(t, output, "debug message")

	// Reset for other tests
	Default = nil
	once = sync.Once{}
}

func TestGlobalFunctions_WithNilDefault(t *testing.T) {
	Default = nil
	once = sync.Once{}

	// Should not panic with nil Default
	assert.NotPanics(t, func() {
		Debug("test")
		Info("test")
		Warn("test")
		Error("test")
	})

	// Reset
	Default = nil
	once = sync.Once{}
}

func TestWithPrefix_GlobalFunction(t *testing.T) {
	buf := &bytes.Buffer{}

	Default = nil
	once = sync.Once{}
	Init(false, false)
	Default.SetOutput(buf)

	prefixedLogger := WithPrefix("GLOBAL")
	assert.NotNil(t, prefixedLogger)
	assert.Equal(t, "GLOBAL", prefixedLogger.prefix)

	prefixedLogger.Info("test message")
	output := buf.String()
	assert.Contains(t, output, "[GLOBAL]")
	assert.Contains(t, output, "test message")

	// Reset
	Default = nil
	once = sync.Once{}
}

func TestWithPrefix_NilDefault(t *testing.T) {
	Default = nil
	once = sync.Once{}

	prefixedLogger := WithPrefix("TEST")
	assert.NotNil(t, prefixedLogger)
	assert.Equal(t, "TEST", prefixedLogger.prefix)

	// Reset
	Default = nil
	once = sync.Once{}
}

func TestLogger_Log_ThreadSafety(t *testing.T) {
	buf := &bytes.Buffer{}
	logger := NewLogger(InfoLevel, buf, false, false)

	// Test concurrent logging
	done := make(chan bool)
	for i := 0; i < 10; i++ {
		go func(id int) {
			logger.Info("message %d", id)
			done <- true
		}(i)
	}

	// Wait for all goroutines
	for i := 0; i < 10; i++ {
		<-done
	}

	output := buf.String()
	lines := strings.Split(strings.TrimSpace(output), "\n")
	assert.Equal(t, 10, len(lines))
}

func TestLogger_MultipleLogLevels(t *testing.T) {
	tests := []struct {
		name       string
		logLevel   Level
		verbose    bool
		quiet      bool
		shouldShow map[Level]bool
	}{
		{
			name:     "debug level verbose",
			logLevel: DebugLevel,
			verbose:  true,
			quiet:    false,
			shouldShow: map[Level]bool{
				DebugLevel: true,
				InfoLevel:  true,
				WarnLevel:  true,
				ErrorLevel: true,
			},
		},
		{
			name:     "info level",
			logLevel: InfoLevel,
			verbose:  false,
			quiet:    false,
			shouldShow: map[Level]bool{
				DebugLevel: false,
				InfoLevel:  true,
				WarnLevel:  true,
				ErrorLevel: true,
			},
		},
		{
			name:     "info level quiet",
			logLevel: InfoLevel,
			verbose:  false,
			quiet:    true,
			shouldShow: map[Level]bool{
				DebugLevel: false,
				InfoLevel:  false,
				WarnLevel:  true,
				ErrorLevel: true,
			},
		},
		{
			name:     "error level only",
			logLevel: ErrorLevel,
			verbose:  false,
			quiet:    false,
			shouldShow: map[Level]bool{
				DebugLevel: false,
				InfoLevel:  false,
				WarnLevel:  false,
				ErrorLevel: true,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			buf := &bytes.Buffer{}
			logger := NewLogger(tt.logLevel, buf, tt.verbose, tt.quiet)

			logger.Debug("debug")
			logger.Info("info")
			logger.Warn("warn")
			logger.Error("error")

			output := buf.String()

			if tt.shouldShow[DebugLevel] {
				assert.Contains(t, output, "debug")
			} else {
				assert.NotContains(t, output, "debug")
			}

			if tt.shouldShow[InfoLevel] {
				assert.Contains(t, output, "info")
			} else {
				assert.NotContains(t, output, "info")
			}

			if tt.shouldShow[WarnLevel] {
				assert.Contains(t, output, "warn")
			} else {
				assert.NotContains(t, output, "warn")
			}

			if tt.shouldShow[ErrorLevel] {
				assert.Contains(t, output, "error")
			}
		})
	}
}

func TestInit_MultipleCalls(t *testing.T) {
	// Reset
	Default = nil
	once = sync.Once{}

	// First call
	Init(true, false)
	firstDefault := Default
	assert.NotNil(t, firstDefault)

	// Second call should not change Default (once semantics)
	Init(false, true)
	assert.Equal(t, firstDefault, Default)
	assert.True(t, Default.verbose) // Should still be from first Init

	// Reset
	Default = nil
	once = sync.Once{}
}
