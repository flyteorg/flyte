package config

import (
	"context"
	"os"
	"path/filepath"
	"testing"

	"github.com/flyteorg/flyte/flyteadmin/pkg/repositories/models"
	"github.com/flyteorg/flyte/flyteidl/gen/pb-go/flyteidl/admin"
	"github.com/golang/protobuf/proto"
	"github.com/stretchr/testify/assert"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

func GetDbForResourcesTest(t *testing.T) *gorm.DB {
	tempDir, err := os.MkdirTemp("", "test_db_")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}

	gormDb, err := gorm.Open(sqlite.Open(filepath.Join(tempDir, "gorm.db")), &gorm.Config{})
	if err != nil {
		t.Fatalf("Failed to open db: %v", err)
	}

	err = gormDb.AutoMigrate(&models.Resource{})
	if err != nil {
		t.Fatalf("Failed to migrate db: %v", err)
	}

	t.Cleanup(func() {
		os.RemoveAll(tempDir)
	})

	return gormDb
}

func TestSeedMatchableResourcesFromConfig(t *testing.T) {
	ctx := context.TODO()
	t.Run("Empty config", func(t *testing.T) {
		gormDb := GetDbForResourcesTest(t)

		config := MatcheableResourcesConfig{
			ExecutionClusterLabels: ExecutionClusterLabelConfig{
				Values: []ExecutionClusterLabel{},
			},
			TaskResourceAttributes: TaskResourceAttributesConfig{
				Values: []TaskResourceAttributes{},
			},
		}

		err := SeedMatchableResources(ctx, gormDb, config)
		assert.NoError(t, err)

		var modelsInDB []models.Resource
		gormDb.Where(models.Resource{}).Find(&modelsInDB)
		assert.Len(t, modelsInDB, 0)
	})

	t.Run("Seed empty db", func(t *testing.T) {
		gormDb := GetDbForResourcesTest(t)

		config := MatcheableResourcesConfig{
			ExecutionClusterLabels: ExecutionClusterLabelConfig{
				Values: []ExecutionClusterLabel{
					{
						Project:  "project",
						Domain:   "domain",
						Workflow: "workflow",
						Label:    "label",
					},
				},
			},
			TaskResourceAttributes: TaskResourceAttributesConfig{
				Values: []TaskResourceAttributes{
					{
						Project: "project2",
						Defaults: TaskResourceSpec{
							Cpu: "1",
						},
						Limits: TaskResourceSpec{
							Cpu: "2",
						},
					},
				},
			},
		}

		err := SeedMatchableResources(ctx, gormDb, config)
		assert.NoError(t, err)

		var modelsInDB []models.Resource
		gormDb.Where(models.Resource{}).Find(&modelsInDB)

		assert.Len(t, modelsInDB, 2)
		assert.Equal(t, "project", modelsInDB[0].Project)
		assert.Equal(t, "domain", modelsInDB[0].Domain)
		assert.Equal(t, "workflow", modelsInDB[0].Workflow)

		labelInDB := admin.MatchingAttributes{
			Target: &admin.MatchingAttributes_ExecutionClusterLabel{
				ExecutionClusterLabel: &admin.ExecutionClusterLabel{},
			},
		}
		err = proto.Unmarshal(modelsInDB[0].Attributes, &labelInDB)
		assert.NoError(t, err)
		assert.Equal(t, "label", labelInDB.GetExecutionClusterLabel().Value)

		assert.Equal(t, "project2", modelsInDB[1].Project)
		assert.Equal(t, "", modelsInDB[1].Domain)
		assert.Equal(t, "", modelsInDB[1].Workflow)

		taskResourceInDB := admin.MatchingAttributes{
			Target: &admin.MatchingAttributes_TaskResourceAttributes{
				TaskResourceAttributes: &admin.TaskResourceAttributes{},
			},
		}
		err = proto.Unmarshal(modelsInDB[1].Attributes, &taskResourceInDB)
		assert.NoError(t, err)
		assert.Equal(t, "1", taskResourceInDB.GetTaskResourceAttributes().GetDefaults().GetCpu())
	})

	t.Run("Test imperative", func(t *testing.T) {
		gormDb := GetDbForResourcesTest(t)

		config := MatcheableResourcesConfig{
			ExecutionClusterLabels: ExecutionClusterLabelConfig{
				Values: []ExecutionClusterLabel{
					{
						Project: "project",
						Label:   "label",
					},
				},
			},
			TaskResourceAttributes: TaskResourceAttributesConfig{
				Values: []TaskResourceAttributes{
					{
						Project: "project",
						Defaults: TaskResourceSpec{
							Cpu: "1",
						},
					},
				},
			},
		}

		err := SeedMatchableResources(ctx, gormDb, config)
		assert.NoError(t, err)

		var modelsInDB []models.Resource
		gormDb.Where(models.Resource{}).Find(&modelsInDB)

		assert.Len(t, modelsInDB, 2)

		newConfig := MatcheableResourcesConfig{
			ExecutionClusterLabels: ExecutionClusterLabelConfig{
				Values: []ExecutionClusterLabel{
					{
						Project: "project",
						Label:   "labelAfterUpdate", // Update the value
					},
				},
			},
			TaskResourceAttributes: TaskResourceAttributesConfig{
				Values: []TaskResourceAttributes{
					{
						Project: "newProject", // Add resource for a new project
						Defaults: TaskResourceSpec{
							Cpu: "1",
						},
					},
				},
			},
		}

		err = SeedMatchableResources(ctx, gormDb, newConfig)
		assert.NoError(t, err)

		var modelsInDBAfterUpdate []models.Resource
		gormDb.Where(models.Resource{}).Find(&modelsInDBAfterUpdate)

		// One existing resource was updated, one was added
		assert.Len(t, modelsInDBAfterUpdate, 3)

	})

	t.Run("Test declarative", func(t *testing.T) {
		gormDb := GetDbForResourcesTest(t)

		config := MatcheableResourcesConfig{
			ExecutionClusterLabels: ExecutionClusterLabelConfig{
				Values: []ExecutionClusterLabel{
					{
						Project: "project",
						Label:   "label",
					},
				},
			},
			TaskResourceAttributes: TaskResourceAttributesConfig{
				Values: []TaskResourceAttributes{
					{
						Project: "project",
						Defaults: TaskResourceSpec{
							Cpu: "1",
						},
					},
				},
			},
		}

		err := SeedMatchableResources(ctx, gormDb, config)
		assert.NoError(t, err)

		var modelsInDB []models.Resource
		gormDb.Where(models.Resource{}).Find(&modelsInDB)

		assert.Len(t, modelsInDB, 2)

		newConfig := MatcheableResourcesConfig{
			ExecutionClusterLabels: ExecutionClusterLabelConfig{
				Declarative: true, // Causes all resources which are not part of the new config to be deleted
				Values: []ExecutionClusterLabel{
					{
						Project: "newProject",
						Label:   "label",
					},
				},
			},
			TaskResourceAttributes: TaskResourceAttributesConfig{
				Declarative: true,
				Values: []TaskResourceAttributes{
					{
						Project: "newProject", // Add resource for a new project
						Defaults: TaskResourceSpec{
							Cpu: "1",
						},
					},
				},
			},
		}

		err = SeedMatchableResources(ctx, gormDb, newConfig)
		assert.NoError(t, err)

		var modelsInDBAfterUpdate []models.Resource
		gormDb.Where(models.Resource{}).Find(&modelsInDBAfterUpdate)

		assert.Len(t, modelsInDBAfterUpdate, 2)
		assert.Equal(t, "newProject", modelsInDBAfterUpdate[0].Project)
		assert.Equal(t, "newProject", modelsInDBAfterUpdate[1].Project)
	})
}

