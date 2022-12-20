package flytek8s

import (
	"testing"

	"github.com/flyteorg/flyteidl/gen/pb-go/flyteidl/core"
	"github.com/flyteorg/flyteplugins/go/tasks/pluginmachinery/core/mocks"

	"github.com/stretchr/testify/assert"
)

func TestGetServiceAccountNameFromTaskExecutionMetadata(t *testing.T) {
	mockTaskExecMetadata := mocks.TaskExecutionMetadata{}
	mockTaskExecMetadata.OnGetSecurityContext().Return(core.SecurityContext{
		RunAs: &core.Identity{K8SServiceAccount: "service-account"},
	})
	result := GetServiceAccountNameFromTaskExecutionMetadata(&mockTaskExecMetadata)
	assert.Equal(t, "service-account", result)
}

func TestGetServiceAccountNameFromServiceAccount(t *testing.T) {
	mockTaskExecMetadata := mocks.TaskExecutionMetadata{}
	mockTaskExecMetadata.OnGetSecurityContext().Return(core.SecurityContext{})
	mockTaskExecMetadata.OnGetK8sServiceAccount().Return("service-account")
	result := GetServiceAccountNameFromTaskExecutionMetadata(&mockTaskExecMetadata)
	assert.Equal(t, "service-account", result)
}
