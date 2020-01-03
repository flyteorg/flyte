package validation

import (
	"context"
	"errors"
	"fmt"
	"testing"

	"github.com/lyft/flyteadmin/pkg/repositories/mocks"
	"github.com/lyft/flyteidl/gen/pb-go/flyteidl/admin"

	"github.com/golang/protobuf/proto"

	"github.com/lyft/flyteidl/gen/pb-go/flyteidl/core"
	"k8s.io/apimachinery/pkg/api/resource"

	"github.com/lyft/flyteadmin/pkg/manager/impl/testutils"
	runtimeInterfaces "github.com/lyft/flyteadmin/pkg/runtime/interfaces"
	runtimeMocks "github.com/lyft/flyteadmin/pkg/runtime/mocks"
	"github.com/stretchr/testify/assert"
)

func getMockTaskConfigProvider() runtimeInterfaces.TaskResourceConfiguration {
	var taskConfig = runtimeMocks.MockTaskResourceConfiguration{}
	taskConfig.Limits = runtimeInterfaces.TaskResourceSet{
		Memory: "500Mi",
		CPU:    "200m",
		GPU:    "8",
	}

	return &taskConfig
}

var mockWhitelistConfigProvider = runtimeMocks.NewMockWhitelistConfiguration()
var taskApplicationConfigProvider = testutils.GetApplicationConfigWithDefaultProjects()

func TestValidateTaskEmptyProject(t *testing.T) {
	request := testutils.GetValidTaskRequest()
	request.Id.Project = ""
	err := ValidateTask(context.Background(), request, testutils.GetRepoWithDefaultProject(),
		getMockTaskConfigProvider(), mockWhitelistConfigProvider, taskApplicationConfigProvider)
	assert.EqualError(t, err, "missing project")
}

func TestValidateTaskInvalidProjectAndDomain(t *testing.T) {
	request := testutils.GetValidTaskRequest()
	err := ValidateTask(context.Background(), request, testutils.GetRepoWithDefaultProjectAndErr(errors.New("foo")),
		getMockTaskConfigProvider(), mockWhitelistConfigProvider, taskApplicationConfigProvider)
	assert.EqualError(t, err, "failed to validate that project [project] and domain [domain] are registered, err: [foo]")
}

func TestValidateTaskEmptyDomain(t *testing.T) {
	request := testutils.GetValidTaskRequest()
	request.Id.Domain = ""
	err := ValidateTask(context.Background(), request, testutils.GetRepoWithDefaultProject(),
		getMockTaskConfigProvider(), mockWhitelistConfigProvider, taskApplicationConfigProvider)
	assert.EqualError(t, err, "missing domain")
}

func TestValidateTaskEmptyName(t *testing.T) {
	request := testutils.GetValidTaskRequest()
	request.Id.Name = ""
	err := ValidateTask(context.Background(), request, testutils.GetRepoWithDefaultProject(),
		getMockTaskConfigProvider(), mockWhitelistConfigProvider, taskApplicationConfigProvider)
	assert.EqualError(t, err, "missing name")
}

func TestValidateTaskEmptyVersion(t *testing.T) {
	request := testutils.GetValidTaskRequest()
	request.Id.Version = ""
	err := ValidateTask(context.Background(), request, testutils.GetRepoWithDefaultProject(),
		getMockTaskConfigProvider(), mockWhitelistConfigProvider, taskApplicationConfigProvider)
	assert.EqualError(t, err, "missing version")
}

func TestValidateTaskEmptyType(t *testing.T) {
	request := testutils.GetValidTaskRequest()
	request.Spec.Template.Type = ""
	err := ValidateTask(context.Background(), request, testutils.GetRepoWithDefaultProject(),
		getMockTaskConfigProvider(), mockWhitelistConfigProvider, taskApplicationConfigProvider)
	assert.EqualError(t, err, "missing type")
}

func TestValidateTaskEmptyMetadata(t *testing.T) {
	request := testutils.GetValidTaskRequest()
	request.Spec.Template.Metadata = nil
	err := ValidateTask(context.Background(), request, testutils.GetRepoWithDefaultProject(),
		getMockTaskConfigProvider(), mockWhitelistConfigProvider, taskApplicationConfigProvider)
	assert.EqualError(t, err, "missing metadata")
}

func TestValidateTaskEmptyRuntimeVersion(t *testing.T) {
	request := testutils.GetValidTaskRequest()
	request.Spec.Template.Metadata.Runtime.Version = ""
	err := ValidateTask(context.Background(), request, testutils.GetRepoWithDefaultProject(),
		getMockTaskConfigProvider(), mockWhitelistConfigProvider, taskApplicationConfigProvider)
	assert.EqualError(t, err, "missing runtime version")
}

