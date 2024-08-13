package launchplan

import (
	"context"
	"reflect"
	"testing"
	"time"

	"github.com/golang/protobuf/proto"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	"github.com/flyteorg/flyte/flyteidl/clients/go/admin/mocks"
	"github.com/flyteorg/flyte/flyteidl/clients/go/coreutils"
	"github.com/flyteorg/flyte/flyteidl/gen/pb-go/flyteidl/admin"
	"github.com/flyteorg/flyte/flyteidl/gen/pb-go/flyteidl/core"
	"github.com/flyteorg/flyte/flytestdlib/cache"
	mocks2 "github.com/flyteorg/flyte/flytestdlib/cache/mocks"
	"github.com/flyteorg/flyte/flytestdlib/contextutils"
	"github.com/flyteorg/flyte/flytestdlib/promutils"
	"github.com/flyteorg/flyte/flytestdlib/promutils/labeled"
	"github.com/flyteorg/flyte/flytestdlib/storage"
	storageMocks "github.com/flyteorg/flyte/flytestdlib/storage/mocks"
)

func TestAdminLaunchPlanExecutor_GetStatus(t *testing.T) {
	ctx := context.TODO()
	id := &core.WorkflowExecutionIdentifier{
		Name:    "n",
		Domain:  "d",
		Project: "p",
	}
	var result *admin.ExecutionClosure

	memStore, err := storage.NewDataStore(&storage.Config{Type: storage.TypeMemory}, promutils.NewTestScope())
	assert.NoError(t, err)

	t.Run("happy", func(t *testing.T) {
		mockClient := &mocks.AdminServiceClient{}
		exec, err := NewAdminLaunchPlanExecutor(ctx, mockClient, time.Millisecond, defaultAdminConfig, promutils.NewTestScope(), memStore)
		assert.NoError(t, err)
		mockClient.On("GetExecution",
			ctx,
			mock.MatchedBy(func(o *admin.WorkflowExecutionGetRequest) bool { return true }),
		).Return(result, nil)
		assert.NoError(t, err)
		s, _, err := exec.GetStatus(ctx, id)
		assert.NoError(t, err)
		assert.Equal(t, result, s)
	})

	t.Run("notFound", func(t *testing.T) {
		mockClient := &mocks.AdminServiceClient{}

		mockClient.On("CreateExecution",
			ctx,
			mock.MatchedBy(func(o *admin.ExecutionCreateRequest) bool {
				return o.Project == "p" && o.Domain == "d" && o.Name == "n" && o.Spec.Inputs == nil
			}),
		).Return(nil, nil)

		mockClient.On("GetExecution",
			mock.Anything,
			mock.MatchedBy(func(o *admin.WorkflowExecutionGetRequest) bool { return true }),
		).Return(nil, status.Error(codes.NotFound, ""))

		exec, err := NewAdminLaunchPlanExecutor(ctx, mockClient, time.Millisecond, defaultAdminConfig, promutils.NewTestScope(), memStore)
		assert.NoError(t, err)

		assert.NoError(t, exec.Initialize(ctx))

		err = exec.Launch(ctx,
			LaunchContext{
				ParentNodeExecution: &core.NodeExecutionIdentifier{
					NodeId: "node-id",
					ExecutionId: &core.WorkflowExecutionIdentifier{
						Project: "p",
						Domain:  "d",
						Name:    "w",
					},
				},
			},
			id,
			&core.Identifier{},
			nil,
		)
		assert.NoError(t, err)

		// Allow for sync to be called
		time.Sleep(time.Second)

		s, _, err := exec.GetStatus(ctx, id)
		assert.Error(t, err)
		assert.Nil(t, s)
		assert.True(t, IsNotFound(err))
	})

	t.Run("other", func(t *testing.T) {
		mockClient := &mocks.AdminServiceClient{}

		mockClient.On("CreateExecution",
			ctx,
			mock.MatchedBy(func(o *admin.ExecutionCreateRequest) bool {
				return o.Project == "p" && o.Domain == "d" && o.Name == "n" && o.Spec.Inputs == nil
			}),
		).Return(nil, nil)

		mockClient.On("GetExecution",
			mock.Anything,
			mock.MatchedBy(func(o *admin.WorkflowExecutionGetRequest) bool { return true }),
		).Return(nil, status.Error(codes.Canceled, ""))

		exec, err := NewAdminLaunchPlanExecutor(ctx, mockClient, time.Millisecond, defaultAdminConfig, promutils.NewTestScope(), memStore)
		assert.NoError(t, err)

		assert.NoError(t, exec.Initialize(ctx))

		err = exec.Launch(ctx,
			LaunchContext{
				ParentNodeExecution: &core.NodeExecutionIdentifier{
					NodeId: "node-id",
					ExecutionId: &core.WorkflowExecutionIdentifier{
						Project: "p",
						Domain:  "d",
						Name:    "w",
					},
				},
			},
			id,
			&core.Identifier{},
			nil,
		)
		assert.NoError(t, err)

		// Allow for sync to be called
		time.Sleep(time.Second)

		s, _, err := exec.GetStatus(ctx, id)
		assert.Error(t, err)
		assert.Nil(t, s)
		assert.False(t, IsNotFound(err))
	})

}

