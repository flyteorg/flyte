package manager

import (
	"os"
	"testing"

	"github.com/go-gormigrate/gormigrate/v2"
	"gorm.io/gorm"

	"github.com/flyteorg/flyte/v2/cache_service/migrations"
	"github.com/flyteorg/flyte/v2/flytestdlib/database"
)

var testDB *gorm.DB

func TestMain(m *testing.M) {
	os.Exit(database.RunTestMain(m, 15434, "flyte_cache_test", &testDB, func(db *gorm.DB) error {
		return gormigrate.New(db, gormigrate.DefaultOptions, migrations.CacheServiceMigrations).Migrate()
	}))
}
