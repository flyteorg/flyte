package sidecar

import (
	"context"
	"errors"
	"io/ioutil"
	"os"
	"path"
	"testing"

	errors2 "github.com/lyft/flyteplugins/go/tasks/errors"

	"github.com/lyft/flytestdlib/storage"
	"github.com/stretchr/testify/mock"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"

	"github.com/golang/protobuf/jsonpb"
	structpb "github.com/golang/protobuf/ptypes/struct"
	"github.com/lyft/flyteidl/gen/pb-go/flyteidl/core"
	"github.com/lyft/flyteidl/gen/pb-go/flyteidl/plugins"
	"github.com/stretchr/testify/assert"
	v1 "k8s.io/api/core/v1"

	"github.com/lyft/flyteplugins/go/tasks/pluginmachinery/flytek8s/config"
	"github.com/lyft/flyteplugins/go/tasks/pluginmachinery/utils"

	pluginsCore "github.com/lyft/flyteplugins/go/tasks/pluginmachinery/core"
	pluginsCoreMock "github.com/lyft/flyteplugins/go/tasks/pluginmachinery/core/mocks"
	pluginsIOMock "github.com/lyft/flyteplugins/go/tasks/pluginmachinery/io/mocks"
)

const ResourceNvidiaGPU = "nvidia.com/gpu"

var resourceRequirements = &v1.ResourceRequirements{
	Limits: v1.ResourceList{
		v1.ResourceCPU:     resource.MustParse("1024m"),
		v1.ResourceStorage: resource.MustParse("100M"),
	},
}

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

func dummyContainerTaskMetadata(resources *v1.ResourceRequirements) pluginsCore.TaskExecutionMetadata {
	taskMetadata := &pluginsCoreMock.TaskExecutionMetadata{}
	taskMetadata.On("GetNamespace").Return("test-namespace")
	taskMetadata.On("GetAnnotations").Return(map[string]string{"annotation-1": "val1"})
	taskMetadata.On("GetLabels").Return(map[string]string{"label-1": "val1"})
	taskMetadata.On("GetOwnerReference").Return(metav1.OwnerReference{
		Kind: "node",
		Name: "blah",
	})
	taskMetadata.On("GetK8sServiceAccount").Return("service-account")
	taskMetadata.On("GetOwnerID").Return(types.NamespacedName{
		Namespace: "test-namespace",
		Name:      "test-owner-name",
	})

	tID := &pluginsCoreMock.TaskExecutionID{}
	tID.On("GetID").Return(core.TaskExecutionIdentifier{
		NodeExecutionId: &core.NodeExecutionIdentifier{
			ExecutionId: &core.WorkflowExecutionIdentifier{
				Name:    "my_name",
				Project: "my_project",
				Domain:  "my_domain",
			},
		},
	})
	tID.On("GetGeneratedName").Return("my_project:my_domain:my_name")
	taskMetadata.On("GetTaskExecutionID").Return(tID)

	to := &pluginsCoreMock.TaskOverrides{}
	to.On("GetResources").Return(resources)
	taskMetadata.On("GetOverrides").Return(to)

	return taskMetadata
}

func getDummySidecarTaskContext(taskTemplate *core.TaskTemplate, resources *v1.ResourceRequirements) pluginsCore.TaskExecutionContext {
	taskCtx := &pluginsCoreMock.TaskExecutionContext{}
	dummyTaskMetadata := dummyContainerTaskMetadata(resources)
	inputReader := &pluginsIOMock.InputReader{}
	inputReader.On("GetInputPrefixPath").Return(storage.DataReference("test-data-prefix"))
	inputReader.On("GetInputPath").Return(storage.DataReference("test-data-reference"))
	inputReader.On("Get", mock.Anything).Return(&core.LiteralMap{}, nil)
	taskCtx.On("InputReader").Return(inputReader)

	outputReader := &pluginsIOMock.OutputWriter{}
	outputReader.On("GetOutputPath").Return(storage.DataReference("/data/outputs.pb"))
	outputReader.On("GetOutputPrefixPath").Return(storage.DataReference("/data/"))
	taskCtx.On("OutputWriter").Return(outputReader)

	taskReader := &pluginsCoreMock.TaskReader{}
	taskReader.On("Read", mock.Anything).Return(taskTemplate, nil)
	taskCtx.On("TaskReader").Return(taskReader)

	taskCtx.On("TaskExecutionMetadata").Return(dummyTaskMetadata)
	return taskCtx
}

