package impl

import (
	"context"
	"database/sql"

	"github.com/jmoiron/sqlx"

	"github.com/flyteorg/flyte/v2/cache_service/repository/interfaces"
	"github.com/flyteorg/flyte/v2/cache_service/repository/models"
)

var _ interfaces.CachedOutputRepo = (*CachedOutputRepo)(nil)

type CachedOutputRepo struct {
	db *sqlx.DB
}

func NewCachedOutputRepo(db *sqlx.DB) *CachedOutputRepo {
	return &CachedOutputRepo{db: db}
}

func (r *CachedOutputRepo) Get(ctx context.Context, key string) (*models.CachedOutput, error) {
	var output models.CachedOutput
	err := sqlx.GetContext(ctx, r.db, &output,
		"SELECT * FROM cache_service_outputs WHERE key = $1", key)
	if err != nil {
		return nil, err
	}
	return &output, nil
}

func (r *CachedOutputRepo) Put(ctx context.Context, output *models.CachedOutput) error {
	_, err := r.db.ExecContext(ctx,
		`INSERT INTO cache_service_outputs (key, output_uri, metadata, last_updated, created_at, updated_at)
		 VALUES ($1, $2, $3, $4, CURRENT_TIMESTAMP, CURRENT_TIMESTAMP)
		 ON CONFLICT (key) DO UPDATE SET
		   output_uri = EXCLUDED.output_uri,
		   metadata = EXCLUDED.metadata,
		   last_updated = EXCLUDED.last_updated,
		   updated_at = CURRENT_TIMESTAMP`,
		output.Key, output.OutputURI, output.Metadata, output.LastUpdated)
	return err
}

func (r *CachedOutputRepo) Delete(ctx context.Context, key string) error {
	result, err := r.db.ExecContext(ctx,
		"DELETE FROM cache_service_outputs WHERE key = $1", key)
	if err != nil {
		return err
	}
	rows, _ := result.RowsAffected()
	if rows == 0 {
		return sql.ErrNoRows
	}
	return nil
}
