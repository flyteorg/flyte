package flytek8s

import (
	"context"
	"encoding/json"
	"io/ioutil"
	"path/filepath"
	"testing"
	"time"

	config1 "github.com/flyteorg/flytestdlib/config"
	"github.com/flyteorg/flytestdlib/config/viper"

	"github.com/flyteorg/flytestdlib/storage"
	"github.com/stretchr/testify/mock"

	"github.com/flyteorg/flyteplugins/go/tasks/pluginmachinery/flytek8s/config"
	"github.com/flyteorg/flyteplugins/go/tasks/pluginmachinery/io"

	"github.com/flyteorg/flyteidl/gen/pb-go/flyteidl/core"
	"github.com/stretchr/testify/assert"
	v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	metaV1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	pluginsCore "github.com/flyteorg/flyteplugins/go/tasks/pluginmachinery/core"
	pluginsCoreMock "github.com/flyteorg/flyteplugins/go/tasks/pluginmachinery/core/mocks"
	pluginsIOMock "github.com/flyteorg/flyteplugins/go/tasks/pluginmachinery/io/mocks"
)

func dummyTaskExecutionMetadata(resources *v1.ResourceRequirements) pluginsCore.TaskExecutionMetadata {
	taskExecutionMetadata := &pluginsCoreMock.TaskExecutionMetadata{}
	taskExecutionMetadata.On("GetNamespace").Return("test-namespace")
	taskExecutionMetadata.On("GetAnnotations").Return(map[string]string{"annotation-1": "val1"})
	taskExecutionMetadata.On("GetLabels").Return(map[string]string{"label-1": "val1"})
	taskExecutionMetadata.On("GetOwnerReference").Return(metaV1.OwnerReference{
		Kind: "node",
		Name: "blah",
	})
	taskExecutionMetadata.On("GetK8sServiceAccount").Return("service-account")
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
	tID.On("GetGeneratedName").Return("some-acceptable-name")
	taskExecutionMetadata.On("GetTaskExecutionID").Return(tID)

	to := &pluginsCoreMock.TaskOverrides{}
	to.On("GetResources").Return(resources)
	taskExecutionMetadata.On("GetOverrides").Return(to)
	taskExecutionMetadata.On("IsInterruptible").Return(true)
	return taskExecutionMetadata
}

func dummyTaskReader() pluginsCore.TaskReader {
	taskReader := &pluginsCoreMock.TaskReader{}
	task := &core.TaskTemplate{
		Type: "test",
		Target: &core.TaskTemplate_Container{
			Container: &core.Container{
				Command: []string{"command"},
				Args:    []string{"{{.Input}}"},
			},
		},
	}
	taskReader.On("Read", mock.Anything).Return(task, nil)
	return taskReader
}

func dummyInputReader() io.InputReader {
	inputReader := &pluginsIOMock.InputReader{}
	inputReader.OnGetInputPath().Return(storage.DataReference("test-data-reference"))
	inputReader.OnGetInputPrefixPath().Return(storage.DataReference("test-data-reference-prefix"))
	inputReader.OnGetMatch(mock.Anything).Return(&core.LiteralMap{}, nil)
	return inputReader
}

func dummyExecContext(r *v1.ResourceRequirements) pluginsCore.TaskExecutionContext {
	ow := &pluginsIOMock.OutputWriter{}
	ow.OnGetOutputPrefixPath().Return("")
	ow.OnGetRawOutputPrefix().Return("")

	tCtx := &pluginsCoreMock.TaskExecutionContext{}
	tCtx.OnTaskExecutionMetadata().Return(dummyTaskExecutionMetadata(r))
	tCtx.OnInputReader().Return(dummyInputReader())
	tCtx.OnTaskReader().Return(dummyTaskReader())
	tCtx.OnOutputWriter().Return(ow)
	return tCtx
}

func TestPodSetup(t *testing.T) {
	configAccessor := viper.NewAccessor(config1.Options{
		StrictMode:  true,
		SearchPaths: []string{"testdata/config.yaml"},
	})
	err := configAccessor.UpdateConfig(context.TODO())
	assert.NoError(t, err)

	t.Run("ApplyInterruptibleNodeAffinity", TestApplyInterruptibleNodeAffinity)
	t.Run("UpdatePod", updatePod)
	t.Run("ToK8sPodInterruptible", toK8sPodInterruptible)
}

