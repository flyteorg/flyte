package k8s

import (
	"testing"

	"github.com/flyteorg/flyteidl/gen/pb-go/flyteidl/core"
	"github.com/stretchr/testify/assert"
)

func TestComputeRetryStrategy(t *testing.T) {

	tests := []struct {
		name            string
		nodeRetries     int
		taskRetries     int
		expectedRetries int
	}{
		{"node-only", 1, 0, 2},
		{"task-only", 0, 1, 2},
		{"node-task", 2, 3, 3},
		{"no-retries", 0, 0, 0},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {

			var node *core.Node
			if test.nodeRetries != 0 {
				node = &core.Node{
					Metadata: &core.NodeMetadata{
						Retries: &core.RetryStrategy{
							Retries: uint32(test.nodeRetries),
						},
					},
				}
			}

			var tmpl *core.TaskTemplate
			if test.taskRetries != 0 {
				tmpl = &core.TaskTemplate{
					Metadata: &core.TaskMetadata{
						Retries: &core.RetryStrategy{
							Retries: uint32(test.taskRetries),
						},
					},
				}
			}

			r := computeRetryStrategy(node, tmpl)
			if test.expectedRetries != 0 {
				assert.NotNil(t, r)
				assert.Equal(t, test.expectedRetries, *r.MinAttempts)
			} else {
				assert.Nil(t, r)
			}
		})
	}

}
