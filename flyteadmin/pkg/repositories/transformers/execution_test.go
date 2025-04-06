package transformers

import (
	"context"
	"fmt"
	"google.golang.org/protobuf/types/known/durationpb"
	"google.golang.org/protobuf/types/known/timestamppb"
	"testing"
	"time"
	"unicode/utf8"

	"github.com/stretchr/testify/assert"
	"google.golang.org/grpc/codes"
	"google.golang.org/protobuf/proto"

	"github.com/flyteorg/flyte/flyteadmin/pkg/common"
	commonMocks "github.com/flyteorg/flyte/flyteadmin/pkg/common/mocks"
	"github.com/flyteorg/flyte/flyteadmin/pkg/errors"
	"github.com/flyteorg/flyte/flyteadmin/pkg/manager/impl/testutils"
	"github.com/flyteorg/flyte/flyteadmin/pkg/repositories/models"
	"github.com/flyteorg/flyte/flyteadmin/pkg/runtime/interfaces"
	"github.com/flyteorg/flyte/flyteidl/gen/pb-go/flyteidl/admin"
	"github.com/flyteorg/flyte/flyteidl/gen/pb-go/flyteidl/core"
	"github.com/flyteorg/flyte/flyteidl/gen/pb-go/flyteidl/event"
	"github.com/flyteorg/flyte/flytestdlib/storage"
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
	t.Run("successful execution", func(t *testing.T) {
		execution, err := CreateExecutionModel(CreateExecutionModelInput{
			WorkflowExecutionID: &core.WorkflowExecutionIdentifier{
				Project: "project",
				Domain:  "domain",
				Name:    "name",
			},
			RequestSpec:           execRequest.GetSpec(),
			LaunchPlanID:          lpID,
			WorkflowID:            wfID,
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
		assert.Equal(t, execution.Phase, core.WorkflowExecution_UNDEFINED.String())
		expectedSpec := execRequest.GetSpec()
		expectedSpec.Metadata.Principal = principal
		expectedSpec.Metadata.SystemMetadata = &admin.SystemMetadata{
			ExecutionCluster: cluster,
			Namespace:        namespace,
		}
		expectedSpec.SecurityContext = securityCtx
		expectedSpecBytes, _ := proto.Marshal(expectedSpec)
		assert.Equal(t, expectedSpecBytes, execution.Spec)
		assert.Equal(t, execution.User, principal)

		expectedCreatedAt := timestamppb.New(createdAt)
		expectedClosure, _ := proto.Marshal(&admin.ExecutionClosure{
			Phase:      core.WorkflowExecution_UNDEFINED,
			CreatedAt:  expectedCreatedAt,
			UpdatedAt:  expectedCreatedAt,
			WorkflowId: workflowIdentifier,
			StateChangeDetails: &admin.ExecutionStateChangeDetails{
				State:      admin.ExecutionState_EXECUTION_ACTIVE,
				OccurredAt: expectedCreatedAt,
				Principal:  principal,
			},
		})
		assert.Equal(t, expectedClosure, execution.Closure)
	})
	t.Run("failed with unknown error", func(t *testing.T) {
		execErr := fmt.Errorf("bla-bla")
		execution, err := CreateExecutionModel(CreateExecutionModelInput{
			WorkflowExecutionID: &core.WorkflowExecutionIdentifier{
				Project: "project",
				Domain:  "domain",
				Name:    "name",
			},
			RequestSpec:           execRequest.GetSpec(),
			LaunchPlanID:          lpID,
			WorkflowID:            wfID,
			CreatedAt:             createdAt,
			WorkflowIdentifier:    workflowIdentifier,
			ParentNodeExecutionID: nodeID,
			SourceExecutionID:     sourceID,
			Cluster:               cluster,
			SecurityContext:       securityCtx,
			LaunchEntity:          core.ResourceType_LAUNCH_PLAN,
			Namespace:             namespace,
			Error:                 execErr,
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
		assert.Equal(t, core.WorkflowExecution_FAILED.String(), execution.Phase)
		expectedSpec := execRequest.GetSpec()
		expectedSpec.Metadata.Principal = principal
		expectedSpec.Metadata.SystemMetadata = &admin.SystemMetadata{
			ExecutionCluster: cluster,
			Namespace:        namespace,
		}
		expectedSpec.SecurityContext = securityCtx
		expectedSpecBytes, _ := proto.Marshal(expectedSpec)
		assert.Equal(t, expectedSpecBytes, execution.Spec)
		assert.Equal(t, execution.User, principal)

		expectedCreatedAt := timestamppb.New(createdAt)
		expectedClosure, _ := proto.Marshal(&admin.ExecutionClosure{
			Phase: core.WorkflowExecution_FAILED,
			OutputResult: &admin.ExecutionClosure_Error{
				Error: &core.ExecutionError{
					Code:    "Unknown",
					Message: execErr.Error(),
					Kind:    core.ExecutionError_SYSTEM,
				},
			},
			CreatedAt:  expectedCreatedAt,
			UpdatedAt:  expectedCreatedAt,
			WorkflowId: workflowIdentifier,
			StateChangeDetails: &admin.ExecutionStateChangeDetails{
				State:      admin.ExecutionState_EXECUTION_ACTIVE,
				OccurredAt: expectedCreatedAt,
				Principal:  principal,
			},
		})
		assert.Equal(t, string(expectedClosure), string(execution.Closure))
	})
	t.Run("failed with invalid argument error", func(t *testing.T) {
		execErr := errors.NewFlyteAdminError(codes.InvalidArgument, "task validation failed")
		execution, err := CreateExecutionModel(CreateExecutionModelInput{
			WorkflowExecutionID: &core.WorkflowExecutionIdentifier{
				Project: "project",
				Domain:  "domain",
				Name:    "name",
			},
			RequestSpec:           execRequest.GetSpec(),
			LaunchPlanID:          lpID,
			WorkflowID:            wfID,
			CreatedAt:             createdAt,
			WorkflowIdentifier:    workflowIdentifier,
			ParentNodeExecutionID: nodeID,
			SourceExecutionID:     sourceID,
			Cluster:               cluster,
			SecurityContext:       securityCtx,
			LaunchEntity:          core.ResourceType_LAUNCH_PLAN,
			Namespace:             namespace,
			Error:                 execErr,
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
		assert.Equal(t, core.WorkflowExecution_FAILED.String(), execution.Phase)
		expectedSpec := execRequest.GetSpec()
		expectedSpec.Metadata.Principal = principal
		expectedSpec.Metadata.SystemMetadata = &admin.SystemMetadata{
			ExecutionCluster: cluster,
			Namespace:        namespace,
		}
		expectedSpec.SecurityContext = securityCtx
		expectedSpecBytes, _ := proto.Marshal(expectedSpec)
		assert.Equal(t, expectedSpecBytes, execution.Spec)
		assert.Equal(t, execution.User, principal)

		expectedCreatedAt := timestamppb.New(createdAt)
		expectedClosure, _ := proto.Marshal(&admin.ExecutionClosure{
			Phase: core.WorkflowExecution_FAILED,
			OutputResult: &admin.ExecutionClosure_Error{
				Error: &core.ExecutionError{
					Code:    execErr.Code().String(),
					Message: execErr.Error(),
					Kind:    core.ExecutionError_USER,
				},
			},
			CreatedAt:  expectedCreatedAt,
			UpdatedAt:  expectedCreatedAt,
			WorkflowId: workflowIdentifier,
			StateChangeDetails: &admin.ExecutionStateChangeDetails{
				State:      admin.ExecutionState_EXECUTION_ACTIVE,
				OccurredAt: expectedCreatedAt,
				Principal:  principal,
			},
		})
		assert.Equal(t, expectedClosure, execution.Closure)
	})
	t.Run("failed with internal error", func(t *testing.T) {
		execErr := errors.NewFlyteAdminError(codes.Internal, "db failed")
		execution, err := CreateExecutionModel(CreateExecutionModelInput{
			WorkflowExecutionID: &core.WorkflowExecutionIdentifier{
				Project: "project",
				Domain:  "domain",
				Name:    "name",
			},
			RequestSpec:           execRequest.GetSpec(),
			LaunchPlanID:          lpID,
			WorkflowID:            wfID,
			CreatedAt:             createdAt,
			WorkflowIdentifier:    workflowIdentifier,
			ParentNodeExecutionID: nodeID,
			SourceExecutionID:     sourceID,
			Cluster:               cluster,
			SecurityContext:       securityCtx,
			LaunchEntity:          core.ResourceType_LAUNCH_PLAN,
			Namespace:             namespace,
			Error:                 execErr,
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
		assert.Equal(t, core.WorkflowExecution_FAILED.String(), execution.Phase)
		expectedSpec := execRequest.GetSpec()
		expectedSpec.Metadata.Principal = principal
		expectedSpec.Metadata.SystemMetadata = &admin.SystemMetadata{
			ExecutionCluster: cluster,
			Namespace:        namespace,
		}
		expectedSpec.SecurityContext = securityCtx
		expectedSpecBytes, _ := proto.Marshal(expectedSpec)
		assert.Equal(t, expectedSpecBytes, execution.Spec)
		assert.Equal(t, execution.User, principal)

		expectedCreatedAt := timestamppb.New(createdAt)
		expectedClosure, _ := proto.Marshal(&admin.ExecutionClosure{
			Phase: core.WorkflowExecution_FAILED,
			OutputResult: &admin.ExecutionClosure_Error{
				Error: &core.ExecutionError{
					Code:    execErr.Code().String(),
					Message: execErr.Error(),
					Kind:    core.ExecutionError_SYSTEM,
				},
			},
			CreatedAt:  expectedCreatedAt,
			UpdatedAt:  expectedCreatedAt,
			WorkflowId: workflowIdentifier,
			StateChangeDetails: &admin.ExecutionStateChangeDetails{
				State:      admin.ExecutionState_EXECUTION_ACTIVE,
				OccurredAt: expectedCreatedAt,
				Principal:  principal,
			},
		})
		assert.Equal(t, expectedClosure, execution.Closure)
	})
}

func TestUpdateModelState_UnknownToRunning(t *testing.T) {

	createdAt := time.Date(2018, 10, 29, 16, 0, 0, 0, time.UTC)
	createdAtProto := timestamppb.New(createdAt)
	existingClosure := admin.ExecutionClosure{
		ComputedInputs: &core.LiteralMap{
			Literals: map[string]*core.Literal{
				"foo": {},
			},
		},
		Phase:     core.WorkflowExecution_UNDEFINED,
		CreatedAt: createdAtProto,
	}
	spec := testutils.GetExecutionRequest().GetSpec()
	specBytes, _ := proto.Marshal(spec)
	existingClosureBytes, _ := proto.Marshal(&existingClosure)
	startedAt := time.Now()
	executionModel := getRunningExecutionModel(specBytes, existingClosureBytes, startedAt)

	occurredAt := time.Date(2018, 10, 29, 16, 10, 0, 0, time.UTC)
	occurredAtProto := timestamppb.New(occurredAt)
	err := UpdateExecutionModelState(context.TODO(), &executionModel, &admin.WorkflowExecutionEventRequest{
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
	startedAtProto := timestamppb.New(startedAt)
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
	spec := testutils.GetExecutionRequest().GetSpec()
	specBytes, _ := proto.Marshal(spec)
	existingClosureBytes, _ := proto.Marshal(&existingClosure)
	executionModel := getRunningExecutionModel(specBytes, existingClosureBytes, startedAt)
	duration := time.Minute
	occurredAt := startedAt.Add(duration).UTC()
	occurredAtProto := timestamppb.New(occurredAt)
	executionError := core.ExecutionError{
		Code:    ec,
		Kind:    ek,
		Message: "bar baz",
	}
	err := UpdateExecutionModelState(context.TODO(), &executionModel, &admin.WorkflowExecutionEventRequest{
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
	durationProto := durationpb.New(duration)
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
	startedAtProto := timestamppb.New(startedAt)
	existingClosure := admin.ExecutionClosure{
		ComputedInputs: &core.LiteralMap{
			Literals: map[string]*core.Literal{
				"foo": {},
			},
		},
		Phase:     core.WorkflowExecution_RUNNING,
		StartedAt: startedAtProto,
	}
	spec := testutils.GetExecutionRequest().GetSpec()
	specBytes, _ := proto.Marshal(spec)
	existingClosureBytes, _ := proto.Marshal(&existingClosure)
	executionModel := getRunningExecutionModel(specBytes, existingClosureBytes, startedAt)
	duration := time.Minute
	durationProto := durationpb.New(duration)
	occurredAt := startedAt.Add(duration).UTC()
	occurredAtProto := timestamppb.New(occurredAt)

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
		err := UpdateExecutionModelState(context.TODO(), &executionModel, &admin.WorkflowExecutionEventRequest{
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
		err := UpdateExecutionModelState(context.TODO(), &executionModel, &admin.WorkflowExecutionEventRequest{
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
		err := UpdateExecutionModelState(context.TODO(), &executionModel, &admin.WorkflowExecutionEventRequest{
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
	spec := testutils.GetExecutionRequest().GetSpec()
	specBytes, _ := proto.Marshal(spec)
	phase := core.WorkflowExecution_RUNNING.String()
	startedAt := time.Date(2018, 8, 30, 0, 0, 0, 0, time.UTC)
	createdAt := time.Date(2022, 01, 18, 0, 0, 0, 0, time.UTC)
	startedAtProto := timestamppb.New(startedAt)
	createdAtProto := timestamppb.New(createdAt)
	closure := admin.ExecutionClosure{
		ComputedInputs: spec.GetInputs(),
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
	assert.Equal(t, core.WorkflowExecution_ABORTED, execution.GetClosure().GetPhase())
	assert.True(t, proto.Equal(&admin.AbortMetadata{
		Cause: abortCause,
	}, execution.GetClosure().GetAbortMetadata()))

	executionModel.Phase = core.WorkflowExecution_RUNNING.String()
	execution, err = FromExecutionModel(context.TODO(), executionModel, DefaultExecutionTransformerOptions)
	assert.Nil(t, err)
	assert.Empty(t, execution.GetClosure().GetAbortCause())
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
	assert.Equal(t, core.WorkflowExecution_FAILED, execution.GetClosure().GetPhase())
	assert.True(t, proto.Equal(expectedExecErr, execution.GetClosure().GetError()))
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
	assert.Equal(t, execution.GetSpec().GetMetadata().GetSystemMetadata().GetNamespace(), overwrittenNamespace)
}

func TestFromExecutionModels(t *testing.T) {
	spec := testutils.GetExecutionRequest().GetSpec()
	specBytes, _ := proto.Marshal(spec)
	phase := core.WorkflowExecution_SUCCEEDED.String()
	startedAt := time.Date(2018, 8, 30, 0, 0, 0, 0, time.UTC)
	createdAt := time.Date(2022, 01, 18, 0, 0, 0, 0, time.UTC)
	startedAtProto := timestamppb.New(startedAt)
	createdAtProto := timestamppb.New(createdAt)
	duration := 2 * time.Minute
	durationProto := durationpb.New(duration)
	closure := admin.ExecutionClosure{
		ComputedInputs: spec.GetInputs(),
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
	createdAtProto := timestamppb.New(createdAt)
	existingClosure := admin.ExecutionClosure{
		ComputedInputs: &core.LiteralMap{
			Literals: map[string]*core.Literal{
				"foo": {},
			},
		},
		Phase:     core.WorkflowExecution_UNDEFINED,
		CreatedAt: createdAtProto,
	}
	spec := testutils.GetExecutionRequest().GetSpec()
	specBytes, _ := proto.Marshal(spec)
	existingClosureBytes, _ := proto.Marshal(&existingClosure)
	startedAt := time.Now()
	executionModel := getRunningExecutionModel(specBytes, existingClosureBytes, startedAt)
	testCluster := "C1"
	altCluster := "C2"
	executionModel.Cluster = testCluster
	occurredAt := time.Date(2018, 10, 29, 16, 10, 0, 0, time.UTC)
	occurredAtProto := timestamppb.New(occurredAt)
	t.Run("update", func(t *testing.T) {
		executionModel.Cluster = altCluster
		err := UpdateExecutionModelState(context.TODO(), &executionModel, &admin.WorkflowExecutionEventRequest{
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
		err := UpdateExecutionModelState(context.TODO(), &executionModel, &admin.WorkflowExecutionEventRequest{
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
		err := UpdateExecutionModelState(context.TODO(), &executionModel, &admin.WorkflowExecutionEventRequest{
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
		err := UpdateExecutionModelState(context.TODO(), &executionModel, &admin.WorkflowExecutionEventRequest{
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
		spec := testutils.GetExecutionRequest().GetSpec()
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
		assert.Equal(t, newCluster, updatedSpec.GetMetadata().GetSystemMetadata().GetExecutionCluster())
	})
	t.Run("happy case - initialize cluster", func(t *testing.T) {
		spec := testutils.GetExecutionRequest().GetSpec()
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
		assert.Equal(t, newCluster, updatedSpec.GetMetadata().GetSystemMetadata().GetExecutionCluster())
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
	createdAtProto := timestamppb.New(createdAt)

	t.Run("supporting older executions", func(t *testing.T) {
		executionModel := models.Execution{
			BaseModel: models.BaseModel{
				CreatedAt: createdAt,
			},
		}
		executionStatus, err := PopulateDefaultStateChangeDetails(executionModel)
		assert.Nil(t, err)
		assert.NotNil(t, executionStatus)
		assert.Equal(t, admin.ExecutionState_EXECUTION_ACTIVE, executionStatus.GetState())
		assert.NotNil(t, executionStatus.GetOccurredAt())
		assert.Equal(t, createdAtProto, executionStatus.GetOccurredAt())
	})
}

func TestUpdateExecutionModelStateChangeDetails(t *testing.T) {
	t.Run("empty closure", func(t *testing.T) {
		execModel := &models.Execution{}
		stateUpdatedAt := time.Now()
		statetUpdateAtProto := timestamppb.New(stateUpdatedAt)
		err := UpdateExecutionModelStateChangeDetails(execModel, admin.ExecutionState_EXECUTION_ARCHIVED,
			stateUpdatedAt, "dummyUser")
		assert.Nil(t, err)
		stateInt := int32(admin.ExecutionState_EXECUTION_ARCHIVED)
		assert.Equal(t, execModel.State, &stateInt)
		closure := &admin.ExecutionClosure{}
		err = proto.Unmarshal(execModel.Closure, closure)
		assert.Nil(t, err)
		assert.NotNil(t, closure)
		assert.NotNil(t, closure.GetStateChangeDetails())
		assert.Equal(t, admin.ExecutionState_EXECUTION_ARCHIVED, closure.GetStateChangeDetails().GetState())
		assert.Equal(t, "dummyUser", closure.GetStateChangeDetails().GetPrincipal())
		assert.Equal(t, statetUpdateAtProto, closure.GetStateChangeDetails().GetOccurredAt())

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
}

func TestTrimErrorMessage(t *testing.T) {
	errMsg := "[1/1] currentAttempt done. Last Error: USER::                   │\n│ ❱  760 │   │   │   │   return __callback(*args, **kwargs)                    │\n││"
	trimmedErrMessage := TrimErrorMessage(errMsg)
	errMsgAreValidUTF8 := utf8.Valid([]byte(trimmedErrMessage))
	assert.True(t, errMsgAreValidUTF8)
}
