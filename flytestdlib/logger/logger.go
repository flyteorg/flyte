// Defines global context-aware logger.
// The default implementation uses logrus. This package registers "logger" config section on init(). The structure of the
// config section is expected to be un-marshal-able to Config struct.
package logger

import (
	"context"

	"github.com/lyft/flytestdlib/contextutils"

	"fmt"
	"runtime"
	"strings"

	"github.com/sirupsen/logrus"
)

//go:generate gotests -w -all $FILE

const indentLevelKey contextutils.Key = "LoggerIndentLevel"

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
	if globalConfig.IncludeSourceCode {
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

		return fmt.Sprintf("[%v:%v] ", file, line)
	}

	return ""
}

func wrapHeader(ctx context.Context, args ...interface{}) []interface{} {
	args = append([]interface{}{getIndent(ctx)}, args...)

	if globalConfig.IncludeSourceCode {
		return append(
			[]interface{}{
				fmt.Sprintf("%v", getSourceLocation()),
			},
			args...)
	}

	return args
}

func wrapHeaderForMessage(ctx context.Context, message string) string {
	message = fmt.Sprintf("%v%v", getIndent(ctx), message)

	if globalConfig.IncludeSourceCode {
		return fmt.Sprintf("%v%v", getSourceLocation(), message)
	}

	return message
}

