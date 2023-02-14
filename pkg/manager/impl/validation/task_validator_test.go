package validation

import (
	"context"
	"encoding/json"
	"errors"
	"testing"

	"github.com/flyteorg/flyteadmin/pkg/manager/impl/testutils"
	runtimeInterfaces "github.com/flyteorg/flyteadmin/pkg/runtime/interfaces"
	runtimeMocks "github.com/flyteorg/flyteadmin/pkg/runtime/mocks"
	workflowengineInterfaces "github.com/flyteorg/flyteadmin/pkg/workflowengine/interfaces"
	"github.com/flyteorg/flyteidl/gen/pb-go/flyteidl/core"

	"github.com/stretchr/testify/assert"

	"google.golang.org/protobuf/types/known/structpb"

	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
)

func getMockTaskResources() workflowengineInterfaces.TaskResources {
	return workflowengineInterfaces.TaskResources{
		Limits: runtimeInterfaces.TaskResourceSet{
			Memory: resource.MustParse("500Mi"),
			CPU:    resource.MustParse("200m"),
			GPU:    resource.MustParse("8"),
		},
	}
}

var mockWhitelistConfigProvider = runtimeMocks.NewMockWhitelistConfiguration()
var taskApplicationConfigProvider = testutils.GetApplicationConfigWithDefaultDomains()

func TestValidateTask(t *testing.T) {
	request := testutils.GetValidTaskRequest()
	resources := []*core.Resources_ResourceEntry{
		{
			Name:  core.Resources_CPU,
			Value: "1.5Gi",
		},
		{
			Name:  core.Resources_MEMORY,
			Value: "200m",
		},
	}
	request.Spec.Template.GetContainer().Resources = &core.Resources{Requests: resources}
	err := ValidateTask(context.Background(), request, testutils.GetRepoWithDefaultProject(),
		getMockTaskResources(), mockWhitelistConfigProvider, taskApplicationConfigProvider)
	assert.EqualError(t, err, "Requested CPU default [1536Mi] is greater than current limit set in the platform configuration [200m]. Please contact Flyte Admins to change these limits or consult the configuration")

	request.Spec.Template.Target = &core.TaskTemplate_K8SPod{K8SPod: &core.K8SPod{}}
	err = ValidateTask(context.Background(), request, testutils.GetRepoWithDefaultProject(),
		getMockTaskResources(), mockWhitelistConfigProvider, taskApplicationConfigProvider)
	assert.EqualError(t, err, "invalid TaskSpecification, pod tasks should specify their target as a K8sPod with a defined pod spec")

	resourceList := corev1.ResourceList{corev1.ResourceCPU: resource.MustParse("1.5Gi")}
	podSpec := &corev1.PodSpec{Containers: []corev1.Container{{Resources: corev1.ResourceRequirements{Requests: resourceList}}}}
	request.Spec.Template.Target = &core.TaskTemplate_K8SPod{K8SPod: &core.K8SPod{PodSpec: transformStructToStructPB(t, podSpec)}}
	err = ValidateTask(context.Background(), request, testutils.GetRepoWithDefaultProject(),
		getMockTaskResources(), mockWhitelistConfigProvider, taskApplicationConfigProvider)
	assert.EqualError(t, err, "Requested CPU default [1536Mi] is greater than current limit set in the platform configuration [200m]. Please contact Flyte Admins to change these limits or consult the configuration")

	resourceList = corev1.ResourceList{corev1.ResourceCPU: resource.MustParse("200m")}
	podSpec = &corev1.PodSpec{Containers: []corev1.Container{{Resources: corev1.ResourceRequirements{Requests: resourceList}}}}
	request.Spec.Template.Target = &core.TaskTemplate_K8SPod{K8SPod: &core.K8SPod{PodSpec: transformStructToStructPB(t, podSpec)}}
	err = ValidateTask(context.Background(), request, testutils.GetRepoWithDefaultProject(),
		getMockTaskResources(), mockWhitelistConfigProvider, taskApplicationConfigProvider)
	assert.Nil(t, err)
}

