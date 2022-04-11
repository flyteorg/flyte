package repositories

import (
	"context"
	"fmt"

	"github.com/flyteorg/flytestdlib/database"

	"github.com/flyteorg/datacatalog/pkg/repositories/config"
	"github.com/flyteorg/datacatalog/pkg/repositories/errors"
	"github.com/flyteorg/datacatalog/pkg/repositories/interfaces"
	"github.com/flyteorg/flytestdlib/promutils"
)

type RepoConfig int32

const (
	POSTGRES RepoConfig = 0
)

var RepositoryConfigurationName = map[RepoConfig]string{
	POSTGRES: "POSTGRES",
}

// The RepositoryInterface indicates the methods that each Repository must support.
// A Repository indicates a Database which is collection of Tables/models.
// The goal is allow databases to be Plugged in easily.
type RepositoryInterface interface {
	DatasetRepo() interfaces.DatasetRepo
	ArtifactRepo() interfaces.ArtifactRepo
	TagRepo() interfaces.TagRepo
	ReservationRepo() interfaces.ReservationRepo
}

func GetRepository(ctx context.Context, repoType RepoConfig, dbConfig database.DbConfig, scope promutils.Scope) RepositoryInterface {
	switch repoType {
	case POSTGRES:
		db, err := config.OpenDbConnection(ctx, config.NewPostgresConfigProvider(dbConfig, scope.NewSubScope("postgres")))
		if err != nil {
			panic(err)
		}
		return NewPostgresRepo(
			db,
			errors.NewPostgresErrorTransformer(),
			scope.NewSubScope("repositories"))
	default:
		panic(fmt.Sprintf("Invalid repoType %v", repoType))
	}
}
