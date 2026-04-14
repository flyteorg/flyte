package executor

import (
	"context"
	"fmt"
	"hash/fnv"
	"net/http"
	"time"

	"connectrpc.com/connect"
	"golang.org/x/time/rate"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/types/known/timestamppb"

	"github.com/flyteorg/flyte/v2/flytestdlib/logger"
	"github.com/flyteorg/flyte/v2/gen/go/flyteidl2/common"
	corepb "github.com/flyteorg/flyte/v2/gen/go/flyteidl2/core"
	taskpb "github.com/flyteorg/flyte/v2/gen/go/flyteidl2/task"
	"github.com/flyteorg/flyte/v2/gen/go/flyteidl2/workflow"
	"github.com/flyteorg/flyte/v2/gen/go/flyteidl2/workflow/workflowconnect"
	"github.com/flyteorg/flyte/v2/runs/repository/models"
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

	runName := runName(t, scheduledAt)

	createReq := &workflow.CreateRunRequest{
		Id: &workflow.CreateRunRequest_RunId{
			RunId: &common.RunIdentifier{
				Project: t.Project,
				Domain:  t.Domain,
				Name:    runName,
			},
		},
		Task: &workflow.CreateRunRequest_TriggerName{
			TriggerName: &common.TriggerName{
				Project:  t.Project,
				Domain:   t.Domain,
				Name:     t.Name,
				TaskName: t.TaskName,
			},
		},
		Source: workflow.RunSource_RUN_SOURCE_SCHEDULE_TRIGGER,
	}

	// Inject the scheduled time as the kickoff input arg if the trigger spec defines one.
	if argName := kickoffTimeInputArg(t); argName != "" {
		createReq.InputWrapper = &workflow.CreateRunRequest_Inputs{
			Inputs: appendKickoffTimeInput(nil, argName, scheduledAt),
		}
	}

	req := connect.NewRequest(createReq)

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

// kickoffTimeInputArg extracts the KickoffTimeInputArg from the trigger's AutomationSpec.
// Returns "" if not set or the spec cannot be parsed.
func kickoffTimeInputArg(t *models.Trigger) string {
	if len(t.AutomationSpec) == 0 {
		return ""
	}
	spec := &taskpb.TriggerAutomationSpec{}
	if err := proto.Unmarshal(t.AutomationSpec, spec); err != nil {
		return ""
	}
	return spec.GetSchedule().GetKickoffTimeInputArg()
}

// runName returns a deterministic run name for a scheduled trigger execution.
func runName(t *models.Trigger, scheduledAt time.Time) string {
	h := fnv.New64()
	_, _ = fmt.Fprintf(h, "%s:%s:%s:%s:%d:%d:%d:%d:%d:%d",
		t.Project, t.Domain, t.TaskName, t.Name,
		scheduledAt.Year(), scheduledAt.Month(), scheduledAt.Day(),
		scheduledAt.Hour(), scheduledAt.Minute(), scheduledAt.Second())
	return fmt.Sprintf("r%x", h.Sum64())
}

// appendKickoffTimeInput adds a datetime literal for the kickoff time to the inputs.
func appendKickoffTimeInput(inputs *taskpb.Inputs, argName string, scheduledAt time.Time) *taskpb.Inputs {
	if inputs == nil {
		inputs = &taskpb.Inputs{}
	}
	inputs.Literals = append(inputs.Literals, &taskpb.NamedLiteral{
		Name: argName,
		Value: &corepb.Literal{
			Value: &corepb.Literal_Scalar{
				Scalar: &corepb.Scalar{
					Value: &corepb.Scalar_Primitive{
						Primitive: &corepb.Primitive{
							Value: &corepb.Primitive_Datetime{
								Datetime: timestamppb.New(scheduledAt),
							},
						},
					},
				},
			},
		},
	})
	return inputs
}
