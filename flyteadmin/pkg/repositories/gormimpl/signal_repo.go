package gormimpl

import (
	"context"
	"errors"

	"google.golang.org/grpc/codes"
	"gorm.io/gorm"

	"github.com/flyteorg/flyte/flyteadmin/auth/isolation"
	"github.com/flyteorg/flyte/flyteadmin/pkg/common"
	adminerrors "github.com/flyteorg/flyte/flyteadmin/pkg/errors"
	"github.com/flyteorg/flyte/flyteadmin/pkg/manager/impl/util"
	flyteAdminDbErrors "github.com/flyteorg/flyte/flyteadmin/pkg/repositories/errors"
	"github.com/flyteorg/flyte/flyteadmin/pkg/repositories/interfaces"
	"github.com/flyteorg/flyte/flyteadmin/pkg/repositories/models"
	"github.com/flyteorg/flyte/flytestdlib/promutils"
)

var (
	signalResourceColumns = common.ResourceColumns{Project: "execution_project", Domain: "execution_domain"}
)

// SignalRepo is an implementation of SignalRepoInterface.
type SignalRepo struct {
	db               *gorm.DB
	errorTransformer flyteAdminDbErrors.ErrorTransformer
	metrics          gormMetrics
}

// Get retrieves a signal model from the database store.
func (s *SignalRepo) Get(ctx context.Context, input models.SignalKey) (models.Signal, error) {
	isolationFilter := util.GetIsolationFilter(ctx, isolation.DomainTargetResourceScopeDepth, signalResourceColumns)
	var signal models.Signal
	timer := s.metrics.GetDuration.Start()
	tx := s.db.WithContext(ctx).Where(&models.Signal{
		SignalKey: input,
	})
	if isolationFilter != nil {
		cleanSession := tx.Session(&gorm.Session{NewDB: true})
		tx = tx.Where(cleanSession.Scopes(isolationFilter.GetScopes()...))
	}
	tx = tx.Take(&signal)
	timer.Stop()
	if errors.Is(tx.Error, gorm.ErrRecordNotFound) {
		return models.Signal{}, adminerrors.NewFlyteAdminError(codes.NotFound, "signal does not exist")
	}
	if tx.Error != nil {
		return models.Signal{}, s.errorTransformer.ToFlyteAdminError(tx.Error)
	}
	return signal, nil
}

// GetOrCreate returns a signal if it already exists, if not it creates a new one given the input
func (s *SignalRepo) GetOrCreate(ctx context.Context, input *models.Signal) error {
	if err := util.FilterResourceMutation(ctx, input.Project, input.Domain); err != nil {
		return err
	}
	timer := s.metrics.CreateDuration.Start()
	tx := s.db.WithContext(ctx).FirstOrCreate(&input, input)
	timer.Stop()
	if tx.Error != nil {
		return s.errorTransformer.ToFlyteAdminError(tx.Error)
	}
	return nil
}

// List fetches all signals that match the provided input
func (s *SignalRepo) List(ctx context.Context, input interfaces.ListResourceInput) ([]models.Signal, error) {
	// First validate input.
	if err := ValidateListInput(input); err != nil {
		return nil, err
	}
	isolationFilter := util.GetIsolationFilter(ctx, isolation.DomainTargetResourceScopeDepth, signalResourceColumns)
	var signals []models.Signal
	tx := s.db.WithContext(ctx).Limit(input.Limit).Offset(input.Offset)

	// Apply filters
	tx, err := applyFilters(tx, input.InlineFilters, input.MapFilters, isolationFilter)
	if err != nil {
		return nil, err
	}
	// Apply sort ordering.
	if input.SortParameter != nil {
		tx = tx.Order(input.SortParameter.GetGormOrderExpr())
	}
	timer := s.metrics.ListDuration.Start()
	tx.Find(&signals)
	timer.Stop()
	if tx.Error != nil {
		return nil, s.errorTransformer.ToFlyteAdminError(tx.Error)
	}

	return signals, nil
}

// Update sets the value field on the specified signal model
func (s *SignalRepo) Update(ctx context.Context, input models.SignalKey, value []byte) error {
	if err := util.FilterResourceMutation(ctx, input.Project, input.Domain); err != nil {
		return err
	}
	signal := models.Signal{
		SignalKey: input,
		Value:     value,
	}

	timer := s.metrics.GetDuration.Start()
	tx := s.db.WithContext(ctx).Model(&signal).Select("value").Updates(signal)
	timer.Stop()
	if tx.Error != nil {
		return s.errorTransformer.ToFlyteAdminError(tx.Error)
	}
	if tx.RowsAffected == 0 {
		return adminerrors.NewFlyteAdminError(codes.NotFound, "signal does not exist")
	}
	return nil
}

// Returns an instance of SignalRepoInterface
func NewSignalRepo(
	db *gorm.DB, errorTransformer flyteAdminDbErrors.ErrorTransformer, scope promutils.Scope) interfaces.SignalRepoInterface {
	metrics := newMetrics(scope)
	return &SignalRepo{
		db:               db,
		errorTransformer: errorTransformer,
		metrics:          metrics,
	}
}
