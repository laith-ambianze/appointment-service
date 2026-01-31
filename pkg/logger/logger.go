package logger

import (
	"os"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

// Logger wraps zap logger
type Logger struct {
	*zap.Logger
}

// New creates a new logger instance
func New(level, format string) (*Logger, error) {
	var config zap.Config

	// Configure based on format
	if format == "json" {
		config = zap.NewProductionConfig()
	} else {
		config = zap.NewDevelopmentConfig()
		config.EncoderConfig.EncodeLevel = zapcore.CapitalColorLevelEncoder
	}

	// Set log level
	logLevel, err := parseLevel(level)
	if err != nil {
		logLevel = zapcore.InfoLevel
	}
	config.Level = zap.NewAtomicLevelAt(logLevel)

	// Build logger
	zapLogger, err := config.Build(
		zap.AddCaller(),
		zap.AddStacktrace(zapcore.ErrorLevel),
	)
	if err != nil {
		return nil, err
	}

	return &Logger{zapLogger}, nil
}

// NewDefault creates a development logger
func NewDefault() *Logger {
	logger, _ := New("debug", "console")
	return logger
}

// parseLevel converts string level to zapcore.Level
func parseLevel(level string) (zapcore.Level, error) {
	switch level {
	case "debug":
		return zapcore.DebugLevel, nil
	case "info":
		return zapcore.InfoLevel, nil
	case "warn", "warning":
		return zapcore.WarnLevel, nil
	case "error":
		return zapcore.ErrorLevel, nil
	case "fatal":
		return zapcore.FatalLevel, nil
	default:
		return zapcore.InfoLevel, nil
	}
}

// Sync flushes any buffered log entries
func (l *Logger) Sync() error {
	return l.Logger.Sync()
}

// WithFields adds fields to the logger
func (l *Logger) WithFields(fields ...zap.Field) *Logger {
	return &Logger{l.Logger.With(fields...)}
}

// InfoF logs a formatted info message
func (l *Logger) InfoF(template string, args ...interface{}) {
	l.Sugar().Infof(template, args...)
}

// ErrorF logs a formatted error message
func (l *Logger) ErrorF(template string, args ...interface{}) {
	l.Sugar().Errorf(template, args...)
}

// DebugF logs a formatted debug message
func (l *Logger) DebugF(template string, args ...interface{}) {
	l.Sugar().Debugf(template, args...)
}

// WarnF logs a formatted warning message
func (l *Logger) WarnF(template string, args ...interface{}) {
	l.Sugar().Warnf(template, args...)
}

// FatalF logs a formatted fatal message and exits
func (l *Logger) FatalF(template string, args ...interface{}) {
	l.Sugar().Fatalf(template, args...)
	os.Exit(1)
}
