package launchplan

import (
	"context"
	"fmt"
	"reflect"
	"testing"
	"time"

	"github.com/golang/protobuf/proto"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	"github.com/flyteorg/flyte/flyteidl/clients/go/admin/mocks"
	"github.com/flyteorg/flyte/flyteidl/gen/pb-go/flyteidl/admin"
	"github.com/flyteorg/flyte/flyteidl/gen/pb-go/flyteidl/core"
	"github.com/flyteorg/flyte/flyteidl/gen/pb-go/flyteidl/watch"
	ctrlConfig "github.com/flyteorg/flyte/flytepropeller/pkg/controller/config"
	"github.com/flyteorg/flyte/flytestdlib/cache"
	cacheMocks "github.com/flyteorg/flyte/flytestdlib/cache/mocks"
	"github.com/flyteorg/flyte/flytestdlib/config"
	"github.com/flyteorg/flyte/flytestdlib/contextutils"
	"github.com/flyteorg/flyte/flytestdlib/errors"
	"github.com/flyteorg/flyte/flytestdlib/promutils"
	"github.com/flyteorg/flyte/flytestdlib/promutils/labeled"
	"github.com/flyteorg/flyte/flytestdlib/storage"
	storageMocks "github.com/flyteorg/flyte/flytestdlib/storage/mocks"
)

var (
	launchPlanWithOutputs = &core.LaunchPlanTemplate{
		Id: &core.Identifier{},
		Interface: &core.TypedInterface{
			Inputs: &core.VariableMap{
				Variables: map[string]*core.Variable{
					"foo": {
						Type: &core.LiteralType{Type: &core.LiteralType_Simple{Simple: core.SimpleType_STRING}},
					},
				},
			},
			Outputs: &core.VariableMap{
				Variables: map[string]*core.Variable{
					"bar": {
						Type: &core.LiteralType{Type: &core.LiteralType_Simple{Simple: core.SimpleType_STRING}},
					},
				},
			},
		},
	}
	parentWorkflowID = "parentwf"
	execID           = &core.WorkflowExecutionIdentifier{
		Name:    "n",
		Domain:  "d",
		Project: "p",
		Org:     "o",
	}
	testURI = "s3://bla"
)