func transformStructToStructPB(t *testing.T, obj interface{}) *structpb.Struct {
	data, err := json.Marshal(obj)
	assert.Nil(t, err)
	podSpecMap := make(map[string]interface{})
	err = json.Unmarshal(data, &podSpecMap)
	assert.Nil(t, err)
	s, err := structpb.NewStruct(podSpecMap)
	assert.Nil(t, err)
	return s
}

func TestValidateTaskEmptyProject(t *testing.T) {
	request := testutils.GetValidTaskRequest()
	request.Id.Project = ""
	err := ValidateTask(context.Background(), request, testutils.GetRepoWithDefaultProject(),
		getMockTaskResources(), mockWhitelistConfigProvider, taskApplicationConfigProvider)
	assert.EqualError(t, err, "missing project")
}

func TestValidateTaskInvalidProjectAndDomain(t *testing.T) {
	request := testutils.GetValidTaskRequest()
	err := ValidateTask(context.Background(), request, testutils.GetRepoWithDefaultProjectAndErr(errors.New("foo")),
		getMockTaskResources(), mockWhitelistConfigProvider, taskApplicationConfigProvider)
	assert.EqualError(t, err, "failed to validate that project [project] and domain [domain] are registered, err: [foo]")
}

func TestValidateTaskEmptyDomain(t *testing.T) {
	request := testutils.GetValidTaskRequest()
	request.Id.Domain = ""
	err := ValidateTask(context.Background(), request, testutils.GetRepoWithDefaultProject(),
		getMockTaskResources(), mockWhitelistConfigProvider, taskApplicationConfigProvider)
	assert.EqualError(t, err, "missing domain")
}

func TestValidateTaskEmptyName(t *testing.T) {
	request := testutils.GetValidTaskRequest()
	request.Id.Name = ""
	err := ValidateTask(context.Background(), request, testutils.GetRepoWithDefaultProject(),
		getMockTaskResources(), mockWhitelistConfigProvider, taskApplicationConfigProvider)
	assert.EqualError(t, err, "missing name")
}

func TestValidateTaskEmptyVersion(t *testing.T) {
	request := testutils.GetValidTaskRequest()
	request.Id.Version = ""
	err := ValidateTask(context.Background(), request, testutils.GetRepoWithDefaultProject(),
		getMockTaskResources(), mockWhitelistConfigProvider, taskApplicationConfigProvider)
	assert.EqualError(t, err, "missing version")
}

func TestValidateTaskEmptyType(t *testing.T) {
	request := testutils.GetValidTaskRequest()
	request.Spec.Template.Type = ""
	err := ValidateTask(context.Background(), request, testutils.GetRepoWithDefaultProject(),
		getMockTaskResources(), mockWhitelistConfigProvider, taskApplicationConfigProvider)
	assert.EqualError(t, err, "missing type")
}

func TestValidateTaskEmptyMetadata(t *testing.T) {
	request := testutils.GetValidTaskRequest()
	request.Spec.Template.Metadata = nil
	err := ValidateTask(context.Background(), request, testutils.GetRepoWithDefaultProject(),
		getMockTaskResources(), mockWhitelistConfigProvider, taskApplicationConfigProvider)
	assert.EqualError(t, err, "missing metadata")
}

func TestValidateTaskEmptyRuntimeVersion(t *testing.T) {
	request := testutils.GetValidTaskRequest()
	request.Spec.Template.Metadata.Runtime.Version = ""
	err := ValidateTask(context.Background(), request, testutils.GetRepoWithDefaultProject(),
		getMockTaskResources(), mockWhitelistConfigProvider, taskApplicationConfigProvider)
	assert.EqualError(t, err, "missing runtime version")
}

func TestValidateTaskEmptyTypedInterface(t *testing.T) {
	request := testutils.GetValidTaskRequest()
	request.Spec.Template.Interface = nil
	err := ValidateTask(context.Background(), request, testutils.GetRepoWithDefaultProject(),
		getMockTaskResources(), mockWhitelistConfigProvider, taskApplicationConfigProvider)
	assert.EqualError(t, err, "missing typed interface")
}

func TestValidateTaskEmptyContainer(t *testing.T) {
	request := testutils.GetValidTaskRequest()
	request.Spec.Template.Target = nil
	err := ValidateTask(context.Background(), request, testutils.GetRepoWithDefaultProject(),
		getMockTaskResources(), mockWhitelistConfigProvider, taskApplicationConfigProvider)
	assert.Nil(t, err)
}