func TestValidateTaskEmptyTypedInterface(t *testing.T) {
	request := testutils.GetValidTaskRequest()
	request.Spec.Template.Interface = nil
	err := ValidateTask(context.Background(), request, testutils.GetRepoWithDefaultProject(),
		getMockTaskConfigProvider(), mockWhitelistConfigProvider, taskApplicationConfigProvider)
	assert.EqualError(t, err, "missing typed interface")
}

func TestValidateTaskEmptyContainer(t *testing.T) {
	request := testutils.GetValidTaskRequest()
	request.Spec.Template.Target = nil
	err := ValidateTask(context.Background(), request, testutils.GetRepoWithDefaultProject(),
		getMockTaskConfigProvider(), mockWhitelistConfigProvider, taskApplicationConfigProvider)
	assert.Nil(t, err)
}

func TestValidateTaskEmptyImage(t *testing.T) {
	request := testutils.GetValidTaskRequest()
	request.Spec.Template.GetContainer().Image = ""
	err := ValidateTask(context.Background(), request, testutils.GetRepoWithDefaultProject(),
		getMockTaskConfigProvider(), mockWhitelistConfigProvider, taskApplicationConfigProvider)
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
		CPU:    "100Mi",
		GPU:    "2",
		Memory: "1.5Gi",
	}
	resourceSetMap := taskResourceSetToMap(resourceSet)
	assert.Len(t, resourceSetMap, 3)
	assert.Equal(t, resourceSetMap[core.Resources_CPU].Value(), int64(104857600))
	assert.Equal(t, resourceSetMap[core.Resources_GPU].Value(), int64(2))
	assert.Equal(t, resourceSetMap[core.Resources_MEMORY].Value(), int64(1610612736))
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
	assert.Contains(t, err.Error(), "Invalid quantity")

	quantity = resourceEntries[core.Resources_CPU]
	val = quantity.Value()
	assert.Equal(t, val, int64(104857600), "Existing values in the resource entry map should not be overwritten")
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
	}
	assert.Nil(t, validateTaskResources(&core.Identifier{}, runtimeInterfaces.TaskResourceSet{},
		requestedTaskResourceDefaults, requestedTaskResourceLimits))
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
	assert.EqualError(t, err, "Resource CPU for [name:\"name\" ] cannot set default > limit")
}

