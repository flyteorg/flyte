package gormimpl

import (
	"time"

	"github.com/flyteorg/flyte/flytestdlib/promutils"
)

// Common metrics emitted by gormimpl repos.
type gormMetrics struct {
	Scope                   promutils.Scope
	CreateDuration          promutils.StopWatch
	GetDuration             promutils.StopWatch
	UpdateDuration          promutils.StopWatch
	ListDuration            promutils.StopWatch
	ListIdentifiersDuration promutils.StopWatch
	DeleteDuration          promutils.StopWatch
	ExistsDuration          promutils.StopWatch
	CountDuration           promutils.StopWatch
	FindFirstCheckpoint     promutils.HistogramStopWatch
	FindNextCheckpoint      promutils.HistogramStopWatch
	FindStatusUpdates       promutils.HistogramStopWatch
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
		ListDuration: scope.MustNewStopWatch(
			"list", "time taken to list entries", time.Millisecond),
		ListIdentifiersDuration: scope.MustNewStopWatch(
			"list_identifiers", "time taken to list identifier entries", time.Millisecond),
		DeleteDuration: scope.MustNewStopWatch(
			"delete", "time taken to delete an individual entry", time.Millisecond),
		ExistsDuration: scope.MustNewStopWatch(
			"exists", "time taken to determine whether an individual entry exists", time.Millisecond),
		CountDuration: scope.MustNewStopWatch(
			"count", "time taken to count entries", time.Millisecond),
		FindFirstCheckpoint: scope.MustNewHistogramStopWatch(
			"find_first_checkpoint", "time taken to find first checkpoint for child execution status updates"),
		FindNextCheckpoint: scope.MustNewHistogramStopWatch(
			"find_next_checkpoint", "time taken to find next checkpoint for child execution status updates"),
		FindStatusUpdates: scope.MustNewHistogramStopWatch(
			"find_status_updates", "time taken to fetch child execution status updates"),
	}
}