func TestApplyInterruptibleNodeAffinity(t *testing.T) {
	t.Run("WithInterruptibleNodeSelectorRequirement", func(t *testing.T) {
		podSpec := v1.PodSpec{}
		ApplyInterruptibleNodeAffinity(true, &podSpec)
		assert.EqualValues(
			t,
			[]v1.NodeSelectorTerm{
				v1.NodeSelectorTerm{
					MatchExpressions: []v1.NodeSelectorRequirement{
						v1.NodeSelectorRequirement{
							Key:      "x/interruptible",
							Operator: v1.NodeSelectorOpIn,
							Values:   []string{"true"},
						},
					},
				},
			},
			podSpec.Affinity.NodeAffinity.RequiredDuringSchedulingIgnoredDuringExecution.NodeSelectorTerms,
		)
	})

	t.Run("WithNonInterruptibleNodeSelectorRequirement", func(t *testing.T) {
		podSpec := v1.PodSpec{}
		ApplyInterruptibleNodeAffinity(false, &podSpec)
		assert.EqualValues(
			t,
			[]v1.NodeSelectorTerm{
				v1.NodeSelectorTerm{
					MatchExpressions: []v1.NodeSelectorRequirement{
						v1.NodeSelectorRequirement{
							Key:      "x/interruptible",
							Operator: v1.NodeSelectorOpDoesNotExist,
						},
					},
				},
			},
			podSpec.Affinity.NodeAffinity.RequiredDuringSchedulingIgnoredDuringExecution.NodeSelectorTerms,
		)
	})

	t.Run("WithExistingAffinityWithInterruptibleNodeSelectorRequirement", func(t *testing.T) {
		podSpec := v1.PodSpec{
			Affinity: &v1.Affinity{
				NodeAffinity: &v1.NodeAffinity{
					RequiredDuringSchedulingIgnoredDuringExecution: &v1.NodeSelector{
						NodeSelectorTerms: []v1.NodeSelectorTerm{
							v1.NodeSelectorTerm{
								MatchExpressions: []v1.NodeSelectorRequirement{
									v1.NodeSelectorRequirement{
										Key:      "node selector requirement",
										Operator: v1.NodeSelectorOpIn,
										Values:   []string{"exists"},
									},
								},
							},
						},
					},
				},
			},
		}
		ApplyInterruptibleNodeAffinity(true, &podSpec)
		assert.EqualValues(
			t,
			[]v1.NodeSelectorTerm{
				v1.NodeSelectorTerm{
					MatchExpressions: []v1.NodeSelectorRequirement{
						v1.NodeSelectorRequirement{
							Key:      "node selector requirement",
							Operator: v1.NodeSelectorOpIn,
							Values:   []string{"exists"},
						},
						v1.NodeSelectorRequirement{
							Key:      "x/interruptible",
							Operator: v1.NodeSelectorOpIn,
							Values:   []string{"true"},
						},
					},
				},
			},
			podSpec.Affinity.NodeAffinity.RequiredDuringSchedulingIgnoredDuringExecution.NodeSelectorTerms,
		)
	})
}

func updatePod(t *testing.T) {
	taskExecutionMetadata := dummyTaskExecutionMetadata(&v1.ResourceRequirements{
		Limits: v1.ResourceList{
			v1.ResourceCPU:     resource.MustParse("1024m"),
			v1.ResourceStorage: resource.MustParse("100M"),
		},
		Requests: v1.ResourceList{
			v1.ResourceCPU:     resource.MustParse("1024m"),
			v1.ResourceStorage: resource.MustParse("100M"),
		},
	})

	pod := &v1.Pod{
		Spec: v1.PodSpec{
			Tolerations: []v1.Toleration{
				{
					Key:   "my toleration key",
					Value: "my toleration value",
				},
			},
			NodeSelector: map[string]string{
				"user": "also configured",
			},
		},
	}
	UpdatePod(taskExecutionMetadata, []v1.ResourceRequirements{}, &pod.Spec)
	assert.Equal(t, v1.RestartPolicyNever, pod.Spec.RestartPolicy)
	for _, tol := range pod.Spec.Tolerations {
		if tol.Key == "x/flyte" {
			assert.Equal(t, tol.Value, "interruptible")
			assert.Equal(t, tol.Operator, v1.TolerationOperator("Equal"))
			assert.Equal(t, tol.Effect, v1.TaintEffect("NoSchedule"))
		} else if tol.Key == "my toleration key" {
			assert.Equal(t, tol.Value, "my toleration value")
		} else {
			t.Fatalf("unexpected toleration [%+v]", tol)
		}
	}
	assert.Equal(t, "service-account", pod.Spec.ServiceAccountName)
	assert.Equal(t, "flyte-scheduler", pod.Spec.SchedulerName)
	assert.Len(t, pod.Spec.Tolerations, 2)
	assert.EqualValues(t, map[string]string{
		"x/interruptible": "true",
		"user":            "also configured",
	}, pod.Spec.NodeSelector)
	assert.EqualValues(
		t,
		[]v1.NodeSelectorTerm{
			v1.NodeSelectorTerm{
				MatchExpressions: []v1.NodeSelectorRequirement{
					v1.NodeSelectorRequirement{
						Key:      "x/interruptible",
						Operator: v1.NodeSelectorOpIn,
						Values:   []string{"true"},
					},
				},
			},
		},
		pod.Spec.Affinity.NodeAffinity.RequiredDuringSchedulingIgnoredDuringExecution.NodeSelectorTerms,
	)
}

