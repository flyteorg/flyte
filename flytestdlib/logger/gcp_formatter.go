package logger

import (
	"encoding/json"
	"fmt"
	"time"

	"github.com/sirupsen/logrus"
)

// GcpFormatter Log formatter compatible with GCP stackdriver logging.
type GcpFormatter struct {
}

type GcpEntry struct {
	Data      map[string]interface{} `json:"data,omitempty"`
	Message   string                 `json:"message,omitempty"`
	Severity  GcpSeverity            `json:"severity,omitempty"`
	Timestamp string                 `json:"timestamp,omitempty"`
}

type GcpSeverity = string

// https://cloud.google.com/logging/docs/reference/v2/rest/v2/LogEntry#LogSeverity
const (
	GcpSeverityDebug    GcpSeverity = "DEBUG"
	GcpSeverityInfo     GcpSeverity = "INFO"
	GcpSeverityWarning  GcpSeverity = "WARNING"
	GcpSeverityError    GcpSeverity = "ERROR"
	GcpSeverityCritical GcpSeverity = "CRITICAL"
	GcpSeverityAlert    GcpSeverity = "ALERT"
)

var (
	logrusToGcp = map[logrus.Level]GcpSeverity{
		logrus.DebugLevel: GcpSeverityDebug,
		logrus.InfoLevel:  GcpSeverityInfo,
		logrus.WarnLevel:  GcpSeverityWarning,
		logrus.ErrorLevel: GcpSeverityError,
		logrus.FatalLevel: GcpSeverityCritical,
		logrus.PanicLevel: GcpSeverityAlert,
	}
)

func toGcpSeverityOrDefault(level logrus.Level) GcpSeverity {
	if severity, found := logrusToGcp[level]; found {
		return severity
	}
	// default value
	return GcpSeverityDebug
}

func (f *GcpFormatter) Format(entry *logrus.Entry) ([]byte, error) {
	var log = &GcpEntry{
		Timestamp: entry.Time.Format(time.RFC3339),
		Message:   entry.Message,
		Severity:  toGcpSeverityOrDefault(entry.Level),
		Data:      entry.Data,
	}

	serialized, err := json.Marshal(log)
	if err != nil {
		return nil, fmt.Errorf("Failed to marshal fields to JSON, %w", err)
	}
	return append(serialized, '\n'), nil
}
