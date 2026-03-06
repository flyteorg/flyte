# Implementation Specification: Queue, Runs, and State Services

## Overview

This document specifies the implementation of three gRPC services using buf connect:
- **QueueService** - Manages execution queue for actions
- **RunService** - Manages workflow runs and their lifecycle
- **StateService** - Manages state persistence for actions

Each service will have a simple implementation backed by PostgreSQL, using pgx for database operations, and leveraging existing flytestdlib packages for database connectivity and configuration management.

## Architecture

```
flyte/
├── queue/                    # QueueService implementation
│   ├── config/
│   │   └── config.go        # Queue-specific configuration
│   ├── repository/
│   │   ├── interfaces.go    # Repository interfaces
│   │   ├── postgres.go      # PostgreSQL implementation
│   │   └── models.go        # Database models
│   ├── service/
│   │   └── queue_service.go # QueueServiceHandler implementation
│   ├── cmd/
│   │   └── main.go          # Standalone queue service binary
│   └── migrations/
│       └── *.sql            # Database migration files
│
├── runs/                     # RunService implementation
│   ├── config/
│   │   └── config.go        # Runs-specific configuration
│   ├── repository/
│   │   ├── interfaces.go    # Repository interfaces
│   │   ├── postgres.go      # PostgreSQL implementation
│   │   └── models.go        # Database models
│   ├── service/
│   │   └── run_service.go   # RunServiceHandler implementation
│   ├── cmd/
│   │   └── main.go          # Standalone runs service binary
│   └── migrations/
│       └── *.sql            # Database migration files
│
├── state/                    # StateService implementation
│   ├── config/
│   │   └── config.go        # State-specific configuration
│   ├── repository/
│   │   ├── interfaces.go    # Repository interfaces
│   │   ├── postgres.go      # PostgreSQL implementation
│   │   └── models.go        # Database models
│   ├── service/
│   │   └── state_service.go # StateServiceHandler implementation
│   ├── cmd/
│   │   └── main.go          # Standalone state service binary
│   └── migrations/
│       └── *.sql            # Database migration files
│
└── cmd/
    └── flyte-services/
        └── main.go           # Unified binary for all services + executor
```

## Technology Stack

### Core Dependencies
- **Protocol**: buf connect (using generated code from `gen/go/flyteidl2/workflow/workflowconnect/`)
- **Database**: PostgreSQL
- **Database Driver**: pgx/v5 (for raw queries)
- **ORM**: gorm (for migrations and basic operations)
- **Config Management**: `github.com/flyteorg/flyte/v2/flytestdlib/config`
- **Database Utils**: `github.com/flyteorg/flyte/v2/flytestdlib/database`
- **Logging**: `github.com/flyteorg/flyte/v2/flytestdlib/logger`
- **CLI**: `github.com/spf13/cobra`

### Service Framework
- HTTP server using `net/http`
- buf connect handlers mounted on HTTP server
- Graceful shutdown support
- Health check endpoints

## Service Specifications

### 1. QueueService

**Location**: `queue/`

**Proto Definition**: `flyteidl2/workflow/queue_service.proto`

**Connect Interface**: `workflowconnect.QueueServiceHandler`

#### RPCs to Implement:
1. `EnqueueAction(EnqueueActionRequest) -> EnqueueActionResponse`
   - Validates request
   - Persists action to queue table
   - Returns immediately (async processing)

2. `AbortQueuedRun(AbortQueuedRunRequest) -> AbortQueuedRunResponse`
   - Marks all actions in a run as aborted
   - Updates abort metadata

3. `AbortQueuedAction(AbortQueuedActionRequest) -> AbortQueuedActionResponse`
   - Marks specific action as aborted
   - Updates abort metadata

#### Database Schema:

```sql
-- queue/migrations/001_create_queue_tables.sql

CREATE TABLE IF NOT EXISTS queued_actions (
    id BIGSERIAL PRIMARY KEY,

    -- Action Identifier
    org VARCHAR(255) NOT NULL,
    project VARCHAR(255) NOT NULL,
    domain VARCHAR(255) NOT NULL,
    run_name VARCHAR(255) NOT NULL,
    action_name VARCHAR(255) NOT NULL,

    -- Parent reference
    parent_action_name VARCHAR(255),

    -- Action group
    action_group VARCHAR(255),

    -- Subject who created the action
    subject VARCHAR(255),

    -- Serialized action spec (protobuf or JSON)
    action_spec JSONB NOT NULL,

    -- Input/Output paths
    input_uri TEXT NOT NULL,
    run_output_base TEXT NOT NULL,

    -- Queue metadata
    enqueued_at TIMESTAMP NOT NULL DEFAULT NOW(),
    processed_at TIMESTAMP,

    -- Status tracking
    status VARCHAR(50) NOT NULL DEFAULT 'queued', -- queued, processing, completed, aborted, failed
    abort_reason TEXT,
    error_message TEXT,

    -- Indexing
    created_at TIMESTAMP NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP NOT NULL DEFAULT NOW(),

    -- Unique constraint on action identifier
    UNIQUE(org, project, domain, run_name, action_name)
);

CREATE INDEX idx_queued_actions_run ON queued_actions(org, project, domain, run_name);
CREATE INDEX idx_queued_actions_status ON queued_actions(status);
CREATE INDEX idx_queued_actions_enqueued_at ON queued_actions(enqueued_at) WHERE status = 'queued';
CREATE INDEX idx_queued_actions_parent ON queued_actions(parent_action_name) WHERE parent_action_name IS NOT NULL;
```

#### Configuration:

```go
// queue/config/config.go

type Config struct {
    // HTTP server configuration
    Server ServerConfig `json:"server"`

    // Database configuration (reuses flytestdlib)
    Database database.DbConfig `json:"database"`

    // Queue specific settings
    MaxQueueSize int `json:"maxQueueSize" pflag:",Maximum number of queued actions"`
    WorkerCount  int `json:"workerCount" pflag:",Number of worker goroutines for processing queue"`
}

type ServerConfig struct {
    Port int    `json:"port" pflag:",Port to bind the HTTP server"`
    Host string `json:"host" pflag:",Host to bind the HTTP server"`
}
```

---

### 2. RunService

**Location**: `runs/`

**Proto Definition**: `flyteidl2/workflow/run_service.proto`

**Connect Interface**: `workflowconnect.RunServiceHandler`

#### RPCs to Implement:
1. `CreateRun(CreateRunRequest) -> CreateRunResponse`
   - Creates a new run record
   - Initializes root action
   - Returns run metadata

2. `AbortRun(AbortRunRequest) -> AbortRunResponse`
   - Marks run as aborted
   - Cascades to all actions

3. `GetRunDetails(GetRunDetailsRequest) -> GetRunDetailsResponse`
   - Fetches complete run information
   - Includes root action details

4. `WatchRunDetails(WatchRunDetailsRequest) -> stream WatchRunDetailsResponse`
   - Streams run updates
   - Uses PostgreSQL LISTEN/NOTIFY for real-time updates

5. `GetActionDetails(GetActionDetailsRequest) -> GetActionDetailsResponse`
   - Fetches detailed action information
   - Includes all attempts

6. `WatchActionDetails(WatchActionDetailsRequest) -> stream WatchActionDetailsResponse`
   - Streams action updates

7. `GetActionData(GetActionDataRequest) -> GetActionDataResponse`
   - Returns input/output references
   - Does NOT load actual data (just URIs)

8. `ListRuns(ListRunsRequest) -> ListRunsResponse`
   - Paginated run listing
   - Supports filtering by org/project/trigger

9. `WatchRuns(WatchRunsRequest) -> stream WatchRunsResponse`
   - Streams run updates matching filter

10. `ListActions(ListActionsRequest) -> ListActionsResponse`
    - Lists actions for a run
    - Paginated

11. `WatchActions(WatchActionsRequest) -> stream WatchActionsResponse`
    - Streams action updates for a run
    - Supports filtering

12. `WatchClusterEvents(WatchClusterEventsRequest) -> stream WatchClusterEventsResponse`
    - Streams cluster events for an action attempt

13. `AbortAction(AbortActionRequest) -> AbortActionResponse`
    - Aborts a specific action

#### Database Schema:

```sql
-- runs/migrations/001_create_runs_tables.sql

CREATE TYPE phase AS ENUM (
    'PHASE_UNSPECIFIED',
    'PHASE_QUEUED',
    'PHASE_WAITING_FOR_RESOURCES',
    'PHASE_INITIALIZING',
    'PHASE_RUNNING',
    'PHASE_SUCCEEDED',
    'PHASE_FAILED',
    'PHASE_ABORTED',
    'PHASE_TIMED_OUT'
);

CREATE TYPE action_type AS ENUM (
    'ACTION_TYPE_UNSPECIFIED',
    'ACTION_TYPE_TASK',
    'ACTION_TYPE_TRACE',
    'ACTION_TYPE_CONDITION'
);

CREATE TYPE error_kind AS ENUM (
    'KIND_UNSPECIFIED',
    'KIND_USER',
    'KIND_SYSTEM'
);

CREATE TABLE IF NOT EXISTS runs (
    id BIGSERIAL PRIMARY KEY,

    -- Run Identifier
    org VARCHAR(255) NOT NULL,
    project VARCHAR(255) NOT NULL,
    domain VARCHAR(255) NOT NULL,
    name VARCHAR(255) NOT NULL,

    -- Root action reference
    root_action_name VARCHAR(255) NOT NULL,

    -- Trigger reference (if applicable)
    trigger_org VARCHAR(255),
    trigger_project VARCHAR(255),
    trigger_domain VARCHAR(255),
    trigger_name VARCHAR(255),

    -- Run spec (serialized)
    run_spec JSONB NOT NULL,

    -- Metadata
    created_by_principal VARCHAR(255),
    created_by_k8s_service_account VARCHAR(255),

    -- Timestamps
    created_at TIMESTAMP NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP NOT NULL DEFAULT NOW(),

    -- Unique constraint
    UNIQUE(org, project, domain, name)
);

CREATE TABLE IF NOT EXISTS actions (
    id BIGSERIAL PRIMARY KEY,

    -- Action Identifier
    org VARCHAR(255) NOT NULL,
    project VARCHAR(255) NOT NULL,
    domain VARCHAR(255) NOT NULL,
    run_name VARCHAR(255) NOT NULL,
    name VARCHAR(255) NOT NULL,

    -- Foreign key to run
    run_id BIGINT NOT NULL REFERENCES runs(id) ON DELETE CASCADE,

    -- Parent action (for nested actions)
    parent_action_name VARCHAR(255),

    -- Action type and metadata
    action_type action_type NOT NULL,
    action_group VARCHAR(255),

    -- Task action metadata (nullable)
    task_org VARCHAR(255),
    task_project VARCHAR(255),
    task_domain VARCHAR(255),
    task_name VARCHAR(255),
    task_version VARCHAR(255),
    task_type VARCHAR(255),
    task_short_name VARCHAR(255),

    -- Trace action metadata (nullable)
    trace_name VARCHAR(255),

    -- Condition action metadata (nullable)
    condition_name VARCHAR(255),
    condition_scope_type VARCHAR(50), -- run_id, action_id, global
    condition_scope_value TEXT,

    -- Execution metadata
    executed_by_principal VARCHAR(255),
    executed_by_k8s_service_account VARCHAR(255),

    -- Status
    phase phase NOT NULL DEFAULT 'PHASE_QUEUED',
    start_time TIMESTAMP,
    end_time TIMESTAMP,
    attempts_count INT NOT NULL DEFAULT 0,

    -- Cache status
    cache_status VARCHAR(50),

    -- Error info (if failed)
    error_kind error_kind,
    error_message TEXT,

    -- Abort info (if aborted)
    abort_reason TEXT,
    aborted_by_principal VARCHAR(255),
    aborted_by_k8s_service_account VARCHAR(255),

    -- Serialized specs
    task_spec JSONB,
    trace_spec JSONB,
    condition_spec JSONB,

    -- Input/Output references
    input_uri TEXT,
    run_output_base TEXT,

    -- Timestamps
    created_at TIMESTAMP NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP NOT NULL DEFAULT NOW(),

    -- Unique constraint
    UNIQUE(org, project, domain, run_name, name)
);

CREATE TABLE IF NOT EXISTS action_attempts (
    id BIGSERIAL PRIMARY KEY,

    -- Foreign key to action
    action_id BIGINT NOT NULL REFERENCES actions(id) ON DELETE CASCADE,

    -- Attempt number (1-indexed)
    attempt_number INT NOT NULL,

    -- Phase tracking
    phase phase NOT NULL,
    start_time TIMESTAMP NOT NULL,
    end_time TIMESTAMP,

    -- Error info (if failed)
    error_kind error_kind,
    error_message TEXT,

    -- Output references
    outputs JSONB,

    -- Logs
    log_info JSONB, -- Array of TaskLog
    logs_available BOOLEAN DEFAULT FALSE,

    -- Cache status
    cache_status VARCHAR(50),

    -- Cluster assignment
    cluster VARCHAR(255),

    -- Log context (k8s pod/container info)
    log_context JSONB,

    -- Timestamps
    created_at TIMESTAMP NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP NOT NULL DEFAULT NOW(),

    -- Unique constraint
    UNIQUE(action_id, attempt_number)
);

CREATE TABLE IF NOT EXISTS cluster_events (
    id BIGSERIAL PRIMARY KEY,

    -- Foreign key to attempt
    attempt_id BIGINT NOT NULL REFERENCES action_attempts(id) ON DELETE CASCADE,

    -- Event details
    occurred_at TIMESTAMP NOT NULL,
    message TEXT NOT NULL,

    -- Timestamps
    created_at TIMESTAMP NOT NULL DEFAULT NOW()
);

CREATE TABLE IF NOT EXISTS phase_transitions (
    id BIGSERIAL PRIMARY KEY,

    -- Foreign key to attempt
    attempt_id BIGINT NOT NULL REFERENCES action_attempts(id) ON DELETE CASCADE,

    -- Phase transition
    phase phase NOT NULL,
    start_time TIMESTAMP NOT NULL,
    end_time TIMESTAMP,

    -- Timestamps
    created_at TIMESTAMP NOT NULL DEFAULT NOW()
);

-- Indexes for performance
CREATE INDEX idx_runs_org_project ON runs(org, project);
CREATE INDEX idx_runs_trigger ON runs(trigger_org, trigger_project, trigger_domain, trigger_name) WHERE trigger_name IS NOT NULL;
CREATE INDEX idx_actions_run_id ON actions(run_id);
CREATE INDEX idx_actions_phase ON actions(phase);
CREATE INDEX idx_actions_parent ON actions(parent_action_name) WHERE parent_action_name IS NOT NULL;
CREATE INDEX idx_attempts_action_id ON action_attempts(action_id);
CREATE INDEX idx_cluster_events_attempt_id ON cluster_events(attempt_id);
CREATE INDEX idx_phase_transitions_attempt_id ON phase_transitions(attempt_id);

-- For real-time updates using LISTEN/NOTIFY
CREATE OR REPLACE FUNCTION notify_action_change() RETURNS TRIGGER AS $$
BEGIN
    PERFORM pg_notify('action_updates', NEW.id::TEXT);
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

CREATE TRIGGER action_update_trigger
AFTER INSERT OR UPDATE ON actions
FOR EACH ROW EXECUTE FUNCTION notify_action_change();

CREATE OR REPLACE FUNCTION notify_run_change() RETURNS TRIGGER AS $$
BEGIN
    PERFORM pg_notify('run_updates', NEW.id::TEXT);
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

CREATE TRIGGER run_update_trigger
AFTER INSERT OR UPDATE ON runs
FOR EACH ROW EXECUTE FUNCTION notify_run_change();
```

