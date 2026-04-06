package database

import (
	"fmt"

	embeddedpostgres "github.com/fergusstrange/embedded-postgres"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

// StartEmbeddedPostgres starts an embedded PostgreSQL instance on the given port,
// connects via GORM, and returns the DB and a stop function.
// Intended for use in TestMain functions. The caller must call stop() when done.
func StartEmbeddedPostgres(port uint32, dbName string) (*gorm.DB, func() error, error) {
	pg := embeddedpostgres.NewDatabase(
		embeddedpostgres.DefaultConfig().
			Port(port).
			Database(dbName).
			Username("postgres").
			Password("postgres").
			RuntimePath(fmt.Sprintf("/tmp/embedded-postgres-%d", port)),
	)
	if err := pg.Start(); err != nil {
		return nil, nil, fmt.Errorf("failed to start embedded postgres on port %d: %w", port, err)
	}

	dsn := fmt.Sprintf(
		"host=localhost port=%d user=postgres password=postgres dbname=%s sslmode=disable",
		port, dbName,
	)
	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{TranslateError: true})
	if err != nil {
		_ = pg.Stop()
		return nil, nil, fmt.Errorf("failed to connect to embedded postgres: %w", err)
	}

	return db, pg.Stop, nil
}
