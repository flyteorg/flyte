package implementations

import (
	"context"
	"github.com/flyteorg/flyteidl/gen/pb-go/flyteidl/admin"
	"testing"

	runtimeInterfaces "github.com/flyteorg/flyteadmin/pkg/runtime/interfaces"
	"github.com/flyteorg/flytestdlib/promutils"
	"github.com/stretchr/testify/assert"
)

func TestSlackWebhook(t *testing.T) {
	webhook := NewSlackWebhook(runtimeInterfaces.WebHookConfig{}, promutils.NewTestScope())
	err := webhook.Post(context.Background(), admin.WebhookPayload{Message: "hello world"})
	assert.Nil(t, err)
}