func TestCreateAttributeModel(t *testing.T) {
	t.Run("Test missing project", func(t *testing.T) {
		matchingAttributes := &admin.MatchingAttributes{
			Target: &admin.MatchingAttributes_ExecutionClusterLabel{
				ExecutionClusterLabel: &admin.ExecutionClusterLabel{},
			},
		}
		_, err := CreateAttributeModel("", "", "workflow", matchingAttributes, admin.MatchableResource_EXECUTION_CLUSTER_LABEL)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "project is required")
	})
	t.Run("Test missing domain", func(t *testing.T) {

		matchingAttributes := &admin.MatchingAttributes{
			Target: &admin.MatchingAttributes_ExecutionClusterLabel{
				ExecutionClusterLabel: &admin.ExecutionClusterLabel{},
			},
		}
		_, err := CreateAttributeModel("project", "", "workflow", matchingAttributes, admin.MatchableResource_EXECUTION_CLUSTER_LABEL)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "domain is required")
	})
	t.Run("Test project only", func(t *testing.T) {
		matchingAttributes := &admin.MatchingAttributes{
			Target: &admin.MatchingAttributes_ExecutionClusterLabel{
				ExecutionClusterLabel: &admin.ExecutionClusterLabel{},
			},
		}
		model, err := CreateAttributeModel("project", "", "", matchingAttributes, admin.MatchableResource_EXECUTION_CLUSTER_LABEL)
		assert.NoError(t, err)
		assert.Equal(t, "project", model.Project)
		assert.Equal(t, "", model.Domain)
		assert.Equal(t, "", model.Workflow)
		assert.Equal(t, models.ResourcePriorityProjectLevel, model.Priority)
	})
	t.Run("Test project and domain", func(t *testing.T) {
		matchingAttributes := &admin.MatchingAttributes{
			Target: &admin.MatchingAttributes_ExecutionClusterLabel{
				ExecutionClusterLabel: &admin.ExecutionClusterLabel{},
			},
		}
		model, err := CreateAttributeModel("project", "domain", "", matchingAttributes, admin.MatchableResource_EXECUTION_CLUSTER_LABEL)
		assert.NoError(t, err)
		assert.Equal(t, "project", model.Project)
		assert.Equal(t, "domain", model.Domain)
		assert.Equal(t, "", model.Workflow)
		assert.Equal(t, models.ResourcePriorityProjectDomainLevel, model.Priority)
	})
	t.Run("Test project, domain, and workflow", func(t *testing.T) {
		matchingAttributes := &admin.MatchingAttributes{
			Target: &admin.MatchingAttributes_ExecutionClusterLabel{
				ExecutionClusterLabel: &admin.ExecutionClusterLabel{},
			},
		}
		model, err := CreateAttributeModel("project", "domain", "workflow", matchingAttributes, admin.MatchableResource_EXECUTION_CLUSTER_LABEL)
		assert.NoError(t, err)
		assert.Equal(t, "project", model.Project)
		assert.Equal(t, "domain", model.Domain)
		assert.Equal(t, "workflow", model.Workflow)
		assert.Equal(t, models.ResourcePriorityWorkflowLevel, model.Priority)
	})
}

