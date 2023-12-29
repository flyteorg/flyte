//go:build local_integration

// Eduardo - add this build tag to run this test

package db

import (
	"context"
	"fmt"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/flyteorg/flyte/flyteartifacts/pkg/models"
	"github.com/flyteorg/flyte/flyteidl/gen/pb-go/flyteidl/artifact"
	"github.com/flyteorg/flyte/flyteidl/gen/pb-go/flyteidl/core"
	"github.com/flyteorg/flyte/flytestdlib/config"
	"github.com/flyteorg/flyte/flytestdlib/config/viper"
	"github.com/flyteorg/flyte/flytestdlib/promutils"
)

func TestBasicWrite(t *testing.T) {
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

	ak := &core.ArtifactKey{
		Project: "demotst",
		Domain:  "unit",
		Name:    "artfname 10",
	}
	source := &artifact.ArtifactSource{
		WorkflowExecution: &core.WorkflowExecutionIdentifier{
			Project: "demotst",
			Domain:  "unit",
			Name:    "exectest1",
		},
		Principal: "userone",
		NodeId:    "testnodeid",
		TaskId: &core.Identifier{
			ResourceType: core.ResourceType_TASK,
			Project:      "demotst",
			Domain:       "unit",
			Name:         "testtaskname 2",
			Version:      "testtaskversion",
		},
	}
	spec := &artifact.ArtifactSpec{
		Value:            lit,
		Type:             lt,
		ShortDescription: "",
		UserMetadata:     nil,
		MetadataType:     "",
	}
	partitions := map[string]string{
		"area": "51",
		"ds":   "2023-05-01",
	}

	// Create one
	am, err := models.CreateArtifactModelFromRequest(ctx, ak, spec, "abc123/1/n0/1", partitions, "tag", source)
	assert.NoError(t, err)

	newModel, err := rds.CreateArtifact(ctx, am)
	assert.NoError(t, err)
	fmt.Println(newModel)

	// Create another
	am, err = models.CreateArtifactModelFromRequest(ctx, ak, spec, "abc123/1/n0/2", partitions, "tag", source)
	assert.NoError(t, err)

	newModel, err = rds.CreateArtifact(ctx, am)
	assert.NoError(t, err)
	fmt.Println(newModel)
}

func TestBasicRead(t *testing.T) {
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
	ak := &core.ArtifactKey{
		Project: "demotst",
		Domain:  "unit",
		Name:    "artfname 10",
	}
	partitions := map[string]string{
		"area": "51",
		"ds":   "2023-05-01",
	}
	query := core.ArtifactQuery{
		Identifier: &core.ArtifactQuery_ArtifactId{
			ArtifactId: &core.ArtifactID{
				ArtifactKey: ak,
				Version:     "abc123/1/n0/1",
			},
		},
	}
	_, err = rds.GetArtifact(context.Background(), query)
	assert.Error(t, err)

	pidl := models.PartitionsToIdl(partitions)
	query = core.ArtifactQuery{
		Identifier: &core.ArtifactQuery_ArtifactId{
			ArtifactId: &core.ArtifactID{
				ArtifactKey: ak,
				Version:     "abc123/1/n0/1",
				Dimensions: &core.ArtifactID_Partitions{
					Partitions: pidl,
				},
			},
		},
	}
	ga, err := rds.GetArtifact(context.Background(), query)
	assert.NoError(t, err)

	fmt.Println(ga)
}
