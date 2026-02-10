package main

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"sync"
	"syscall"
	"time"

	"github.com/spf13/cobra"
	"golang.org/x/net/http2"
	"golang.org/x/net/http2/h2c"
	"gorm.io/gorm"
	corev1 "k8s.io/api/core/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	utilruntime "k8s.io/apimachinery/pkg/util/runtime"
	clientgoscheme "k8s.io/client-go/kubernetes/scheme"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/healthz"
	"sigs.k8s.io/controller-runtime/pkg/log/zap"
	metricsserver "sigs.k8s.io/controller-runtime/pkg/metrics/server"

	dataproxyconfig "github.com/flyteorg/flyte/v2/dataproxy/config"
	dataproxyservice "github.com/flyteorg/flyte/v2/dataproxy/service"
	flyteorgv1 "github.com/flyteorg/flyte/v2/executor/api/v1"
	executorcontroller "github.com/flyteorg/flyte/v2/executor/pkg/controller"
	"github.com/flyteorg/flyte/v2/executor/pkg/plugin"
	"github.com/flyteorg/flyte/v2/flyteplugins/go/tasks/pluginmachinery"
	"github.com/flyteorg/flyte/v2/flytestdlib/config"
	"github.com/flyteorg/flyte/v2/flytestdlib/config/viper"
	"github.com/flyteorg/flyte/v2/flytestdlib/contextutils"
	"github.com/flyteorg/flyte/v2/flytestdlib/database"
	"github.com/flyteorg/flyte/v2/flytestdlib/logger"
	"github.com/flyteorg/flyte/v2/flytestdlib/promutils"
	"github.com/flyteorg/flyte/v2/flytestdlib/promutils/labeled"
	"github.com/flyteorg/flyte/v2/flytestdlib/storage"
	"github.com/flyteorg/flyte/v2/gen/go/flyteidl2/dataproxy/dataproxyconnect"
	"github.com/flyteorg/flyte/v2/gen/go/flyteidl2/workflow/workflowconnect"
	managerconfig "github.com/flyteorg/flyte/v2/manager/config"
	queuek8s "github.com/flyteorg/flyte/v2/queue/k8s"
	queueservice "github.com/flyteorg/flyte/v2/queue/service"
	runsconfig "github.com/flyteorg/flyte/v2/runs/config"
	"github.com/flyteorg/flyte/v2/runs/migrations"
	"github.com/flyteorg/flyte/v2/runs/repository"
	runsservice "github.com/flyteorg/flyte/v2/runs/service"
	statek8s "github.com/flyteorg/flyte/v2/state/k8s"
	stateservice "github.com/flyteorg/flyte/v2/state/service"

	// Plugin registrations -- blank imports trigger init() which registers plugins with the global registry.
	_ "github.com/flyteorg/flyte/v2/flyteplugins/go/tasks/plugins/k8s/pod"
)

var (
	cfgFile        string
	configAccessor config.Accessor
	scheme         = runtime.NewScheme()
)

func init() {
	// Register Kubernetes schemes
	utilruntime.Must(clientgoscheme.AddToScheme(scheme))
	utilruntime.Must(flyteorgv1.AddToScheme(scheme))
}

func newRootCmd() *cobra.Command {
	rootCmd := &cobra.Command{
		Use:   "flyte",
		Short: "Unified Flyte Service - Runs all services and operator",
		PersistentPreRunE: func(cmd *cobra.Command, args []string) error {
			return initConfig(cmd)
		},
		RunE: func(cmd *cobra.Command, args []string) error {
			return serve(cmd.Context())
		},
	}

	rootCmd.PersistentFlags().StringVar(&cfgFile, "config", "", "config file (default is ./config.yaml)")
	configAccessor = viper.NewAccessor(config.Options{StrictMode: false})
	configAccessor.InitializePflags(rootCmd.PersistentFlags())

	return rootCmd
}

func main() {
	if err := newRootCmd().Execute(); err != nil {
		os.Exit(1)
	}
}