func TestUpdatePodWithDefaultAffinityAndInterruptibleNodeSelectorRequirement(t *testing.T) {
	taskExecutionMetadata := dummyTaskExecutionMetadata(&v1.ResourceRequirements{})
	assert.NoError(t, config.SetK8sPluginConfig(&config.K8sPluginConfig{
		DefaultAffinity: &v1.Affinity{
			NodeAffinity: &v1.NodeAffinity{
				RequiredDuringSchedulingIgnoredDuringExecution: &v1.NodeSelector{
					NodeSelectorTerms: []v1.NodeSelectorTerm{
						v1.NodeSelectorTerm{
							MatchExpressions: []v1.NodeSelectorRequirement{
								v1.NodeSelectorRequirement{
									Key:      "default node affinity",
									Operator: v1.NodeSelectorOpIn,
									Values:   []string{"exists"},
								},
							},
						},
					},
				},
			},
		},
		InterruptibleNodeSelectorRequirement: &v1.NodeSelectorRequirement{
			Key:      "x/interruptible",
			Operator: v1.NodeSelectorOpIn,
			Values:   []string{"true"},
		},
	}))
	for i := 0; i < 3; i++ {
		podSpec := v1.PodSpec{}
		UpdatePod(taskExecutionMetadata, []v1.ResourceRequirements{}, &podSpec)
		assert.EqualValues(
			t,
			[]v1.NodeSelectorTerm{
				v1.NodeSelectorTerm{
					MatchExpressions: []v1.NodeSelectorRequirement{
						v1.NodeSelectorRequirement{
							Key:      "default node affinity",
							Operator: v1.NodeSelectorOpIn,
							Values:   []string{"exists"},
						},
						v1.NodeSelectorRequirement{
							Key:      "x/interruptible",
							Operator: v1.NodeSelectorOpIn,
							Values:   []string{"true"},
						},
					},
				},
			},
			podSpec.Affinity.NodeAffinity.RequiredDuringSchedulingIgnoredDuringExecution.NodeSelectorTerms,
		)
	}
}

