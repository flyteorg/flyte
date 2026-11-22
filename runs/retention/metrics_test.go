package retention

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
	"go.opentelemetry.io/otel/attribute"
	sdkmetric "go.opentelemetry.io/otel/sdk/metric"
	"go.opentelemetry.io/otel/sdk/metric/metricdata"
)

func testMetrics(t *testing.T) (*Metrics, *sdkmetric.ManualReader) {
	t.Helper()
	reader := sdkmetric.NewManualReader()
	provider := sdkmetric.NewMeterProvider(sdkmetric.WithReader(reader))
	t.Cleanup(func() { require.NoError(t, provider.Shutdown(context.Background())) })
	m, err := NewMetrics(provider)
	require.NoError(t, err)
	return m, reader
}

func deletedTotals(t *testing.T, reader *sdkmetric.ManualReader) map[string]int64 {
	t.Helper()
	var data metricdata.ResourceMetrics
	require.NoError(t, reader.Collect(context.Background(), &data))
	totals := map[string]int64{}
	for _, scope := range data.ScopeMetrics {
		for _, m := range scope.Metrics {
			if m.Name != "runs.retention.deleted" {
				continue
			}
			for _, p := range m.Data.(metricdata.Sum[int64]).DataPoints {
				entity, _ := p.Attributes.Value(attribute.Key("entity"))
				totals[entity.AsString()] += p.Value
			}
		}
	}
	return totals
}

func TestInvocationMetrics(t *testing.T) {
	for _, tc := range []struct {
		name    string
		result  Result
		err     error
		outcome string
	}{
		{"Success", Result{}, nil, "success"},
		{"Failure", Result{}, errors.New("database failure"), "failure"},
		{"DryRun", Result{DryRun: true}, nil, "success"},
		{"Skipped", Result{LockUnavailable: true}, nil, "skipped"},
		{"Cancelled", Result{}, context.Canceled, "failure"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			m, reader := testMetrics(t)
			ctx, cancel := context.WithCancel(context.Background())
			cancel()
			m.RecordInvocation(ctx, 2*time.Second, tc.result, tc.err)
			var data metricdata.ResourceMetrics
			require.NoError(t, reader.Collect(context.Background(), &data))
			seen := map[string]bool{}
			for _, scope := range data.ScopeMetrics {
				for _, metric := range scope.Metrics {
					seen[metric.Name] = true
					switch points := metric.Data.(type) {
					case metricdata.Histogram[float64]:
						require.Equal(t, "runs.retention.duration", metric.Name)
						require.Equal(t, "s", metric.Unit)
						require.Len(t, points.DataPoints, 1)
						p := points.DataPoints[0]
						require.EqualValues(t, 1, p.Count)
						require.Equal(t, 2.0, p.Sum)
						value, _ := p.Attributes.Value(attribute.Key("result"))
						require.Equal(t, tc.outcome, value.AsString())
						dry, _ := p.Attributes.Value(attribute.Key("dry_run"))
						require.Equal(t, tc.result.DryRun, dry.AsBool())
					default:
						t.Fatalf("unexpected metric %s", metric.Name)
					}
				}
			}
			require.False(t, seen["runs.retention.invocations"])
			require.True(t, seen["runs.retention.duration"])
		})
	}
}
