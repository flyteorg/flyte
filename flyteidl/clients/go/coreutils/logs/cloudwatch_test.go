package logs

import (
	"testing"

	"github.com/flyteorg/flyteidl/gen/pb-go/flyteidl/core"
	"github.com/stretchr/testify/assert"
)

func TestCloudwatchLogMaker_CriO(t *testing.T) {
	p := NewCloudwatchLogPlugin("us-east-1", "/flyte-production/kubernetes")
	tl, err := p.GetTaskLog(
		"f-uuid-driver",
		"flyteexamples-production",
		"spark-kubernetes-driver",
		"cri-o://abc",
		"main_logs")
	assert.NoError(t, err)
	assert.Equal(t, tl.GetName(), "main_logs")
	assert.Equal(t, tl.GetMessageFormat(), core.TaskLog_JSON)
	assert.Equal(t, "https://console.aws.amazon.com/cloudwatch/home?region=us-east-1#logEventViewer:group=/flyte-production/kubernetes;stream=var.log.containers.f-uuid-driver_flyteexamples-production_spark-kubernetes-driver-abc.log", tl.Uri)
}

func TestCloudwatchLogMaker_Docker(t *testing.T) {
	p := NewCloudwatchLogPlugin("us-east-1", "/flyte-staging/kubernetes")
	tl, err := p.GetTaskLog(
		"f-uuid-driver",
		"flyteexamples-staging",
		"spark-kubernetes-driver",
		"docker://abc123",
		"main_logs")
	assert.NoError(t, err)
	assert.Equal(t, tl.GetName(), "main_logs")
	assert.Equal(t, tl.GetMessageFormat(), core.TaskLog_JSON)
	assert.Equal(t, "https://console.aws.amazon.com/cloudwatch/home?region=us-east-1#logEventViewer:group=/flyte-staging/kubernetes;stream=var.log.containers.f-uuid-driver_flyteexamples-staging_spark-kubernetes-driver-abc123.log", tl.Uri)
}

func TestCloudwatchLogMaker_NoPrefix(t *testing.T) {
	p := NewCloudwatchLogPlugin("us-east-1", "/flyte-staging/kubernetes")
	tl, err := p.GetTaskLog(
		"f-uuid-driver",
		"flyteexamples-staging",
		"spark-kubernetes-driver",
		"123abc",
		"main_logs")
	assert.NoError(t, err)
	assert.Equal(t, tl.GetName(), "main_logs")
	assert.Equal(t, tl.GetMessageFormat(), core.TaskLog_JSON)
	assert.Equal(t, "https://console.aws.amazon.com/cloudwatch/home?region=us-east-1#logEventViewer:group=/flyte-staging/kubernetes;stream=var.log.containers.f-uuid-driver_flyteexamples-staging_spark-kubernetes-driver-123abc.log", tl.Uri)
}
