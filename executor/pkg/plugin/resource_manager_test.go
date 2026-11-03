package plugin

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"

	pluginsCore "github.com/flyteorg/flyte/v2/flyteplugins/go/tasks/pluginmachinery/core"
)

func TestNoopResourceRegistrar(t *testing.T) {
	r := NewNoopResourceRegistrar()
	assert.NotNil(t, r)
	// Accepts a quota declaration (the connector ships ResourceQuotas{"default":1000}) without error.
	assert.NoError(t, r.RegisterResourceQuota(context.Background(), "default", 1000))
}

func TestNoopResourceManager(t *testing.T) {
	m := NewNoopResourceManager()
	assert.NotNil(t, m)
	assert.Equal(t, "executor-noop-resource-manager", m.GetID())

	// Every allocation is granted, matching v1 propeller with no quota backend configured.
	status, err := m.AllocateResource(context.Background(), "default", "token", pluginsCore.ResourceConstraintsSpec{})
	assert.NoError(t, err)
	assert.Equal(t, pluginsCore.AllocationStatusGranted, status)

	assert.NoError(t, m.ReleaseResource(context.Background(), "default", "token"))
}
