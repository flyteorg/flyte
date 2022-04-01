package nodes

import (
	"testing"

	"github.com/flyteorg/flyteidl/gen/pb-go/flyteidl/core"
	"github.com/flyteorg/flyteidl/gen/pb-go/flyteidl/event"
	"github.com/flyteorg/flytepropeller/pkg/apis/flyteworkflow/v1alpha1"
	"github.com/flyteorg/flytepropeller/pkg/apis/flyteworkflow/v1alpha1/mocks"
	mocks2 "github.com/flyteorg/flytepropeller/pkg/controller/executors/mocks"
	"github.com/flyteorg/flytepropeller/pkg/controller/nodes/handler"
	"github.com/flyteorg/flytestdlib/storage"
	"github.com/stretchr/testify/assert"
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
		}, info, "inputPath", &status, v1alpha1.EventVersion2, &parentInfo, &node, "clusterID", v1alpha1.DynamicNodePhaseParentFinalized)
		assert.NoError(t, err)
		assert.True(t, nev.IsDynamic)
		assert.True(t, nev.IsParent)
		assert.Equal(t, nodeExecutionEventVersion, nev.EventVersion)
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
		}, info, "inputPath", &status, v1alpha1.EventVersion2, &parentInfo, &node, "clusterID", v1alpha1.DynamicNodePhaseNone)
		assert.NoError(t, err)
		assert.False(t, nev.IsDynamic)
		assert.True(t, nev.IsParent)
		assert.Equal(t, nodeExecutionEventVersion, nev.EventVersion)
	})
}
