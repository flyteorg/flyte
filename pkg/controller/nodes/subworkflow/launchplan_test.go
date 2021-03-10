package subworkflow

import (
	"context"
	"fmt"
	"reflect"
	"testing"

	"github.com/flyteorg/flytestdlib/errors"

	"github.com/flyteorg/flyteidl/gen/pb-go/flyteidl/admin"
	"github.com/flyteorg/flyteidl/gen/pb-go/flyteidl/core"
	"github.com/flyteorg/flytestdlib/promutils"
	"github.com/flyteorg/flytestdlib/storage"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"

	"github.com/flyteorg/flytepropeller/pkg/apis/flyteworkflow/v1alpha1"
	mocks2 "github.com/flyteorg/flytepropeller/pkg/apis/flyteworkflow/v1alpha1/mocks"
	execMocks "github.com/flyteorg/flytepropeller/pkg/controller/executors/mocks"
	"github.com/flyteorg/flytepropeller/pkg/controller/nodes/handler"
	"github.com/flyteorg/flytepropeller/pkg/controller/nodes/subworkflow/launchplan"
	"github.com/flyteorg/flytepropeller/pkg/controller/nodes/subworkflow/launchplan/mocks"
	"github.com/flyteorg/flytepropeller/pkg/utils"
)

func createInmemoryStore(t testing.TB) *storage.DataStore {
	cfg := storage.Config{
		Type: storage.TypeMemory,
	}

	d, err := storage.NewDataStore(&cfg, promutils.NewTestScope())
	assert.NoError(t, err)

	return d
}

