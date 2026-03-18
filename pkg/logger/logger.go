package logger

import (
	"os"
	"time"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

// Logger wraps zap logger
type Logger struct {
	*zap.Logger
}

// New creates a new logger instance
func New(level, format string) (*Logger, error) {
	// Set log level
	logLevel, err := parseLevel(level)
	if err != nil {
		logLevel = zapcore.InfoLevel
	}

	var encoder zapcore.Encoder
	var config zap.Config

	if format == "json" {
		// JSON format for production/log aggregation
		config = zap.NewProductionConfig()
		config.Level = zap.NewAtomicLevelAt(logLevel)
		zapLogger, err := config.Build(
			zap.AddCaller(),
			zap.AddStacktrace(zapcore.ErrorLevel),
		)
		if err != nil {
			return nil, err
		}
		return &Logger{zapLogger}, nil
	}

	// Console format - human readable for Docker Desktop / local development
	encoderConfig := zapcore.EncoderConfig{
		TimeKey:       "time",
		LevelKey:      "level",
		NameKey:       "logger",
		CallerKey:     "caller",
		FunctionKey:   zapcore.OmitKey,
		MessageKey:    "msg",
		StacktraceKey: "stacktrace",
		LineEnding:    zapcore.DefaultLineEnding,
		// Human-readable time format
		EncodeTime: func(t time.Time, enc zapcore.PrimitiveArrayEncoder) {
			enc.AppendString(t.Format("15:04:05"))
		},
		// Clean level format with padding for alignment
		EncodeLevel: func(l zapcore.Level, enc zapcore.PrimitiveArrayEncoder) {
			switch l {
			case zapcore.DebugLevel:
				enc.AppendString("[DEBUG]")
			case zapcore.InfoLevel:
				enc.AppendString("[INFO] ")
			case zapcore.WarnLevel:
				enc.AppendString("[WARN] ")
			case zapcore.ErrorLevel:
				enc.AppendString("[ERROR]")
			case zapcore.FatalLevel:
				enc.AppendString("[FATAL]")
			default:
				enc.AppendString("[" + l.CapitalString() + "]")
			}
		},
		EncodeDuration: zapcore.StringDurationEncoder,
		EncodeCaller:   zapcore.ShortCallerEncoder,
	}

	encoder = zapcore.NewConsoleEncoder(encoderConfig)

	core := zapcore.NewCore(
		encoder,
		zapcore.AddSync(os.Stdout),
		zap.NewAtomicLevelAt(logLevel),
	)

	zapLogger := zap.New(core,
		zap.AddCaller(),
		zap.AddStacktrace(zapcore.ErrorLevel),
	)

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
