package hive

import (
	idlCore "github.com/flyteorg/flyteidl/gen/pb-go/flyteidl/core"
	"github.com/flyteorg/flyteidl/gen/pb-go/flyteidl/plugins"
	"github.com/flyteorg/flyteplugins/go/tasks/pluginmachinery/core"
	coreMock "github.com/flyteorg/flyteplugins/go/tasks/pluginmachinery/core/mocks"
	ioMock "github.com/flyteorg/flyteplugins/go/tasks/pluginmachinery/io/mocks"
	"github.com/flyteorg/flyteplugins/go/tasks/pluginmachinery/utils"
	"github.com/flyteorg/flytestdlib/storage"
	structpb "github.com/golang/protobuf/ptypes/struct"
	"github.com/stretchr/testify/mock"
	v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
)

func GetSingleHiveQueryTaskTemplate() idlCore.TaskTemplate {
	hiveJob := plugins.QuboleHiveJob{
		ClusterLabel: "default",
		Tags:         []string{"flyte_plugin_test"},
		Query: &plugins.HiveQuery{
			TimeoutSec: 500,
			Query:      "select 'one'",
			RetryCount: 0,
		},
		// Even though it's deprecated, we might have one element in the query collection for backwards compatibility
		QueryCollection: &plugins.HiveQueryCollection{
			Queries: []*plugins.HiveQuery{
				{
					TimeoutSec: 500,
					Query:      "select 'one'",
					RetryCount: 0,
				},
			},
		},
	}
	stObj := &structpb.Struct{}
	_ = utils.MarshalStruct(&hiveJob, stObj)
	tt := idlCore.TaskTemplate{
		Type:   "hive",
		Custom: stObj,
		Id: &idlCore.Identifier{
			Name:         "sample_hive_task_test_name",
			Project:      "flyteplugins",
			Version:      "1",
			ResourceType: idlCore.ResourceType_TASK,
		},
		Interface: &idlCore.TypedInterface{
			Outputs: &idlCore.VariableMap{
				Variables: map[string]*idlCore.Variable{
					"results": &idlCore.Variable{
						Type: &idlCore.LiteralType{
							Type: &idlCore.LiteralType_Schema{Schema: &idlCore.SchemaType{}},
						},
					},
				},
			},
		},
	}

	return tt
}

var resourceRequirements = &v1.ResourceRequirements{
	Limits: v1.ResourceList{
		v1.ResourceCPU:     resource.MustParse("1024m"),
		v1.ResourceStorage: resource.MustParse("100M"),
	},
}

func GetMockTaskExecutionMetadata() core.TaskExecutionMetadata {
	taskMetadata := &coreMock.TaskExecutionMetadata{}
	taskMetadata.On("GetNamespace").Return("test-namespace")
	taskMetadata.On("GetAnnotations").Return(map[string]string{"annotation-1": "val1"})
	taskMetadata.On("GetLabels").Return(map[string]string{"label-1": "val1"})
	taskMetadata.On("GetMaxAttempts").Return(uint32(1))
	taskMetadata.On("GetOwnerReference").Return(metav1.OwnerReference{
		Kind: "node",
		Name: "blah",
	})
	taskMetadata.On("IsInterruptible").Return(true)
	taskMetadata.On("GetK8sServiceAccount").Return("service-account")
	taskMetadata.On("GetOwnerID").Return(types.NamespacedName{
		Namespace: "test-namespace",
		Name:      "test-owner-name",
	})

	tID := &coreMock.TaskExecutionID{}
	tID.On("GetID").Return(idlCore.TaskExecutionIdentifier{
		TaskId: &idlCore.Identifier{
			Domain: "production",
		},
		NodeExecutionId: &idlCore.NodeExecutionIdentifier{
			ExecutionId: &idlCore.WorkflowExecutionIdentifier{
				Name:    "my_wf_exec_name",
				Project: "my_wf_exec_project",
				Domain:  "my_wf_exec_domain",
			},
		},
		RetryAttempt: 1,
	})
	tID.On("GetGeneratedName").Return("my_wf_exec_project:my_wf_exec_domain:my_wf_exec_name")
	taskMetadata.On("GetTaskExecutionID").Return(tID)

	to := &coreMock.TaskOverrides{}
	to.On("GetResources").Return(resourceRequirements)
	taskMetadata.On("GetOverrides").Return(to)

	return taskMetadata
}

func GetMockTaskExecutionContext() core.TaskExecutionContext {
	tt := GetSingleHiveQueryTaskTemplate()

	dummyTaskMetadata := GetMockTaskExecutionMetadata()
	taskCtx := &coreMock.TaskExecutionContext{}
	inputReader := &ioMock.InputReader{}
	inputReader.On("GetInputPrefixPath").Return(storage.DataReference("s3://test-input-prefix"))
	inputReader.On("GetInputPath").Return(storage.DataReference("test-data-reference"))
	inputReader.On("Get", mock.Anything).Return(&idlCore.LiteralMap{}, nil)
	taskCtx.On("InputReader").Return(inputReader)

	outputReader := &ioMock.OutputWriter{}
	outputReader.On("GetOutputPath").Return(storage.DataReference("/data/outputs.pb"))
	outputReader.On("GetOutputPrefixPath").Return(storage.DataReference("/data/"))
	outputReader.On("GetRawOutputPrefix").Return(storage.DataReference("gs://custom-output-bucket/b"))
	outputReader.OnGetCheckpointPrefix().Return("/checkpoint")
	outputReader.OnGetPreviousCheckpointsPrefix().Return("/prev")
	taskCtx.On("OutputWriter").Return(outputReader)

	taskReader := &coreMock.TaskReader{}
	taskReader.On("Read", mock.Anything).Return(&tt, nil)
	taskCtx.On("TaskReader").Return(taskReader)

	resourceManager := &coreMock.ResourceManager{}
	taskCtx.On("ResourceManager").Return(resourceManager)

	taskCtx.On("TaskExecutionMetadata").Return(dummyTaskMetadata)
	mockSecretManager := &coreMock.SecretManager{}
	mockSecretManager.On("Get", mock.Anything, mock.Anything).Return("fake key", nil)
	taskCtx.On("SecretManager").Return(mockSecretManager)

	return taskCtx
}