func TestValidateTaskEmptyImage(t *testing.T) {
	request := testutils.GetValidTaskRequest()
	request.Spec.Template.GetContainer().Image = ""
	err := ValidateTask(context.Background(), request, testutils.GetRepoWithDefaultProject(),
		getMockTaskResources(), mockWhitelistConfigProvider, taskApplicationConfigProvider)
	assert.EqualError(t, err, "missing image")
}

func TestValidateTaskTypeWhitelist(t *testing.T) {
	whitelistConfig := runtimeMocks.NewMockWhitelistConfiguration()
	whitelistConfig.(*runtimeMocks.MockWhitelistConfiguration).TaskTypeWhitelist = runtimeInterfaces.TaskTypeWhitelist{
		"every_type": {},
		"type_a": {
			{
				Project: "proj_a",
			},
		},
		"type_b": {
			{
				Project: "proj_b",
				Domain:  "domain_b",
			},
			{
				Project: "proj_c",
			},
		},
	}
	err := validateTaskType(core.Identifier{
		Project: "proj_a",
		Domain:  "domain_a",
	}, "type_a", whitelistConfig)
	assert.Nil(t, err)

	err = validateTaskType(core.Identifier{
		Project: "proj_b",
		Domain:  "domain_a",
	}, "type_b", whitelistConfig)
	assert.NotNil(t, err)

	err = validateTaskType(core.Identifier{
		Project: "proj_b",
		Domain:  "domain_b",
	}, "type_a", whitelistConfig)
	assert.NotNil(t, err)

	err = validateTaskType(core.Identifier{
		Project: "proj_b",
		Domain:  "domain_b",
	}, "type_b", whitelistConfig)
	assert.Nil(t, err)

	err = validateTaskType(core.Identifier{
		Project: "proj_c",
	}, "every_type", whitelistConfig)
	assert.Nil(t, err)

	err = validateTaskType(core.Identifier{
		Project: "proj_c",
	}, "type_b", whitelistConfig)
	assert.Nil(t, err)

	err = validateTaskType(core.Identifier{}, "some_generally_supported_type", whitelistConfig)
	assert.Nil(t, err)
}

func TestTaskResourceSetToMap(t *testing.T) {
	resourceSet := runtimeInterfaces.TaskResourceSet{
		CPU:              resource.MustParse("100Mi"),
		GPU:              resource.MustParse("2"),
		Memory:           resource.MustParse("1.5Gi"),
		EphemeralStorage: resource.MustParse("500Mi"),
	}
	resourceSetMap := taskResourceSetToMap(resourceSet)
	assert.Len(t, resourceSetMap, 4)
	assert.Equal(t, resourceSetMap[core.Resources_CPU].Value(), int64(104857600))
	assert.Equal(t, resourceSetMap[core.Resources_GPU].Value(), int64(2))
	assert.Equal(t, resourceSetMap[core.Resources_MEMORY].Value(), int64(1610612736))
	assert.Equal(t, resourceSetMap[core.Resources_EPHEMERAL_STORAGE].Value(), int64(524288000))
}

func TestAddResourceEntryToMap(t *testing.T) {
	resourceEntries := make(map[core.Resources_ResourceName]resource.Quantity)
	resourceEntries[core.Resources_CPU] = resource.MustParse("100Mi")

	err := addResourceEntryToMap(&core.Identifier{}, &core.Resources_ResourceEntry{
		Name:  core.Resources_GPU,
		Value: "2",
	}, &resourceEntries)
	assert.Nil(t, err)
	quantity := resourceEntries[core.Resources_GPU]
	val := quantity.Value()
	assert.Equal(t, val, int64(2))

	err = addResourceEntryToMap(&core.Identifier{}, &core.Resources_ResourceEntry{
		Name:  core.Resources_GPU,
		Value: "2",
	}, &resourceEntries)
	assert.Contains(t, err.Error(), "multiple times")

	err = addResourceEntryToMap(&core.Identifier{}, &core.Resources_ResourceEntry{
		Name:  core.Resources_MEMORY,
		Value: "foo",
	}, &resourceEntries)
	assert.Contains(t, err.Error(), "Parsing of MEMORY request failed")

	quantity = resourceEntries[core.Resources_CPU]
	val = quantity.Value()
	assert.Equal(t, val, int64(104857600), "Existing values in the resource entry map should not be overwritten")
}

