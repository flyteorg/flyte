package config

import (
	"context"
	"fmt"

	"gorm.io/gorm"

	"github.com/flyteorg/flyte/flyteadmin/pkg/repositories/models"
	"github.com/flyteorg/flyte/flyteadmin/pkg/repositories/transformers"
	"github.com/flyteorg/flyte/flyteidl/gen/pb-go/flyteidl/admin"
	"github.com/flyteorg/flyte/flyteidl/gen/pb-go/flyteidl/core"
	"github.com/flyteorg/flyte/flytestdlib/logger"
)

// Returns a function to seed the database with default values.
func SeedProjects(db *gorm.DB, projects []string) error {
	tx := db.Begin()
	for _, project := range projects {
		projectModel := models.Project{
			Identifier:  project,
			Name:        project,
			Description: fmt.Sprintf("%s description", project),
		}
		if err := tx.Where(models.Project{Identifier: project}).Omit("id").FirstOrCreate(&projectModel).Error; err != nil {
			logger.Warningf(context.Background(), "failed to save project [%s]", project)
			tx.Rollback()
			return err
		}
	}
	return tx.Commit().Error
}

func ValidateMatcheableResourceLevel(project string, domain string, workflow string) error {
	if project == "" {
		return fmt.Errorf("project is required")
	}
	if domain == "" && workflow != "" {
		return fmt.Errorf("domain is required: %s %s", project, workflow)
	}
	return nil
}

func CreateAttributeModel(project string, domain string, workflow string, attribute *admin.MatchingAttributes, resourceType admin.MatchableResource) (models.Resource, error) {
	err := ValidateMatcheableResourceLevel(project, domain, workflow)
	if err != nil {
		return models.Resource{}, err
	}

	var model models.Resource

	if workflow != "" {
		model, err = transformers.WorkflowAttributesToResourceModel(
			admin.WorkflowAttributes{
				Project:            project,
				Domain:             domain,
				Workflow:           workflow,
				MatchingAttributes: attribute,
			},
			resourceType,
		)

	} else if domain != "" {
		model, err = transformers.ProjectDomainAttributesToResourceModel(
			admin.ProjectDomainAttributes{
				Project:            project,
				Domain:             domain,
				MatchingAttributes: attribute,
			},
			resourceType,
		)
	} else {
		model, err = transformers.ProjectAttributesToResourceModel(
			admin.ProjectAttributes{
				Project:            project,
				MatchingAttributes: attribute,
			},
			resourceType,
		)

	}
	return model, err
}

func CreateKey(project string, domain string, workflow string, resourceType string) string {
	return fmt.Sprintf("%s-%s-%s-%s", project, domain, workflow, resourceType)
}

func RetrieveExistingResourceModelsMap(db *gorm.DB, resourceType admin.MatchableResource) (map[string]models.Resource, error) {
	var modelsInDB []models.Resource
	db.Where(models.Resource{ResourceType: resourceType.String()}).Find(&modelsInDB)

	dbMap := make(map[string]models.Resource)
	for _, model := range modelsInDB {
		key := CreateKey(model.Project, model.Domain, model.Workflow, model.ResourceType)
		dbMap[key] = model
	}
	return dbMap, nil
}

func CreateWhereQuery(model *models.Resource) map[string]interface{} {
	whereMap := make(map[string]interface{})

	whereMap["project"] = model.Project

	if model.Domain != "" {
		whereMap["domain"] = model.Domain
	}
	if model.Workflow != "" {
		whereMap["workflow"] = model.Workflow
	}

	whereMap["resource_type"] = model.ResourceType

	return whereMap
}

func CreateOrUpdateResourceModel(db *gorm.DB, configuredModels map[string]models.Resource, existingModels map[string]models.Resource) error {
	for key, model := range configuredModels {
		if _, exists := existingModels[key]; exists {
			if err := db.Model(&models.Resource{}).Where(CreateWhereQuery(&model)).Updates(models.Resource{
				Attributes: model.Attributes,
			}).Error; err != nil {
				logger.Warningf(context.Background(), "failed to update matcheable resource [%s]", model)
				return err
			}
		} else {
			if err := db.Create(&model).Error; err != nil {
				logger.Warningf(context.Background(), "failed to create matcheable resource [%s]", model)
				return err
			}
		}
	}
	return nil
}

func DeleteNonConfiguredResourceModels(db *gorm.DB, configuredModels map[string]models.Resource, existingModels map[string]models.Resource) error {
	for key, model := range existingModels {
		if _, exists := configuredModels[key]; !exists {
			if err := db.Delete(&model).Error; err != nil {
				logger.Warningf(context.Background(), "failed to delete matcheable resource [%s]", model)
				return err
			}
		}
	}
	return nil
}

