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
		recordedFilter: testFilter(),
		bufferSize:     10,
		subscribers:    make(map[string]map[chan *ActionUpdate]struct{}),
	}
}

func runID(project, runName string) *common.RunIdentifier {
	return &common.RunIdentifier{Project: project, Domain: "development", Name: runName}
}

func childUpdate(run *common.RunIdentifier, childName, parentName string) *ActionUpdate {
	return &ActionUpdate{
		ActionID: &common.ActionIdentifier{
			Run:  run,
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

	chA := c.Subscribe(runID("p", "runA"), "a0")
	chB := c.Subscribe(runID("p", "runB"), "a0")

	// An update for runA's child must reach only runA's subscriber.
	c.notifySubscribers(context.Background(), childUpdate(runID("p", "runA"), "child1", "a0"))

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

	ch := c.Subscribe(runID("p", "runA"), "a0")
	c.Unsubscribe(runID("p", "runA"), "a0", ch)

	// Unsubscribing the only subscriber should drop the key entirely.
	if _, ok := c.subscribers[subscriberKey(runID("p", "runA"), "a0")]; ok {
		t.Fatal("subscriber key not cleaned up after Unsubscribe")
	}

	// Notifying after unsubscribe must not panic or deliver.
	c.notifySubscribers(context.Background(), childUpdate(runID("p", "runA"), "child1", "a0"))
}

// Run names are only unique within a project and domain, so two projects can each
// have a run with the same name. Their watch streams must stay separate.
func TestSubscribeIsScopedByProject(t *testing.T) {
	c := newSubscribeTestClient()

	chA := c.Subscribe(runID("projA", "r1"), "a0")
	chB := c.Subscribe(runID("projB", "r1"), "a0")

	c.notifySubscribers(context.Background(), childUpdate(runID("projA", "r1"), "child1", "a0"))

	select {
	case got := <-chA:
		assert.Equal(t, "projA", got.ActionID.GetRun().GetProject())
	default:
		t.Fatal("projA subscriber did not receive its own update")
	}

	select {
	case leaked := <-chB:
		t.Fatalf("projB subscriber leaked a projA update: %+v", leaked)
	default:
	}
}
