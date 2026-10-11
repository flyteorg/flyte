package main

import (
	"context"
	"encoding/json"
	"errors"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/flyteorg/flyte/v2/flytestdlib/app"
	"github.com/flyteorg/flyte/v2/flytestdlib/database"
	"github.com/flyteorg/flyte/v2/flytestdlib/logger"
	_ "github.com/flyteorg/flyte/v2/runs/config"
	"github.com/flyteorg/flyte/v2/runs/retention"
	"github.com/jmoiron/sqlx"
	"github.com/spf13/cobra"
)

func command() *cobra.Command {
	// Reuse config initialization while replacing the server run loop.
	cmd := (&app.App{Name: "runs-retention", Short: "Clean up completed run history"}).Command()
	cmd.Args = cobra.NoArgs
	o := retention.Options{}
	cmd.Flags().DurationVar(&o.Age, "retention-age", 90*24*time.Hour, "Age since run completion")
	cmd.Flags().IntVar(&o.BatchSize, "batch-size", 100, "Candidate runs per page")
	cmd.Flags().DurationVar(&o.MaxRuntime, "max-runtime", 20*time.Minute, "Maximum cleanup runtime")
	cmd.Flags().BoolVar(&o.DryRun, "dry-run", true, "Report eligible rows without deleting them")
	cmd.RunE = func(cmd *cobra.Command, _ []string) error {
		if err := o.Validate(); err != nil {
			return err
		}
		if err := logger.SetConfig(logger.GetConfig()); err != nil {
			return err
		}
		cfg := database.GetConfig()
		if cfg.Postgres.IsEmpty() {
			return errors.New("PostgreSQL database configuration is required")
		}
		// One overall deadline covers connection setup and cleanup.
		ctx, cancel := context.WithTimeout(cmd.Context(), o.MaxRuntime)
		defer cancel()
		// Connect to the existing primary; maintenance must never create a DB.
		db, err := sqlx.ConnectContext(ctx, "pgx", database.GetPostgresDsn(ctx, cfg.Postgres))
		if err != nil {
			return err
		}
		defer db.Close()
		db.SetMaxOpenConns(cfg.MaxOpenConnections)
		db.SetMaxIdleConns(cfg.MaxIdleConnections)
		db.SetConnMaxLifetime(cfg.ConnMaxLifeTime.Duration)
		result, err := retention.Run(ctx, db, o)
		// Emit committed progress even on failure.
		if encodeErr := json.NewEncoder(cmd.OutOrStdout()).Encode(result); err == nil {
			err = encodeErr
		}
		return err
	}
	return cmd
}

func main() {
	ctx, cancel := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer cancel()
	if err := command().ExecuteContext(ctx); err != nil {
		os.Exit(1)
	}
}
