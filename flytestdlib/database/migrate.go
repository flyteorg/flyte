package database

import (
	"context"
	"fmt"
	"io/fs"
	"sort"
	"strings"
	"time"

	"github.com/jmoiron/sqlx"

	"github.com/flyteorg/flyte/v2/flytestdlib/logger"
)

const schemaMigrationsTable = `
CREATE TABLE IF NOT EXISTS schema_migrations (
	version    TEXT PRIMARY KEY,
	applied_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
)`

// migrationAdvisoryLockID is a fixed key for pg_advisory_lock to prevent
// concurrent replicas from racing on migration application.
const migrationAdvisoryLockID = 5432

// Migrate applies all unapplied SQL migration files from the provided fs.FS.
// Migration files are expected to be located under "sql/" and follow the naming
// convention "NNNN_description.sql". Files ending with "_down.sql" are skipped.
// Migrations are applied in lexicographic order, each within its own transaction.
//
// The prefix parameter namespaces versions in the shared schema_migrations table
// so that multiple services sharing the same database do not collide
// (e.g. prefix "runs" records version "runs/001_init_schema").
func Migrate(ctx context.Context, db *sqlx.DB, prefix string, migrations fs.FS) error {
	// Acquire an advisory lock so that concurrent replicas do not race on migrations.
	if _, err := db.ExecContext(ctx, "SELECT pg_advisory_lock($1)", migrationAdvisoryLockID); err != nil {
		return fmt.Errorf("failed to acquire migration advisory lock: %w", err)
	}
	defer func() {
		if _, err := db.ExecContext(ctx, "SELECT pg_advisory_unlock($1)", migrationAdvisoryLockID); err != nil {
			logger.Errorf(ctx, "failed to release migration advisory lock: %v", err)
		}
	}()

	if _, err := db.ExecContext(ctx, schemaMigrationsTable); err != nil {
		return fmt.Errorf("failed to create schema_migrations table: %w", err)
	}

	files, err := collectUpMigrations(migrations)
	if err != nil {
		return err
	}

	for _, f := range files {
		version := prefix + "/" + migrationVersion(f)

		var count int
		if err := db.GetContext(ctx, &count, "SELECT COUNT(*) FROM schema_migrations WHERE version = $1", version); err != nil {
			return fmt.Errorf("failed to check migration status for %s: %w", version, err)
		}
		if count > 0 {
			continue
		}

		content, err := fs.ReadFile(migrations, f)
		if err != nil {
			return fmt.Errorf("failed to read migration file %s: %w", f, err)
		}

		tx, err := db.BeginTxx(ctx, nil)
		if err != nil {
			return fmt.Errorf("failed to begin transaction for migration %s: %w", version, err)
		}

		if _, err := tx.ExecContext(ctx, string(content)); err != nil {
			_ = tx.Rollback()
			return fmt.Errorf("failed to execute migration %s: %w", version, err)
		}

		if _, err := tx.ExecContext(ctx, "INSERT INTO schema_migrations (version, applied_at) VALUES ($1, $2)", version, time.Now().UTC()); err != nil {
			_ = tx.Rollback()
			return fmt.Errorf("failed to record migration %s: %w", version, err)
		}

		if err := tx.Commit(); err != nil {
			return fmt.Errorf("failed to commit migration %s: %w", version, err)
		}

		logger.Infof(ctx, "Applied migration: %s", version)
	}

	logger.Infof(ctx, "All migrations applied successfully")
	return nil
}

// Rollback rolls back the most recently applied migration for the given prefix
// by executing its corresponding _down.sql file from the provided fs.FS.
func Rollback(ctx context.Context, db *sqlx.DB, prefix string, migrations fs.FS) error {
	if _, err := db.ExecContext(ctx, schemaMigrationsTable); err != nil {
		return fmt.Errorf("failed to create schema_migrations table: %w", err)
	}

	var version string
	err := db.GetContext(ctx, &version, "SELECT version FROM schema_migrations WHERE version LIKE $1 ORDER BY version DESC LIMIT 1", prefix+"/%")
	if err != nil {
		return fmt.Errorf("failed to get latest migration: %w", err)
	}

	// Find the corresponding down file
	downFile, err := findDownMigration(migrations, version)
	if err != nil {
		return err
	}

	content, err := fs.ReadFile(migrations, downFile)
	if err != nil {
		return fmt.Errorf("failed to read down migration file %s: %w", downFile, err)
	}

	tx, err := db.BeginTxx(ctx, nil)
	if err != nil {
		return fmt.Errorf("failed to begin transaction for rollback %s: %w", version, err)
	}

	if _, err := tx.ExecContext(ctx, string(content)); err != nil {
		_ = tx.Rollback()
		return fmt.Errorf("failed to execute rollback %s: %w", version, err)
	}

	if _, err := tx.ExecContext(ctx, "DELETE FROM schema_migrations WHERE version = $1", version); err != nil {
		_ = tx.Rollback()
		return fmt.Errorf("failed to remove migration record %s: %w", version, err)
	}

	if err := tx.Commit(); err != nil {
		return fmt.Errorf("failed to commit rollback %s: %w", version, err)
	}

	logger.Infof(ctx, "Rolled back migration: %s", version)
	return nil
}

// collectUpMigrations returns sorted list of "up" migration file paths under "sql/".
func collectUpMigrations(migrations fs.FS) ([]string, error) {
	var files []string
	err := fs.WalkDir(migrations, "sql", func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if d.IsDir() {
			return nil
		}
		if !strings.HasSuffix(path, ".sql") {
			return nil
		}
		if strings.HasSuffix(path, "_down.sql") {
			return nil
		}
		files = append(files, path)
		return nil
	})
	if err != nil {
		return nil, fmt.Errorf("failed to walk migration files: %w", err)
	}
	sort.Strings(files)
	return files, nil
}

// migrationVersion extracts the version identifier from a migration file path.
// e.g. "sql/0001_initial.sql" -> "0001_initial"
func migrationVersion(path string) string {
	// Remove directory prefix
	base := path
	if idx := strings.LastIndex(path, "/"); idx >= 0 {
		base = path[idx+1:]
	}
	// Remove .sql extension
	return strings.TrimSuffix(base, ".sql")
}

// findDownMigration locates the _down.sql file for a given prefixed version.
// The version has the format "prefix/001_init_schema"; the file is "sql/001_init_schema_down.sql".
func findDownMigration(migrations fs.FS, version string) (string, error) {
	bare := version
	if idx := strings.Index(version, "/"); idx >= 0 {
		bare = version[idx+1:]
	}
	downFile := "sql/" + bare + "_down.sql"
	if _, err := fs.Stat(migrations, downFile); err != nil {
		return "", fmt.Errorf("down migration file not found for version %s: %w", version, err)
	}
	return downFile, nil
}
