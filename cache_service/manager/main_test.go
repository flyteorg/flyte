package manager

import (
	"fmt"
	"log"
	"os"
	"testing"

	embeddedpostgres "github.com/fergusstrange/embedded-postgres"
	"github.com/go-gormigrate/gormigrate/v2"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"

	"github.com/flyteorg/flyte/v2/cache_service/migrations"
)

const testDBPort = 15434

var testDB *gorm.DB

func TestMain(m *testing.M) {
	pg := embeddedpostgres.NewDatabase(
		embeddedpostgres.DefaultConfig().
			Port(testDBPort).
			Database("flyte_cache_test").
			Username("postgres").
			Password("postgres").
			RuntimePath(fmt.Sprintf("/tmp/embedded-postgres-%d", testDBPort)),
	)

	if err := pg.Start(); err != nil {
		log.Fatalf("failed to start embedded postgres: %v", err)
	}

	dsn := fmt.Sprintf(
		"host=localhost port=%d user=postgres password=postgres dbname=flyte_cache_test sslmode=disable",
		testDBPort,
	)
	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{})
	if err != nil {
		pg.Stop()
		log.Fatalf("failed to connect to embedded postgres: %v", err)
	}

	mg := gormigrate.New(db, gormigrate.DefaultOptions, migrations.CacheServiceMigrations)
	if err := mg.Migrate(); err != nil {
		pg.Stop()
		log.Fatalf("failed to run migrations: %v", err)
	}

	testDB = db
	code := m.Run()

	if err := pg.Stop(); err != nil {
		log.Printf("warning: failed to stop embedded postgres: %v", err)
	}
	os.Exit(code)
}
