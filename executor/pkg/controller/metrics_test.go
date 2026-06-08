package controller

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	flyteorgv1 "github.com/flyteorg/flyte/v2/executor/api/v1"
)

func TestRegisterTaskActionMetrics(t *testing.T) {
	// The "executor" meter provider is not registered in unit tests, so this
	// exercises the no-op meter path: registration must still succeed and return
	// usable instruments. The async callback's client is only invoked on
	// collection (never under the no-op provider), so a nil client is safe here.
	m, err := registerTaskActionMetrics(nil)
	require.NoError(t, err)
	require.NotNil(t, m)
}

func TestObserveCRDSize(t *testing.T) {
	m, err := registerTaskActionMetrics(nil)
	require.NoError(t, err)

	// Records without panicking for a populated CRD...
	assert.NotPanics(t, func() {
		m.observeCRDSize(context.Background(), &flyteorgv1.TaskAction{
			Status: flyteorgv1.TaskActionStatus{PluginPhase: "Executing"},
		})
	})

	// ...and is a safe no-op when metrics registration failed (nil receiver).
	var nilMetrics *taskActionMetrics
	assert.NotPanics(t, func() {
		nilMetrics.observeCRDSize(context.Background(), &flyteorgv1.TaskAction{})
	})
}
