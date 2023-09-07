package database

import (
	"context"
	"log"
	"os"
	"time"

	"github.com/flyteorg/flytestdlib/logger"

	gormLogger "gorm.io/gorm/logger"
)

// GetGormLogger converts between the flytestdlib configured log level to the equivalent gorm log level and outputs
// a gorm/logger implementation accordingly configured.
func GetGormLogger(ctx context.Context, logConfig *logger.Config) gormLogger.Interface {
	logConfigLevel := logger.ErrorLevel
	if logConfig != nil {
		logConfigLevel = logConfig.Level
	} else {
		logger.Debugf(ctx, "No log config block found, setting gorm db log level to: error")
	}
	var logLevel gormLogger.LogLevel
	ignoreRecordNotFoundError := true
	switch logConfigLevel {
	case logger.PanicLevel:
		fallthrough
	case logger.FatalLevel:
		fallthrough
	case logger.ErrorLevel:
		logLevel = gormLogger.Error
	case logger.WarnLevel:
		fallthrough
	case logger.InfoLevel:
		logLevel = gormLogger.Warn
	case logger.DebugLevel:
		logLevel = gormLogger.Info
		ignoreRecordNotFoundError = false
	default:
		logLevel = gormLogger.Silent
	}
	// Copied from gormLogger.Default initialization. The gormLogger interface only allows modifying the LogLevel
	// and not IgnoreRecordNotFoundError.
	return gormLogger.New(log.New(os.Stdout, "\r\n", log.LstdFlags), gormLogger.Config{
		SlowThreshold:             200 * time.Millisecond,
		LogLevel:                  logLevel,
		IgnoreRecordNotFoundError: ignoreRecordNotFoundError,
		Colorful:                  true,
	})
}
