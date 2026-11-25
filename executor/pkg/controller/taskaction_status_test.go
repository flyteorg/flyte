package controller

import (
	"testing"

	"github.com/stretchr/testify/assert"

	flyteorgv1 "github.com/flyteorg/flyte/v2/executor/api/v1"
)

// The reconciler's status write keeps an address the actions service stored
// after the reconcile read the TaskAction (review of #8201, point 1).
func TestApplyReconciledStatus_KeepsTheActionsServiceAddress(t *testing.T) {
	// What the reconcile read: no address yet.
	reconciled := &flyteorgv1.TaskAction{}
	reconciled.Status.Attempts = 1
	reconciled.Status.PluginPhaseVersion = 7

	// Meanwhile PutStatus stored one.
	latest := &flyteorgv1.TaskAction{}
	latest.Status.Attempts = 1
	latest.Status.Address = &flyteorgv1.ActionAddress{Path: "/s/session", Attempt: 1}

	applyReconciledStatus(latest, reconciled)
	assert.Equal(t, uint32(7), latest.Status.PluginPhaseVersion, "the reconciler's fields are written")
	if assert.NotNil(t, latest.Status.Address, "the address survives the write") {
		assert.Equal(t, "/s/session", latest.Status.Address.Path)
	}
}
