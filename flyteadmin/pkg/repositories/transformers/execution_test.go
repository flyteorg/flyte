package transformers

import (
	"fmt"
	"testing"
	"time"

	"github.com/flyteorg/flyteidl/gen/pb-go/flyteidl/event"
	"github.com/golang/protobuf/ptypes"

	"github.com/flyteorg/flyteidl/gen/pb-go/flyteidl/admin"
	"github.com/flyteorg/flyteidl/gen/pb-go/flyteidl/core"
	"github.com/golang/protobuf/proto"
	"github.com/stretchr/testify/assert"

	"github.com/flyteorg/flyteadmin/pkg/manager/impl/testutils"
	"github.com/flyteorg/flyteadmin/pkg/repositories/models"
)

func getRunningExecutionModel(specBytes []byte, existingClosureBytes []byte, startedAt time.Time) models.Execution {
	executionModel := models.Execution{
		ExecutionKey: models.ExecutionKey{
			Project: "project",
			Domain:  "domain",
			Name:    "name",
		},
		Spec:               specBytes,
		Phase:              core.WorkflowExecution_RUNNING.String(),
		Closure:            existingClosureBytes,
		LaunchPlanID:       uint(1),
		WorkflowID:         uint(2),
		StartedAt:          &startedAt,
		ExecutionCreatedAt: &startedAt,
		ExecutionUpdatedAt: &startedAt,
	}
	return executionModel
}

func TestCreateExecutionModel(t *testing.T) {
	execRequest := testutils.GetExecutionRequest()
	execRequest.Spec.Metadata = &admin.ExecutionMetadata{
		Mode: admin.ExecutionMetadata_SYSTEM,
	}
	lpID := uint(33)
	wfID := uint(23)
	nodeID := uint(11)
	sourceID := uint(12)
	createdAt := time.Now()
	workflowIdentifier := &core.Identifier{
		Project: "project",
		Domain:  "domain",
		Name:    "workflow name",
		Version: "version",
	}

	principal := "principal"
	cluster := "cluster"
	execution, err := CreateExecutionModel(CreateExecutionModelInput{
		WorkflowExecutionID: core.WorkflowExecutionIdentifier{
			Project: "project",
			Domain:  "domain",
			Name:    "name",
		},
		RequestSpec:           execRequest.Spec,
		LaunchPlanID:          lpID,
		WorkflowID:            wfID,
		Phase:                 core.WorkflowExecution_RUNNING,
		CreatedAt:             createdAt,
		WorkflowIdentifier:    workflowIdentifier,
		ParentNodeExecutionID: nodeID,
		SourceExecutionID:     sourceID,
		Principal:             principal,
		Cluster:               cluster,
	})
	assert.NoError(t, err)
	assert.Equal(t, "project", execution.Project)
	assert.Equal(t, "domain", execution.Domain)
	assert.Equal(t, "name", execution.Name)
	assert.Equal(t, lpID, execution.LaunchPlanID)
	assert.Equal(t, wfID, execution.WorkflowID)
	assert.EqualValues(t, createdAt, *execution.ExecutionCreatedAt)
	assert.EqualValues(t, createdAt, *execution.ExecutionUpdatedAt)
	assert.Equal(t, int32(admin.ExecutionMetadata_SYSTEM), execution.Mode)
	assert.Equal(t, nodeID, execution.ParentNodeExecutionID)
	assert.Equal(t, sourceID, execution.SourceExecutionID)
	expectedSpec := execRequest.Spec
	expectedSpec.Metadata.Principal = principal
	expectedSpec.Metadata.SystemMetadata = &admin.SystemMetadata{
		ExecutionCluster: cluster,
	}
	expectedSpecBytes, _ := proto.Marshal(expectedSpec)
	assert.Equal(t, expectedSpecBytes, execution.Spec)
	assert.Equal(t, execution.User, principal)

	expectedCreatedAt, _ := ptypes.TimestampProto(createdAt)
	expectedClosure, _ := proto.Marshal(&admin.ExecutionClosure{
		Phase:      core.WorkflowExecution_RUNNING,
		CreatedAt:  expectedCreatedAt,
		StartedAt:  expectedCreatedAt,
		UpdatedAt:  expectedCreatedAt,
		WorkflowId: workflowIdentifier,
	})
	assert.Equal(t, expectedClosure, execution.Closure)
}