#### Configuration:

```go
// runs/config/config.go

type Config struct {
    // HTTP server configuration
    Server ServerConfig `json:"server"`

    // Database configuration
    Database database.DbConfig `json:"database"`

    // Watch/streaming settings
    WatchBufferSize int `json:"watchBufferSize" pflag:",Buffer size for watch streams"`
}

type ServerConfig struct {
    Port int    `json:"port" pflag:",Port to bind the HTTP server"`
    Host string `json:"host" pflag:",Host to bind the HTTP server"`
}
```

---

### 3. StateService

**Location**: `state/`

**Proto Definition**: `flyteidl2/workflow/state_service.proto`

**Connect Interface**: `workflowconnect.StateServiceHandler`

#### RPCs to Implement:
1. `Put(stream PutRequest) -> stream PutResponse`
   - Bidirectional streaming
   - Persists action state (NodeStatus JSON)
   - Returns status for each request

2. `Get(stream GetRequest) -> stream GetResponse`
   - Bidirectional streaming
   - Retrieves action state

3. `Watch(WatchRequest) -> stream WatchResponse`
   - Server streaming
   - Watches state changes for child actions
   - Uses PostgreSQL LISTEN/NOTIFY

#### Database Schema:

```sql
-- state/migrations/001_create_state_tables.sql

CREATE TABLE IF NOT EXISTS action_states (
    id BIGSERIAL PRIMARY KEY,

    -- Action Identifier
    org VARCHAR(255) NOT NULL,
    project VARCHAR(255) NOT NULL,
    domain VARCHAR(255) NOT NULL,
    run_name VARCHAR(255) NOT NULL,
    action_name VARCHAR(255) NOT NULL,

    -- Parent reference
    parent_action_name VARCHAR(255),

    -- State data (NodeStatus as JSON)
    state JSONB NOT NULL,

    -- Phase (extracted from state for indexing/filtering)
    phase VARCHAR(50),

    -- Output URI (extracted from state)
    output_uri TEXT,

    -- Error (extracted from state)
    error JSONB,

    -- Version tracking for optimistic locking
    version BIGINT NOT NULL DEFAULT 1,

    -- Timestamps
    created_at TIMESTAMP NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP NOT NULL DEFAULT NOW(),

    -- Unique constraint
    UNIQUE(org, project, domain, run_name, action_name)
);

CREATE INDEX idx_action_states_parent ON action_states(parent_action_name) WHERE parent_action_name IS NOT NULL;
CREATE INDEX idx_action_states_phase ON action_states(phase);

-- For real-time updates using LISTEN/NOTIFY
CREATE OR REPLACE FUNCTION notify_state_change() RETURNS TRIGGER AS $$
BEGIN
    PERFORM pg_notify('state_updates',
        json_build_object(
            'action_id', NEW.id,
            'parent_action_name', NEW.parent_action_name
        )::TEXT
    );
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

CREATE TRIGGER state_update_trigger
AFTER INSERT OR UPDATE ON action_states
FOR EACH ROW EXECUTE FUNCTION notify_state_change();
```

