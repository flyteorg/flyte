package launchplan

import (
	"context"
	"testing"

	"github.com/flyteorg/flyteidl/gen/pb-go/flyteidl/core"
	"github.com/stretchr/testify/assert"
)

func TestFailFastWorkflowLauncher(t *testing.T) {
	ctx := context.TODO()
	f := NewFailFastLaunchPlanExecutor()
	t.Run("getStatus", func(t *testing.T) {
		a, _, err := f.GetStatus(ctx, &core.WorkflowExecutionIdentifier{
			Project: "p",
			Domain:  "d",
			Name:    "n",
		})
		assert.Nil(t, a)
		assert.Error(t, err)
	})

	t.Run("launch", func(t *testing.T) {
		err := f.Launch(ctx, LaunchContext{
			ParentNodeExecution: &core.NodeExecutionIdentifier{
				NodeId: "node-id",
				ExecutionId: &core.WorkflowExecutionIdentifier{
					Project: "p",
					Domain:  "d",
					Name:    "n",
				},
			},
		}, &core.WorkflowExecutionIdentifier{
			Project: "p",
			Domain:  "d",
			Name:    "n",
		}, &core.Identifier{},
			nil)
		assert.Error(t, err)
	})

	t.Run("kill", func(t *testing.T) {
		err := f.Kill(ctx, &core.WorkflowExecutionIdentifier{
			Project: "p",
			Domain:  "d",
			Name:    "n",
		}, "reason")
		assert.NoError(t, err)
	})
}
