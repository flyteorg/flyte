package db

import (
	"time"

	"github.com/flyteorg/flyte/flytestdlib/promutils"
)

// Common metrics emitted by gormimpl repos.
type gormMetrics struct {
	Scope          promutils.Scope
	CreateDuration promutils.StopWatch
	GetDuration    promutils.StopWatch
	UpdateDuration promutils.StopWatch
	SearchDuration promutils.StopWatch
}

func newMetrics(scope promutils.Scope) gormMetrics {
	return gormMetrics{
		Scope: scope,
		CreateDuration: scope.MustNewStopWatch(
			"create", "time taken to create a new entry", time.Millisecond),
		GetDuration: scope.MustNewStopWatch(
			"get", "time taken to get an entry", time.Millisecond),
		UpdateDuration: scope.MustNewStopWatch(
			"update", "time taken to update an entry", time.Millisecond),
		SearchDuration: scope.MustNewStopWatch(
			"search", "time taken for searching", time.Millisecond),
	}
}
