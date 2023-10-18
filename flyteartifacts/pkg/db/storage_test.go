package db

import (
	"context"
	"fmt"
	"testing"

	"github.com/flyteorg/flyte/flytestdlib/config"
	"github.com/flyteorg/flyte/flytestdlib/config/viper"
	"github.com/flyteorg/flyte/flytestdlib/promutils"
	"github.com/golang/protobuf/proto"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/stretchr/testify/assert"

	"github.com/flyteorg/flyte/flyteidl/gen/pb-go/flyteidl/core"
)

func TestWriteOne(t *testing.T) {
	ctx := context.Background()
	configAccessor := viper.NewAccessor(config.Options{
		// TODO: find an idiomatic way to point to this file
		SearchPaths: []string{"/Users/eduardo/repos/flyte/flyteartifacts/sandbox.yaml"},
		StrictMode:  false,
	})
	err := configAccessor.UpdateConfig(ctx)

	fmt.Println("Local integration testing using: ", configAccessor.ConfigFilesUsed())
	scope := promutils.NewTestScope()
	rds := NewStorage(ctx, scope)

	one := uint32(1)
	pval1 := "51"
	p := pgtype.Hstore{
		"area": &pval1,
	}
	//p := postgres.Hstore{"area": &pval1}

	lt := &core.LiteralType{
		Type: &core.LiteralType_Simple{Simple: core.SimpleType_INTEGER},
	}
	lit := &core.Literal{
		Value: &core.Literal_Scalar{
			Scalar: &core.Scalar{
				Value: &core.Scalar_Primitive{
					Primitive: &core.Primitive{
						Value: &core.Primitive_Integer{Integer: 15},
					},
				},
			},
		},
	}

	ltBytes, err := proto.Marshal(lt)
	assert.NoError(t, err)
	litBytes, err := proto.Marshal(lit)
	assert.NoError(t, err)

	gormA := Artifact{
		ArtifactKey: ArtifactKey{
			Project: "demotst",
			Domain:  "unit",
			Name:    "testname 2",
		},
		Version:       "abc123/1/n0/7",
		Partitions:    p,
		LiteralType:   ltBytes,
		LiteralValue:  litBytes,
		ExecutionName: "ddd",
		RetryAttempt:  &one,
	}

	a, err := rds.WriteOne(ctx, gormA)
	fmt.Println(a, err)
}
