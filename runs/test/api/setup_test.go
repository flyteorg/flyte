package api

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"reflect"
	"testing"
	"time"

	"golang.org/x/net/http2"
	"golang.org/x/net/http2/h2c"
	"gorm.io/gorm"

	"github.com/flyteorg/flyte/v2/flytestdlib/database"
	"github.com/flyteorg/flyte/v2/flytestdlib/logger"
	"github.com/flyteorg/flyte/v2/gen/go/flyteidl2/task/taskconnect"
	"github.com/flyteorg/flyte/v2/runs/migrations"
	"github.com/flyteorg/flyte/v2/runs/repository"
	"github.com/flyteorg/flyte/v2/runs/service"
)

const (
	testPort   = 8091                      // Different port to avoid conflicts with main service
	testDBFile = "/tmp/flyte-runs-test.db" // Temporary database file for tests
)

var (
	endpoint   string
	testServer *http.Server
	testDB     *gorm.DB // Expose DB for cleanup
)

// TestMain sets up the test environment with SQLite database and runs service
func TestMain(m *testing.M) {
	ctx := context.Background()
	var exitCode int

	defer func() {
		// Stop server if it was started
		if testServer != nil {
			shutdownCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
			defer cancel()
			if err := testServer.Shutdown(shutdownCtx); err != nil {
				log.Printf("Test server shutdown error: %v", err)
			}
			log.Println("Test server stopped")
		}

		// Remove test database file
		if err := os.Remove(testDBFile); err != nil && !os.IsNotExist(err) {
			log.Printf("Warning: Failed to remove test database: %v", err)
		}

		os.Exit(exitCode)
	}()

	// Setup: Create SQLite database
	// NOTE: Using a temp file instead of :memory: to avoid GORM AutoMigrate issues
	// with index creation on fresh in-memory databases
	dbConfig := &database.DbConfig{
		SQLite: database.SQLiteConfig{
			File: testDBFile,
		},
	}
	logCfg := logger.GetConfig()
	var err error
	testDB, err = database.GetDB(ctx, dbConfig, logCfg)
	if err != nil {
		log.Printf("Failed to initialize database: %v", err)
		exitCode = 1
		return
	}
	log.Println("Database initialized")

	// Run migrations
	if err := migrations.RunMigrations(testDB); err != nil {
		log.Printf("Failed to run migrations: %v", err)
		exitCode = 1
		return
	}
	log.Println("Database migrations completed")

	// Create repository and services
	repo := repository.NewRepository(testDB)
	taskSvc := service.NewTaskService(repo)

	// Setup HTTP server
	mux := http.NewServeMux()
	taskPath, taskHandler := taskconnect.NewTaskServiceHandler(taskSvc)
	mux.Handle(taskPath, taskHandler)

	// Add health check
	mux.HandleFunc("/healthz", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte("OK"))
	})

	endpoint = fmt.Sprintf("http://localhost:%d", testPort)
	testServer = &http.Server{
		Addr:    fmt.Sprintf(":%d", testPort),
		Handler: h2c.NewHandler(mux, &http2.Server{}),
	}

	// Start server in background
	errChan := make(chan error, 1)
	go func() {
		log.Printf("Test server starting on %s", endpoint)
		if err := testServer.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			errChan <- err
		}
	}()

	// Wait for either server readiness or startup error
	readyChan := make(chan bool, 1)
	go func() {
		readyChan <- waitForServer(endpoint, 10*time.Second)
	}()

	select {
	case err := <-errChan:
		log.Printf("Test server failed to start: %v", err)
		exitCode = 1
		return
	case ready := <-readyChan:
		if !ready {
			log.Printf("Test server failed to start (health check timeout)")
			exitCode = 1
			return
		}
	}
	log.Println("Test server is ready")

	// Run tests
	exitCode = m.Run()
}

// waitForServer waits for the server to be ready
func waitForServer(url string, timeout time.Duration) bool {
	client := &http.Client{Timeout: 1 * time.Second}
	deadline := time.Now().Add(timeout)

	for time.Now().Before(deadline) {
		resp, err := client.Get(url + "/healthz")
		if err == nil && resp.StatusCode == http.StatusOK {
			resp.Body.Close()
			return true
		}
		if resp != nil {
			resp.Body.Close()
		}
		time.Sleep(100 * time.Millisecond)
	}
	return false
}

// cleanupTestDB clears all tables in the test database
// This ensures each test starts with a clean state
func cleanupTestDB(t *testing.T) {
	t.Helper()

	if testDB == nil {
		t.Log("Warning: testDB is nil, skipping cleanup")
		return
	}

	// Loop through all models defined in migrations
	for _, model := range migrations.AllModels {
		tableName := testDB.NamingStrategy.TableName(reflect.TypeOf(model).Elem().Name())

		if err := testDB.Exec(fmt.Sprintf("DELETE FROM %s", tableName)).Error; err != nil {
			t.Logf("Warning: Failed to cleanup table %s: %v", tableName, err)
		}
	}

	t.Log("Test database cleaned up")
}