func TestUpdateModelState_UnknownToRunning(t *testing.T) {

	createdAt := time.Date(2018, 10, 29, 16, 0, 0, 0, time.UTC)
	createdAtProto, _ := ptypes.TimestampProto(createdAt)
	existingClosure := admin.ExecutionClosure{
		ComputedInputs: &core.LiteralMap{
			Literals: map[string]*core.Literal{
				"foo": {},
			},
		},
		Phase:     core.WorkflowExecution_UNDEFINED,
		CreatedAt: createdAtProto,
	}
	spec := testutils.GetExecutionRequest().Spec
	specBytes, _ := proto.Marshal(spec)
	existingClosureBytes, _ := proto.Marshal(&existingClosure)
	startedAt := time.Now()
	executionModel := getRunningExecutionModel(specBytes, existingClosureBytes, startedAt)

	occurredAt := time.Date(2018, 10, 29, 16, 10, 0, 0, time.UTC)
	occurredAtProto, _ := ptypes.TimestampProto(occurredAt)
	err := UpdateExecutionModelState(&executionModel, admin.WorkflowExecutionEventRequest{
		Event: &event.WorkflowExecutionEvent{
			Phase:      core.WorkflowExecution_RUNNING,
			OccurredAt: occurredAtProto,
		},
	})
	assert.Nil(t, err)

	expectedClosure := admin.ExecutionClosure{
		ComputedInputs: &core.LiteralMap{
			Literals: map[string]*core.Literal{
				"foo": {},
			},
		},
		Phase:     core.WorkflowExecution_RUNNING,
		StartedAt: occurredAtProto,
		UpdatedAt: occurredAtProto,
		CreatedAt: createdAtProto,
	}
	expectedClosureBytes, _ := proto.Marshal(&expectedClosure)
	expectedModel := models.Execution{
		ExecutionKey: models.ExecutionKey{
			Project: "project",
			Domain:  "domain",
			Name:    "name",
		},
		Spec:               specBytes,
		Phase:              core.WorkflowExecution_RUNNING.String(),
		Closure:            expectedClosureBytes,
		LaunchPlanID:       uint(1),
		WorkflowID:         uint(2),
		StartedAt:          &occurredAt,
		ExecutionCreatedAt: executionModel.ExecutionCreatedAt,
		ExecutionUpdatedAt: &occurredAt,
	}
	assert.EqualValues(t, expectedModel, executionModel)
}

func TestUpdateModelState_RunningToFailed(t *testing.T) {
	startedAt := time.Now()
	startedAtProto, _ := ptypes.TimestampProto(startedAt)
	existingClosure := admin.ExecutionClosure{
		ComputedInputs: &core.LiteralMap{
			Literals: map[string]*core.Literal{
				"foo": {},
			},
		},
		Phase:     core.WorkflowExecution_RUNNING,
		StartedAt: startedAtProto,
	}
	ec := "foo"
	ek := core.ExecutionError_SYSTEM
	spec := testutils.GetExecutionRequest().Spec
	specBytes, _ := proto.Marshal(spec)
	existingClosureBytes, _ := proto.Marshal(&existingClosure)
	executionModel := getRunningExecutionModel(specBytes, existingClosureBytes, startedAt)
	duration := time.Minute
	occurredAt := startedAt.Add(duration).UTC()
	occurredAtProto, _ := ptypes.TimestampProto(occurredAt)
	executionError := core.ExecutionError{
		Code:    ec,
		Kind:    ek,
		Message: "bar baz",
	}
	err := UpdateExecutionModelState(&executionModel, admin.WorkflowExecutionEventRequest{
		Event: &event.WorkflowExecutionEvent{
			Phase:      core.WorkflowExecution_ABORTED,
			OccurredAt: occurredAtProto,
			OutputResult: &event.WorkflowExecutionEvent_Error{
				Error: &executionError,
			},
		},
	})
	assert.Nil(t, err)

	ekString := ek.String()
	durationProto := ptypes.DurationProto(duration)
	expectedClosure := admin.ExecutionClosure{
		ComputedInputs: &core.LiteralMap{
			Literals: map[string]*core.Literal{
				"foo": {},
			},
		},
		Phase:     core.WorkflowExecution_ABORTED,
		StartedAt: startedAtProto,
		UpdatedAt: occurredAtProto,
		Duration:  durationProto,
		OutputResult: &admin.ExecutionClosure_Error{
			Error: &executionError,
		},
	}
	expectedClosureBytes, _ := proto.Marshal(&expectedClosure)
	expectedModel := models.Execution{
		ExecutionKey: models.ExecutionKey{
			Project: "project",
			Domain:  "domain",
			Name:    "name",
		},
		Spec:               specBytes,
		Phase:              core.WorkflowExecution_ABORTED.String(),
		Closure:            expectedClosureBytes,
		LaunchPlanID:       uint(1),
		WorkflowID:         uint(2),
		StartedAt:          &startedAt,
		Duration:           duration,
		ExecutionCreatedAt: executionModel.ExecutionCreatedAt,
		ExecutionUpdatedAt: &occurredAt,
		ErrorCode:          &ec,
		ErrorKind:          &ekString,
	}
	assert.EqualValues(t, expectedModel, executionModel)
}