func TestResourceListToQuantity(t *testing.T) {
	cpuResources := resourceListToQuantity(corev1.ResourceList{corev1.ResourceCPU: resource.MustParse("100Mi")})
	cpuQuantity := cpuResources[core.Resources_CPU]
	val := cpuQuantity.Value()
	assert.Equal(t, val, int64(104857600))

	gpuResources := resourceListToQuantity(corev1.ResourceList{corev1.ResourceCPU: resource.MustParse("2")})
	gpuQuantity := gpuResources[core.Resources_CPU]
	val = gpuQuantity.Value()
	assert.Equal(t, val, int64(2))
}

func TestRequestedResourcesToQuantity(t *testing.T) {
	resources, err := requestedResourcesToQuantity(&core.Identifier{}, []*core.Resources_ResourceEntry{
		{
			Name:  core.Resources_CPU,
			Value: "100Mi",
		},
		{
			Name:  core.Resources_GPU,
			Value: "2",
		},
	})
	assert.Nil(t, err)
	cpuQuantity := resources[core.Resources_CPU]
	val := cpuQuantity.Value()
	assert.Equal(t, val, int64(104857600))
	gpuQuantity := resources[core.Resources_GPU]
	val = gpuQuantity.Value()
	assert.Equal(t, val, int64(2))
}

func TestRequestedResourcesToQuantity_InvalidValues(t *testing.T) {
	_, err := requestedResourcesToQuantity(&core.Identifier{}, []*core.Resources_ResourceEntry{
		{
			Name:  core.Resources_CPU,
			Value: "100foo",
		},
	})
	assert.NotNil(t, err)

	_, err = requestedResourcesToQuantity(&core.Identifier{}, []*core.Resources_ResourceEntry{
		{
			Name:  core.Resources_GPU,
			Value: "100n",
		},
	})
	assert.NotNil(t, err)
}

func TestValidateTaskResources(t *testing.T) {
	requestedTaskResourceDefaults := []*core.Resources_ResourceEntry{
		{
			Name:  core.Resources_CPU,
			Value: "100Mi",
		},
		{
			Name:  core.Resources_GPU,
			Value: "2",
		},
		{
			Name:  core.Resources_EPHEMERAL_STORAGE,
			Value: "500Mi",
		},
	}

	requestedTaskResourceLimits := []*core.Resources_ResourceEntry{
		{
			Name:  core.Resources_CPU,
			Value: "200Mi",
		},
		{
			Name:  core.Resources_GPU,
			Value: "2",
		},
		{
			Name:  core.Resources_EPHEMERAL_STORAGE,
			Value: "501Mi",
		},
	}
	assert.Nil(t, validateTaskResources(&core.Identifier{}, runtimeInterfaces.TaskResourceSet{},
		requestedTaskResourceDefaults, requestedTaskResourceLimits))
}

func TestValidateTaskResources_ParsingIssue(t *testing.T) {
	err := validateTaskResources(&core.Identifier{
		Name: "name",
	}, runtimeInterfaces.TaskResourceSet{},
		[]*core.Resources_ResourceEntry{
			{
				Name:  core.Resources_CPU,
				Value: "200Q",
			},
		}, []*core.Resources_ResourceEntry{
			{
				Name:  core.Resources_CPU,
				Value: "200Q",
			},
		})
	assert.EqualError(t, err, "Parsing of CPU request failed for value 200Q - reason  quantities must match the regular expression '^([+-]?[0-9.]+)([eEinumkKMGTP]*[-+]?[0-9]*)$'. Please follow K8s conventions for resources https://kubernetes.io/docs/concepts/configuration/manage-resources-containers/")
}

func TestValidateTaskResources_LimitLessThanRequested(t *testing.T) {
	err := validateTaskResources(&core.Identifier{
		Name: "name",
	}, runtimeInterfaces.TaskResourceSet{},
		[]*core.Resources_ResourceEntry{
			{
				Name:  core.Resources_CPU,
				Value: "1.5Gi",
			},
		}, []*core.Resources_ResourceEntry{
			{
				Name:  core.Resources_CPU,
				Value: "1Gi",
			},
		})
	assert.EqualError(t, err, "Requested CPU default [1536Mi] is greater than the limit [1Gi]. Please fix your configuration")
}