func TestSubWorkflowHandler_StartLaunchPlan(t *testing.T) {
	ctx := context.TODO()

	attempts := uint32(1)

	lpID := &core.Identifier{
		Project:      "p",
		Domain:       "d",
		Name:         "n",
		Version:      "v",
		ResourceType: core.ResourceType_LAUNCH_PLAN,
	}
	mockWfNode := &mocks2.ExecutableWorkflowNode{}
	mockWfNode.On("GetLaunchPlanRefID").Return(&v1alpha1.Identifier{
		Identifier: lpID,
	})

	mockNode := &mocks2.ExecutableNode{}
	mockNode.On("GetID").Return("n1")
	mockNode.On("GetWorkflowNode").Return(mockWfNode)

	mockNodeStatus := &mocks2.ExecutableNodeStatus{}
	mockNodeStatus.On("GetAttempts").Return(attempts)

	t.Run("happy", func(t *testing.T) {

		mockLPExec := &mocks.Executor{}

		h := launchPlanHandler{
			launchPlan: mockLPExec,
		}
		mockLPExec.On("Launch",
			ctx,
			mock.MatchedBy(func(o launchplan.LaunchContext) bool {
				return o.ParentNodeExecution.NodeId == mockNode.GetID() &&
					o.ParentNodeExecution.ExecutionId == wfExecID
			}),
			mock.MatchedBy(func(o *core.WorkflowExecutionIdentifier) bool {
				return assert.Equal(t, wfExecID.Project, o.Project) && assert.Equal(t, wfExecID.Domain, o.Domain)
			}),
			mock.MatchedBy(func(o *core.Identifier) bool { return lpID == o }),
			mock.MatchedBy(func(o *core.LiteralMap) bool { return o.Literals == nil }),
		).Return(nil)

		wfStatus := &mocks2.MutableWorkflowNodeStatus{}
		mockNodeStatus.On("GetOrCreateWorkflowStatus").Return(wfStatus)

		nCtx := createNodeContext(v1alpha1.WorkflowNodePhaseUndefined, mockNode, mockNodeStatus)
		s, err := h.StartLaunchPlan(ctx, nCtx)
		assert.NoError(t, err)
		assert.Equal(t, s.Info().GetPhase(), handler.EPhaseRunning)
	})

	t.Run("alreadyExists", func(t *testing.T) {

		mockLPExec := &mocks.Executor{}

		h := launchPlanHandler{
			launchPlan: mockLPExec,
		}
		mockLPExec.On("Launch",
			ctx,
			mock.MatchedBy(func(o launchplan.LaunchContext) bool {
				return o.ParentNodeExecution.NodeId == mockNode.GetID() &&
					o.ParentNodeExecution.ExecutionId == wfExecID
			}),
			mock.MatchedBy(func(o *core.WorkflowExecutionIdentifier) bool {
				return assert.Equal(t, wfExecID.Project, o.Project) && assert.Equal(t, wfExecID.Domain, o.Domain)
			}),
			mock.MatchedBy(func(o *core.Identifier) bool { return lpID == o }),
			mock.MatchedBy(func(o *core.LiteralMap) bool { return o.Literals == nil }),
		).Return(errors.Wrapf(launchplan.RemoteErrorAlreadyExists, fmt.Errorf("blah"), "failed"))

		nCtx := createNodeContext(v1alpha1.WorkflowNodePhaseUndefined, mockNode, mockNodeStatus)
		s, err := h.StartLaunchPlan(ctx, nCtx)
		assert.NoError(t, err)
		assert.Equal(t, s.Info().GetPhase(), handler.EPhaseRunning)
	})

	t.Run("systemError", func(t *testing.T) {

		mockLPExec := &mocks.Executor{}

		h := launchPlanHandler{
			launchPlan: mockLPExec,
		}
		mockLPExec.On("Launch",
			ctx,
			mock.MatchedBy(func(o launchplan.LaunchContext) bool {
				return o.ParentNodeExecution.NodeId == mockNode.GetID() &&
					o.ParentNodeExecution.ExecutionId == wfExecID
			}),
			mock.MatchedBy(func(o *core.WorkflowExecutionIdentifier) bool {
				return assert.Equal(t, wfExecID.Project, o.Project) && assert.Equal(t, wfExecID.Domain, o.Domain)
			}),
			mock.MatchedBy(func(o *core.Identifier) bool { return lpID == o }),
			mock.MatchedBy(func(o *core.LiteralMap) bool { return o.Literals == nil }),
		).Return(errors.Wrapf(launchplan.RemoteErrorSystem, fmt.Errorf("blah"), "failed"))

		nCtx := createNodeContext(v1alpha1.WorkflowNodePhaseExecuting, mockNode, mockNodeStatus)
		s, err := h.StartLaunchPlan(ctx, nCtx)
		assert.Error(t, err)
		assert.Equal(t, handler.EPhaseUndefined, s.Info().GetPhase())
	})

	t.Run("userError", func(t *testing.T) {

		mockLPExec := &mocks.Executor{}

		h := launchPlanHandler{
			launchPlan: mockLPExec,
		}
		mockLPExec.On("Launch",
			ctx,
			mock.MatchedBy(func(o launchplan.LaunchContext) bool {
				return o.ParentNodeExecution.NodeId == mockNode.GetID() &&
					o.ParentNodeExecution.ExecutionId == wfExecID
			}),
			mock.MatchedBy(func(o *core.WorkflowExecutionIdentifier) bool {
				return assert.Equal(t, wfExecID.Project, o.Project) && assert.Equal(t, wfExecID.Domain, o.Domain)
			}),
			mock.MatchedBy(func(o *core.Identifier) bool { return lpID == o }),
			mock.MatchedBy(func(o *core.LiteralMap) bool { return o.Literals == nil }),
		).Return(errors.Wrapf(launchplan.RemoteErrorUser, fmt.Errorf("blah"), "failed"))

		nCtx := createNodeContext(v1alpha1.WorkflowNodePhaseExecuting, mockNode, mockNodeStatus)
		s, err := h.StartLaunchPlan(ctx, nCtx)
		assert.NoError(t, err)
		assert.Equal(t, handler.EPhaseFailed, s.Info().GetPhase())
	})
}