func TestAdminLaunchPlanExecutor_Launch(t *testing.T) {
	ctx := context.TODO()
	id := &core.WorkflowExecutionIdentifier{
		Name:    "n",
		Domain:  "d",
		Project: "p",
	}
	memStore, err := storage.NewDataStore(&storage.Config{Type: storage.TypeMemory}, promutils.NewTestScope())
	assert.NoError(t, err)

	t.Run("happy", func(t *testing.T) {

		mockClient := &mocks.AdminServiceClient{}
		exec, err := NewAdminLaunchPlanExecutor(ctx, mockClient, time.Second, defaultAdminConfig, promutils.NewTestScope(), memStore)
		mockClient.On("CreateExecution",
			ctx,
			mock.MatchedBy(func(o *admin.ExecutionCreateRequest) bool {
				return o.Project == "p" && o.Domain == "d" && o.Name == "n" && o.Spec.Inputs == nil &&
					o.Spec.Metadata.Mode == admin.ExecutionMetadata_CHILD_WORKFLOW &&
					reflect.DeepEqual(o.Spec.Labels.Values, map[string]string{"foo": "bar"}) // Ensure shard-key was removed.
			}),
		).Return(nil, nil)
		assert.NoError(t, err)

		var labels = map[string]string{"foo": "bar", "shard-key": "1"}

		err = exec.Launch(ctx,
			LaunchContext{
				ParentNodeExecution: &core.NodeExecutionIdentifier{
					NodeId: "node-id",
					ExecutionId: &core.WorkflowExecutionIdentifier{
						Project: "p",
						Domain:  "d",
						Name:    "w",
					},
				},
				Labels: labels,
			},
			id,
			&core.Identifier{},
			nil,
		)
		assert.NoError(t, err)
		// Ensure we haven't mutated the state of the parent workflow.
		assert.True(t, reflect.DeepEqual(labels, map[string]string{"foo": "bar", "shard-key": "1"}))
	})

	t.Run("happy recover", func(t *testing.T) {

		mockClient := &mocks.AdminServiceClient{}
		parentNodeExecution := &core.NodeExecutionIdentifier{
			NodeId: "node-id",
			ExecutionId: &core.WorkflowExecutionIdentifier{
				Project: "p",
				Domain:  "d",
				Name:    "orig",
			},
		}
		exec, err := NewAdminLaunchPlanExecutor(ctx, mockClient, time.Second, defaultAdminConfig, promutils.NewTestScope(), memStore)
		mockClient.On("RecoverExecution",
			ctx,
			mock.MatchedBy(func(o *admin.ExecutionRecoverRequest) bool {
				return o.Id.Project == "p" && o.Id.Domain == "d" && o.Id.Name == "w" && o.Name == "n" &&
					proto.Equal(o.Metadata.ParentNodeExecution, parentNodeExecution)
			}),
		).Return(nil, nil)
		assert.NoError(t, err)
		err = exec.Launch(ctx,
			LaunchContext{
				RecoveryExecution: &core.WorkflowExecutionIdentifier{
					Project: "p",
					Domain:  "d",
					Name:    "w",
				},
				ParentNodeExecution: parentNodeExecution,
			},
			id,
			&core.Identifier{},
			nil,
		)
		assert.NoError(t, err)
	})

	t.Run("recovery fails", func(t *testing.T) {

		mockClient := &mocks.AdminServiceClient{}
		parentNodeExecution := &core.NodeExecutionIdentifier{
			NodeId: "node-id",
			ExecutionId: &core.WorkflowExecutionIdentifier{
				Project: "p",
				Domain:  "d",
				Name:    "orig",
			},
		}
		exec, err := NewAdminLaunchPlanExecutor(ctx, mockClient, time.Second, defaultAdminConfig, promutils.NewTestScope(), memStore)
		assert.NoError(t, err)

		recoveryErr := status.Error(codes.NotFound, "foo")
		mockClient.On("RecoverExecution",
			ctx,
			mock.MatchedBy(func(o *admin.ExecutionRecoverRequest) bool {
				return o.Id.Project == "p" && o.Id.Domain == "d" && o.Id.Name == "w" && o.Name == "n" &&
					proto.Equal(o.Metadata.ParentNodeExecution, parentNodeExecution)
			}),
		).Return(nil, recoveryErr)

		var createCalled = false
		mockClient.On("CreateExecution",
			ctx,
			mock.MatchedBy(func(o *admin.ExecutionCreateRequest) bool {
				createCalled = true
				return o.Project == "p" && o.Domain == "d" && o.Name == "n" && o.Spec.Inputs == nil &&
					o.Spec.Metadata.Mode == admin.ExecutionMetadata_CHILD_WORKFLOW
			}),
		).Return(nil, nil)

		err = exec.Launch(ctx,
			LaunchContext{
				RecoveryExecution: &core.WorkflowExecutionIdentifier{
					Project: "p",
					Domain:  "d",
					Name:    "w",
				},
				ParentNodeExecution: parentNodeExecution,
			},
			id,
			&core.Identifier{},
			nil,
		)
		assert.NoError(t, err)
		assert.True(t, createCalled)
	})

	t.Run("notFound", func(t *testing.T) {

		mockClient := &mocks.AdminServiceClient{}
		exec, err := NewAdminLaunchPlanExecutor(ctx, mockClient, time.Second, defaultAdminConfig, promutils.NewTestScope(), memStore)
		mockClient.On("CreateExecution",
			ctx,
			mock.MatchedBy(func(o *admin.ExecutionCreateRequest) bool { return true }),
		).Return(nil, status.Error(codes.AlreadyExists, ""))
		assert.NoError(t, err)
		err = exec.Launch(ctx,
			LaunchContext{
				ParentNodeExecution: &core.NodeExecutionIdentifier{
					NodeId: "node-id",
					ExecutionId: &core.WorkflowExecutionIdentifier{
						Project: "p",
						Domain:  "d",
						Name:    "w",
					},
				},
			},
			id,
			&core.Identifier{},
			nil,
		)
		assert.Error(t, err)
		assert.True(t, IsAlreadyExists(err))
	})

	t.Run("other", func(t *testing.T) {

		mockClient := &mocks.AdminServiceClient{}
		exec, err := NewAdminLaunchPlanExecutor(ctx, mockClient, time.Second, defaultAdminConfig, promutils.NewTestScope(), memStore)
		mockClient.On("CreateExecution",
			ctx,
			mock.MatchedBy(func(o *admin.ExecutionCreateRequest) bool { return true }),
		).Return(nil, status.Error(codes.Canceled, ""))
		assert.NoError(t, err)
		err = exec.Launch(ctx,
			LaunchContext{
				ParentNodeExecution: &core.NodeExecutionIdentifier{
					NodeId: "node-id",
					ExecutionId: &core.WorkflowExecutionIdentifier{
						Project: "p",
						Domain:  "d",
						Name:    "w",
					},
				},
			},
			id,
			&core.Identifier{},
			nil,
		)
		assert.Error(t, err)
		assert.False(t, IsAlreadyExists(err))
	})
}

