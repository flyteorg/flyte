package flytek8s

import (
	"context"
	"testing"

	"github.com/lyft/flyteidl/gen/pb-go/flyteidl/core"
	"github.com/lyft/flytestdlib/storage"
	"github.com/stretchr/testify/assert"
	"k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	metaV1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/lyft/flyteplugins/go/tasks/v1/flytek8s/config"
	"github.com/lyft/flyteplugins/go/tasks/v1/types"
	"github.com/lyft/flyteplugins/go/tasks/v1/types/mocks"
)

func dummyContainerTaskContext(resources *v1.ResourceRequirements) types.TaskContext {
	taskCtx := &mocks.TaskContext{}
	taskCtx.On("GetNamespace").Return("test-namespace")
	taskCtx.On("GetAnnotations").Return(map[string]string{"annotation-1": "val1"})
	taskCtx.On("GetLabels").Return(map[string]string{"label-1": "val1"})
	taskCtx.On("GetOwnerReference").Return(metaV1.OwnerReference{
		Kind: "node",
		Name: "blah",
	})
	taskCtx.On("GetDataDir").Return(storage.DataReference("/data/"))
	taskCtx.On("GetInputsFile").Return(storage.DataReference("/input"))
	taskCtx.On("GetK8sServiceAccount").Return("service-account")

	tID := &mocks.TaskExecutionID{}
	tID.On("GetID").Return(core.TaskExecutionIdentifier{
		NodeExecutionId: &core.NodeExecutionIdentifier{
			ExecutionId: &core.WorkflowExecutionIdentifier{
				Name:    "my_name",
				Project: "my_project",
				Domain:  "my_domain",
			},
		},
	})
	tID.On("GetGeneratedName").Return("some-acceptable-name")
	taskCtx.On("GetTaskExecutionID").Return(tID)

	to := &mocks.TaskOverrides{}
	to.On("GetResources").Return(resources)
	taskCtx.On("GetOverrides").Return(to)

	return taskCtx
}

func TestToK8sPod(t *testing.T) {
	ctx := context.TODO()
	command := []string{"command"}
	args := []string{"{{.Input}}"}
	container := &core.Container{
		Command: command,
		Args:    args,
	}

	tolGPU := v1.Toleration{
		Key:      "flyte/gpu",
		Value:    "dedicated",
		Operator: v1.TolerationOpEqual,
		Effect:   v1.TaintEffectNoSchedule,
	}

	tolStorage := v1.Toleration{
		Key:      "storage",
		Value:    "dedicated",
		Operator: v1.TolerationOpExists,
		Effect:   v1.TaintEffectNoSchedule,
	}

	assert.NoError(t, config.SetK8sPluginConfig(&config.K8sPluginConfig{
		ResourceTolerations: map[v1.ResourceName][]v1.Toleration{
			v1.ResourceStorage: {tolStorage},
			ResourceNvidiaGPU:  {tolGPU},
		}}),
	)

	t.Run("WithGPU", func(t *testing.T) {
		x := dummyContainerTaskContext(&v1.ResourceRequirements{
			Limits: v1.ResourceList{
				v1.ResourceCPU:     resource.MustParse("1024m"),
				v1.ResourceStorage: resource.MustParse("100M"),
				ResourceNvidiaGPU:  resource.MustParse("1"),
			},
			Requests: v1.ResourceList{
				v1.ResourceCPU:     resource.MustParse("1024m"),
				v1.ResourceStorage: resource.MustParse("100M"),
			},
		})

		p, err := ToK8sPod(ctx, x, container, nil)
		assert.NoError(t, err)
		assert.Equal(t, len(p.Tolerations), 1)
	})

	t.Run("NoGPU", func(t *testing.T) {
		x := dummyContainerTaskContext(&v1.ResourceRequirements{
			Limits: v1.ResourceList{
				v1.ResourceCPU:     resource.MustParse("1024m"),
				v1.ResourceStorage: resource.MustParse("100M"),
			},
			Requests: v1.ResourceList{
				v1.ResourceCPU:     resource.MustParse("1024m"),
				v1.ResourceStorage: resource.MustParse("100M"),
			},
		})

		p, err := ToK8sPod(ctx, x, container, nil)
		assert.NoError(t, err)
		assert.Equal(t, len(p.Tolerations), 0)
		assert.Equal(t, "some-acceptable-name", p.Containers[0].Name)
	})
}

