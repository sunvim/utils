package logger

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"sync/atomic"
	"time"

	"github.com/rs/zerolog"
)

// Logger interface defines the logging methods
type Logger interface {
	Debug(msg string)
	Debugf(format string, args ...interface{})
	Info(msg string)
	Infof(format string, args ...interface{})
	Warn(msg string)
	Warnf(format string, args ...interface{})
	Error(msg string)
	Errorf(format string, args ...interface{})
	WithField(key string, value interface{}) Logger
	WithFields(fields map[string]interface{}) Logger
	WithError(err error) Logger
}

type LoggingConfig struct {
	Level      string `mapstructure:"level"`
	Format     string `mapstructure:"format"`
	Output     string `mapstructure:"output"` // stdout, stderr, file
	FilePath   string `mapstructure:"file_path"`
	MaxSize    int    `mapstructure:"max_size"`
	MaxBackups int    `mapstructure:"max_backups"`
	MaxAge     int    `mapstructure:"max_age"`
}

func DefaultLoggingConfig() LoggingConfig {

	return LoggingConfig{
		Level:      "info",
		Format:     "text",
		Output:     "stdout",
		FilePath:   "app.log",
		MaxSize:    100, // megabytes
		MaxBackups: 7,
		MaxAge:     30, // days
	}
}

// ZerologLogger implements Logger interface using zerolog
type ZerologLogger struct {
	logger zerolog.Logger
}

// getRelativeCaller returns the caller information with relative path
func getRelativeCaller(skip int) string {
	_, file, line, ok := runtime.Caller(skip + 2) // +2 to skip this function and the logger method
	if !ok {
		return "unknown:0"
	}

	// Try to find project root by looking for go.mod file
	projectRoot := findProjectRoot(file)
	if projectRoot == "" {
		// Fallback to current working directory
		wd, err := os.Getwd()
		if err != nil {
			return filepath.Base(file) + ":" + fmt.Sprintf("%d", line)
		}
		projectRoot = wd
	}

	relPath, err := filepath.Rel(projectRoot, file)
	if err != nil {
		return filepath.Base(file) + ":" + fmt.Sprintf("%d", line)
	}

	// If the relative path starts with "../", it means the file is outside the project
	// In this case, just use the base name
	if strings.HasPrefix(relPath, "../") {
		return filepath.Base(file) + ":" + fmt.Sprintf("%d", line)
	}

	return relPath + ":" + fmt.Sprintf("%d", line)
}

// findProjectRoot finds the project root by looking for go.mod file
func findProjectRoot(startPath string) string {
	dir := filepath.Dir(startPath)
	for {
		goModPath := filepath.Join(dir, "go.mod")
		if _, err := os.Stat(goModPath); err == nil {
			return dir
		}

		parent := filepath.Dir(dir)
		if parent == dir {
			// Reached root directory
			break
		}
		dir = parent
	}
	return ""
}

// NewZerologLogger creates a new ZerologLogger instance
func NewZerologLogger(logger zerolog.Logger) *ZerologLogger {
	return &ZerologLogger{
		logger: logger,
	}
}

// LoggerManager manages logger instances and configuration
type LoggerManager struct {
	config        *LoggingConfig
	logger        zerolog.Logger
	zerologLogger *ZerologLogger
	outputFile    *os.File
}

// NewLoggerManager creates a new logger manager with the given configuration
func NewLoggerManager(cfg *LoggingConfig) (*LoggerManager, error) {
	if cfg == nil {
		return nil, fmt.Errorf("logging config is required")
	}

	m := &LoggerManager{
		config: cfg,
	}

	if err := m.initialize(); err != nil {
		return nil, err
	}

	return m, nil
}

