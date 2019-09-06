package subworkflow

import (
	"testing"

	"github.com/lyft/flyteidl/gen/pb-go/flyteidl/core"
	"github.com/stretchr/testify/assert"
)

func TestGetChildWorkflowExecutionID(t *testing.T) {
	id, err := GetChildWorkflowExecutionID(
		&core.WorkflowExecutionIdentifier{
			Project: "project",
			Domain:  "domain",
			Name:    "first-name-is-pretty-large",
		}, "hello-world", 1)
	assert.Equal(t, id.Name, "fav2uxxi")
	assert.NoError(t, err)
}
