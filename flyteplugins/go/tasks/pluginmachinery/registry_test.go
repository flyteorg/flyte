package pluginmachinery

import (
	"testing"

	rayv1 "github.com/ray-project/kuberay/ray-operator/apis/ray/v1"
	"github.com/stretchr/testify/assert"

	pluginsCore "github.com/flyteorg/flyte/v2/flyteplugins/go/tasks/pluginmachinery/core"
	"github.com/flyteorg/flyte/v2/flyteplugins/go/tasks/pluginmachinery/k8s"
)

// stubPlugin / stubClusterPlugin embed the interface to get nil-method stubs sufficient for
// construction-only validation tests.
type stubPlugin struct{ k8s.Plugin }
type stubClusterPlugin struct{ k8s.ClusterPlugin }

func TestRegisterK8sPlugin_ClusterPluginValidation(t *testing.T) {
	t.Run("cluster plugin requires ClusterResourceToWatch", func(t *testing.T) {
		reg := &taskPluginRegistry{}
		assert.Panics(t, func() {
			reg.RegisterK8sPlugin(k8s.PluginEntry{
				ID:                  "x",
				RegisteredTaskTypes: []pluginsCore.TaskType{"x"},
				ResourceToWatch:     &rayv1.RayJob{},
				ClusterPlugin:       stubClusterPlugin{},
			})
		})
	})

	t.Run("cannot set both Plugin and ClusterPlugin", func(t *testing.T) {
		reg := &taskPluginRegistry{}
		assert.Panics(t, func() {
			reg.RegisterK8sPlugin(k8s.PluginEntry{
				ID:                     "x",
				RegisteredTaskTypes:    []pluginsCore.TaskType{"x"},
				ResourceToWatch:        &rayv1.RayJob{},
				ClusterResourceToWatch: &rayv1.RayCluster{},
				Plugin:                 stubPlugin{},
				ClusterPlugin:          stubClusterPlugin{},
			})
		})
	})

	t.Run("valid cluster plugin registers", func(t *testing.T) {
		reg := &taskPluginRegistry{}
		assert.NotPanics(t, func() {
			reg.RegisterK8sPlugin(k8s.PluginEntry{
				ID:                     "x",
				RegisteredTaskTypes:    []pluginsCore.TaskType{"x"},
				ResourceToWatch:        &rayv1.RayJob{},
				ClusterResourceToWatch: &rayv1.RayCluster{},
				ClusterPlugin:          stubClusterPlugin{},
			})
		})
		assert.Len(t, reg.GetK8sPlugins(), 1)
	})
}