// initialize sets up the logger based on configuration
func (m *LoggerManager) initialize() error {
	// Parse log level
	level, err := m.parseLevel(m.config.Level)
	if err != nil {
		return fmt.Errorf("invalid log level '%s': %w", m.config.Level, err)
	}

	// Set global log level
	zerolog.SetGlobalLevel(level)

	// Get output writer
	writer, err := m.getWriter()
	if err != nil {
		return fmt.Errorf("failed to get writer: %w", err)
	}

	// Create logger based on format
	var logger zerolog.Logger
	switch m.config.Format {
	case "json":
		logger = zerolog.New(writer).With().Timestamp().Logger()
	case "text", "console":
		// Use console writer for human-readable output
		consoleWriter := zerolog.ConsoleWriter{
			Out:        writer,
			TimeFormat: time.RFC3339,
		}
		logger = zerolog.New(consoleWriter).With().Timestamp().Logger()
	default:
		return fmt.Errorf("unsupported log format: %s", m.config.Format)
	}

	m.logger = logger
	m.zerologLogger = NewZerologLogger(logger)

	return nil
}

// parseLevel converts string level to zerolog.Level
func (m *LoggerManager) parseLevel(level string) (zerolog.Level, error) {
	switch level {
	case "trace":
		return zerolog.TraceLevel, nil
	case "debug":
		return zerolog.DebugLevel, nil
	case "info":
		return zerolog.InfoLevel, nil
	case "warn", "warning":
		return zerolog.WarnLevel, nil
	case "error":
		return zerolog.ErrorLevel, nil
	case "fatal":
		return zerolog.FatalLevel, nil
	case "panic":
		return zerolog.PanicLevel, nil
	case "disabled":
		return zerolog.Disabled, nil
	default:
		return zerolog.InfoLevel, fmt.Errorf("invalid log level: %s", level)
	}
}

// getWriter returns the appropriate writer based on output configuration
func (m *LoggerManager) getWriter() (io.Writer, error) {
	switch m.config.Output {
	case "stdout":
		return os.Stdout, nil
	case "stderr":
		return os.Stderr, nil
	case "file":
		if m.config.FilePath == "" {
			return nil, fmt.Errorf("file path is required for file output")
		}
		file, err := os.OpenFile(m.config.FilePath, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0666)
		if err != nil {
			return nil, fmt.Errorf("failed to open log file: %w", err)
		}
		m.outputFile = file
		return file, nil
	default:
		return nil, fmt.Errorf("unsupported log output: %s", m.config.Output)
	}
}

// GetLogger returns the logger instance
func (m *LoggerManager) GetLogger() Logger {
	return m.zerologLogger
}

// SetLevel changes the log level
func (m *LoggerManager) SetLevel(level string) error {
	zerologLevel, err := m.parseLevel(level)
	if err != nil {
		return err
	}
	zerolog.SetGlobalLevel(zerologLevel)
	return nil
}

// Close closes the logger and flushes any buffered log entries
func (m *LoggerManager) Close() error {
	if m.outputFile != nil {
		return m.outputFile.Close()
	}
	return nil
}

// ZerologLogger method implementations

// Debug logs a debug message
func (l *ZerologLogger) Debug(msg string) {
	l.logger.Debug().Str("caller", getRelativeCaller(0)).Msg(msg)
}

// Debugf logs a formatted debug message
func (l *ZerologLogger) Debugf(format string, args ...interface{}) {
	l.logger.Debug().Str("caller", getRelativeCaller(0)).Msgf(format, args...)
}

// Info logs an info message
func (l *ZerologLogger) Info(msg string) {
	l.logger.Info().Str("caller", getRelativeCaller(0)).Msg(msg)
}

// Infof logs a formatted info message
func (l *ZerologLogger) Infof(format string, args ...interface{}) {
	l.logger.Info().Str("caller", getRelativeCaller(0)).Msgf(format, args...)
}

// Warn logs a warning message
func (l *ZerologLogger) Warn(msg string) {
	l.logger.Warn().Str("caller", getRelativeCaller(0)).Msg(msg)
}

// Warnf logs a formatted warning message
func (l *ZerologLogger) Warnf(format string, args ...interface{}) {
	l.logger.Warn().Str("caller", getRelativeCaller(0)).Msgf(format, args...)
}

// Error logs an error message
func (l *ZerologLogger) Error(msg string) {
	l.logger.Error().Str("caller", getRelativeCaller(0)).Msg(msg)
}

// Errorf logs a formatted error message
func (l *ZerologLogger) Errorf(format string, args ...interface{}) {
	l.logger.Error().Str("caller", getRelativeCaller(0)).Msgf(format, args...)
}