func toK8sPodInterruptible(t *testing.T) {
	ctx := context.TODO()

	x := dummyExecContext(&v1.ResourceRequirements{
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

	p, err := ToK8sPodSpec(ctx, x)
	assert.NoError(t, err)
	assert.Len(t, p.Tolerations, 2)
	assert.Equal(t, "x/flyte", p.Tolerations[1].Key)
	assert.Equal(t, "interruptible", p.Tolerations[1].Value)
	assert.Equal(t, 1, len(p.NodeSelector))
	assert.Equal(t, "true", p.NodeSelector["x/interruptible"])
	assert.EqualValues(
		t,
		[]v1.NodeSelectorTerm{
			v1.NodeSelectorTerm{
				MatchExpressions: []v1.NodeSelectorRequirement{
					v1.NodeSelectorRequirement{
						Key:      "x/interruptible",
						Operator: v1.NodeSelectorOpIn,
						Values:   []string{"true"},
					},
				},
			},
		},
		p.Affinity.NodeAffinity.RequiredDuringSchedulingIgnoredDuringExecution.NodeSelectorTerms,
	)
}

func TestToK8sPod(t *testing.T) {
	ctx := context.TODO()

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
		DefaultCPURequest:    "1024m",
		DefaultMemoryRequest: "1024Mi",
	}))

	op := &pluginsIOMock.OutputFilePaths{}
	op.On("GetOutputPrefixPath").Return(storage.DataReference(""))
	op.On("GetRawOutputPrefix").Return(storage.DataReference(""))

	t.Run("WithGPU", func(t *testing.T) {
		x := dummyExecContext(&v1.ResourceRequirements{
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

		p, err := ToK8sPodSpec(ctx, x)
		assert.NoError(t, err)
		assert.Equal(t, len(p.Tolerations), 1)
	})

	t.Run("NoGPU", func(t *testing.T) {
		x := dummyExecContext(&v1.ResourceRequirements{
			Limits: v1.ResourceList{
				v1.ResourceCPU:     resource.MustParse("1024m"),
				v1.ResourceStorage: resource.MustParse("100M"),
			},
			Requests: v1.ResourceList{
				v1.ResourceCPU:     resource.MustParse("1024m"),
				v1.ResourceStorage: resource.MustParse("100M"),
			},
		})

		p, err := ToK8sPodSpec(ctx, x)
		assert.NoError(t, err)
		assert.Equal(t, len(p.Tolerations), 0)
		assert.Equal(t, "some-acceptable-name", p.Containers[0].Name)
	})

	t.Run("Default toleration, selector, scheduler", func(t *testing.T) {
		x := dummyExecContext(&v1.ResourceRequirements{
			Limits: v1.ResourceList{
				v1.ResourceCPU:     resource.MustParse("1024m"),
				v1.ResourceStorage: resource.MustParse("100M"),
			},
			Requests: v1.ResourceList{
				v1.ResourceCPU:     resource.MustParse("1024m"),
				v1.ResourceStorage: resource.MustParse("100M"),
			},
		})

		assert.NoError(t, config.SetK8sPluginConfig(&config.K8sPluginConfig{
			DefaultTolerations: []v1.Toleration{
				{
					Key:   "tolerationKey",
					Value: flyteDataConfigVolume,
				},
			},
			DefaultNodeSelector: map[string]string{
				"nodeId": "123",
			},
			SchedulerName:        "myScheduler",
			DefaultCPURequest:    "1024m",
			DefaultMemoryRequest: "1024Mi",
		}))

		p, err := ToK8sPodSpec(ctx, x)
		assert.NoError(t, err)
		assert.Equal(t, 1, len(p.Tolerations))
		assert.Equal(t, 1, len(p.NodeSelector))
		assert.Equal(t, "myScheduler", p.SchedulerName)
		assert.Equal(t, "some-acceptable-name", p.Containers[0].Name)
	})
}

func TestDemystifyPending(t *testing.T) {
	assert.NoError(t, config.SetK8sPluginConfig(&config.K8sPluginConfig{
		CreateContainerErrorGracePeriod: config1.Duration{
			Duration: time.Minute * 3,
		},
	}))

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
		assert.Equal(t, pluginsCore.PhaseQueued, taskStatus.Phase())
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
		assert.Equal(t, pluginsCore.PhaseQueued, taskStatus.Phase())
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
		assert.Equal(t, pluginsCore.PhaseQueued, taskStatus.Phase())
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
		assert.Equal(t, pluginsCore.PhaseQueued, taskStatus.Phase())
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
		assert.Equal(t, pluginsCore.PhaseInitializing, taskStatus.Phase())
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
		assert.Equal(t, pluginsCore.PhaseInitializing, taskStatus.Phase())
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
		assert.Equal(t, pluginsCore.PhaseInitializing, taskStatus.Phase())
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
		assert.Equal(t, pluginsCore.PhaseRetryableFailure, taskStatus.Phase())
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
		assert.Equal(t, pluginsCore.PhasePermanentFailure, taskStatus.Phase())
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
		assert.Equal(t, pluginsCore.PhaseRetryableFailure, taskStatus.Phase())
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
		assert.Equal(t, pluginsCore.PhaseRetryableFailure, taskStatus.Phase())
	})

	t.Run("CreateContainerConfigError", func(t *testing.T) {
		s.ContainerStatuses = []v1.ContainerStatus{
			{
				Ready: false,
				State: v1.ContainerState{
					Waiting: &v1.ContainerStateWaiting{
						Reason:  "CreateContainerConfigError",
						Message: "this an error",
					},
				},
			},
		}
		taskStatus, err := DemystifyPending(s)
		assert.NoError(t, err)
		assert.Equal(t, pluginsCore.PhasePermanentFailure, taskStatus.Phase())
	})

	t.Run("CreateContainerErrorWithinGracePeriod", func(t *testing.T) {
		s2 := *s.DeepCopy()
		s2.Conditions[0].LastTransitionTime = metaV1.Now()
		s2.ContainerStatuses = []v1.ContainerStatus{
			{
				Ready: false,
				State: v1.ContainerState{
					Waiting: &v1.ContainerStateWaiting{
						Reason:  "CreateContainerError",
						Message: "this is a transient error",
					},
				},
			},
		}
		taskStatus, err := DemystifyPending(s2)
		assert.NoError(t, err)
		assert.Equal(t, pluginsCore.PhaseInitializing, taskStatus.Phase())
	})

	t.Run("CreateContainerErrorOutsideGracePeriod", func(t *testing.T) {
		s2 := *s.DeepCopy()
		s2.Conditions[0].LastTransitionTime.Time = metaV1.Now().Add(-config.GetK8sPluginConfig().CreateContainerErrorGracePeriod.Duration)
		s2.ContainerStatuses = []v1.ContainerStatus{
			{
				Ready: false,
				State: v1.ContainerState{
					Waiting: &v1.ContainerStateWaiting{
						Reason:  "CreateContainerError",
						Message: "this a permanent error",
					},
				},
			},
		}
		taskStatus, err := DemystifyPending(s2)
		assert.NoError(t, err)
		assert.Equal(t, pluginsCore.PhasePermanentFailure, taskStatus.Phase())
	})
}

