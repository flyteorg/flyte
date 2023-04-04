// Defines global context-aware logger.
// The default implementation uses logrus. This package registers "logger" config section on init(). The structure of the
// config section is expected to be un-marshal-able to Config struct.
package logger

import (
	"context"
	"fmt"
	"io"
	"runtime"
	"strings"

	"github.com/flyteorg/flytestdlib/contextutils"

	"github.com/sirupsen/logrus"
)

const (
	indentLevelKey contextutils.Key = "LoggerIndentLevel"
	sourceCodeKey  string           = "src"
)

var noopLogger = NoopLogger{}

func onConfigUpdated(cfg Config) {
	logrus.SetLevel(logrus.Level(cfg.Level))

	switch cfg.Formatter.Type {
	case FormatterText:
		if _, isText := logrus.StandardLogger().Formatter.(*logrus.JSONFormatter); !isText {
			logrus.SetFormatter(&logrus.TextFormatter{
				FieldMap: logrus.FieldMap{
					logrus.FieldKeyTime: "ts",
				},
			})
		}
	case FormatterGCP:
		if _, isGCP := logrus.StandardLogger().Formatter.(*GcpFormatter); !isGCP {
			logrus.SetFormatter(&GcpFormatter{})
		}
	default:
		if _, isJSON := logrus.StandardLogger().Formatter.(*logrus.JSONFormatter); !isJSON {
			logrus.SetFormatter(&logrus.JSONFormatter{
				DataKey: jsonDataKey,
				FieldMap: logrus.FieldMap{
					logrus.FieldKeyTime: "ts",
				},
			})
		}
	}
}

func getSourceLocation() string {
	// The reason we pass 3 here: 0 means this function (getSourceLocation), 1 means the getLogger function (only caller
	// to getSourceLocation, 2 means the logging function (e.g. Debugln), and 3 means the caller for the logging function.
	_, file, line, ok := runtime.Caller(3)
	if !ok {
		file = "???"
		line = 1
	} else {
		slash := strings.LastIndex(file, "/")
		if slash >= 0 {
			file = file[slash+1:]
		}
	}

	return fmt.Sprintf("%v:%v", file, line)
}

func getLogger(ctx context.Context) logrus.FieldLogger {
	cfg := GetConfig()
	if cfg.Mute {
		return noopLogger
	}

	entry := logrus.WithFields(logrus.Fields(contextutils.GetLogFields(ctx)))
	if cfg.IncludeSourceCode {
		entry = entry.WithField(sourceCodeKey, getSourceLocation())
	}

	entry.Level = logrus.Level(cfg.Level)

	return entry
}

// Returns a standard io.PipeWriter that logs using the same logger configurations in this package.
func GetLogWriter(ctx context.Context) *io.PipeWriter {
	logger := getLogger(ctx)
	return logger.(*logrus.Entry).Writer()
}

func WithIndent(ctx context.Context, additionalIndent string) context.Context {
	indentLevel := getIndent(ctx) + additionalIndent
	return context.WithValue(ctx, indentLevelKey, indentLevel)
}

func getIndent(ctx context.Context) string {
	if existing := ctx.Value(indentLevelKey); existing != nil {
		return existing.(string)
	}

	return ""
}

// Gets a value indicating whether logs at this level will be written to the logger. This is particularly useful to avoid
// computing log messages unnecessarily.
func IsLoggable(_ context.Context, level Level) bool {
	return GetConfig().Level >= level
}

// Debug logs a message at level Debug on the standard logger.
func Debug(ctx context.Context, args ...interface{}) {
	getLogger(ctx).Debug(args...)
}

// Print logs a message at level Info on the standard logger.
func Print(ctx context.Context, args ...interface{}) {
	getLogger(ctx).Print(args...)
}

// Info logs a message at level Info on the standard logger.
func Info(ctx context.Context, args ...interface{}) {
	getLogger(ctx).Info(args...)
}

// Warn logs a message at level Warn on the standard logger.
func Warn(ctx context.Context, args ...interface{}) {
	getLogger(ctx).Warn(args...)
}

// Warning logs a message at level Warn on the standard logger.
func Warning(ctx context.Context, args ...interface{}) {
	getLogger(ctx).Warning(args...)
}

// Error logs a message at level Error on the standard logger.
func Error(ctx context.Context, args ...interface{}) {
	getLogger(ctx).Error(args...)
}

// Panic logs a message at level Panic on the standard logger.
func Panic(ctx context.Context, args ...interface{}) {
	getLogger(ctx).Panic(args...)
}

// Fatal logs a message at level Fatal on the standard logger.
func Fatal(ctx context.Context, args ...interface{}) {
	getLogger(ctx).Fatal(args...)
}

