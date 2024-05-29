package k8s

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/flyteorg/flyte/flyteidl/gen/pb-go/flyteidl/core"
	"github.com/flyteorg/flyte/flyteplugins/go/tasks/pluginmachinery/core/mocks"
)

func Test_newTaskExecutionMetadata(t *testing.T) {
	t.Run("No Secret", func(t *testing.T) {
		existingMetadata := &mocks.TaskExecutionMetadata{}
		existingAnnotations := map[string]string{
			"existingKey": "existingValue",
		}
		existingMetadata.OnGetAnnotations().Return(existingAnnotations)

		existingLabels := map[string]string{
			"existingLabel": "existingLabelValue",
		}
		existingMetadata.OnGetLabels().Return(existingLabels)
		existingMetadata.OnGetSecurityContext().Return(core.SecurityContext{RunAs: &core.Identity{}})

		actual, err := newTaskExecutionMetadata(existingMetadata, &core.TaskTemplate{})
		assert.NoError(t, err)

		assert.Equal(t, existingAnnotations, actual.GetAnnotations())
		assert.Equal(t, existingLabels, actual.GetLabels())
	})

	t.Run("Secret", func(t *testing.T) {
		existingMetadata := &mocks.TaskExecutionMetadata{}
		existingAnnotations := map[string]string{
			"existingKey": "existingValue",
		}
		existingMetadata.OnGetAnnotations().Return(existingAnnotations)

		existingLabels := map[string]string{
			"existingLabel": "existingLabelValue",
		}
		existingMetadata.OnGetLabels().Return(existingLabels)
		existingMetadata.OnGetSecurityContext().Return(core.SecurityContext{RunAs: &core.Identity{}})

		actual, err := newTaskExecutionMetadata(existingMetadata, &core.TaskTemplate{
			SecurityContext: &core.SecurityContext{
				Secrets: []*core.Secret{
					{
						Group:            "my_group",
						Key:              "my_key",
						MountRequirement: core.Secret_ENV_VAR,
					},
				},
			},
		})
		assert.NoError(t, err)

		assert.Equal(t, map[string]string{
			"existingKey":      "existingValue",
			"flyte.secrets/s0": "m4zg54lqhiqce2lzl4txe22voarau12fpe4caitnpfpwwzlzeifg122vnz1f53tfof1ws3tfnvsw34b1ebcu3vs6kzavecq",
		}, actual.GetAnnotations())

		assert.Equal(t, map[string]string{
			"existingLabel":        "existingLabelValue",
			"inject-flyte-secrets": "true",
		}, actual.GetLabels())
	})

	t.Run("Inject exec identity", func(t *testing.T) {

		existingMetadata := &mocks.TaskExecutionMetadata{}
		existingAnnotations := map[string]string{}
		existingMetadata.OnGetAnnotations().Return(existingAnnotations)

		existingMetadata.OnGetSecurityContext().Return(core.SecurityContext{RunAs: &core.Identity{ExecutionIdentity: "test-exec-identity"}})

		existingLabels := map[string]string{
			"existingLabel": "existingLabelValue",
		}
		existingMetadata.OnGetLabels().Return(existingLabels)

		actual, err := newTaskExecutionMetadata(existingMetadata, &core.TaskTemplate{})
		assert.NoError(t, err)

		assert.Equal(t, 2, len(actual.GetLabels()))
		assert.Equal(t, "test-exec-identity", actual.GetLabels()[executionIdentityVariable])
	})
	t.Run("Inject exec identity K8s label sanitation", func(t *testing.T) {

		existingMetadata := &mocks.TaskExecutionMetadata{}
		existingAnnotations := map[string]string{}
		existingMetadata.OnGetAnnotations().Return(existingAnnotations)

		existingMetadata.OnGetSecurityContext().Return(core.SecurityContext{RunAs: &core.Identity{ExecutionIdentity: "name@company.com"}})

		existingLabels := map[string]string{
			"existingLabel": "existingLabelValue",
		}
		existingMetadata.OnGetLabels().Return(existingLabels)

		actual, err := newTaskExecutionMetadata(existingMetadata, &core.TaskTemplate{})
		assert.NoError(t, err)

		assert.Equal(t, 2, len(actual.GetLabels()))
		assert.Equal(t, "name-company-com", actual.GetLabels()[executionIdentityVariable])
	})
}

func Test_newTaskExecutionContext(t *testing.T) {
	existing := &mocks.TaskExecutionContext{}
	existing.OnTaskExecutionMetadata().Panic("Unexpected")

	newMetadata := &mocks.TaskExecutionMetadata{}
	actualCtx := newTaskExecutionContext(existing, newMetadata)
	assert.Equal(t, newMetadata, actualCtx.TaskExecutionMetadata())
}