func TestSubWorkflowHandler_CheckLaunchPlanStatus(t *testing.T) {

	ctx := context.TODO()

	attempts := uint32(1)
	dataDir := storage.DataReference("data")

	lpID := &core.Identifier{
		Project:      "p",
		Domain:       "d",
		Name:         "n",
		Version:      "v",
		ResourceType: core.ResourceType_LAUNCH_PLAN,
	}
	mockWfNode := &mocks2.ExecutableWorkflowNode{}
	mockWfNode.On("GetLaunchPlanRefID").Return(&v1alpha1.Identifier{
		Identifier: lpID,
	})

	mockNode := &mocks2.ExecutableNode{}
	mockNode.On("GetID").Return("n1")
	mockNode.On("GetWorkflowNode").Return(mockWfNode)

	mockNodeStatus := &mocks2.ExecutableNodeStatus{}
	mockNodeStatus.On("GetAttempts").Return(attempts)
	mockNodeStatus.On("GetDataDir").Return(dataDir)
	mockNodeStatus.On("GetOutputDir").Return(dataDir)

	t.Run("stillRunning", func(t *testing.T) {

		mockLPExec := &mocks.Executor{}

		h := launchPlanHandler{
			launchPlan: mockLPExec,
		}
		mockLPExec.On("GetStatus",
			ctx,
			mock.MatchedBy(func(o *core.WorkflowExecutionIdentifier) bool {
				return assert.Equal(t, wfExecID.Project, o.Project) && assert.Equal(t, wfExecID.Domain, o.Domain)
			}),
		).Return(&admin.ExecutionClosure{
			Phase: core.WorkflowExecution_RUNNING,
		}, nil)

		nCtx := createNodeContext(v1alpha1.WorkflowNodePhaseExecuting, mockNode, mockNodeStatus)
		s, err := h.CheckLaunchPlanStatus(ctx, nCtx)

		assert.NoError(t, err)
		assert.Equal(t, handler.EPhaseRunning, s.Info().GetPhase())
	})

	t.Run("successNoOutputs", func(t *testing.T) {

		mockLPExec := &mocks.Executor{}

		h := launchPlanHandler{
			launchPlan: mockLPExec,
		}
		mockLPExec.On("GetStatus",
			ctx,
			mock.MatchedBy(func(o *core.WorkflowExecutionIdentifier) bool {
				return assert.Equal(t, wfExecID.Project, o.Project) && assert.Equal(t, wfExecID.Domain, o.Domain)
			}),
		).Return(&admin.ExecutionClosure{
			Phase: core.WorkflowExecution_SUCCEEDED,
		}, nil)

		nCtx := createNodeContext(v1alpha1.WorkflowNodePhaseExecuting, mockNode, mockNodeStatus)
		s, err := h.CheckLaunchPlanStatus(ctx, nCtx)
		assert.NoError(t, err)
		assert.Equal(t, handler.EPhaseSuccess, s.Info().GetPhase())
	})

	t.Run("successOutputURI", func(t *testing.T) {

		mockStore := createInmemoryStore(t)
		mockLPExec := &mocks.Executor{}
		uri := storage.DataReference("uri")

		// tODO ssingh: do we need mockStore
		h := launchPlanHandler{
			launchPlan: mockLPExec,
		}

		op := &core.LiteralMap{
			Literals: map[string]*core.Literal{
				"x": utils.MustMakePrimitiveLiteral(1),
			},
		}
		err := mockStore.WriteProtobuf(ctx, uri, storage.Options{}, op)
		assert.NoError(t, err)

		mockLPExec.On("GetStatus",
			ctx,
			mock.MatchedBy(func(o *core.WorkflowExecutionIdentifier) bool {
				return assert.Equal(t, wfExecID.Project, o.Project) && assert.Equal(t, wfExecID.Domain, o.Domain)
			}),
		).Return(&admin.ExecutionClosure{
			Phase: core.WorkflowExecution_SUCCEEDED,
			OutputResult: &admin.ExecutionClosure_Outputs{
				Outputs: &admin.LiteralMapBlob{
					Data: &admin.LiteralMapBlob_Uri{
						Uri: uri.String(),
					},
				},
			},
		}, nil)

		nCtx := createNodeContext(v1alpha1.WorkflowNodePhaseExecuting, mockNode, mockNodeStatus)
		nCtx.OnDataStore().Return(mockStore)
		s, err := h.CheckLaunchPlanStatus(ctx, nCtx)
		assert.NoError(t, err)
		assert.Equal(t, handler.EPhaseSuccess, s.Info().GetPhase())
		final := &core.LiteralMap{}
		assert.NoError(t, mockStore.ReadProtobuf(ctx, v1alpha1.GetOutputsFile(dataDir), final), mockStore)
		v, ok := final.GetLiterals()["x"]
		assert.True(t, ok)
		assert.Equal(t, int64(1), v.GetScalar().GetPrimitive().GetInteger())
	})

	t.Run("successOutputs", func(t *testing.T) {

		mockStore := createInmemoryStore(t)
		mockLPExec := &mocks.Executor{}
		h := launchPlanHandler{
			launchPlan: mockLPExec,
		}

		op := &core.LiteralMap{
			Literals: map[string]*core.Literal{
				"x": utils.MustMakePrimitiveLiteral(1),
			},
		}
		mockLPExec.On("GetStatus",
			ctx,
			mock.MatchedBy(func(o *core.WorkflowExecutionIdentifier) bool {
				return assert.Equal(t, wfExecID.Project, o.Project) && assert.Equal(t, wfExecID.Domain, o.Domain)
			}),
		).Return(&admin.ExecutionClosure{
			Phase: core.WorkflowExecution_SUCCEEDED,
			OutputResult: &admin.ExecutionClosure_Outputs{
				Outputs: &admin.LiteralMapBlob{
					Data: &admin.LiteralMapBlob_Values{
						Values: op,
					},
				},
			},
		}, nil)

		nCtx := createNodeContext(v1alpha1.WorkflowNodePhaseExecuting, mockNode, mockNodeStatus)
		nCtx.OnDataStore().Return(mockStore)
		s, err := h.CheckLaunchPlanStatus(ctx, nCtx)
		assert.NoError(t, err)
		assert.Equal(t, s.Info().GetPhase(), handler.EPhaseSuccess)
		final := &core.LiteralMap{}
		assert.NoError(t, mockStore.ReadProtobuf(ctx, v1alpha1.GetOutputsFile(dataDir), final))
		v, ok := final.GetLiterals()["x"]
		assert.True(t, ok)
		assert.Equal(t, int64(1), v.GetScalar().GetPrimitive().GetInteger())
	})

	t.Run("failureError", func(t *testing.T) {

		mockLPExec := &mocks.Executor{}

		h := launchPlanHandler{
			launchPlan: mockLPExec,
		}

		mockLPExec.On("GetStatus",
			ctx,
			mock.MatchedBy(func(o *core.WorkflowExecutionIdentifier) bool {
				return assert.Equal(t, wfExecID.Project, o.Project) && assert.Equal(t, wfExecID.Domain, o.Domain)
			}),
		).Return(&admin.ExecutionClosure{
			Phase: core.WorkflowExecution_FAILED,
			OutputResult: &admin.ExecutionClosure_Error{
				Error: &core.ExecutionError{
					Message: "msg",
					Code:    "code",
				},
			},
		}, nil)

		nCtx := createNodeContext(v1alpha1.WorkflowNodePhaseExecuting, mockNode, mockNodeStatus)
		s, err := h.CheckLaunchPlanStatus(ctx, nCtx)
		assert.NoError(t, err)
		assert.Equal(t, s.Info().GetPhase(), handler.EPhaseFailed)
	})

	t.Run("failureNoError", func(t *testing.T) {

		mockLPExec := &mocks.Executor{}

		h := launchPlanHandler{
			launchPlan: mockLPExec,
		}

		mockLPExec.On("GetStatus",
			ctx,
			mock.MatchedBy(func(o *core.WorkflowExecutionIdentifier) bool {
				return assert.Equal(t, wfExecID.Project, o.Project) && assert.Equal(t, wfExecID.Domain, o.Domain)
			}),
		).Return(&admin.ExecutionClosure{
			Phase: core.WorkflowExecution_FAILED,
		}, nil)

		nCtx := createNodeContext(v1alpha1.WorkflowNodePhaseExecuting, mockNode, mockNodeStatus)
		s, err := h.CheckLaunchPlanStatus(ctx, nCtx)
		assert.NoError(t, err)
		assert.Equal(t, s.Info().GetPhase(), handler.EPhaseFailed)
	})

	t.Run("aborted", func(t *testing.T) {

		mockLPExec := &mocks.Executor{}

		h := launchPlanHandler{
			launchPlan: mockLPExec,
		}

		mockLPExec.On("GetStatus",
			ctx,
			mock.MatchedBy(func(o *core.WorkflowExecutionIdentifier) bool {
				return assert.Equal(t, wfExecID.Project, o.Project) && assert.Equal(t, wfExecID.Domain, o.Domain)
			}),
		).Return(&admin.ExecutionClosure{
			Phase: core.WorkflowExecution_ABORTED,
		}, nil)

		nCtx := createNodeContext(v1alpha1.WorkflowNodePhaseExecuting, mockNode, mockNodeStatus)
		s, err := h.CheckLaunchPlanStatus(ctx, nCtx)
		assert.NoError(t, err)
		assert.Equal(t, s.Info().GetPhase(), handler.EPhaseFailed)
	})

	t.Run("notFound", func(t *testing.T) {

		mockLPExec := &mocks.Executor{}

		h := launchPlanHandler{
			launchPlan: mockLPExec,
		}

		mockLPExec.On("GetStatus",
			ctx,
			mock.MatchedBy(func(o *core.WorkflowExecutionIdentifier) bool {
				return assert.Equal(t, wfExecID.Project, o.Project) && assert.Equal(t, wfExecID.Domain, o.Domain)
			}),
		).Return(nil, errors.Wrapf(launchplan.RemoteErrorNotFound, fmt.Errorf("some error"), "not found"))

		nCtx := createNodeContext(v1alpha1.WorkflowNodePhaseExecuting, mockNode, mockNodeStatus)
		s, err := h.CheckLaunchPlanStatus(ctx, nCtx)
		assert.NoError(t, err)
		assert.Equal(t, s.Info().GetPhase(), handler.EPhaseFailed)
	})

	t.Run("systemError", func(t *testing.T) {

		mockLPExec := &mocks.Executor{}

		h := launchPlanHandler{
			launchPlan: mockLPExec,
		}

		mockLPExec.On("GetStatus",
			ctx,
			mock.MatchedBy(func(o *core.WorkflowExecutionIdentifier) bool {
				return assert.Equal(t, wfExecID.Project, o.Project) && assert.Equal(t, wfExecID.Domain, o.Domain)
			}),
		).Return(nil, errors.Wrapf(launchplan.RemoteErrorSystem, fmt.Errorf("some error"), "not found"))

		nCtx := createNodeContext(v1alpha1.WorkflowNodePhaseExecuting, mockNode, mockNodeStatus)
		s, err := h.CheckLaunchPlanStatus(ctx, nCtx)
		assert.Error(t, err)
		assert.Equal(t, s.Info().GetPhase(), handler.EPhaseUndefined)
	})

	t.Run("dataStoreFailure", func(t *testing.T) {

		mockStore := storage.NewCompositeDataStore(storage.URLPathConstructor{}, storage.NewDefaultProtobufStore(utils.FailingRawStore{}, promutils.NewTestScope()))
		mockLPExec := &mocks.Executor{}

		h := launchPlanHandler{
			launchPlan: mockLPExec,
		}

		op := &core.LiteralMap{
			Literals: map[string]*core.Literal{
				"x": utils.MustMakePrimitiveLiteral(1),
			},
		}
		mockLPExec.On("GetStatus",
			ctx,
			mock.MatchedBy(func(o *core.WorkflowExecutionIdentifier) bool {
				return assert.Equal(t, wfExecID.Project, o.Project) && assert.Equal(t, wfExecID.Domain, o.Domain)
			}),
		).Return(&admin.ExecutionClosure{
			Phase: core.WorkflowExecution_SUCCEEDED,
			OutputResult: &admin.ExecutionClosure_Outputs{
				Outputs: &admin.LiteralMapBlob{
					Data: &admin.LiteralMapBlob_Values{
						Values: op,
					},
				},
			},
		}, nil)

		nCtx := createNodeContext(v1alpha1.WorkflowNodePhaseExecuting, mockNode, mockNodeStatus)
		nCtx.OnDataStore().Return(mockStore)
		s, err := h.CheckLaunchPlanStatus(ctx, nCtx)
		assert.Error(t, err)
		assert.Equal(t, s.Info().GetPhase(), handler.EPhaseUndefined)
	})

	t.Run("outputURINotFound", func(t *testing.T) {

		mockStore := createInmemoryStore(t)
		mockLPExec := &mocks.Executor{}
		uri := storage.DataReference("uri")

		h := launchPlanHandler{
			launchPlan: mockLPExec,
		}

		mockLPExec.On("GetStatus",
			ctx,
			mock.MatchedBy(func(o *core.WorkflowExecutionIdentifier) bool {
				return assert.Equal(t, wfExecID.Project, o.Project) && assert.Equal(t, wfExecID.Domain, o.Domain)
			}),
		).Return(&admin.ExecutionClosure{
			Phase: core.WorkflowExecution_SUCCEEDED,
			OutputResult: &admin.ExecutionClosure_Outputs{
				Outputs: &admin.LiteralMapBlob{
					Data: &admin.LiteralMapBlob_Uri{
						Uri: uri.String(),
					},
				},
			},
		}, nil)

		nCtx := createNodeContext(v1alpha1.WorkflowNodePhaseExecuting, mockNode, mockNodeStatus)
		nCtx.OnDataStore().Return(mockStore)
		s, err := h.CheckLaunchPlanStatus(ctx, nCtx)
		assert.NotNil(t, err)
		assert.Equal(t, handler.EPhaseUndefined, s.Info().GetPhase())
	})

	t.Run("outputURISystemError", func(t *testing.T) {

		mockStore := storage.NewCompositeDataStore(storage.URLPathConstructor{}, storage.NewDefaultProtobufStore(utils.FailingRawStore{}, promutils.NewTestScope()))
		mockLPExec := &mocks.Executor{}
		uri := storage.DataReference("uri")

		h := launchPlanHandler{
			launchPlan: mockLPExec,
		}

		mockLPExec.On("GetStatus",
			ctx,
			mock.MatchedBy(func(o *core.WorkflowExecutionIdentifier) bool {
				return assert.Equal(t, wfExecID.Project, o.Project) && assert.Equal(t, wfExecID.Domain, o.Domain)
			}),
		).Return(&admin.ExecutionClosure{
			Phase: core.WorkflowExecution_SUCCEEDED,
			OutputResult: &admin.ExecutionClosure_Outputs{
				Outputs: &admin.LiteralMapBlob{
					Data: &admin.LiteralMapBlob_Uri{
						Uri: uri.String(),
					},
				},
			},
		}, nil)

		nCtx := createNodeContext(v1alpha1.WorkflowNodePhaseExecuting, mockNode, mockNodeStatus)
		nCtx.OnDataStore().Return(mockStore)
		s, err := h.CheckLaunchPlanStatus(ctx, nCtx)
		assert.Error(t, err)
		assert.Equal(t, s.Info().GetPhase().String(), handler.EPhaseUndefined.String())
	})
}