func SeedMatchableResourceModels(db *gorm.DB, configuredModels map[string]models.Resource, existingModels map[string]models.Resource, declarative bool) error {
	tx := db.Begin()

	if err := CreateOrUpdateResourceModel(tx, configuredModels, existingModels); err != nil {
		db.Rollback()
		return err
	}

	if declarative {
		if err := DeleteNonConfiguredResourceModels(tx, configuredModels, existingModels); err != nil {
			db.Rollback()
			return err
		}
	}

	return tx.Commit().Error
}

type ResourceProcessor[T any] func(item T, resourceType admin.MatchableResource) (models.Resource, error)

func CreateConfiguredResourceModelsMap[T any](items []T, processor ResourceProcessor[T], resourceType admin.MatchableResource) (map[string]models.Resource, error) {
	configMap := make(map[string]models.Resource)
	for _, item := range items {
		model, err := processor(item, resourceType)
		if err != nil {
			return nil, err
		}
		key := CreateKey(model.Project, model.Domain, model.Workflow, model.ResourceType)
		if _, exists := configMap[key]; exists {
			return nil, fmt.Errorf("duplicate resource of type %s for %s %s %s", resourceType, model.Project, model.Domain, model.Workflow)
		}
		configMap[key] = model
	}
	return configMap, nil
}

func SeedMatchableResourcesSingleType[T any](db *gorm.DB, items []T, processor ResourceProcessor[T], resourceType admin.MatchableResource, declarative bool) error {
	configuredModels, err := CreateConfiguredResourceModelsMap(items, processor, resourceType)
	if err != nil {
		return fmt.Errorf("invalid configuration: %v", err)
	}
	existingModels, err := RetrieveExistingResourceModelsMap(db, resourceType)
	if err != nil {
		return fmt.Errorf("failed to retrieve existing resources: %v", err)
	}

	return SeedMatchableResourceModels(db, configuredModels, existingModels, declarative)
}

func ExecutionClusterLabelProcessor(item ExecutionClusterLabel, resourceType admin.MatchableResource) (models.Resource, error) {
	// Directly use item as ExecutionClusterLabel, no need for type assertion
	model, err := CreateAttributeModel(
		item.Project,
		item.Domain,
		item.Workflow,
		&admin.MatchingAttributes{
			Target: &admin.MatchingAttributes_ExecutionClusterLabel{
				ExecutionClusterLabel: &admin.ExecutionClusterLabel{
					Value: item.Label,
				},
			},
		},
		resourceType,
	)
	if err != nil {
		return models.Resource{}, err
	}
	return model, nil
}

func TaskResourceAttributesProcessor(item TaskResourceAttributes, resourceType admin.MatchableResource) (models.Resource, error) {
	model, err := CreateAttributeModel(
		item.Project,
		item.Domain,
		item.Workflow,
		&admin.MatchingAttributes{
			Target: &admin.MatchingAttributes_TaskResourceAttributes{
				TaskResourceAttributes: &admin.TaskResourceAttributes{
					Defaults: &admin.TaskResourceSpec{
						Cpu:              item.Defaults.Cpu,
						Gpu:              item.Defaults.Gpu,
						Memory:           item.Defaults.Memory,
						Storage:          item.Defaults.Storage,
						EphemeralStorage: item.Defaults.EphemeralStorage,
					},
					Limits: &admin.TaskResourceSpec{
						Cpu:              item.Limits.Cpu,
						Gpu:              item.Limits.Gpu,
						Memory:           item.Limits.Memory,
						Storage:          item.Limits.Storage,
						EphemeralStorage: item.Limits.EphemeralStorage,
					},
				},
			},
		},
		resourceType,
	)
	if err != nil {
		return models.Resource{}, err
	}
	return model, nil
}

func ClusterResourceAttributesProcessor(item ClusterResourceAttributes, resourceType admin.MatchableResource) (models.Resource, error) {
	attributes := make(map[string]string)
	for key, value := range item.Attributes {
		attributes[key] = value
	}

	model, err := CreateAttributeModel(
		item.Project,
		item.Domain,
		item.Workflow,
		&admin.MatchingAttributes{
			Target: &admin.MatchingAttributes_ClusterResourceAttributes{
				ClusterResourceAttributes: &admin.ClusterResourceAttributes{
					Attributes: attributes,
				},
			},
		},
		resourceType,
	)
	if err != nil {
		return models.Resource{}, err
	}
	return model, nil
}

