package api

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"testing"

	"time"

	embeddedpostgres "github.com/fergusstrange/embedded-postgres"
	"github.com/jmoiron/sqlx"
	"golang.org/x/net/http2"
	"golang.org/x/net/http2/h2c"

	"github.com/flyteorg/flyte/v2/flytestdlib/database"
	"github.com/flyteorg/flyte/v2/gen/go/flyteidl2/actions/actionsconnect"
	"github.com/flyteorg/flyte/v2/gen/go/flyteidl2/task/taskconnect"
	"github.com/flyteorg/flyte/v2/gen/go/flyteidl2/workflow/workflowconnect"
	"github.com/flyteorg/flyte/v2/runs/migrations"
	"github.com/flyteorg/flyte/v2/runs/repository"
	"github.com/flyteorg/flyte/v2/runs/service"
)

const (
	testPort = 8091 // Different port to avoid conflicts with main service
)

var (
	endpoint   string
	testServer *http.Server
	testDB     *sqlx.DB // Expose DB for cleanup
)

// TestMain sets up the test environment with PostgreSQL database and runs service
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

		os.Exit(exitCode)
	}()

	// Setup: Start embedded PostgreSQL
	const embeddedPGPort = 15435
	pg := embeddedpostgres.NewDatabase(
		embeddedpostgres.DefaultConfig().
			Port(embeddedPGPort).
			Database("flyte_runs_test").
			Username("postgres").
			Password("postgres").
			RuntimePath(fmt.Sprintf("/tmp/embedded-postgres-%d", embeddedPGPort)),
	)
	if err := pg.Start(); err != nil {
		log.Printf("Failed to start embedded postgres: %v", err)
		exitCode = 1
		return
	}
	defer func() {
		if err := pg.Stop(); err != nil {
			log.Printf("Warning: failed to stop embedded postgres: %v", err)
		}
	}()

	dbConfig := &database.DbConfig{
		Postgres: database.PostgresConfig{
			Host:         "localhost",
			Port:         embeddedPGPort,
			DbName:       "flyte_runs_test",
			User:         "postgres",
			Password:     "postgres",
			ExtraOptions: "sslmode=disable",
		},
		MaxIdleConnections: 10,
		MaxOpenConnections: 100,
	}
	var err error
	testDB, err = database.GetDB(ctx, dbConfig)
	if err != nil {
		log.Printf("Failed to initialize database: %v", err)
		exitCode = 1
		return
	}
	log.Println("Database initialized")

	// Run migrations
	if err := migrations.RunMigrations(ctx, testDB); err != nil {
		log.Printf("Failed to run migrations: %v", err)
		exitCode = 1
		return
	}
	log.Println("Database migrations completed")

	// Create repository and services
	repo, err := repository.NewRepository(testDB, *dbConfig)
	if err != nil {
		log.Printf("Failed to create repository: %v", err)
		exitCode = 1
		return
	}
	taskSvc := service.NewTaskService(repo, nil)

	// Create RunService with a no-op actions client (points at test server; not used by watch tests)
	endpointURL := fmt.Sprintf("http://localhost:%d", testPort)
	actionsClient := actionsconnect.NewActionsServiceClient(http.DefaultClient, endpointURL)
	runSvc := service.NewRunService(repo, actionsClient, nil, nil, "", nil, nil)

	// Setup HTTP server
	mux := http.NewServeMux()
	taskPath, taskHandler := taskconnect.NewTaskServiceHandler(taskSvc)
	mux.Handle(taskPath, taskHandler)

	runPath, runHandler := workflowconnect.NewRunServiceHandler(runSvc)
	mux.Handle(runPath, runHandler)

	internalRunPath, internalRunHandler := workflowconnect.NewInternalRunServiceHandler(runSvc)
	mux.Handle(internalRunPath, internalRunHandler)

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

	// Truncate known tables
	tables := []string{
		"action_events", "actions", "runs", "tasks", "projects",
	}
	for _, table := range tables {
		if _, err := testDB.Exec(fmt.Sprintf("DELETE FROM %s", table)); err != nil {
			t.Logf("Warning: Failed to cleanup table %s: %v", table, err)
		}
	}

	t.Log("Test database cleaned up")
}
