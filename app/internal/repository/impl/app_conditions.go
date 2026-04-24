package impl

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"time"

	"github.com/jmoiron/sqlx"
	"google.golang.org/protobuf/proto"

	"github.com/flyteorg/flyte/v2/app/internal/repository/interfaces"
	"github.com/flyteorg/flyte/v2/flytestdlib/logger"
	flyteapp "github.com/flyteorg/flyte/v2/gen/go/flyteidl2/app"
)

var _ interfaces.AppConditionsRepo = (*AppConditionsRepo)(nil)

type AppConditionsRepo struct {
	db *sqlx.DB
}

func NewAppConditionsRepo(db *sqlx.DB) *AppConditionsRepo {
	return &AppConditionsRepo{db: db}
}

// AppendCondition appends cond to the stored conditions for the given app within a
// single transaction. If the total exceeds maxConditions, the oldest entries are
// trimmed. Creates the row if it does not exist.
func (r *AppConditionsRepo) AppendCondition(ctx context.Context, appID *flyteapp.Identifier, cond *flyteapp.Condition, maxConditions int) error {
	tx, err := r.db.BeginTxx(ctx, nil)
	if err != nil {
		return fmt.Errorf("failed to begin transaction for app conditions: %w", err)
	}
	defer func() {
		if err != nil {
			_ = tx.Rollback()
		}
	}()

	// Read existing conditions inside the transaction.
	var raw []byte
	err = tx.QueryRowContext(ctx,
		"SELECT conditions FROM app_conditions WHERE project = $1 AND domain = $2 AND name = $3 FOR UPDATE",
		appID.GetProject(), appID.GetDomain(), appID.GetName(),
	).Scan(&raw)
	if err != nil && !errors.Is(err, sql.ErrNoRows) {
		return fmt.Errorf("failed to read app conditions for %s/%s/%s: %w", appID.GetProject(), appID.GetDomain(), appID.GetName(), err)
	}

	conditions, err := unmarshalConditions(raw)
	if err != nil {
		return fmt.Errorf("failed to unmarshal app conditions for %s/%s/%s: %w", appID.GetProject(), appID.GetDomain(), appID.GetName(), err)
	}

	// Append and trim.
	conditions = append(conditions, cond)
	if maxConditions > 0 && len(conditions) > maxConditions {
		logger.Debugf(ctx, "Trimming conditions for app %s/%s/%s from %d to %d",
			appID.GetProject(), appID.GetDomain(), appID.GetName(), len(conditions), maxConditions)
		conditions = conditions[len(conditions)-maxConditions:]
	}

	raw, err = marshalConditions(conditions)
	if err != nil {
		return fmt.Errorf("failed to marshal app conditions for %s/%s/%s: %w", appID.GetProject(), appID.GetDomain(), appID.GetName(), err)
	}

	_, err = tx.ExecContext(ctx,
		`INSERT INTO app_conditions (project, domain, name, conditions, updated_at)
		 VALUES ($1, $2, $3, $4, $5)
		 ON CONFLICT (project, domain, name) DO UPDATE SET
		   conditions = EXCLUDED.conditions,
		   updated_at = EXCLUDED.updated_at`,
		appID.GetProject(), appID.GetDomain(), appID.GetName(), raw, time.Now().UTC(),
	)
	if err != nil {
		return fmt.Errorf("failed to upsert app conditions for %s/%s/%s: %w", appID.GetProject(), appID.GetDomain(), appID.GetName(), err)
	}

	return tx.Commit()
}

// GetConditions returns the stored conditions for the given app.
// Returns nil if no conditions have been recorded yet.
func (r *AppConditionsRepo) GetConditions(ctx context.Context, appID *flyteapp.Identifier) ([]*flyteapp.Condition, error) {
	var raw []byte
	err := r.db.QueryRowContext(ctx,
		"SELECT conditions FROM app_conditions WHERE project = $1 AND domain = $2 AND name = $3",
		appID.GetProject(), appID.GetDomain(), appID.GetName(),
	).Scan(&raw)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get app conditions for %s/%s/%s: %w", appID.GetProject(), appID.GetDomain(), appID.GetName(), err)
	}

	conditions, err := unmarshalConditions(raw)
	if err != nil {
		return nil, fmt.Errorf("failed to unmarshal app conditions for %s/%s/%s: %w", appID.GetProject(), appID.GetDomain(), appID.GetName(), err)
	}
	return conditions, nil
}

// DeleteConditions removes the conditions row for the given app.
// No-ops if the row does not exist.
func (r *AppConditionsRepo) DeleteConditions(ctx context.Context, appID *flyteapp.Identifier) error {
	_, err := r.db.ExecContext(ctx,
		"DELETE FROM app_conditions WHERE project = $1 AND domain = $2 AND name = $3",
		appID.GetProject(), appID.GetDomain(), appID.GetName(),
	)
	if err != nil {
		return fmt.Errorf("failed to delete app conditions for %s/%s/%s: %w", appID.GetProject(), appID.GetDomain(), appID.GetName(), err)
	}
	return nil
}

// marshalConditions serializes a slice of Condition using Status as a proto wrapper.
func marshalConditions(conditions []*flyteapp.Condition) ([]byte, error) {
	return proto.Marshal(&flyteapp.Status{Conditions: conditions})
}

// unmarshalConditions deserializes conditions from bytes produced by marshalConditions.
// Returns nil for empty/nil input.
func unmarshalConditions(raw []byte) ([]*flyteapp.Condition, error) {
	if len(raw) == 0 {
		return nil, nil
	}
	status := &flyteapp.Status{}
	if err := proto.Unmarshal(raw, status); err != nil {
		return nil, err
	}
	return status.GetConditions(), nil
}
