package service

import (
	"log"
	"os"
	"testing"

	"gorm.io/gorm"

	"github.com/flyteorg/flyte/v2/flytestdlib/database"
	runsmigrations "github.com/flyteorg/flyte/v2/runs/migrations"
)

const testDBPort = 15433

var testDB *gorm.DB

func TestMain(m *testing.M) {
	db, stop, err := database.StartEmbeddedPostgres(testDBPort, "flyte_runs_test")
	if err != nil {
		log.Fatalf("failed to start embedded postgres: %v", err)
	}

	if err := db.AutoMigrate(runsmigrations.AllModels...); err != nil {
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
