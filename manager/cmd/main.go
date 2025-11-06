package main

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"sync"
	"syscall"

	"github.com/spf13/cobra"
	"golang.org/x/net/http2"
	"golang.org/x/net/http2/h2c"
	"gorm.io/gorm"
	"k8s.io/apimachinery/pkg/runtime"
	utilruntime "k8s.io/apimachinery/pkg/util/runtime"
	clientgoscheme "k8s.io/client-go/kubernetes/scheme"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/healthz"
	"sigs.k8s.io/controller-runtime/pkg/log/zap"

	flyteorgv1 "github.com/flyteorg/flyte/v2/executor/api/v1"
	executorcontroller "github.com/flyteorg/flyte/v2/executor/pkg/controller"
	"github.com/flyteorg/flyte/v2/flytestdlib/config"
	"github.com/flyteorg/flyte/v2/flytestdlib/config/viper"
	"github.com/flyteorg/flyte/v2/flytestdlib/database"
	"github.com/flyteorg/flyte/v2/flytestdlib/logger"
	"github.com/flyteorg/flyte/v2/gen/go/flyteidl2/workflow/workflowconnect"
	managerconfig "github.com/flyteorg/flyte/v2/manager/config"
	queuek8s "github.com/flyteorg/flyte/v2/queue/k8s"
	queueservice "github.com/flyteorg/flyte/v2/queue/service"
	"github.com/flyteorg/flyte/v2/runs/migrations"
	"github.com/flyteorg/flyte/v2/runs/repository"
	runsservice "github.com/flyteorg/flyte/v2/runs/service"
)

var (
	cfgFile string
	rootCmd = &cobra.Command{
		Use:   "flyte",
		Short: "Unified Flyte Service - Runs all services and operator",
		RunE: func(cmd *cobra.Command, args []string) error {
			return serve(cmd.Context())
		},
	}
	scheme = runtime.NewScheme()
)

func init() {
	rootCmd.PersistentFlags().StringVar(&cfgFile, "config", "", "config file (default is ./config.yaml)")

	// Register Kubernetes schemes
	utilruntime.Must(clientgoscheme.AddToScheme(scheme))
	utilruntime.Must(flyteorgv1.AddToScheme(scheme))
}

func main() {
	if err := rootCmd.Execute(); err != nil {
		os.Exit(1)
	}
}