func ExecutionQueueAttributesProcessor(item ExecutionQueueAttributes, resourceType admin.MatchableResource) (models.Resource, error) {
	model, err := CreateAttributeModel(
		item.Project,
		item.Domain,
		item.Workflow,
		&admin.MatchingAttributes{
			Target: &admin.MatchingAttributes_ExecutionQueueAttributes{
				ExecutionQueueAttributes: &admin.ExecutionQueueAttributes{
					Tags: item.Tags,
				},
			},
		},
		resourceType,
	)
	if err != nil {
		return models.Resource{}, err
	}
	return model, nil
}

func PluginOverridesProcessor(item PluginOverrides, resourceType admin.MatchableResource) (models.Resource, error) {
	overrides := make([]*admin.PluginOverride, 0)
	for _, override := range item.Overrides {
		overrides = append(overrides, &admin.PluginOverride{
			PluginId:              override.PluginId,
			TaskType:              override.TaskType,
			MissingPluginBehavior: admin.PluginOverride_MissingPluginBehavior(override.MissingPluginBehavior),
		})
	}

	model, err := CreateAttributeModel(
		item.Project,
		item.Domain,
		item.Workflow,
		&admin.MatchingAttributes{
			Target: &admin.MatchingAttributes_PluginOverrides{
				PluginOverrides: &admin.PluginOverrides{
					Overrides: overrides,
				},
			},
		},
		resourceType,
	)
	if err != nil {
		return models.Resource{}, err
	}
	return model, nil
}

func WorkflowExecutionConfigProcessor(item WorkflowExecutionConfig, resourceType admin.MatchableResource) (models.Resource, error) {
	model, err := CreateAttributeModel(
		item.Project,
		item.Domain,
		item.Workflow,
		&admin.MatchingAttributes{
			Target: &admin.MatchingAttributes_WorkflowExecutionConfig{
				WorkflowExecutionConfig: &admin.WorkflowExecutionConfig{
					MaxParallelism: item.MaxParallelism,
					SecurityContext: &core.SecurityContext{
						RunAs: &core.Identity{
							K8SServiceAccount: item.SecurityContext.RunAs.K8sServiceAccount,
						},
					},
				},
			},
		},
		resourceType,
	)
	if err != nil {
		return models.Resource{}, err
	}
	return model, nil
}

func SeedMatchableResources(ctx context.Context, db *gorm.DB, config MatcheableResourcesConfig) error {

	if err := SeedMatchableResourcesSingleType(db, config.ExecutionClusterLabels.Values, ExecutionClusterLabelProcessor, admin.MatchableResource_EXECUTION_CLUSTER_LABEL, config.ExecutionClusterLabels.Declarative); err != nil {

		return fmt.Errorf("could not add execution cluster labels to database with err: %v", err)
	}
	logger.Infof(ctx, "Successfully added execution cluster labels to database")

	if err := SeedMatchableResourcesSingleType(db, config.TaskResourceAttributes.Values, TaskResourceAttributesProcessor, admin.MatchableResource_TASK_RESOURCE, config.TaskResourceAttributes.Declarative); err != nil {
		return fmt.Errorf("could not add task resource attributes to database with err: %v", err)
	}

	if err := SeedMatchableResourcesSingleType(db, config.ClusterResourceAttributes.Values, ClusterResourceAttributesProcessor, admin.MatchableResource_CLUSTER_RESOURCE, config.ClusterResourceAttributes.Declarative); err != nil {
		return fmt.Errorf("could not add cluster resource attributes to database with err: %v", err)
	}

	if err := SeedMatchableResourcesSingleType(db, config.ExecutionQueueAttributes.Values, ExecutionQueueAttributesProcessor, admin.MatchableResource_EXECUTION_QUEUE, config.ExecutionQueueAttributes.Declarative); err != nil {
		return fmt.Errorf("could not add execution queue attributes to database with err: %v", err)
	}
	if err := SeedMatchableResourcesSingleType(db, config.PluginOverrides.Values, PluginOverridesProcessor, admin.MatchableResource_PLUGIN_OVERRIDE, config.PluginOverrides.Declarative); err != nil {
		return fmt.Errorf("could not add plugin overrides to database with err: %v", err)
	}

	if err := SeedMatchableResourcesSingleType(db, config.WorkflowExecutionConfigs.Values, WorkflowExecutionConfigProcessor, admin.MatchableResource_WORKFLOW_EXECUTION_CONFIG, config.WorkflowExecutionConfigs.Declarative); err != nil {
		return fmt.Errorf("could not add workflow execution configs to database with err: %v", err)
	}
	return nil
}
