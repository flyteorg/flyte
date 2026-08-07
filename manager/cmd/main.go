package main

import (
	"context"
	"fmt"
	"net/http"
	"os"

	"github.com/flyteorg/flyte/v2/actions"
	actionsconfig "github.com/flyteorg/flyte/v2/actions/config"
	flyteapp "github.com/flyteorg/flyte/v2/app"
	"github.com/flyteorg/flyte/v2/cache_service"
	"github.com/flyteorg/flyte/v2/dataproxy"
	"github.com/flyteorg/flyte/v2/events"
	"github.com/flyteorg/flyte/v2/executor"
	"github.com/flyteorg/flyte/v2/flytestdlib/app"
	"github.com/flyteorg/flyte/v2/flytestdlib/contextutils"
	"github.com/flyteorg/flyte/v2/flytestdlib/database"
	"github.com/flyteorg/flyte/v2/flytestdlib/promutils"
	"github.com/flyteorg/flyte/v2/flytestdlib/promutils/labeled"
	"github.com/flyteorg/flyte/v2/flytestdlib/storage"
	"github.com/flyteorg/flyte/v2/runs"
	runsconfig "github.com/flyteorg/flyte/v2/runs/config"
	"github.com/flyteorg/flyte/v2/secret"
	servingv1 "knative.dev/serving/pkg/apis/serving/v1"
)

func main() {
	a := &app.App{
		Name:  "flyte",
		Short: "Unified Flyte Service - Runs all services and operator",
		Setup: setup,
	}
	if err := a.Run(); err != nil {
		os.Exit(1)
	}
}

func setup(ctx context.Context, sc *app.SetupContext) error {
	// Reuse the run service's server config and the actions service's Kubernetes
	// config rather than maintaining a duplicate manager-specific section.
	serverCfg := runsconfig.GetConfig().Server
	k8sCfg := actionsconfig.GetConfig().Kubernetes
	sc.Host = serverCfg.Host
	sc.Port = serverCfg.Port
	sc.Namespace = k8sCfg.Namespace
	sc.Middleware = corsMiddleware
	sc.BaseURL = fmt.Sprintf("http://localhost:%d", serverCfg.Port)

	// Initialize database
	dbCfg := &runsconfig.GetConfig().Database
	db, err := database.GetDB(ctx, dbCfg)
	if err != nil {
		return fmt.Errorf("failed to initialize database: %w", err)
	}
	sc.DB = db

	// Register Knative types into the shared scheme so the K8s client can manage KServices.
	if err := servingv1.AddToScheme(executor.Scheme()); err != nil {
		return fmt.Errorf("failed to register Knative scheme: %w", err)
	}

	// Initialize Kubernetes client
	k8sClient, k8sConfig, err := app.InitKubernetesClient(ctx, k8sCfg, executor.Scheme())
	if err != nil {
		return fmt.Errorf("failed to initialize Kubernetes client: %w", err)
	}
	sc.K8sClient = k8sClient
	sc.K8sConfig = k8sConfig

	// Initialize metrics scope
	sc.Scope = promutils.NewScope("flyte")

	// Initialize labeled metrics (required for storage)
	labeled.SetMetricKeys(contextutils.ProjectKey, contextutils.DomainKey, contextutils.WorkflowIDKey, contextutils.TaskIDKey)

	// Initialize storage
	storageCfg := storage.GetConfig()
	dataStore, err := storage.NewDataStore(storageCfg, sc.Scope.NewSubScope("storage"))
	if err != nil {
		return fmt.Errorf("failed to initialize storage: %w", err)
	}
	sc.DataStore = dataStore

	// Setup all services
	if err := runs.Setup(ctx, sc); err != nil {
		return err
	}
	if err := dataproxy.Setup(ctx, sc); err != nil {
		return err
	}
	if err := events.Setup(ctx, sc); err != nil {
		return err
	}
	if err := cache_service.Setup(ctx, sc); err != nil {
		return err
	}
	// executor.Setup sets sc.K8sCache (via mgr.GetCache()); services that depend
	// on the cache (InternalAppService, Actions) must be set up after this.
	if err := executor.Setup(ctx, sc); err != nil {
		return err
	}
	if err := actions.Setup(ctx, sc); err != nil {
		return err
	}
	// InternalAppService must be mounted before AppService so the proxy can reach it.
	if err := flyteapp.SetupInternal(ctx, sc); err != nil {
		return err
	}
	if err := flyteapp.Setup(ctx, sc); err != nil {
		return err
	}
	if err := secret.Setup(ctx, sc); err != nil {
		return err
	}

	return nil
}

// corsMiddleware wraps an http.Handler with permissive CORS headers for local
// development (UI on localhost:8080 -> manager on localhost:8090).
func corsMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		origin := r.Header.Get("Origin")
		if origin == "" {
			origin = "*"
		}
		w.Header().Set("Access-Control-Allow-Origin", origin)
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers",
			"Accept, Authorization, Content-Type, "+
				"Connect-Protocol-Version, Connect-Timeout-Ms, "+
				"Grpc-Timeout, X-Grpc-Web, X-User-Agent")
		w.Header().Set("Access-Control-Expose-Headers",
			"Grpc-Status, Grpc-Message, Grpc-Status-Details-Bin")
		w.Header().Set("Access-Control-Allow-Credentials", "true")

		if r.Method == http.MethodOptions {
			w.Header().Set("Access-Control-Max-Age", "86400")
			w.WriteHeader(http.StatusNoContent)
			return
		}

		next.ServeHTTP(w, r)
	})
}