func TestValidateTaskResources_LimitGreaterThanConfig(t *testing.T) {
	err := validateTaskResources(&core.Identifier{
		Name: "name",
	}, runtimeInterfaces.TaskResourceSet{
		CPU: resource.MustParse("1Gi"),
	},
		[]*core.Resources_ResourceEntry{
			{
				Name:  core.Resources_CPU,
				Value: "1.5Gi",
			},
		}, []*core.Resources_ResourceEntry{
			{
				Name:  core.Resources_CPU,
				Value: "1.5Gi",
			},
		})
	assert.EqualError(t, err, "Requested CPU limit [1536Mi] is greater than current limit set in the platform configuration [1Gi]. Please contact Flyte Admins to change these limits or consult the configuration")
}

func TestValidateTaskResources_DefaultGreaterThanConfig(t *testing.T) {
	err := validateTaskResources(&core.Identifier{
		Name: "name",
	}, runtimeInterfaces.TaskResourceSet{
		CPU: resource.MustParse("1Gi"),
	},
		[]*core.Resources_ResourceEntry{
			{
				Name:  core.Resources_CPU,
				Value: "1.5Gi",
			},
		}, []*core.Resources_ResourceEntry{})
	assert.EqualError(t, err, "Requested CPU default [1536Mi] is greater than current limit set in the platform configuration [1Gi]. Please contact Flyte Admins to change these limits or consult the configuration")
}

func TestValidateTaskResources_GPULimitNotEqualToRequested(t *testing.T) {
	err := validateTaskResources(&core.Identifier{
		Name: "name",
	}, runtimeInterfaces.TaskResourceSet{},
		[]*core.Resources_ResourceEntry{
			{
				Name:  core.Resources_GPU,
				Value: "2",
			},
		}, []*core.Resources_ResourceEntry{
			{
				Name:  core.Resources_GPU,
				Value: "1",
			},
		})
	assert.EqualError(t, err,
		"For extended resource 'gpu' the default value must equal the limit value for task [name:\"name\" ]")
}

func TestValidateTaskResources_GPULimitGreaterThanConfig(t *testing.T) {
	err := validateTaskResources(&core.Identifier{
		Name: "name",
	}, runtimeInterfaces.TaskResourceSet{
		GPU: resource.MustParse("1"),
	},
		[]*core.Resources_ResourceEntry{
			{
				Name:  core.Resources_GPU,
				Value: "2",
			},
		}, []*core.Resources_ResourceEntry{
			{
				Name:  core.Resources_GPU,
				Value: "2",
			},
		})
	assert.EqualError(t, err, "Requested GPU default [2] is greater than current limit set in the platform configuration [1]. Please contact Flyte Admins to change these limits or consult the configuration")
}

func TestValidateTaskResources_GPUDefaultGreaterThanConfig(t *testing.T) {
	err := validateTaskResources(&core.Identifier{
		Name: "name",
	}, runtimeInterfaces.TaskResourceSet{
		GPU: resource.MustParse("1"),
	},
		[]*core.Resources_ResourceEntry{
			{
				Name:  core.Resources_GPU,
				Value: "2",
			},
		}, []*core.Resources_ResourceEntry{})
	assert.EqualError(t, err, "Requested GPU default [2] is greater than current limit set in the platform configuration [1]. Please contact Flyte Admins to change these limits or consult the configuration")
}

func TestIsWholeNumber(t *testing.T) {
	wholeNumbers := []string{
		"1Mi",
		"1M",
		"2Gi",
		"2G",
		"3Ti",
		"3T",
		"4",
		"7000m",
	}
	fractions := []string{
		"100n",
		"7m",
	}
	for _, wholeNumber := range wholeNumbers {
		assert.True(t, isWholeNumber(resource.MustParse(wholeNumber)),
			"%s should be treated as a whole number", wholeNumber)
	}
	for _, fraction := range fractions {
		assert.False(t, isWholeNumber(resource.MustParse(fraction)),
			"%s should not be treated as a whole number", fraction)
	}
}
