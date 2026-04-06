package database

import (
	"fmt"
	"log"
	"testing"

	embeddedpostgres "github.com/fergusstrange/embedded-postgres"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

// RunTestMain starts an embedded PostgreSQL instance, runs migrate, executes m.Run(),
// and returns an exit code suitable for os.Exit. The connected *gorm.DB is written to db.
// Intended for use in TestMain functions:
//
//	func TestMain(m *testing.M) {
//	    os.Exit(database.RunTestMain(m, 15432, "mydb", &testDB, func(db *gorm.DB) error {
//	        return db.AutoMigrate(...)
//	    }))
//	}
func RunTestMain(m *testing.M, port uint32, dbName string, db **gorm.DB, migrate func(*gorm.DB) error) int {
	pg := embeddedpostgres.NewDatabase(
		embeddedpostgres.DefaultConfig().
			Port(port).
			Database(dbName).
			Username("postgres").
			Password("postgres").
			RuntimePath(fmt.Sprintf("/tmp/embedded-postgres-%d", port)),
	)
	if err := pg.Start(); err != nil {
		log.Printf("failed to start embedded postgres on port %d: %v", port, err)
		return 1
	}

	dsn := fmt.Sprintf(
		"host=localhost port=%d user=postgres password=postgres dbname=%s sslmode=disable",
		port, dbName,
	)
	conn, err := gorm.Open(postgres.Open(dsn), &gorm.Config{TranslateError: true})
	if err != nil {
		_ = pg.Stop()
		log.Printf("failed to connect to embedded postgres: %v", err)
		return 1
	}

	if err := migrate(conn); err != nil {
		_ = pg.Stop()
		log.Printf("failed to run migrations: %v", err)
		return 1
	}

	*db = conn
	code := m.Run()

	if err := pg.Stop(); err != nil {
		log.Printf("warning: failed to stop embedded postgres: %v", err)
	}
	return code
}
