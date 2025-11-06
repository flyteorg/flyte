package main

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/spf13/cobra"
	"golang.org/x/net/http2"
	"golang.org/x/net/http2/h2c"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
	"sigs.k8s.io/controller-runtime/pkg/client"

	"github.com/flyteorg/flyte/v2/flytestdlib/config"
	"github.com/flyteorg/flyte/v2/flytestdlib/config/viper"
	"github.com/flyteorg/flyte/v2/flytestdlib/logger"
	"github.com/flyteorg/flyte/v2/gen/go/flyteidl2/workflow/workflowconnect"
	queueconfig "github.com/flyteorg/flyte/v2/queue/config"
	"github.com/flyteorg/flyte/v2/queue/k8s"
	"github.com/flyteorg/flyte/v2/queue/service"
)

var (
	cfgFile string
	rootCmd = &cobra.Command{
		Use:   "queue-service",
		Short: "Queue Service for Flyte",
		RunE: func(cmd *cobra.Command, args []string) error {
			return serve(cmd.Context())
		},
	}
)

func init() {
	rootCmd.PersistentFlags().StringVar(&cfgFile, "config", "", "config file (default is $HOME/.flyte/config.yaml)")
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

	logger.Infof(ctx, "Starting Queue Service")

	// Get configuration
	cfg := queueconfig.GetConfig()

	// Initialize Kubernetes client
	k8sClient, err := initKubernetesClient(ctx, &cfg.Kubernetes)
	if err != nil {
		return fmt.Errorf("failed to initialize Kubernetes client: %w", err)
	}

	// Initialize the executor API scheme
	if err := k8s.InitScheme(); err != nil {
		return fmt.Errorf("failed to initialize scheme: %w", err)
	}

	// Create queue client
	queueClient := k8s.NewQueueClient(k8sClient, cfg.Kubernetes.Namespace)
	logger.Infof(ctx, "Kubernetes client initialized for namespace: %s", cfg.Kubernetes.Namespace)

	// Create service
	queueSvc := service.NewQueueService(queueClient)

	// Setup HTTP server with Connect handlers
	mux := http.NewServeMux()

	// Mount the Queue Service
	path, handler := workflowconnect.NewQueueServiceHandler(queueSvc)
	mux.Handle(path, handler)

	// Add health check endpoint
	mux.HandleFunc("/healthz", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte("OK"))
	})

	// Add readiness check endpoint
	mux.HandleFunc("/readyz", func(w http.ResponseWriter, r *http.Request) {
		// Queue service is always ready (no database to check)
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte("OK"))
	})

	// Setup HTTP/2 support (required for gRPC)
	addr := fmt.Sprintf("%s:%d", cfg.Server.Host, cfg.Server.Port)
	server := &http.Server{
		Addr:    addr,
		Handler: h2c.NewHandler(mux, &http2.Server{}),
	}

	// Start server in a goroutine
	errCh := make(chan error, 1)
	go func() {
		logger.Infof(ctx, "Queue Service listening on %s", addr)
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			errCh <- fmt.Errorf("server error: %w", err)
		}
	}()

	// Wait for interrupt signal or error
	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, syscall.SIGINT, syscall.SIGTERM)

	select {
	case sig := <-sigCh:
		logger.Infof(ctx, "Received signal %v, shutting down gracefully...", sig)
	case err := <-errCh:
		return err
	}

	// Graceful shutdown
	shutdownCtx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	if err := server.Shutdown(shutdownCtx); err != nil {
		return fmt.Errorf("server shutdown failed: %w", err)
	}

	logger.Infof(ctx, "Queue Service stopped")
	return nil
}

func initConfig() error {
	// Use viper to load config
	configAccessor := viper.NewAccessor(config.Options{
		SearchPaths: []string{cfgFile, ".", "/etc/flyte/config"},
		StrictMode:  false,
	})

	return configAccessor.UpdateConfig(context.Background())
}

func initKubernetesClient(ctx context.Context, cfg *queueconfig.KubernetesConfig) (client.Client, error) {
	var restConfig *rest.Config
	var err error

	if cfg.KubeConfig != "" {
		// Use explicitly configured kubeconfig file
		logger.Infof(ctx, "Using kubeconfig from: %s", cfg.KubeConfig)
		restConfig, err = clientcmd.BuildConfigFromFlags("", cfg.KubeConfig)
		if err != nil {
			return nil, fmt.Errorf("failed to build k8s config from flags: %w", err)
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
				return nil, fmt.Errorf("failed to get Kubernetes config (tried in-cluster and default kubeconfig): %w", err)
			}

			logger.Infof(ctx, "Using default kubeconfig from standard locations (~/.kube/config)")
		}
	}

	// Create the controller-runtime client
	k8sClient, err := client.New(restConfig, client.Options{})
	if err != nil {
		return nil, fmt.Errorf("failed to create Kubernetes client: %w", err)
	}

	logger.Infof(ctx, "Kubernetes client initialized successfully")
	return k8sClient, nil
}
