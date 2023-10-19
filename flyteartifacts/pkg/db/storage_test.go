package db

import (
	"context"
	"fmt"
	"github.com/flyteorg/flyte/flyteartifacts/pkg/models"
	"github.com/flyteorg/flyte/flyteidl/gen/pb-go/flyteidl/artifact"
	"os"
	"testing"

	"github.com/flyteorg/flyte/flytestdlib/config"
	"github.com/flyteorg/flyte/flytestdlib/config/viper"
	"github.com/flyteorg/flyte/flytestdlib/promutils"
	"github.com/stretchr/testify/assert"

	"github.com/flyteorg/flyte/flyteidl/gen/pb-go/flyteidl/core"
)

func TestWriteOne(t *testing.T) {
	ctx := context.Background()
	sandboxCfgFile := os.ExpandEnv("$GOPATH/src/github.com/flyteorg/flyte/flyteartifacts/sandbox.yaml")
	configAccessor := viper.NewAccessor(config.Options{
		SearchPaths: []string{sandboxCfgFile},
		StrictMode:  false,
	})
	err := configAccessor.UpdateConfig(ctx)

	fmt.Println("Local integration testing using: ", configAccessor.ConfigFilesUsed())
	scope := promutils.NewTestScope()
	rds := NewStorage(ctx, scope)

	//one := uint32(1)
	//pval1 := "51"
	//p := pgtype.Hstore{
	//	"area": &pval1,
	//}

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

	//ltBytes, err := proto.Marshal(lt)
	//assert.NoError(t, err)
	//litBytes, err := proto.Marshal(lit)
	//assert.NoError(t, err)

	//gormA := Artifact{
	//	ArtifactKey: ArtifactKey{
	//		Project: "demotst",
	//		Domain:  "unit",
	//		Name:    "testname 2",
	//	},
	//	Version:       "abc123/1/n0/7",
	//	Partitions:    p,
	//	LiteralType:   ltBytes,
	//	LiteralValue:  litBytes,
	//	ExecutionName: "ddd",
	//	RetryAttempt:  &one,
	//}
	ak := &core.ArtifactKey{
		Project: "demotst",
		Domain:  "unit",
		Name:    "testname 2",
	}
	spec := &artifact.ArtifactSpec{
		Value: lit,
		Type:  lt,
		TaskExecution: &core.TaskExecutionIdentifier{
			TaskId: &core.Identifier{
				ResourceType: core.ResourceType_TASK,
				Project:      "demotst",
				Domain:       "unit",
				Name:         "testname 2",
				Version:      "testtaskversion",
			},
			NodeExecutionId: &core.NodeExecutionIdentifier{
				NodeId: "testnodeid",
			},
		},
		Execution: &core.WorkflowExecutionIdentifier{
			Project: "demotst",
			Domain:  "unit",
			Name:    "exectest1",
		},
		Principal:        "userone",
		ShortDescription: "",
		UserMetadata:     nil,
		MetadataType:     "",
	}
	partitions := map[string]string{
		"area": "51",
		"ds":   "2023-05-01",
	}

	am, err := models.CreateArtifactModelFromRequest(ctx, ak, spec, "abc123/1/n0/9", partitions, "tag", "principal")
	assert.NoError(t, err)

	newModel, err := rds.CreateArtifact(ctx, am)
	assert.NoError(t, err)
	fmt.Println(newModel)
}
