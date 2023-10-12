package shared

import (
	stdlibLogger "github.com/flyteorg/flyte/flytestdlib/logger"
	"github.com/flyteorg/flyte/flytestdlib/promutils"
	"github.com/go-gormigrate/gormigrate/v2"
	"google.golang.org/grpc/codes"
)

func PGMigrate(cfg *DbConfig, logCfg *stdlibLogger.Config, _ promutils.Scope, migrations []*gormigrate.Migration) error {
	db, err := connection.OpenDbConnection(connection.NewPostgresDialector(cfg), cfg, logCfg)
	if err != nil {
		return serverErr.NewServerErrorf(codes.InvalidArgument, "Failed to open postgres connection to [%+s] with err: %+v", cfg.DbName, err)
	}

	if err := sharedmigrations.CreateDB(db, cfg.DbName); err != nil {
		return serverErr.NewServerErrorf(codes.Internal, "Failed to ensure db [%s] exists wth err: %v", cfg.DbName, err)
	}

	if len(migrations) > 0 {
		migrator := gormigrate.New(db, gormigrate.DefaultOptions, migrations)
		err = migrator.Migrate()
		if err != nil {
			return serverErr.NewServerErrorf(codes.Internal, "Failed to migrate database: %v", err)
		}
	}
	return nil
}
