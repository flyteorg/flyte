// Package sentryutils reports Flyte operation counts and server-side failures
// to Sentry. The DSN is hardcoded and is not user configuration;
// DO_NOT_TRACK=1 opts out (FLYTE_DISABLE_TELEMETRY=true and the legacy
// FLYTE_DISABLE_SENTRY=true are honored too).
package sentryutils

import (
	"context"
	"sync"
	"time"

	"connectrpc.com/connect"
	"github.com/getsentry/sentry-go"
	"github.com/getsentry/sentry-go/attribute"

	"github.com/flyteorg/flyte/v2/flytestdlib/logger"
	"github.com/flyteorg/flyte/v2/flytestdlib/telemetry"
)

// dsn is hardcoded and is not user configuration
const dsn = "https://d0e3f0a470b8e1333411eff583cf4004@o4507249423810560.ingest.us.sentry.io/4511135180128256"

// operationMetric is the counter every instrumented operation increments.
const operationMetric = "flyte.operation"

var (
	initOnce sync.Once
	enabled  bool
)

// Disabled honors the shared telemetry opt-out (DO_NOT_TRACK).
func Disabled() bool {
	return telemetry.Disabled()
}

// Init initializes the process-wide Sentry client and reports whether
// reporting is on.
func Init(ctx context.Context, environment string) bool {
	initOnce.Do(func() {
		if Disabled() {
			return
		}
		if err := sentry.Init(sentry.ClientOptions{Dsn: dsn, Environment: environment}); err != nil {
			logger.Errorf(ctx, "failed to initialize sentry, continuing without it: %v", err)
			return
		}
		enabled = true
		logger.Infof(ctx, "Sentry error reporting enabled")
	})
	return enabled
}

// Flush drains buffered events and metrics. Sentry sends async, so a service
// shutting down has to wait or lose whatever is still queued.
func Flush() {
	sentry.Flush(2 * time.Second)
}

// Count records n successful occurrences of operation.
func Count(ctx context.Context, operation string, n int64) {
	sentry.NewMeter(ctx).Count(operationMetric, n, sentry.WithAttributes(
		attribute.String("operation", operation),
		attribute.String("status", "success"),
	))
}

// Gauge records the current value of the metric name. Unlike Count, each
// gauge is its own metric with no attributes.
func Gauge(ctx context.Context, name string, value float64) {
	sentry.NewMeter(ctx).Gauge(name, value)
}

// Interceptor counts one operation per call for each procedure in operations
// (keyed by procedure, valued by operation name) and reports server-side
// failures as exceptions. Client-caused errors (invalid argument, not found,
// ...) are intentionally not reported as exceptions.
func Interceptor(operations map[string]string) connect.UnaryInterceptorFunc {
	return func(next connect.UnaryFunc) connect.UnaryFunc {
		return func(ctx context.Context, req connect.AnyRequest) (connect.AnyResponse, error) {
			resp, err := next(ctx, req)
			operation, ok := operations[req.Spec().Procedure]
			if !ok {
				return resp, err
			}
			attrs := []attribute.Builder{attribute.String("operation", operation)}
			if err != nil {
				attrs = append(attrs,
					attribute.String("status", "error"),
					attribute.String("error_code", connect.CodeOf(err).String()),
				)
				switch connect.CodeOf(err) {
				case connect.CodeInternal, connect.CodeUnknown, connect.CodeDataLoss:
					hub := sentry.CurrentHub().Clone()
					hub.Scope().SetTag("procedure", req.Spec().Procedure)
					hub.CaptureException(err)
				}
			} else {
				attrs = append(attrs, attribute.String("status", "success"))
			}
			sentry.NewMeter(ctx).Count(operationMetric, 1, sentry.WithAttributes(attrs...))
			return resp, err
		}
	}
}
