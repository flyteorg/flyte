package impl

import (
	"fmt"
	"log"
	"os"
	"testing"

	embeddedpostgres "github.com/fergusstrange/embedded-postgres"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"

	runsmigrations "github.com/flyteorg/flyte/v2/runs/migrations"
)

const testDBPort = 15432

var testDB *gorm.DB

func TestMain(m *testing.M) {
	pg := embeddedpostgres.NewDatabase(
		embeddedpostgres.DefaultConfig().
			Port(testDBPort).
			Database("flyte_runs_test").
			Username("postgres").
			Password("postgres").
			RuntimePath(fmt.Sprintf("/tmp/embedded-postgres-%d", testDBPort)),
	)

	if err := pg.Start(); err != nil {
		log.Fatalf("failed to start embedded postgres: %v", err)
	}

	dsn := fmt.Sprintf(
		"host=localhost port=%d user=postgres password=postgres dbname=flyte_runs_test sslmode=disable",
		testDBPort,
	)
	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{TranslateError: true})
	if err != nil {
		pg.Stop()
		log.Fatalf("failed to connect to embedded postgres: %v", err)
	}

	if err := db.AutoMigrate(runsmigrations.AllModels...); err != nil {
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
