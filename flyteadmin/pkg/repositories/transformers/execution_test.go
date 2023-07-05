package transformers

import (
	"context"
	"math"
	"strings"
	"testing"
	"time"
	"unicode/utf8"

	"github.com/flyteorg/flyteadmin/pkg/common"

	"github.com/flyteorg/flyteadmin/pkg/errors"
	"google.golang.org/grpc/codes"

	commonMocks "github.com/flyteorg/flyteadmin/pkg/common/mocks"
	"github.com/flyteorg/flyteadmin/pkg/runtime/interfaces"
	"github.com/flyteorg/flytestdlib/storage"

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
	principal := "principal"
	execRequest.Spec.Metadata = &admin.ExecutionMetadata{
		Mode:      admin.ExecutionMetadata_SYSTEM,
		Principal: principal,
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

	cluster := "cluster"
	securityCtx := &core.SecurityContext{
		RunAs: &core.Identity{
			IamRole: "iam_role",
		},
	}
	namespace := "ns"
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
		Cluster:               cluster,
		SecurityContext:       securityCtx,
		LaunchEntity:          core.ResourceType_LAUNCH_PLAN,
		Namespace:             namespace,
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
	assert.Equal(t, "launch_plan", execution.LaunchEntity)
	expectedSpec := execRequest.Spec
	expectedSpec.Metadata.Principal = principal
	expectedSpec.Metadata.SystemMetadata = &admin.SystemMetadata{
		ExecutionCluster: cluster,
		Namespace:        namespace,
	}
	expectedSpec.SecurityContext = securityCtx
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
		StateChangeDetails: &admin.ExecutionStateChangeDetails{
			State:      admin.ExecutionState_EXECUTION_ACTIVE,
			OccurredAt: expectedCreatedAt,
			Principal:  principal,
		},
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
	err := UpdateExecutionModelState(context.TODO(), &executionModel, admin.WorkflowExecutionEventRequest{
		Event: &event.WorkflowExecutionEvent{
			Phase:      core.WorkflowExecution_RUNNING,
			OccurredAt: occurredAtProto,
		},
	}, interfaces.InlineEventDataPolicyStoreInline, commonMocks.GetMockStorageClient())
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
	err := UpdateExecutionModelState(context.TODO(), &executionModel, admin.WorkflowExecutionEventRequest{
		Event: &event.WorkflowExecutionEvent{
			Phase:      core.WorkflowExecution_ABORTED,
			OccurredAt: occurredAtProto,
			OutputResult: &event.WorkflowExecutionEvent_Error{
				Error: &executionError,
			},
		},
	}, interfaces.InlineEventDataPolicyStoreInline, commonMocks.GetMockStorageClient())
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
	durationProto := ptypes.DurationProto(duration)
	occurredAt := startedAt.Add(duration).UTC()
	occurredAtProto, _ := ptypes.TimestampProto(occurredAt)

	expectedModel := models.Execution{
		ExecutionKey: models.ExecutionKey{
			Project: "project",
			Domain:  "domain",
			Name:    "name",
		},
		Spec:               specBytes,
		Phase:              core.WorkflowExecution_SUCCEEDED.String(),
		LaunchPlanID:       uint(1),
		WorkflowID:         uint(2),
		StartedAt:          &startedAt,
		Duration:           duration,
		ExecutionCreatedAt: executionModel.ExecutionCreatedAt,
		ExecutionUpdatedAt: &occurredAt,
	}

	t.Run("output URI set", func(t *testing.T) {
		err := UpdateExecutionModelState(context.TODO(), &executionModel, admin.WorkflowExecutionEventRequest{
			Event: &event.WorkflowExecutionEvent{
				Phase:      core.WorkflowExecution_SUCCEEDED,
				OccurredAt: occurredAtProto,
				OutputResult: &event.WorkflowExecutionEvent_OutputUri{
					OutputUri: "output.pb",
				},
			},
		}, interfaces.InlineEventDataPolicyStoreInline, commonMocks.GetMockStorageClient())
		assert.Nil(t, err)

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
		closureBytes, _ := proto.Marshal(&expectedClosure)
		expectedModel.Closure = closureBytes
		assert.EqualValues(t, expectedModel, executionModel)
	})
	t.Run("output data set", func(t *testing.T) {
		outputData := &core.LiteralMap{
			Literals: map[string]*core.Literal{
				"foo": {
					Value: &core.Literal_Scalar{
						Scalar: &core.Scalar{
							Value: &core.Scalar_Primitive{
								Primitive: &core.Primitive{
									Value: &core.Primitive_Integer{
										Integer: 4,
									},
								},
							},
						},
					},
				},
			},
		}
		err := UpdateExecutionModelState(context.TODO(), &executionModel, admin.WorkflowExecutionEventRequest{
			Event: &event.WorkflowExecutionEvent{
				Phase:      core.WorkflowExecution_SUCCEEDED,
				OccurredAt: occurredAtProto,
				OutputResult: &event.WorkflowExecutionEvent_OutputData{
					OutputData: outputData,
				},
			},
		}, interfaces.InlineEventDataPolicyStoreInline, commonMocks.GetMockStorageClient())
		assert.Nil(t, err)

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
			OutputResult: &admin.ExecutionClosure_OutputData{
				OutputData: outputData,
			},
		}
		closureBytes, _ := proto.Marshal(&expectedClosure)
		expectedModel.Closure = closureBytes
		assert.EqualValues(t, expectedModel, executionModel)
	})
	t.Run("output data offloaded", func(t *testing.T) {
		outputData := &core.LiteralMap{
			Literals: map[string]*core.Literal{
				"foo": {
					Value: &core.Literal_Scalar{
						Scalar: &core.Scalar{
							Value: &core.Scalar_Primitive{
								Primitive: &core.Primitive{
									Value: &core.Primitive_Integer{
										Integer: 4,
									},
								},
							},
						},
					},
				},
			},
		}
		mockStorage := commonMocks.GetMockStorageClient()
		mockStorage.ComposedProtobufStore.(*commonMocks.TestDataStore).WriteProtobufCb = func(ctx context.Context, reference storage.DataReference, opts storage.Options, msg proto.Message) error {
			assert.Equal(t, reference.String(), "s3://bucket/metadata/project/domain/name/offloaded_outputs")
			return nil
		}
		err := UpdateExecutionModelState(context.TODO(), &executionModel, admin.WorkflowExecutionEventRequest{
			Event: &event.WorkflowExecutionEvent{
				ExecutionId: &core.WorkflowExecutionIdentifier{
					Project: "project",
					Domain:  "domain",
					Name:    "name",
				},
				Phase:      core.WorkflowExecution_SUCCEEDED,
				OccurredAt: occurredAtProto,
				OutputResult: &event.WorkflowExecutionEvent_OutputData{
					OutputData: outputData,
				},
			},
		}, interfaces.InlineEventDataPolicyOffload, mockStorage)
		assert.Nil(t, err)

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
						Uri: "s3://bucket/metadata/project/domain/name/offloaded_outputs",
					},
				},
			},
		}
		closureBytes, _ := proto.Marshal(&expectedClosure)
		expectedModel.Closure = closureBytes
		assert.EqualValues(t, expectedModel, executionModel)
	})
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
	err := SetExecutionAborting(&existingModel, cause, principal)
	assert.NoError(t, err)
	var actualClosure admin.ExecutionClosure
	err = proto.Unmarshal(existingModel.Closure, &actualClosure)
	if err != nil {
		t.Fatalf("Failed to marshal execution closure: %v", err)
	}
	assert.True(t, proto.Equal(&admin.ExecutionClosure{
		OutputResult: &admin.ExecutionClosure_AbortMetadata{
			AbortMetadata: &admin.AbortMetadata{
				Cause:     cause,
				Principal: principal,
			}},
		// The execution abort metadata is recorded but the phase is not actually updated *until* the abort event is
		// propagated by flytepropeller.
		Phase: core.WorkflowExecution_ABORTING,
	}, &actualClosure))
	assert.Equal(t, existingModel.AbortCause, cause)
	assert.Equal(t, existingModel.Phase, core.WorkflowExecution_ABORTING.String())
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
	createdAt := time.Date(2022, 01, 18, 0, 0, 0, 0, time.UTC)
	startedAtProto, _ := ptypes.TimestampProto(startedAt)
	createdAtProto, _ := ptypes.TimestampProto(createdAt)
	closure := admin.ExecutionClosure{
		ComputedInputs: spec.Inputs,
		Phase:          core.WorkflowExecution_RUNNING,
		StartedAt:      startedAtProto,
		StateChangeDetails: &admin.ExecutionStateChangeDetails{
			State:      admin.ExecutionState_EXECUTION_ACTIVE,
			OccurredAt: createdAtProto,
		},
	}
	closureBytes, _ := proto.Marshal(&closure)
	stateInt := int32(admin.ExecutionState_EXECUTION_ACTIVE)
	executionModel := models.Execution{
		BaseModel: models.BaseModel{
			CreatedAt: createdAt,
		},
		ExecutionKey: models.ExecutionKey{
			Project: "project",
			Domain:  "domain",
			Name:    "name",
		},
		User:         "",
		Spec:         specBytes,
		Phase:        phase,
		Closure:      closureBytes,
		LaunchPlanID: uint(1),
		WorkflowID:   uint(2),
		StartedAt:    &startedAt,
		State:        &stateInt,
	}
	execution, err := FromExecutionModel(context.TODO(), executionModel, DefaultExecutionTransformerOptions)
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
	execution, err := FromExecutionModel(context.TODO(), executionModel, DefaultExecutionTransformerOptions)
	assert.Nil(t, err)
	assert.Equal(t, core.WorkflowExecution_ABORTED, execution.Closure.Phase)
	assert.True(t, proto.Equal(&admin.AbortMetadata{
		Cause: abortCause,
	}, execution.Closure.GetAbortMetadata()))

	executionModel.Phase = core.WorkflowExecution_RUNNING.String()
	execution, err = FromExecutionModel(context.TODO(), executionModel, DefaultExecutionTransformerOptions)
	assert.Nil(t, err)
	assert.Empty(t, execution.Closure.GetAbortCause())
}

