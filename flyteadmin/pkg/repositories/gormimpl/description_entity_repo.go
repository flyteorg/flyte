package gormimpl

import (
	"context"

	"github.com/flyteorg/flyteidl/gen/pb-go/flyteidl/core"
	"github.com/flyteorg/flytestdlib/promutils"
	"gorm.io/gorm"

	"github.com/flyteorg/flyteadmin/pkg/common"
	flyteAdminDbErrors "github.com/flyteorg/flyteadmin/pkg/repositories/errors"
	"github.com/flyteorg/flyteadmin/pkg/repositories/interfaces"
	"github.com/flyteorg/flyteadmin/pkg/repositories/models"
)

// DescriptionEntityRepo Implementation of DescriptionEntityRepoInterface.
type DescriptionEntityRepo struct {
	db               *gorm.DB
	errorTransformer flyteAdminDbErrors.ErrorTransformer
	metrics          gormMetrics
}

func (r *DescriptionEntityRepo) Get(ctx context.Context, input interfaces.GetDescriptionEntityInput) (models.DescriptionEntity, error) {
	var descriptionEntity models.DescriptionEntity

	filters, err := getDescriptionEntityFilters(input.ResourceType, input.Project, input.Domain, input.Name, input.Version)
	if err != nil {
		return models.DescriptionEntity{}, err
	}

	tx := r.db.Table(descriptionEntityTableName)
	// Apply filters
	tx, err = applyFilters(tx, filters, nil)
	if err != nil {
		return models.DescriptionEntity{}, err
	}

	timer := r.metrics.GetDuration.Start()
	tx = tx.Take(&descriptionEntity)
	timer.Stop()

	if tx.Error != nil {
		return models.DescriptionEntity{}, r.errorTransformer.ToFlyteAdminError(tx.Error)
	}

	return descriptionEntity, nil
}

func (r *DescriptionEntityRepo) List(
	ctx context.Context, input interfaces.ListResourceInput) (interfaces.DescriptionEntityCollectionOutput, error) {
	// First validate input.
	if err := ValidateListInput(input); err != nil {
		return interfaces.DescriptionEntityCollectionOutput{}, err
	}
	var descriptionEntities []models.DescriptionEntity
	tx := r.db.Limit(input.Limit).Offset(input.Offset)

	// Apply filters
	tx, err := applyFilters(tx, input.InlineFilters, input.MapFilters)
	if err != nil {
		return interfaces.DescriptionEntityCollectionOutput{}, err
	}
	// Apply sort ordering.
	if input.SortParameter != nil {
		tx = tx.Order(input.SortParameter.GetGormOrderExpr())
	}
	timer := r.metrics.ListDuration.Start()
	tx.Find(&descriptionEntities)
	timer.Stop()
	if tx.Error != nil {
		return interfaces.DescriptionEntityCollectionOutput{}, r.errorTransformer.ToFlyteAdminError(tx.Error)
	}
	return interfaces.DescriptionEntityCollectionOutput{
		Entities: descriptionEntities,
	}, nil
}

func getDescriptionEntityFilters(resourceType core.ResourceType, project string, domain string, name string, version string) ([]common.InlineFilter, error) {
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
	versionFilter, err := common.NewSingleValueFilter(entity, common.Equal, Version, version)
	if err != nil {
		return nil, err
	}
	filters = append(filters, versionFilter)

	return filters, nil
}

// NewDescriptionEntityRepo Returns an instance of DescriptionRepoInterface
func NewDescriptionEntityRepo(
	db *gorm.DB, errorTransformer flyteAdminDbErrors.ErrorTransformer, scope promutils.Scope) interfaces.DescriptionEntityRepoInterface {
	metrics := newMetrics(scope)
	return &DescriptionEntityRepo{
		db:               db,
		errorTransformer: errorTransformer,
		metrics:          metrics,
	}
}
