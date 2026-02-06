package transformers

import (
	"context"
	"fmt"
	"time"

	"google.golang.org/protobuf/encoding/protojson"
	"google.golang.org/protobuf/types/known/timestamppb"
	"gorm.io/datatypes"

	"github.com/flyteorg/flyte/v2/flytestdlib/logger"
	"github.com/flyteorg/flyte/v2/gen/go/flyteidl2/common"
	"github.com/flyteorg/flyte/v2/gen/go/flyteidl2/core"
	"github.com/flyteorg/flyte/v2/gen/go/flyteidl2/workflow"
	"github.com/flyteorg/flyte/v2/runs/repository/models"
)

const InitialPhase = "PHASE_QUEUED"

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
	actionSpecBytes, err := protojson.Marshal(actionSpec)
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
		Phase:            InitialPhase,
		ActionSpec:       datatypes.JSON(actionSpecBytes),
		ActionDetails:    datatypes.JSON([]byte("{}")), // Empty details initially
	}

	logger.Infof(ctx, "Created run model:  %s/%s/%s/%s", run.Org, run.Project, run.Domain, run.Name)
	return run, nil
}

// RunModelToCreateRunResponse converts a domain model Run to a CreateRunResponse
func RunModelToCreateRunResponse(run *models.Run, source workflow.RunSource) *workflow.CreateRunResponse {
	if run == nil {
		return nil
	}

	// Build the action identifier
	actionID := &common.ActionIdentifier{
		Run: &common.RunIdentifier{
			Org:     run.Org,
			Project: run.Project,
			Domain:  run.Domain,
			Name:    run.Name,
		},
		Name: run.Name, // For root actions, action name = run name
	}

	// Build action status
	actionStatus := &workflow.ActionStatus{
		Phase:       DBPhaseToProtobufPhase(run.Phase),
		StartTime:   timestamppb.New(run.CreatedAt),
		Attempts:    0,
		CacheStatus: core.CatalogCacheStatus_CACHE_DISABLED,
	}

	// Build action metadata
	actionMetadata := &workflow.ActionMetadata{
		Source:     source, // ← Use the passed-in source
		Parent:     "",
		ActionType: workflow.ActionType_ACTION_TYPE_TASK,
	}

	// Build the complete response
	return &workflow.CreateRunResponse{
		Run: &workflow.Run{
			Action: &workflow.Action{
				Id:       actionID,
				Status:   actionStatus,
				Metadata: actionMetadata,
			},
		},
	}
}

func DBPhaseToProtobufPhase(dbPhase string) common.ActionPhase {
	protoPhaseStr := "ACTION_" + dbPhase // "PHASE_QUEUED" → "ACTION_PHASE_QUEUED"

	if val, ok := common.ActionPhase_value[protoPhaseStr]; ok {
		return common.ActionPhase(val)
	}

	return common.ActionPhase_ACTION_PHASE_UNSPECIFIED
}