func serve(ctx context.Context) error {
	// Initialize logger
	logConfig := logger.GetConfig()
	if err := logger.SetConfig(logConfig); err != nil {
		return fmt.Errorf("failed to initialize logger: %w", err)
	}

	// Set controller-runtime logger
	ctrl.SetLogger(zap.New(zap.UseDevMode(true)))

	logger.Infof(ctx, "Starting Flyte Manager (unified binary)")

	// Get configuration (use defaults if config doesn't load)
	cfg := &managerconfig.Config{
		Server: managerconfig.ServerConfig{
			Host: "0.0.0.0",
			Port: 8090, // Single port for all Connect services
		},
		Executor: managerconfig.ExecutorConfig{
			HealthProbePort: 8081,
		},
		Kubernetes: managerconfig.KubernetesConfig{
			Namespace: "flyte",
		},
	}

	dbCfg := &database.DbConfig{
		SQLite: database.SQLiteConfig{
			File: "flyte.db",
		},
	}

	// Initialize database
	db, err := initDB(ctx, dbCfg)
	if err != nil {
		return fmt.Errorf("failed to initialize database: %w", err)
	}

	// Run migrations
	logger.Infof(ctx, "Running database migrations")
	if err := migrations.RunMigrations(db); err != nil {
		return fmt.Errorf("failed to run migrations: %w", err)
	}

	// Create repository
	repo := repository.NewRepository(db)

	// Initialize the executor API scheme
	if err := queuek8s.InitScheme(); err != nil {
		return fmt.Errorf("failed to initialize scheme: %w", err)
	}

	// Initialize Kubernetes client with watch support
	k8sClient, k8sConfig, err := initKubernetesClient(ctx, &cfg.Kubernetes)
	if err != nil {
		return fmt.Errorf("failed to initialize Kubernetes client: %w", err)
	}

	// Create a client.Client from the WithWatch client for services that don't need watch
	var k8sClientWithoutWatch client.Client = k8sClient

	// Initialize labeled metrics (required for storage)
	labeled.SetMetricKeys(contextutils.ProjectKey, contextutils.DomainKey, contextutils.WorkflowIDKey, contextutils.TaskIDKey)

	// Initialize storage
	storageCfg := storage.GetConfig()
	metricsScope := promutils.NewTestScope()
	dataStore, err := storage.NewDataStore(storageCfg, metricsScope)
	if err != nil {
		return fmt.Errorf("failed to initialize storage: %w", err)
	}
	logger.Infof(ctx, "Storage initialized with type: %s", storageCfg.Type)

	// Create queue client (for Kubernetes operations)
	queueK8sClient := queuek8s.NewQueueClient(k8sClientWithoutWatch, cfg.Kubernetes.Namespace)
	logger.Infof(ctx, "Kubernetes client initialized for namespace: %s", cfg.Kubernetes.Namespace)

	// Create state client (K8s-based, for watching TaskAction CRs)
	stateK8sClient := statek8s.NewStateClient(k8sClient, cfg.Kubernetes.Namespace, 100)

	// Start watching TaskActions for state service
	if err := stateK8sClient.StartWatching(ctx); err != nil {
		return fmt.Errorf("failed to start TaskAction watcher: %w", err)
	}
	defer stateK8sClient.StopWatching()

	// Create queue service client (points to same server)
	queueClient := workflowconnect.NewQueueServiceClient(
		http.DefaultClient,
		fmt.Sprintf("http://localhost:%d", cfg.Server.Port),
	)

	// Create all services
	runsCfg := runsconfig.GetConfig()
	runsSvc := runsservice.NewRunService(repo, queueClient, runsCfg.StoragePrefix, dataStore)
	stateSvc := stateservice.NewStateService(stateK8sClient) // K8s-based state service
	queueSvc := queueservice.NewQueueService(queueK8sClient)
	dataProxyCfg := dataproxyconfig.GetConfig()
	dataProxySvc := dataproxyservice.NewService(*dataProxyCfg, dataStore)

	// Setup single HTTP server with all services mounted
	mux := http.NewServeMux()

	// Mount all Connect services on the same mux
	runsPath, runsHandler := workflowconnect.NewRunServiceHandler(runsSvc)
	mux.Handle(runsPath, runsHandler)
	logger.Infof(ctx, "Mounted RunService at %s", runsPath)

	statePath, stateHandler := workflowconnect.NewStateServiceHandler(stateSvc)
	mux.Handle(statePath, stateHandler)
	logger.Infof(ctx, "Mounted StateService at %s (K8s-based)", statePath)

	queuePath, queueHandler := workflowconnect.NewQueueServiceHandler(queueSvc)
	mux.Handle(queuePath, queueHandler)
	logger.Infof(ctx, "Mounted QueueService at %s", queuePath)

	dataProxyPath, dataProxyHandler := dataproxyconnect.NewDataProxyServiceHandler(dataProxySvc)
	mux.Handle(dataProxyPath, dataProxyHandler)
	logger.Infof(ctx, "Mounted DataProxyService at %s", dataProxyPath)

	// Health checks
	mux.HandleFunc("/healthz", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte("OK"))
	})

	mux.HandleFunc("/readyz", func(w http.ResponseWriter, r *http.Request) {
		// Check database
		sqlDB, err := db.DB()
		if err != nil || sqlDB.Ping() != nil {
			w.WriteHeader(http.StatusServiceUnavailable)
			_, _ = w.Write([]byte("Database unavailable"))
			return
		}
		// Check storage
		baseContainer := dataStore.GetBaseContainerFQN(r.Context())
		if baseContainer == "" {
			w.WriteHeader(http.StatusServiceUnavailable)
			_, _ = w.Write([]byte("Storage connection error"))
			return
		}
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte("OK"))
	})

	// Wait group for both server and executor
	var wg sync.WaitGroup
	errCh := make(chan error, 2)

	// 1. Start unified HTTP server with all Connect services
	wg.Add(1)
	go func() {
		defer wg.Done()
		addr := fmt.Sprintf("%s:%d", cfg.Server.Host, cfg.Server.Port)
		server := &http.Server{
			Addr:    addr,
			Handler: h2c.NewHandler(corsMiddleware(mux), &http2.Server{}),
		}

		logger.Infof(ctx, "Flyte Connect Server listening on %s", addr)
		logger.Infof(ctx, "All services (Runs, State, Queue, DataProxy) available on port %d", cfg.Server.Port)
		if err := server.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			errCh <- fmt.Errorf("server error: %w", err)
		}
	}()

	// 2. Start Executor/Operator
	wg.Add(1)
	go func() {
		defer wg.Done()
		logger.Infof(ctx, "Starting Executor/Operator")

		// Metrics endpoint is enabled in 'config/default/kustomization.yaml'. The Metrics options configure the server.
		// More info:
		// - https://pkg.go.dev/sigs.k8s.io/controller-runtime@v0.22.1/pkg/metrics/server
		// - https://book.kubebuilder.io/reference/metrics.html
		metricsServerOptions := metricsserver.Options{
			BindAddress: ":10254",
		}

		// Create controller manager
		mgr, err := ctrl.NewManager(k8sConfig, ctrl.Options{
			Scheme:                 scheme,
			HealthProbeBindAddress: fmt.Sprintf(":%d", cfg.Executor.HealthProbePort),
			Metrics:                metricsServerOptions,
			LeaderElection:         false, // Single instance for now
		})
		if err != nil {
			errCh <- fmt.Errorf("failed to create controller manager: %w", err)
			return
		}

		// Create SetupContext and initialize plugin registry
		setupCtx := plugin.NewSetupContext(
			mgr, // KubeClient
			nil, // SecretManager -- TODO: implement
			nil, // ResourceRegistrar -- not needed for manager
			nil, // EnqueueOwner -- not needed, controller-runtime handles reconciliation
			nil, // EnqueueLabels
			"TaskAction",
			promutils.NewScope("executor"),
		)
		registry := plugin.NewRegistry(setupCtx, pluginmachinery.PluginRegistry())
		if err := registry.Initialize(ctx); err != nil {
			errCh <- fmt.Errorf("failed to initialize plugin registry: %w", err)
			return
		}

		// Setup TaskAction controller
		reconciler := executorcontroller.NewTaskActionReconciler(
			mgr.GetClient(),
			mgr.GetScheme(),
			registry,
			dataStore,
		)
		reconciler.Recorder = mgr.GetEventRecorderFor("taskaction-controller")
		if err := reconciler.SetupWithManager(mgr); err != nil {
			errCh <- fmt.Errorf("failed to setup controller: %w", err)
			return
		}

		// Add health checks
		if err := mgr.AddHealthzCheck("healthz", healthz.Ping); err != nil {
			errCh <- fmt.Errorf("failed to add health check: %w", err)
			return
		}
		if err := mgr.AddReadyzCheck("readyz", healthz.Ping); err != nil {
			errCh <- fmt.Errorf("failed to add ready check: %w", err)
			return
		}

		logger.Infof(ctx, "Executor controller starting (updates TaskAction CRs directly)")
		if err := mgr.Start(ctrl.SetupSignalHandler()); err != nil {
			errCh <- fmt.Errorf("executor controller error: %w", err)
		}
	}()

	// Wait for interrupt signal or error
	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, syscall.SIGINT, syscall.SIGTERM)

	select {
	case sig := <-sigCh:
		logger.Infof(ctx, "Received signal %v, shutting down gracefully...", sig)
	case err := <-errCh:
		logger.Errorf(ctx, "Service error: %v", err)
		return err
	}

	logger.Infof(ctx, "Flyte Manager stopped")
	return nil
}