func TestCreateConfiguredResourceModelsMap(t *testing.T) {
	t.Run("Test valid config", func(t *testing.T) {
		executionClusterLabels := []ExecutionClusterLabel{
			{
				Project: "project",
				Label:   "label",
			},
		}

		modelsMap, err := CreateConfiguredResourceModelsMap(executionClusterLabels, ExecutionClusterLabelProcessor, admin.MatchableResource_EXECUTION_CLUSTER_LABEL)
		assert.NoError(t, err)
		assert.Len(t, modelsMap, 1)
	})
	t.Run("Test duplicate config", func(t *testing.T) {
		executionClusterLabels := []ExecutionClusterLabel{
			{
				Project: "project",
				Label:   "label",
			},
			{
				Project: "project",
				Label:   "label",
			},
		}
		_, err := CreateConfiguredResourceModelsMap(executionClusterLabels, ExecutionClusterLabelProcessor, admin.MatchableResource_EXECUTION_CLUSTER_LABEL)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "duplicate resource of type")
	})
}

func TestExecutionClusterLabelProcessor(t *testing.T) {
	executionClusterLabel := ExecutionClusterLabel{
		Project: "project",
		Domain:  "domain",
		Label:   "label",
	}

	model, err := ExecutionClusterLabelProcessor(executionClusterLabel, admin.MatchableResource_EXECUTION_CLUSTER_LABEL)
	assert.NoError(t, err)
	assert.Equal(t, "project", model.Project)
	assert.Equal(t, "domain", model.Domain)
	assert.Equal(t, "", model.Workflow)
	assert.Equal(t, models.ResourcePriorityProjectDomainLevel, model.Priority)

	matchingAttributes := admin.MatchingAttributes{}
	err = proto.Unmarshal(model.Attributes, &matchingAttributes)
	assert.NoError(t, err)
	assert.Equal(t, "label", matchingAttributes.GetExecutionClusterLabel().Value)
}

func TestTaskResourceAttributesProcessor(t *testing.T) {
	taskResourceAttributes := TaskResourceAttributes{
		Project:  "project",
		Domain:   "domain",
		Workflow: "workflow",
		Defaults: TaskResourceSpec{
			Cpu: "1",
		},
		Limits: TaskResourceSpec{
			Cpu: "2",
		},
	}

	model, err := TaskResourceAttributesProcessor(taskResourceAttributes, admin.MatchableResource_TASK_RESOURCE)
	assert.NoError(t, err)
	assert.Equal(t, "project", model.Project)
	assert.Equal(t, "domain", model.Domain)
	assert.Equal(t, "workflow", model.Workflow)
	assert.Equal(t, models.ResourcePriorityWorkflowLevel, model.Priority)

	matchingAttributes := admin.MatchingAttributes{}
	err = proto.Unmarshal(model.Attributes, &matchingAttributes)
	assert.NoError(t, err)
	assert.Equal(t, "1", matchingAttributes.GetTaskResourceAttributes().GetDefaults().GetCpu())
	assert.Equal(t, "2", matchingAttributes.GetTaskResourceAttributes().GetLimits().GetCpu())
}

