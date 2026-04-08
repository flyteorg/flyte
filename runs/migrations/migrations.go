package migrations

import (
	"context"
	"embed"

	"github.com/jmoiron/sqlx"

	"github.com/flyteorg/flyte/v2/flytestdlib/database"
)

//go:embed sql/*.sql
var migrationFS embed.FS

// RunMigrations applies all pending runs service migrations.
func RunMigrations(ctx context.Context, db *sqlx.DB) error {
	return database.Migrate(ctx, db, "runs", migrationFS)
}
