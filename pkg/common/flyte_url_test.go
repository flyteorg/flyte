package common

import (
	"testing"

	"github.com/flyteorg/flyteidl/gen/pb-go/flyteidl/core"
	"github.com/golang/protobuf/proto"
	"github.com/stretchr/testify/assert"
)

func TestParseFlyteUrl(t *testing.T) {
	t.Run("valid", func(t *testing.T) {
		ne, attempt, kind, err := ParseFlyteURL("flyte://v1/fs/dev/abc/n0/0/o")
		assert.NoError(t, err)
		assert.Equal(t, 0, *attempt)
		assert.Equal(t, ArtifactTypeO, kind)
		assert.True(t, proto.Equal(&core.NodeExecutionIdentifier{
			NodeId: "n0",
			ExecutionId: &core.WorkflowExecutionIdentifier{
				Project: "fs",
				Domain:  "dev",
				Name:    "abc",
			},
		}, &ne))
		ne, attempt, kind, err = ParseFlyteURL("flyte://v1/fs/dev/abc/n0/i")
		assert.NoError(t, err)
		assert.Nil(t, attempt)
		assert.Equal(t, ArtifactTypeI, kind)
		assert.True(t, proto.Equal(&core.NodeExecutionIdentifier{
			NodeId: "n0",
			ExecutionId: &core.WorkflowExecutionIdentifier{
				Project: "fs",
				Domain:  "dev",
				Name:    "abc",
			},
		}, &ne))

		ne, attempt, kind, err = ParseFlyteURL("flyte://v1/fs/dev/abc/n0/d")
		assert.NoError(t, err)
		assert.Nil(t, attempt)
		assert.Equal(t, ArtifactTypeD, kind)
		assert.True(t, proto.Equal(&core.NodeExecutionIdentifier{
			NodeId: "n0",
			ExecutionId: &core.WorkflowExecutionIdentifier{
				Project: "fs",
				Domain:  "dev",
				Name:    "abc",
			},
		}, &ne))

		ne, attempt, kind, err = ParseFlyteURL("flyte://v1/fs/dev/abc/n0-dn0-9-n0-n0/d")
		assert.NoError(t, err)
		assert.Nil(t, attempt)
		assert.Equal(t, ArtifactTypeD, kind)
		assert.True(t, proto.Equal(&core.NodeExecutionIdentifier{
			NodeId: "n0-dn0-9-n0-n0",
			ExecutionId: &core.WorkflowExecutionIdentifier{
				Project: "fs",
				Domain:  "dev",
				Name:    "abc",
			},
		}, &ne))
	})

	t.Run("invalid", func(t *testing.T) {
		// more than one character
		_, attempt, kind, err := ParseFlyteURL("flyte://v1/fs/dev/abc/n0/0/od")
		assert.Error(t, err)
		assert.Nil(t, attempt)
		assert.Equal(t, ArtifactTypeUndefined, kind)

		_, attempt, kind, err = ParseFlyteURL("flyte://v1/fs/dev/abc/n0/input")
		assert.Error(t, err)
		assert.Nil(t, attempt)
		assert.Equal(t, ArtifactTypeUndefined, kind)

		// non integer for attempt
		_, attempt, kind, err = ParseFlyteURL("flyte://v1/fs/dev/ab/n0/a/i")
		assert.Error(t, err)
		assert.Nil(t, attempt)
		assert.Equal(t, ArtifactTypeUndefined, kind)
	})
}

func TestFlyteURLsFromNodeExecutionID(t *testing.T) {
	t.Run("with deck", func(t *testing.T) {
		ne := core.NodeExecutionIdentifier{
			NodeId: "n0-dn0-n1",
			ExecutionId: &core.WorkflowExecutionIdentifier{
				Project: "fs",
				Domain:  "dev",
				Name:    "abc",
			},
		}
		urls := FlyteURLsFromNodeExecutionID(ne, true)
		assert.Equal(t, "flyte://v1/fs/dev/abc/n0-dn0-n1/i", urls.GetInputs())
		assert.Equal(t, "flyte://v1/fs/dev/abc/n0-dn0-n1/o", urls.GetOutputs())
		assert.Equal(t, "flyte://v1/fs/dev/abc/n0-dn0-n1/d", urls.GetDeck())
	})

	t.Run("without deck", func(t *testing.T) {
		ne := core.NodeExecutionIdentifier{
			NodeId: "n0-dn0-n1",
			ExecutionId: &core.WorkflowExecutionIdentifier{
				Project: "fs",
				Domain:  "dev",
				Name:    "abc",
			},
		}
		urls := FlyteURLsFromNodeExecutionID(ne, false)
		assert.Equal(t, "flyte://v1/fs/dev/abc/n0-dn0-n1/i", urls.GetInputs())
		assert.Equal(t, "flyte://v1/fs/dev/abc/n0-dn0-n1/o", urls.GetOutputs())
		assert.Equal(t, "", urls.GetDeck())
	})
}

func TestFlyteURLsFromTaskExecutionID(t *testing.T) {
	t.Run("with deck", func(t *testing.T) {
		te := core.TaskExecutionIdentifier{
			TaskId: &core.Identifier{
				ResourceType: core.ResourceType_TASK,
				Project:      "fs",
				Domain:       "dev",
				Name:         "abc",
				Version:      "v1",
			},
			NodeExecutionId: &core.NodeExecutionIdentifier{
				NodeId: "n0",
				ExecutionId: &core.WorkflowExecutionIdentifier{
					Project: "fs",
					Domain:  "dev",
					Name:    "abc",
				},
			},
			RetryAttempt: 1,
		}
		urls := FlyteURLsFromTaskExecutionID(te, true)
		assert.Equal(t, "flyte://v1/fs/dev/abc/n0/1/i", urls.GetInputs())
		assert.Equal(t, "flyte://v1/fs/dev/abc/n0/1/o", urls.GetOutputs())
		assert.Equal(t, "flyte://v1/fs/dev/abc/n0/1/d", urls.GetDeck())
	})

	t.Run("without deck", func(t *testing.T) {
		te := core.TaskExecutionIdentifier{
			TaskId: &core.Identifier{
				ResourceType: core.ResourceType_TASK,
				Project:      "fs",
				Domain:       "dev",
				Name:         "abc",
				Version:      "v1",
			},
			NodeExecutionId: &core.NodeExecutionIdentifier{
				NodeId: "n0",
				ExecutionId: &core.WorkflowExecutionIdentifier{
					Project: "fs",
					Domain:  "dev",
					Name:    "abc",
				},
			},
		}
		urls := FlyteURLsFromTaskExecutionID(te, false)
		assert.Equal(t, "flyte://v1/fs/dev/abc/n0/0/i", urls.GetInputs())
		assert.Equal(t, "flyte://v1/fs/dev/abc/n0/0/o", urls.GetOutputs())
		assert.Equal(t, "", urls.GetDeck())
	})
}

func TestMatchRegexDirectly(t *testing.T) {
	result := MatchRegex(re, "flyte://v1/fs/dev/abc/n0-dn0-9-n0-n0/i")
	assert.Equal(t, "", result["attempt"])

	result = MatchRegex(re, "flyteff://v2/fs/dfdsaev/abc/n0-dn0-9-n0-n0/i")
	assert.Nil(t, result)
}
