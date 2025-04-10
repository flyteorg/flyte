package nodes

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"google.golang.org/protobuf/proto"

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
		status.EXPECT().GetOutputDir().Return(storage.DataReference("s3://foo/bar"))
		status.EXPECT().GetParentNodeID().Return(nil)
		parentInfo := mocks2.ImmutableParentInfo{}
		parentInfo.EXPECT().CurrentAttempt().Return(0)
		parentInfo.EXPECT().GetUniqueID().Return("u")
		parentInfo.EXPECT().IsInDynamicChain().Return(true)
		node := mocks.ExecutableNode{}
		node.EXPECT().GetID().Return("n")
		node.EXPECT().GetName().Return("nodey")
		node.EXPECT().GetKind().Return(v1alpha1.NodeKindTask)

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
		assert.True(t, nev.GetIsDynamic())
		assert.True(t, nev.GetIsParent())
		assert.Equal(t, nodeExecutionEventVersion, nev.GetEventVersion())
		assert.True(t, nev.GetIsInDynamicChain())
	})
	t.Run("is parent", func(t *testing.T) {
		info := handler.PhaseInfoDynamicRunning(&handler.ExecutionInfo{TaskNodeInfo: &handler.TaskNodeInfo{
			TaskNodeMetadata: &event.TaskNodeMetadata{},
		}})
		status := mocks.ExecutableNodeStatus{}
		status.EXPECT().GetOutputDir().Return(storage.DataReference("s3://foo/bar"))
		status.EXPECT().GetParentNodeID().Return(nil)
		parentInfo := mocks2.ImmutableParentInfo{}
		parentInfo.EXPECT().CurrentAttempt().Return(0)
		parentInfo.EXPECT().GetUniqueID().Return("u")
		parentInfo.EXPECT().IsInDynamicChain().Return(false)
		node := mocks.ExecutableNode{}
		node.EXPECT().GetID().Return("n")
		node.EXPECT().GetName().Return("nodey")
		node.EXPECT().GetKind().Return(v1alpha1.NodeKindWorkflow)
		executableWorkflowNode := mocks.ExecutableWorkflowNode{}
		subworkflowRef := "ref"
		executableWorkflowNode.EXPECT().GetSubWorkflowRef().Return(&subworkflowRef)
		node.EXPECT().GetWorkflowNode().Return(&executableWorkflowNode)

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
		assert.False(t, nev.GetIsDynamic())
		assert.True(t, nev.GetIsParent())
		assert.Equal(t, nodeExecutionEventVersion, nev.GetEventVersion())
	})
	t.Run("inline events", func(t *testing.T) {
		inputs := &core.LiteralMap{
			Literals: map[string]*core.Literal{
				"foo": coreutils.MustMakeLiteral("bar"),
			},
		}
		info := handler.PhaseInfoQueued("z", inputs)
		status := mocks.ExecutableNodeStatus{}
		status.EXPECT().GetOutputDir().Return(storage.DataReference("s3://foo/bar"))
		status.EXPECT().GetParentNodeID().Return(nil)
		node := mocks.ExecutableNode{}
		node.EXPECT().GetID().Return("n")
		node.EXPECT().GetName().Return("nodey")
		node.EXPECT().GetKind().Return(v1alpha1.NodeKindTask)
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