func TestDemystifyPending(t *testing.T) {

	t.Run("PodNotScheduled", func(t *testing.T) {
		s := v1.PodStatus{
			Phase: v1.PodPending,
			Conditions: []v1.PodCondition{
				{
					Type:   v1.PodScheduled,
					Status: v1.ConditionFalse,
				},
			},
		}
		taskStatus, err := DemystifyPending(s)
		assert.NoError(t, err)
		assert.Equal(t, types.TaskPhaseQueued, taskStatus.Phase)
	})

	t.Run("PodUnschedulable", func(t *testing.T) {
		s := v1.PodStatus{
			Phase: v1.PodPending,
			Conditions: []v1.PodCondition{
				{
					Type:   v1.PodReasonUnschedulable,
					Status: v1.ConditionFalse,
				},
			},
		}
		taskStatus, err := DemystifyPending(s)
		assert.NoError(t, err)
		assert.Equal(t, types.TaskPhaseQueued, taskStatus.Phase)
	})

	t.Run("PodNotScheduled", func(t *testing.T) {
		s := v1.PodStatus{
			Phase: v1.PodPending,
			Conditions: []v1.PodCondition{
				{
					Type:   v1.PodScheduled,
					Status: v1.ConditionTrue,
				},
			},
		}
		taskStatus, err := DemystifyPending(s)
		assert.NoError(t, err)
		assert.Equal(t, types.TaskPhaseQueued, taskStatus.Phase)
	})

	t.Run("PodUnschedulable", func(t *testing.T) {
		s := v1.PodStatus{
			Phase: v1.PodPending,
			Conditions: []v1.PodCondition{
				{
					Type:   v1.PodReasonUnschedulable,
					Status: v1.ConditionUnknown,
				},
			},
		}
		taskStatus, err := DemystifyPending(s)
		assert.NoError(t, err)
		assert.Equal(t, types.TaskPhaseQueued, taskStatus.Phase)
	})

	s := v1.PodStatus{
		Phase: v1.PodPending,
		Conditions: []v1.PodCondition{
			{
				Type:   v1.PodReady,
				Status: v1.ConditionFalse,
			},
			{
				Type:   v1.PodReasonUnschedulable,
				Status: v1.ConditionUnknown,
			},
			{
				Type:   v1.PodScheduled,
				Status: v1.ConditionTrue,
			},
		},
	}

	t.Run("ContainerCreating", func(t *testing.T) {
		s.ContainerStatuses = []v1.ContainerStatus{
			{
				Ready: false,
				State: v1.ContainerState{
					Waiting: &v1.ContainerStateWaiting{
						Reason:  "ContainerCreating",
						Message: "this is not an error",
					},
				},
			},
		}
		taskStatus, err := DemystifyPending(s)
		assert.NoError(t, err)
		assert.Equal(t, types.TaskPhaseQueued, taskStatus.Phase)
	})

	t.Run("ErrImagePull", func(t *testing.T) {
		s.ContainerStatuses = []v1.ContainerStatus{
			{
				Ready: false,
				State: v1.ContainerState{
					Waiting: &v1.ContainerStateWaiting{
						Reason:  "ErrImagePull",
						Message: "this is not an error",
					},
				},
			},
		}
		taskStatus, err := DemystifyPending(s)
		assert.NoError(t, err)
		assert.Equal(t, types.TaskPhaseQueued, taskStatus.Phase)
	})

	t.Run("PodInitializing", func(t *testing.T) {
		s.ContainerStatuses = []v1.ContainerStatus{
			{
				Ready: false,
				State: v1.ContainerState{
					Waiting: &v1.ContainerStateWaiting{
						Reason:  "PodInitializing",
						Message: "this is not an error",
					},
				},
			},
		}
		taskStatus, err := DemystifyPending(s)
		assert.NoError(t, err)
		assert.Equal(t, types.TaskPhaseQueued, taskStatus.Phase)
	})

	t.Run("ImagePullBackOff", func(t *testing.T) {
		s.ContainerStatuses = []v1.ContainerStatus{
			{
				Ready: false,
				State: v1.ContainerState{
					Waiting: &v1.ContainerStateWaiting{
						Reason:  "ImagePullBackOff",
						Message: "this is an error",
					},
				},
			},
		}
		taskStatus, err := DemystifyPending(s)
		assert.NoError(t, err)
		assert.Equal(t, types.TaskPhaseRetryableFailure, taskStatus.Phase)
	})

	t.Run("InvalidImageName", func(t *testing.T) {
		s.ContainerStatuses = []v1.ContainerStatus{
			{
				Ready: false,
				State: v1.ContainerState{
					Waiting: &v1.ContainerStateWaiting{
						Reason:  "InvalidImageName",
						Message: "this is an error",
					},
				},
			},
		}
		taskStatus, err := DemystifyPending(s)
		assert.NoError(t, err)
		assert.Equal(t, types.TaskPhaseRetryableFailure, taskStatus.Phase)
	})

	t.Run("RegistryUnavailable", func(t *testing.T) {
		s.ContainerStatuses = []v1.ContainerStatus{
			{
				Ready: false,
				State: v1.ContainerState{
					Waiting: &v1.ContainerStateWaiting{
						Reason:  "RegistryUnavailable",
						Message: "this is an error",
					},
				},
			},
		}
		taskStatus, err := DemystifyPending(s)
		assert.NoError(t, err)
		assert.Equal(t, types.TaskPhaseRetryableFailure, taskStatus.Phase)
	})

	t.Run("RandomError", func(t *testing.T) {
		s.ContainerStatuses = []v1.ContainerStatus{
			{
				Ready: false,
				State: v1.ContainerState{
					Waiting: &v1.ContainerStateWaiting{
						Reason:  "RandomError",
						Message: "this is an error",
					},
				},
			},
		}
		taskStatus, err := DemystifyPending(s)
		assert.NoError(t, err)
		assert.Equal(t, types.TaskPhaseRetryableFailure, taskStatus.Phase)
	})
}
