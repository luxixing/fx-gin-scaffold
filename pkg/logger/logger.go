package logger

import (
	"os"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

var (
	// Global logger instance
	logger *zap.Logger
	sugar  *zap.SugaredLogger
)

// Config defines logger configuration
type Config struct {
	Level  string // debug, info, warn, error
	Format string // json, console
	Output string // stdout, stderr, file path
}

// Initialize sets up the global logger
func Initialize(config Config) error {
	var err error
	logger, err = NewLogger(config)
	if err != nil {
		return err
	}

	sugar = logger.Sugar()

	// Replace the global logger
	zap.ReplaceGlobals(logger)

	return nil
}

// NewLogger creates a new zap logger with the given configuration
func NewLogger(config Config) (*zap.Logger, error) {
	// Parse log level
	level, err := zapcore.ParseLevel(config.Level)
	if err != nil {
		level = zapcore.InfoLevel
	}

	// Create encoder config
	encoderConfig := zap.NewProductionEncoderConfig()
	encoderConfig.TimeKey = "timestamp"
	encoderConfig.EncodeTime = zapcore.ISO8601TimeEncoder
	encoderConfig.StacktraceKey = ""

	// Create encoder
	var encoder zapcore.Encoder
	if config.Format == "console" {
		encoder = zapcore.NewConsoleEncoder(encoderConfig)
	} else {
		encoder = zapcore.NewJSONEncoder(encoderConfig)
	}

	// Create writer syncer
	var writeSyncer zapcore.WriteSyncer
	switch config.Output {
	case "stderr":
		writeSyncer = zapcore.Lock(os.Stderr)
	case "stdout", "":
		writeSyncer = zapcore.Lock(os.Stdout)
	default:
		// Assume it's a file path
		file, err := os.OpenFile(config.Output, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
		if err != nil {
			return nil, err
		}
		writeSyncer = zapcore.AddSync(file)
	}

	// Create core
	core := zapcore.NewCore(encoder, writeSyncer, level)

	// Create logger
	logger := zap.New(core, zap.AddCaller(), zap.AddCallerSkip(1))

	return logger, nil
}

// GetLogger returns the global logger instance
func GetLogger() *zap.Logger {
	if logger == nil {
		// Fallback to a default logger if not initialized
		logger, _ = zap.NewProduction()
	}
	return logger
}

// GetSugar returns the global sugared logger instance
func GetSugar() *zap.SugaredLogger {
	if sugar == nil {
		sugar = GetLogger().Sugar()
	}
	return sugar
}

// Sync flushes any buffered log entries
func Sync() {
	if logger != nil {
		_ = logger.Sync()
	}
}

// Convenience functions that use the global logger
func Debug(msg string, fields ...zap.Field) {
	GetLogger().Debug(msg, fields...)
}

func Info(msg string, fields ...zap.Field) {
	GetLogger().Info(msg, fields...)
}

func Warn(msg string, fields ...zap.Field) {
	GetLogger().Warn(msg, fields...)
}

func Error(msg string, fields ...zap.Field) {
	GetLogger().Error(msg, fields...)
}

func Fatal(msg string, fields ...zap.Field) {
	GetLogger().Fatal(msg, fields...)
}

// Sugared convenience functions
func Debugf(template string, args ...any) {
	GetSugar().Debugf(template, args...)
}

func Infof(template string, args ...any) {
	GetSugar().Infof(template, args...)
}

func Warnf(template string, args ...any) {
	GetSugar().Warnf(template, args...)
}

func Errorf(template string, args ...any) {
	GetSugar().Errorf(template, args...)
}

func Fatalf(template string, args ...any) {
	GetSugar().Fatalf(template, args...)
}

func Debugw(msg string, keysAndValues ...any) {
	GetSugar().Debugw(msg, keysAndValues...)
}

func Infow(msg string, keysAndValues ...any) {
	GetSugar().Infow(msg, keysAndValues...)
}

func Warnw(msg string, keysAndValues ...any) {
	GetSugar().Warnw(msg, keysAndValues...)
}

func Errorw(msg string, keysAndValues ...any) {
	GetSugar().Errorw(msg, keysAndValues...)
}

func Fatalw(msg string, keysAndValues ...any) {
	GetSugar().Fatalw(msg, keysAndValues...)
}