// WithField adds a field to the logger
func (l *ZerologLogger) WithField(key string, value interface{}) Logger {
	newLogger := l.logger.With().Interface(key, value).Logger()
	return &ZerologLogger{logger: newLogger}
}

// WithFields adds multiple fields to the logger
func (l *ZerologLogger) WithFields(fields map[string]interface{}) Logger {
	ctx := l.logger.With()
	for key, value := range fields {
		ctx = ctx.Interface(key, value)
	}
	newLogger := ctx.Logger()
	return &ZerologLogger{logger: newLogger}
}

// WithError adds an error field to the logger
func (l *ZerologLogger) WithError(err error) Logger {
	newLogger := l.logger.With().Err(err).Logger()
	return &ZerologLogger{logger: newLogger}
}

// LoggerHook interface for custom hooks
type LoggerHook interface {
	Fire(level zerolog.Level, msg string) error
}

// MetricsHook tracks log metrics
type MetricsHook struct {
	errorCount int64
	warnCount  int64
	infoCount  int64
	debugCount int64
}

// NewMetricsHook creates a new metrics hook
func NewMetricsHook() *MetricsHook {
	return &MetricsHook{}
}

// Fire is called when a log entry is made
func (h *MetricsHook) Fire(level zerolog.Level, msg string) error {
	switch level {
	case zerolog.ErrorLevel:
		atomic.AddInt64(&h.errorCount, 1)
	case zerolog.WarnLevel:
		atomic.AddInt64(&h.warnCount, 1)
	case zerolog.InfoLevel:
		atomic.AddInt64(&h.infoCount, 1)
	case zerolog.DebugLevel:
		atomic.AddInt64(&h.debugCount, 1)
	}
	return nil
}

// GetCounts returns the current counts
func (h *MetricsHook) GetCounts() (errors, warns, infos, debugs int64) {
	return atomic.LoadInt64(&h.errorCount),
		atomic.LoadInt64(&h.warnCount),
		atomic.LoadInt64(&h.infoCount),
		atomic.LoadInt64(&h.debugCount)
}

// RequestIDHook transforms request_id to req_id
type RequestIDHook struct{}

// NewRequestIDHook creates a new request ID hook
func NewRequestIDHook() *RequestIDHook {
	return &RequestIDHook{}
}

// Fire transforms request_id field to req_id
func (h *RequestIDHook) Fire(level zerolog.Level, msg string) error {
	// In zerolog, field transformation would be handled differently
	// This is a simplified implementation for compatibility
	return nil
}

// Global logger management
var (
	globalLoggerManager *LoggerManager
	globalLogger        Logger
)

// InitLogger initializes the global logger
func InitLogger(cfg *LoggingConfig) error {
	manager, err := NewLoggerManager(cfg)
	if err != nil {
		return err
	}

	globalLoggerManager = manager
	globalLogger = manager.GetLogger()
	return nil
}

// GetGlobalLogger returns the global logger instance
func GetGlobalLogger() Logger {
	if globalLogger == nil {
		// Create a basic logger if not initialized
		logger := zerolog.New(os.Stderr).With().Timestamp().Logger()
		return NewZerologLogger(logger)
	}
	return globalLogger
}

// GetLoggerWithComponent returns a logger with component field
func GetLoggerWithComponent(component string) Logger {
	return GetGlobalLogger().WithField("component", component)
}

// GetLoggerWithModule returns a logger with module field
func GetLoggerWithModule(module string) Logger {
	return GetGlobalLogger().WithField("module", module)
}

// SetGlobalLogLevel sets the global log level
func SetGlobalLogLevel(level string) error {
	if globalLoggerManager != nil {
		return globalLoggerManager.SetLevel(level)
	}
	return fmt.Errorf("logger manager not initialized")
}

// AddHook adds a hook to the global logger
func AddHook(hook LoggerHook) error {
	if globalLoggerManager == nil {
		return fmt.Errorf("logger manager not initialized")
	}
	// In zerolog, hooks are handled differently
	// This is a placeholder for compatibility
	return nil
}

// CloseGlobalLogger closes the global logger
func CloseGlobalLogger() error {
	if globalLoggerManager != nil {
		err := globalLoggerManager.Close()
		globalLoggerManager = nil
		globalLogger = nil
		return err
	}
	return nil
}