// Debugf logs a message at level Debug on the standard logger.
func Debugf(ctx context.Context, format string, args ...interface{}) {
	getLogger(ctx).Debugf(format, args...)
}

// Printf logs a message at level Info on the standard logger.
func Printf(ctx context.Context, format string, args ...interface{}) {
	getLogger(ctx).Printf(format, args...)
}

// Infof logs a message at level Info on the standard logger.
func Infof(ctx context.Context, format string, args ...interface{}) {
	getLogger(ctx).Infof(format, args...)
}

// InfofNoCtx logs a formatted message without context.
func InfofNoCtx(format string, args ...interface{}) {
	getLogger(context.TODO()).Infof(format, args...)
}

// Warnf logs a message at level Warn on the standard logger.
func Warnf(ctx context.Context, format string, args ...interface{}) {
	getLogger(ctx).Warnf(format, args...)
}

// Warningf logs a message at level Warn on the standard logger.
func Warningf(ctx context.Context, format string, args ...interface{}) {
	getLogger(ctx).Warningf(format, args...)
}

// Errorf logs a message at level Error on the standard logger.
func Errorf(ctx context.Context, format string, args ...interface{}) {
	getLogger(ctx).Errorf(format, args...)
}

// Panicf logs a message at level Panic on the standard logger.
func Panicf(ctx context.Context, format string, args ...interface{}) {
	getLogger(ctx).Panicf(format, args...)
}

// Fatalf logs a message at level Fatal on the standard logger.
func Fatalf(ctx context.Context, format string, args ...interface{}) {
	getLogger(ctx).Fatalf(format, args...)
}

// Debugln logs a message at level Debug on the standard logger.
func Debugln(ctx context.Context, args ...interface{}) {
	getLogger(ctx).Debugln(args...)
}

// Println logs a message at level Info on the standard logger.
func Println(ctx context.Context, args ...interface{}) {
	getLogger(ctx).Println(args...)
}

// Infoln logs a message at level Info on the standard logger.
func Infoln(ctx context.Context, args ...interface{}) {
	getLogger(ctx).Infoln(args...)
}

// Warnln logs a message at level Warn on the standard logger.
func Warnln(ctx context.Context, args ...interface{}) {
	getLogger(ctx).Warnln(args...)
}

// Warningln logs a message at level Warn on the standard logger.
func Warningln(ctx context.Context, args ...interface{}) {
	getLogger(ctx).Warningln(args...)
}

// Errorln logs a message at level Error on the standard logger.
func Errorln(ctx context.Context, args ...interface{}) {
	getLogger(ctx).Errorln(args...)
}

// Panicln logs a message at level Panic on the standard logger.
func Panicln(ctx context.Context, args ...interface{}) {
	getLogger(ctx).Panicln(args...)
}

// Fatalln logs a message at level Fatal on the standard logger.
func Fatalln(ctx context.Context, args ...interface{}) {
	getLogger(ctx).Fatalln(args...)
}

type NoopLogger struct {
}

func (NoopLogger) WithField(key string, value interface{}) *logrus.Entry {
	return nil
}

func (NoopLogger) WithFields(fields logrus.Fields) *logrus.Entry {
	return nil
}

func (NoopLogger) WithError(err error) *logrus.Entry {
	return nil
}

func (NoopLogger) Debugf(format string, args ...interface{}) {
}

func (NoopLogger) Infof(format string, args ...interface{}) {
}

func (NoopLogger) Warnf(format string, args ...interface{}) {
}

func (NoopLogger) Warningf(format string, args ...interface{}) {
}

func (NoopLogger) Errorf(format string, args ...interface{}) {
}

func (NoopLogger) Debug(args ...interface{}) {
}

func (NoopLogger) Info(args ...interface{}) {
}

func (NoopLogger) Warn(args ...interface{}) {
}

func (NoopLogger) Warning(args ...interface{}) {
}

func (NoopLogger) Error(args ...interface{}) {
}

func (NoopLogger) Debugln(args ...interface{}) {
}

func (NoopLogger) Infoln(args ...interface{}) {
}

func (NoopLogger) Warnln(args ...interface{}) {
}

func (NoopLogger) Warningln(args ...interface{}) {
}

func (NoopLogger) Errorln(args ...interface{}) {
}

func (NoopLogger) Print(...interface{}) {
}

func (NoopLogger) Printf(string, ...interface{}) {
}

func (NoopLogger) Println(...interface{}) {
}

func (NoopLogger) Fatal(...interface{}) {
}

func (NoopLogger) Fatalf(string, ...interface{}) {
}

func (NoopLogger) Fatalln(...interface{}) {
}

func (NoopLogger) Panic(...interface{}) {
}

func (NoopLogger) Panicf(string, ...interface{}) {
}

func (NoopLogger) Panicln(...interface{}) {
}