func TestAdminLaunchPlanExecutor_GetStatus(t *testing.T) {
	ctx := context.TODO()
	adminConfig := GetAdminConfig()
	adminConfig.Workers = 1
	adminConfig.WatchConfig.Enabled = false

	id := &core.WorkflowExecutionIdentifier{
		Name:    "n",
		Domain:  "d",
		Project: "p",
		Org:     "o",
	}

	memStore, err := storage.NewDataStore(&storage.Config{Type: storage.TypeMemory}, promutils.NewTestScope())
	assert.NoError(t, err)

	t.Run("fetch SUCCEEDED and execution outputs", func(t *testing.T) {
		mockClient := &mocks.AdminServiceClient{}
		mockClient.Test(t)
		defer mockClient.AssertExpectations(t)
		exec, err := NewAdminLaunchPlanExecutor(ctx, ctrlConfig.GetConfig(), mockClient, nil, adminConfig, promutils.NewTestScope(), memStore, func(string) {})
		assert.NoError(t, err)

		initCtx, done := context.WithCancel(ctx)
		require.NoError(t, exec.Initialize(initCtx))

		execution := &admin.Execution{
			Closure: &admin.ExecutionClosure{
				Phase: core.WorkflowExecution_SUCCEEDED,
				OutputResult: &admin.ExecutionClosure_Outputs{
					Outputs: &admin.LiteralMapBlob{
						Data: &admin.LiteralMapBlob_Values{
							Values: &core.LiteralMap{
								Literals: map[string]*core.Literal{"foo": nil},
							},
						},
					},
				},
			},
		}
		mockClient.
			OnGetExecutionMatch(mock.Anything, &admin.WorkflowExecutionGetRequest{Id: id}).
			Run(func(args mock.Arguments) {
				done()
			}).
			Return(execution, nil).
			Once()
		execData := &admin.WorkflowExecutionGetDataResponse{
			FullOutputs: &core.LiteralMap{
				Literals: map[string]*core.Literal{"foo": nil},
			},
		}
		mockClient.
			OnGetExecutionDataMatch(mock.Anything, &admin.WorkflowExecutionGetDataRequest{Id: id}).
			Return(execData, nil).
			Once()

		assert.NoError(t, err)
		st, err := exec.GetStatus(
			ctx,
			id,
			launchPlanWithOutputs,
			parentWorkflowID,
		)
		assert.NoError(t, err)

		// return immediately default values
		assert.Equal(t, core.WorkflowExecution_UNDEFINED.String(), st.Phase.String())
		assert.Nil(t, st.Outputs)

		// Allow for sync to be called
		assert.Eventually(t, func() bool {
			st, err := exec.GetStatus(
				ctx,
				id,
				launchPlanWithOutputs,
				parentWorkflowID,
			)
			assert.NoError(t, err)
			return execution.Closure.Phase == st.Phase
		}, time.Second, time.Millisecond, "timeout waiting for GetExecution to return expected phase")

		st, err = exec.GetStatus(
			ctx,
			id,
			launchPlanWithOutputs,
			parentWorkflowID,
		)
		assert.NoError(t, err)

		assert.Equal(t, execution.Closure.Phase.String(), st.Phase.String())
		assert.Equal(t, execData.FullOutputs, st.Outputs)

		item, err := exec.(*adminLaunchPlanExecutor).cache.Get(id.String())
		assert.NoError(t, err)
		require.IsType(t, executionCacheItem{}, item)
		e := item.(executionCacheItem)
		assert.True(t, e.HasOutputs)
		assert.Equal(t, parentWorkflowID, e.ParentWorkflowID)
		assert.Equal(t, execution.Closure.Phase.String(), e.Phase.String())
		assert.Equal(t, execData.FullOutputs, e.ExecutionOutputs)
	})

	t.Run("GetExecution returns NotFound asynchronously", func(t *testing.T) {
		mockClient := &mocks.AdminServiceClient{}
		mockClient.Test(t)
		defer mockClient.AssertExpectations(t)

		exec, err := NewAdminLaunchPlanExecutor(ctx, ctrlConfig.GetConfig(), mockClient, nil, adminConfig, promutils.NewTestScope(), memStore, func(string) {})
		assert.NoError(t, err)

		initCtx, done := context.WithCancel(ctx)
		assert.NoError(t, exec.Initialize(initCtx))

		mockClient.
			OnCreateExecutionMatch(
				ctx,
				mock.MatchedBy(func(o *admin.ExecutionCreateRequest) bool {
					return o.Project == id.Project && o.Domain == id.Domain && o.Name == id.Name && o.Org == id.Org && o.Spec.Inputs == nil
				}),
			).
			Return(nil, nil).
			Once()
		mockClient.
			OnGetExecutionMatch(mock.Anything, &admin.WorkflowExecutionGetRequest{Id: id}).
			Run(func(args mock.Arguments) {
				done()
			}).
			Return(nil, status.Error(codes.NotFound, "")).
			Once()

		err = exec.Launch(ctx,
			LaunchContext{
				ParentNodeExecution: &core.NodeExecutionIdentifier{
					NodeId:      "node-id",
					ExecutionId: id,
				},
			},
			id,
			launchPlanWithOutputs,
			nil,
			parentWorkflowID,
		)
		assert.NoError(t, err)

		// Allow for sync to be called
		assert.Eventually(t, func() bool {
			_, err := exec.GetStatus(ctx, id, launchPlanWithOutputs, parentWorkflowID)
			return IsNotFound(err)
		}, time.Second, time.Millisecond, "timeout waiting for GetExecution to be called")

		st, err := exec.GetStatus(ctx, id, launchPlanWithOutputs, parentWorkflowID)
		assert.Error(t, err)
		assert.Empty(t, st)
		assert.True(t, IsNotFound(err))
	})

	t.Run("GetExecution returns unknown error", func(t *testing.T) {
		mockClient := &mocks.AdminServiceClient{}
		mockClient.Test(t)
		defer mockClient.AssertExpectations(t)

		exec, err := NewAdminLaunchPlanExecutor(ctx, ctrlConfig.GetConfig(), mockClient, nil, adminConfig, promutils.NewTestScope(), memStore, func(string) {})
		assert.NoError(t, err)

		initCtx, done := context.WithCancel(ctx)
		assert.NoError(t, exec.Initialize(initCtx))

		mockClient.
			OnCreateExecutionMatch(
				ctx,
				mock.MatchedBy(func(o *admin.ExecutionCreateRequest) bool {
					return o.Project == id.Project && o.Domain == id.Domain && o.Name == id.Name && o.Org == id.Org && o.Spec.Inputs == nil
				}),
			).
			Return(nil, nil).
			Once()
		mockClient.
			OnGetExecutionMatch(mock.Anything, &admin.WorkflowExecutionGetRequest{Id: id}).
			Run(func(args mock.Arguments) {
				done()
			}).
			Return(nil, status.Error(codes.Canceled, "")).
			Once()

		err = exec.Launch(ctx,
			LaunchContext{
				ParentNodeExecution: &core.NodeExecutionIdentifier{
					NodeId:      "node-id",
					ExecutionId: id,
				},
			},
			id,
			launchPlanWithOutputs,
			nil,
			parentWorkflowID,
		)
		assert.NoError(t, err)

		// Allow for sync to be called
		assert.Eventually(t, func() bool {
			_, err := exec.GetStatus(ctx, id, launchPlanWithOutputs, parentWorkflowID)
			stCode, ok := status.FromError(err)
			return ok && stCode.Code() == codes.Canceled
		}, time.Second, time.Millisecond, "timeout waiting for GetExecution to be called")

		st, err := exec.GetStatus(ctx, id, launchPlanWithOutputs, parentWorkflowID)
		assert.Error(t, err)
		assert.Empty(t, st)
		stCode, ok := status.FromError(err)
		assert.True(t, ok)
		assert.Equal(t, codes.Canceled, stCode.Code())
	})
}

