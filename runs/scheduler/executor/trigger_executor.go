package executor

import (
	"context"
	"net/http"
	"time"

	"connectrpc.com/connect"
	"golang.org/x/time/rate"

	"github.com/flyteorg/flyte/v2/flytestdlib/logger"
	"github.com/flyteorg/flyte/v2/gen/go/flyteidl2/common"
	"github.com/flyteorg/flyte/v2/gen/go/flyteidl2/workflow"
	"github.com/flyteorg/flyte/v2/gen/go/flyteidl2/workflow/workflowconnect"
	"github.com/flyteorg/flyte/v2/runs/repository/models"
	"github.com/flyteorg/flyte/v2/runs/scheduler/core"
)

// TriggerExecutor fires a CreateRun call for a scheduled trigger execution.
type TriggerExecutor struct {
	runClient workflowconnect.RunServiceClient
	limiter   *rate.Limiter
}

// TriggerExecutorConfig holds tunables for the executor.
type TriggerExecutorConfig struct {
	// BaseURL is the URL of the runs service (e.g. "http://localhost:8090").
	BaseURL string
	// QPS is the token-bucket rate for CreateRun calls (tokens/second).
	QPS float64
	// Burst is the token-bucket burst size.
	Burst int
}

// NewTriggerExecutor constructs a TriggerExecutor.
func NewTriggerExecutor(cfg TriggerExecutorConfig) *TriggerExecutor {
	return &TriggerExecutor{
		runClient: workflowconnect.NewRunServiceClient(http.DefaultClient, cfg.BaseURL),
		limiter:   rate.NewLimiter(rate.Limit(cfg.QPS), cfg.Burst),
	}
}

// Execute fires a single CreateRun for the trigger at the given scheduled time.
// It uses a deterministic run name so that duplicate executions are idempotent.
func (e *TriggerExecutor) Execute(ctx context.Context, t *models.Trigger, scheduledAt time.Time) error {
	if err := e.limiter.Wait(ctx); err != nil {
		return err
	}

	runName := core.NameHash(t.Project, t.Domain, t.TaskName, t.Name, scheduledAt)

	req := connect.NewRequest(&workflow.CreateRunRequest{
		Id: &workflow.CreateRunRequest_RunId{
			RunId: &common.RunIdentifier{
				Project: t.Project,
				Domain:  t.Domain,
				Name:    runName,
			},
		},
		Task: &workflow.CreateRunRequest_TriggerName{
			TriggerName: &common.TriggerName{
				Project: t.Project,
				Domain:  t.Domain,
				Name:    t.Name,
			},
		},
		Source: workflow.RunSource_RUN_SOURCE_SCHEDULE_TRIGGER,
	})

	_, err := e.runClient.CreateRun(ctx, req)
	if err != nil {
		// AlreadyExists means the run was already fired — treat as success.
		if connect.CodeOf(err) == connect.CodeAlreadyExists {
			logger.Infof(ctx, "scheduler: run %s already exists, skipping", runName)
			return nil
		}
		return err
	}

	logger.Infof(ctx, "scheduler: fired run %s for trigger %s/%s/%s at %s",
		runName, t.Project, t.Domain, t.Name, scheduledAt.Format(time.RFC3339))
	return nil
}
