package k8splugins

import (
	"context"
	"io/ioutil"
	"os"
	"path"
	"testing"

	"github.com/golang/protobuf/jsonpb"
	"github.com/golang/protobuf/ptypes/struct"
	"github.com/lyft/flyteidl/gen/pb-go/flyteidl/core"
	"github.com/lyft/flyteidl/gen/pb-go/flyteidl/plugins"
	"github.com/stretchr/testify/assert"
	"k8s.io/api/core/v1"

	"github.com/lyft/flyteplugins/go/tasks/v1/flytek8s"
	"github.com/lyft/flyteplugins/go/tasks/v1/flytek8s/config"
	"github.com/lyft/flyteplugins/go/tasks/v1/types"
	"github.com/lyft/flyteplugins/go/tasks/v1/utils"
)

const ResourceNvidiaGPU = "nvidia.com/gpu"

func getSidecarTaskTemplateForTest(sideCarJob plugins.SidecarJob) *core.TaskTemplate {
	sidecarJSON, err := utils.MarshalToString(&sideCarJob)
	if err != nil {
		panic(err)
	}
	structObj := structpb.Struct{}
	err = jsonpb.UnmarshalString(sidecarJSON, &structObj)
	if err != nil {
		panic(err)
	}
	return &core.TaskTemplate{
		Custom: &structObj,
	}
}

func TestBuildSidecarResource(t *testing.T) {
	dir, err := os.Getwd()
	if err != nil {
		t.Fatal(err)
	}
	sidecarCustomJSON, err := ioutil.ReadFile(path.Join(dir, "mocks", "sidecar_custom"))
	if err != nil {
		t.Fatal(sidecarCustomJSON)
	}
	sidecarCustom := structpb.Struct{}
	if err := jsonpb.UnmarshalString(string(sidecarCustomJSON), &sidecarCustom); err != nil {
		t.Fatal(err)
	}
	task := core.TaskTemplate{
		Custom: &sidecarCustom,
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
		},
	}))
	handler := &sidecarResourceHandler{}
	taskCtx := dummyContainerTaskContext(resourceRequirements)
	resource, err := handler.BuildResource(context.TODO(), taskCtx, &task, nil)
	assert.Nil(t, err)
	assert.EqualValues(t, map[string]string{
		primaryContainerKey: "a container",
	}, resource.GetAnnotations())
	assert.Contains(t, resource.(*v1.Pod).Spec.Tolerations, tolGPU)
}

func TestBuildSidecarResourceMissingPrimary(t *testing.T) {
	sideCarJob := plugins.SidecarJob{
		PrimaryContainerName: "PrimaryContainer",
		PodSpec: &v1.PodSpec{
			Containers: []v1.Container{
				{
					Name: "SecondaryContainer",
				},
			},
		},
	}

	task := getSidecarTaskTemplateForTest(sideCarJob)

	handler := &sidecarResourceHandler{}
	taskCtx := dummyContainerTaskContext(resourceRequirements)
	_, err := handler.BuildResource(context.TODO(), taskCtx, task, nil)
	assert.EqualError(t, err,
		"task failed, BadTaskSpecification: invalid Sidecar task, primary container [PrimaryContainer] not defined")
}

func TestGetTaskSidecarStatus(t *testing.T) {
	var testCases = map[v1.PodPhase]types.TaskPhase{
		v1.PodSucceeded:           types.TaskPhaseSucceeded,
		v1.PodFailed:              types.TaskPhaseRetryableFailure,
		v1.PodReasonUnschedulable: types.TaskPhaseQueued,
		v1.PodUnknown:             types.TaskPhaseUnknown,
	}

	for podPhase, expectedTaskPhase := range testCases {
		var resource flytek8s.K8sResource
		resource = &v1.Pod{
			Status: v1.PodStatus{
				Phase: podPhase,
			},
		}
		handler := &sidecarResourceHandler{}
		taskCtx := dummyContainerTaskContext(resourceRequirements)
		status, _, err := handler.GetTaskStatus(context.TODO(), taskCtx, resource)
		assert.Nil(t, err)
		assert.Equal(t, expectedTaskPhase, status.Phase)
	}
}

func TestDemystifiedSidecarStatus_PrimaryFailed(t *testing.T) {
	var resource flytek8s.K8sResource
	resource = &v1.Pod{
		Status: v1.PodStatus{
			Phase: v1.PodRunning,
			ContainerStatuses: []v1.ContainerStatus{
				{
					Name: "Primary",
					State: v1.ContainerState{
						Terminated: &v1.ContainerStateTerminated{
							ExitCode: 1,
						},
					},
				},
			},
		},
	}
	resource.SetAnnotations(map[string]string{
		primaryContainerKey: "Primary",
	})
	handler := &sidecarResourceHandler{}
	taskCtx := dummyContainerTaskContext(resourceRequirements)
	status, _, err := handler.GetTaskStatus(context.TODO(), taskCtx, resource)
	assert.Nil(t, err)
	assert.Equal(t, types.TaskPhaseRetryableFailure, status.Phase)
}

func TestDemystifiedSidecarStatus_PrimarySucceeded(t *testing.T) {
	var resource flytek8s.K8sResource
	resource = &v1.Pod{
		Status: v1.PodStatus{
			Phase: v1.PodRunning,
			ContainerStatuses: []v1.ContainerStatus{
				{
					Name: "Primary",
					State: v1.ContainerState{
						Terminated: &v1.ContainerStateTerminated{
							ExitCode: 0,
						},
					},
				},
			},
		},
	}
	resource.SetAnnotations(map[string]string{
		primaryContainerKey: "Primary",
	})
	handler := &sidecarResourceHandler{}
	taskCtx := dummyContainerTaskContext(resourceRequirements)
	status, _, err := handler.GetTaskStatus(context.TODO(), taskCtx, resource)
	assert.Nil(t, err)
	assert.Equal(t, types.TaskPhaseSucceeded, status.Phase)
}

func TestDemystifiedSidecarStatus_PrimaryRunning(t *testing.T) {
	var resource flytek8s.K8sResource
	resource = &v1.Pod{
		Status: v1.PodStatus{
			Phase: v1.PodRunning,
			ContainerStatuses: []v1.ContainerStatus{
				{
					Name: "Primary",
					State: v1.ContainerState{
						Waiting: &v1.ContainerStateWaiting{
							Reason: "stay patient",
						},
					},
				},
			},
		},
	}
	resource.SetAnnotations(map[string]string{
		primaryContainerKey: "Primary",
	})
	handler := &sidecarResourceHandler{}
	taskCtx := dummyContainerTaskContext(resourceRequirements)
	status, _, err := handler.GetTaskStatus(context.TODO(), taskCtx, resource)
	assert.Nil(t, err)
	assert.Equal(t, types.TaskPhaseRunning, status.Phase)
}

func TestDemystifiedSidecarStatus_PrimaryMissing(t *testing.T) {
	var resource flytek8s.K8sResource
	resource = &v1.Pod{
		Status: v1.PodStatus{
			Phase: v1.PodRunning,
			ContainerStatuses: []v1.ContainerStatus{
				{
					Name: "Secondary",
				},
			},
		},
	}
	resource.SetAnnotations(map[string]string{
		primaryContainerKey: "Primary",
	})
	handler := &sidecarResourceHandler{}
	taskCtx := dummyContainerTaskContext(resourceRequirements)
	status, _, err := handler.GetTaskStatus(context.TODO(), taskCtx, resource)
	assert.Nil(t, err)
	assert.Equal(t, types.TaskPhasePermanentFailure, status.Phase)
}
