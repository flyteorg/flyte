package logger

import (
	"testing"
	"time"

	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
)

func Test_gcpFormatter(t *testing.T) {
	tests := []struct {
		name     string
		entry    *logrus.Entry
		expected string
	}{
		{"test", &logrus.Entry{
			Logger:  nil,
			Data:    map[string]interface{}{"src": "some-src"},
			Time:    time.Date(2000, 01, 01, 01, 01, 000, 000, time.UTC),
			Level:   1,
			Caller:  nil,
			Message: "some-message",
			Buffer:  nil,
			Context: nil,
		}, "{\"data\":{\"src\":\"some-src\"},\"message\":\"some-message\",\"severity\":\"CRITICAL\",\"timestamp\":\"2000-01-01T01:01:00Z\"}\n"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			formatter := GcpFormatter{}
			bytelog, err := formatter.Format(tt.entry)
			assert.Nil(t, err)
			assert.Equal(t, tt.expected, string(bytelog))
		})
	}
}

func Test_gcpFormatterFails(t *testing.T) {
	tests := []struct {
		name   string
		entry  *logrus.Entry
		errMsg string
	}{
		{"test", &logrus.Entry{
			Logger:  nil,
			Data:    map[string]interface{}{"src": make(chan int)},
			Time:    time.Date(2000, 01, 01, 01, 01, 000, 000, time.UTC),
			Level:   1,
			Caller:  nil,
			Message: "some-message",
			Buffer:  nil,
			Context: nil,
		}, "Failed to marshal fields to JSON, json: unsupported type: chan int"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			formatter := GcpFormatter{}
			_, err := formatter.Format(tt.entry)
			assert.Equal(t, tt.errMsg, err.Error())
		})
	}
}