func TestAdminLaunchPlanExecutor_Kill(t *testing.T) {
	ctx := context.TODO()
	id := &core.WorkflowExecutionIdentifier{
		Name:    "n",
		Domain:  "d",
		Project: "p",
	}
	memStore, err := storage.NewDataStore(&storage.Config{Type: storage.TypeMemory}, promutils.NewTestScope())
	assert.NoError(t, err)

	const reason = "reason"
	t.Run("happy", func(t *testing.T) {

		mockClient := &mocks.AdminServiceClient{}
		exec, err := NewAdminLaunchPlanExecutor(ctx, mockClient, time.Second, defaultAdminConfig, promutils.NewTestScope(), memStore)
		mockClient.On("TerminateExecution",
			ctx,
			mock.MatchedBy(func(o *admin.ExecutionTerminateRequest) bool { return o.Id == id && o.Cause == reason }),
		).Return(&admin.ExecutionTerminateResponse{}, nil)
		assert.NoError(t, err)
		err = exec.Kill(ctx, id, reason)
		assert.NoError(t, err)
	})

	t.Run("notFound", func(t *testing.T) {

		mockClient := &mocks.AdminServiceClient{}
		exec, err := NewAdminLaunchPlanExecutor(ctx, mockClient, time.Second, defaultAdminConfig, promutils.NewTestScope(), memStore)
		mockClient.On("TerminateExecution",
			ctx,
			mock.MatchedBy(func(o *admin.ExecutionTerminateRequest) bool { return o.Id == id && o.Cause == reason }),
		).Return(nil, status.Error(codes.NotFound, ""))
		assert.NoError(t, err)
		err = exec.Kill(ctx, id, reason)
		assert.NoError(t, err)
	})

	t.Run("other", func(t *testing.T) {

		mockClient := &mocks.AdminServiceClient{}
		exec, err := NewAdminLaunchPlanExecutor(ctx, mockClient, time.Second, defaultAdminConfig, promutils.NewTestScope(), memStore)
		mockClient.On("TerminateExecution",
			ctx,
			mock.MatchedBy(func(o *admin.ExecutionTerminateRequest) bool { return o.Id == id && o.Cause == reason }),
		).Return(nil, status.Error(codes.Canceled, ""))
		assert.NoError(t, err)
		err = exec.Kill(ctx, id, reason)
		assert.Error(t, err)
		assert.False(t, IsNotFound(err))
	})
}