#### Configuration:

```go
// state/config/config.go

type Config struct {
    // HTTP server configuration
    Server ServerConfig `json:"server"`

    // Database configuration
    Database database.DbConfig `json:"database"`

    // State specific settings
    MaxStateSizeBytes int `json:"maxStateSizeBytes" pflag:",Maximum size of state JSON in bytes"`
}

type ServerConfig struct {
    Port int    `json:"port" pflag:",Port to bind the HTTP server"`
    Host string `json:"host" pflag:",Host to bind the HTTP server"`
}
```

---

## Unified Binary

**Location**: `cmd/flyte-services/main.go`

The unified binary provides a single entrypoint that can run:
1. **queue** - QueueService only
2. **runs** - RunService only
3. **state** - StateService only
4. **executor** - Kubernetes controller only
5. **all** - All services together on different ports

### Command Structure:

```bash
# Run queue service only
flyte-services queue --config config.yaml

# Run runs service only
flyte-services runs --config config.yaml

# Run state service only
flyte-services state --config config.yaml

# Run executor only
flyte-services executor --config config.yaml

# Run all services
flyte-services all --config config.yaml
```

### Implementation:

```go
// cmd/flyte-services/main.go

package main

import (
    "context"
    "os"
    "os/signal"
    "syscall"

    "github.com/spf13/cobra"
    "github.com/flyteorg/flyte/v2/flytestdlib/config"
    "github.com/flyteorg/flyte/v2/flytestdlib/logger"

    queuecmd "github.com/flyteorg/flyte/v2/queue/cmd"
    runscmd "github.com/flyteorg/flyte/v2/runs/cmd"
    statecmd "github.com/flyteorg/flyte/v2/state/cmd"
    executorcmd "github.com/flyteorg/flyte/v2/executor/cmd"
)

var rootCmd = &cobra.Command{
    Use:   "flyte-services",
    Short: "Flyte services unified binary",
}

func init() {
    rootCmd.AddCommand(queuecmd.NewQueueCommand())
    rootCmd.AddCommand(runscmd.NewRunsCommand())
    rootCmd.AddCommand(statecmd.NewStateCommand())
    rootCmd.AddCommand(executorcmd.NewExecutorCommand())
    rootCmd.AddCommand(newAllCommand())
}

func main() {
    if err := rootCmd.Execute(); err != nil {
        os.Exit(1)
    }
}

func newAllCommand() *cobra.Command {
    return &cobra.Command{
        Use:   "all",
        Short: "Run all services (queue, runs, state, executor)",
        RunE: func(cmd *cobra.Command, args []string) error {
            ctx, cancel := context.WithCancel(context.Background())
            defer cancel()

            // Setup signal handling
            sigCh := make(chan os.Signal, 1)
            signal.Notify(sigCh, syscall.SIGINT, syscall.SIGTERM)

            // Start all services in goroutines
            errCh := make(chan error, 4)

            go func() {
                errCh <- queuecmd.RunQueue(ctx)
            }()

            go func() {
                errCh <- runscmd.RunRuns(ctx)
            }()

            go func() {
                errCh <- statecmd.RunState(ctx)
            }()

            go func() {
                errCh <- executorcmd.RunExecutor(ctx)
            }()

            // Wait for signal or error
            select {
            case sig := <-sigCh:
                logger.Infof(ctx, "Received signal %v, shutting down...", sig)
                cancel()
            case err := <-errCh:
                logger.Errorf(ctx, "Service error: %v", err)
                cancel()
                return err
            }

            return nil
        },
    }
}
```