func TestUpdateModelState_RunningToSuccess(t *testing.T) {
	startedAt := time.Now()
	startedAtProto, _ := ptypes.TimestampProto(startedAt)
	existingClosure := admin.ExecutionClosure{
		ComputedInputs: &core.LiteralMap{
			Literals: map[string]*core.Literal{
				"foo": {},
			},
		},
		Phase:     core.WorkflowExecution_RUNNING,
		StartedAt: startedAtProto,
	}
	spec := testutils.GetExecutionRequest().Spec
	specBytes, _ := proto.Marshal(spec)
	existingClosureBytes, _ := proto.Marshal(&existingClosure)
	executionModel := getRunningExecutionModel(specBytes, existingClosureBytes, startedAt)
	duration := time.Minute
	occurredAt := startedAt.Add(duration).UTC()
	occurredAtProto, _ := ptypes.TimestampProto(occurredAt)

	err := UpdateExecutionModelState(&executionModel, admin.WorkflowExecutionEventRequest{
		Event: &event.WorkflowExecutionEvent{
			Phase:      core.WorkflowExecution_SUCCEEDED,
			OccurredAt: occurredAtProto,
			OutputResult: &event.WorkflowExecutionEvent_OutputUri{
				OutputUri: "output.pb",
			},
		},
	})
	assert.Nil(t, err)

	durationProto := ptypes.DurationProto(duration)
	expectedClosure := admin.ExecutionClosure{
		ComputedInputs: &core.LiteralMap{
			Literals: map[string]*core.Literal{
				"foo": {},
			},
		},
		Phase:     core.WorkflowExecution_SUCCEEDED,
		StartedAt: startedAtProto,
		UpdatedAt: occurredAtProto,
		Duration:  durationProto,
		OutputResult: &admin.ExecutionClosure_Outputs{
			Outputs: &admin.LiteralMapBlob{
				Data: &admin.LiteralMapBlob_Uri{
					Uri: "output.pb",
				},
			},
		},
	}
	expectedClosureBytes, _ := proto.Marshal(&expectedClosure)
	expectedModel := models.Execution{
		ExecutionKey: models.ExecutionKey{
			Project: "project",
			Domain:  "domain",
			Name:    "name",
		},
		Spec:               specBytes,
		Phase:              core.WorkflowExecution_SUCCEEDED.String(),
		Closure:            expectedClosureBytes,
		LaunchPlanID:       uint(1),
		WorkflowID:         uint(2),
		StartedAt:          &startedAt,
		Duration:           duration,
		ExecutionCreatedAt: executionModel.ExecutionCreatedAt,
		ExecutionUpdatedAt: &occurredAt,
	}
	assert.EqualValues(t, expectedModel, executionModel)
}

func TestSetExecutionAborted(t *testing.T) {
	existingClosure := admin.ExecutionClosure{
		Phase: core.WorkflowExecution_RUNNING,
	}
	existingClosureBytes, _ := proto.Marshal(&existingClosure)
	existingModel := models.Execution{
		Phase:   core.WorkflowExecution_RUNNING.String(),
		Closure: existingClosureBytes,
	}
	cause := "a snafoo occurred"
	principal := "principal"
	err := SetExecutionAborted(&existingModel, cause, principal)
	assert.NoError(t, err)
	var actualClosure admin.ExecutionClosure
	err = proto.Unmarshal(existingModel.Closure, &actualClosure)
	if err != nil {
		t.Fatal(fmt.Sprintf("Failed to marshal execution closure: %v", err))
	}
	assert.True(t, proto.Equal(&admin.ExecutionClosure{
		OutputResult: &admin.ExecutionClosure_AbortMetadata{
			AbortMetadata: &admin.AbortMetadata{
				Cause:     cause,
				Principal: principal,
			}},
		// The execution abort metadata is recorded but the phase is not actually updated *until* the abort event is
		// propagated by flytepropeller.
		Phase: core.WorkflowExecution_RUNNING,
	}, &actualClosure))
}