func TestNewAdminLaunchPlanExecutor_GetLaunchPlan(t *testing.T) {
	ctx := context.TODO()
	id := &core.Identifier{
		ResourceType: core.ResourceType_LAUNCH_PLAN,
		Name:         "n",
		Domain:       "d",
		Project:      "p",
		Version:      "v",
	}
	memStore, err := storage.NewDataStore(&storage.Config{Type: storage.TypeMemory}, promutils.NewTestScope())
	assert.NoError(t, err)

	t.Run("launch plan found", func(t *testing.T) {
		mockClient := &mocks.AdminServiceClient{}
		exec, err := NewAdminLaunchPlanExecutor(ctx, mockClient, time.Second, defaultAdminConfig, promutils.NewTestScope(), memStore)
		assert.NoError(t, err)
		mockClient.OnGetLaunchPlanMatch(
			ctx,
			mock.MatchedBy(func(o *admin.ObjectGetRequest) bool { return true }),
		).Return(&admin.LaunchPlan{Id: id}, nil)
		lp, err := exec.GetLaunchPlan(ctx, id)
		assert.NoError(t, err)
		assert.Equal(t, lp.Id, id)
	})

	t.Run("launch plan not found", func(t *testing.T) {
		mockClient := &mocks.AdminServiceClient{}
		exec, err := NewAdminLaunchPlanExecutor(ctx, mockClient, time.Second, defaultAdminConfig, promutils.NewTestScope(), memStore)
		assert.NoError(t, err)
		mockClient.OnGetLaunchPlanMatch(
			ctx,
			mock.MatchedBy(func(o *admin.ObjectGetRequest) bool { return true }),
		).Return(nil, status.Error(codes.NotFound, ""))
		lp, err := exec.GetLaunchPlan(ctx, id)
		assert.Nil(t, lp)
		assert.Error(t, err)
	})
}