---

## Database Connection Management

All services use flytestdlib for database management:

```go
// Example from queue/service/queue_service.go

import (
    "github.com/flyteorg/flyte/v2/flytestdlib/database"
    "gorm.io/gorm"
)

func initDB(ctx context.Context, cfg *database.DbConfig) (*gorm.DB, error) {
    gormConfig := &gorm.Config{
        // Configuration options
    }

    // Create database if it doesn't exist
    db, err := database.CreatePostgresDbIfNotExists(ctx, gormConfig, cfg.Postgres)
    if err != nil {
        return nil, err
    }

    // Apply connection pool settings
    sqlDB, err := db.DB()
    if err != nil {
        return nil, err
    }

    sqlDB.SetMaxIdleConns(cfg.MaxIdleConnections)
    sqlDB.SetMaxOpenConns(cfg.MaxOpenConnections)
    sqlDB.SetConnMaxLifetime(cfg.ConnMaxLifeTime.Duration)

    return db, nil
}
```

For pgx-specific operations (like LISTEN/NOTIFY for streaming):

```go
import (
    "github.com/jackc/pgx/v5/pgxpool"
)

func initPgxPool(ctx context.Context, cfg *database.PostgresConfig) (*pgxpool.Pool, error) {
    connString := fmt.Sprintf(
        "postgres://%s:%s@%s:%d/%s?%s",
        cfg.User,
        resolvePassword(ctx, cfg.Password, cfg.PasswordPath),
        cfg.Host,
        cfg.Port,
        cfg.DbName,
        cfg.ExtraOptions,
    )

    return pgxpool.New(ctx, connString)
}
```

---

## Configuration Management

All services use flytestdlib config:

```go
// Example from queue/cmd/main.go

import (
    "github.com/flyteorg/flyte/v2/flytestdlib/config"
    queueconfig "github.com/flyteorg/flyte/v2/queue/config"
)

var (
    configSection = config.MustRegisterSection("queue", &queueconfig.Config{})
)

func main() {
    // Parse config from file and flags
    if err := config.LoadConfig(...); err != nil {
        panic(err)
    }

    cfg := configSection.GetConfig().(*queueconfig.Config)
    // Use cfg...
}
```

---

## Service Implementation Pattern

Each service follows this pattern:

```go
// queue/service/queue_service.go

package service

import (
    "context"

    "connectrpc.com/connect"
    "github.com/flyteorg/flyte/v2/gen/go/flyteidl2/workflow"
    "github.com/flyteorg/flyte/v2/gen/go/flyteidl2/workflow/workflowconnect"
    "github.com/flyteorg/flyte/v2/queue/repository"
)

type QueueService struct {
    repo repository.Repository
}

func NewQueueService(repo repository.Repository) *QueueService {
    return &QueueService{repo: repo}
}

// Ensure we implement the interface
var _ workflowconnect.QueueServiceHandler = (*QueueService)(nil)

func (s *QueueService) EnqueueAction(
    ctx context.Context,
    req *connect.Request[workflow.EnqueueActionRequest],
) (*connect.Response[workflow.EnqueueActionResponse], error) {
    // Validate request
    if err := req.Msg.Validate(); err != nil {
        return nil, connect.NewError(connect.CodeInvalidArgument, err)
    }

    // Persist to database
    if err := s.repo.EnqueueAction(ctx, req.Msg); err != nil {
        return nil, connect.NewError(connect.CodeInternal, err)
    }

    return connect.NewResponse(&workflow.EnqueueActionResponse{}), nil
}

// ... other methods
```

---

## HTTP Server Setup

Each service's main.go sets up an HTTP server:

