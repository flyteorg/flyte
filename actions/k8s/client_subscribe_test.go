package k8s

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"

	"github.com/flyteorg/flyte/v2/gen/go/flyteidl2/common"
)

// newSubscribeTestClient builds a minimal ActionsClient exercising only the
// pub/sub machinery (Subscribe/Unsubscribe/notifySubscribers).
func newSubscribeTestClient() *ActionsClient {
	return &ActionsClient{
		recordedFilter: testFilter(),
		bufferSize:     10,
		subscribers:    make(map[string]map[chan *ActionUpdate]*subscription),
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

	got := receive(t, chA)
	assert.Equal(t, "runA", got.ActionID.GetRun().GetName())

	select {
	case leaked := <-chB:
		t.Fatalf("runB subscriber leaked a runA update: %+v", leaked)
	case <-time.After(50 * time.Millisecond):
		// expected: no cross-run delivery
	}
}

// A burst larger than the channel buffer must still deliver every child's update.
func TestNotifySubscribersNeverDrops(t *testing.T) {
	c := newSubscribeTestClient()
	ch := c.Subscribe("runA", "a0")

	const children = 1000 // 100x the test client's bufferSize of 10
	for i := range children {
		c.notifySubscribers(context.Background(), childUpdate("runA", fmt.Sprintf("child%d", i), "a0"))
	}

	seen := make(map[string]bool, children)
	for range children {
		seen[receive(t, ch).ActionID.GetName()] = true
	}
	assert.Len(t, seen, children)
}

// A newer update for a parked child must replace it and keep its place in the order.
func TestNotifySubscribersCoalescesPerChild(t *testing.T) {
	sub := &subscription{
		out:     make(chan *ActionUpdate),
		pending: make(map[string]*ActionUpdate),
		wake:    make(chan struct{}, 1),
		done:    make(chan struct{}),
	}
	running := childUpdate("runA", "child1", "a0")
	running.Phase = common.ActionPhase_ACTION_PHASE_RUNNING
	sub.notify(running)
	sub.notify(childUpdate("runA", "child2", "a0"))
	sub.notify(childUpdate("runA", "child1", "a0"))

	// No pump is running, so all three updates are still parked.
	first, second := sub.next(), sub.next()
	assert.Nil(t, sub.next())
	assert.Equal(t, "child1", first.ActionID.GetName())
	assert.Equal(t, common.ActionPhase_ACTION_PHASE_SUCCEEDED, first.Phase)
	assert.Equal(t, "child2", second.ActionID.GetName())
}

func TestUnsubscribeClosesChannel(t *testing.T) {
	c := newSubscribeTestClient()
	ch := c.Subscribe("runA", "a0")
	c.notifySubscribers(context.Background(), childUpdate("runA", "child1", "a0"))
	c.Unsubscribe("runA", "a0", ch)

	// The pump may deliver updates it already held, but must then close the channel.
	timeout := time.After(time.Second)
	for {
		select {
		case _, ok := <-ch:
			if !ok {
				return
			}
		case <-timeout:
			t.Fatal("channel not closed after Unsubscribe")
		}
	}
}

func receive(t *testing.T, ch chan *ActionUpdate) *ActionUpdate {
	t.Helper()
	select {
	case update := <-ch:
		return update
	case <-time.After(time.Second):
		t.Fatal("subscriber did not receive an update")
		return nil
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
