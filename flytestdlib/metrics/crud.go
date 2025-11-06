package metrics

import (
	"context"
	"time"

	"github.com/prometheus/client_golang/prometheus"

	"github.com/flyteorg/flyte/v2/flytestdlib/promutils"
	"github.com/flyteorg/flyte/v2/flytestdlib/promutils/labeled"
)

type APITimer struct {
	Success labeled.Timer
	Failure labeled.Timer
}

func (a APITimer) Stop(err *error) {
	if err != nil && *err != nil {
		a.Failure.Stop()
	} else {
		a.Success.Stop()
	}
}

type API struct {
	Success labeled.StopWatch
	Failure labeled.StopWatch
}

func (a API) Start(ctx context.Context) APITimer {
	return APITimer{
		Success: a.Success.Start(ctx),
		Failure: a.Failure.Start(ctx),
	}
}

type OperationTimer struct {
	Start   time.Time
	Success prometheus.Observer
	Failure prometheus.Observer
}

func (o OperationTimer) Stop(err *error) {
	seconds := time.Since(o.Start).Seconds()
	if err != nil && *err != nil {
		o.Failure.Observe(seconds)
	} else {
		o.Success.Observe(seconds)
	}
}

type Operation struct {
	Success *prometheus.HistogramVec
	Failure *prometheus.HistogramVec
}

func (o Operation) Start(values ...string) OperationTimer {
	return OperationTimer{
		Start:   time.Now(),
		Success: o.Success.WithLabelValues(values...),
		Failure: o.Failure.WithLabelValues(values...),
	}
}

type CRUD struct {
	Create   API
	Read     API
	Update   API
	Delete   API
	Undelete API
	List     API
	Count    API
}

func NewAPI(scope promutils.Scope) API {
	return API{
		Success: labeled.NewStopWatch("success", "api succeeded", time.Millisecond, scope),
		Failure: labeled.NewStopWatch("failure", "api failed", time.Millisecond, scope),
	}
}

// NewOperationHistogram creates 2 histogram metrics for tracking success & failure durations.
//
// The resulting operations requires len(labels) + 1 labels, where the last label is the respective "operation".
//
// Histograms have higher performance impact on Prometheus server when calculating quantiles:
// https://prometheus.io/docs/practices/histograms/#quantiles
func NewOperationHistogram(scope promutils.Scope, labels ...string) Operation {
	labelSlice := append(labels, "operation")
	return Operation{
		Success: scope.MustNewHistogramVec("success", "operation succeeded", labelSlice...),
		Failure: scope.MustNewHistogramVec("failure", "operation failed", labelSlice...),
	}
}

func NewCRUD(scope promutils.Scope) CRUD {
	return CRUD{
		Create:   NewAPI(scope.NewSubScope("create")),
		Read:     NewAPI(scope.NewSubScope("read")),
		Update:   NewAPI(scope.NewSubScope("update")),
		List:     NewAPI(scope.NewSubScope("list")),
		Delete:   NewAPI(scope.NewSubScope("delete")),
		Undelete: NewAPI(scope.NewSubScope("undelete")),
		Count:    NewAPI(scope.NewSubScope("count")),
	}
}
