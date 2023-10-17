package db

import (
	"context"
	"fmt"
	"github.com/flyteorg/flyte/flytestdlib/config"
	"github.com/flyteorg/flyte/flytestdlib/config/viper"
	"github.com/flyteorg/flyte/flytestdlib/promutils"
	"github.com/golang/protobuf/proto"
	"github.com/jinzhu/gorm/dialects/postgres"
	"github.com/stretchr/testify/assert"
	"testing"

	"github.com/flyteorg/flyte/flyteidl/gen/pb-go/flyteidl/core"
)

func TestWriteOne(t *testing.T) {
	ctx := context.Background()
	configAccessor := viper.NewAccessor(config.Options{
		SearchPaths: []string{"/Users/ytong/go/src/github.com/flyteorg/flyte/flyteartifacts/sandbox.yaml"},
		StrictMode:  false,
	})
	err := configAccessor.UpdateConfig(ctx)

	fmt.Println("Local integration testing using: ", configAccessor.ConfigFilesUsed())
	scope := promutils.NewTestScope()
	rds := NewStorage(ctx, scope)

	one := uint32(1)
	pval1 := "f51f"
	p := postgres.Hstore{"area": &pval1}

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
			Name:    "testname 1",
		},
		Version:       "abc123/1/n0/7",
		Partitions:    &p,
		LiteralType:   ltBytes,
		LiteralValue:  litBytes,
		ExecutionName: "ddd",
		RetryAttempt:  &one,
	}

	a, err := rds.WriteOne(ctx, gormA)
	fmt.Println(a, err)
}
