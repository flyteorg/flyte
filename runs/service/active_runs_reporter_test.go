package service

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestActiveRunsReporter_ReportsRunningCount(t *testing.T) {
	reporter := NewActiveRunsReporter(func(context.Context) (int64, error) {
		return 3, nil
	}, time.Minute)

	var gotName string
	var gotValue float64
	reporter.report = func(_ context.Context, name string, value float64) {
		gotName, gotValue = name, value
	}

	reporter.reportOnce(context.Background())

	assert.Equal(t, "flyte.active_runs", gotName)
	assert.Equal(t, float64(3), gotValue)
}

func TestActiveRunsReporter_SkipsOnError(t *testing.T) {
	reporter := NewActiveRunsReporter(func(context.Context) (int64, error) {
		return 0, errors.New("db down")
	}, time.Minute)

	called := false
	reporter.report = func(context.Context, string, float64) { called = true }

	reporter.reportOnce(context.Background())

	assert.False(t, called)
}
