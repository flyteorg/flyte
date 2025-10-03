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
	"gorm.io/gorm"

	"github.com/flyteorg/flyte/flytestdlib/config"
	"github.com/flyteorg/flyte/flytestdlib/config/viper"
	"github.com/flyteorg/flyte/flytestdlib/database"
	"github.com/flyteorg/flyte/flytestdlib/logger"
	"github.com/flyteorg/flyte/v2/gen/go/flyteidl2/workflow/workflowconnect"
	queueconfig "github.com/flyteorg/flyte/v2/queue/config"
	"github.com/flyteorg/flyte/v2/queue/migrations"
	"github.com/flyteorg/flyte/v2/queue/repository"
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
	// Logger is already initialized via SetConfig

	logger.Infof(ctx, "Starting Queue Service")

	// Get configuration
	cfg := queueconfig.GetConfig()
	dbCfg := database.GetConfig()

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

	// Create service
	queueSvc := service.NewQueueService(repo)

	// Setup HTTP server with Connect handlers
	mux := http.NewServeMux()

	// Mount the Queue Service
	path, handler := workflowconnect.NewQueueServiceHandler(queueSvc)
	mux.Handle(path, handler)

	// Add health check endpoint
	mux.HandleFunc("/healthz", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("OK"))
	})

	// Add readiness check endpoint
	mux.HandleFunc("/readyz", func(w http.ResponseWriter, r *http.Request) {
		// Check database connection
		sqlDB, err := db.DB()
		if err != nil {
			w.WriteHeader(http.StatusServiceUnavailable)
			w.Write([]byte("Database connection error"))
			return
		}
		if err := sqlDB.Ping(); err != nil {
			w.WriteHeader(http.StatusServiceUnavailable)
			w.Write([]byte("Database ping failed"))
			return
		}
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("OK"))
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

func initDB(ctx context.Context, cfg *database.DbConfig) (*gorm.DB, error) {
	gormConfig := &gorm.Config{
		// We can add GORM-specific configuration here
	}

	// Create database if it doesn't exist
	db, err := database.CreatePostgresDbIfNotExists(ctx, gormConfig, cfg.Postgres)
	if err != nil {
		return nil, fmt.Errorf("failed to create database: %w", err)
	}

	// Apply connection pool settings
	sqlDB, err := db.DB()
	if err != nil {
		return nil, fmt.Errorf("failed to get sql.DB: %w", err)
	}

	sqlDB.SetMaxIdleConns(cfg.MaxIdleConnections)
	sqlDB.SetMaxOpenConns(cfg.MaxOpenConnections)
	sqlDB.SetConnMaxLifetime(cfg.ConnMaxLifeTime.Duration)

	logger.Infof(ctx, "Database connection established")
	return db, nil
}
