package server

import (
	"context"
	"fmt"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"

	"github.com/flyteorg/flyte/flyteadmin/pkg/runtime/interfaces"
	"github.com/flyteorg/flyte/flytestdlib/logger"
)

// GetDB initializes and returns a database connection using the application configuration
func GetDB(ctx context.Context, appConfig interfaces.ApplicationConfiguration) (*gorm.DB, error) {
	dbConfig := appConfig.GetDbConfig()
	if dbConfig == nil {
		return nil, fmt.Errorf("no database configuration provided")
	}

	// Use the Postgres configuration
	postgresConfig := dbConfig.Postgres
	logger.Infof(ctx, "Connecting to database: %s@%s:%d/%s",
		postgresConfig.User, postgresConfig.Host, postgresConfig.Port, postgresConfig.DbName)

	dsn := fmt.Sprintf("host=%s port=%d user=%s dbname=%s password=%s %s",
		postgresConfig.Host, postgresConfig.Port, postgresConfig.User, postgresConfig.DbName,
		postgresConfig.Password, postgresConfig.ExtraOptions)

	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{})
	if err != nil {
		return nil, fmt.Errorf("failed to connect to database: %w", err)
	}

	// Test connection
	sqlDB, err := db.DB()
	if err != nil {
		return nil, fmt.Errorf("failed to get database connection: %w", err)
	}

	if err := sqlDB.Ping(); err != nil {
		return nil, fmt.Errorf("failed to ping database: %w", err)
	}

	return db, nil
}

// RepoConfig provides configuration for repository operations
type RepoConfig struct {
	Schema string
}

// GetRepoConfig returns the repository configuration from the application configuration
func GetRepoConfig(appConfig interfaces.ApplicationConfiguration) *RepoConfig {
	dbConfig := appConfig.GetDbConfig()
	if dbConfig == nil {
		return &RepoConfig{
			Schema: "public",
		}
	}

	// Default to public schema if not specified
	schema := "public"

	return &RepoConfig{
		Schema: schema,
	}
}
