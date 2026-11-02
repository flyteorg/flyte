package redis

import (
	"sync"

	redisotelnative "github.com/redis/go-redis/extra/redisotel-native/v9"
	"go.opentelemetry.io/otel/metric"
)

var metricsMu sync.Mutex

// InitMetrics initializes native Redis metrics for all clients in this process.
// Call before creating clients. Subsequent calls preserve the first provider;
// a nil provider leaves metrics uninitialized.
func InitMetrics(provider metric.MeterProvider) error {
	if provider == nil {
		return nil
	}
	metricsMu.Lock()
	defer metricsMu.Unlock()
	observability := redisotelnative.GetObservabilityInstance()
	if observability.IsEnabled() {
		return nil
	}
	return observability.Init(redisotelnative.NewConfig().WithEnabled(true).WithMeterProvider(provider))
}
