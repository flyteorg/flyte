package k8s

import (
	"context"
	"encoding/json"
	"testing"

	"k8s.io/apimachinery/pkg/api/resource"

	"github.com/flyteorg/flytestdlib/storage"

	"github.com/flyteorg/flyteidl/gen/pb-go/flyteidl/core"
	idlPlugins "github.com/flyteorg/flyteidl/gen/pb-go/flyteidl/plugins"
	"github.com/flyteorg/flyteplugins/go/tasks/pluginmachinery/core/mocks"
	mocks2 "github.com/flyteorg/flyteplugins/go/tasks/pluginmachinery/io/mocks"
	"github.com/flyteorg/flyteplugins/go/tasks/pluginmachinery/utils"
	structpb "github.com/golang/protobuf/ptypes/struct"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	v12 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

const testPrimaryContainerName = "primary container"

var podSpec = v1.PodSpec{
	Containers: []v1.Container{
		{
			Name: testPrimaryContainerName,
			Resources: v1.ResourceRequirements{
				Requests: v1.ResourceList{
					v1.ResourceCPU:     resource.MustParse("1"),
					v1.ResourceStorage: resource.MustParse("2"),
				},
			},
		},
		{
			Name: "secondary container",
		},
	},
}

var arrayJob = idlPlugins.ArrayJob{
	Size: 100,
}

func getK8sPodTask(t *testing.T, annotations map[string]string) *core.TaskTemplate {
	marshalledPodspec, err := json.Marshal(podSpec)
	if err != nil {
		t.Fatal(err)
	}

	structObj := &structpb.Struct{}
	if err := json.Unmarshal(marshalledPodspec, structObj); err != nil {
		t.Fatal(err)
	}

	custom := &structpb.Struct{}
	if err := utils.MarshalStruct(&arrayJob, custom); err != nil {
		t.Fatal(err)
	}

	return &core.TaskTemplate{
		TaskTypeVersion: 2,
		Config: map[string]string{
			primaryContainerKey: testPrimaryContainerName,
		},
		Target: &core.TaskTemplate_K8SPod{
			K8SPod: &core.K8SPod{
				PodSpec: structObj,
				Metadata: &core.K8SObjectMetadata{
					Labels: map[string]string{
						"label": "foo",
					},
					Annotations: annotations,
				},
			},
		},
		Custom: custom,
	}
}

func TestBuildPodMapTask(t *testing.T) {
	tMeta := &mocks.TaskExecutionMetadata{}
	tMeta.OnGetSecurityContext().Return(core.SecurityContext{})
	tMeta.OnGetK8sServiceAccount().Return("sa")
	pod, err := buildPodMapTask(getK8sPodTask(t, map[string]string{
		"anno": "bar",
	}), tMeta)
	assert.NoError(t, err)
	var expected = podSpec.DeepCopy()
	expected.RestartPolicy = v1.RestartPolicyNever
	assert.EqualValues(t, *expected, pod.Spec)
	assert.EqualValues(t, map[string]string{
		"label": "foo",
	}, pod.Labels)
	assert.EqualValues(t, map[string]string{
		"anno":                   "bar",
		"primary_container_name": "primary container",
	}, pod.Annotations)
}

func TestBuildPodMapTask_Errors(t *testing.T) {
	t.Run("invalid task template", func(t *testing.T) {
		_, err := buildPodMapTask(&core.TaskTemplate{}, nil)
		assert.EqualError(t, err, "[BadTaskSpecification] Missing pod spec for task")
	})
	b, err := json.Marshal(podSpec)
	if err != nil {
		t.Fatal(err)
	}

	structObj := &structpb.Struct{}
	if err := json.Unmarshal(b, structObj); err != nil {
		t.Fatal(err)
	}
	t.Run("missing primary container annotation", func(t *testing.T) {
		_, err = buildPodMapTask(&core.TaskTemplate{
			Target: &core.TaskTemplate_K8SPod{
				K8SPod: &core.K8SPod{
					PodSpec: structObj,
				},
			},
		}, nil)
		assert.EqualError(t, err, "[BadTaskSpecification] invalid TaskSpecification, config missing [primary_container_name] key in [map[]]")
	})
}

func TestBuildPodMapTask_AddAnnotations(t *testing.T) {
	tMeta := &mocks.TaskExecutionMetadata{}
	tMeta.OnGetSecurityContext().Return(core.SecurityContext{})
	tMeta.OnGetK8sServiceAccount().Return("sa")
	podTask := getK8sPodTask(t, nil)
	pod, err := buildPodMapTask(podTask, tMeta)
	assert.NoError(t, err)
	var expected = podSpec.DeepCopy()
	expected.RestartPolicy = v1.RestartPolicyNever
	assert.EqualValues(t, *expected, pod.Spec)
	assert.EqualValues(t, map[string]string{
		"label": "foo",
	}, pod.Labels)
	assert.EqualValues(t, map[string]string{
		"primary_container_name": "primary container",
	}, pod.Annotations)
}

func TestFlyteArrayJobToK8sPodTemplate(t *testing.T) {
	ctx := context.TODO()
	tr := &mocks.TaskReader{}
	tr.OnRead(ctx).Return(getK8sPodTask(t, map[string]string{
		"anno": "bar",
	}), nil)

	ir := &mocks2.InputReader{}
	ir.OnGetInputPrefixPath().Return("/prefix/")
	ir.OnGetInputPath().Return("/prefix/inputs.pb")
	ir.OnGetMatch(mock.Anything).Return(&core.LiteralMap{}, nil)

	tMeta := &mocks.TaskExecutionMetadata{}
	tMeta.OnGetNamespace().Return("n")
	tMeta.OnGetLabels().Return(map[string]string{
		"tCtx": "label",
	})
	tMeta.OnGetAnnotations().Return(map[string]string{
		"tCtx": "anno",
	})
	tMeta.OnGetOwnerReference().Return(v12.OwnerReference{})
	tMeta.OnGetSecurityContext().Return(core.SecurityContext{})
	tMeta.OnGetK8sServiceAccount().Return("sa")
	tID := &mocks.TaskExecutionID{}
	tID.OnGetID().Return(core.TaskExecutionIdentifier{
		NodeExecutionId: &core.NodeExecutionIdentifier{
			ExecutionId: &core.WorkflowExecutionIdentifier{
				Name:    "my_name",
				Project: "my_project",
				Domain:  "my_domain",
			},
		},
		RetryAttempt: 1,
	})
	tMeta.OnGetTaskExecutionID().Return(tID)

	outputReader := &mocks2.OutputWriter{}
	outputReader.On("GetOutputPath").Return(storage.DataReference("/data/outputs.pb"))
	outputReader.On("GetOutputPrefixPath").Return(storage.DataReference("/data/"))
	outputReader.On("GetRawOutputPrefix").Return(storage.DataReference(""))

	tCtx := &mocks.TaskExecutionContext{}
	tCtx.OnTaskReader().Return(tr)
	tCtx.OnInputReader().Return(ir)
	tCtx.OnTaskExecutionMetadata().Return(tMeta)
	tCtx.OnOutputWriter().Return(outputReader)

	pod, job, err := FlyteArrayJobToK8sPodTemplate(ctx, tCtx, "")
	assert.NoError(t, err)
	assert.EqualValues(t, metav1.ObjectMeta{
		Namespace: "n",
		Labels: map[string]string{
			"tCtx":  "label",
			"label": "foo",
		},
		Annotations: map[string]string{
			"tCtx":                   "anno",
			"anno":                   "bar",
			"primary_container_name": "primary container",
			"cluster-autoscaler.kubernetes.io/safe-to-evict": "false",
		},
		OwnerReferences: []metav1.OwnerReference{
			{},
		},
	}, pod.ObjectMeta)
	assert.EqualValues(t, &arrayJob, job)
	defaultMemoryFromConfig := resource.MustParse("1024Mi")
	assert.EqualValues(t, v1.ResourceRequirements{
		Requests: v1.ResourceList{
			v1.ResourceCPU:    resource.MustParse("1"),
			v1.ResourceMemory: defaultMemoryFromConfig,
		},
		Limits: v1.ResourceList{
			v1.ResourceCPU:    resource.MustParse("1"),
			v1.ResourceMemory: defaultMemoryFromConfig,
		},
	}, pod.Spec.Containers[0].Resources)
	assert.EqualValues(t, []v1.EnvVar{
		{
			Name:  "FLYTE_INTERNAL_EXECUTION_ID",
			Value: "my_name",
		},
		{
			Name:  "FLYTE_INTERNAL_EXECUTION_PROJECT",
			Value: "my_project",
		},
		{
			Name:  "FLYTE_INTERNAL_EXECUTION_DOMAIN",
			Value: "my_domain",
		},
		{
			Name:  "FLYTE_ATTEMPT_NUMBER",
			Value: "1",
		},
	}, pod.Spec.Containers[0].Env)
}
