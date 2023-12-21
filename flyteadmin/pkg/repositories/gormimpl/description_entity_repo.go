package gormimpl

import (
	"context"
	"github.com/flyteorg/flyte/flyteidl/gen/pb-go/flyteidl/admin"

	"gorm.io/gorm"

	"github.com/flyteorg/flyte/flyteadmin/pkg/common"
	flyteAdminDbErrors "github.com/flyteorg/flyte/flyteadmin/pkg/repositories/errors"
	"github.com/flyteorg/flyte/flyteadmin/pkg/repositories/interfaces"
	"github.com/flyteorg/flyte/flyteadmin/pkg/repositories/models"
	"github.com/flyteorg/flyte/flyteidl/gen/pb-go/flyteidl/core"
	"github.com/flyteorg/flyte/flytestdlib/promutils"
)

// DescriptionEntityRepo Implementation of DescriptionEntityRepoInterface.
type DescriptionEntityRepo struct {
	db               *gorm.DB
	errorTransformer flyteAdminDbErrors.ErrorTransformer
	metrics          gormMetrics
}

func (r *DescriptionEntityRepo) Get(ctx context.Context, id *core.Identifier) (models.DescriptionEntity, error) {
	var descriptionEntity models.DescriptionEntity

	entity := common.ResourceTypeToEntity[id.GetResourceType()]
	versionFilter, err := common.NewSingleValueFilter(entity, common.Equal, Version, id.GetVersion())
	if err != nil {
		return models.DescriptionEntity{}, err
	}

	tx := r.db.WithContext(ctx).Table(descriptionEntityTableName)
	// Apply filters
	tx, err = applyFilters(tx, entity, &admin.NamedEntityIdentifier{
		Project: id.GetProject(),
		Domain:  id.GetDomain(),
		Org:     id.GetOrg(),
		Name:    id.GetName(),
	}, []common.InlineFilter{versionFilter}, nil)
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
	ctx context.Context, entity common.Entity, input interfaces.ListResourceInput) (interfaces.DescriptionEntityCollectionOutput, error) {
	// First validate input.
	if err := ValidateListInput(input); err != nil {
		return interfaces.DescriptionEntityCollectionOutput{}, err
	}
	var descriptionEntities []models.DescriptionEntity
	tx := r.db.WithContext(ctx).Limit(input.Limit).Offset(input.Offset)

	// Apply filters
	tx, err := applyFilters(tx, entity, input.IdentifierScope, input.InlineFilters, input.MapFilters)
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
