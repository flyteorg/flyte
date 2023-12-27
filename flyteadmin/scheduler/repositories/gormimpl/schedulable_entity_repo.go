package gormimpl

import (
	"context"
	"errors"
	"fmt"

	"gorm.io/gorm"

	adminErrors "github.com/flyteorg/flyte/flyteadmin/pkg/repositories/errors"
	"github.com/flyteorg/flyte/flyteadmin/scheduler/repositories/interfaces"
	"github.com/flyteorg/flyte/flyteadmin/scheduler/repositories/models"
	"github.com/flyteorg/flyte/flyteidl/gen/pb-go/flyteidl/core"
	"github.com/flyteorg/flyte/flytestdlib/promutils"
)

// SchedulableEntityRepo Implementation of SchedulableEntityRepoInterface.
type SchedulableEntityRepo struct {
	db               *gorm.DB
	errorTransformer adminErrors.ErrorTransformer
	metrics          gormMetrics
}

func (r *SchedulableEntityRepo) Create(ctx context.Context, id *core.Identifier, input models.SchedulableEntity) error {
	timer := r.metrics.GetDuration.Start()
	var record models.SchedulableEntity
	tx := r.db.Omit("id").FirstOrCreate(&record, input)
	timer.Stop()
	if tx.Error != nil {
		return r.errorTransformer.ToFlyteAdminError(tx.Error)
	}
	return nil
}

func (r *SchedulableEntityRepo) Activate(ctx context.Context, id *core.Identifier, input models.SchedulableEntity) error {
	var schedulableEntity models.SchedulableEntity
	timer := r.metrics.GetDuration.Start()
	// Find the existence of a scheduled entity
	tx := r.db.Where(&models.SchedulableEntity{
		SchedulableEntityKey: models.SchedulableEntityKey{
			Project: id.Project,
			Domain:  id.Domain,
			Name:    id.Name,
			Version: id.Version,
		},
	}).Take(&schedulableEntity)
	timer.Stop()

	if tx.Error != nil {
		if errors.Is(tx.Error, gorm.ErrRecordNotFound) {
			// Not found and hence create one
			return r.Create(ctx, id, input)
		}
		return r.errorTransformer.ToFlyteAdminError(tx.Error)
	}

	// Activate the already existing schedule
	return activateOrDeactivate(r, id, true)
}

func (r *SchedulableEntityRepo) Deactivate(ctx context.Context, id *core.Identifier) error {
	// Activate the schedule
	return activateOrDeactivate(r, id, false)
}

func (r *SchedulableEntityRepo) GetAll(ctx context.Context) ([]models.SchedulableEntity, error) {
	var schedulableEntities []models.SchedulableEntity
	timer := r.metrics.GetDuration.Start()

	tx := r.db.Find(&schedulableEntities)

	timer.Stop()

	if tx.Error != nil {
		if errors.Is(tx.Error, gorm.ErrRecordNotFound) {
			return nil,
				fmt.Errorf("no active schedulable entities found")
		}
		return nil, r.errorTransformer.ToFlyteAdminError(tx.Error)
	}

	return schedulableEntities, nil
}

func (r *SchedulableEntityRepo) Get(ctx context.Context, id *core.Identifier) (models.SchedulableEntity, error) {
	var schedulableEntity models.SchedulableEntity
	timer := r.metrics.GetDuration.Start()
	tx := r.db.Where(&models.SchedulableEntity{
		SchedulableEntityKey: models.SchedulableEntityKey{
			Project: id.Project,
			Domain:  id.Domain,
			Name:    id.Name,
			Version: id.Version,
		},
	}).Take(&schedulableEntity)
	timer.Stop()

	if tx.Error != nil {
		if errors.Is(tx.Error, gorm.ErrRecordNotFound) {
			return models.SchedulableEntity{},
				adminErrors.GetMissingEntityError("schedulable entity", id)
		}
		return models.SchedulableEntity{}, r.errorTransformer.ToFlyteAdminError(tx.Error)
	}

	return schedulableEntity, nil
}

// Helper function to activate and deactivate a schedule
func activateOrDeactivate(r *SchedulableEntityRepo, id *core.Identifier, activate bool) error {
	timer := r.metrics.GetDuration.Start()
	tx := r.db.Model(&models.SchedulableEntity{}).Where(&models.SchedulableEntity{
		SchedulableEntityKey: models.SchedulableEntityKey{
			Project: id.Project,
			Domain:  id.Domain,
			Name:    id.Name,
			Version: id.Version,
		},
	}).Update("active", activate)
	timer.Stop()
	if tx.Error != nil {
		if errors.Is(tx.Error, gorm.ErrRecordNotFound) {
			return adminErrors.GetMissingEntityError("schedulable entity", id)
		}
		return r.errorTransformer.ToFlyteAdminError(tx.Error)
	}
	return nil
}

// NewSchedulableEntityRepo Returns an instance of SchedulableEntityRepoInterface
func NewSchedulableEntityRepo(
	db *gorm.DB, errorTransformer adminErrors.ErrorTransformer, scope promutils.Scope) interfaces.SchedulableEntityRepoInterface {
	metrics := newMetrics(scope)
	return &SchedulableEntityRepo{
		db:               db,
		errorTransformer: errorTransformer,
		metrics:          metrics,
	}
}