func initConfig(cmd *cobra.Command) error {
	configAccessor = viper.NewAccessor(config.Options{
		SearchPaths: []string{cfgFile, ".", "/etc/flyte/config"},
		StrictMode:  false,
	})

	// Traverse to root command
	rootCmd := cmd
	for rootCmd.Parent() != nil {
		rootCmd = rootCmd.Parent()
	}

	configAccessor.InitializePflags(rootCmd.PersistentFlags())

	return configAccessor.UpdateConfig(context.Background())
}

func initDB(ctx context.Context, cfg *database.DbConfig) (*gorm.DB, error) {
	logCfg := logger.GetConfig()

	db, err := database.GetDB(ctx, cfg, logCfg)
	if err != nil {
		return nil, fmt.Errorf("failed to initialize database: %w", err)
	}

	logger.Infof(ctx, "Database connection established")
	return db, nil
}

func initKubernetesClient(ctx context.Context, cfg *managerconfig.KubernetesConfig) (client.WithWatch, *rest.Config, error) {
	var restConfig *rest.Config
	var err error

	if cfg.KubeConfig != "" {
		// Use explicitly configured kubeconfig file
		logger.Infof(ctx, "Using kubeconfig from: %s", cfg.KubeConfig)
		restConfig, err = clientcmd.BuildConfigFromFlags("", cfg.KubeConfig)
		if err != nil {
			return nil, nil, fmt.Errorf("failed to build k8s config from flags: %w", err)
		}
	} else {
		// Try in-cluster config first
		logger.Infof(ctx, "Attempting to use in-cluster Kubernetes configuration")
		restConfig, err = rest.InClusterConfig()

		if err != nil {
			// Fall back to default kubeconfig location
			logger.Infof(ctx, "In-cluster config not available, falling back to default kubeconfig")
			loadingRules := clientcmd.NewDefaultClientConfigLoadingRules()
			configOverrides := &clientcmd.ConfigOverrides{}
			kubeConfig := clientcmd.NewNonInteractiveDeferredLoadingClientConfig(loadingRules, configOverrides)
			restConfig, err = kubeConfig.ClientConfig()

			if err != nil {
				return nil, nil, fmt.Errorf("failed to get Kubernetes config: %w", err)
			}

			logger.Infof(ctx, "Using default kubeconfig from standard locations (~/.kube/config)")
		}
	}

	// Apply rate limits
	if cfg.QPS > 0 {
		restConfig.QPS = float32(cfg.QPS)
	}
	if cfg.Burst > 0 {
		restConfig.Burst = cfg.Burst
	}
	if cfg.Timeout != "" {
		d, err := time.ParseDuration(cfg.Timeout)
		if err != nil {
			return nil, nil, fmt.Errorf("invalid kubernetes.timeout %q: %w", cfg.Timeout, err)
		}
		restConfig.Timeout = d
	}

	logger.Infof(ctx, "K8s client rate limits: QPS=%.0f, Burst=%d", restConfig.QPS, restConfig.Burst)

	// Create the controller-runtime client with watch support
	k8sClient, err := client.NewWithWatch(restConfig, client.Options{Scheme: scheme})
	if err != nil {
		return nil, nil, fmt.Errorf("failed to create Kubernetes client: %w", err)
	}

	logger.Infof(ctx, "Kubernetes client initialized successfully")

	// Ensure the namespace exists
	if err := ensureNamespaceExists(ctx, k8sClient, cfg.Namespace); err != nil {
		return nil, nil, fmt.Errorf("failed to ensure namespace exists: %w", err)
	}

	return k8sClient, restConfig, nil
}

func ensureNamespaceExists(ctx context.Context, k8sClient client.Client, namespaceName string) error {
	namespace := &corev1.Namespace{}
	err := k8sClient.Get(ctx, types.NamespacedName{Name: namespaceName}, namespace)

	if err == nil {
		// Namespace already exists
		logger.Infof(ctx, "Namespace '%s' already exists", namespaceName)
		return nil
	}

	if !apierrors.IsNotFound(err) {
		// Some other error occurred
		return fmt.Errorf("failed to check if namespace exists: %w", err)
	}

	// Namespace doesn't exist, create it
	logger.Infof(ctx, "Creating namespace '%s'", namespaceName)
	namespace = &corev1.Namespace{
		ObjectMeta: metav1.ObjectMeta{
			Name: namespaceName,
		},
	}

	if err := k8sClient.Create(ctx, namespace); err != nil {
		return fmt.Errorf("failed to create namespace: %w", err)
	}

	logger.Infof(ctx, "Successfully created namespace '%s'", namespaceName)
	return nil
}

// corsMiddleware wraps an http.Handler with permissive CORS headers for local
// development (UI on localhost:8080 → manager on localhost:8090).
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