```go
// queue/cmd/main.go

package cmd

import (
    "context"
    "fmt"
    "net/http"

    "golang.org/x/net/http2"
    "golang.org/x/net/http2/h2c"

    "github.com/flyteorg/flyte/v2/gen/go/flyteidl2/workflow/workflowconnect"
    "github.com/flyteorg/flyte/v2/queue/service"
    "github.com/flyteorg/flyte/v2/queue/repository"
)

func RunQueue(ctx context.Context) error {
    // Initialize database
    db, err := initDB(ctx)
    if err != nil {
        return err
    }

    // Run migrations
    if err := runMigrations(db); err != nil {
        return err
    }

    // Create repository
    repo := repository.NewPostgresRepository(db)

    // Create service
    svc := service.NewQueueService(repo)

    // Create HTTP handler
    mux := http.NewServeMux()

    path, handler := workflowconnect.NewQueueServiceHandler(svc)
    mux.Handle(path, handler)

    // Add health check
    mux.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
        w.WriteHeader(http.StatusOK)
        w.Write([]byte("OK"))
    })

    // Setup HTTP/2 support
    addr := fmt.Sprintf("%s:%d", cfg.Server.Host, cfg.Server.Port)
    server := &http.Server{
        Addr:    addr,
        Handler: h2c.NewHandler(mux, &http2.Server{}),
    }

    logger.Infof(ctx, "Starting Queue Service on %s", addr)
    return server.ListenAndServe()
}
```

---

## Migration Management

Each service uses golang-migrate or similar:

```go
// queue/repository/migrations.go

func RunMigrations(db *gorm.DB) error {
    return db.AutoMigrate(
        &models.QueuedAction{},
    )
}
```

Or use raw SQL migrations with a migration tool.

---

## Repository Pattern

Each service implements a repository interface:

```go
// queue/repository/interfaces.go

type Repository interface {
    EnqueueAction(ctx context.Context, req *workflow.EnqueueActionRequest) error
    AbortQueuedRun(ctx context.Context, runID *common.RunIdentifier, reason string) error
    AbortQueuedAction(ctx context.Context, actionID *common.ActionIdentifier, reason string) error
    GetQueuedActions(ctx context.Context, limit int) ([]*models.QueuedAction, error)
}
```

```go
// queue/repository/postgres.go

type PostgresRepository struct {
    db *gorm.DB
}

func NewPostgresRepository(db *gorm.DB) Repository {
    return &PostgresRepository{db: db}
}

func (r *PostgresRepository) EnqueueAction(ctx context.Context, req *workflow.EnqueueActionRequest) error {
    action := &models.QueuedAction{
        Org:            req.ActionId.RunId.Org,
        Project:        req.ActionId.RunId.Project,
        Domain:         req.ActionId.RunId.Domain,
        RunName:        req.ActionId.RunId.Name,
        ActionName:     req.ActionId.Name,
        ParentActionName: req.ParentActionName,
        ActionGroup:    req.Group,
        Subject:        req.Subject,
        ActionSpec:     req, // Will be marshaled to JSONB
        InputUri:       req.InputUri,
        RunOutputBase:  req.RunOutputBase,
        Status:         "queued",
    }

    return r.db.WithContext(ctx).Create(action).Error
}
```

---

## Streaming Implementation (Watch/Listen)

For streaming RPCs, use PostgreSQL LISTEN/NOTIFY:

```go
// runs/service/run_service.go

func (s *RunService) WatchRunDetails(
    ctx context.Context,
    req *connect.Request[workflow.WatchRunDetailsRequest],
    stream *connect.ServerStream[workflow.WatchRunDetailsResponse],
) error {
    // Get initial state
    details, err := s.repo.GetRunDetails(ctx, req.Msg.RunId)
    if err != nil {
        return err
    }

    if err := stream.Send(&workflow.WatchRunDetailsResponse{Details: details}); err != nil {
        return err
    }

    // Subscribe to updates via PostgreSQL LISTEN
    updates := make(chan *workflow.RunDetails)
    errs := make(chan error)

    go s.repo.WatchRunDetails(ctx, req.Msg.RunId, updates, errs)

    for {
        select {
        case <-ctx.Done():
            return nil
        case err := <-errs:
            return err
        case details := <-updates:
            if err := stream.Send(&workflow.WatchRunDetailsResponse{Details: details}); err != nil {
                return err
            }
        }
    }
}
```

```go
// runs/repository/postgres.go

func (r *PostgresRepository) WatchRunDetails(
    ctx context.Context,
    runID *common.RunIdentifier,
    updates chan<- *workflow.RunDetails,
    errs chan<- error,
) {
    conn, err := r.pgxPool.Acquire(ctx)
    if err != nil {
        errs <- err
        return
    }
    defer conn.Release()

    // Listen for notifications
    _, err = conn.Exec(ctx, "LISTEN run_updates")
    if err != nil {
        errs <- err
        return
    }

    for {
        notification, err := conn.Conn().WaitForNotification(ctx)
        if err != nil {
            errs <- err
            return
        }

        // Fetch updated run details
        details, err := r.GetRunDetails(ctx, runID)
        if err != nil {
            errs <- err
            return
        }

        select {
        case updates <- details:
        case <-ctx.Done():
            return
        }
    }
}
```

