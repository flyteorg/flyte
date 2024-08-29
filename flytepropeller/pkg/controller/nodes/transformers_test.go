package nodes

import (
	"testing"

	"github.com/golang/protobuf/proto"
	"github.com/stretchr/testify/assert"

	"github.com/flyteorg/flyte/flyteidl/clients/go/coreutils"
	"github.com/flyteorg/flyte/flyteidl/gen/pb-go/flyteidl/core"
	"github.com/flyteorg/flyte/flyteidl/gen/pb-go/flyteidl/event"
	"github.com/flyteorg/flyte/flytepropeller/pkg/apis/flyteworkflow/v1alpha1"
	"github.com/flyteorg/flyte/flytepropeller/pkg/apis/flyteworkflow/v1alpha1/mocks"
	"github.com/flyteorg/flyte/flytepropeller/pkg/controller/config"
	mocks2 "github.com/flyteorg/flyte/flytepropeller/pkg/controller/executors/mocks"
	"github.com/flyteorg/flyte/flytepropeller/pkg/controller/nodes/handler"
	"github.com/flyteorg/flyte/flytestdlib/storage"
)

func TestToNodeExecutionEvent(t *testing.T) {
	project := "project"
	domain := "domain"
	t.Run("dynamic node", func(t *testing.T) {
		info := handler.PhaseInfoDynamicRunning(&handler.ExecutionInfo{TaskNodeInfo: &handler.TaskNodeInfo{
			TaskNodeMetadata: &event.TaskNodeMetadata{
				DynamicWorkflow: &event.DynamicWorkflowNodeMetadata{
					Id: &core.Identifier{
						Project: project,
						Domain:  domain,
						Name:    "wfname",
						Version: "wfversion",
					},
				},
			},
		}})
		status := mocks.ExecutableNodeStatus{}
		status.OnGetOutputDir().Return(storage.DataReference("s3://foo/bar"))
		status.OnGetParentNodeID().Return(nil)
		parentInfo := mocks2.ImmutableParentInfo{}
		parentInfo.OnCurrentAttempt().Return(0)
		parentInfo.OnGetUniqueID().Return("u")
		parentInfo.OnIsInDynamicChain().Return(true)
		node := mocks.ExecutableNode{}
		node.OnGetID().Return("n")
		node.OnGetName().Return("nodey")
		node.OnGetKind().Return(v1alpha1.NodeKindTask)

		nev, err := ToNodeExecutionEvent(&core.NodeExecutionIdentifier{
			NodeId: "nodey",
			ExecutionId: &core.WorkflowExecutionIdentifier{
				Project: "project",
				Domain:  "domain",
				Name:    "exec",
			},
		}, info, "inputPath", &status, v1alpha1.EventVersion2, &parentInfo, &node, "clusterID", v1alpha1.DynamicNodePhaseParentFinalized, &config.EventConfig{
			RawOutputPolicy: config.RawOutputPolicyReference,
		}, nil)
		assert.NoError(t, err)
		assert.True(t, nev.IsDynamic)
		assert.True(t, nev.IsParent)
		assert.Equal(t, nodeExecutionEventVersion, nev.EventVersion)
		assert.True(t, nev.IsInDynamicChain)
	})
	t.Run("is parent", func(t *testing.T) {
		info := handler.PhaseInfoDynamicRunning(&handler.ExecutionInfo{TaskNodeInfo: &handler.TaskNodeInfo{
			TaskNodeMetadata: &event.TaskNodeMetadata{},
		}})
		status := mocks.ExecutableNodeStatus{}
		status.OnGetOutputDir().Return(storage.DataReference("s3://foo/bar"))
		status.OnGetParentNodeID().Return(nil)
		parentInfo := mocks2.ImmutableParentInfo{}
		parentInfo.OnCurrentAttempt().Return(0)
		parentInfo.OnGetUniqueID().Return("u")
		parentInfo.OnIsInDynamicChain().Return(false)
		node := mocks.ExecutableNode{}
		node.OnGetID().Return("n")
		node.OnGetName().Return("nodey")
		node.OnGetKind().Return(v1alpha1.NodeKindWorkflow)
		executableWorkflowNode := mocks.ExecutableWorkflowNode{}
		subworkflowRef := "ref"
		executableWorkflowNode.OnGetSubWorkflowRef().Return(&subworkflowRef)
		node.OnGetWorkflowNode().Return(&executableWorkflowNode)

		nev, err := ToNodeExecutionEvent(&core.NodeExecutionIdentifier{
			NodeId: "nodey",
			ExecutionId: &core.WorkflowExecutionIdentifier{
				Project: "project",
				Domain:  "domain",
				Name:    "exec",
			},
		}, info, "inputPath", &status, v1alpha1.EventVersion2, &parentInfo, &node, "clusterID", v1alpha1.DynamicNodePhaseNone, &config.EventConfig{
			RawOutputPolicy: config.RawOutputPolicyReference,
		}, nil)
		assert.NoError(t, err)
		assert.False(t, nev.IsDynamic)
		assert.True(t, nev.IsParent)
		assert.Equal(t, nodeExecutionEventVersion, nev.EventVersion)
	})
	t.Run("inline events", func(t *testing.T) {
		inputs := &core.LiteralMap{
			Literals: map[string]*core.Literal{
				"foo": coreutils.MustMakeLiteral("bar"),
			},
		}
		info := handler.PhaseInfoQueued("z", inputs)
		status := mocks.ExecutableNodeStatus{}
		status.OnGetOutputDir().Return(storage.DataReference("s3://foo/bar"))
		status.OnGetParentNodeID().Return(nil)
		node := mocks.ExecutableNode{}
		node.OnGetID().Return("n")
		node.OnGetName().Return("nodey")
		node.OnGetKind().Return(v1alpha1.NodeKindTask)
		nev, err := ToNodeExecutionEvent(&core.NodeExecutionIdentifier{
			NodeId: "nodey",
			ExecutionId: &core.WorkflowExecutionIdentifier{
				Project: "project",
				Domain:  "domain",
				Name:    "exec",
			},
		}, info, "inputPath", &status, v1alpha1.EventVersion2, nil, &node, "clusterID", v1alpha1.DynamicNodePhaseParentFinalized, &config.EventConfig{
			RawOutputPolicy: config.RawOutputPolicyInline,
		}, nil)
		assert.NoError(t, err)
		assert.True(t, proto.Equal(inputs, nev.GetInputData()))
	})
}
