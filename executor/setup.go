package executor

import (
	"context"
	"crypto/tls"
	"fmt"
	"net/http"
	"os"

	"k8s.io/apimachinery/pkg/runtime"
	utilruntime "k8s.io/apimachinery/pkg/util/runtime"
	"k8s.io/client-go/kubernetes"
	clientgoscheme "k8s.io/client-go/kubernetes/scheme"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/healthz"
	"sigs.k8s.io/controller-runtime/pkg/log/zap"
	"sigs.k8s.io/controller-runtime/pkg/metrics/filters"
	metricsserver "sigs.k8s.io/controller-runtime/pkg/metrics/server"
	"sigs.k8s.io/controller-runtime/pkg/webhook"

	"github.com/flyteorg/flyte/v2/flytestdlib/app"
	flyteorgv1 "github.com/flyteorg/flyte/v2/executor/api/v1"
	"github.com/flyteorg/flyte/v2/executor/pkg/config"
	"github.com/flyteorg/flyte/v2/executor/pkg/controller"
	"github.com/flyteorg/flyte/v2/executor/pkg/plugin"
	webhookPkg "github.com/flyteorg/flyte/v2/executor/pkg/webhook"
	"github.com/flyteorg/flyte/v2/flyteplugins/go/tasks/pluginmachinery"
	"github.com/flyteorg/flyte/v2/flyteplugins/go/tasks/pluginmachinery/catalog"
	cachecatalog "github.com/flyteorg/flyte/v2/flyteplugins/go/tasks/pluginmachinery/catalog/cache_service"
	webhookConfig "github.com/flyteorg/flyte/v2/flyteplugins/go/tasks/pluginmachinery/secret/config"
	"github.com/flyteorg/flyte/v2/flytestdlib/promutils"
	"github.com/flyteorg/flyte/v2/flytestdlib/storage"
	"github.com/flyteorg/flyte/v2/gen/go/flyteidl2/workflow/workflowconnect"

	// Plugin registrations — blank imports trigger init() which registers
	// plugins with the global registry.
	_ "github.com/flyteorg/flyte/v2/flyteplugins/go/tasks/plugins/k8s/pod"
)

var scheme = runtime.NewScheme()

func init() {
	utilruntime.Must(clientgoscheme.AddToScheme(scheme))
	utilruntime.Must(flyteorgv1.AddToScheme(scheme))
}

// Scheme returns the runtime.Scheme with executor CRDs registered.
// Useful for callers that need to pass the scheme to InitKubernetesClient.
func Scheme() *runtime.Scheme {
	return scheme
}