func serve(ctx context.Context) error {
	// Initialize config
	if err := initConfig(); err != nil {
		return fmt.Errorf("failed to initialize config: %w", err)
	}

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
		RunsService: managerconfig.ServiceConfig{
			Host: "0.0.0.0",
			Port: 8090,
		},
		QueueService: managerconfig.ServiceConfig{
			Host: "0.0.0.0",
			Port: 8089,
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
	repo := repository.NewPostgresRepository(db)

	// Initialize Kubernetes client
	k8sClient, k8sConfig, err := initKubernetesClient(ctx, &cfg.Kubernetes)
	if err != nil {
		return fmt.Errorf("failed to initialize Kubernetes client: %w", err)
	}

	// Initialize the executor API scheme
	if err := queuek8s.InitScheme(); err != nil {
		return fmt.Errorf("failed to initialize scheme: %w", err)
	}

	// Create queue client (for Kubernetes operations)
	queueK8sClient := queuek8s.NewQueueClient(k8sClient, cfg.Kubernetes.Namespace)
	logger.Infof(ctx, "Kubernetes client initialized for namespace: %s", cfg.Kubernetes.Namespace)

	// Wait group for all services
	var wg sync.WaitGroup
	errCh := make(chan error, 3)

	// 1. Start Runs Service (includes State Service)
	wg.Add(1)
	go func() {
		defer wg.Done()
		logger.Infof(ctx, "Starting Runs Service on %s:%d", cfg.RunsService.Host, cfg.RunsService.Port)

		// Create queue service client (for enqueuing actions)
		queueClient := workflowconnect.NewQueueServiceClient(
			http.DefaultClient,
			fmt.Sprintf("http://localhost:%d", cfg.QueueService.Port),
		)

		// Create services
		runsSvc := runsservice.NewRunService(repo, queueClient)
		stateSvc := runsservice.NewStateService(repo)

		// Setup HTTP server
		mux := http.NewServeMux()

		// Mount services
		runsPath, runsHandler := workflowconnect.NewRunServiceHandler(runsSvc)
		mux.Handle(runsPath, runsHandler)

		statePath, stateHandler := workflowconnect.NewStateServiceHandler(stateSvc)
		mux.Handle(statePath, stateHandler)

		// Health checks
		mux.HandleFunc("/healthz", func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusOK)
			_, _ = w.Write([]byte("OK"))
		})

		mux.HandleFunc("/readyz", func(w http.ResponseWriter, r *http.Request) {
			sqlDB, err := db.DB()
			if err != nil || sqlDB.Ping() != nil {
				w.WriteHeader(http.StatusServiceUnavailable)
				_, _ = w.Write([]byte("Database unavailable"))
				return
			}
			w.WriteHeader(http.StatusOK)
			_, _ = w.Write([]byte("OK"))
		})

		addr := fmt.Sprintf("%s:%d", cfg.RunsService.Host, cfg.RunsService.Port)
		server := &http.Server{
			Addr:    addr,
			Handler: h2c.NewHandler(mux, &http2.Server{}),
		}

		logger.Infof(ctx, "Runs Service listening on %s", addr)
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			errCh <- fmt.Errorf("runs service error: %w", err)
		}
	}()

	// 2. Start Queue Service
	wg.Add(1)
	go func() {
		defer wg.Done()
		logger.Infof(ctx, "Starting Queue Service on %s:%d", cfg.QueueService.Host, cfg.QueueService.Port)

		// Create queue service
		queueSvc := queueservice.NewQueueService(queueK8sClient)

		// Setup HTTP server
		mux := http.NewServeMux()

		// Mount queue service
		path, handler := workflowconnect.NewQueueServiceHandler(queueSvc)
		mux.Handle(path, handler)

		// Health checks
		mux.HandleFunc("/healthz", func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusOK)
			_, err = w.Write([]byte("OK"))
			if err != nil {
				logger.Info(ctx, err)
			}
		})

		mux.HandleFunc("/readyz", func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusOK)
			_, err = w.Write([]byte("OK"))
			if err != nil {
				logger.Info(ctx, err)
			}
		})

		addr := fmt.Sprintf("%s:%d", cfg.QueueService.Host, cfg.QueueService.Port)
		server := &http.Server{
			Addr:    addr,
			Handler: h2c.NewHandler(mux, &http2.Server{}),
		}

		logger.Infof(ctx, "Queue Service listening on %s", addr)
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			errCh <- fmt.Errorf("queue service error: %w", err)
		}
	}()

	// 3. Start Executor/Operator
	wg.Add(1)
	go func() {
		defer wg.Done()
		logger.Infof(ctx, "Starting Executor/Operator")

		// Create controller manager
		mgr, err := ctrl.NewManager(k8sConfig, ctrl.Options{
			Scheme:                 scheme,
			HealthProbeBindAddress: fmt.Sprintf(":%d", cfg.Executor.HealthProbePort),
			LeaderElection:         false, // Single instance for now
		})
		if err != nil {
			errCh <- fmt.Errorf("failed to create controller manager: %w", err)
			return
		}

		// Setup TaskAction controller
		stateServiceURL := fmt.Sprintf("http://localhost:%d", cfg.RunsService.Port)
		if err := executorcontroller.NewTaskActionReconciler(
			mgr.GetClient(),
			mgr.GetScheme(),
			stateServiceURL,
		).SetupWithManager(mgr); err != nil {
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

		logger.Infof(ctx, "Executor controller starting")
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

func initConfig() error {
	configAccessor := viper.NewAccessor(config.Options{
		SearchPaths: []string{cfgFile, ".", "/etc/flyte/config"},
		StrictMode:  false,
	})

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

func initKubernetesClient(ctx context.Context, cfg *managerconfig.KubernetesConfig) (client.Client, *rest.Config, error) {
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

	// Create the controller-runtime client
	k8sClient, err := client.New(restConfig, client.Options{Scheme: scheme})
	if err != nil {
		return nil, nil, fmt.Errorf("failed to create Kubernetes client: %w", err)
	}

	logger.Infof(ctx, "Kubernetes client initialized successfully")
	return k8sClient, restConfig, nil
}