func TestFromExecutionModel_Error(t *testing.T) {
	extraLongErrMsg := string(make([]byte, 2*trimmedErrMessageLen))
	execErr := &core.ExecutionError{
		Code:    "CODE",
		Message: extraLongErrMsg,
		Kind:    core.ExecutionError_USER,
	}
	executionClosureBytes, _ := proto.Marshal(&admin.ExecutionClosure{
		Phase:        core.WorkflowExecution_FAILED,
		OutputResult: &admin.ExecutionClosure_Error{Error: execErr},
	})
	executionModel := models.Execution{
		ExecutionKey: models.ExecutionKey{
			Project: "project",
			Domain:  "domain",
			Name:    "name",
		},
		Phase:   core.WorkflowExecution_FAILED.String(),
		Closure: executionClosureBytes,
	}
	execution, err := FromExecutionModel(context.TODO(), executionModel, &ExecutionTransformerOptions{
		TrimErrorMessage: true,
	})
	expectedExecErr := execErr
	expectedExecErr.Message = string(make([]byte, trimmedErrMessageLen))
	assert.Nil(t, err)
	assert.Equal(t, core.WorkflowExecution_FAILED, execution.Closure.Phase)
	assert.True(t, proto.Equal(expectedExecErr, execution.Closure.GetError()))
}

