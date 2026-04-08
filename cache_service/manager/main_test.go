package manager

import (
	"context"
	"os"
	"testing"

	"github.com/jmoiron/sqlx"

	"github.com/flyteorg/flyte/v2/cache_service/migrations"
	"github.com/flyteorg/flyte/v2/flytestdlib/database"
)

var testDB *sqlx.DB

func TestMain(m *testing.M) {
	os.Exit(database.RunTestMain(m, 15434, "flyte_cache_test", &testDB, func(db *sqlx.DB) error {
		return migrations.RunMigrations(context.Background(), db)
	}))
}