---

## Testing Strategy

### Unit Tests
- Repository layer: Mock database using testcontainers with PostgreSQL
- Service layer: Mock repository interface
- Use table-driven tests

### Integration Tests
- End-to-end tests with real PostgreSQL
- Use docker-compose for local testing
- Test streaming with multiple concurrent clients

### Example:

```go
// queue/service/queue_service_test.go

func TestEnqueueAction(t *testing.T) {
    mockRepo := &mocks.Repository{}
    svc := service.NewQueueService(mockRepo)

    req := connect.NewRequest(&workflow.EnqueueActionRequest{
        // ... populate request
    })

    mockRepo.On("EnqueueAction", mock.Anything, req.Msg).Return(nil)

    resp, err := svc.EnqueueAction(context.Background(), req)
    assert.NoError(t, err)
    assert.NotNil(t, resp)
}
```

---

## Deployment Considerations

### Configuration Files

Example `config.yaml`:

```yaml
database:
  postgres:
    host: postgres.flyte.svc.cluster.local
    port: 5432
    dbname: flyte_queue
    username: flyte
    passwordPath: /etc/secrets/db-password
    extraOptions: "sslmode=require"
  maxIdleConnections: 10
  maxOpenConnections: 100
  connMaxLifeTime: 1h

queue:
  server:
    host: 0.0.0.0
    port: 8089
  maxQueueSize: 10000
  workerCount: 10

runs:
  server:
    host: 0.0.0.0
    port: 8090
  watchBufferSize: 100

state:
  server:
    host: 0.0.0.0
    port: 8091
  maxStateSizeBytes: 1048576  # 1MB
```

### Docker Compose (for local development)

```yaml
version: '3.8'
services:
  postgres:
    image: postgres:15
    environment:
      POSTGRES_USER: flyte
      POSTGRES_PASSWORD: flyte
      POSTGRES_DB: flyte
    ports:
      - "5432:5432"

  queue:
    build: .
    command: queue --config /etc/flyte/config.yaml
    ports:
      - "8089:8089"
    depends_on:
      - postgres

  runs:
    build: .
    command: runs --config /etc/flyte/config.yaml
    ports:
      - "8090:8090"
    depends_on:
      - postgres

  state:
    build: .
    command: state --config /etc/flyte/config.yaml
    ports:
      - "8091:8091"
    depends_on:
      - postgres
```

---

## Implementation Phases

### Phase 1: Core Infrastructure
1. Setup project structure
2. Implement database schemas and migrations
3. Implement repository interfaces and PostgreSQL implementations
4. Setup configuration management using flytestdlib

### Phase 2: Service Implementation
1. Implement QueueService
2. Implement RunService (non-streaming RPCs first)
3. Implement StateService (non-streaming RPCs first)

### Phase 3: Streaming Support
1. Add PostgreSQL LISTEN/NOTIFY support
2. Implement streaming RPCs (Watch*)
3. Test concurrent streaming clients

### Phase 4: Integration
1. Implement unified binary command structure
2. Add health checks and metrics
3. Integration testing
4. Documentation

### Phase 5: Production Readiness
1. Add observability (metrics, tracing, logging)
2. Performance testing and optimization
3. Security audit
4. Deployment documentation

---

## Open Questions

1. **Migration Strategy**: Should we use golang-migrate, gorm AutoMigrate, or custom SQL scripts?
2. **Protobuf Serialization**: Store protobuf as JSONB or use binary serialization?
3. **Queue Processing**: Should QueueService also include worker implementation for processing queued actions?
4. **Multi-tenancy**: How to handle org/project isolation at the database level?
5. **Metrics**: What metrics should each service expose?
6. **Rate Limiting**: Should services implement rate limiting per org/project?

---

## References

- Protocol Buffers: [queue_service.proto](flyteidl2/workflow/queue_service.proto), [run_service.proto](flyteidl2/workflow/run_service.proto), [state_service.proto](flyteidl2/workflow/state_service.proto)
- Generated Code: `gen/go/flyteidl2/workflow/workflowconnect/`
- Database Utils: `flytestdlib/database/`
- Config Management: `flytestdlib/config/`
- Buf Connect: https://connectrpc.com/docs/go/getting-started
- PostgreSQL LISTEN/NOTIFY: https://www.postgresql.org/docs/current/sql-notify.html