type test struct {
	name                  string
	cacheItem             executionCacheItem
	getExecutionResp      *admin.Execution
	getExecutionError     error
	getExecutionDataResp  *admin.WorkflowExecutionGetDataResponse
	getExecutionDataError error
	storageReadError      error
	expectSuccess         bool
	expectError           bool
	expectedOutputs       *core.LiteralMap
	expectedErrorContains string
}

func TestAdminLaunchPlanExecutorScenarios(t *testing.T) {
	ctx := context.TODO()

	mockExecutionRespWithOutputs := &admin.Execution{
		Closure: &admin.ExecutionClosure{
			Phase: core.WorkflowExecution_SUCCEEDED,
			OutputResult: &admin.ExecutionClosure_Outputs{
				Outputs: &admin.LiteralMapBlob{
					Data: &admin.LiteralMapBlob_Uri{
						Uri: "s3://foo/bar",
					},
				},
			},
		},
	}
	mockExecutionRespWithoutOutputs := &admin.Execution{
		Closure: &admin.ExecutionClosure{
			Phase: core.WorkflowExecution_SUCCEEDED,
		},
	}
	outputLiteral := &core.LiteralMap{
		Literals: map[string]*core.Literal{
			"foo-1": coreutils.MustMakeLiteral("foo-value-1"),
		},
	}

	tests := []test{
		{
			name:          "terminal-sync",
			expectSuccess: true,
			cacheItem: executionCacheItem{
				ExecutionClosure: &admin.ExecutionClosure{
					Phase:      core.WorkflowExecution_SUCCEEDED,
					WorkflowId: &core.Identifier{Project: "p"}},
				ExecutionOutputs: outputLiteral,
			},
			expectedOutputs: outputLiteral,
		},
		{
			name:                  "GetExecution-fails",
			expectError:           true,
			cacheItem:             executionCacheItem{},
			getExecutionError:     status.Error(codes.NotFound, ""),
			expectedErrorContains: RemoteErrorNotFound,
		},
		{
			name:                  "GetExecution-fails-system",
			expectError:           true,
			cacheItem:             executionCacheItem{},
			getExecutionError:     status.Error(codes.Internal, ""),
			expectedErrorContains: RemoteErrorSystem,
		},
		{
			name:            "GetExecutionData-succeeds",
			expectSuccess:   true,
			cacheItem:       executionCacheItem{},
			expectedOutputs: outputLiteral,
			getExecutionDataResp: &admin.WorkflowExecutionGetDataResponse{
				FullOutputs: outputLiteral,
			},
			getExecutionDataError: nil,
			getExecutionResp:      mockExecutionRespWithOutputs,
		},
		{
			name:                  "GetExecutionData-error-no-retry",
			expectError:           true,
			cacheItem:             executionCacheItem{},
			getExecutionDataError: status.Error(codes.NotFound, ""),
			expectedErrorContains: codes.NotFound.String(),
			getExecutionResp:      mockExecutionRespWithoutOutputs,
		},
		{
			name:                  "GetExecutionData-error-retry-fails",
			expectError:           true,
			cacheItem:             executionCacheItem{},
			getExecutionDataError: status.Error(codes.NotFound, ""),
			storageReadError:      status.Error(codes.Internal, ""),
			expectedErrorContains: codes.Internal.String(),
			getExecutionResp:      mockExecutionRespWithOutputs,
		},
		{
			name:        "GetExecutionData-empty-retry-fails",
			expectError: true,
			cacheItem:   executionCacheItem{},
			getExecutionDataResp: &admin.WorkflowExecutionGetDataResponse{
				FullOutputs: &core.LiteralMap{},
			},
			getExecutionDataError: nil,
			storageReadError:      status.Error(codes.Internal, ""),
			expectedErrorContains: codes.Internal.String(),
			getExecutionResp:      mockExecutionRespWithOutputs,
		},
		{
			name:          "GetExecutionData-empty-retry-succeeds",
			expectSuccess: true,
			cacheItem:     executionCacheItem{},
			getExecutionDataResp: &admin.WorkflowExecutionGetDataResponse{
				FullOutputs: &core.LiteralMap{},
			},
			getExecutionDataError: nil,
			expectedOutputs:       &core.LiteralMap{},
			getExecutionResp:      mockExecutionRespWithOutputs,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			mockClient := &mocks.AdminServiceClient{}
			pbStore := &storageMocks.ComposedProtobufStore{}
			pbStore.On("ReadProtobuf", mock.Anything, mock.Anything, mock.Anything).Return(tc.storageReadError)
			storageClient := &storage.DataStore{
				ComposedProtobufStore: pbStore,
				ReferenceConstructor:  &storageMocks.ReferenceConstructor{},
			}
			exec, err := NewAdminLaunchPlanExecutor(ctx, mockClient, time.Millisecond, defaultAdminConfig, promutils.NewTestScope(), storageClient)
			assert.NoError(t, err)

			iwMock := &mocks2.ItemWrapper{}
			i := tc.cacheItem
			iwMock.OnGetItem().Return(i)
			iwMock.OnGetID().Return("id")

			mockClient.On("GetExecution", mock.Anything, mock.Anything).Return(tc.getExecutionResp, tc.getExecutionError)
			mockClient.On("GetExecutionData", mock.Anything, mock.Anything).Return(tc.getExecutionDataResp, tc.getExecutionDataError)

			adminExec := exec.(*adminLaunchPlanExecutor)

			v, err := adminExec.syncItem(ctx, cache.Batch{iwMock})
			assert.NoError(t, err)
			assert.Len(t, v, 1)
			item, ok := v[0].Item.(executionCacheItem)
			assert.True(t, ok)

			if tc.expectSuccess {
				assert.Nil(t, item.SyncError)
				assert.True(t, proto.Equal(tc.expectedOutputs, item.ExecutionOutputs))
			}
			if tc.expectError {
				assert.NotNil(t, item.SyncError)
				assert.Contains(t, item.SyncError.Error(), tc.expectedErrorContains)
			}
		})
	}
}

func TestIsWorkflowTerminated(t *testing.T) {
	assert.True(t, IsWorkflowTerminated(core.WorkflowExecution_SUCCEEDED))
	assert.True(t, IsWorkflowTerminated(core.WorkflowExecution_ABORTED))
	assert.True(t, IsWorkflowTerminated(core.WorkflowExecution_FAILED))
	assert.True(t, IsWorkflowTerminated(core.WorkflowExecution_TIMED_OUT))
	assert.False(t, IsWorkflowTerminated(core.WorkflowExecution_SUCCEEDING))
	assert.False(t, IsWorkflowTerminated(core.WorkflowExecution_FAILING))
	assert.False(t, IsWorkflowTerminated(core.WorkflowExecution_RUNNING))
	assert.False(t, IsWorkflowTerminated(core.WorkflowExecution_QUEUED))
	assert.False(t, IsWorkflowTerminated(core.WorkflowExecution_UNDEFINED))
}

func init() {
	labeled.SetMetricKeys(contextutils.ProjectKey, contextutils.DomainKey, contextutils.WorkflowIDKey, contextutils.TaskIDKey)
}
