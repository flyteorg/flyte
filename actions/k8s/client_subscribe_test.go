package k8s

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/flyteorg/flyte/v2/gen/go/flyteidl2/common"
)

// newSubscribeTestClient builds a minimal ActionsClient exercising only the
// pub/sub machinery (Subscribe/Unsubscribe/notifySubscribers).
func newSubscribeTestClient() *ActionsClient {
	return &ActionsClient{
		bufferSize:  10,
		subscribers: make(map[string]map[chan *ActionUpdate]struct{}),
	}
}

func childUpdate(runName, childName, parentName string) *ActionUpdate {
	return &ActionUpdate{
		ActionID: &common.ActionIdentifier{
			Run:  &common.RunIdentifier{Name: runName},
			Name: childName,
		},
		ParentActionName: parentName,
		Phase:            common.ActionPhase_ACTION_PHASE_SUCCEEDED,
	}
}

// Every run's root action is named "a0", so keying subscriptions on the parent
// action name alone leaks one run's child updates into every other run's watch.
// Subscriptions must be scoped by (run, parent action).
func TestSubscribeIsScopedByRun(t *testing.T) {
	c := newSubscribeTestClient()

	chA := c.Subscribe("runA", "a0")
	chB := c.Subscribe("runB", "a0")

	// An update for runA's child must reach only runA's subscriber.
	c.notifySubscribers(context.Background(), childUpdate("runA", "child1", "a0"))

	select {
	case got := <-chA:
		assert.Equal(t, "runA", got.ActionID.GetRun().GetName())
	default:
		t.Fatal("runA subscriber did not receive its own update")
	}

	select {
	case leaked := <-chB:
		t.Fatalf("runB subscriber leaked a runA update: %+v", leaked)
	default:
		// expected: no cross-run delivery
	}
}

func TestUnsubscribeScopedByRun(t *testing.T) {
	c := newSubscribeTestClient()

	ch := c.Subscribe("runA", "a0")
	c.Unsubscribe("runA", "a0", ch)

	// Unsubscribing the only subscriber should drop the key entirely.
	if _, ok := c.subscribers[subscriberKey("runA", "a0")]; ok {
		t.Fatal("subscriber key not cleaned up after Unsubscribe")
	}

	// Notifying after unsubscribe must not panic or deliver.
	c.notifySubscribers(context.Background(), childUpdate("runA", "child1", "a0"))
}