func TestClusterResourceAttributesProcessor(t *testing.T) {
	clusterResourceAttributes := ClusterResourceAttributes{
		Project: "project",
		Attributes: map[string]string{
			"foo": "bar",
		},
	}

	model, err := ClusterResourceAttributesProcessor(clusterResourceAttributes, admin.MatchableResource_CLUSTER_RESOURCE)
	assert.NoError(t, err)
	assert.Equal(t, "project", model.Project)
	assert.Equal(t, "", model.Domain)
	assert.Equal(t, "", model.Workflow)
	assert.Equal(t, models.ResourcePriorityProjectLevel, model.Priority)

	matchingAttributes := admin.MatchingAttributes{}
	err = proto.Unmarshal(model.Attributes, &matchingAttributes)
	assert.NoError(t, err)
	assert.Equal(t, "bar", matchingAttributes.GetClusterResourceAttributes().GetAttributes()["foo"])
}

func TestExecutionQueueAttributesProcessor(t *testing.T) {
	executionQueueAttributes := ExecutionQueueAttributes{
		Project:  "project",
		Domain:   "domain",
		Workflow: "workflow",
		Tags:     []string{"tag1", "tag2"},
	}

	model, err := ExecutionQueueAttributesProcessor(executionQueueAttributes, admin.MatchableResource_EXECUTION_QUEUE)
	assert.NoError(t, err)
	assert.Equal(t, "project", model.Project)
	assert.Equal(t, "domain", model.Domain)
	assert.Equal(t, "workflow", model.Workflow)
	assert.Equal(t, models.ResourcePriorityWorkflowLevel, model.Priority)

	matchingAttributes := admin.MatchingAttributes{}
	err = proto.Unmarshal(model.Attributes, &matchingAttributes)
	assert.NoError(t, err)
	assert.Equal(t, []string{"tag1", "tag2"}, matchingAttributes.GetExecutionQueueAttributes().GetTags())
}

func TestPluginOverridesProcessor(t *testing.T) {
	pluginOverrides := PluginOverrides{
		Project:  "project",
		Domain:   "domain",
		Workflow: "workflow",
		Overrides: []PluginOverride{
			{
				TaskType:              "taskType",
				PluginId:              []string{"pluginId"},
				MissingPluginBehavior: 0,
			},
		},
	}

	model, err := PluginOverridesProcessor(pluginOverrides, admin.MatchableResource_PLUGIN_OVERRIDE)
	assert.NoError(t, err)
	assert.Equal(t, "project", model.Project)
	assert.Equal(t, "domain", model.Domain)
	assert.Equal(t, "workflow", model.Workflow)
	assert.Equal(t, models.ResourcePriorityWorkflowLevel, model.Priority)

	matchingAttributes := admin.MatchingAttributes{}
	err = proto.Unmarshal(model.Attributes, &matchingAttributes)
	assert.NoError(t, err)
	assert.Equal(t, "taskType", matchingAttributes.GetPluginOverrides().GetOverrides()[0].GetTaskType())
	assert.Equal(t, []string{"pluginId"}, matchingAttributes.GetPluginOverrides().GetOverrides()[0].GetPluginId())
	assert.Equal(t, 0, int(matchingAttributes.GetPluginOverrides().GetOverrides()[0].GetMissingPluginBehavior()))
}

func TestWorkflowExecutionConfigProcessor(t *testing.T) {
	workflowExecutionConfig := WorkflowExecutionConfig{
		Project:        "project",
		Domain:         "domain",
		MaxParallelism: 10,
		SecurityContext: SecurityContext{
			RunAs: Identity{
				K8sServiceAccount: "serviceAccount",
			},
		},
	}

	model, err := WorkflowExecutionConfigProcessor(workflowExecutionConfig, admin.MatchableResource_WORKFLOW_EXECUTION_CONFIG)
	assert.NoError(t, err)
	assert.Equal(t, "project", model.Project)
	assert.Equal(t, "domain", model.Domain)
	assert.Equal(t, "", model.Workflow)
	assert.Equal(t, models.ResourcePriorityProjectDomainLevel, model.Priority)

	matchingAttributes := admin.MatchingAttributes{}
	err = proto.Unmarshal(model.Attributes, &matchingAttributes)
	assert.NoError(t, err)
	assert.Equal(t, int32(10), matchingAttributes.GetWorkflowExecutionConfig().GetMaxParallelism())
	assert.Equal(t, "serviceAccount", matchingAttributes.GetWorkflowExecutionConfig().GetSecurityContext().GetRunAs().K8SServiceAccount)
}