// Setup registers the executor as a background worker on the SetupContext.
// Requires sc.K8sConfig and sc.DataStore to be set.
func Setup(ctx context.Context, sc *app.SetupContext) error {
	ctrl.SetLogger(zap.New(zap.UseDevMode(true)))

	cfg := config.GetConfig()

	var tlsOpts []func(*tls.Config)
	if !cfg.EnableHTTP2 {
		tlsOpts = append(tlsOpts, func(c *tls.Config) {
			c.NextProtos = []string{"http/1.1"}
		})
	}

	wCfg := webhookConfig.GetConfig()
	webhookServerOptions := webhook.Options{TLSOpts: tlsOpts}
	webhookServerOptions.CertDir = wCfg.ExpandCertDir()
	webhookServerOptions.CertName = webhookPkg.ServerCertKey
	webhookServerOptions.KeyName = webhookPkg.ServerCertPrivateKey
	webhookServerOptions.Port = wCfg.ListenPort

	metricsServerOptions := metricsserver.Options{
		BindAddress:   cfg.MetricsBindAddress,
		SecureServing: cfg.MetricsSecure,
		TLSOpts:       tlsOpts,
	}
	if cfg.MetricsSecure {
		metricsServerOptions.FilterProvider = filters.WithAuthenticationAndAuthorization
	}
	if len(cfg.MetricsCertPath) > 0 {
		metricsServerOptions.CertDir = cfg.MetricsCertPath
		metricsServerOptions.CertName = cfg.MetricsCertName
		metricsServerOptions.KeyName = cfg.MetricsCertKey
	}

	mgr, err := ctrl.NewManager(sc.K8sConfig, ctrl.Options{
		Scheme:                 scheme,
		Metrics:                metricsServerOptions,
		WebhookServer:          webhook.NewServer(webhookServerOptions),
		HealthProbeBindAddress: cfg.HealthProbeBindAddress,
		LeaderElection:         cfg.LeaderElect,
		LeaderElectionID:       "abf369a8.flyte.org",
	})
	if err != nil {
		return fmt.Errorf("executor: failed to create controller manager: %w", err)
	}
	sc.K8sCache = mgr.GetCache()

	kubeClient, err := kubernetes.NewForConfig(sc.K8sConfig)
	if err != nil {
		return fmt.Errorf("executor: failed to create kubernetes client for webhook: %w", err)
	}

	podNamespace := os.Getenv(webhookPkg.PodNamespaceEnvVar)
	if podNamespace == "" {
		podNamespace = sc.Namespace
	}

	if err := webhookPkg.Setup(ctx, kubeClient, wCfg, podNamespace, promutils.NewScope("executor"), mgr); err != nil {
		return fmt.Errorf("executor: webhook setup failed: %w", err)
	}

	dataStore, err := storage.NewDataStore(storage.GetConfig(), promutils.NewScope("executor:storage"))
	if err != nil {
		return fmt.Errorf("executor: failed to create data store: %w", err)
	}

	setupCtx := plugin.NewSetupContext(
		mgr, nil, nil, nil, nil,
		"TaskAction",
		promutils.NewScope("executor"),
	)
	registry := plugin.NewRegistry(setupCtx, pluginmachinery.PluginRegistry())
	if err := registry.Initialize(ctx); err != nil {
		return fmt.Errorf("executor: failed to initialize plugin registry: %w", err)
	}

	eventsServiceURL := sc.BaseURL
	if eventsServiceURL == "" {
		eventsServiceURL = cfg.EventsServiceURL
	}
	eventsClient := workflowconnect.NewEventsProxyServiceClient(http.DefaultClient, eventsServiceURL)
	catalogCfg := catalog.GetConfig()
	cacheServiceURL := sc.BaseURL
	if cacheServiceURL == "" {
		cacheServiceURL = cfg.CacheServiceURL
	}
	cacheClient := cachecatalog.NewHTTPClient(dataStore, cacheServiceURL, catalogCfg.MaxCacheAge.Duration)
	asyncCatalogClient, err := catalog.NewAsyncClient(cacheClient, *catalogCfg, promutils.NewScope("executor:catalog"))
	if err != nil {
		return fmt.Errorf("executor: failed to create catalog cache client: %w", err)
	}
	if err := asyncCatalogClient.Start(ctx); err != nil {
		return fmt.Errorf("executor: failed to start catalog cache client: %w", err)
	}

	reconciler := controller.NewTaskActionReconciler(
		mgr.GetClient(), mgr.GetScheme(), registry, dataStore, eventsClient, cfg.Cluster,
	)
	reconciler.CatalogClient = asyncCatalogClient
	reconciler.Catalog = cacheClient
	reconciler.Recorder = mgr.GetEventRecorderFor("taskaction-controller")
	reconciler.MaxSystemFailures = cfg.MaxSystemFailures
	if err := reconciler.SetupWithManager(mgr); err != nil {
		return fmt.Errorf("executor: failed to setup controller: %w", err)
	}

	if cfg.GC.Interval.Duration > 0 {
		if cfg.GC.MaxTTL.Duration <= 0 {
			return fmt.Errorf("executor: gc.maxTTL must be positive when gc is enabled, got %v", cfg.GC.MaxTTL.Duration)
		}
		gc := controller.NewGarbageCollector(mgr.GetClient(), cfg.GC.Interval.Duration, cfg.GC.MaxTTL.Duration)
		if err := mgr.Add(gc); err != nil {
			return fmt.Errorf("executor: failed to add garbage collector: %w", err)
		}
	}

	if err := mgr.AddHealthzCheck("healthz", healthz.Ping); err != nil {
		return fmt.Errorf("executor: failed to add health check: %w", err)
	}
	if err := mgr.AddReadyzCheck("readyz", healthz.Ping); err != nil {
		return fmt.Errorf("executor: failed to add ready check: %w", err)
	}

	sc.AddWorker("executor", func(ctx context.Context) error {
		return mgr.Start(ctx)
	})

	return nil
}
