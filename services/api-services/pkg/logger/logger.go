package logger

import (
	"fmt"
	"os"
	"strings"
	"sync"
	"time"

	"github.com/samaasi/uptime-application/services/api-services/internal/config"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"gopkg.in/natefinch/lumberjack.v2"
)

var (
	globalLogger *zap.Logger
	once         sync.Once
)

// Field is a type alias for zap.Field, allowing external packages to use logger.Field
// without directly importing go.uber.org/zap.
type Field = zap.Field

// LoggerConfig holds configuration for the logger
type LoggerConfig struct {
	Level       string   `json:"level" yaml:"level"`
	Encoding    string   `json:"encoding" yaml:"encoding"`
	Development bool     `json:"development" yaml:"development"`
	StackTrace  bool     `json:"stackTrace" yaml:"stackTrace"`
	Caller      bool     `json:"caller" yaml:"caller"`
	OutputPaths []string `json:"outputPaths" yaml:"outputPaths"`

	// Configuration for log rotation
	MaxSize    int  `json:"maxSize" yaml:"maxSize"`
	MaxBackups int  `json:"maxBackups" yaml:"maxBackups"`
	MaxAge     int  `json:"maxAge" yaml:"maxAge"`
	Compress   bool `json:"compress" yaml:"compress"`
}

// DefaultConfig returns a default logger configuration
func DefaultConfig() *LoggerConfig {
	return &LoggerConfig{
		Level:       "info",
		Encoding:    "json",
		Development: false,
		StackTrace:  true,
		Caller:      true,
		OutputPaths: []string{"stdout"},
		MaxSize:     100,
		MaxBackups:  3,
		MaxAge:      30,
		Compress:    true,
	}
}

// NewLogger creates a new logger instance with the given configuration
func NewLogger(cfg *LoggerConfig) (*zap.Logger, error) {
	if cfg == nil {
		cfg = DefaultConfig()
	}

	return InitGlobalLogger(cfg)
}

// InitFromConfig initializes the logger from the unified application config
func InitFromConfig(cfg config.LoggingConfig) error {
	var initErr error
	once.Do(func() {
		logLevel, err := zap.ParseAtomicLevel(cfg.Level)
		if err != nil {
			initErr = fmt.Errorf("failed to parse log level: %w", err)
			return
		}

		encoderConfig := zap.NewProductionEncoderConfig()
		if cfg.Development {
			encoderConfig = zap.NewDevelopmentEncoderConfig()
		}
		encoderConfig.TimeKey = "ts"
		encoderConfig.EncodeTime = zapcore.ISO8601TimeEncoder
		encoderConfig.EncodeLevel = zapcore.LowercaseLevelEncoder
		encoderConfig.CallerKey = "caller"
		encoderConfig.EncodeCaller = zapcore.ShortCallerEncoder
		encoderConfig.MessageKey = "msg"
		encoderConfig.StacktraceKey = "stacktrace"
		encoderConfig.LineEnding = zapcore.DefaultLineEnding
		encoderConfig.FunctionKey = zapcore.OmitKey

		if cfg.Development {
			encoderConfig.EncodeLevel = zapcore.CapitalColorLevelEncoder
		}

		encoder := zapcore.NewJSONEncoder(encoderConfig)
		if cfg.Encoding == "console" {
			encoder = zapcore.NewConsoleEncoder(encoderConfig)
		}

		var writers []zapcore.WriteSyncer
		for _, path := range cfg.OutputPaths {
			if strings.EqualFold(path, "stdout") || strings.EqualFold(path, "stderr") {
				writers = append(writers, zapcore.Lock(os.Stdout))
			} else {
				// Implement log rotation with Lumberjack for file outputs
				lj := &lumberjack.Logger{
					Filename:   path,
					MaxSize:    cfg.MaxSize,
					MaxBackups: cfg.MaxBackups,
					MaxAge:     cfg.MaxAge,
					Compress:   cfg.Compress,
				}
				writers = append(writers, zapcore.AddSync(lj))
			}
		}
		syncer := zapcore.NewMultiWriteSyncer(writers...)

		core := zapcore.NewCore(encoder, syncer, logLevel)

		options := []zap.Option{zap.ErrorOutput(zapcore.AddSync(os.Stderr))}
		if cfg.Caller {
			options = append(options, zap.AddCaller())
		}
		if cfg.StackTrace {
			options = append(options, zap.AddStacktrace(zap.ErrorLevel))
		}

		logger := zap.New(core, options...)

		globalLogger = logger
		zap.ReplaceGlobals(logger)
	})
	return initErr
}

