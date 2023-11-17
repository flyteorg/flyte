package db

import (
	"context"
	"database/sql"
	"github.com/DATA-DOG/go-sqlmock"
	"github.com/flyteorg/flyte/flyteartifacts/pkg/models"
	"github.com/flyteorg/flyte/flyteidl/gen/pb-go/flyteidl/artifact"
	"github.com/flyteorg/flyte/flyteidl/gen/pb-go/flyteidl/core"
	"github.com/flyteorg/flyte/flytestdlib/database"
	"github.com/flyteorg/flyte/flytestdlib/promutils"
	"github.com/stretchr/testify/assert"
	"testing"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

func getMockRds(t *testing.T) (*sql.DB, sqlmock.Sqlmock, RDSStorage) {
	dbCfg := database.DbConfig{}

	mockDb, mock, err := sqlmock.New()
	assert.NoError(t, err)
	dialector := postgres.New(postgres.Config{
		Conn:       mockDb,
		DriverName: "postgres",
	})
	db, _ := gorm.Open(dialector, &gorm.Config{})

	scope := promutils.NewTestScope()
	rds := RDSStorage{
		config:  dbCfg,
		db:      db,
		metrics: newMetrics(scope),
	}
	return mockDb, mock, rds
}

func TestQuery1(t *testing.T) {
	ctx := context.Background()

	mockDb, mock, rds := getMockRds(t)
	defer func(mockDb *sql.DB) {
		err := mockDb.Close()
		assert.NoError(t, err)
	}(mockDb)

	//rows := sqlmock.NewRows([]string{"Code", "Price"}).AddRow("D43", 100)
	//mock.ExpectQuery(`SELECT`).WillReturnRows(rows)

	query := core.ArtifactQuery{
		Identifier: &core.ArtifactQuery_ArtifactId{
			ArtifactId: &core.ArtifactID{
				ArtifactKey: &core.ArtifactKey{
					Project: "pp",
					Domain:  "dd",
					Name:    "nn",
				},
				Version:    "abc",
				Dimensions: nil,
			},
		},
	}
	_, err := rds.GetArtifact(ctx, query)
	assert.NoError(t, err)

	//mock.ExpectQuery("SELECT")
	mock.ExpectClose()
}

func TestQuery2_Create(t *testing.T) {
	ctx := context.Background()

	mockDb, mock, rds := getMockRds(t)
	defer func(mockDb *sql.DB) {
		err := mockDb.Close()
		assert.NoError(t, err)
	}(mockDb)

	a := artifact.Artifact{
		ArtifactId: &core.ArtifactID{
			ArtifactKey: &core.ArtifactKey{
				Project: "pp",
				Domain:  "dd",
				Name:    "nn",
			},
			Version:    "abc",
			Dimensions: nil,
		},
		Spec: &artifact.ArtifactSpec{
			Value: nil,
			Type:  nil,
		},
		Source: &artifact.ArtifactSource{
			WorkflowExecution: &core.WorkflowExecutionIdentifier{
				Project: "pp",
				Domain:  "dd",
				Name:    "artifact_name",
			},
			NodeId:    "node1",
			Principal: "foo",
		},
	}

	mock.ExpectBegin()
	_, err := rds.CreateArtifact(ctx, models.Artifact{
		Artifact:          a,
		LiteralTypeBytes:  []byte("hello"),
		LiteralValueBytes: []byte("world"),
	})
	assert.NoError(t, err)

	//mock.ExpectQuery("SELECT")
	mock.ExpectClose()
}
