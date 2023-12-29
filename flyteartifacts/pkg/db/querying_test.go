package db

import (
	"context"
	"database/sql"
	"fmt"
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

func TestQuery3_Find(t *testing.T) {
	ctx := context.Background()

	mockDb, mock, rds := getMockRds(t)
	defer func(mockDb *sql.DB) {
		err := mockDb.Close()
		assert.NoError(t, err)
	}(mockDb)

	p := models.PartitionsToIdl(map[string]string{"region": "LAX"})

	s := artifact.SearchArtifactsRequest{
		ArtifactKey: &core.ArtifactKey{
			Domain: "development",
		},
		Partitions: p,
		Principal:  "",
		Version:    "",
		Options:    nil,
	}

	res, ct, err := rds.SearchArtifacts(ctx, s)
	assert.NoError(t, err)
	assert.Equal(t, "", ct)
	assert.Equal(t, nil, res)

	s = artifact.SearchArtifactsRequest{
		ArtifactKey: &core.ArtifactKey{
			Domain: "development",
		},
		Partitions: p,
		Principal:  "",
		Version:    "",
		Options: &artifact.SearchOptions{
			StrictPartitions: true,
		},
	}

	_, ct, err = rds.SearchArtifacts(ctx, s)
	assert.NoError(t, err)
	assert.Equal(t, "", ct)

	s = artifact.SearchArtifactsRequest{
		ArtifactKey: &core.ArtifactKey{
			Domain: "development",
		},
		Principal: "abc",
		Version:   "vxyz",
		Options: &artifact.SearchOptions{
			StrictPartitions: true,
		},
	}

	res, ct, err = rds.SearchArtifacts(ctx, s)
	assert.NoError(t, err)

	fmt.Println(res, ct)

	mock.ExpectClose()
}

func TestQuery4_Find(t *testing.T) {
	ctx := context.Background()

	mockDb, mock, rds := getMockRds(t)
	defer func(mockDb *sql.DB) {
		err := mockDb.Close()
		assert.NoError(t, err)
	}(mockDb)

	s := artifact.SearchArtifactsRequest{
		ArtifactKey: &core.ArtifactKey{
			Domain: "development",
		},
		Principal: "",
		Version:   "",
		Options:   nil,
	}

	res, ct, err := rds.SearchArtifacts(ctx, s)
	assert.NoError(t, err)

	fmt.Println(res, ct)

	mock.ExpectClose()
}

func TestQuery5_Find(t *testing.T) {
	ctx := context.Background()

	mockDb, mock, rds := getMockRds(t)
	defer func(mockDb *sql.DB) {
		err := mockDb.Close()
		assert.NoError(t, err)
	}(mockDb)

	s := artifact.SearchArtifactsRequest{
		ArtifactKey: &core.ArtifactKey{
			Domain: "development",
		},
		Principal: "",
		Version:   "",
		Token:     "1",
		Options: &artifact.SearchOptions{
			LatestByKey: true,
		},
	}

	res, ct, err := rds.SearchArtifacts(ctx, s)
	assert.NoError(t, err)

	fmt.Println(res, ct)

	mock.ExpectClose()
}

func TestLineage1_Input(t *testing.T) {
	ctx := context.Background()

	mockDb, mock, rds := getMockRds(t)
	defer func(mockDb *sql.DB) {
		err := mockDb.Close()
		assert.NoError(t, err)
	}(mockDb)

	r := artifact.FindByWorkflowExecRequest{
		ExecId: &core.WorkflowExecutionIdentifier{
			Project: "pr",
			Domain:  "do",
			Name:    "nnn",
		},
		Direction: 0,
	}

	res, err := rds.FindByWorkflowExec(ctx, &r)
	assert.NoError(t, err)

	fmt.Println(res)

	mock.ExpectClose()
}

func TestLineage2_Output(t *testing.T) {
	ctx := context.Background()

	mockDb, mock, rds := getMockRds(t)
	defer func(mockDb *sql.DB) {
		err := mockDb.Close()
		assert.NoError(t, err)
	}(mockDb)

	r := artifact.FindByWorkflowExecRequest{
		ExecId: &core.WorkflowExecutionIdentifier{
			Project: "pr",
			Domain:  "do",
			Name:    "nnn",
		},
		Direction: 1,
	}

	res, err := rds.FindByWorkflowExec(ctx, &r)
	assert.NoError(t, err)

	fmt.Println(res)

	mock.ExpectClose()
}

func TestLineage3_Set(t *testing.T) {
	ctx := context.Background()

	mockDb, mock, rds := getMockRds(t)
	defer func(mockDb *sql.DB) {
		err := mockDb.Close()
		assert.NoError(t, err)
	}(mockDb)

	ak := core.ArtifactKey{
		Project: "pr",
		Domain:  "do",
		Name:    "nnn",
	}
	ids := []*core.ArtifactID{
		{
			ArtifactKey: &ak,
			Version:     "v1",
		},
		{
			ArtifactKey: &ak,
			Version:     "v2",
		},
	}

	r := artifact.ExecutionInputsRequest{
		ExecutionId: &core.WorkflowExecutionIdentifier{
			Project: "pr",
			Domain:  "do",
			Name:    "nnn",
		},
		Inputs: ids,
	}
	mock.ExpectBegin()
	err := rds.SetExecutionInputs(ctx, &r)
	assert.NoError(t, err)
	mock.ExpectClose()
}

func TestLineagedd3_Set(t *testing.T) {
	mockDb, mock, rds := getMockRds(t)
	defer func(mockDb *sql.DB) {
		err := mockDb.Close()
		assert.NoError(t, err)
	}(mockDb)
	defer func(mockDb *sql.DB) {
		err := mockDb.Close()
		assert.NoError(t, err)
	}(mockDb)

	type Language struct {
		gorm.Model
		Name string //`gorm:"uniqueIndex:idx_lang_name"`
	}
	type User struct {
		gorm.Model
		Name      string
		Languages []Language `gorm:"many2many:user_languages;"`
	}
	var userLanguages []Language
	var dbUser User
	err := rds.db.Model(&dbUser).Where("name = ?", "chinese").Association("Languages").Find(&userLanguages)

	assert.NoError(t, err)
	mock.ExpectClose()
}