func TestAdminLaunchPlanExecutor_Launch(t *testing.T) {
	ctx := context.TODO()
	adminConfig := GetAdminConfig()
	adminConfig.Workers = 1
	adminConfig.WatchConfig.Enabled = false
	id := &core.WorkflowExecutionIdentifier{
		Name:    "n",
		Domain:  "d",
		Project: "p",
		Org:     "o",
	}
	memStore, err := storage.NewDataStore(&storage.Config{Type: storage.TypeMemory}, promutils.NewTestScope())
	assert.NoError(t, err)

	t.Run("happy", func(t *testing.T) {
		mockClient := &mocks.AdminServiceClient{}
		mockClient.Test(t)
		defer mockClient.AssertExpectations(t)

		mockClient.
			OnCreateExecutionMatch(
				ctx,
				mock.MatchedBy(func(o *admin.ExecutionCreateRequest) bool {
					return o.Project == id.Project && o.Domain == id.Domain && o.Name == id.Name && o.Org == id.Org && o.Spec.Inputs == nil &&
						o.Spec.Metadata.Mode == admin.ExecutionMetadata_CHILD_WORKFLOW &&
						reflect.DeepEqual(o.Spec.Labels.Values, map[string]string{
							"foo":            "bar",
							"parent-cluster": "propeller",
							"parent-shard":   "1",
						}) // Ensure shard-key was removed.
				}),
			).
			Return(nil, nil).
			Once()

		exec, err := NewAdminLaunchPlanExecutor(ctx, ctrlConfig.GetConfig(), mockClient, nil, adminConfig, promutils.NewTestScope(), memStore, func(string) {})
		assert.NoError(t, err)

		var labels = map[string]string{"foo": "bar", "shard-key": "1"}

		err = exec.Launch(ctx,
			LaunchContext{
				ParentNodeExecution: &core.NodeExecutionIdentifier{
					NodeId:      "node-id",
					ExecutionId: id,
				},
				Labels: labels,
			},
			id,
			launchPlanWithOutputs,
			nil,
			parentWorkflowID,
		)
		assert.NoError(t, err)
		// Ensure we haven't mutated the state of the parent workflow.
		assert.True(t, reflect.DeepEqual(labels, map[string]string{"foo": "bar", "shard-key": "1"}))
	})

	t.Run("happy recover", func(t *testing.T) {
		mockClient := &mocks.AdminServiceClient{}
		mockClient.Test(t)
		defer mockClient.AssertExpectations(t)

		parentNodeExecution := &core.NodeExecutionIdentifier{
			NodeId: "node-id",
			ExecutionId: &core.WorkflowExecutionIdentifier{
				Project: "p",
				Domain:  "d",
				Name:    "orig",
			},
		}
		exec, err := NewAdminLaunchPlanExecutor(ctx, ctrlConfig.GetConfig(), mockClient, nil, adminConfig, promutils.NewTestScope(), memStore, func(string) {})
		mockClient.
			OnRecoverExecutionMatch(
				ctx,
				mock.MatchedBy(func(o *admin.ExecutionRecoverRequest) bool {
					return o.Id.Project == "p" && o.Id.Domain == "d" && o.Id.Name == "w" && o.Id.Org == "o" && o.Name == "n" &&
						proto.Equal(o.Metadata.ParentNodeExecution, parentNodeExecution)
				}),
			).
			Return(nil, nil).
			Once()
		assert.NoError(t, err)
		err = exec.Launch(ctx,
			LaunchContext{
				RecoveryExecution: &core.WorkflowExecutionIdentifier{
					Project: "p",
					Domain:  "d",
					Name:    "w",
					Org:     "o",
				},
				ParentNodeExecution: parentNodeExecution,
			},
			id,
			launchPlanWithOutputs,
			nil,
			parentWorkflowID,
		)
		assert.NoError(t, err)
	})

	t.Run("recovery fails", func(t *testing.T) {
		mockClient := &mocks.AdminServiceClient{}
		mockClient.Test(t)
		defer mockClient.AssertExpectations(t)

		parentNodeExecution := &core.NodeExecutionIdentifier{
			NodeId: "node-id",
			ExecutionId: &core.WorkflowExecutionIdentifier{
				Project: "p",
				Domain:  "d",
				Name:    "orig",
			},
		}
		exec, err := NewAdminLaunchPlanExecutor(ctx, ctrlConfig.GetConfig(), mockClient, nil, adminConfig, promutils.NewTestScope(), memStore, func(string) {})
		assert.NoError(t, err)

		recoveryErr := status.Error(codes.NotFound, "foo")
		mockClient.
			OnRecoverExecutionMatch(
				ctx,
				mock.MatchedBy(func(o *admin.ExecutionRecoverRequest) bool {
					return o.Id.Project == "p" && o.Id.Domain == "d" && o.Id.Name == "w" && o.Id.Org == "o" && o.Name == "n" &&
						proto.Equal(o.Metadata.ParentNodeExecution, parentNodeExecution)
				}),
			).
			Return(nil, recoveryErr).
			Once()

		var createCalled = false
		mockClient.
			OnCreateExecutionMatch(
				ctx,
				mock.MatchedBy(func(o *admin.ExecutionCreateRequest) bool {
					createCalled = true
					return o.Project == "p" && o.Domain == "d" && o.Name == "n" && o.Org == "o" && o.Spec.Inputs == nil &&
						o.Spec.Metadata.Mode == admin.ExecutionMetadata_CHILD_WORKFLOW
				}),
			).
			Return(nil, nil).
			Once()

		err = exec.Launch(ctx,
			LaunchContext{
				RecoveryExecution: &core.WorkflowExecutionIdentifier{
					Project: "p",
					Domain:  "d",
					Name:    "w",
					Org:     "o",
				},
				ParentNodeExecution: parentNodeExecution,
			},
			id,
			launchPlanWithOutputs,
			nil,
			parentWorkflowID,
		)
		assert.NoError(t, err)
		assert.True(t, createCalled)
	})

	t.Run("notFound", func(t *testing.T) {
		mockClient := &mocks.AdminServiceClient{}
		mockClient.Test(t)
		defer mockClient.AssertExpectations(t)

		exec, err := NewAdminLaunchPlanExecutor(ctx, ctrlConfig.GetConfig(), mockClient, nil, adminConfig, promutils.NewTestScope(), memStore, func(string) {})
		mockClient.
			OnCreateExecutionMatch(
				ctx,
				mock.MatchedBy(func(o *admin.ExecutionCreateRequest) bool { return true }),
			).
			Return(nil, status.Error(codes.AlreadyExists, "")).
			Once()
		assert.NoError(t, err)
		err = exec.Launch(ctx,
			LaunchContext{
				ParentNodeExecution: &core.NodeExecutionIdentifier{
					NodeId: "node-id",
					ExecutionId: &core.WorkflowExecutionIdentifier{
						Project: "p",
						Domain:  "d",
						Name:    "w",
						Org:     "o",
					},
				},
			},
			id,
			launchPlanWithOutputs,
			nil,
			parentWorkflowID,
		)
		assert.Error(t, err)
		assert.True(t, IsAlreadyExists(err))
	})

	t.Run("other", func(t *testing.T) {
		mockClient := &mocks.AdminServiceClient{}
		mockClient.Test(t)
		defer mockClient.AssertExpectations(t)

		exec, err := NewAdminLaunchPlanExecutor(ctx, ctrlConfig.GetConfig(), mockClient, nil, adminConfig, promutils.NewTestScope(), memStore, func(string) {})
		mockClient.
			OnCreateExecutionMatch(
				ctx,
				mock.MatchedBy(func(o *admin.ExecutionCreateRequest) bool { return true }),
			).
			Return(nil, status.Error(codes.Canceled, "")).
			Once()
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
			launchPlanWithOutputs,
			nil,
			parentWorkflowID,
		)
		assert.Error(t, err)
		assert.False(t, IsAlreadyExists(err))
	})
}

