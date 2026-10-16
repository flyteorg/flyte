package main

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"

	"github.com/flyteorg/flyte/flyteadmin/pkg/retention"
	"github.com/flyteorg/flyte/flytestdlib/config"
	"github.com/flyteorg/flyte/flytestdlib/config/viper"
	"github.com/flyteorg/flyte/flytestdlib/database"
	"github.com/flyteorg/flyte/flytestdlib/logger"
	"github.com/flyteorg/flyte/flytestdlib/otelutils"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/stdlib"
	"github.com/spf13/cobra"
)

func command() *cobra.Command {
	var configPath string
	o := retention.Options{}
	cmd := &cobra.Command{Use: "flyteadmin-retention", Short: "Clean up completed Flyte execution history", Args: cobra.NoArgs, SilenceUsage: true}
	cmd.PersistentFlags().StringVar(&configPath, "config", "", "Configuration file or glob")
	accessor := viper.NewAccessor(config.Options{StrictMode: false})
	accessor.InitializePflags(cmd.PersistentFlags())
	cmd.Flags().DurationVar(&o.Age, "retention-age", 90*24*time.Hour, "Age since execution completion")
	cmd.Flags().IntVar(&o.BatchSize, "batch-size", 100, "Candidate executions per page")
	cmd.Flags().DurationVar(&o.MaxRuntime, "max-runtime", 20*time.Minute, "Overall runtime limit")
	cmd.Flags().BoolVar(&o.DryRun, "dry-run", true, "Report eligible history without deleting it")
	cmd.PreRunE = func(cmd *cobra.Command, _ []string) error {
		accessor = viper.NewAccessor(config.Options{SearchPaths: []string{configPath, ".", "/etc/flyte/config"}, StrictMode: false})
		accessor.InitializePflags(cmd.Flags())
		return accessor.UpdateConfig(cmd.Context())
	}
	cmd.RunE = func(cmd *cobra.Command, _ []string) (err error) {
		if err = logger.SetConfig(logger.GetConfig()); err != nil {
			return err
		}
		provider, shutdown, err := retention.NewMeterProvider(cmd.Context(), otelutils.GetConfig())
		if err != nil {
			return err
		}
		defer func() {
			ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
			defer cancel()
			if e := shutdown(ctx); e != nil {
				logger.Errorf(ctx, "retention telemetry shutdown: %v", e)
			}
		}()
		metrics, err := retention.NewMetrics(provider)
		if err != nil {
			return err
		}
		started := time.Now()
		result := retention.Result{DryRun: o.DryRun}
		defer func() { metrics.RecordInvocation(cmd.Context(), time.Since(started), result, err) }()
		if err = o.Validate(); err != nil {
			return err
		}
		ctx, cancel := context.WithTimeout(cmd.Context(), o.MaxRuntime)
		defer cancel()
		cfg := database.GetConfig()
		if !cfg.SQLite.IsEmpty() || cfg.Postgres.IsEmpty() {
			return errors.New("retention requires PostgreSQL configuration")
		}
		password := cfg.Postgres.Password
		if cfg.Postgres.PasswordPath != "" {
			value, e := os.ReadFile(cfg.Postgres.PasswordPath)
			if e != nil {
				return e
			}
			password = strings.TrimSpace(string(value))
		}
		// Quote libpq values before parsing so credentials with spaces or
		// quotes are preserved and TLS is configured for the actual host.
		dsn := fmt.Sprintf("host=%s port=%d dbname=%s user=%s password=%s %s",
			quoteDSN(cfg.Postgres.Host), cfg.Postgres.Port, quoteDSN(cfg.Postgres.DbName),
			quoteDSN(cfg.Postgres.User), quoteDSN(password), cfg.Postgres.ExtraOptions)
		pg, err := pgx.ParseConfig(dsn)
		if err != nil {
			return errors.New("invalid PostgreSQL connection configuration")
		}

		db := stdlib.OpenDB(*pg)
		defer db.Close()
		db.SetMaxOpenConns(cfg.MaxOpenConnections)
		db.SetMaxIdleConns(cfg.MaxIdleConnections)
		db.SetConnMaxLifetime(cfg.ConnMaxLifeTime.Duration)
		if err = db.PingContext(ctx); err != nil {
			return err
		}
		result, err = retention.Run(ctx, db, o, metrics)
		if e := json.NewEncoder(cmd.OutOrStdout()).Encode(result); err == nil {
			err = e
		}
		return err
	}
	return cmd
}
func quoteDSN(value string) string {
	return "'" + strings.ReplaceAll(strings.ReplaceAll(value, `\`, `\\`), "'", `\'`) + "'"
}

func main() {
	ctx, cancel := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer cancel()
	if err := command().ExecuteContext(ctx); err != nil {
		os.Exit(1)
	}
}
