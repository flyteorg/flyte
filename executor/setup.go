package executor

import (
	"context"
	"crypto/tls"
	"fmt"
	"net/http"
	"os"
	"time"

	"connectrpc.com/connect"
	"connectrpc.com/otelconnect"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/apimachinery/pkg/runtime"
	utilruntime "k8s.io/apimachinery/pkg/util/runtime"
	"k8s.io/client-go/informers"
	"k8s.io/client-go/kubernetes"
	clientgoscheme "k8s.io/client-go/kubernetes/scheme"
	k8scache "k8s.io/client-go/tools/cache"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/cache"
	"sigs.k8s.io/controller-runtime/pkg/client"
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
	cachecatalog "github.com/flyteorg/flyte/v2/flyteplugins/go/tasks/pluginmachinery/catalog/cache_service"
	"github.com/flyteorg/flyte/v2/flyteplugins/go/tasks/pluginmachinery/flytek8s"
	webhookConfig "github.com/flyteorg/flyte/v2/flyteplugins/go/tasks/pluginmachinery/secret/config"
	connectorplugin "github.com/flyteorg/flyte/v2/flyteplugins/go/tasks/plugins/webapi/connector"
	"github.com/flyteorg/flyte/v2/flytestdlib/app"
	"github.com/flyteorg/flyte/v2/flytestdlib/otelutils"
	"github.com/flyteorg/flyte/v2/flytestdlib/promutils"
	"github.com/flyteorg/flyte/v2/flytestdlib/serviceclient"
	"github.com/flyteorg/flyte/v2/flytestdlib/storage"
	"github.com/flyteorg/flyte/v2/gen/go/flyteidl2/workflow/workflowconnect"

	_ "github.com/flyteorg/flyte/v2/executor/plugins"
	_ "github.com/flyteorg/flyte/v2/flyteplugins/go/tasks/plugins/k8s/clustered"
	_ "github.com/flyteorg/flyte/v2/flyteplugins/go/tasks/plugins/k8s/pod"
)

var scheme = runtime.NewScheme()

const otelServiceName = "executor"

// podTemplateSyncTimeout bounds the initial PodTemplate informer sync; matches
// controller-runtime's default cache sync timeout.
const podTemplateSyncTimeout = 2 * time.Minute

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

	// controller-runtime caches every watched object across all namespaces, so the
	// manager would otherwise hold every Pod in the cluster. On a large multi-tenant
	// cluster that costs several GB of resident memory and OOMKills the executor.
	//
	// Restrict the cache to the Pods belonging to this executor. NewTaskExecutionMetadata
	// puts the label on every task's execution labels, which each plugin merges into the
	// object it builds, so both the Pods the executor creates directly and the ones an
	// operator derives from a plugin's CRD (Ray head/worker, Kubeflow replicas, Dask
	// scheduler/worker) carry it.
	cacheOptions := cache.Options{
		ByObject: map[client.Object]cache.ByObject{
			&corev1.Pod{}: {
				Label: labels.SelectorFromSet(labels.Set{
					flytek8s.ManagedLabelKey: flytek8s.ManagedLabelValue,
				}),
			},
		},
	}

	mgr, err := ctrl.NewManager(sc.K8sConfig, ctrl.Options{
		Scheme:                 scheme,
		Metrics:                metricsServerOptions,
		Cache:                  cacheOptions,
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

	informerFactory := informers.NewSharedInformerFactory(kubeClient, 0)
	if err := watchPodTemplates(informerFactory, podNamespace); err != nil {
		return fmt.Errorf("executor: failed to register PodTemplate event handler: %w", err)
	}
	sc.AddWorker("podtemplate-informer", func(ctx context.Context) error {
		informerFactory.Start(ctx.Done())
		syncCtx, cancel := context.WithTimeout(ctx, podTemplateSyncTimeout)
		defer cancel()
		if !k8scache.WaitForCacheSync(syncCtx.Done(), informerFactory.Core().V1().PodTemplates().Informer().HasSynced) {
			return fmt.Errorf("executor: PodTemplate informer failed to sync within %v; "+
				"verify the service account can get/list/watch core/v1 podtemplates", podTemplateSyncTimeout)
		}
		<-ctx.Done()
		return nil
	})

	executorScope := promutils.NewScope("executor")

	podMutator, err := webhookPkg.Setup(ctx, kubeClient, wCfg, podNamespace, executorScope.NewSubScope("webhook"), mgr)
	if err != nil {
		return fmt.Errorf("executor: webhook setup failed: %w", err)
	}

	// Serve cache invalidation so the secret service can drop cached secret values on write.
	if wCfg.CacheInvalidationPort > 0 {
		sc.AddWorker("secret-cache-invalidation", func(ctx context.Context) error {
			return webhookPkg.StartCacheInvalidationServer(ctx, wCfg.CacheInvalidationPort, podMutator.SecretsMutator())
		})
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

	eventsServiceCfg := cfg.EventsService
	cacheServiceCfg := cfg.CacheService
	if sc.BaseURL != "" {
		eventsServiceCfg.URL = sc.BaseURL
		cacheServiceCfg.URL = sc.BaseURL
	}

	eventsHTTPClient, err := serviceclient.NewHTTPClient(ctx, http.DefaultClient, eventsServiceCfg)
	if err != nil {
		return fmt.Errorf("executor: configure events service client: %w", err)
	}
	cacheHTTPClient, err := serviceclient.NewHTTPClient(ctx, http.DefaultClient, cacheServiceCfg)
	if err != nil {
		return fmt.Errorf("executor: configure cache service client: %w", err)
	}
	eventsClient := workflowconnect.NewEventsProxyServiceClient(eventsHTTPClient, eventsServiceCfg.URL, connect.WithInterceptors(otelInterceptor))
	catalogCfg := catalog.GetConfig()
	cacheClient := cachecatalog.NewClient(cacheHTTPClient, dataStore, cacheServiceCfg.URL, catalogCfg.MaxCacheAge.Duration, connect.WithInterceptors(otelInterceptor))
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
	if cfg.RequeueDuration.Duration < 0 {
		return fmt.Errorf("executor: requeueDuration must not be negative, got %v", cfg.RequeueDuration.Duration)
	}
	reconciler.RequeueDuration = cfg.RequeueDuration.Duration
	if err := reconciler.SetupWithManager(mgr, cfg.MaxConcurrentReconciles); err != nil {
		return fmt.Errorf("executor: failed to setup controller: %w", err)
	}

	if cfg.GC.Interval.Duration > 0 {
		gc := controller.NewGarbageCollector(mgr.GetClient(), mgr.GetAPIReader(), cfg.GC.Interval.Duration, cfg.GC.MaxTTL.Duration, otelutils.GetMeterProvider(otelServiceName))
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
