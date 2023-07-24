package athena

import (
	"testing"

	awsSdk "github.com/aws/aws-sdk-go-v2/aws"
	idlCore "github.com/flyteorg/flyteidl/gen/pb-go/flyteidl/core"
	"github.com/stretchr/testify/assert"
)

func TestCreateTaskInfo(t *testing.T) {
	taskInfo := createTaskInfo("query_id", awsSdk.Config{
		Region: "us-east-1",
	})
	assert.EqualValues(t, []*idlCore.TaskLog{
		{
			Uri:  "https://us-east-1.console.aws.amazon.com/athena/home?force&region=us-east-1#query/history/query_id",
			Name: "Athena Query Console",
		},
	}, taskInfo.Logs)
	assert.Len(t, taskInfo.ExternalResources, 1)
	assert.Equal(t, taskInfo.ExternalResources[0].ExternalID, "query_id")
}
