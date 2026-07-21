package executor

import (
	"context"
	"crypto/tls"
	"fmt"
	"net/http"
	"os"

	"connectrpc.com/connect"
	"connectrpc.com/otelconnect"
	"k8s.io/apimachinery/pkg/runtime"
	utilruntime "k8s.io/apimachinery/pkg/util/runtime"
	"k8s.io/client-go/informers"
	"k8s.io/client-go/kubernetes"
	clientgoscheme "k8s.io/client-go/kubernetes/scheme"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/healthz"
	"sigs.k8s.io/controller-runtime/pkg/log/zap"
	"sigs.k8s.io/controller-runtime/pkg/metrics/filters"
	metricsserver "sigs.k8s.io/controller-runtime/pkg/metrics/server"
	"sigs.k8s.io/controller-runtime/pkg/webhook"

	flyteorgv1 "github.com/flyteorg/flyte/v2/executor/api/v1"
	"github.com/flyteorg/flyte/v2/executor/pkg/config"
	"github.com/flyteorg/flyte/v2/executor/pkg/controller"
	"github.com/flyteorg/flyte/v2/executor/pkg/plugin"
	webhookPkg "github.com/flyteorg/flyte/v2/executor/pkg/webhook"
	"github.com/flyteorg/flyte/v2/flyteplugins/go/tasks/pluginmachinery"
	"github.com/flyteorg/flyte/v2/flyteplugins/go/tasks/pluginmachinery/catalog"
	"github.com/flyteorg/flyte/v2/flyteplugins/go/tasks/pluginmachinery/flytek8s"
	cachecatalog "github.com/flyteorg/flyte/v2/flyteplugins/go/tasks/pluginmachinery/catalog/cache_service"
	webhookConfig "github.com/flyteorg/flyte/v2/flyteplugins/go/tasks/pluginmachinery/secret/config"
	connectorplugin "github.com/flyteorg/flyte/v2/flyteplugins/go/tasks/plugins/webapi/connector"
	"github.com/flyteorg/flyte/v2/flytestdlib/app"
	"github.com/flyteorg/flyte/v2/flytestdlib/otelutils"
	"github.com/flyteorg/flyte/v2/flytestdlib/promutils"
	"github.com/flyteorg/flyte/v2/flytestdlib/storage"
	"github.com/flyteorg/flyte/v2/gen/go/flyteidl2/workflow/workflowconnect"

	_ "github.com/flyteorg/flyte/v2/executor/plugins"
	_ "github.com/flyteorg/flyte/v2/flyteplugins/go/tasks/plugins/k8s/clustered"
	_ "github.com/flyteorg/flyte/v2/flyteplugins/go/tasks/plugins/k8s/pod"
)

var scheme = runtime.NewScheme()

const otelServiceName = "executor"

func init() {
	utilruntime.Must(clientgoscheme.AddToScheme(scheme))
	utilruntime.Must(flyteorgv1.AddToScheme(scheme))
}

// Scheme returns the runtime.Scheme with executor CRDs registered.
// Useful for callers that need to pass the scheme to InitKubernetesClient.
func Scheme() *runtime.Scheme {
	return scheme
}

// watchPodTemplates wires a PodTemplate informer into the flytek8s.DefaultPodTemplateStore,
// with defaultNamespace as the fallback namespace for template lookups.
func watchPodTemplates(informerFactory informers.SharedInformerFactory, defaultNamespace string) error {
	flytek8s.DefaultPodTemplateStore.SetDefaultNamespace(defaultNamespace)
	_, err := informerFactory.Core().V1().PodTemplates().Informer().AddEventHandler(
		flytek8s.GetPodTemplateUpdatesHandler(&flytek8s.DefaultPodTemplateStore))
	return err
}