func TestBuildSidecarResource(t *testing.T) {
	dir, err := os.Getwd()
	if err != nil {
		t.Fatal(err)
	}
	sidecarCustomJSON, err := ioutil.ReadFile(path.Join(dir, "testdata", "sidecar_custom"))
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
	taskCtx := getDummySidecarTaskContext(&task, resourceRequirements)
	res, err := handler.BuildResource(context.TODO(), taskCtx)
	assert.Nil(t, err)
	assert.EqualValues(t, map[string]string{
		primaryContainerKey: "a container",
	}, res.GetAnnotations())
	assert.Contains(t, res.(*v1.Pod).Spec.Tolerations, tolGPU)
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
	taskCtx := getDummySidecarTaskContext(task, resourceRequirements)
	_, err := handler.BuildResource(context.TODO(), taskCtx)
	assert.True(t, errors.Is(err, errors2.Errorf("BadTaskSpecification", "")))
}

func TestGetTaskSidecarStatus(t *testing.T) {
	sideCarJob := plugins.SidecarJob{
		PrimaryContainerName: "PrimaryContainer",
		PodSpec: &v1.PodSpec{
			Containers: []v1.Container{
				{
					Name: "PrimaryContainer",
				},
			},
		},
	}

	task := getSidecarTaskTemplateForTest(sideCarJob)

	var testCases = map[v1.PodPhase]pluginsCore.Phase{
		v1.PodSucceeded:           pluginsCore.PhaseSuccess,
		v1.PodFailed:              pluginsCore.PhaseRetryableFailure,
		v1.PodReasonUnschedulable: pluginsCore.PhaseQueued,
		v1.PodUnknown:             pluginsCore.PhaseUndefined,
	}

	for podPhase, expectedTaskPhase := range testCases {
		res := &v1.Pod{
			Status: v1.PodStatus{
				Phase: podPhase,
			},
		}
		res.SetAnnotations(map[string]string{
			primaryContainerKey: "PrimaryContainer",
		})
		handler := &sidecarResourceHandler{}
		taskCtx := getDummySidecarTaskContext(task, resourceRequirements)
		phaseInfo, err := handler.GetTaskPhase(context.TODO(), taskCtx, res)
		assert.Nil(t, err)
		assert.Equal(t, expectedTaskPhase, phaseInfo.Phase(),
			"Expected [%v] got [%v] instead, for podPhase [%v]", expectedTaskPhase, phaseInfo.Phase(), podPhase)
	}
}

func TestDemystifiedSidecarStatus_PrimaryFailed(t *testing.T) {
	res := &v1.Pod{
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
	res.SetAnnotations(map[string]string{
		primaryContainerKey: "Primary",
	})
	handler := &sidecarResourceHandler{}
	taskCtx := getDummySidecarTaskContext(&core.TaskTemplate{}, resourceRequirements)
	phaseInfo, err := handler.GetTaskPhase(context.TODO(), taskCtx, res)
	assert.Nil(t, err)
	assert.Equal(t, pluginsCore.PhaseRetryableFailure, phaseInfo.Phase())
}

func TestDemystifiedSidecarStatus_PrimarySucceeded(t *testing.T) {
	res := &v1.Pod{
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
	res.SetAnnotations(map[string]string{
		primaryContainerKey: "Primary",
	})
	handler := &sidecarResourceHandler{}
	taskCtx := getDummySidecarTaskContext(&core.TaskTemplate{}, resourceRequirements)
	phaseInfo, err := handler.GetTaskPhase(context.TODO(), taskCtx, res)
	assert.Nil(t, err)
	assert.Equal(t, pluginsCore.PhaseSuccess, phaseInfo.Phase())
}

func TestDemystifiedSidecarStatus_PrimaryRunning(t *testing.T) {
	res := &v1.Pod{
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
	res.SetAnnotations(map[string]string{
		primaryContainerKey: "Primary",
	})
	handler := &sidecarResourceHandler{}
	taskCtx := getDummySidecarTaskContext(&core.TaskTemplate{}, resourceRequirements)
	phaseInfo, err := handler.GetTaskPhase(context.TODO(), taskCtx, res)
	assert.Nil(t, err)
	assert.Equal(t, pluginsCore.PhaseRunning, phaseInfo.Phase())
}

func TestDemystifiedSidecarStatus_PrimaryMissing(t *testing.T) {
	res := &v1.Pod{
		Status: v1.PodStatus{
			Phase: v1.PodRunning,
			ContainerStatuses: []v1.ContainerStatus{
				{
					Name: "Secondary",
				},
			},
		},
	}
	res.SetAnnotations(map[string]string{
		primaryContainerKey: "Primary",
	})
	handler := &sidecarResourceHandler{}
	taskCtx := getDummySidecarTaskContext(&core.TaskTemplate{}, resourceRequirements)
	phaseInfo, err := handler.GetTaskPhase(context.TODO(), taskCtx, res)
	assert.Nil(t, err)
	assert.Equal(t, pluginsCore.PhasePermanentFailure, phaseInfo.Phase())
}
