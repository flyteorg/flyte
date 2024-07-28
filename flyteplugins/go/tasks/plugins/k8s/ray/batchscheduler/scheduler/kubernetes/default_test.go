package kubernetes

import (
	"testing"

	"github.com/stretchr/testify/assert"
	v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/flyteorg/flyte/flyteidl/gen/pb-go/flyteidl/plugins"
	schedulerConfig "github.com/flyteorg/flyte/flyteplugins/go/tasks/plugins/k8s/ray/batchscheduler/config"
)

var (
	metadata = &metav1.ObjectMeta{
		Labels:      map[string]string{"others": "extra"},
		Annotations: map[string]string{"others": "extra"},
	}
	res = v1.ResourceList{
		"cpu":    resource.MustParse("500m"),
		"memory": resource.MustParse("1Gi"),
	}
	podSpec = &v1.PodSpec{
		Containers: []v1.Container{
			{
				Resources: v1.ResourceRequirements{
					Requests: res,
				},
			},
		},
		NodeSelector:              nil,
		Affinity:                  nil,
		TopologySpreadConstraints: nil,
	}
)

func TestNewDefaultPlugin(t *testing.T) {
	t.Run("New default scheduler plugin", func(t *testing.T) {
		p := NewDefaultPlugin()
		assert.NotNil(t, p)
		assert.Equal(t, DefaultScheduler, p.GetSchedulerName())
	})
}

func TestParseJob(t *testing.T) {
	t.Run("Default scheduler plugin parse job", func(t *testing.T) {
		p := schedulerConfig.NewConfig()
		rayWorkersSpec := []*plugins.WorkerGroupSpec{
			{
				GroupName:   "g1",
				Replicas:    int32(2),
				MinReplicas: int32(1),
				MaxReplicas: int32(3),
				RayStartParams: map[string]string{
					"parameters": "specific parameters",
				},
			},
		}
		index := 0
		err := NewDefaultPlugin().ParseJob(&p, metadata, rayWorkersSpec, podSpec, index)
		assert.Nil(t, err)
		workerspec := rayWorkersSpec[0]
		assert.Equal(t, "g1", workerspec.GroupName)
		assert.Equal(t, int32(2), workerspec.Replicas)
		assert.Equal(t, int32(1), workerspec.MinReplicas)
		assert.Equal(t, int32(3), workerspec.MaxReplicas)
		assert.Equal(t, map[string]string{"parameters": "specific parameters"}, workerspec.RayStartParams)
		assert.Equal(t, map[string]string{"others": "extra"}, metadata.Annotations)
		assert.Equal(t, map[string]string{"others": "extra"}, metadata.Labels)
		assert.Equal(t, res, podSpec.Containers[index].Resources.Requests)
		assert.Equal(t, "", p.GetScheduler())
		assert.Equal(t, "", p.GetParameters())
	})
}

func TestProcessHead(t *testing.T) {
	t.Run("Default scheduler plugin process head", func(t *testing.T) {
		index := 0
		NewDefaultPlugin().ProcessHead(metadata, podSpec, index)
		assert.Equal(t, map[string]string{"others": "extra"}, metadata.Annotations)
		assert.Equal(t, map[string]string{"others": "extra"}, metadata.Labels)
		assert.Equal(t, res, podSpec.Containers[index].Resources.Requests)
	})
}

func TestProcessWorker(t *testing.T) {
	t.Run("Default scheduler plugin preprocess worker", func(t *testing.T) {
		index := 0
		NewDefaultPlugin().ProcessWorker(metadata, podSpec, index)
		assert.Equal(t, map[string]string{"others": "extra"}, metadata.Annotations)
		assert.Equal(t, map[string]string{"others": "extra"}, metadata.Labels)
		assert.Equal(t, res, podSpec.Containers[index].Resources.Requests)
	})
}

func TestAfterProcess(t *testing.T) {
	t.Run("Default scheduler plugin afterly process worker", func(t *testing.T) {
		NewDefaultPlugin().AfterProcess(metadata)
		assert.Equal(t, map[string]string{"others": "extra"}, metadata.Annotations)
		assert.Equal(t, map[string]string{"others": "extra"}, metadata.Labels)
	})
}
