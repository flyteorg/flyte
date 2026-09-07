package main

import (
	"context"
	"fmt"
	"net/http"

	"github.com/flyteorg/flyte/v2/actions"
	actionsconfig "github.com/flyteorg/flyte/v2/actions/config"
	actionsk8s "github.com/flyteorg/flyte/v2/actions/k8s"
	flyteapp "github.com/flyteorg/flyte/v2/app"
	appconfig "github.com/flyteorg/flyte/v2/app/config"
	"github.com/flyteorg/flyte/v2/cache_service"
	cacheserviceconfig "github.com/flyteorg/flyte/v2/cache_service/config"
	"github.com/flyteorg/flyte/v2/dataproxy"
	dataproxyconfig "github.com/flyteorg/flyte/v2/dataproxy/config"
	"github.com/flyteorg/flyte/v2/events"
	eventsconfig "github.com/flyteorg/flyte/v2/events/config"
	"github.com/flyteorg/flyte/v2/executor"
	stdlibapp "github.com/flyteorg/flyte/v2/flytestdlib/app"
	"github.com/flyteorg/flyte/v2/flytestdlib/contextutils"
	"github.com/flyteorg/flyte/v2/flytestdlib/database"
	"github.com/flyteorg/flyte/v2/flytestdlib/promutils"
	"github.com/flyteorg/flyte/v2/flytestdlib/promutils/labeled"
	"github.com/flyteorg/flyte/v2/flytestdlib/storage"
	"github.com/flyteorg/flyte/v2/runs"
	runsconfig "github.com/flyteorg/flyte/v2/runs/config"
	"github.com/flyteorg/flyte/v2/secret"
	secretconfig "github.com/flyteorg/flyte/v2/secret/config"
	"k8s.io/client-go/kubernetes/scheme"
	servingv1 "knative.dev/serving/pkg/apis/serving/v1"
)

const componentAll = "all"

type componentSetup func(context.Context, *stdlibapp.SetupContext) error

var componentSetups = map[string]componentSetup{
	componentAll: setupAll,
	"runs":       setupRuns,
	"actions":    setupActions,
	"events":     setupEvents,
	"secret":     setupSecret,
	"cache":      setupCache,
	"app":        setupApp,
	"dataproxy":  setupDataproxy,
	"executor":   setupExecutor,
}

func setupComponent(ctx context.Context, sc *stdlibapp.SetupContext, component string) error {
	setup, ok := componentSetups[component]
	if !ok {
		return fmt.Errorf("unknown component %q", component)
	}
	return setup(ctx, sc)
}

func setupRuns(ctx context.Context, sc *stdlibapp.SetupContext) error {
	cfg := runsconfig.GetConfig()
	sc.Host = cfg.Server.Host
	sc.Port = cfg.Server.Port

	db, err := database.GetDB(ctx, database.GetConfig())
	if err != nil {
		return fmt.Errorf("failed to initialize database: %w", err)
	}
	sc.DB = db

	setMetricKeys()
	dataStore, err := storage.NewDataStore(storage.GetConfig(), promutils.NewTestScope())
	if err != nil {
		return fmt.Errorf("failed to initialize storage: %w", err)
	}
	sc.DataStore = dataStore

	return runs.Setup(ctx, sc)
}

func setupActions(ctx context.Context, sc *stdlibapp.SetupContext) error {
	cfg := actionsconfig.GetConfig()
	sc.Host = cfg.Server.Host
	sc.Port = cfg.Server.Port
	sc.Scope = promutils.NewScope("actions-service")

	if err := actionsk8s.InitScheme(); err != nil {
		return fmt.Errorf("failed to initialize Kubernetes scheme: %w", err)
	}
	k8sClient, k8sConfig, err := stdlibapp.InitKubernetesClient(ctx, cfg.Kubernetes, scheme.Scheme)
	if err != nil {
		return fmt.Errorf("failed to initialize Kubernetes client: %w", err)
	}
	k8sCache, err := stdlibapp.InitKubernetesCache(k8sConfig, scheme.Scheme)
	if err != nil {
		return err
	}
	sc.K8sClient = k8sClient
	sc.K8sConfig = k8sConfig
	sc.K8sCache = k8sCache
	sc.Namespace = cfg.Kubernetes.Namespace
	sc.AddWorker("kubernetes-cache", k8sCache.Start)

	return actions.Setup(ctx, sc)
}

func setupEvents(ctx context.Context, sc *stdlibapp.SetupContext) error {
	cfg := eventsconfig.GetConfig()
	sc.Host = cfg.Server.Host
	sc.Port = cfg.Server.Port
	return events.Setup(ctx, sc)
}

func setupSecret(ctx context.Context, sc *stdlibapp.SetupContext) error {
	cfg := secretconfig.GetConfig()
	sc.Host = cfg.Server.Host
	sc.Port = cfg.Server.Port

	k8sClient, _, err := stdlibapp.InitKubernetesClient(ctx, cfg.Kubernetes, nil)
	if err != nil {
		return fmt.Errorf("failed to initialize Kubernetes client: %w", err)
	}
	sc.K8sClient = k8sClient
	sc.Namespace = cfg.Kubernetes.Namespace

	return secret.Setup(ctx, sc)
}

