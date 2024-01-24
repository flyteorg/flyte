package impl

import (
	"context"
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/flyteorg/flyte/flyteadmin/pkg/artifacts"
	eventWriterMocks "github.com/flyteorg/flyte/flyteadmin/pkg/async/events/mocks"
	"github.com/flyteorg/flyte/flyteidl/gen/pb-go/flyteidl/core"
	mockScope "github.com/flyteorg/flyte/flytestdlib/promutils"
)

func TestResolveNotWorking(t *testing.T) {
	mockConfig := getMockExecutionsConfigProvider()

	execManager := NewExecutionManager(nil, nil, mockConfig, nil, mockScope.NewTestScope(), mockScope.NewTestScope(), nil, nil, nil, nil, nil, nil, &eventWriterMocks.WorkflowExecutionEventWriter{}, artifacts.NewArtifactRegistry(context.Background(), nil)).(*ExecutionManager)

	pm, artifactIDs, err := execManager.ResolveParameterMapArtifacts(context.Background(), nil, nil)
	assert.Nil(t, err)
	fmt.Println(pm, artifactIDs)

}

func TestTrackingBitExtract(t *testing.T) {
	mockConfig := getMockExecutionsConfigProvider()

	execManager := NewExecutionManager(nil, nil, mockConfig, nil, mockScope.NewTestScope(), mockScope.NewTestScope(), nil, nil, nil, nil, nil, nil, &eventWriterMocks.WorkflowExecutionEventWriter{}, artifacts.NewArtifactRegistry(context.Background(), nil)).(*ExecutionManager)

	lit := core.Literal{
		Value: &core.Literal_Scalar{
			Scalar: &core.Scalar{
				Value: &core.Scalar_Primitive{
					Primitive: &core.Primitive{
						Value: &core.Primitive_Integer{
							Integer: 1,
						},
					},
				},
			},
		},
		Metadata: map[string]string{"_ua": "proj/domain/name@version"},
	}
	inputMap := core.LiteralMap{
		Literals: map[string]*core.Literal{
			"a": &lit,
		},
	}
	inputColl := core.LiteralCollection{
		Literals: []*core.Literal{
			&lit,
		},
	}

	var trackers = make(map[string]string)
	execManager.ExtractArtifactTrackers(trackers, &lit)
	assert.Equal(t, 1, len(trackers))

	trackers = make(map[string]string)
	execManager.ExtractArtifactTrackers(trackers, &core.Literal{Value: &core.Literal_Map{Map: &inputMap}})
	assert.Equal(t, 1, len(trackers))

	trackers = make(map[string]string)
	execManager.ExtractArtifactTrackers(trackers, &core.Literal{Value: &core.Literal_Collection{Collection: &inputColl}})
	assert.Equal(t, 1, len(trackers))
	assert.Equal(t, "", trackers["proj/domain/name@version"])
}
