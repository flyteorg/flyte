package manager

import (
	"log"
	"os"
	"testing"

	"github.com/go-gormigrate/gormigrate/v2"
	"gorm.io/gorm"

	"github.com/flyteorg/flyte/v2/cache_service/migrations"
	"github.com/flyteorg/flyte/v2/flytestdlib/database"
)

const testDBPort = 15434

var testDB *gorm.DB

func TestMain(m *testing.M) {
	db, stop, err := database.StartEmbeddedPostgres(testDBPort, "flyte_cache_test")
	if err != nil {
		log.Fatalf("failed to start embedded postgres: %v", err)
	}

	if err := gormigrate.New(db, gormigrate.DefaultOptions, migrations.CacheServiceMigrations).Migrate(); err != nil {
		_ = stop()
		log.Fatalf("failed to run migrations: %v", err)
	}

	testDB = db
	code := m.Run()

	if err := stop(); err != nil {
		log.Printf("warning: failed to stop embedded postgres: %v", err)
	}
	os.Exit(code)
}
