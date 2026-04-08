package service

import (
	"context"
	"os"
	"testing"

	"github.com/jmoiron/sqlx"

	"github.com/flyteorg/flyte/v2/flytestdlib/database"
	runsmigrations "github.com/flyteorg/flyte/v2/runs/migrations"
)

var testDB *sqlx.DB

func TestMain(m *testing.M) {
	os.Exit(database.RunTestMain(m, 15433, "flyte_runs_test", &testDB, func(db *sqlx.DB) error {
		return runsmigrations.RunMigrations(context.Background(), db)
	}))
}
