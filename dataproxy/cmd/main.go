package main

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/spf13/cobra"
	"golang.org/x/net/http2"
	"golang.org/x/net/http2/h2c"

	"github.com/flyteorg/flyte/v2/dataproxy/config"
	"github.com/flyteorg/flyte/v2/dataproxy/service"
	stdconfig "github.com/flyteorg/flyte/v2/flytestdlib/config"
	"github.com/flyteorg/flyte/v2/flytestdlib/config/viper"
	"github.com/flyteorg/flyte/v2/flytestdlib/contextutils"
	"github.com/flyteorg/flyte/v2/flytestdlib/logger"
	"github.com/flyteorg/flyte/v2/flytestdlib/promutils"
	"github.com/flyteorg/flyte/v2/flytestdlib/promutils/labeled"
	"github.com/flyteorg/flyte/v2/flytestdlib/storage"
	"github.com/flyteorg/flyte/v2/gen/go/flyteidl2/dataproxy/dataproxyconnect"
)

var (
	cfgFile string
	port    int
	host    string

	rootCmd = &cobra.Command{
		Use:   "dataproxy-service",
		Short: "Data Proxy Service for Flyte",
		RunE: func(cmd *cobra.Command, args []string) error {
			return serve(cmd.Context())
		},
	}
)

func init() {
	rootCmd.PersistentFlags().StringVar(&cfgFile, "config", "", "config file (default is $HOME/.flyte/config.yaml)")
	rootCmd.PersistentFlags().IntVar(&port, "port", 8088, "server port")
	rootCmd.PersistentFlags().StringVar(&host, "host", "0.0.0.0", "server host")
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

	logger.Infof(ctx, "Starting Data Proxy Service")

	// Initialize labeled metrics (required before using storage)
	labeled.SetMetricKeys(contextutils.ProjectKey, contextutils.DomainKey, contextutils.WorkflowIDKey, contextutils.TaskIDKey)

	// Get configuration
	dataProxyCfg := config.GetConfig()
	storageCfg := storage.GetConfig()

	// Initialize storage with metrics scope
	metricsScope := promutils.NewTestScope()
	dataStore, err := storage.NewDataStore(storageCfg, metricsScope)
	if err != nil {
		return fmt.Errorf("failed to initialize storage: %w", err)
	}
	logger.Infof(ctx, "Storage initialized with type: %s", storageCfg.Type)

	// Create data proxy service
	dataProxySvc := service.NewService(*dataProxyCfg, dataStore)

	// Setup HTTP server with Connect handlers
	mux := http.NewServeMux()

	// Mount the Data Proxy Service
	path, handler := dataproxyconnect.NewDataProxyServiceHandler(dataProxySvc)
	mux.Handle(path, handler)
	logger.Infof(ctx, "Mounted DataProxyService at %s", path)

	// Add health check endpoint
	mux.HandleFunc("/healthz", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte("OK"))
	})

	// Add readiness check endpoint
	mux.HandleFunc("/readyz", func(w http.ResponseWriter, r *http.Request) {
		// Check storage connection by getting base container
		baseContainer := dataStore.GetBaseContainerFQN(r.Context())
		if baseContainer == "" {
			w.WriteHeader(http.StatusServiceUnavailable)
			_, _ = w.Write([]byte("Storage connection error"))
			return
		}
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte("OK"))
	})

	// Setup HTTP/2 support (required for gRPC and Connect)
	addr := fmt.Sprintf("%s:%d", host, port)
	server := &http.Server{
		Addr:    addr,
		Handler: h2c.NewHandler(mux, &http2.Server{}),
	}

	// Start server in a goroutine
	errCh := make(chan error, 1)
	go func() {
		logger.Infof(ctx, "Data Proxy Service listening on %s", addr)
		if err := server.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
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

	logger.Infof(ctx, "Data Proxy Service stopped")
	return nil
}

func initConfig() error {
	// Use viper to load config
	configAccessor := viper.NewAccessor(stdconfig.Options{
		SearchPaths: []string{cfgFile, ".", "/etc/flyte/config"},
		StrictMode:  false,
	})

	return configAccessor.UpdateConfig(context.Background())
}
