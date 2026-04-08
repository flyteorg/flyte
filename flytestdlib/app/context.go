package app

import (
	"context"
	"net/http"

	"github.com/jmoiron/sqlx"
	"k8s.io/client-go/rest"
	"sigs.k8s.io/controller-runtime/pkg/cache"
	"sigs.k8s.io/controller-runtime/pkg/client"

	"github.com/flyteorg/flyte/v2/flytestdlib/promutils"
	"github.com/flyteorg/flyte/v2/flytestdlib/storage"
)

// WorkerFunc is a long-running function (e.g. controller-runtime manager)
// that runs alongside the HTTP server. It should block until ctx is cancelled.
type WorkerFunc func(ctx context.Context) error

// ReadyCheckFunc is called on /readyz requests. Return nil for healthy.
type ReadyCheckFunc func(r *http.Request) error

type namedWorker struct {
	name string
	fn   WorkerFunc
}

// SetupContext carries shared resources to service Setup functions.
// New fields can be added here without changing any Setup() signature.
type SetupContext struct {
	// Server address. Defaults to "0.0.0.0" and 8080.
	Host string
	Port int

	// Mux is the HTTP serve mux to register handlers on.
	Mux *http.ServeMux

	// DB is the shared database connection (may be nil if service doesn't need it).
	DB *sqlx.DB

	// K8sClient is the Kubernetes client with watch support (may be nil).
	K8sClient client.WithWatch

	// K8sConfig is the Kubernetes REST config (may be nil).
	K8sConfig *rest.Config

	// K8sCache is an optional shared controller-runtime cache.
	K8sCache cache.Cache

	// DataStore is the object storage client (may be nil).
	DataStore *storage.DataStore

	// Namespace is the Kubernetes namespace for CRs.
	Namespace string

	// BaseURL is the base URL for intra-service calls (e.g. "http://localhost:8090").
	// In unified mode the manager sets this so all services call each other
	// via the same mux. Standalone binaries leave it empty and use their own
	// per-service config for remote URLs.
	BaseURL string

	// Scope is the metrics scope.
	Scope promutils.Scope

	// Middleware wraps the final HTTP handler (e.g. CORS).
	Middleware func(http.Handler) http.Handler

	workers     []namedWorker
	readyChecks []ReadyCheckFunc
}

// AddWorker registers a background goroutine that runs alongside the HTTP
// server. The worker should block until ctx is cancelled.
func (sc *SetupContext) AddWorker(name string, fn WorkerFunc) {
	sc.workers = append(sc.workers, namedWorker{name: name, fn: fn})
}

// AddReadyCheck registers a readiness check invoked on /readyz.
func (sc *SetupContext) AddReadyCheck(fn ReadyCheckFunc) {
	sc.readyChecks = append(sc.readyChecks, fn)
}
