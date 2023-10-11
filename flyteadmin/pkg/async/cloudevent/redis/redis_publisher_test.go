package redis

import (
	"context"
	"github.com/flyteorg/flyteidl/gen/pb-go/flyteidl/core"
	"github.com/flyteorg/flyteidl/gen/pb-go/flyteidl/event"
	"github.com/golang/protobuf/proto"
	"github.com/stretchr/testify/assert"
	"testing"

	redisAPI "github.com/redis/go-redis/v9"
)

func TestRjkl(t *testing.T) {
	ctx := context.TODO()

	client := redisAPI.NewClient(&redisAPI.Options{
		Addr: "localhost:6379",
	})

	e := event.WorkflowExecutionEvent{
		ExecutionId: &core.WorkflowExecutionIdentifier{
			Project: "project",
			Domain:  "domain",
			Name:    "name",
		},
		ProducerId: "",
		Phase:      0,
	}
	ee, _ := proto.Marshal(&e)
	r := client.Publish(ctx, "channel007", ee)
	assert.NoError(t, r.Err())
}
