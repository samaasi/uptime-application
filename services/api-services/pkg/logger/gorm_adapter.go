package logger

import (
	"context"
	"fmt"
	"time"

	"go.uber.org/zap"
	"gorm.io/gorm"
	gormLogger "gorm.io/gorm/logger"
)

// GormZapLogger adapts Zap to GORM's logger.Interface
type GormZapLogger struct {
	cfg      gormLogger.Config
	logLevel gormLogger.LogLevel
}

// NewGormLogger creates a new Zap-backed GORM logger adapter
func NewGormLogger(cfg gormLogger.Config) gormLogger.Interface {
	return &GormZapLogger{
		cfg:      cfg,
		logLevel: cfg.LogLevel,
	}
}

// LogMode sets the logging level and returns a new logger
func (l *GormZapLogger) LogMode(level gormLogger.LogLevel) gormLogger.Interface {
	nl := *l
	nl.logLevel = level
	nl.cfg.LogLevel = level
	return &nl
}

func (l *GormZapLogger) Info(ctx context.Context, msg string, data ...interface{}) {
	if l.logLevel < gormLogger.Info {
		return
	}
	Get().WithOptions(zap.AddCallerSkip(1)).Info(fmt.Sprintf(msg, data...))
}

func (l *GormZapLogger) Warn(ctx context.Context, msg string, data ...interface{}) {
	if l.logLevel < gormLogger.Warn {
		return
	}
	Get().WithOptions(zap.AddCallerSkip(1)).Warn(fmt.Sprintf(msg, data...))
}

func (l *GormZapLogger) Error(ctx context.Context, msg string, data ...interface{}) {
	if l.logLevel < gormLogger.Error {
		return
	}
	Get().WithOptions(zap.AddCallerSkip(1)).Error(fmt.Sprintf(msg, data...))
}

// Trace logs SQL execution details with elapsed time, rows affected, and error
func (l *GormZapLogger) Trace(ctx context.Context, begin time.Time, fc func() (string, int64), err error) {
	elapsed := time.Since(begin)
	sql, rows := fc()

	if err == gorm.ErrRecordNotFound && l.cfg.IgnoreRecordNotFoundError {
		return
	}

	if err != nil {
		if l.logLevel >= gormLogger.Error {
			Get().WithOptions(zap.AddCallerSkip(1)).Error("gorm query error",
				zap.Error(err),
				zap.String("sql", sql),
				zap.Int64("rows", rows),
				zap.Duration("elapsed", elapsed),
			)
		}
		return
	}

	if l.cfg.SlowThreshold > 0 && elapsed > l.cfg.SlowThreshold {
		if l.logLevel >= gormLogger.Warn {
			Get().WithOptions(zap.AddCallerSkip(1)).Warn("gorm slow query",
				zap.String("sql", sql),
				zap.Int64("rows", rows),
				zap.Duration("elapsed", elapsed),
			)
		}
		return
	}

	if l.logLevel >= gormLogger.Info {
		Get().WithOptions(zap.AddCallerSkip(1)).Info("gorm query",
			zap.String("sql", sql),
			zap.Int64("rows", rows),
			zap.Duration("elapsed", elapsed),
		)
	}
}