func TestFromExecutionModel_ValidUTF8TrimmedErrorMsg(t *testing.T) {
	errMsg := "[1/1] currentAttempt done. Last Error: USER::                   │\n│ ❱  760 │   │   │   │   return __callback(*args, **kwargs)                    │\n││"

	executionClosureBytes, _ := proto.Marshal(&admin.ExecutionClosure{
		Phase:        core.WorkflowExecution_FAILED,
		OutputResult: &admin.ExecutionClosure_Error{Error: &core.ExecutionError{Message: errMsg}},
	})
	executionModel := models.Execution{
		ExecutionKey: models.ExecutionKey{
			Project: "project",
			Domain:  "domain",
			Name:    "name",
		},
		Phase:   core.WorkflowExecution_FAILED.String(),
		Closure: executionClosureBytes,
	}
	execution, err := FromExecutionModel(context.TODO(), executionModel, &ExecutionTransformerOptions{
		TrimErrorMessage: true,
	})
	assert.NoError(t, err)
	errMsgAreValidUTF8 := utf8.Valid([]byte(execution.GetClosure().GetError().GetMessage()))
	assert.True(t, errMsgAreValidUTF8)
}

func TestFromExecutionModel_OverwriteNamespace(t *testing.T) {
	abortCause := "abort cause"
	executionClosureBytes, _ := proto.Marshal(&admin.ExecutionClosure{
		Phase: core.WorkflowExecution_RUNNING,
	})
	executionModel := models.Execution{
		ExecutionKey: models.ExecutionKey{
			Project: "project",
			Domain:  "domain",
			Name:    "name",
		},
		Phase:      core.WorkflowExecution_RUNNING.String(),
		AbortCause: abortCause,
		Closure:    executionClosureBytes,
	}
	overwrittenNamespace := "ns"
	execution, err := FromExecutionModel(context.TODO(), executionModel, &ExecutionTransformerOptions{
		DefaultNamespace: overwrittenNamespace,
	})
	assert.NoError(t, err)
	assert.Equal(t, execution.GetSpec().GetMetadata().GetSystemMetadata().Namespace, overwrittenNamespace)
}