func TestDemystifySuccess(t *testing.T) {
	t.Run("OOMKilled", func(t *testing.T) {
		phaseInfo, err := DemystifySuccess(v1.PodStatus{
			ContainerStatuses: []v1.ContainerStatus{
				{
					State: v1.ContainerState{
						Terminated: &v1.ContainerStateTerminated{
							Reason: OOMKilled,
						},
					},
				},
			},
		}, pluginsCore.TaskInfo{})
		assert.Nil(t, err)
		assert.Equal(t, pluginsCore.PhaseRetryableFailure, phaseInfo.Phase())
		assert.Equal(t, "OOMKilled", phaseInfo.Err().Code)
	})

	t.Run("InitContainer OOMKilled", func(t *testing.T) {
		phaseInfo, err := DemystifySuccess(v1.PodStatus{
			InitContainerStatuses: []v1.ContainerStatus{
				{
					State: v1.ContainerState{
						Terminated: &v1.ContainerStateTerminated{
							Reason: OOMKilled,
						},
					},
				},
			},
		}, pluginsCore.TaskInfo{})
		assert.Nil(t, err)
		assert.Equal(t, pluginsCore.PhaseRetryableFailure, phaseInfo.Phase())
		assert.Equal(t, "OOMKilled", phaseInfo.Err().Code)
	})

	t.Run("success", func(t *testing.T) {
		phaseInfo, err := DemystifySuccess(v1.PodStatus{}, pluginsCore.TaskInfo{})
		assert.Nil(t, err)
		assert.Equal(t, pluginsCore.PhaseSuccess, phaseInfo.Phase())
	})
}

func TestConvertPodFailureToError(t *testing.T) {
	t.Run("unknown-error", func(t *testing.T) {
		code, _ := ConvertPodFailureToError(v1.PodStatus{})
		assert.Equal(t, code, "UnknownError")
	})

	t.Run("known-error", func(t *testing.T) {
		code, _ := ConvertPodFailureToError(v1.PodStatus{Reason: "hello"})
		assert.Equal(t, code, "hello")
	})

	t.Run("OOMKilled", func(t *testing.T) {
		code, _ := ConvertPodFailureToError(v1.PodStatus{
			ContainerStatuses: []v1.ContainerStatus{
				{
					State: v1.ContainerState{
						Terminated: &v1.ContainerStateTerminated{
							Reason:   OOMKilled,
							ExitCode: 137,
						},
					},
				},
			},
		})
		assert.Equal(t, code, "OOMKilled")
	})

	t.Run("OOMKilled", func(t *testing.T) {
		code, _ := ConvertPodFailureToError(v1.PodStatus{
			ContainerStatuses: []v1.ContainerStatus{
				{
					LastTerminationState: v1.ContainerState{
						Terminated: &v1.ContainerStateTerminated{
							Reason:   OOMKilled,
							ExitCode: 137,
						},
					},
				},
			},
		})
		assert.Equal(t, code, "OOMKilled")
	})

	t.Run("SIGKILL", func(t *testing.T) {
		code, _ := ConvertPodFailureToError(v1.PodStatus{
			ContainerStatuses: []v1.ContainerStatus{
				{
					LastTerminationState: v1.ContainerState{
						Terminated: &v1.ContainerStateTerminated{
							Reason:   "some reason",
							ExitCode: SIGKILL,
						},
					},
				},
			},
		})
		assert.Equal(t, code, "Interrupted")
	})
}

