package gormimpl

import (
	"context"
	"fmt"

	"github.com/flyteorg/flyteidl/gen/pb-go/flyteidl/core"
	"github.com/flyteorg/flytestdlib/promutils"
	"google.golang.org/grpc/codes"
	"gorm.io/gorm"

	"github.com/flyteorg/flyteadmin/pkg/common"
	adminErrors "github.com/flyteorg/flyteadmin/pkg/errors"
	"github.com/flyteorg/flyteadmin/pkg/repositories/errors"
	"github.com/flyteorg/flyteadmin/pkg/repositories/interfaces"
	"github.com/flyteorg/flyteadmin/pkg/repositories/models"
)

const innerJoinTableAlias = "entities"

var resourceTypeToTableName = map[core.ResourceType]string{
	core.ResourceType_LAUNCH_PLAN: launchPlanTableName,
	core.ResourceType_WORKFLOW:    workflowTableName,
	core.ResourceType_TASK:        taskTableName,
}

var joinString = "RIGHT JOIN (?) AS entities ON named_entity_metadata.resource_type = %d AND " +
	"named_entity_metadata.project = entities.project AND named_entity_metadata.domain = entities.domain AND " +
	"named_entity_metadata.name = entities.name"

func getSubQueryJoin(db *gorm.DB, tableName string, input interfaces.ListNamedEntityInput) *gorm.DB {
	tx := db.Select([]string{Project, Domain, Name}).
		Table(tableName).
		Where(map[string]interface{}{Project: input.Project, Domain: input.Domain}).
		Limit(input.Limit).
		Offset(input.Offset).
		Group(identifierGroupBy)

	// Apply consistent sort ordering.
	if input.SortParameter != nil {
		tx = tx.Order(input.SortParameter.GetGormOrderExpr())
	}

	return db.Joins(fmt.Sprintf(joinString, input.ResourceType), tx)
}

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

var resourceTypeToMetadataJoin = map[core.ResourceType]string{
	core.ResourceType_LAUNCH_PLAN: leftJoinLaunchPlanNameToMetadata,
	core.ResourceType_WORKFLOW:    leftJoinWorkflowNameToMetadata,
	core.ResourceType_TASK:        leftJoinTaskNameToMetadata,
}

var getGroupByForNamedEntity = fmt.Sprintf("%s.%s, %s.%s, %s.%s, %s.%s, %s.%s",
	innerJoinTableAlias, Project, innerJoinTableAlias, Domain, innerJoinTableAlias, Name, namedEntityMetadataTableName,
	Description,
	namedEntityMetadataTableName, State)

func getSelectForNamedEntity(tableName string, resourceType core.ResourceType) []string {
	return []string{
		fmt.Sprintf("%s.%s", tableName, Project),
		fmt.Sprintf("%s.%s", tableName, Domain),
		fmt.Sprintf("%s.%s", tableName, Name),
		fmt.Sprintf("'%d' AS %s", resourceType, ResourceType),
		fmt.Sprintf("%s.%s", namedEntityMetadataTableName, Description),
		fmt.Sprintf("%s.%s", namedEntityMetadataTableName, State),
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
	}).Assign(input.NamedEntityMetadataFields).Omit("id").FirstOrCreate(&metadata)
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
		return models.NamedEntity{}, adminErrors.NewFlyteAdminErrorf(codes.InvalidArgument, "Cannot get NamedEntityMetadata for resource type: %v", input.ResourceType)
	}

	tx := r.db.Table(tableName).Joins(joinString)

	// Apply filters
	tx, err = applyScopedFilters(tx, filters, nil)
	if err != nil {
		return models.NamedEntity{}, err
	}

	timer := r.metrics.GetDuration.Start()
	tx = tx.Select(getSelectForNamedEntity(tableName, input.ResourceType)).Take(&namedEntity)
	timer.Stop()

	if tx.Error != nil {
		return models.NamedEntity{}, r.errorTransformer.ToFlyteAdminError(tx.Error)
	}

	return namedEntity, nil
}

func (r *NamedEntityRepo) List(ctx context.Context, input interfaces.ListNamedEntityInput) (
	interfaces.NamedEntityCollectionOutput, error) {

	// Validate input. Filters aren't required because they're implicit in the Project & Domain specified by the input.
	if len(input.Project) == 0 {
		return interfaces.NamedEntityCollectionOutput{}, errors.GetInvalidInputError(Project)
	}
	if len(input.Domain) == 0 {
		return interfaces.NamedEntityCollectionOutput{}, errors.GetInvalidInputError(Domain)
	}
	if input.Limit == 0 {
		return interfaces.NamedEntityCollectionOutput{}, errors.GetInvalidInputError(limit)
	}

	tableName, tableFound := resourceTypeToTableName[input.ResourceType]
	if !tableFound {
		return interfaces.NamedEntityCollectionOutput{}, adminErrors.NewFlyteAdminErrorf(codes.InvalidArgument,
			"Cannot list entity names for resource type: %v", input.ResourceType)
	}

	tx := getSubQueryJoin(r.db, tableName, input)

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

	tx.Select(getSelectForNamedEntity(innerJoinTableAlias, input.ResourceType)).Table(namedEntityMetadataTableName).Group(getGroupByForNamedEntity).Scan(&entities)

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
