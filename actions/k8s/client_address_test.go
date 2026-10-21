package k8s

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"google.golang.org/protobuf/proto"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/watch"
	"sigs.k8s.io/controller-runtime/pkg/client"

	executorv1 "github.com/flyteorg/flyte/v2/executor/api/v1"
	"github.com/flyteorg/flyte/v2/gen/go/flyteidl2/common"
	"github.com/flyteorg/flyte/v2/gen/go/flyteidl2/workflow"
)

// rootActionID addresses the "run1-a0" TaskAction newConditionClient preloads.
var rootActionID = &common.ActionIdentifier{
	Run:  &common.RunIdentifier{Project: "proj", Domain: "dev", Name: "run1"},
	Name: "run1",
}

func getRoot(t *testing.T, c *ActionsClient) *executorv1.TaskAction {
	t.Helper()
	ta := &executorv1.TaskAction{}
	require.NoError(t, c.k8sClient.Get(context.Background(), client.ObjectKey{Name: "run1-a0", Namespace: "flyte"}, ta))
	return ta
}

func running(attempts uint32) *workflow.ActionStatus {
	return &workflow.ActionStatus{Phase: common.ActionPhase_ACTION_PHASE_RUNNING, Attempts: attempts}
}

func sessionAddress(path string) *common.ActionAddress {
	return &common.ActionAddress{
		Pod:     &common.PodAddress{Namespace: "ns", Name: "agent-0", Ip: "10.0.0.7", Port: 8080},
		Path:    path,
		Cluster: "c1",
	}
}

func TestPutStatus_Address(t *testing.T) {
	ctx := context.Background()
	c := newConditionClient(t)
	addr := sessionAddress("/s/run1")

	require.NoError(t, c.PutStatus(ctx, rootActionID, 1, running(1), addr))
	ta := getRoot(t, c)
	require.NotNil(t, ta.Status.Address, "address stored on the CR status")
	assert.Equal(t, "/s/run1", ta.Status.Address.Path)
	assert.Equal(t, "10.0.0.7", ta.Status.Address.Pod.IP)
	assert.True(t, proto.Equal(addr, AddressFromStatus(ta)), "round trips to the proto")

	// No address in the same attempt keeps it.
	require.NoError(t, c.PutStatus(ctx, rootActionID, 1, running(1), nil))
	assert.True(t, proto.Equal(addr, AddressFromStatus(getRoot(t, c))))

	// A new address replaces it.
	other := sessionAddress("/s/run1-b")
	require.NoError(t, c.PutStatus(ctx, rootActionID, 1, running(1), other))
	assert.True(t, proto.Equal(other, AddressFromStatus(getRoot(t, c))))

	// A new attempt clears it.
	require.NoError(t, c.PutStatus(ctx, rootActionID, 2, running(2), nil))
	assert.Nil(t, AddressFromStatus(getRoot(t, c)))
}

// The address belongs to the attempt and pod that reported it (review of #8201, point 2).
func TestAddress_TiedToTheReportingAttemptAndPod(t *testing.T) {
	ctx := context.Background()

	t.Run("a retry the controller starts hides it", func(t *testing.T) {
		c := newConditionClient(t)
		require.NoError(t, c.PutStatus(ctx, rootActionID, 1, running(1), sessionAddress("/s/attempt1")))
		ta := getRoot(t, c)
		ta.Status.Attempts = 2 // the controller's in-place or max-runtime restart
		require.NoError(t, c.k8sClient.Status().Update(ctx, ta))
		// The worker's next update for the new attempt, with no address yet.
		require.NoError(t, c.PutStatus(ctx, rootActionID, 2, running(2), nil))
		assert.Nil(t, AddressFromStatus(getRoot(t, c)), "attempt 1's address must not reach watchers")
	})

	t.Run("a late report from an older attempt is ignored", func(t *testing.T) {
		c := newConditionClient(t)
		require.NoError(t, c.PutStatus(ctx, rootActionID, 2, running(2), sessionAddress("/s/attempt2")))
		require.NoError(t, c.PutStatus(ctx, rootActionID, 1, running(1), sessionAddress("/s/attempt1")))
		ta := getRoot(t, c)
		assert.Equal(t, uint32(2), ta.Status.Attempts, "a stale report does not move attempts back")
		assert.Equal(t, "/s/attempt2", AddressFromStatus(ta).GetPath())
	})

	t.Run("a system-retry relaunch hides it", func(t *testing.T) {
		c := newConditionClient(t)
		require.NoError(t, c.PutStatus(ctx, rootActionID, 1, running(1), sessionAddress("/s/pod1")))
		ta := getRoot(t, c)
		ta.Status.SystemRetries++ // the controller relaunches the pod in place
		require.NoError(t, c.k8sClient.Status().Update(ctx, ta))
		assert.Nil(t, AddressFromStatus(getRoot(t, c)), "the old pod's address must not reach watchers")
		// The relaunched pod reports its own.
		require.NoError(t, c.PutStatus(ctx, rootActionID, 1, running(1), sessionAddress("/s/pod2")))
		assert.Equal(t, "/s/pod2", AddressFromStatus(getRoot(t, c)).GetPath())
	})
}

// An empty address withdraws the held one (review of #8201, point 4).
func TestPutStatus_EmptyAddressWithdraws(t *testing.T) {
	ctx := context.Background()
	c := newConditionClient(t)
	require.NoError(t, c.PutStatus(ctx, rootActionID, 1, running(1), sessionAddress("/s/x")))
	require.NoError(t, c.PutStatus(ctx, rootActionID, 1, running(1), &common.ActionAddress{}))
	ta := getRoot(t, c)
	assert.Nil(t, ta.Status.Address)
	assert.Nil(t, AddressFromStatus(ta))
}

func TestBuildActionUpdate_Address(t *testing.T) {
	ctx := context.Background()
	ta, _ := newTestActionUpdate("session1")
	assert.Nil(t, buildActionUpdate(ctx, ta, watch.Modified).Address)

	ta.Status.Address = &executorv1.ActionAddress{
		Pod:     &executorv1.PodAddress{Namespace: "ns", Name: "agent-0", IP: "10.0.0.7", Port: 8080},
		Path:    "/s/session1",
		Cluster: "c1",
	}
	ta.Status.Conditions = []metav1.Condition{{
		Type: string(executorv1.ConditionTypeProgressing), Status: "True", Reason: string(executorv1.ConditionReasonExecuting),
	}}
	update := buildActionUpdate(ctx, ta, watch.Modified)
	require.NotNil(t, update.Address, "a running task's address reaches watchers")
	assert.Equal(t, "/s/session1", update.Address.GetPath())
	assert.Equal(t, "10.0.0.7", update.Address.GetPod().GetIp())
	assert.Equal(t, "c1", update.Address.GetCluster())
	assert.Nil(t, update.SignalValue)

	// Once terminal, the address is no longer delivered.
	ta.Status.Conditions = []metav1.Condition{{Type: string(executorv1.ConditionTypeSucceeded), Status: "True"}}
	assert.Nil(t, buildActionUpdate(ctx, ta, watch.Modified).Address)
}

func TestNextAddress(t *testing.T) {
	held := &executorv1.ActionAddress{Path: "/a"}
	assert.Equal(t, "/b", nextAddress(held, 1, 1, &common.ActionAddress{Path: "/b"}).Path, "replaced")
	assert.Same(t, held, nextAddress(held, 1, 1, nil), "kept within the attempt")
	assert.Nil(t, nextAddress(held, 1, 2, nil), "cleared by a new attempt")
}