func TestLaunchPlanHandler_HandleAbort(t *testing.T) {

	ctx := context.TODO()

	nodeID := "n1"
	attempts := uint32(1)
	dataDir := storage.DataReference("data")

	lpID := &core.Identifier{
		Project:      "p",
		Domain:       "d",
		Name:         "n",
		Version:      "v",
		ResourceType: core.ResourceType_LAUNCH_PLAN,
	}
	mockWfNode := &mocks2.ExecutableWorkflowNode{}
	mockWfNode.On("GetLaunchPlanRefID").Return(&v1alpha1.Identifier{
		Identifier: lpID,
	})

	mockNode := &mocks2.ExecutableNode{}
	mockNode.On("GetID").Return(nodeID)
	mockNode.On("GetWorkflowNode").Return(mockWfNode)

	mockNodeStatus := &mocks2.ExecutableNodeStatus{}
	mockNodeStatus.On("GetAttempts").Return(attempts)
	mockNodeStatus.On("GetDataDir").Return(dataDir)

	t.Run("abort-success", func(t *testing.T) {
		mockLPExec := &mocks.Executor{}
		mockLPExec.On("Kill",
			ctx,
			mock.MatchedBy(func(o *core.WorkflowExecutionIdentifier) bool {
				return assert.Equal(t, wfExecID.Project, o.Project) && assert.Equal(t, wfExecID.Domain, o.Domain)
			}),
			mock.AnythingOfType(reflect.String.String()),
		).Return(nil)

		h := launchPlanHandler{
			launchPlan: mockLPExec,
		}
		nCtx := createNodeContext(v1alpha1.WorkflowNodePhaseExecuting, mockNode, mockNodeStatus)
		eCtx := &execMocks.ExecutionContext{}
		eCtx.OnGetName().Return("name")
		nCtx.OnExecutionContext().Return(eCtx)
		err := h.HandleAbort(ctx, nCtx, "some reason")
		assert.NoError(t, err)
	})

	t.Run("abort-fail", func(t *testing.T) {
		expectedErr := fmt.Errorf("fail")
		mockLPExec := &mocks.Executor{}
		mockLPExec.On("Kill",
			ctx,
			mock.MatchedBy(func(o *core.WorkflowExecutionIdentifier) bool {
				return assert.Equal(t, wfExecID.Project, o.Project) && assert.Equal(t, wfExecID.Domain, o.Domain)
			}),
			mock.AnythingOfType(reflect.String.String()),
		).Return(expectedErr)

		h := launchPlanHandler{
			launchPlan: mockLPExec,
		}
		nCtx := createNodeContext(v1alpha1.WorkflowNodePhaseExecuting, mockNode, mockNodeStatus)
		err := h.HandleAbort(ctx, nCtx, "reason")
		assert.Error(t, err)
		assert.Equal(t, err, expectedErr)
	})
}