func getLogger(ctx context.Context) *logrus.Entry {
	entry := logrus.WithFields(logrus.Fields(contextutils.GetLogFields(ctx)))
	entry.Level = logrus.Level(globalConfig.Level)
	return entry
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
func IsLoggable(ctx context.Context, level Level) bool {
	return getLogger(ctx).Level >= logrus.Level(level)
}

// Debug logs a message at level Debug on the standard logger.
func Debug(ctx context.Context, args ...interface{}) {
	if globalConfig.Mute {
		return
	}

	getLogger(ctx).Debug(wrapHeader(ctx, args)...)
}

// Print logs a message at level Info on the standard logger.
func Print(ctx context.Context, args ...interface{}) {
	if globalConfig.Mute {
		return
	}

	getLogger(ctx).Print(wrapHeader(ctx, args)...)
}

// Info logs a message at level Info on the standard logger.
func Info(ctx context.Context, args ...interface{}) {
	if globalConfig.Mute {
		return
	}

	getLogger(ctx).Info(wrapHeader(ctx, args)...)
}

// Warn logs a message at level Warn on the standard logger.
func Warn(ctx context.Context, args ...interface{}) {
	if globalConfig.Mute {
		return
	}

	getLogger(ctx).Warn(wrapHeader(ctx, args)...)
}

// Warning logs a message at level Warn on the standard logger.
func Warning(ctx context.Context, args ...interface{}) {
	if globalConfig.Mute {
		return
	}

	getLogger(ctx).Warning(wrapHeader(ctx, args)...)
}

// Error logs a message at level Error on the standard logger.
func Error(ctx context.Context, args ...interface{}) {
	if globalConfig.Mute {
		return
	}

	getLogger(ctx).Error(wrapHeader(ctx, args)...)
}

// Panic logs a message at level Panic on the standard logger.
func Panic(ctx context.Context, args ...interface{}) {
	if globalConfig.Mute {
		return
	}

	getLogger(ctx).Panic(wrapHeader(ctx, args)...)
}

// Fatal logs a message at level Fatal on the standard logger.
func Fatal(ctx context.Context, args ...interface{}) {
	if globalConfig.Mute {
		return
	}

	getLogger(ctx).Fatal(wrapHeader(ctx, args)...)
}

// Debugf logs a message at level Debug on the standard logger.
func Debugf(ctx context.Context, format string, args ...interface{}) {
	if globalConfig.Mute {
		return
	}

	getLogger(ctx).Debugf(wrapHeaderForMessage(ctx, format), args...)
}

// Printf logs a message at level Info on the standard logger.
func Printf(ctx context.Context, format string, args ...interface{}) {
	if globalConfig.Mute {
		return
	}

	getLogger(ctx).Printf(wrapHeaderForMessage(ctx, format), args...)
}

// Infof logs a message at level Info on the standard logger.
func Infof(ctx context.Context, format string, args ...interface{}) {
	if globalConfig.Mute {
		return
	}

	getLogger(ctx).Infof(wrapHeaderForMessage(ctx, format), args...)
}

// InfofNoCtx logs a formatted message without context.
func InfofNoCtx(format string, args ...interface{}) {
	if globalConfig.Mute {
		return
	}

	getLogger(context.TODO()).Infof(format, args...)
}

// Warnf logs a message at level Warn on the standard logger.
func Warnf(ctx context.Context, format string, args ...interface{}) {
	if globalConfig.Mute {
		return
	}

	getLogger(ctx).Warnf(wrapHeaderForMessage(ctx, format), args...)
}

// Warningf logs a message at level Warn on the standard logger.
func Warningf(ctx context.Context, format string, args ...interface{}) {
	if globalConfig.Mute {
		return
	}

	getLogger(ctx).Warningf(wrapHeaderForMessage(ctx, format), args...)
}

// Errorf logs a message at level Error on the standard logger.
func Errorf(ctx context.Context, format string, args ...interface{}) {
	if globalConfig.Mute {
		return
	}

	getLogger(ctx).Errorf(wrapHeaderForMessage(ctx, format), args...)
}

// Panicf logs a message at level Panic on the standard logger.
func Panicf(ctx context.Context, format string, args ...interface{}) {
	if globalConfig.Mute {
		return
	}

	getLogger(ctx).Panicf(wrapHeaderForMessage(ctx, format), args...)
}

// Fatalf logs a message at level Fatal on the standard logger.
func Fatalf(ctx context.Context, format string, args ...interface{}) {
	if globalConfig.Mute {
		return
	}

	getLogger(ctx).Fatalf(wrapHeaderForMessage(ctx, format), args...)
}

// Debugln logs a message at level Debug on the standard logger.
func Debugln(ctx context.Context, args ...interface{}) {
	if globalConfig.Mute {
		return
	}

	getLogger(ctx).Debugln(wrapHeader(ctx, args)...)
}

// Println logs a message at level Info on the standard logger.
func Println(ctx context.Context, args ...interface{}) {
	if globalConfig.Mute {
		return
	}

	getLogger(ctx).Println(wrapHeader(ctx, args)...)
}

// Infoln logs a message at level Info on the standard logger.
func Infoln(ctx context.Context, args ...interface{}) {
	if globalConfig.Mute {
		return
	}

	getLogger(ctx).Infoln(wrapHeader(ctx, args)...)
}

// Warnln logs a message at level Warn on the standard logger.
func Warnln(ctx context.Context, args ...interface{}) {
	if globalConfig.Mute {
		return
	}

	getLogger(ctx).Warnln(wrapHeader(ctx, args)...)
}

// Warningln logs a message at level Warn on the standard logger.
func Warningln(ctx context.Context, args ...interface{}) {
	if globalConfig.Mute {
		return
	}

	getLogger(ctx).Warningln(wrapHeader(ctx, args)...)
}

// Errorln logs a message at level Error on the standard logger.
func Errorln(ctx context.Context, args ...interface{}) {
	if globalConfig.Mute {
		return
	}

	getLogger(ctx).Errorln(wrapHeader(ctx, args)...)
}

// Panicln logs a message at level Panic on the standard logger.
func Panicln(ctx context.Context, args ...interface{}) {
	if globalConfig.Mute {
		return
	}

	getLogger(ctx).Panicln(wrapHeader(ctx, args)...)
}

// Fatalln logs a message at level Fatal on the standard logger.
func Fatalln(ctx context.Context, args ...interface{}) {
	if globalConfig.Mute {
		return
	}

	getLogger(ctx).Fatalln(wrapHeader(ctx, args)...)
}