func TestDemystifyPending_testcases(t *testing.T) {

	tests := []struct {
		name     string
		filename string
		isErr    bool
		errCode  string
		message  string
	}{
		{"ImagePullBackOff", "imagepull-failurepod.json", false, "ContainersNotReady|ImagePullBackOff", "containers with unready status: [fdf98e4ed2b524dc3bf7-get-flyte-id-task-0]|Back-off pulling image \"image\""},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			testFile := filepath.Join("testdata", tt.filename)
			data, err := ioutil.ReadFile(testFile)
			assert.NoError(t, err, "failed to read file %s", testFile)
			pod := &v1.Pod{}
			if assert.NoError(t, json.Unmarshal(data, pod), "failed to unmarshal json in %s. Expected of type v1.Pod", testFile) {
				p, err := DemystifyPending(pod.Status)
				if tt.isErr {
					assert.Error(t, err, "Error expected from method")
				} else {
					assert.NoError(t, err, "Error not expected")
					assert.NotNil(t, p)
					assert.Equal(t, p.Phase(), pluginsCore.PhaseRetryableFailure)
					if assert.NotNil(t, p.Err()) {
						assert.Equal(t, p.Err().Code, tt.errCode)
						assert.Equal(t, p.Err().Message, tt.message)
					}
				}
			}
		})
	}
}

func TestDeterminePrimaryContainerPhase(t *testing.T) {
	primaryContainerName := "primary"
	secondaryContainer := v1.ContainerStatus{
		Name: "secondary",
		State: v1.ContainerState{
			Terminated: &v1.ContainerStateTerminated{
				ExitCode: 0,
			},
		},
	}
	var info = &pluginsCore.TaskInfo{}
	t.Run("primary container waiting", func(t *testing.T) {
		phaseInfo := DeterminePrimaryContainerPhase(primaryContainerName, []v1.ContainerStatus{
			secondaryContainer, {
				Name: primaryContainerName,
				State: v1.ContainerState{
					Waiting: &v1.ContainerStateWaiting{
						Reason: "just dawdling",
					},
				},
			},
		}, info)
		assert.Equal(t, pluginsCore.PhaseRunning, phaseInfo.Phase())
	})
	t.Run("primary container running", func(t *testing.T) {
		phaseInfo := DeterminePrimaryContainerPhase(primaryContainerName, []v1.ContainerStatus{
			secondaryContainer, {
				Name: primaryContainerName,
				State: v1.ContainerState{
					Running: &v1.ContainerStateRunning{
						StartedAt: metaV1.Now(),
					},
				},
			},
		}, info)
		assert.Equal(t, pluginsCore.PhaseRunning, phaseInfo.Phase())
	})
	t.Run("primary container failed", func(t *testing.T) {
		phaseInfo := DeterminePrimaryContainerPhase(primaryContainerName, []v1.ContainerStatus{
			secondaryContainer, {
				Name: primaryContainerName,
				State: v1.ContainerState{
					Terminated: &v1.ContainerStateTerminated{
						ExitCode: 1,
						Reason:   "foo",
						Message:  "foo failed",
					},
				},
			},
		}, info)
		assert.Equal(t, pluginsCore.PhaseRetryableFailure, phaseInfo.Phase())
		assert.Equal(t, "foo", phaseInfo.Err().Code)
		assert.Equal(t, "foo failed", phaseInfo.Err().Message)
	})
	t.Run("primary container succeeded", func(t *testing.T) {
		phaseInfo := DeterminePrimaryContainerPhase(primaryContainerName, []v1.ContainerStatus{
			secondaryContainer, {
				Name: primaryContainerName,
				State: v1.ContainerState{
					Terminated: &v1.ContainerStateTerminated{
						ExitCode: 0,
					},
				},
			},
		}, info)
		assert.Equal(t, pluginsCore.PhaseSuccess, phaseInfo.Phase())
	})
	t.Run("missing primary container", func(t *testing.T) {
		phaseInfo := DeterminePrimaryContainerPhase(primaryContainerName, []v1.ContainerStatus{
			secondaryContainer,
		}, info)
		assert.Equal(t, pluginsCore.PhasePermanentFailure, phaseInfo.Phase())
		assert.Equal(t, "Primary container [primary] not found in pod's container statuses", phaseInfo.Err().Message)
	})
}
