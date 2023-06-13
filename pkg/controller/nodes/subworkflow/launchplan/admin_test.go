package launchplan

import (
	"context"
	"testing"
	"time"

	"github.com/golang/protobuf/proto"

	"github.com/flyteorg/flytestdlib/cache"
	mocks2 "github.com/flyteorg/flytestdlib/cache/mocks"

	"github.com/flyteorg/flytestdlib/promutils"

	"github.com/flyteorg/flyteidl/clients/go/admin/mocks"
	"github.com/flyteorg/flyteidl/gen/pb-go/flyteidl/admin"
	"github.com/flyteorg/flyteidl/gen/pb-go/flyteidl/core"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func TestAdminLaunchPlanExecutor_GetStatus(t *testing.T) {
	ctx := context.TODO()
	id := &core.WorkflowExecutionIdentifier{
		Name:    "n",
		Domain:  "d",
		Project: "p",
	}
	var result *admin.ExecutionClosure

	t.Run("happy", func(t *testing.T) {
		mockClient := &mocks.AdminServiceClient{}
		exec, err := NewAdminLaunchPlanExecutor(ctx, mockClient, time.Millisecond, defaultAdminConfig, promutils.NewTestScope())
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

	t.Run("terminal-sync", func(t *testing.T) {
		mockClient := &mocks.AdminServiceClient{}
		exec, err := NewAdminLaunchPlanExecutor(ctx, mockClient, time.Millisecond, defaultAdminConfig, promutils.NewTestScope())
		assert.NoError(t, err)
		iwMock := &mocks2.ItemWrapper{}
		i := executionCacheItem{ExecutionClosure: &admin.ExecutionClosure{Phase: core.WorkflowExecution_SUCCEEDED, WorkflowId: &core.Identifier{Project: "p"}}}
		iwMock.OnGetItem().Return(i)
		iwMock.OnGetID().Return("id")
		adminExec := exec.(*adminLaunchPlanExecutor)
		v, err := adminExec.syncItem(ctx, cache.Batch{
			iwMock,
		})
		assert.NoError(t, err)
		assert.NotNil(t, v)
		assert.Len(t, v, 1)
		assert.Equal(t, v[0].ID, "id")
		assert.Equal(t, v[0].Item, i)
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

		exec, err := NewAdminLaunchPlanExecutor(ctx, mockClient, time.Millisecond, defaultAdminConfig, promutils.NewTestScope())
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

		exec, err := NewAdminLaunchPlanExecutor(ctx, mockClient, time.Millisecond, defaultAdminConfig, promutils.NewTestScope())
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

	t.Run("happy", func(t *testing.T) {

		mockClient := &mocks.AdminServiceClient{}
		exec, err := NewAdminLaunchPlanExecutor(ctx, mockClient, time.Second, defaultAdminConfig, promutils.NewTestScope())
		mockClient.On("CreateExecution",
			ctx,
			mock.MatchedBy(func(o *admin.ExecutionCreateRequest) bool {
				return o.Project == "p" && o.Domain == "d" && o.Name == "n" && o.Spec.Inputs == nil &&
					o.Spec.Metadata.Mode == admin.ExecutionMetadata_CHILD_WORKFLOW
			}),
		).Return(nil, nil)
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
		assert.NoError(t, err)
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
		exec, err := NewAdminLaunchPlanExecutor(ctx, mockClient, time.Second, defaultAdminConfig, promutils.NewTestScope())
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
		exec, err := NewAdminLaunchPlanExecutor(ctx, mockClient, time.Second, defaultAdminConfig, promutils.NewTestScope())
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
		exec, err := NewAdminLaunchPlanExecutor(ctx, mockClient, time.Second, defaultAdminConfig, promutils.NewTestScope())
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
		exec, err := NewAdminLaunchPlanExecutor(ctx, mockClient, time.Second, defaultAdminConfig, promutils.NewTestScope())
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

	const reason = "reason"
	t.Run("happy", func(t *testing.T) {

		mockClient := &mocks.AdminServiceClient{}
		exec, err := NewAdminLaunchPlanExecutor(ctx, mockClient, time.Second, defaultAdminConfig, promutils.NewTestScope())
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
		exec, err := NewAdminLaunchPlanExecutor(ctx, mockClient, time.Second, defaultAdminConfig, promutils.NewTestScope())
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
		exec, err := NewAdminLaunchPlanExecutor(ctx, mockClient, time.Second, defaultAdminConfig, promutils.NewTestScope())
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

	t.Run("launch plan found", func(t *testing.T) {
		mockClient := &mocks.AdminServiceClient{}
		exec, err := NewAdminLaunchPlanExecutor(ctx, mockClient, time.Second, defaultAdminConfig, promutils.NewTestScope())
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
		exec, err := NewAdminLaunchPlanExecutor(ctx, mockClient, time.Second, defaultAdminConfig, promutils.NewTestScope())
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
