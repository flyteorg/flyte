package implementations

import (
	"context"
	runtimeInterfaces "github.com/flyteorg/flyteadmin/pkg/runtime/interfaces"
	"github.com/flyteorg/flytestdlib/promutils"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestSlackWebhook(t *testing.T) {
	webhook := NewSlackWebhook(runtimeInterfaces.WebhooksConfig{}, promutils.NewTestScope())
	err := webhook.Post(context.Background(), "message")
	assert.Nil(t, err)
}