func TestGetExecutionIdentifier(t *testing.T) {
	executionModel := models.Execution{
		ExecutionKey: models.ExecutionKey{
			Project: "project",
			Domain:  "domain",
			Name:    "name",
		},
	}
	actualIdentifier := GetExecutionIdentifier(&executionModel)
	assert.True(t, proto.Equal(&core.WorkflowExecutionIdentifier{
		Project: "project",
		Domain:  "domain",
		Name:    "name",
	}, &actualIdentifier))
}

func TestFromExecutionModel(t *testing.T) {
	spec := testutils.GetExecutionRequest().Spec
	specBytes, _ := proto.Marshal(spec)
	phase := core.WorkflowExecution_RUNNING.String()
	startedAt := time.Date(2018, 8, 30, 0, 0, 0, 0, time.UTC)
	startedAtProto, _ := ptypes.TimestampProto(startedAt)
	closure := admin.ExecutionClosure{
		ComputedInputs: spec.Inputs,
		Phase:          core.WorkflowExecution_RUNNING,
		StartedAt:      startedAtProto,
	}
	closureBytes, _ := proto.Marshal(&closure)

	executionModel := models.Execution{
		ExecutionKey: models.ExecutionKey{
			Project: "project",
			Domain:  "domain",
			Name:    "name",
		},
		Spec:         specBytes,
		Phase:        phase,
		Closure:      closureBytes,
		LaunchPlanID: uint(1),
		WorkflowID:   uint(2),
		StartedAt:    &startedAt,
	}
	execution, err := FromExecutionModel(executionModel)
	assert.Nil(t, err)
	assert.True(t, proto.Equal(&admin.Execution{
		Id: &core.WorkflowExecutionIdentifier{
			Project: "project",
			Domain:  "domain",
			Name:    "name",
		},
		Spec:    spec,
		Closure: &closure,
	}, execution))
}

func TestFromExecutionModel_Aborted(t *testing.T) {
	abortCause := "abort cause"
	executionClosureBytes, _ := proto.Marshal(&admin.ExecutionClosure{
		Phase: core.WorkflowExecution_ABORTED,
	})
	executionModel := models.Execution{
		ExecutionKey: models.ExecutionKey{
			Project: "project",
			Domain:  "domain",
			Name:    "name",
		},
		Phase:      core.WorkflowExecution_ABORTED.String(),
		AbortCause: abortCause,
		Closure:    executionClosureBytes,
	}
	execution, err := FromExecutionModel(executionModel)
	assert.Nil(t, err)
	assert.Equal(t, core.WorkflowExecution_ABORTED, execution.Closure.Phase)
	assert.True(t, proto.Equal(&admin.AbortMetadata{
		Cause: abortCause,
	}, execution.Closure.GetAbortMetadata()))

	executionModel.Phase = core.WorkflowExecution_RUNNING.String()
	execution, err = FromExecutionModel(executionModel)
	assert.Nil(t, err)
	assert.Empty(t, execution.Closure.GetAbortCause())
}

func TestFromExecutionModels(t *testing.T) {
	spec := testutils.GetExecutionRequest().Spec
	specBytes, _ := proto.Marshal(spec)
	phase := core.WorkflowExecution_SUCCEEDED.String()
	startedAt := time.Date(2018, 8, 30, 0, 0, 0, 0, time.UTC)
	startedAtProto, _ := ptypes.TimestampProto(startedAt)
	duration := 2 * time.Minute
	durationProto := ptypes.DurationProto(duration)
	closure := admin.ExecutionClosure{
		ComputedInputs: spec.Inputs,
		Phase:          core.WorkflowExecution_RUNNING,
		StartedAt:      startedAtProto,
		Duration:       durationProto,
	}
	closureBytes, _ := proto.Marshal(&closure)

	executionModels := []models.Execution{
		{
			ExecutionKey: models.ExecutionKey{
				Project: "project",
				Domain:  "domain",
				Name:    "name",
			},
			Spec:         specBytes,
			Phase:        phase,
			Closure:      closureBytes,
			LaunchPlanID: uint(1),
			WorkflowID:   uint(2),
			StartedAt:    &startedAt,
			Duration:     duration,
		},
	}
	executions, err := FromExecutionModels(executionModels)
	assert.Nil(t, err)
	assert.Len(t, executions, 1)
	assert.True(t, proto.Equal(&admin.Execution{
		Id: &core.WorkflowExecutionIdentifier{
			Project: "project",
			Domain:  "domain",
			Name:    "name",
		},
		Spec:    spec,
		Closure: &closure,
	}, executions[0]))
}