func TestFromExecutionModels(t *testing.T) {
	spec := testutils.GetExecutionRequest().Spec
	specBytes, _ := proto.Marshal(spec)
	phase := core.WorkflowExecution_SUCCEEDED.String()
	startedAt := time.Date(2018, 8, 30, 0, 0, 0, 0, time.UTC)
	createdAt := time.Date(2022, 01, 18, 0, 0, 0, 0, time.UTC)
	startedAtProto, _ := ptypes.TimestampProto(startedAt)
	createdAtProto, _ := ptypes.TimestampProto(createdAt)
	duration := 2 * time.Minute
	durationProto := ptypes.DurationProto(duration)
	closure := admin.ExecutionClosure{
		ComputedInputs: spec.Inputs,
		Phase:          core.WorkflowExecution_RUNNING,
		StartedAt:      startedAtProto,
		Duration:       durationProto,
		StateChangeDetails: &admin.ExecutionStateChangeDetails{
			State:      admin.ExecutionState_EXECUTION_ACTIVE,
			OccurredAt: createdAtProto,
		},
	}
	closureBytes, _ := proto.Marshal(&closure)
	stateInt := int32(admin.ExecutionState_EXECUTION_ACTIVE)
	executionModels := []models.Execution{
		{
			BaseModel: models.BaseModel{
				CreatedAt: createdAt,
			},
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
			State:        &stateInt,
		},
	}
	executions, err := FromExecutionModels(context.TODO(), executionModels, DefaultExecutionTransformerOptions)
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

func TestUpdateModelState_WithClusterInformation(t *testing.T) {
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
	testCluster := "C1"
	altCluster := "C2"
	executionModel.Cluster = testCluster
	occurredAt := time.Date(2018, 10, 29, 16, 10, 0, 0, time.UTC)
	occurredAtProto, _ := ptypes.TimestampProto(occurredAt)
	t.Run("update", func(t *testing.T) {
		executionModel.Cluster = altCluster
		err := UpdateExecutionModelState(context.TODO(), &executionModel, admin.WorkflowExecutionEventRequest{
			Event: &event.WorkflowExecutionEvent{
				Phase:      core.WorkflowExecution_QUEUED,
				OccurredAt: occurredAtProto,
				ProducerId: testCluster,
			},
		}, interfaces.InlineEventDataPolicyStoreInline, commonMocks.GetMockStorageClient())
		assert.NoError(t, err)
		assert.Equal(t, testCluster, executionModel.Cluster)
		executionModel.Cluster = testCluster
	})
	t.Run("do not update", func(t *testing.T) {
		err := UpdateExecutionModelState(context.TODO(), &executionModel, admin.WorkflowExecutionEventRequest{
			Event: &event.WorkflowExecutionEvent{
				Phase:      core.WorkflowExecution_RUNNING,
				OccurredAt: occurredAtProto,
				ProducerId: altCluster,
			},
		}, interfaces.InlineEventDataPolicyStoreInline, commonMocks.GetMockStorageClient())
		assert.Equal(t, err.(errors.FlyteAdminError).Code(), codes.FailedPrecondition)
	})
	t.Run("matches recorded", func(t *testing.T) {
		executionModel.Cluster = testCluster
		err := UpdateExecutionModelState(context.TODO(), &executionModel, admin.WorkflowExecutionEventRequest{
			Event: &event.WorkflowExecutionEvent{
				Phase:      core.WorkflowExecution_RUNNING,
				OccurredAt: occurredAtProto,
				ProducerId: testCluster,
			},
		}, interfaces.InlineEventDataPolicyStoreInline, commonMocks.GetMockStorageClient())
		assert.NoError(t, err)
	})
	t.Run("default cluster value", func(t *testing.T) {
		executionModel.Cluster = testCluster
		err := UpdateExecutionModelState(context.TODO(), &executionModel, admin.WorkflowExecutionEventRequest{
			Event: &event.WorkflowExecutionEvent{
				Phase:      core.WorkflowExecution_RUNNING,
				OccurredAt: occurredAtProto,
				ProducerId: common.DefaultProducerID,
			},
		}, interfaces.InlineEventDataPolicyStoreInline, commonMocks.GetMockStorageClient())
		assert.NoError(t, err)
	})
}

func TestReassignCluster(t *testing.T) {
	oldCluster := "old_cluster"
	newCluster := "new_cluster"

	workflowExecutionID := core.WorkflowExecutionIdentifier{
		Project: "project",
		Domain:  "domain",
		Name:    "name",
	}

	t.Run("happy case", func(t *testing.T) {
		spec := testutils.GetExecutionRequest().Spec
		spec.Metadata = &admin.ExecutionMetadata{
			SystemMetadata: &admin.SystemMetadata{
				ExecutionCluster: oldCluster,
			},
		}
		specBytes, _ := proto.Marshal(spec)
		executionModel := models.Execution{
			Spec:    specBytes,
			Cluster: oldCluster,
		}
		err := reassignCluster(context.TODO(), newCluster, &workflowExecutionID, &executionModel)
		assert.NoError(t, err)
		assert.Equal(t, newCluster, executionModel.Cluster)

		var updatedSpec admin.ExecutionSpec
		err = proto.Unmarshal(executionModel.Spec, &updatedSpec)
		assert.NoError(t, err)
		assert.Equal(t, newCluster, updatedSpec.Metadata.SystemMetadata.ExecutionCluster)
	})
	t.Run("happy case - initialize cluster", func(t *testing.T) {
		spec := testutils.GetExecutionRequest().Spec
		specBytes, _ := proto.Marshal(spec)
		executionModel := models.Execution{
			Spec: specBytes,
		}
		err := reassignCluster(context.TODO(), newCluster, &workflowExecutionID, &executionModel)
		assert.NoError(t, err)
		assert.Equal(t, newCluster, executionModel.Cluster)

		var updatedSpec admin.ExecutionSpec
		err = proto.Unmarshal(executionModel.Spec, &updatedSpec)
		assert.NoError(t, err)
		assert.Equal(t, newCluster, updatedSpec.Metadata.SystemMetadata.ExecutionCluster)
	})
	t.Run("invalid existing spec", func(t *testing.T) {
		executionModel := models.Execution{
			Spec:    []byte("I'm invalid"),
			Cluster: oldCluster,
		}
		err := reassignCluster(context.TODO(), newCluster, &workflowExecutionID, &executionModel)
		assert.Equal(t, err.(errors.FlyteAdminError).Code(), codes.Internal)
	})
}

func TestGetExecutionStateFromModel(t *testing.T) {
	createdAt := time.Date(2022, 01, 90, 16, 0, 0, 0, time.UTC)
	createdAtProto, _ := ptypes.TimestampProto(createdAt)

	t.Run("supporting older executions", func(t *testing.T) {
		executionModel := models.Execution{
			BaseModel: models.BaseModel{
				CreatedAt: createdAt,
			},
		}
		executionStatus, err := PopulateDefaultStateChangeDetails(executionModel)
		assert.Nil(t, err)
		assert.NotNil(t, executionStatus)
		assert.Equal(t, admin.ExecutionState_EXECUTION_ACTIVE, executionStatus.State)
		assert.NotNil(t, executionStatus.OccurredAt)
		assert.Equal(t, createdAtProto, executionStatus.OccurredAt)
	})
	t.Run("incorrect created at", func(t *testing.T) {
		createdAt := time.Unix(math.MinInt64, math.MinInt32).UTC()
		executionModel := models.Execution{
			BaseModel: models.BaseModel{
				CreatedAt: createdAt,
			},
		}
		executionStatus, err := PopulateDefaultStateChangeDetails(executionModel)
		assert.NotNil(t, err)
		assert.Nil(t, executionStatus)
	})
}

func TestUpdateExecutionModelStateChangeDetails(t *testing.T) {
	t.Run("empty closure", func(t *testing.T) {
		execModel := &models.Execution{}
		stateUpdatedAt := time.Now()
		statetUpdateAtProto, err := ptypes.TimestampProto(stateUpdatedAt)
		assert.Nil(t, err)
		err = UpdateExecutionModelStateChangeDetails(execModel, admin.ExecutionState_EXECUTION_ARCHIVED,
			stateUpdatedAt, "dummyUser")
		assert.Nil(t, err)
		stateInt := int32(admin.ExecutionState_EXECUTION_ARCHIVED)
		assert.Equal(t, execModel.State, &stateInt)
		var closure admin.ExecutionClosure
		err = proto.Unmarshal(execModel.Closure, &closure)
		assert.Nil(t, err)
		assert.NotNil(t, closure)
		assert.NotNil(t, closure.StateChangeDetails)
		assert.Equal(t, admin.ExecutionState_EXECUTION_ARCHIVED, closure.StateChangeDetails.State)
		assert.Equal(t, "dummyUser", closure.StateChangeDetails.Principal)
		assert.Equal(t, statetUpdateAtProto, closure.StateChangeDetails.OccurredAt)

	})
	t.Run("bad closure", func(t *testing.T) {
		execModel := &models.Execution{
			Closure: []byte{1, 2, 3},
		}
		err := UpdateExecutionModelStateChangeDetails(execModel, admin.ExecutionState_EXECUTION_ARCHIVED,
			time.Now(), "dummyUser")
		assert.NotNil(t, err)
		assert.Contains(t, err.Error(), "Failed to unmarshal execution closure")
	})
	t.Run("bad stateUpdatedAt time", func(t *testing.T) {
		execModel := &models.Execution{}
		badTimeData := time.Unix(math.MinInt64, math.MinInt32).UTC()
		err := UpdateExecutionModelStateChangeDetails(execModel, admin.ExecutionState_EXECUTION_ARCHIVED,
			badTimeData, "dummyUser")
		assert.NotNil(t, err)
		assert.False(t, strings.Contains(err.Error(), "Failed to unmarshal execution closure"))
	})
}

func TestTrimErrorMessage(t *testing.T) {
	errMsg := "[1/1] currentAttempt done. Last Error: USER::                   │\n│ ❱  760 │   │   │   │   return __callback(*args, **kwargs)                    │\n││"
	trimmedErrMessage := TrimErrorMessage(errMsg)
	errMsgAreValidUTF8 := utf8.Valid([]byte(trimmedErrMessage))
	assert.True(t, errMsgAreValidUTF8)
}
