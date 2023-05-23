package implementations

import (
	"context"
	"testing"

	runtimeInterfaces "github.com/flyteorg/flyteadmin/pkg/runtime/interfaces"
	"github.com/flyteorg/flytestdlib/promutils"
	"github.com/stretchr/testify/assert"
)

func TestSlackWebhook(t *testing.T) {
	webhook := NewSlackWebhook(runtimeInterfaces.WebhooksConfig{}, promutils.NewTestScope())
	err := webhook.Post(context.Background(), "message")
	assert.Nil(t, err)
}
