package logs

import (
	"testing"

	"github.com/flyteorg/flyteidl/gen/pb-go/flyteidl/core"
	"github.com/stretchr/testify/assert"
)

func TestKubernetesLogMaker(t *testing.T) {
	p := NewKubernetesLogPlugin("https://dashboard.k8s.net")
	tl, err := p.GetTaskLog(
		"flyteexamples-development-task-name",
		"flyteexamples-development",
		"ignore",
		"ignore",
		"main_logs",
	)
	assert.NoError(t, err)
	assert.Equal(t, tl.GetName(), "main_logs")
	assert.Equal(t, tl.GetMessageFormat(), core.TaskLog_UNKNOWN)
	assert.Equal(t, "https://dashboard.k8s.net/#!/log/flyteexamples-development/flyteexamples-development-task-name/pod?namespace=flyteexamples-development", tl.Uri)
}