// Setup registers the executor as a background worker on the SetupContext.
// Requires sc.K8sConfig and sc.DataStore to be set.
func Setup(ctx context.Context, sc *app.SetupContext) error {
	ctrl.SetLogger(zap.New(zap.UseDevMode(true)))

	cfg := config.GetConfig()

	for _, reg := range pluginmachinery.PluginRegistry().GetSchemeRegisters() {
		utilruntime.Must(reg.AddToScheme(scheme))
	}

	// Register the connector (webapi) backend plugin so task types backed by an external connector
	// service are routed to it. This must run before plugin.NewRegistry below, which snapshots the
	// core plugins once.
	connectorplugin.RegisterConnectorPlugin(&connectorplugin.ConnectorService{})

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

	// Populate the flytek8s.DefaultPodTemplateStore read during pod construction, so both the
	// global plugins.k8s default-pod-template-name and per-task pod_template_name resolve.
	// Uses a standalone informer rather than the manager cache so a missing `podtemplates`
	// RBAC grant degrades to watch-error logs instead of failing manager cache sync.
	informerFactory := informers.NewSharedInformerFactory(kubeClient, 0)
	if err := watchPodTemplates(informerFactory, podNamespace); err != nil {
		return fmt.Errorf("executor: failed to register PodTemplate event handler: %w", err)
	}
	sc.AddWorker("podtemplate-informer", func(ctx context.Context) error {
		informerFactory.Start(ctx.Done())
		<-ctx.Done()
		return nil
	})

	executorScope := promutils.NewScope("executor")

	if err := webhookPkg.Setup(ctx, kubeClient, wCfg, podNamespace, executorScope.NewSubScope("webhook"), mgr); err != nil {
		return fmt.Errorf("executor: webhook setup failed: %w", err)
	}

	dataStore, err := storage.NewDataStore(storage.GetConfig(), promutils.NewScope("executor:storage"))
	if err != nil {
		return fmt.Errorf("executor: failed to create data store: %w", err)
	}

	setupCtx := plugin.NewSetupContext(
		mgr, plugin.NewNoopSecretManager(), plugin.NewNoopResourceRegistrar(), nil, nil,
		"TaskAction",
		executorScope.NewSubScope("plugin"),
	)
	registry := plugin.NewRegistry(setupCtx, pluginmachinery.PluginRegistry())
	if err := registry.Initialize(ctx); err != nil {
		return fmt.Errorf("executor: failed to initialize plugin registry: %w", err)
	}

	otelCfg := otelutils.GetConfig()
	if err := otelutils.RegisterProvidersWithContext(ctx, otelServiceName, otelCfg); err != nil {
		return fmt.Errorf("registering otel providers: %w", err)
	}
	otelInterceptor, err := otelconnect.NewInterceptor(
		otelconnect.WithTracerProvider(otelutils.GetTracerProvider(otelServiceName)),
		otelconnect.WithMeterProvider(otelutils.GetMeterProvider(otelServiceName)),
		otelconnect.WithoutServerPeerAttributes(),
	)
	if err != nil {
		return fmt.Errorf("creating otel interceptor: %w", err)
	}

	eventsServiceURL := sc.BaseURL
	if eventsServiceURL == "" {
		eventsServiceURL = cfg.EventsServiceURL
	}
	eventsClient := workflowconnect.NewEventsProxyServiceClient(http.DefaultClient, eventsServiceURL, connect.WithInterceptors(otelInterceptor))
	catalogCfg := catalog.GetConfig()
	cacheServiceURL := sc.BaseURL
	if cacheServiceURL == "" {
		cacheServiceURL = cfg.CacheServiceURL
	}
	cacheClient := cachecatalog.NewHTTPClient(dataStore, cacheServiceURL, catalogCfg.MaxCacheAge.Duration, connect.WithInterceptors(otelInterceptor))
	asyncCatalogClient, err := catalog.NewAsyncClient(cacheClient, *catalogCfg, promutils.NewScope("executor:catalog"))
	if err != nil {
		return fmt.Errorf("executor: failed to create catalog cache client: %w", err)
	}
	if err := asyncCatalogClient.Start(ctx); err != nil {
		return fmt.Errorf("executor: failed to start catalog cache client: %w", err)
	}

	reconciler := controller.NewTaskActionReconciler(
		mgr.GetClient(), mgr.GetScheme(), registry, dataStore, eventsClient, cfg.Cluster,
		otelutils.GetMeterProvider(otelServiceName), mgr.GetCache(),
	)
	reconciler.CatalogClient = asyncCatalogClient
	reconciler.Catalog = cacheClient
	reconciler.Recorder = mgr.GetEventRecorder("taskaction-controller")
	// Supply a ResourceManager for the webapi allocation-token path, used by connector-backed task
	// types that declare ResourceQuotas. It grants every allocation by default, matching
	// FlytePropeller with no quota backend. Swap in a real one to enforce quotas.
	reconciler.ResourceManager = plugin.NewNoopResourceManager()
	// Supply a SecretManager so connector tasks that reference secrets do not nil-deref at execution
	// time. It has no backend and fails lookups with a clear error. Swap in a real one to resolve
	// secrets.
	reconciler.SecretManager = plugin.NewNoopSecretManager()
	if cfg.MaxSystemFailures < 0 {
		return fmt.Errorf("executor: maxSystemFailures must be non-negative, got %d", cfg.MaxSystemFailures)
	}
	reconciler.MaxSystemFailures = uint32(cfg.MaxSystemFailures)
	if err := reconciler.SetupWithManager(mgr, cfg.MaxConcurrentReconciles); err != nil {
		return fmt.Errorf("executor: failed to setup controller: %w", err)
	}

	if cfg.GC.Interval.Duration > 0 {
		if cfg.GC.MaxTTL.Duration <= 0 {
			return fmt.Errorf("executor: gc.maxTTL must be positive when gc is enabled, got %v", cfg.GC.MaxTTL.Duration)
		}
		gc := controller.NewGarbageCollector(mgr.GetClient(), mgr.GetAPIReader(), cfg.GC.Interval.Duration, cfg.GC.MaxTTL.Duration)
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
