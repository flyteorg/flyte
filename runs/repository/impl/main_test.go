package impl

import (
	"os"
	"testing"

	"gorm.io/gorm"

	"github.com/flyteorg/flyte/v2/flytestdlib/database"
	runsmigrations "github.com/flyteorg/flyte/v2/runs/migrations"
)

var testDB *gorm.DB

func TestMain(m *testing.M) {
	os.Exit(database.RunTestMain(m, 15432, "flyte_runs_test", &testDB, func(db *gorm.DB) error {
		return db.AutoMigrate(runsmigrations.AllModels...)
	}))
}
