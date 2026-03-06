package k8s

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/flyteorg/flyte/v2/gen/go/flyteidl2/common"
)

func TestBuildTaskActionName(t *testing.T) {
	runID := &common.RunIdentifier{
		Org:     "org",
		Project: "project",
		Domain:  "development",
		Name:    "rabc123",
	}

	t.Run("root action uses a0-0 suffix", func(t *testing.T) {
		// Root: action name == run name
		actionID := &common.ActionIdentifier{
			Run:  runID,
			Name: runID.Name,
		}
		assert.Equal(t, "rabc123-a0-0", buildTaskActionName(actionID))
	})

	t.Run("child action includes action name", func(t *testing.T) {
		actionID := &common.ActionIdentifier{
			Run:  runID,
			Name: "train",
		}
		assert.Equal(t, "rabc123-train-0", buildTaskActionName(actionID))
	})
}

func TestBuildNamespace(t *testing.T) {
	t.Run("combines project and domain", func(t *testing.T) {
		runID := &common.RunIdentifier{
			Project: "flytesnacks",
			Domain:  "development",
		}
		assert.Equal(t, "flytesnacks-development", buildNamespace(runID))
	})

	t.Run("different project and domain", func(t *testing.T) {
		runID := &common.RunIdentifier{
			Project: "myproject",
			Domain:  "production",
		}
		assert.Equal(t, "myproject-production", buildNamespace(runID))
	})
}