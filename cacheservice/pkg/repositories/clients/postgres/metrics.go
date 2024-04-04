package postgres

import (
	"time"

	"github.com/flyteorg/flyte/flytestdlib/promutils"
	"github.com/flyteorg/flyte/flytestdlib/promutils/labeled"
)

type gormMetrics struct {
	Scope                     promutils.Scope
	PutDuration               labeled.StopWatch
	GetDuration               labeled.StopWatch
	DeleteDuration            labeled.StopWatch
	CreateReservationDuration labeled.StopWatch
	UpdateReservationDuration labeled.StopWatch
	DeleteReservationDuration labeled.StopWatch
	GetReservationDuration    labeled.StopWatch
}

func newPostgresRepoMetrics(scope promutils.Scope) gormMetrics {
	return gormMetrics{
		Scope: scope,
		PutDuration: labeled.NewStopWatch(
			"put", "time taken to create/update a new entry", time.Millisecond, scope),
		GetDuration: labeled.NewStopWatch(
			"get", "time taken to get an entry", time.Millisecond, scope),
		DeleteDuration: labeled.NewStopWatch(
			"delete", "time taken to delete an individual entry", time.Millisecond, scope),
		CreateReservationDuration: labeled.NewStopWatch(
			"createReservation", "time taken to create a new entry", time.Millisecond, scope),
		UpdateReservationDuration: labeled.NewStopWatch(
			"updateReservation", "time taken to update an entry", time.Millisecond, scope),
		DeleteReservationDuration: labeled.NewStopWatch(
			"deleteReservation", "time taken to delete an individual entry", time.Millisecond, scope),
		GetReservationDuration: labeled.NewStopWatch(
			"getReservation", "time taken to get an entry", time.Millisecond, scope),
	}
}