func TestAdminLaunchPlanExecutor_Kill(t *testing.T) {
	ctx := context.TODO()

	adminConfig := GetAdminConfig()
	adminConfig.CacheResyncDuration = config.Duration{Duration: time.Second}
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
		exec, err := NewAdminLaunchPlanExecutor(ctx, ctrlConfig.GetConfig(), mockClient, nil, adminConfig, promutils.NewTestScope(), memStore, func(string) {})
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
		exec, err := NewAdminLaunchPlanExecutor(ctx, ctrlConfig.GetConfig(), mockClient, nil, adminConfig, promutils.NewTestScope(), memStore, func(string) {})
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
		exec, err := NewAdminLaunchPlanExecutor(ctx, ctrlConfig.GetConfig(), mockClient, nil, adminConfig, promutils.NewTestScope(), memStore, func(string) {})
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
	adminConfig := GetAdminConfig()
	adminConfig.CacheResyncDuration = config.Duration{Duration: time.Second}
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
		exec, err := NewAdminLaunchPlanExecutor(ctx, ctrlConfig.GetConfig(), mockClient, nil, adminConfig, promutils.NewTestScope(), memStore, func(string) {})
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
		exec, err := NewAdminLaunchPlanExecutor(ctx, ctrlConfig.GetConfig(), mockClient, nil, adminConfig, promutils.NewTestScope(), memStore, func(string) {})
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

type LPExecutorSuite struct {
	suite.Suite
	adminClient        *mocks.AdminServiceClient
	watchServiceClient *mocks.WatchServiceClient
	adminEventsSrv     *mocks.WatchService_WatchExecutionStatusUpdatesClient
	enqueueWorkflow    *mock.Mock
	pbStore            *storageMocks.ComposedProtobufStore
	executor           *adminLaunchPlanExecutor
	ctx                context.Context
	cancel             func()
	adminConfig        *AdminConfig
}

func (s *LPExecutorSuite) SetupTest() {
	s.ctx, s.cancel = context.WithCancel(context.TODO())
	s.adminClient = &mocks.AdminServiceClient{}
	s.adminClient.Test(s.T())
	s.watchServiceClient = &mocks.WatchServiceClient{}
	s.watchServiceClient.Test(s.T())
	s.enqueueWorkflow = &mock.Mock{}
	s.enqueueWorkflow.Test(s.T())
	s.adminEventsSrv = &mocks.WatchService_WatchExecutionStatusUpdatesClient{}
	s.adminEventsSrv.Test(s.T())
	s.pbStore = &storageMocks.ComposedProtobufStore{}
	s.pbStore.Test(s.T())

	cfg := ctrlConfig.GetConfig()
	cfg.ClusterID = "foo-cluster"
	s.watchServiceClient.
		OnWatchExecutionStatusUpdatesMatch(s.ctx, &watch.WatchExecutionStatusUpdatesRequest{Cluster: cfg.ClusterID}, mock.Anything).
		Return(s.adminEventsSrv, nil).
		Maybe()

	s.adminConfig = GetAdminConfig()
	s.adminConfig.Workers = 1
	s.adminConfig.WatchConfig.Enabled = true
	storageClient := &storage.DataStore{
		ComposedProtobufStore: s.pbStore,
		ReferenceConstructor:  &storageMocks.ReferenceConstructor{},
	}
	exec, err := NewAdminLaunchPlanExecutor(s.ctx, cfg, s.adminClient, s.watchServiceClient, s.adminConfig, promutils.NewTestScope(), storageClient, func(id string) {
		s.enqueueWorkflow.MethodCalled("enqueueWorkflow", id)
	})
	s.NoError(err)
	s.Require().IsType(&adminLaunchPlanExecutor{}, exec)
	s.executor = exec.(*adminLaunchPlanExecutor)
}

func (s *LPExecutorSuite) TearDownTest() {
	s.adminClient.AssertExpectations(s.T())
	s.watchServiceClient.AssertExpectations(s.T())
	s.enqueueWorkflow.AssertExpectations(s.T())
	s.adminEventsSrv.AssertExpectations(s.T())
	s.pbStore.AssertExpectations(s.T())
}

func (s *LPExecutorSuite) Test_syncItem_SUCCEEDED() {
	cacheItem := executionCacheItem{
		WorkflowExecutionIdentifier: *execID,
		ParentWorkflowID:            parentWorkflowID,
		Phase:                       core.WorkflowExecution_RUNNING,
		UpdatedAt:                   time.Now().Add(-4 * time.Minute),
		HasOutputs:                  true,
	}

	itemWrapper := &cacheMocks.ItemWrapper{}
	defer itemWrapper.AssertExpectations(s.T())
	itemWrapper.OnGetItem().Return(cacheItem).Once()
	itemWrapper.OnGetID().Return("id").Once()

	s.adminClient.
		OnGetExecutionMatch(s.ctx, &admin.WorkflowExecutionGetRequest{Id: execID}).
		Return(&admin.Execution{Closure: &admin.ExecutionClosure{
			Phase: core.WorkflowExecution_SUCCEEDED,
		}}, nil).
		Once()

	execData := &admin.WorkflowExecutionGetDataResponse{
		FullOutputs: &core.LiteralMap{
			Literals: map[string]*core.Literal{"foo": nil},
		},
	}
	s.adminClient.
		OnGetExecutionDataMatch(mock.Anything, &admin.WorkflowExecutionGetDataRequest{Id: execID}).
		Return(execData, nil).
		Once()
	s.enqueueWorkflow.On("enqueueWorkflow", parentWorkflowID).Return().Once()

	resp, err := s.executor.syncItem(s.ctx, cache.Batch{itemWrapper})

	s.NoError(err)
	s.Require().Len(resp, 1)
	respItem := resp[0]
	s.Equal(cache.Update, respItem.Action)
	actualCacheItem := respItem.Item.(executionCacheItem)
	s.Equal(core.WorkflowExecution_SUCCEEDED, actualCacheItem.Phase)
	s.NoError(actualCacheItem.SyncError)
}

func (s *LPExecutorSuite) Test_syncItem_FAILED() {
	cacheItem := executionCacheItem{
		WorkflowExecutionIdentifier: *execID,
		ParentWorkflowID:            parentWorkflowID,
		Phase:                       core.WorkflowExecution_RUNNING,
		UpdatedAt:                   time.Now().Add(-4 * time.Minute),
	}

	itemWrapper := &cacheMocks.ItemWrapper{}
	defer itemWrapper.AssertExpectations(s.T())
	itemWrapper.OnGetItem().Return(cacheItem).Once()
	itemWrapper.OnGetID().Return("id").Once()

	expectedErr := &core.ExecutionError{Code: "fail"}
	s.adminClient.
		OnGetExecutionMatch(s.ctx, &admin.WorkflowExecutionGetRequest{Id: execID}).
		Return(&admin.Execution{Closure: &admin.ExecutionClosure{
			Phase:        core.WorkflowExecution_FAILED,
			OutputResult: &admin.ExecutionClosure_Error{Error: expectedErr},
		}}, nil).
		Once()
	s.enqueueWorkflow.On("enqueueWorkflow", parentWorkflowID).Return().Once()

	resp, err := s.executor.syncItem(s.ctx, cache.Batch{itemWrapper})

	s.NoError(err)
	s.Require().Len(resp, 1)
	respItem := resp[0]
	s.Equal(cache.Update, respItem.Action)
	actualCacheItem := respItem.Item.(executionCacheItem)
	s.Equal(core.WorkflowExecution_FAILED, actualCacheItem.Phase)
	s.Equal(expectedErr, actualCacheItem.ExecutionError)
}

func (s *LPExecutorSuite) Test_syncItem_RUNNING() {
	cacheItem := executionCacheItem{
		WorkflowExecutionIdentifier: *execID,
		ParentWorkflowID:            parentWorkflowID,
		Phase:                       core.WorkflowExecution_QUEUED,
		UpdatedAt:                   time.Now().Add(-4 * time.Minute),
		HasOutputs:                  true,
	}

	itemWrapper := &cacheMocks.ItemWrapper{}
	defer itemWrapper.AssertExpectations(s.T())
	itemWrapper.OnGetItem().Return(cacheItem).Once()
	itemWrapper.OnGetID().Return("id").Once()

	s.adminClient.
		OnGetExecutionMatch(s.ctx, &admin.WorkflowExecutionGetRequest{Id: execID}).
		Return(&admin.Execution{Closure: &admin.ExecutionClosure{
			Phase: core.WorkflowExecution_RUNNING,
		}}, nil).
		Once()

	resp, err := s.executor.syncItem(s.ctx, cache.Batch{itemWrapper})

	s.NoError(err)
	s.Require().Len(resp, 1)
	respItem := resp[0]
	s.Equal(cache.Update, respItem.Action)
	actualCacheItem := respItem.Item.(executionCacheItem)
	s.Equal(core.WorkflowExecution_RUNNING, actualCacheItem.Phase)
	s.Empty(actualCacheItem.ExecutionOutputs)
	s.NoError(actualCacheItem.SyncError)
}

func (s *LPExecutorSuite) Test_syncItem_AlreadyTerminal() {
	cacheItem := executionCacheItem{
		WorkflowExecutionIdentifier: *execID,
		ParentWorkflowID:            parentWorkflowID,
		Phase:                       core.WorkflowExecution_ABORTED,
		UpdatedAt:                   time.Now().Add(-4 * time.Minute),
		HasOutputs:                  true,
	}

	itemWrapper := &cacheMocks.ItemWrapper{}
	defer itemWrapper.AssertExpectations(s.T())
	itemWrapper.OnGetItem().Return(cacheItem).Once()

	resp, err := s.executor.syncItem(s.ctx, cache.Batch{itemWrapper})

	s.NoError(err)
	s.Empty(resp)
}

func (s *LPExecutorSuite) Test_syncItem_RecentlyUpdated() {
	cacheItem := executionCacheItem{
		WorkflowExecutionIdentifier: *execID,
		ParentWorkflowID:            parentWorkflowID,
		Phase:                       core.WorkflowExecution_RUNNING,
		UpdatedAt:                   time.Now(),
		HasOutputs:                  true,
	}

	itemWrapper := &cacheMocks.ItemWrapper{}
	defer itemWrapper.AssertExpectations(s.T())
	itemWrapper.OnGetItem().Return(cacheItem).Once()

	resp, err := s.executor.syncItem(s.ctx, cache.Batch{itemWrapper})

	s.NoError(err)
	s.Empty(resp)
}

func (s *LPExecutorSuite) Test_syncItem_SyncError() {
	cacheItem := executionCacheItem{
		WorkflowExecutionIdentifier: *execID,
		ParentWorkflowID:            parentWorkflowID,
		Phase:                       core.WorkflowExecution_RUNNING,
		UpdatedAt:                   time.Now().Add(-4 * time.Minute),
		HasOutputs:                  true,
	}

	itemWrapper := &cacheMocks.ItemWrapper{}
	defer itemWrapper.AssertExpectations(s.T())
	itemWrapper.OnGetItem().Return(cacheItem).Once()
	itemWrapper.OnGetID().Return("id").Once()

	expectedErr := fmt.Errorf("get exec fail")
	s.adminClient.
		OnGetExecutionMatch(s.ctx, &admin.WorkflowExecutionGetRequest{Id: execID}).
		Return(nil, expectedErr).
		Once()

	resp, err := s.executor.syncItem(s.ctx, cache.Batch{itemWrapper})

	s.NoError(err)
	s.Require().Len(resp, 1)
	respItem := resp[0]
	s.Equal(cache.Update, respItem.Action)
	actualCacheItem := respItem.Item.(executionCacheItem)
	s.Equal(core.WorkflowExecution_UNDEFINED, actualCacheItem.Phase)
	s.ErrorIs(actualCacheItem.SyncError, expectedErr)
}

func (s *LPExecutorSuite) Test_syncItem_SUCCEEDED_SyncDataError() {
	cacheItem := executionCacheItem{
		WorkflowExecutionIdentifier: *execID,
		ParentWorkflowID:            parentWorkflowID,
		Phase:                       core.WorkflowExecution_RUNNING,
		UpdatedAt:                   time.Now().Add(-4 * time.Minute),
		HasOutputs:                  true,
	}

	itemWrapper := &cacheMocks.ItemWrapper{}
	defer itemWrapper.AssertExpectations(s.T())
	itemWrapper.OnGetItem().Return(cacheItem).Once()
	itemWrapper.OnGetID().Return("id").Once()

	s.adminClient.
		OnGetExecutionMatch(s.ctx, &admin.WorkflowExecutionGetRequest{Id: execID}).
		Return(&admin.Execution{Closure: &admin.ExecutionClosure{
			Phase: core.WorkflowExecution_SUCCEEDED,
		}}, nil).
		Once()

	expectedErr := fmt.Errorf("sync error")
	s.adminClient.
		OnGetExecutionDataMatch(mock.Anything, &admin.WorkflowExecutionGetDataRequest{Id: execID}).
		Return(nil, expectedErr).
		Once()
	s.enqueueWorkflow.On("enqueueWorkflow", parentWorkflowID).Return().Once()

	resp, err := s.executor.syncItem(s.ctx, cache.Batch{itemWrapper})

	s.NoError(err)
	s.Require().Len(resp, 1)
	respItem := resp[0]
	s.Equal(cache.Update, respItem.Action)
	actualCacheItem := respItem.Item.(executionCacheItem)
	s.Equal(core.WorkflowExecution_SUCCEEDED, actualCacheItem.Phase)
	s.Equal(expectedErr, actualCacheItem.SyncError)
}

func (s *LPExecutorSuite) Test_syncItem_SUCCEEDED_FetchDataFromURI_Exhausted() {
	cacheItem := executionCacheItem{
		WorkflowExecutionIdentifier: *execID,
		ParentWorkflowID:            parentWorkflowID,
		Phase:                       core.WorkflowExecution_RUNNING,
		UpdatedAt:                   time.Now().Add(-4 * time.Minute),
		HasOutputs:                  true,
	}

	itemWrapper := &cacheMocks.ItemWrapper{}
	defer itemWrapper.AssertExpectations(s.T())
	itemWrapper.OnGetItem().Return(cacheItem).Once()
	itemWrapper.OnGetID().Return("id").Once()

	s.adminClient.
		OnGetExecutionMatch(s.ctx, &admin.WorkflowExecutionGetRequest{Id: execID}).
		Return(&admin.Execution{Closure: &admin.ExecutionClosure{
			Phase: core.WorkflowExecution_SUCCEEDED,
			OutputResult: &admin.ExecutionClosure_Outputs{
				Outputs: &admin.LiteralMapBlob{Data: &admin.LiteralMapBlob_Uri{Uri: testURI}},
			},
		}}, nil).
		Once()

	s.adminClient.
		OnGetExecutionDataMatch(mock.Anything, &admin.WorkflowExecutionGetDataRequest{Id: execID}).
		Return(nil, status.Error(codes.ResourceExhausted, "")).
		Once()
	expectedOutputs := &core.LiteralMap{
		Literals: map[string]*core.Literal{"foo": nil},
	}
	s.pbStore.OnReadProtobufMatch(s.ctx, storage.DataReference(testURI), mock.Anything).
		Run(func(args mock.Arguments) {
			lm := args[2].(*core.LiteralMap)
			*lm = *expectedOutputs
		}).
		Return(nil).
		Once()
	s.enqueueWorkflow.On("enqueueWorkflow", parentWorkflowID).Return().Once()

	resp, err := s.executor.syncItem(s.ctx, cache.Batch{itemWrapper})

	s.NoError(err)
	s.Require().Len(resp, 1)
	respItem := resp[0]
	s.Equal(cache.Update, respItem.Action)
	actualCacheItem := respItem.Item.(executionCacheItem)
	s.Equal(core.WorkflowExecution_SUCCEEDED, actualCacheItem.Phase)
	s.Equal(expectedOutputs, actualCacheItem.ExecutionOutputs)
}

func (s *LPExecutorSuite) Test_syncItem_SUCCEEDED_FetchDataFromURI_MissingOutputs() {
	cacheItem := executionCacheItem{
		WorkflowExecutionIdentifier: *execID,
		ParentWorkflowID:            parentWorkflowID,
		Phase:                       core.WorkflowExecution_RUNNING,
		UpdatedAt:                   time.Now().Add(-4 * time.Minute),
		HasOutputs:                  true,
	}

	itemWrapper := &cacheMocks.ItemWrapper{}
	defer itemWrapper.AssertExpectations(s.T())
	itemWrapper.OnGetItem().Return(cacheItem).Once()
	itemWrapper.OnGetID().Return("id").Once()

	s.adminClient.
		OnGetExecutionMatch(s.ctx, &admin.WorkflowExecutionGetRequest{Id: execID}).
		Return(&admin.Execution{Closure: &admin.ExecutionClosure{
			Phase: core.WorkflowExecution_SUCCEEDED,
			OutputResult: &admin.ExecutionClosure_Outputs{
				Outputs: &admin.LiteralMapBlob{Data: &admin.LiteralMapBlob_Uri{Uri: testURI}},
			},
		}}, nil).
		Once()

	execData := &admin.WorkflowExecutionGetDataResponse{
		FullOutputs: &core.LiteralMap{
			Literals: nil,
		},
	}
	s.adminClient.
		OnGetExecutionDataMatch(mock.Anything, &admin.WorkflowExecutionGetDataRequest{Id: execID}).
		Return(execData, nil).
		Once()
	expectedOutputs := &core.LiteralMap{
		Literals: map[string]*core.Literal{"foo": nil},
	}
	s.pbStore.OnReadProtobufMatch(s.ctx, storage.DataReference(testURI), mock.Anything).
		Run(func(args mock.Arguments) {
			lm := args[2].(*core.LiteralMap)
			*lm = *expectedOutputs
		}).
		Return(nil).
		Once()
	s.enqueueWorkflow.On("enqueueWorkflow", parentWorkflowID).Return().Once()

	resp, err := s.executor.syncItem(s.ctx, cache.Batch{itemWrapper})

	s.NoError(err)
	s.Require().Len(resp, 1)
	respItem := resp[0]
	s.Equal(cache.Update, respItem.Action)
	actualCacheItem := respItem.Item.(executionCacheItem)
	s.Equal(core.WorkflowExecution_SUCCEEDED, actualCacheItem.Phase)
	s.Equal(expectedOutputs, actualCacheItem.ExecutionOutputs)
}

func (s *LPExecutorSuite) Test_syncItem_SUCCEEDED_FetchDataFromURI_ReadProtoError() {
	cacheItem := executionCacheItem{
		WorkflowExecutionIdentifier: *execID,
		ParentWorkflowID:            parentWorkflowID,
		Phase:                       core.WorkflowExecution_RUNNING,
		UpdatedAt:                   time.Now().Add(-4 * time.Minute),
		HasOutputs:                  true,
	}

	itemWrapper := &cacheMocks.ItemWrapper{}
	defer itemWrapper.AssertExpectations(s.T())
	itemWrapper.OnGetItem().Return(cacheItem).Once()
	itemWrapper.OnGetID().Return("id").Once()

	s.adminClient.
		OnGetExecutionMatch(s.ctx, &admin.WorkflowExecutionGetRequest{Id: execID}).
		Return(&admin.Execution{Closure: &admin.ExecutionClosure{
			Phase: core.WorkflowExecution_SUCCEEDED,
			OutputResult: &admin.ExecutionClosure_Outputs{
				Outputs: &admin.LiteralMapBlob{Data: &admin.LiteralMapBlob_Uri{Uri: testURI}},
			},
		}}, nil).
		Once()

	execData := &admin.WorkflowExecutionGetDataResponse{
		FullOutputs: &core.LiteralMap{
			Literals: nil,
		},
	}
	s.adminClient.
		OnGetExecutionDataMatch(mock.Anything, &admin.WorkflowExecutionGetDataRequest{Id: execID}).
		Return(execData, nil).
		Once()
	expectedErr := fmt.Errorf("fail proto")
	s.pbStore.OnReadProtobufMatch(s.ctx, storage.DataReference(testURI), mock.Anything).
		Return(expectedErr).
		Once()
	s.enqueueWorkflow.On("enqueueWorkflow", parentWorkflowID).Return().Once()

	resp, err := s.executor.syncItem(s.ctx, cache.Batch{itemWrapper})

	s.NoError(err)
	s.Require().Len(resp, 1)
	respItem := resp[0]
	s.Equal(cache.Update, respItem.Action)
	actualCacheItem := respItem.Item.(executionCacheItem)
	s.Equal(core.WorkflowExecution_SUCCEEDED, actualCacheItem.Phase)
	s.ErrorIs(actualCacheItem.SyncError, expectedErr)
}

func (s *LPExecutorSuite) Test_watchExecutionStatusUpdates_SUCCEEDED() {
	s.adminClient.
		OnCreateExecutionMatch(
			s.ctx,
			mock.MatchedBy(func(o *admin.ExecutionCreateRequest) bool {
				return o.Project == execID.Project && o.Domain == execID.Domain && o.Name == execID.Name && o.Org == execID.Org && o.Spec.Inputs == nil
			}),
		).
		Return(nil, nil).
		Once()
	err := s.executor.Launch(s.ctx,
		LaunchContext{
			ParentNodeExecution: &core.NodeExecutionIdentifier{
				NodeId:      "node-execID",
				ExecutionId: execID,
			},
		},
		execID,
		launchPlanWithOutputs,
		nil,
		parentWorkflowID,
	)
	s.NoError(err)
	s.adminEventsSrv.
		OnRecv().
		Run(func(_ mock.Arguments) {
			s.cancel()
		}).
		Return(&watch.WatchExecutionStatusUpdatesResponse{Id: execID, Phase: core.WorkflowExecution_SUCCEEDED}, nil).
		Once()
	execData := &admin.WorkflowExecutionGetDataResponse{
		FullOutputs: &core.LiteralMap{
			Literals: map[string]*core.Literal{"foo": nil},
		},
	}
	s.adminClient.
		OnGetExecutionDataMatch(mock.Anything, &admin.WorkflowExecutionGetDataRequest{Id: execID}).
		Return(execData, nil).
		Once()
	s.enqueueWorkflow.On("enqueueWorkflow", parentWorkflowID).Return().Once()
	s.adminEventsSrv.OnCloseSend().Return(nil).Once()

	s.executor.watchExecutionStatusUpdates(s.ctx)

	item, err := s.executor.cache.Get(execID.String())
	s.NoError(err)
	cacheItem := item.(executionCacheItem)
	s.Equal(execID, &cacheItem.WorkflowExecutionIdentifier)
	s.Equal(core.WorkflowExecution_SUCCEEDED, cacheItem.Phase)
	s.Equal(execData.FullOutputs, cacheItem.ExecutionOutputs)
}

func (s *LPExecutorSuite) Test_watchExecutionStatusUpdates_RUNNING() {
	s.adminClient.
		OnCreateExecutionMatch(
			s.ctx,
			mock.MatchedBy(func(o *admin.ExecutionCreateRequest) bool {
				return o.Project == execID.Project && o.Domain == execID.Domain && o.Name == execID.Name && o.Org == execID.Org && o.Spec.Inputs == nil
			}),
		).
		Return(nil, nil).
		Once()
	err := s.executor.Launch(s.ctx,
		LaunchContext{
			ParentNodeExecution: &core.NodeExecutionIdentifier{
				NodeId:      "node-execID",
				ExecutionId: execID,
			},
		},
		execID,
		launchPlanWithOutputs,
		nil,
		parentWorkflowID,
	)
	s.NoError(err)
	s.adminEventsSrv.
		OnRecv().
		Run(func(_ mock.Arguments) {
			s.cancel()
		}).
		Return(&watch.WatchExecutionStatusUpdatesResponse{Id: execID, Phase: core.WorkflowExecution_RUNNING}, nil).
		Once()
	s.adminEventsSrv.OnCloseSend().Return(nil).Once()

	s.executor.watchExecutionStatusUpdates(s.ctx)

	item, err := s.executor.cache.Get(execID.String())
	s.NoError(err)
	cacheItem := item.(executionCacheItem)
	s.Equal(execID, &cacheItem.WorkflowExecutionIdentifier)
	s.Equal(core.WorkflowExecution_RUNNING, cacheItem.Phase)
	s.Nil(cacheItem.ExecutionOutputs)
}

func (s *LPExecutorSuite) Test_watchExecutionStatusUpdates_ExecutionNotFoundInCache() {
	s.adminEventsSrv.
		OnRecv().
		Run(func(_ mock.Arguments) {
			s.cancel()
		}).
		Return(&watch.WatchExecutionStatusUpdatesResponse{Id: execID, Phase: core.WorkflowExecution_RUNNING}, nil).
		Once()
	s.adminEventsSrv.OnCloseSend().Return(nil).Once()

	s.executor.watchExecutionStatusUpdates(s.ctx)

	_, err := s.executor.cache.Get(execID.String())
	s.Error(err)
	s.True(errors.IsCausedBy(err, cache.ErrNotFound))
}

func (s *LPExecutorSuite) Test_watchExecutionStatusUpdates_ExecutionAlreadyTerminated() {
	s.adminEventsSrv.
		OnRecv().
		Run(func(_ mock.Arguments) {
			s.cancel()
		}).
		Return(&watch.WatchExecutionStatusUpdatesResponse{Id: execID, Phase: core.WorkflowExecution_SUCCEEDED}, nil).
		Once()
	s.adminEventsSrv.OnCloseSend().Return(nil).Once()
	_, err := s.executor.cache.GetOrCreate(execID.String(), executionCacheItem{
		Phase: core.WorkflowExecution_FAILED,
	})
	s.Require().NoError(err)

	s.executor.watchExecutionStatusUpdates(s.ctx)

	item, err := s.executor.cache.Get(execID.String())
	s.NoError(err)
	cacheItem := item.(executionCacheItem)
	s.Equal(core.WorkflowExecution_FAILED, cacheItem.Phase)
}

func (s *LPExecutorSuite) Test_watchExecutionStatusUpdates_FetchExecutionDataError() {
	s.adminClient.
		OnCreateExecutionMatch(
			s.ctx,
			mock.MatchedBy(func(o *admin.ExecutionCreateRequest) bool {
				return o.Project == execID.Project && o.Domain == execID.Domain && o.Name == execID.Name && o.Org == execID.Org && o.Spec.Inputs == nil
			}),
		).
		Return(nil, nil).
		Once()
	err := s.executor.Launch(s.ctx,
		LaunchContext{
			ParentNodeExecution: &core.NodeExecutionIdentifier{
				NodeId:      "node-execID",
				ExecutionId: execID,
			},
		},
		execID,
		launchPlanWithOutputs,
		nil,
		parentWorkflowID,
	)
	s.NoError(err)
	s.adminEventsSrv.
		OnRecv().
		Run(func(_ mock.Arguments) {
			s.cancel()
		}).
		Return(&watch.WatchExecutionStatusUpdatesResponse{Id: execID, Phase: core.WorkflowExecution_SUCCEEDED}, nil).
		Once()
	expectedErr := fmt.Errorf("exec data fail")
	s.adminClient.
		OnGetExecutionDataMatch(mock.Anything, &admin.WorkflowExecutionGetDataRequest{Id: execID}).
		Return(nil, expectedErr).
		Once()
	s.enqueueWorkflow.On("enqueueWorkflow", parentWorkflowID).Return().Once()
	s.adminEventsSrv.OnCloseSend().Return(nil).Once()

	s.executor.watchExecutionStatusUpdates(s.ctx)

	item, err := s.executor.cache.Get(execID.String())
	s.NoError(err)
	cacheItem := item.(executionCacheItem)
	s.Equal(execID, &cacheItem.WorkflowExecutionIdentifier)
	s.Equal(core.WorkflowExecution_SUCCEEDED, cacheItem.Phase)
	s.Equal(expectedErr, cacheItem.SyncError)
}

func TestLPExecutorSuite(t *testing.T) {
	suite.Run(t, new(LPExecutorSuite))
}

func init() {
	labeled.SetMetricKeys(contextutils.ProjectKey, contextutils.DomainKey, contextutils.WorkflowIDKey, contextutils.TaskIDKey)
}
