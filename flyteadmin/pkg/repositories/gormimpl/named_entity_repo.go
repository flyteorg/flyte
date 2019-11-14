package gormimpl

import (
	"context"
	"fmt"

	"github.com/lyft/flyteidl/gen/pb-go/flyteidl/core"
	"google.golang.org/grpc/codes"

	"github.com/jinzhu/gorm"
	"github.com/lyft/flyteadmin/pkg/common"
	adminErrors "github.com/lyft/flyteadmin/pkg/errors"
	"github.com/lyft/flyteadmin/pkg/repositories/errors"
	"github.com/lyft/flyteadmin/pkg/repositories/interfaces"
	"github.com/lyft/flyteadmin/pkg/repositories/models"
	"github.com/lyft/flytestdlib/promutils"
)

var leftJoinWorkflowNameToMetadata = fmt.Sprintf(
	"LEFT JOIN %s ON %s.resource_type = %d AND %s.project = %s.project AND %s.domain = %s.domain AND %s.name = %s.name", namedEntityMetadataTableName, namedEntityMetadataTableName, core.ResourceType_WORKFLOW, namedEntityMetadataTableName, workflowTableName,
	namedEntityMetadataTableName, workflowTableName,
	namedEntityMetadataTableName, workflowTableName)

var leftJoinLaunchPlanNameToMetadata = fmt.Sprintf(
	"LEFT JOIN %s ON %s.resource_type = %d AND %s.project = %s.project AND %s.domain = %s.domain AND %s.name = %s.name", namedEntityMetadataTableName, namedEntityMetadataTableName, core.ResourceType_LAUNCH_PLAN, namedEntityMetadataTableName, launchPlanTableName,
	namedEntityMetadataTableName, launchPlanTableName,
	namedEntityMetadataTableName, launchPlanTableName)

var leftJoinTaskNameToMetadata = fmt.Sprintf(
	"LEFT JOIN %s ON %s.resource_type = %d AND %s.project = %s.project AND %s.domain = %s.domain AND %s.name = %s.name", namedEntityMetadataTableName, namedEntityMetadataTableName, core.ResourceType_TASK, namedEntityMetadataTableName, taskTableName,
	namedEntityMetadataTableName, taskTableName,
	namedEntityMetadataTableName, taskTableName)

var resourceTypeToTableName = map[core.ResourceType]string{
	core.ResourceType_LAUNCH_PLAN: launchPlanTableName,
	core.ResourceType_WORKFLOW:    workflowTableName,
	core.ResourceType_TASK:        taskTableName,
}

var resourceTypeToMetadataJoin = map[core.ResourceType]string{
	core.ResourceType_LAUNCH_PLAN: leftJoinLaunchPlanNameToMetadata,
	core.ResourceType_WORKFLOW:    leftJoinWorkflowNameToMetadata,
	core.ResourceType_TASK:        leftJoinTaskNameToMetadata,
}

func getGroupByForNamedEntity(tableName string) string {
	return fmt.Sprintf("%s.%s, %s.%s, %s.%s, %s.%s", tableName, Project, tableName, Domain, tableName, Name, namedEntityMetadataTableName, Description)
}

func getSelectForNamedEntity(tableName string, resourceType core.ResourceType) []string {
	return []string{
		fmt.Sprintf("%s.%s", tableName, Project),
		fmt.Sprintf("%s.%s", tableName, Domain),
		fmt.Sprintf("%s.%s", tableName, Name),
		fmt.Sprintf("'%d' AS %s", resourceType, ResourceType),
		fmt.Sprintf("%s.%s", namedEntityMetadataTableName, Description),
	}
}

func getNamedEntityFilters(resourceType core.ResourceType, project string, domain string, name string) ([]common.InlineFilter, error) {
	entity := common.ResourceTypeToEntity[resourceType]

	filters := make([]common.InlineFilter, 0)
	projectFilter, err := common.NewSingleValueFilter(entity, common.Equal, Project, project)
	if err != nil {
		return nil, err
	}
	filters = append(filters, projectFilter)
	domainFilter, err := common.NewSingleValueFilter(entity, common.Equal, Domain, domain)
	if err != nil {
		return nil, err
	}
	filters = append(filters, domainFilter)
	nameFilter, err := common.NewSingleValueFilter(entity, common.Equal, Name, name)
	if err != nil {
		return nil, err
	}
	filters = append(filters, nameFilter)

	return filters, nil
}