func TestValidateTaskResources_LimitGreaterThanConfig(t *testing.T) {
	err := validateTaskResources(&core.Identifier{
		Name: "name",
	}, runtimeInterfaces.TaskResourceSet{
		CPU: "1Gi",
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
	assert.EqualError(t, err, "Resource CPU for [name:\"name\" ] cannot set limit > platform limit")
}

func TestValidateTaskResources_DefaultGreaterThanConfig(t *testing.T) {
	err := validateTaskResources(&core.Identifier{
		Name: "name",
	}, runtimeInterfaces.TaskResourceSet{
		CPU: "1Gi",
	},
		[]*core.Resources_ResourceEntry{
			{
				Name:  core.Resources_CPU,
				Value: "1.5Gi",
			},
		}, []*core.Resources_ResourceEntry{})
	assert.EqualError(t, err, "Resource CPU for [name:\"name\" ] cannot set default > platform limit")
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
		GPU: "1",
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
	assert.EqualError(t, err, "Resource GPU for [name:\"name\" ] cannot set default > platform limit")
}

func TestValidateTaskResources_GPUDefaultGreaterThanConfig(t *testing.T) {
	err := validateTaskResources(&core.Identifier{
		Name: "name",
	}, runtimeInterfaces.TaskResourceSet{
		GPU: "1",
	},
		[]*core.Resources_ResourceEntry{
			{
				Name:  core.Resources_GPU,
				Value: "2",
			},
		}, []*core.Resources_ResourceEntry{})
	assert.EqualError(t, err, "Resource GPU for [name:\"name\" ] cannot set default > platform limit")
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

func TestAssignResourcesIfUnset(t *testing.T) {
	platformValues := runtimeInterfaces.TaskResourceSet{
		CPU:    "200m",
		GPU:    "8",
		Memory: "200Gi",
	}
	taskResourceSpec := &admin.TaskResourceSpec{
		Cpu:    "400m",
		Memory: "400Gi",
	}
	assignedResources := assignResourcesIfUnset(context.Background(), &core.Identifier{
		Project: "project",
		Domain:  "domain",
		Name:    "name",
		Version: "version",
	}, platformValues, []*core.Resources_ResourceEntry{}, taskResourceSpec)

	assert.EqualValues(t, []*core.Resources_ResourceEntry{
		{
			Name:  core.Resources_CPU,
			Value: taskResourceSpec.Cpu,
		},
		{
			Name:  core.Resources_MEMORY,
			Value: taskResourceSpec.Memory,
		},
	}, assignedResources)
}

func TestSetDefaults(t *testing.T) {
	task := &core.CompiledTask{
		Template: &core.TaskTemplate{
			Target: &core.TaskTemplate_Container{
				Container: &core.Container{
					Resources: &core.Resources{
						Requests: []*core.Resources_ResourceEntry{
							{
								Name:  core.Resources_CPU,
								Value: "200m",
							},
						},
					},
				},
			},
			Id: &core.Identifier{
				Project: "project",
				Domain:  "domain",
				Name:    "task_name",
				Version: "version",
			},
		},
	}

	taskConfig := runtimeMocks.MockTaskResourceConfiguration{}
	taskConfig.Defaults = runtimeInterfaces.TaskResourceSet{
		CPU:    "200m",
		GPU:    "8",
		Memory: "200Gi",
	}
	taskConfig.Limits = runtimeInterfaces.TaskResourceSet{
		CPU:    "300m",
		GPU:    "8",
		Memory: "500Gi",
	}
	SetDefaults(context.Background(), &taskConfig, task, mocks.NewMockRepository(), "workflow")
	assert.True(t, proto.Equal(
		&core.Container{
			Resources: &core.Resources{
				Requests: []*core.Resources_ResourceEntry{
					{
						Name:  core.Resources_CPU,
						Value: "200m",
					},
					{
						Name:  core.Resources_MEMORY,
						Value: "200Gi",
					},
				},
				Limits: []*core.Resources_ResourceEntry{
					{
						Name:  core.Resources_CPU,
						Value: "200m",
					},
					{
						Name:  core.Resources_MEMORY,
						Value: "200Gi",
					},
				},
			},
		},
		task.Template.GetContainer()), fmt.Sprintf("%+v", task.Template.GetContainer()))
}

func TestSetDefaults_MissingDefaults(t *testing.T) {
	task := &core.CompiledTask{
		Template: &core.TaskTemplate{
			Target: &core.TaskTemplate_Container{
				Container: &core.Container{
					Resources: &core.Resources{
						Requests: []*core.Resources_ResourceEntry{
							{
								Name:  core.Resources_CPU,
								Value: "200m",
							},
						},
					},
				},
			},
			Id: &core.Identifier{
				Project: "project",
				Domain:  "domain",
				Name:    "task_name",
				Version: "version",
			},
		},
	}

	taskConfig := runtimeMocks.MockTaskResourceConfiguration{}
	taskConfig.Defaults = runtimeInterfaces.TaskResourceSet{
		CPU:    "200m",
		GPU:    "8",
		Memory: "200Gi",
	}
	taskConfig.Limits = runtimeInterfaces.TaskResourceSet{
		CPU: "300m",
		GPU: "8",
	}
	SetDefaults(context.Background(), &taskConfig, task, mocks.NewMockRepository(), "workflow")
	assert.True(t, proto.Equal(
		&core.Container{
			Resources: &core.Resources{
				Requests: []*core.Resources_ResourceEntry{
					{
						Name:  core.Resources_CPU,
						Value: "200m",
					},
					{
						Name:  core.Resources_MEMORY,
						Value: "200Gi",
					},
				},
				Limits: []*core.Resources_ResourceEntry{
					{
						Name:  core.Resources_CPU,
						Value: "200m",
					},
					{
						Name:  core.Resources_MEMORY,
						Value: "200Gi",
					},
				},
			},
		},
		task.Template.GetContainer()), fmt.Sprintf("%+v", task.Template.GetContainer()))
}

func TestCreateTaskDefaultLimits(t *testing.T) {
	task := &core.CompiledTask{
		Template: &core.TaskTemplate{
			Target: &core.TaskTemplate_Container{
				Container: &core.Container{
					Resources: &core.Resources{
						Requests: []*core.Resources_ResourceEntry{
							{
								Name:  core.Resources_CPU,
								Value: "200m",
							},
							{
								Name:  core.Resources_MEMORY,
								Value: "200Mi",
							},
						},
					},
				},
			},
		},
	}

	defaultLimits := createTaskDefaultLimits(context.Background(), task)
	assert.Equal(t, "200Mi", defaultLimits.Memory)
	assert.Equal(t, "200m", defaultLimits.CPU)
}
