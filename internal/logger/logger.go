package logger

import (
	"fmt"
	"io"
	"log"
	"os"
	"sync"
	"time"
)

// Level represents log level
type Level int

const (
	// DebugLevel for debug messages
	DebugLevel Level = iota
	// InfoLevel for info messages
	InfoLevel
	// WarnLevel for warning messages
	WarnLevel
	// ErrorLevel for error messages
	ErrorLevel
)

// String returns string representation of log level
func (l Level) String() string {
	switch l {
	case DebugLevel:
		return "DEBUG"
	case InfoLevel:
		return "INFO"
	case WarnLevel:
		return "WARN"
	case ErrorLevel:
		return "ERROR"
	default:
		return "UNKNOWN"
	}
}

// Logger provides structured logging
type Logger struct {
	mu      sync.Mutex
	level   Level
	output  io.Writer
	verbose bool
	quiet   bool
	prefix  string
}

var (
	// Default is the default logger instance
	Default *Logger
	once    sync.Once
)

// Init initializes the default logger
func Init(verbose, quiet bool) {
	once.Do(func() {
		level := InfoLevel
		if verbose {
			level = DebugLevel
		}
		Default = &Logger{
			level:   level,
			output:  os.Stderr,
			verbose: verbose,
			quiet:   quiet,
		}
	})
}

// NewLogger creates a new logger with the given configuration
func NewLogger(level Level, output io.Writer, verbose, quiet bool) *Logger {
	return &Logger{
		level:   level,
		output:  output,
		verbose: verbose,
		quiet:   quiet,
	}
}

// WithPrefix returns a new logger with the given prefix
func (l *Logger) WithPrefix(prefix string) *Logger {
	return &Logger{
		level:   l.level,
		output:  l.output,
		verbose: l.verbose,
		quiet:   l.quiet,
		prefix:  prefix,
	}
}

func (l *Logger) log(level Level, format string, v ...interface{}) {
	if l == nil {
		return
	}

	// Skip if level is below threshold
	if level < l.level {
		return
	}

	// Skip info messages if quiet mode
	if l.quiet && level == InfoLevel {
		return
	}

	l.mu.Lock()
	defer l.mu.Unlock()

	timestamp := time.Now().Format("2006-01-02 15:04:05")
	prefix := ""
	if l.prefix != "" {
		prefix = fmt.Sprintf("[%s] ", l.prefix)
	}

	message := fmt.Sprintf(format, v...)
	logLine := fmt.Sprintf("%s [%s] %s%s\n", timestamp, level.String(), prefix, message)

	_, _ = l.output.Write([]byte(logLine))
}

// Debug logs debug message (only when verbose is enabled)
func (l *Logger) Debug(format string, v ...interface{}) {
	l.log(DebugLevel, format, v...)
}

// Info logs info message
func (l *Logger) Info(format string, v ...interface{}) {
	l.log(InfoLevel, format, v...)
}

// Warn logs warning message
func (l *Logger) Warn(format string, v ...interface{}) {
	l.log(WarnLevel, format, v...)
}

// Error logs error message
func (l *Logger) Error(format string, v ...interface{}) {
	l.log(ErrorLevel, format, v...)
}

// Fatal logs error message and exits
func (l *Logger) Fatal(format string, v ...interface{}) {
	l.log(ErrorLevel, format, v...)
	os.Exit(1)
}

// SetLevel sets the log level
func (l *Logger) SetLevel(level Level) {
	l.mu.Lock()
	defer l.mu.Unlock()
	l.level = level
}

// SetOutput sets the output writer
func (l *Logger) SetOutput(output io.Writer) {
	l.mu.Lock()
	defer l.mu.Unlock()
	l.output = output
}

// IsDebug returns true if debug logging is enabled
func (l *Logger) IsDebug() bool {
	return l != nil && l.level == DebugLevel
}

// IsVerbose returns true if verbose mode is enabled
func (l *Logger) IsVerbose() bool {
	return l != nil && l.verbose
}

// IsQuiet returns true if quiet mode is enabled
func (l *Logger) IsQuiet() bool {
	return l != nil && l.quiet
}

// Global convenience functions that use the default logger

// Debug logs debug message using default logger
func Debug(format string, v ...interface{}) {
	if Default != nil {
		Default.Debug(format, v...)
	}
}

// Info logs info message using default logger
func Info(format string, v ...interface{}) {
	if Default != nil {
		Default.Info(format, v...)
	}
}

// Warn logs warning message using default logger
func Warn(format string, v ...interface{}) {
	if Default != nil {
		Default.Warn(format, v...)
	}
}

// Error logs error message using default logger
func Error(format string, v ...interface{}) {
	if Default != nil {
		Default.Error(format, v...)
	}
}

// Fatal logs error message and exits using default logger
func Fatal(format string, v ...interface{}) {
	if Default != nil {
		Default.Fatal(format, v...)
	} else {
		log.Fatalf(format, v...)
	}
}

// WithPrefix returns a new logger with prefix using default logger
func WithPrefix(prefix string) *Logger {
	if Default != nil {
		return Default.WithPrefix(prefix)
	}
	return NewLogger(InfoLevel, os.Stderr, false, false).WithPrefix(prefix)
}