func parseZapLevel(levelStr string) zapcore.Level {
	switch strings.ToLower(levelStr) {
	case "debug":
		return zap.DebugLevel
	case "info":
		return zap.InfoLevel
	case "warn":
		return zap.WarnLevel
	case "error":
		return zap.ErrorLevel
	case "fatal":
		return zap.FatalLevel
	default:
		return zap.InfoLevel
	}
}

// InitGlobalLogger initializes the global logger instance
func InitGlobalLogger(cfg *LoggerConfig) (*zap.Logger, error) {
	// This function can be simplified or deprecated in favor of InitFromConfig
	// For now, we ensure it doesn't error out.
	var logger *zap.Logger
	var err error
	once.Do(func() {
		// Simplified init, should be adapted to use full features if needed
		logger, err = NewLogger(cfg)
		if err != nil {
			return
		}
		globalLogger = logger
		zap.ReplaceGlobals(logger)
	})
	return globalLogger, err
}

// Sync flushes any buffered log entries
func Sync() error {
	if globalLogger == nil {
		return nil
	}
	return globalLogger.Sync()
}

// Get returns the global logger instance
func Get() *zap.Logger {
	if globalLogger == nil {
		_, _ = InitGlobalLogger(DefaultConfig())
	}

	return globalLogger
}

// Wrapper functions for common log levels with structured logging support
// Using WithOptions to set the correct caller skip for these wrapper functions.

func Debug(msg string, fields ...zap.Field) {
	Get().WithOptions(zap.AddCallerSkip(1)).Debug(msg, fields...)
}

func Info(msg string, fields ...zap.Field) {
	Get().WithOptions(zap.AddCallerSkip(1)).Info(msg, fields...)
}

func Warn(msg string, fields ...zap.Field) {
	Get().WithOptions(zap.AddCallerSkip(1)).Warn(msg, fields...)
}

func Error(msg string, fields ...zap.Field) {
	Get().WithOptions(zap.AddCallerSkip(1)).Error(msg, fields...)
}

func Fatal(msg string, fields ...zap.Field) {
	Get().WithOptions(zap.AddCallerSkip(1)).Fatal(msg, fields...)
}

// Helper functions for common field types

func String(key, val string) zap.Field {
	return zap.String(key, val)
}

func Int(key string, val int) zap.Field {
	return zap.Int(key, val)
}

func Int32(key string, val int32) zap.Field {
	return zap.Int32(key, val)
}

func Int64(key string, val int64) zap.Field {
	return zap.Int64(key, val)
}

func Float64(key string, val float64) zap.Field {
	return zap.Float64(key, val)
}

func Bool(key string, val bool) zap.Field {
	return zap.Bool(key, val)
}

func Duration(key string, val time.Duration) zap.Field {
	return zap.Duration(key, val)
}

func Time(key string, val time.Time) zap.Field {
	return zap.Time(key, val)
}

func ErrorField(err error) zap.Field {
	return zap.Error(err)
}

func Any(key string, val interface{}) Field {
	return zap.Any(key, val)
}

func Object(key string, val interface{}) Field {
	return Any(key, val)
}

// With creates a child logger and adds structured context to it
func With(fields ...zap.Field) *zap.Logger {
	return Get().With(fields...)
}

func IsBrokenPipeError(err error) bool {
	if err == nil {
		return false
	}

	return strings.Contains(err.Error(), "The handle is invalid") ||
		strings.Contains(err.Error(), "broken pipe") ||
		strings.Contains(err.Error(), "sync /dev/stdout")
}