func setupCache(ctx context.Context, sc *stdlibapp.SetupContext) error {
	cfg := cacheserviceconfig.GetConfig()
	sc.Host = cfg.Server.Host
	sc.Port = cfg.Server.Port

	db, err := database.GetDB(ctx, database.GetConfig())
	if err != nil {
		return fmt.Errorf("failed to initialize database: %w", err)
	}
	sc.DB = db

	setMetricKeys()
	dataStore, err := storage.NewDataStore(storage.GetConfig(), promutils.NewScope("cache-service"))
	if err != nil {
		return fmt.Errorf("failed to initialize storage: %w", err)
	}
	sc.DataStore = dataStore

	return cache_service.Setup(ctx, sc)
}

func setupApp(ctx context.Context, sc *stdlibapp.SetupContext) error {
	cfg := appconfig.GetAppConfig()
	k8sCfg := actionsconfig.GetConfig().Kubernetes
	sc.Host = cfg.Server.Host
	sc.Port = cfg.Server.Port
	sc.BaseURL = fmt.Sprintf("http://localhost:%d", cfg.Server.Port)

	if err := stdlibapp.InitAppScheme(); err != nil {
		return fmt.Errorf("failed to initialize Kubernetes scheme: %w", err)
	}
	k8sClient, k8sConfig, err := stdlibapp.InitKubernetesClient(ctx, k8sCfg, scheme.Scheme)
	if err != nil {
		return fmt.Errorf("failed to initialize Kubernetes client: %w", err)
	}
	k8sCache, err := stdlibapp.InitKubernetesCache(k8sConfig, scheme.Scheme)
	if err != nil {
		return err
	}
	sc.K8sClient = k8sClient
	sc.K8sConfig = k8sConfig
	sc.K8sCache = k8sCache
	sc.Namespace = k8sCfg.Namespace
	sc.AddWorker("kubernetes-cache", k8sCache.Start)

	if err := flyteapp.SetupInternal(ctx, sc); err != nil {
		return err
	}
	return flyteapp.Setup(ctx, sc)
}

func setupDataproxy(ctx context.Context, sc *stdlibapp.SetupContext) error {
	cfg := dataproxyconfig.GetConfig()
	sc.Host = cfg.Server.Host
	sc.Port = cfg.Server.Port

	_, k8sConfig, err := stdlibapp.InitKubernetesClient(ctx, actionsconfig.GetConfig().Kubernetes, nil)
	if err != nil {
		return fmt.Errorf("failed to initialize Kubernetes client: %w", err)
	}
	sc.K8sConfig = k8sConfig

	setMetricKeys()
	dataStore, err := storage.NewDataStore(storage.GetConfig(), promutils.NewScope("dataproxy-service"))
	if err != nil {
		return fmt.Errorf("failed to initialize storage: %w", err)
	}
	sc.DataStore = dataStore

	return dataproxy.Setup(ctx, sc)
}

func setupExecutor(ctx context.Context, sc *stdlibapp.SetupContext) error {
	sc.Port = 0
	setMetricKeys()
	k8sCfg := actionsconfig.GetConfig().Kubernetes
	_, k8sConfig, err := stdlibapp.InitKubernetesClient(ctx, k8sCfg, executor.Scheme())
	if err != nil {
		return fmt.Errorf("failed to initialize Kubernetes client: %w", err)
	}
	sc.K8sConfig = k8sConfig
	sc.Namespace = k8sCfg.Namespace

	if err := executor.Setup(ctx, sc); err != nil {
		return fmt.Errorf("executor setup failed: %w", err)
	}
	return nil
}

func setupAll(ctx context.Context, sc *stdlibapp.SetupContext) error {
	serverCfg := runsconfig.GetConfig().Server
	k8sCfg := actionsconfig.GetConfig().Kubernetes
	sc.Host = serverCfg.Host
	sc.Port = serverCfg.Port
	sc.Namespace = k8sCfg.Namespace
	sc.Middleware = corsMiddleware
	sc.BaseURL = fmt.Sprintf("http://localhost:%d", serverCfg.Port)

	dbCfg := &runsconfig.GetConfig().Database
	db, err := database.GetDB(ctx, dbCfg)
	if err != nil {
		return fmt.Errorf("failed to initialize database: %w", err)
	}
	sc.DB = db

	if err := servingv1.AddToScheme(executor.Scheme()); err != nil {
		return fmt.Errorf("failed to register Knative scheme: %w", err)
	}
	k8sClient, k8sConfig, err := stdlibapp.InitKubernetesClient(ctx, k8sCfg, executor.Scheme())
	if err != nil {
		return fmt.Errorf("failed to initialize Kubernetes client: %w", err)
	}
	sc.K8sClient = k8sClient
	sc.K8sConfig = k8sConfig
	sc.Scope = promutils.NewScope("flyte")

	setMetricKeys()
	dataStore, err := storage.NewDataStore(storage.GetConfig(), sc.Scope.NewSubScope("storage"))
	if err != nil {
		return fmt.Errorf("failed to initialize storage: %w", err)
	}
	sc.DataStore = dataStore

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
	if err := executor.Setup(ctx, sc); err != nil {
		return err
	}
	if err := actions.Setup(ctx, sc); err != nil {
		return err
	}
	if err := flyteapp.SetupInternal(ctx, sc); err != nil {
		return err
	}
	if err := flyteapp.Setup(ctx, sc); err != nil {
		return err
	}
	return secret.Setup(ctx, sc)
}

func setMetricKeys() {
	labeled.SetMetricKeys(contextutils.ProjectKey, contextutils.DomainKey, contextutils.WorkflowIDKey, contextutils.TaskIDKey)
}

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
