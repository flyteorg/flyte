package athena

import (
	"testing"

	awsSdk "github.com/aws/aws-sdk-go-v2/aws"
	"github.com/stretchr/testify/assert"

	idlCore "github.com/flyteorg/flyte/v2/gen/go/flyteidl2/core"
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

func TestCreateTaskInfoGovAWS(t *testing.T) {
	taskInfo := createTaskInfo("query_id", awsSdk.Config{
		Region: "us-gov-east-1",
	})
	assert.EqualValues(t, []*idlCore.TaskLog{
		{
			Uri:  "https://us-gov-east-1.console.amazonaws-us-gov.com/athena/home?force&region=us-gov-east-1#query/history/query_id",
			Name: "Athena Query Console",
		},
	}, taskInfo.Logs)
	assert.Len(t, taskInfo.ExternalResources, 1)
	assert.Equal(t, taskInfo.ExternalResources[0].ExternalID, "query_id")
}