// Implementation of NamedEntityRepoInterface.
type NamedEntityRepo struct {
	db               *gorm.DB
	errorTransformer errors.ErrorTransformer
	metrics          gormMetrics
}

func (r *NamedEntityRepo) Update(ctx context.Context, input models.NamedEntity) error {
	timer := r.metrics.UpdateDuration.Start()
	var metadata models.NamedEntityMetadata
	tx := r.db.Where(&models.NamedEntityMetadata{
		NamedEntityMetadataKey: models.NamedEntityMetadataKey{
			ResourceType: input.ResourceType,
			Project:      input.Project,
			Domain:       input.Domain,
			Name:         input.Name,
		},
	}).Assign(input.NamedEntityMetadataFields).FirstOrCreate(&metadata)
	timer.Stop()
	if tx.Error != nil {
		return r.errorTransformer.ToFlyteAdminError(tx.Error)
	}
	return nil
}

func (r *NamedEntityRepo) Get(ctx context.Context, input interfaces.GetNamedEntityInput) (models.NamedEntity, error) {
	var namedEntity models.NamedEntity

	filters, err := getNamedEntityFilters(input.ResourceType, input.Project, input.Domain, input.Name)
	if err != nil {
		return models.NamedEntity{}, err
	}

	tableName, tableFound := resourceTypeToTableName[input.ResourceType]
	joinString, joinFound := resourceTypeToMetadataJoin[input.ResourceType]
	if !tableFound || !joinFound {
		return models.NamedEntity{}, adminErrors.NewFlyteAdminErrorf(codes.InvalidArgument, "Cannot get NamedEntity for resource type: %v", input.ResourceType)
	}

	tx := r.db.Table(tableName).Joins(joinString)

	// Apply filters
	tx, err = applyScopedFilters(tx, filters, nil)
	if err != nil {
		return models.NamedEntity{}, err
	}

	timer := r.metrics.GetDuration.Start()
	tx = tx.Select(getSelectForNamedEntity(tableName, input.ResourceType)).First(&namedEntity)
	timer.Stop()

	if tx.Error != nil {
		return models.NamedEntity{}, r.errorTransformer.ToFlyteAdminError(tx.Error)
	}

	return namedEntity, nil
}

func (r *NamedEntityRepo) List(ctx context.Context, resourceType core.ResourceType, input interfaces.ListResourceInput) (
	interfaces.NamedEntityCollectionOutput, error) {

	// Validate input.
	if err := ValidateListInput(input); err != nil {
		return interfaces.NamedEntityCollectionOutput{}, err
	}

	tableName, tableFound := resourceTypeToTableName[resourceType]
	joinString, joinFound := resourceTypeToMetadataJoin[resourceType]
	if !tableFound || !joinFound {
		return interfaces.NamedEntityCollectionOutput{}, adminErrors.NewFlyteAdminErrorf(codes.InvalidArgument, "Cannot list entity names for resource type: %v", resourceType)
	}

	tx := r.db.Table(tableName).Limit(input.Limit).Offset(input.Offset)
	tx = tx.Joins(joinString)

	// Apply filters
	tx, err := applyScopedFilters(tx, input.InlineFilters, input.MapFilters)
	if err != nil {
		return interfaces.NamedEntityCollectionOutput{}, err
	}

	// Apply sort ordering.
	if input.SortParameter != nil {
		tx = tx.Order(input.SortParameter.GetGormOrderExpr())
	}

	// Scan the results into a list of named entities
	var entities []models.NamedEntity
	timer := r.metrics.ListDuration.Start()
	tx.Select(getSelectForNamedEntity(tableName, resourceType)).Group(getGroupByForNamedEntity(tableName)).Scan(&entities)
	timer.Stop()
	if tx.Error != nil {
		return interfaces.NamedEntityCollectionOutput{}, r.errorTransformer.ToFlyteAdminError(tx.Error)
	}

	return interfaces.NamedEntityCollectionOutput{
		Entities: entities,
	}, nil
}

// Returns an instance of NamedEntityRepoInterface
func NewNamedEntityRepo(
	db *gorm.DB, errorTransformer errors.ErrorTransformer, scope promutils.Scope) interfaces.NamedEntityRepoInterface {
	metrics := newMetrics(scope)

	return &NamedEntityRepo{
		db:               db,
		errorTransformer: errorTransformer,
		metrics:          metrics,
	}
}
