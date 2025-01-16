package subworkflow

import (
	"context"
	"fmt"
	"reflect"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"

	"github.com/flyteorg/flyte/flyteidl/clients/go/coreutils"
	"github.com/flyteorg/flyte/flyteidl/gen/pb-go/flyteidl/admin"
	"github.com/flyteorg/flyte/flyteidl/gen/pb-go/flyteidl/core"
	mocks4 "github.com/flyteorg/flyte/flyteplugins/go/tasks/pluginmachinery/io/mocks"
	"github.com/flyteorg/flyte/flytepropeller/pkg/apis/flyteworkflow/v1alpha1"
	mocks2 "github.com/flyteorg/flyte/flytepropeller/pkg/apis/flyteworkflow/v1alpha1/mocks"
	execMocks "github.com/flyteorg/flyte/flytepropeller/pkg/controller/executors/mocks"
	"github.com/flyteorg/flyte/flytepropeller/pkg/controller/nodes/handler"
	mocks3 "github.com/flyteorg/flyte/flytepropeller/pkg/controller/nodes/interfaces/mocks"
	recoveryMocks "github.com/flyteorg/flyte/flytepropeller/pkg/controller/nodes/recovery/mocks"
	"github.com/flyteorg/flyte/flytepropeller/pkg/controller/nodes/subworkflow/launchplan"
	"github.com/flyteorg/flyte/flytepropeller/pkg/controller/nodes/subworkflow/launchplan/mocks"
	"github.com/flyteorg/flyte/flytepropeller/pkg/utils"
	"github.com/flyteorg/flyte/flytestdlib/errors"
	"github.com/flyteorg/flyte/flytestdlib/promutils"
	"github.com/flyteorg/flyte/flytestdlib/storage"
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
	k8sWorkflowID := types.NamespacedName{
		Namespace: "namespace",
		Name:      "name",
	}
	mockWfNode := &mocks2.ExecutableWorkflowNode{}
	mockWfNode.On("GetLaunchPlanRefID").Return(&v1alpha1.Identifier{
		Identifier: lpID,
	})

	mockNode := &mocks2.ExecutableNode{}
	mockNode.OnGetID().Return("n1")
	mockNode.OnGetWorkflowNode().Return(mockWfNode)
	mockNode.OnGetConfig().Return(make(map[string]string))

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
			mock.MatchedBy(func(o v1alpha1.ExecutableLaunchPlan) bool { return true }),
			mock.MatchedBy(func(o *core.LiteralMap) bool { return o.Literals == nil }),
			mock.MatchedBy(func(o string) bool { return o == k8sWorkflowID.String() }),
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
			mock.MatchedBy(func(o v1alpha1.ExecutableLaunchPlan) bool { return true }),
			mock.MatchedBy(func(o *core.LiteralMap) bool { return o.Literals == nil }),
			mock.MatchedBy(func(o string) bool { return o == k8sWorkflowID.String() }),
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
			mock.MatchedBy(func(o v1alpha1.ExecutableLaunchPlan) bool { return true }),
			mock.MatchedBy(func(o *core.LiteralMap) bool { return o.Literals == nil }),
			mock.MatchedBy(func(o string) bool { return o == k8sWorkflowID.String() }),
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
			//mock.MatchedBy(func(o *core.Identifier) bool { return lpID == o }),
			mock.MatchedBy(func(o v1alpha1.ExecutableLaunchPlan) bool { return true }),
			mock.MatchedBy(func(o *core.LiteralMap) bool { return o.Literals == nil }),
			mock.MatchedBy(func(o string) bool { return o == k8sWorkflowID.String() }),
		).Return(errors.Wrapf(launchplan.RemoteErrorUser, fmt.Errorf("blah"), "failed"))

		nCtx := createNodeContext(v1alpha1.WorkflowNodePhaseExecuting, mockNode, mockNodeStatus)
		s, err := h.StartLaunchPlan(ctx, nCtx)
		assert.NoError(t, err)
		assert.Equal(t, handler.EPhaseFailed, s.Info().GetPhase())
	})
	t.Run("recover successfully", func(t *testing.T) {
		recoveredExecID := &core.WorkflowExecutionIdentifier{
			Project: "p",
			Domain:  "d",
			Name:    "n",
		}

		mockLPExec := &mocks.Executor{}
		mockLPExec.On("Launch", mock.Anything, launchplan.LaunchContext{
			ParentNodeExecution: &core.NodeExecutionIdentifier{
				NodeId: "n",
				ExecutionId: &core.WorkflowExecutionIdentifier{
					Project: "project",
					Domain:  "domain",
					Name:    "name",
				},
			},
			RecoveryExecution: recoveredExecID,
		}, mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(nil)

		recoveryClient := recoveryMocks.Client{}
		recoveryClient.On("RecoverNodeExecution", mock.Anything, recoveredExecID, mock.Anything).Return(&admin.NodeExecution{
			Closure: &admin.NodeExecutionClosure{
				Phase: core.NodeExecution_SUCCEEDED,
				TargetMetadata: &admin.NodeExecutionClosure_WorkflowNodeMetadata{
					WorkflowNodeMetadata: &admin.WorkflowNodeMetadata{
						ExecutionId: recoveredExecID,
					},
				},
			},
		}, nil)

		h := launchPlanHandler{
			launchPlan:     mockLPExec,
			recoveryClient: &recoveryClient,
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
			mock.MatchedBy(func(o v1alpha1.ExecutableLaunchPlan) bool { return true }),
			mock.MatchedBy(func(o *core.LiteralMap) bool { return o.Literals == nil }),
			mock.MatchedBy(func(o string) bool { return o == k8sWorkflowID.String() }),
		).Return(nil)

		wfStatus := &mocks2.MutableWorkflowNodeStatus{}
		mockNodeStatus.On("GetOrCreateWorkflowStatus").Return(wfStatus)

		nCtx := &mocks3.NodeExecutionContext{}

		ir := &mocks4.InputReader{}
		inputs := &core.LiteralMap{}
		ir.OnGetMatch(mock.Anything).Return(inputs, nil)
		nCtx.OnInputReader().Return(ir)

		nm := &mocks3.NodeExecutionMetadata{}
		nm.OnGetAnnotations().Return(map[string]string{})
		nm.OnGetNodeExecutionID().Return(&core.NodeExecutionIdentifier{
			ExecutionId: wfExecID,
			NodeId:      "n",
		})
		nm.OnGetK8sServiceAccount().Return("service-account")
		nm.OnGetLabels().Return(map[string]string{})
		nm.OnGetNamespace().Return("namespace")
		nm.OnGetOwnerID().Return(types.NamespacedName{Namespace: "namespace", Name: "name"})
		nm.OnGetOwnerReference().Return(v1.OwnerReference{
			Kind: "sample",
			Name: "name",
		})

		nCtx.OnNodeExecutionMetadata().Return(nm)
		ectx := &execMocks.ExecutionContext{}
		ectx.OnGetDefinitionVersion().Return(v1alpha1.WorkflowDefinitionVersion1)
		ectx.OnGetEventVersion().Return(1)
		ectx.OnGetParentInfo().Return(nil)
		ectx.OnGetExecutionConfig().Return(v1alpha1.ExecutionConfig{
			RecoveryExecution: v1alpha1.WorkflowExecutionIdentifier{
				WorkflowExecutionIdentifier: recoveredExecID,
			},
		})
		ectx.OnIncrementParallelism().Return(1)
		ectx.OnGetSecurityContext().Return(core.SecurityContext{})
		ectx.OnGetRawOutputDataConfig().Return(v1alpha1.RawOutputDataConfig{})
		ectx.OnGetLabels().Return(nil)
		ectx.OnGetAnnotations().Return(nil)
		ectx.OnFindLaunchPlanMatch(mock.Anything).Return(&core.LaunchPlanTemplate{})

		nCtx.OnExecutionContext().Return(ectx)
		nCtx.OnCurrentAttempt().Return(uint32(1))
		nCtx.OnNode().Return(mockNode)

		s, err := h.StartLaunchPlan(ctx, nCtx)
		assert.NoError(t, err)
		assert.Equal(t, s.Info().GetPhase(), handler.EPhaseRunning)
		assert.Equal(t, len(recoveryClient.Calls), 1)
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
	mockNode.OnGetID().Return("n1")
	mockNode.OnGetWorkflowNode().Return(mockWfNode)
	mockNode.OnGetConfig().Return(make(map[string]string))

	mockNodeStatus := &mocks2.ExecutableNodeStatus{}
	mockNodeStatus.On("GetAttempts").Return(attempts)
	mockNodeStatus.On("GetDataDir").Return(dataDir)
	mockNodeStatus.On("GetOutputDir").Return(dataDir)

	t.Run("stillRunning", func(t *testing.T) {
		mockLPExec := &mocks.Executor{}
		mockLPExec.
			OnGetStatusMatch(ctx,
				mock.MatchedBy(func(o *core.WorkflowExecutionIdentifier) bool {
					return assert.Equal(t, wfExecID.Project, o.Project) && assert.Equal(t, wfExecID.Domain, o.Domain)
				}),
				mock.MatchedBy(func(o v1alpha1.ExecutableLaunchPlan) bool { return true }),
				mock.MatchedBy(func(o v1alpha1.WorkflowID) bool { return true }),
			).
			Return(launchplan.ExecutionStatus{Phase: core.WorkflowExecution_RUNNING}, nil).
			Once()
		h := launchPlanHandler{launchPlan: mockLPExec}
		nCtx := createNodeContext(v1alpha1.WorkflowNodePhaseExecuting, mockNode, mockNodeStatus)

		s, err := h.CheckLaunchPlanStatus(ctx, nCtx)

		assert.NoError(t, err)
		assert.Equal(t, handler.EPhaseRunning, s.Info().GetPhase())
		mockLPExec.AssertExpectations(t)
	})

	t.Run("successNoOutputs", func(t *testing.T) {
		mockLPExec := &mocks.Executor{}
		mockLPExec.
			OnGetStatusMatch(ctx,
				mock.MatchedBy(func(o *core.WorkflowExecutionIdentifier) bool {
					return assert.Equal(t, wfExecID.Project, o.Project) && assert.Equal(t, wfExecID.Domain, o.Domain)
				}),
				mock.MatchedBy(func(o v1alpha1.ExecutableLaunchPlan) bool { return true }),
				mock.MatchedBy(func(o v1alpha1.WorkflowID) bool { return true }),
			).
			Return(launchplan.ExecutionStatus{Phase: core.WorkflowExecution_SUCCEEDED}, nil).
			Once()
		h := launchPlanHandler{launchPlan: mockLPExec}
		nCtx := createNodeContext(v1alpha1.WorkflowNodePhaseExecuting, mockNode, mockNodeStatus)

		s, err := h.CheckLaunchPlanStatus(ctx, nCtx)

		assert.NoError(t, err)
		assert.Equal(t, handler.EPhaseSuccess, s.Info().GetPhase())
		mockLPExec.AssertExpectations(t)
	})

	t.Run("successOutputURI", func(t *testing.T) {
		mockStore := createInmemoryStore(t)
		uri := storage.DataReference("uri")
		op := &core.LiteralMap{
			Literals: map[string]*core.Literal{
				"x": coreutils.MustMakePrimitiveLiteral(1),
			},
		}
		mockLPExec := &mocks.Executor{}
		mockLPExec.
			OnGetStatusMatch(
				ctx,
				mock.MatchedBy(func(o *core.WorkflowExecutionIdentifier) bool {
					return assert.Equal(t, wfExecID.Project, o.Project) && assert.Equal(t, wfExecID.Domain, o.Domain)
				}),
				mock.MatchedBy(func(o v1alpha1.ExecutableLaunchPlan) bool { return true }),
				mock.MatchedBy(func(o v1alpha1.WorkflowID) bool { return true }),
			).
			Return(launchplan.ExecutionStatus{Phase: core.WorkflowExecution_SUCCEEDED, Outputs: op}, nil).
			Once()
		// tODO ssingh: do we need mockStore
		h := launchPlanHandler{
			launchPlan:  mockLPExec,
			eventConfig: eventConfig,
		}
		err := mockStore.WriteProtobuf(ctx, uri, storage.Options{}, op)
		assert.NoError(t, err)
		nCtx := createNodeContext(v1alpha1.WorkflowNodePhaseExecuting, mockNode, mockNodeStatus)
		nCtx.OnDataStore().Return(mockStore).Maybe()

		s, err := h.CheckLaunchPlanStatus(ctx, nCtx)
		assert.NoError(t, err)
		assert.Equal(t, handler.EPhaseSuccess, s.Info().GetPhase())
		final := &core.LiteralMap{}
		assert.NoError(t, mockStore.ReadProtobuf(ctx, v1alpha1.GetOutputsFile(dataDir), final), mockStore)
		v, ok := final.GetLiterals()["x"]
		assert.True(t, ok)
		assert.Equal(t, int64(1), v.GetScalar().GetPrimitive().GetInteger())
		mockLPExec.AssertExpectations(t)
	})

	t.Run("successOutputs", func(t *testing.T) {
		mockStore := createInmemoryStore(t)
		op := &core.LiteralMap{
			Literals: map[string]*core.Literal{
				"x": coreutils.MustMakePrimitiveLiteral(1),
			},
		}
		mockLPExec := &mocks.Executor{}
		mockLPExec.
			OnGetStatusMatch(ctx,
				mock.MatchedBy(func(o *core.WorkflowExecutionIdentifier) bool {
					return assert.Equal(t, wfExecID.Project, o.Project) && assert.Equal(t, wfExecID.Domain, o.Domain)
				}),
				mock.MatchedBy(func(o v1alpha1.ExecutableLaunchPlan) bool { return true }),
				mock.MatchedBy(func(o v1alpha1.WorkflowID) bool { return true }),
			).
			Return(launchplan.ExecutionStatus{Phase: core.WorkflowExecution_SUCCEEDED, Outputs: op}, nil).
			Once()
		nCtx := createNodeContext(v1alpha1.WorkflowNodePhaseExecuting, mockNode, mockNodeStatus)
		nCtx.OnDataStore().Return(mockStore)
		h := launchPlanHandler{
			launchPlan:  mockLPExec,
			eventConfig: eventConfig,
		}

		s, err := h.CheckLaunchPlanStatus(ctx, nCtx)

		assert.NoError(t, err)
		assert.Equal(t, s.Info().GetPhase(), handler.EPhaseSuccess)
		final := &core.LiteralMap{}
		assert.NoError(t, mockStore.ReadProtobuf(ctx, v1alpha1.GetOutputsFile(dataDir), final))
		v, ok := final.GetLiterals()["x"]
		assert.True(t, ok)
		assert.Equal(t, int64(1), v.GetScalar().GetPrimitive().GetInteger())
		mockLPExec.AssertExpectations(t)
	})

	t.Run("failureError", func(t *testing.T) {
		mockLPExec := &mocks.Executor{}
		mockLPExec.
			OnGetStatusMatch(
				ctx,
				mock.MatchedBy(func(o *core.WorkflowExecutionIdentifier) bool {
					return assert.Equal(t, wfExecID.Project, o.Project) && assert.Equal(t, wfExecID.Domain, o.Domain)
				}),
				mock.MatchedBy(func(o v1alpha1.ExecutableLaunchPlan) bool { return true }),
				mock.MatchedBy(func(o v1alpha1.WorkflowID) bool { return true }),
			).
			Return(launchplan.ExecutionStatus{
				Phase: core.WorkflowExecution_FAILED,
				Error: &core.ExecutionError{
					Message: "msg",
					Code:    "code",
				},
			}, nil).
			Once()
		h := launchPlanHandler{
			launchPlan:  mockLPExec,
			eventConfig: eventConfig,
		}
		nCtx := createNodeContext(v1alpha1.WorkflowNodePhaseExecuting, mockNode, mockNodeStatus)

		s, err := h.CheckLaunchPlanStatus(ctx, nCtx)

		assert.NoError(t, err)
		assert.Equal(t, s.Info().GetPhase(), handler.EPhaseFailed)
		mockLPExec.AssertExpectations(t)
	})

	t.Run("failureNoError", func(t *testing.T) {
		mockLPExec := &mocks.Executor{}
		mockLPExec.
			OnGetStatusMatch(
				ctx,
				mock.MatchedBy(func(o *core.WorkflowExecutionIdentifier) bool {
					return assert.Equal(t, wfExecID.Project, o.Project) && assert.Equal(t, wfExecID.Domain, o.Domain)
				}),
				mock.MatchedBy(func(o v1alpha1.ExecutableLaunchPlan) bool { return true }),
				mock.MatchedBy(func(o v1alpha1.WorkflowID) bool { return true }),
			).
			Return(launchplan.ExecutionStatus{Phase: core.WorkflowExecution_FAILED}, nil).
			Once()
		h := launchPlanHandler{
			launchPlan:  mockLPExec,
			eventConfig: eventConfig,
		}
		nCtx := createNodeContext(v1alpha1.WorkflowNodePhaseExecuting, mockNode, mockNodeStatus)

		s, err := h.CheckLaunchPlanStatus(ctx, nCtx)

		assert.NoError(t, err)
		assert.Equal(t, s.Info().GetPhase(), handler.EPhaseFailed)
		mockLPExec.AssertExpectations(t)
	})

	t.Run("aborted", func(t *testing.T) {
		mockLPExec := &mocks.Executor{}
		mockLPExec.OnGetStatusMatch(
			ctx,
			mock.MatchedBy(func(o *core.WorkflowExecutionIdentifier) bool {
				return assert.Equal(t, wfExecID.Project, o.Project) && assert.Equal(t, wfExecID.Domain, o.Domain)
			}),
			mock.MatchedBy(func(o v1alpha1.ExecutableLaunchPlan) bool { return true }),
			mock.MatchedBy(func(o v1alpha1.WorkflowID) bool { return true }),
		).
			Return(launchplan.ExecutionStatus{Phase: core.WorkflowExecution_ABORTED}, nil).
			Once()
		h := launchPlanHandler{
			launchPlan:  mockLPExec,
			eventConfig: eventConfig,
		}
		nCtx := createNodeContext(v1alpha1.WorkflowNodePhaseExecuting, mockNode, mockNodeStatus)

		s, err := h.CheckLaunchPlanStatus(ctx, nCtx)

		assert.NoError(t, err)
		assert.Equal(t, s.Info().GetPhase(), handler.EPhaseFailed)
		mockLPExec.AssertExpectations(t)
	})

	t.Run("notFound", func(t *testing.T) {
		mockLPExec := &mocks.Executor{}
		mockLPExec.
			OnGetStatusMatch(ctx,
				mock.MatchedBy(func(o *core.WorkflowExecutionIdentifier) bool {
					return assert.Equal(t, wfExecID.Project, o.Project) && assert.Equal(t, wfExecID.Domain, o.Domain)
				}),
				mock.MatchedBy(func(o v1alpha1.ExecutableLaunchPlan) bool { return true }),
				mock.MatchedBy(func(o v1alpha1.WorkflowID) bool { return true }),
			).
			Return(launchplan.ExecutionStatus{}, errors.Wrapf(launchplan.RemoteErrorNotFound, fmt.Errorf("some error"), "not found")).
			Once()
		h := launchPlanHandler{
			launchPlan:  mockLPExec,
			eventConfig: eventConfig,
		}
		nCtx := createNodeContext(v1alpha1.WorkflowNodePhaseExecuting, mockNode, mockNodeStatus)

		s, err := h.CheckLaunchPlanStatus(ctx, nCtx)

		assert.NoError(t, err)
		assert.Equal(t, s.Info().GetPhase(), handler.EPhaseFailed)
		mockLPExec.AssertExpectations(t)
	})

	t.Run("systemError", func(t *testing.T) {
		mockLPExec := &mocks.Executor{}
		mockLPExec.
			OnGetStatusMatch(
				ctx,
				mock.MatchedBy(func(o *core.WorkflowExecutionIdentifier) bool {
					return assert.Equal(t, wfExecID.Project, o.Project) && assert.Equal(t, wfExecID.Domain, o.Domain)
				}),
				mock.MatchedBy(func(o v1alpha1.ExecutableLaunchPlan) bool { return true }),
				mock.MatchedBy(func(o v1alpha1.WorkflowID) bool { return true }),
			).
			Return(launchplan.ExecutionStatus{}, errors.Wrapf(launchplan.RemoteErrorSystem, fmt.Errorf("some error"), "not found")).
			Once()
		h := launchPlanHandler{
			launchPlan:  mockLPExec,
			eventConfig: eventConfig,
		}
		nCtx := createNodeContext(v1alpha1.WorkflowNodePhaseExecuting, mockNode, mockNodeStatus)

		s, err := h.CheckLaunchPlanStatus(ctx, nCtx)

		assert.Error(t, err)
		assert.Equal(t, s.Info().GetPhase(), handler.EPhaseUndefined)
		mockLPExec.AssertExpectations(t)
	})

	t.Run("dataStoreFailure", func(t *testing.T) {
		mockStore := storage.NewCompositeDataStore(storage.URLPathConstructor{}, storage.NewDefaultProtobufStore(utils.FailingRawStore{}, promutils.NewTestScope()))
		op := &core.LiteralMap{
			Literals: map[string]*core.Literal{
				"x": coreutils.MustMakePrimitiveLiteral(1),
			},
		}
		mockLPExec := &mocks.Executor{}
		mockLPExec.
			OnGetStatusMatch(ctx,
				mock.MatchedBy(func(o *core.WorkflowExecutionIdentifier) bool {
					return assert.Equal(t, wfExecID.Project, o.Project) && assert.Equal(t, wfExecID.Domain, o.Domain)
				}),
				mock.MatchedBy(func(o v1alpha1.ExecutableLaunchPlan) bool { return true }),
				mock.MatchedBy(func(o v1alpha1.WorkflowID) bool { return true }),
			).
			Return(launchplan.ExecutionStatus{
				Phase:   core.WorkflowExecution_SUCCEEDED,
				Outputs: op,
			}, nil).
			Once()
		h := launchPlanHandler{
			launchPlan:  mockLPExec,
			eventConfig: eventConfig,
		}
		nCtx := createNodeContext(v1alpha1.WorkflowNodePhaseExecuting, mockNode, mockNodeStatus)
		nCtx.OnDataStore().Return(mockStore)

		s, err := h.CheckLaunchPlanStatus(ctx, nCtx)

		assert.Error(t, err)
		assert.Equal(t, s.Info().GetPhase(), handler.EPhaseUndefined)
		mockLPExec.AssertExpectations(t)
	})

	t.Run("outputURISystemError", func(t *testing.T) {
		mockStore := storage.NewCompositeDataStore(storage.URLPathConstructor{}, storage.NewDefaultProtobufStore(utils.FailingRawStore{}, promutils.NewTestScope()))
		mockLPExec := &mocks.Executor{}
		mockLPExec.
			OnGetStatusMatch(ctx,
				mock.MatchedBy(func(o *core.WorkflowExecutionIdentifier) bool {
					return assert.Equal(t, wfExecID.Project, o.Project) && assert.Equal(t, wfExecID.Domain, o.Domain)
				}),
				mock.MatchedBy(func(o v1alpha1.ExecutableLaunchPlan) bool { return true }),
				mock.MatchedBy(func(o v1alpha1.WorkflowID) bool { return true }),
			).
			Return(launchplan.ExecutionStatus{
				Phase:   core.WorkflowExecution_SUCCEEDED,
				Outputs: &core.LiteralMap{},
			}, nil).
			Once()
		h := launchPlanHandler{
			launchPlan:  mockLPExec,
			eventConfig: eventConfig,
		}
		nCtx := createNodeContext(v1alpha1.WorkflowNodePhaseExecuting, mockNode, mockNodeStatus)
		nCtx.OnDataStore().Return(mockStore)

		s, err := h.CheckLaunchPlanStatus(ctx, nCtx)

		assert.Error(t, err)
		assert.Equal(t, s.Info().GetPhase().String(), handler.EPhaseUndefined.String())
		mockLPExec.AssertExpectations(t)
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
	mockNode.OnGetID().Return(nodeID)
	mockNode.OnGetWorkflowNode().Return(mockWfNode)
	mockNode.OnGetConfig().Return(make(map[string]string))

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
