package plugin

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/require"
	"google.golang.org/protobuf/proto"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"

	executorv1 "github.com/flyteorg/flyte/v2/executor/api/v1"
	executorconfig "github.com/flyteorg/flyte/v2/executor/pkg/config"
	"github.com/flyteorg/flyte/v2/flyteplugins/go/tasks/pluginmachinery/flytek8s"
	"github.com/flyteorg/flyte/v2/gen/go/flyteidl2/core"
)

func TestTaskNamespace(t *testing.T) {
	cfg := executorconfig.GetConfig()
	original := cfg.AllowedTaskNamespaces
	t.Cleanup(func() { cfg.AllowedTaskNamespaces = original })
	for _, tc := range []struct {
		name, namespace, want string
		allowed               []string
		wantError             bool
	}{
		{name: "unset", want: "flyte"},
		{name: "same namespace", namespace: "flyte", want: "flyte"},
		{name: "allowed", namespace: "gpu-workers", allowed: []string{"gpu-workers"}, want: "gpu-workers"},
		{name: "wildcard", namespace: "gpu-workers", allowed: []string{"*"}, want: "gpu-workers"},
		{name: "disabled by default", namespace: "gpu-workers", wantError: true},
		{name: "not allowed", namespace: "other-workers", allowed: []string{"gpu-workers"}, wantError: true},
		{name: "invalid", namespace: "Bad_namespace", allowed: []string{"*"}, wantError: true},
		{name: "too long", namespace: strings.Repeat("a", 64), allowed: []string{"*"}, wantError: true},
	} {
		t.Run(tc.name, func(t *testing.T) {
			cfg.AllowedTaskNamespaces = tc.allowed
			data, err := proto.Marshal(&core.TaskTemplate{Config: map[string]string{flytek8s.TaskNamespaceConfigKey: tc.namespace}})
			require.NoError(t, err)
			ta := &executorv1.TaskAction{ObjectMeta: metav1.ObjectMeta{Namespace: "flyte"}, Spec: executorv1.TaskActionSpec{TaskTemplate: data}}
			err = ValidateTaskNamespace(ta)
			if tc.wantError {
				require.Error(t, err)
				return
			}
			require.NoError(t, err)
			got, err := TaskNamespace(ta)
			require.NoError(t, err)
			require.Equal(t, tc.want, got)
		})
	}
}

func TestCrossNamespaceMetadataKeepsControlIdentity(t *testing.T) {
	cfg := executorconfig.GetConfig()
	original := cfg.AllowedTaskNamespaces
	cfg.AllowedTaskNamespaces = []string{"gpu-workers"}
	t.Cleanup(func() { cfg.AllowedTaskNamespaces = original })
	data, err := proto.Marshal(&core.TaskTemplate{Config: map[string]string{flytek8s.TaskNamespaceConfigKey: "gpu-workers"}})
	require.NoError(t, err)
	ta := &executorv1.TaskAction{
		ObjectMeta: metav1.ObjectMeta{Name: "task-a0", Namespace: "flyte", UID: types.UID("task-uid"),
			Annotations: map[string]string{flytek8s.TaskActionNamespaceAnnotation: "spoofed", flytek8s.TaskActionNameAnnotation: "spoofed"}},
		Spec: executorv1.TaskActionSpec{TaskTemplate: data},
	}
	metadata, err := NewTaskExecutionMetadata(ta)
	require.NoError(t, err)
	require.Equal(t, "gpu-workers", metadata.GetNamespace())
	require.Equal(t, types.NamespacedName{Namespace: "flyte", Name: "task-a0"}, metadata.GetOwnerID())
	require.Empty(t, metadata.GetOwnerReference().Name, "cross-namespace owner references are invalid")
	require.Equal(t, "flyte", metadata.GetAnnotations()[flytek8s.TaskActionNamespaceAnnotation])
	require.Equal(t, "task-a0", metadata.GetAnnotations()[flytek8s.TaskActionNameAnnotation])
	require.Equal(t, "spoofed", ta.Annotations[flytek8s.TaskActionNamespaceAnnotation], "metadata projection must not mutate the TaskAction")

	ta.Spec.TaskTemplate = nil
	metadata, err = NewTaskExecutionMetadata(ta)
	require.NoError(t, err)
	require.Equal(t, "flyte", metadata.GetNamespace())
	require.Equal(t, ta.UID, metadata.GetOwnerReference().UID)
	require.NotContains(t, metadata.GetAnnotations(), flytek8s.TaskActionNamespaceAnnotation)
}
