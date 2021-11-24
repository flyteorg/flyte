package sagemaker

import (
	"github.com/flyteorg/flyteidl/clients/go/coreutils"
	"github.com/flyteorg/flytestdlib/promutils"
	"github.com/pkg/errors"

	"github.com/golang/protobuf/proto"

	flyteIdlCore "github.com/flyteorg/flyteidl/gen/pb-go/flyteidl/core"
	sagemakerIdl "github.com/flyteorg/flyteidl/gen/pb-go/flyteidl/plugins/sagemaker"
	pluginsCore "github.com/flyteorg/flyteplugins/go/tasks/pluginmachinery/core"
	"github.com/flyteorg/flyteplugins/go/tasks/pluginmachinery/core/mocks"
	"github.com/flyteorg/flyteplugins/go/tasks/pluginmachinery/flytek8s"
	pluginIOMocks "github.com/flyteorg/flyteplugins/go/tasks/pluginmachinery/io/mocks"
	"github.com/flyteorg/flyteplugins/go/tasks/pluginmachinery/utils"
	"github.com/flyteorg/flytestdlib/storage"
	"github.com/golang/protobuf/jsonpb"
	structpb "github.com/golang/protobuf/ptypes/struct"
	"github.com/stretchr/testify/mock"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

const testImage = "image://"
const serviceAccount = "sagemaker_sa"

var (
	dummyEnvVars = []*flyteIdlCore.KeyValuePair{
		{Key: "Env_Var", Value: "Env_Val"},
	}

	testArgs = []string{
		"service_venv",
		"pyflyte-execute",
		"--test-opt1",
		"value1",
		"--test-opt2",
		"value2",
		"--test-flag",
	}

	testCmds = []string{
		"test-cmds1",
		"test-cmds2",
	}

	resourceRequirements = &corev1.ResourceRequirements{
		Limits: corev1.ResourceList{
			corev1.ResourceCPU:         resource.MustParse("1000m"),
			corev1.ResourceMemory:      resource.MustParse("1Gi"),
			flytek8s.ResourceNvidiaGPU: resource.MustParse("1"),
		},
		Requests: corev1.ResourceList{
			corev1.ResourceCPU:         resource.MustParse("100m"),
			corev1.ResourceMemory:      resource.MustParse("512Mi"),
			flytek8s.ResourceNvidiaGPU: resource.MustParse("1"),
		},
	}
)

func generateMockTrainingJobTaskTemplate(id string, trainingJobCustomObj *sagemakerIdl.TrainingJob) *flyteIdlCore.TaskTemplate {

	tjObjJSON, err := utils.MarshalToString(trainingJobCustomObj)
	if err != nil {
		panic(err)
	}
	structObj := structpb.Struct{}

	err = jsonpb.UnmarshalString(tjObjJSON, &structObj)
	if err != nil {
		panic(err)
	}

	return &flyteIdlCore.TaskTemplate{
		Id:   &flyteIdlCore.Identifier{Name: id},
		Type: "container",
		Target: &flyteIdlCore.TaskTemplate_Container{
			Container: &flyteIdlCore.Container{
				Command: testCmds,
				Image:   testImage,
				Args:    testArgs,
				Env:     dummyEnvVars,
			},
		},
		Custom: &structObj,
	}
}

func generateMockHyperparameterTuningJobTaskTemplate(id string, hpoJobCustomObj *sagemakerIdl.HyperparameterTuningJob) *flyteIdlCore.TaskTemplate {

	htObjJSON, err := utils.MarshalToString(hpoJobCustomObj)
	if err != nil {
		panic(err)
	}
	structObj := structpb.Struct{}

	err = jsonpb.UnmarshalString(htObjJSON, &structObj)
	if err != nil {
		panic(err)
	}

	return &flyteIdlCore.TaskTemplate{
		Id:   &flyteIdlCore.Identifier{Name: id},
		Type: "container",
		Target: &flyteIdlCore.TaskTemplate_Container{
			Container: &flyteIdlCore.Container{
				Image: testImage,
				Args:  testArgs,
				Env:   dummyEnvVars,
			},
		},
		Custom: &structObj,
		Interface: &flyteIdlCore.TypedInterface{
			Inputs: &flyteIdlCore.VariableMap{
				Variables: map[string]*flyteIdlCore.Variable{
					"input": {
						Type: &flyteIdlCore.LiteralType{
							Type: &flyteIdlCore.LiteralType_CollectionType{
								CollectionType: &flyteIdlCore.LiteralType{Type: &flyteIdlCore.LiteralType_Simple{Simple: flyteIdlCore.SimpleType_INTEGER}},
							},
						},
					},
				},
			},
			Outputs: &flyteIdlCore.VariableMap{
				Variables: map[string]*flyteIdlCore.Variable{
					"output": {
						Type: &flyteIdlCore.LiteralType{
							Type: &flyteIdlCore.LiteralType_CollectionType{
								CollectionType: &flyteIdlCore.LiteralType{Type: &flyteIdlCore.LiteralType_Simple{Simple: flyteIdlCore.SimpleType_INTEGER}},
							},
						},
					},
				},
			},
		},
	}
}

// nolint
func generateMockCustomTrainingJobTaskContext(taskTemplate *flyteIdlCore.TaskTemplate, outputReaderPutError bool) pluginsCore.TaskExecutionContext {
	taskCtx := &mocks.TaskExecutionContext{}
	inputReader := &pluginIOMocks.InputReader{}
	inputReader.OnGetInputPrefixPath().Return("/input/prefix")
	inputReader.OnGetInputPath().Return("/input")

	trainBlobLoc := storage.DataReference("train-blob-loc")
	validationBlobLoc := storage.DataReference("validation-blob-loc")

	inputReader.OnGetMatch(mock.Anything).Return(
		&flyteIdlCore.LiteralMap{
			Literals: map[string]*flyteIdlCore.Literal{
				"train":      generateMockBlobLiteral(trainBlobLoc),
				"validation": generateMockBlobLiteral(validationBlobLoc),
				"hp_int":     coreutils.MustMakeLiteral(1),
				"hp_float":   coreutils.MustMakeLiteral(1.5),
				"hp_bool":    coreutils.MustMakeLiteral(false),
				"hp_string":  coreutils.MustMakeLiteral("a"),
			},
		}, nil)
	taskCtx.OnInputReader().Return(inputReader)

	outputReader := &pluginIOMocks.OutputWriter{}
	outputReader.OnGetOutputPath().Return("/data/outputs.pb")
	outputReader.OnGetOutputPrefixPath().Return("/data/")
	outputReader.OnGetRawOutputPrefix().Return("/raw/")
	outputReader.OnGetCheckpointPrefix().Return("/checkpoint")
	outputReader.OnGetPreviousCheckpointsPrefix().Return("/prev")

	if outputReaderPutError {
		outputReader.OnPutMatch(mock.Anything, mock.Anything).Return(errors.Errorf("err"))
	} else {
		outputReader.OnPutMatch(mock.Anything, mock.Anything).Return(nil)
	}
	taskCtx.OnOutputWriter().Return(outputReader)

	taskReader := &mocks.TaskReader{}
	taskReader.OnReadMatch(mock.Anything).Return(taskTemplate, nil)
	taskCtx.OnTaskReader().Return(taskReader)

	tID := &mocks.TaskExecutionID{}
	tID.OnGetID().Return(flyteIdlCore.TaskExecutionIdentifier{
		NodeExecutionId: &flyteIdlCore.NodeExecutionIdentifier{
			ExecutionId: &flyteIdlCore.WorkflowExecutionIdentifier{
				Name:    "my_name",
				Project: "my_project",
				Domain:  "my_domain",
			},
		},
	})
	tID.OnGetGeneratedName().Return("some-acceptable-name")

	resources := &mocks.TaskOverrides{}
	resources.OnGetResources().Return(resourceRequirements)

	taskExecutionMetadata := &mocks.TaskExecutionMetadata{}
	taskExecutionMetadata.OnGetTaskExecutionID().Return(tID)
	taskExecutionMetadata.OnGetNamespace().Return("test-namespace")
	taskExecutionMetadata.OnGetAnnotations().Return(map[string]string{"iam.amazonaws.com/role": "metadata_role"})
	taskExecutionMetadata.OnGetSecurityContext().Return(flyteIdlCore.SecurityContext{
		RunAs: &flyteIdlCore.Identity{IamRole: "new-role"},
	})

	taskExecutionMetadata.OnGetLabels().Return(map[string]string{"label-1": "val1"})
	taskExecutionMetadata.OnGetOwnerReference().Return(v1.OwnerReference{
		Kind: "node",
		Name: "blah",
	})
	taskExecutionMetadata.OnIsInterruptible().Return(true)
	taskExecutionMetadata.OnGetOverrides().Return(resources)
	taskExecutionMetadata.OnGetK8sServiceAccount().Return(serviceAccount)
	taskCtx.OnTaskExecutionMetadata().Return(taskExecutionMetadata)

	dataStore, err := storage.NewDataStore(&storage.Config{Type: storage.TypeMemory}, promutils.NewTestScope())
	if err != nil {
		panic(err)
	}
	taskCtx.OnDataStore().Return(dataStore)

	taskCtx.OnMaxDatasetSizeBytes().Return(10000)

	return taskCtx
}

// nolint
func generateMockTrainingJobTaskContext(taskTemplate *flyteIdlCore.TaskTemplate, outputReaderPutError bool) pluginsCore.TaskExecutionContext {
	taskCtx := &mocks.TaskExecutionContext{}
	inputReader := &pluginIOMocks.InputReader{}
	inputReader.OnGetInputPrefixPath().Return("/input/prefix")
	inputReader.OnGetInputPath().Return("/input")

	trainBlobLoc := storage.DataReference("train-blob-loc")
	validationBlobLoc := storage.DataReference("validation-blob-loc")
	shp := map[string]string{"a": "1", "b": "2"}
	shpStructObj, _ := utils.MarshalObjToStruct(shp)

	inputReader.OnGetMatch(mock.Anything).Return(
		&flyteIdlCore.LiteralMap{
			Literals: map[string]*flyteIdlCore.Literal{
				"train":                  generateMockBlobLiteral(trainBlobLoc),
				"validation":             generateMockBlobLiteral(validationBlobLoc),
				"static_hyperparameters": coreutils.MakeGenericLiteral(shpStructObj),
			},
		}, nil)
	taskCtx.OnInputReader().Return(inputReader)

	outputReader := &pluginIOMocks.OutputWriter{}
	outputReader.OnGetOutputPath().Return("/data/outputs.pb")
	outputReader.OnGetOutputPrefixPath().Return("/data/")
	outputReader.OnGetRawOutputPrefix().Return("/raw/")
	outputReader.OnGetCheckpointPrefix().Return("/checkpoint")
	outputReader.OnGetPreviousCheckpointsPrefix().Return("/prev")

	if outputReaderPutError {
		outputReader.OnPutMatch(mock.Anything).Return(errors.Errorf("err"))
	}
	taskCtx.OnOutputWriter().Return(outputReader)

	taskReader := &mocks.TaskReader{}
	taskReader.OnReadMatch(mock.Anything).Return(taskTemplate, nil)
	taskCtx.OnTaskReader().Return(taskReader)

	tID := &mocks.TaskExecutionID{}
	tID.OnGetID().Return(flyteIdlCore.TaskExecutionIdentifier{
		NodeExecutionId: &flyteIdlCore.NodeExecutionIdentifier{
			ExecutionId: &flyteIdlCore.WorkflowExecutionIdentifier{
				Name:    "my_name",
				Project: "my_project",
				Domain:  "my_domain",
			},
		},
	})
	tID.OnGetGeneratedName().Return("some-acceptable-name")

	resources := &mocks.TaskOverrides{}
	resources.OnGetResources().Return(resourceRequirements)

	taskExecutionMetadata := &mocks.TaskExecutionMetadata{}
	taskExecutionMetadata.OnGetTaskExecutionID().Return(tID)
	taskExecutionMetadata.OnGetNamespace().Return("test-namespace")
	taskExecutionMetadata.OnGetAnnotations().Return(map[string]string{"iam.amazonaws.com/role": "metadata_role"})
	taskExecutionMetadata.OnGetSecurityContext().Return(flyteIdlCore.SecurityContext{})
	taskExecutionMetadata.OnGetLabels().Return(map[string]string{"label-1": "val1"})
	taskExecutionMetadata.OnGetOwnerReference().Return(v1.OwnerReference{
		Kind: "node",
		Name: "blah",
	})
	taskExecutionMetadata.OnIsInterruptible().Return(true)
	taskExecutionMetadata.OnGetOverrides().Return(resources)
	taskExecutionMetadata.OnGetK8sServiceAccount().Return(serviceAccount)
	taskCtx.OnTaskExecutionMetadata().Return(taskExecutionMetadata)
	return taskCtx
}

func generateMockBlobLiteral(loc storage.DataReference) *flyteIdlCore.Literal {
	return &flyteIdlCore.Literal{
		Value: &flyteIdlCore.Literal_Scalar{
			Scalar: &flyteIdlCore.Scalar{
				Value: &flyteIdlCore.Scalar_Blob{
					Blob: &flyteIdlCore.Blob{
						Uri: loc.String(),
						Metadata: &flyteIdlCore.BlobMetadata{
							Type: &flyteIdlCore.BlobType{
								Dimensionality: flyteIdlCore.BlobType_SINGLE,
								Format:         "csv",
							},
						},
					},
				},
			},
		},
	}
}

func generateMockHyperparameterTuningJobTaskContext(taskTemplate *flyteIdlCore.TaskTemplate) pluginsCore.TaskExecutionContext {
	taskCtx := &mocks.TaskExecutionContext{}
	inputReader := &pluginIOMocks.InputReader{}
	inputReader.OnGetInputPrefixPath().Return("/input/prefix")
	inputReader.OnGetInputPath().Return("/input")

	trainBlobLoc := storage.DataReference("train-blob-loc")
	validationBlobLoc := storage.DataReference("validation-blob-loc")
	shp := map[string]string{"a": "1", "b": "2"}
	shpStructObj, _ := utils.MarshalObjToStruct(shp)
	hpoJobConfig := sagemakerIdl.HyperparameterTuningJobConfig{
		TuningStrategy: sagemakerIdl.HyperparameterTuningStrategy_BAYESIAN,
		TuningObjective: &sagemakerIdl.HyperparameterTuningObjective{
			ObjectiveType: sagemakerIdl.HyperparameterTuningObjectiveType_MINIMIZE,
			MetricName:    "test:metric",
		},
		TrainingJobEarlyStoppingType: sagemakerIdl.TrainingJobEarlyStoppingType_AUTO,
	}
	hpoJobConfigByteArray, _ := proto.Marshal(&hpoJobConfig)

	intParamRange := &structpb.Struct{}
	err := utils.MarshalStruct(&sagemakerIdl.ParameterRangeOneOf{
		ParameterRangeType: &sagemakerIdl.ParameterRangeOneOf_IntegerParameterRange{
			IntegerParameterRange: &sagemakerIdl.IntegerParameterRange{
				MaxValue:    2,
				MinValue:    1,
				ScalingType: sagemakerIdl.HyperparameterScalingType_LINEAR,
			},
		},
	}, intParamRange)

	if err != nil {
		panic(err)
	}

	inputReader.OnGetMatch(mock.Anything).Return(
		&flyteIdlCore.LiteralMap{
			Literals: map[string]*flyteIdlCore.Literal{
				"train":                            generateMockBlobLiteral(trainBlobLoc),
				"validation":                       generateMockBlobLiteral(validationBlobLoc),
				"static_hyperparameters":           coreutils.MakeGenericLiteral(shpStructObj),
				"hyperparameter_tuning_job_config": coreutils.MakeBinaryLiteral(hpoJobConfigByteArray),
				"a":                                coreutils.MakeGenericLiteral(intParamRange),
			},
		}, nil)
	taskCtx.OnInputReader().Return(inputReader)

	outputReader := &pluginIOMocks.OutputWriter{}
	outputReader.OnGetOutputPath().Return("/data/outputs.pb")
	outputReader.OnGetOutputPrefixPath().Return("/data/")
	outputReader.OnGetRawOutputPrefix().Return("/raw/")

	taskCtx.OnOutputWriter().Return(outputReader)

	taskReader := &mocks.TaskReader{}
	taskReader.OnReadMatch(mock.Anything).Return(taskTemplate, nil)
	taskCtx.OnTaskReader().Return(taskReader)
	taskExecutionMetadata := genMockTaskExecutionMetadata()
	taskCtx.OnTaskExecutionMetadata().Return(taskExecutionMetadata)
	return taskCtx
}

func genMockTaskExecutionMetadata() *mocks.TaskExecutionMetadata {
	tID := &mocks.TaskExecutionID{}
	tID.OnGetID().Return(flyteIdlCore.TaskExecutionIdentifier{
		NodeExecutionId: &flyteIdlCore.NodeExecutionIdentifier{
			ExecutionId: &flyteIdlCore.WorkflowExecutionIdentifier{
				Name:    "my_name",
				Project: "my_project",
				Domain:  "my_domain",
			},
		},
	})

	tID.OnGetGeneratedName().Return("some-acceptable-name")

	resources := &mocks.TaskOverrides{}
	resources.OnGetResources().Return(resourceRequirements)

	taskExecutionMetadata := &mocks.TaskExecutionMetadata{}
	taskExecutionMetadata.OnGetTaskExecutionID().Return(tID)
	taskExecutionMetadata.OnGetNamespace().Return("test-namespace")
	taskExecutionMetadata.OnGetAnnotations().Return(map[string]string{"iam.amazonaws.com/role": "metadata_role"})
	taskExecutionMetadata.OnGetSecurityContext().Return(flyteIdlCore.SecurityContext{
		RunAs: &flyteIdlCore.Identity{IamRole: "default_role"},
	})
	taskExecutionMetadata.OnGetLabels().Return(map[string]string{"label-1": "val1"})
	taskExecutionMetadata.OnGetOwnerReference().Return(v1.OwnerReference{
		Kind: "node",
		Name: "blah",
	})
	taskExecutionMetadata.OnIsInterruptible().Return(true)
	taskExecutionMetadata.OnGetOverrides().Return(resources)
	taskExecutionMetadata.OnGetK8sServiceAccount().Return(serviceAccount)
	return taskExecutionMetadata
}

// nolint
func generateMockTrainingJobCustomObj(
	inputMode sagemakerIdl.InputMode_Value, algName sagemakerIdl.AlgorithmName_Value, algVersion string,
	metricDefinitions []*sagemakerIdl.MetricDefinition, contentType sagemakerIdl.InputContentType_Value,
	instanceCount int64, instanceType string, volumeSizeInGB int64, protocol sagemakerIdl.DistributedProtocol_Value) *sagemakerIdl.TrainingJob {
	return &sagemakerIdl.TrainingJob{
		AlgorithmSpecification: &sagemakerIdl.AlgorithmSpecification{
			InputMode:         inputMode,
			AlgorithmName:     algName,
			AlgorithmVersion:  algVersion,
			MetricDefinitions: metricDefinitions,
			InputContentType:  contentType,
		},
		TrainingJobResourceConfig: &sagemakerIdl.TrainingJobResourceConfig{
			InstanceCount:       instanceCount,
			InstanceType:        instanceType,
			VolumeSizeInGb:      volumeSizeInGB,
			DistributedProtocol: protocol,
		},
	}
}

func generateMockHyperparameterTuningJobCustomObj(
	trainingJob *sagemakerIdl.TrainingJob, maxNumberOfTrainingJobs int64, maxParallelTrainingJobs int64) *sagemakerIdl.HyperparameterTuningJob {
	return &sagemakerIdl.HyperparameterTuningJob{
		TrainingJob:             trainingJob,
		MaxNumberOfTrainingJobs: maxNumberOfTrainingJobs,
		MaxParallelTrainingJobs: maxParallelTrainingJobs,
	}
}
