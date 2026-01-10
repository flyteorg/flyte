package transformers

import (
	"context"
	"fmt"
	"time"

	"google.golang.org/protobuf/proto"
	"gorm.io/datatypes"

	"github.com/flyteorg/flyte/v2/flytestdlib/logger"
	"github.com/flyteorg/flyte/v2/gen/go/flyteidl2/common"
	"github.com/flyteorg/flyte/v2/gen/go/flyteidl2/workflow"
	"github.com/flyteorg/flyte/v2/runs/repository/models"
)

// CreateRunRequestToModel converts CreateRunRequest protobuf to Run domain model
func CreateRunRequestToModel(ctx context.Context, req *workflow.CreateRunRequest) (*models.Run, error) {
	// Determine run ID
	var runID *common.RunIdentifier
	switch id := req.Id.(type) {
	case *workflow.CreateRunRequest_RunId:
		runID = id.RunId
	case *workflow.CreateRunRequest_ProjectId:
		// Generate a run name
		runID = &common.RunIdentifier{
			Org:     id.ProjectId.Organization,
			Project: id.ProjectId.Name,
			Domain:  id.ProjectId.Domain,
			Name:    fmt.Sprintf("run-%d", time.Now().Unix()),
		}
		logger.Debugf(ctx, "Generated run name: %s", runID.Name)
	default:
		return nil, fmt.Errorf("invalid run ID type")
	}

	// Build ActionSpec
	actionSpec := &workflow.ActionSpec{
		ActionId: &common.ActionIdentifier{
			Run:  runID,
			Name: runID.Name,
		},
		ParentActionName: nil,
		RunSpec:          req.RunSpec,
		InputUri:         "",
		RunOutputBase:    "",
	}

	// Set the task spec
	switch taskSpec := req.Task.(type) {
	case *workflow.CreateRunRequest_TaskSpec:
		actionSpec.Spec = &workflow.ActionSpec_Task{
			Task: &workflow.TaskAction{
				Spec: taskSpec.TaskSpec,
			},
		}
	case *workflow.CreateRunRequest_TaskId:
		actionSpec.Spec = &workflow.ActionSpec_Task{
			Task: &workflow.TaskAction{
				Id: taskSpec.TaskId,
			},
		}
	}

	// Serialize ActionSpec
	actionSpecBytes, err := proto.Marshal(actionSpec)
	if err != nil {
		logger.Errorf(ctx, "Failed to marshal ActionSpec: %v", err)
		return nil, fmt.Errorf("failed to marshal action spec:  %w", err)
	}

	// Create Run model
	run := &models.Run{
		Org:              runID.Org,
		Project:          runID.Project,
		Domain:           runID.Domain,
		Name:             runID.Name,
		ParentActionName: nil,
		Phase:            "PHASE_QUEUED",
		ActionSpec:       datatypes.JSON(actionSpecBytes),
		ActionDetails:    datatypes.JSON([]byte("{}")), // Empty details initially
	}

	logger.Infof(ctx, "Created run model:  %s/%s/%s/%s", run.Org, run.Project, run.Domain, run.Name)
	return run, nil
}
