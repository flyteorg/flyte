package service

import (
	"context"
	"errors"
	"fmt"
	"testing"

	"connectrpc.com/connect"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"google.golang.org/protobuf/proto"

	"github.com/flyteorg/flyte/v2/gen/go/flyteidl2/actions"
	"github.com/flyteorg/flyte/v2/gen/go/flyteidl2/actions/actionsconnect"
	"github.com/flyteorg/flyte/v2/gen/go/flyteidl2/common"
	"github.com/flyteorg/flyte/v2/gen/go/flyteidl2/task"
	"github.com/flyteorg/flyte/v2/runs/config"
	"github.com/flyteorg/flyte/v2/runs/repository/models"
)

// fakeActions records which cluster got each call.
type fakeActions struct {
	actionsconnect.ActionsServiceClient
	name     string
	calls    *[]string
	abortErr error
}

func (f *fakeActions) Enqueue(context.Context, *connect.Request[actions.EnqueueRequest]) (*connect.Response[actions.EnqueueResponse], error) {
	*f.calls = append(*f.calls, f.name)
	return connect.NewResponse(&actions.EnqueueResponse{}), nil
}

func (f *fakeActions) Abort(context.Context, *connect.Request[actions.AbortRequest]) (*connect.Response[actions.AbortResponse], error) {
	*f.calls = append(*f.calls, f.name)
	if f.abortErr != nil {
		return nil, f.abortErr
	}
	return connect.NewResponse(&actions.AbortResponse{}), nil
}

// fakeRuns stores run specs by run name.
type fakeRuns map[string]*task.RunSpec

func (f fakeRuns) GetRun(_ context.Context, runID *common.RunIdentifier) (*models.Run, error) {
	spec, err := proto.Marshal(f[runID.GetName()])
	return &models.Run{RunSpec: spec}, err
}

var testClusterConfig = &config.MultiClusterConfig{
	HomeCluster: "home",
	QueueClusterMap: map[string][]config.ClusterEntity{
		"a":  {{ID: "a", Weight: 1}},
		"ab": {{ID: "a", Weight: 0.5}, {ID: "b", Weight: 0.5}},
	},
}

func newTestRouter(t *testing.T, runs fakeRuns) (*ActionsRouter, map[string]*fakeActions, *[]string) {
	calls := &[]string{}
	fakes := map[string]*fakeActions{}
	clients := map[string]actionsconnect.ActionsServiceClient{}
	for _, name := range []string{"home", "a", "b"} {
		fakes[name] = &fakeActions{name: name, calls: calls}
		clients[name] = fakes[name]
	}
	r, err := newActionsRouter(context.Background(), testClusterConfig, clients, runs)
	require.NoError(t, err)
	return r, fakes, calls
}

func queued(queue string) *task.RunSpec {
	return &task.RunSpec{Queue: queue}
}

func enqueueRun(r *ActionsRouter, runName string, spec *task.RunSpec) error {
	_, err := r.Enqueue(context.Background(), connect.NewRequest(&actions.EnqueueRequest{
		Action:  &actions.Action{ActionId: &common.ActionIdentifier{Run: &common.RunIdentifier{Name: runName}, Name: runName}},
		RunSpec: spec,
	}))
	return err
}

func TestActionsRouterEnqueue(t *testing.T) {
	for _, tc := range []struct {
		name  string
		queue string
		want  string
	}{
		{name: "no queue goes home", want: "home"},
		{name: "queue picks its cluster", queue: "a", want: "a"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			r, _, calls := newTestRouter(t, nil)
			require.NoError(t, enqueueRun(r, "run1", queued(tc.queue)))
			assert.Equal(t, []string{tc.want}, *calls)
		})
	}

	t.Run("unknown queue fails the run", func(t *testing.T) {
		r, _, calls := newTestRouter(t, nil)
		err := enqueueRun(r, "run1", queued("nope"))
		assert.Equal(t, connect.CodeInvalidArgument, connect.CodeOf(err))
		assert.Empty(t, *calls)
	})
}

func TestActionsRouterWeightedPick(t *testing.T) {
	r, _, calls := newTestRouter(t, nil)
	for i := 0; i < 100; i++ {
		require.NoError(t, enqueueRun(r, fmt.Sprintf("run%d", i), queued("ab")))
	}
	assert.Contains(t, *calls, "a")
	assert.Contains(t, *calls, "b")

	// The same run always lands on the same cluster.
	first, err := r.placement("run7", queued("ab"))
	require.NoError(t, err)
	again, err := r.placement("run7", queued("ab"))
	require.NoError(t, err)
	assert.Equal(t, first, again)
}

func TestActionsRouterAbort(t *testing.T) {
	runs := fakeRuns{"run1": queued("ab")}
	abort := func(r *ActionsRouter) error {
		_, err := r.Abort(context.Background(), connect.NewRequest(&actions.AbortRequest{
			ActionId: &common.ActionIdentifier{Run: &common.RunIdentifier{Name: "run1"}, Name: "run1"},
		}))
		return err
	}

	t.Run("goes to the run's cluster", func(t *testing.T) {
		r, _, calls := newTestRouter(t, runs)
		placed, err := r.placement("run1", queued("ab"))
		require.NoError(t, err)
		require.NoError(t, abort(r))
		assert.Equal(t, placed[:1], *calls)
	})

	t.Run("tries the queue's other cluster when not found", func(t *testing.T) {
		r, fakes, calls := newTestRouter(t, runs)
		placed, err := r.placement("run1", queued("ab"))
		require.NoError(t, err)
		fakes[placed[0]].abortErr = connect.NewError(connect.CodeInternal, errors.New("TaskAction not found"))
		require.NoError(t, abort(r))
		assert.Equal(t, placed, *calls)
	})

	t.Run("returns other errors without trying further", func(t *testing.T) {
		r, fakes, calls := newTestRouter(t, runs)
		placed, err := r.placement("run1", queued("ab"))
		require.NoError(t, err)
		fakes[placed[0]].abortErr = connect.NewError(connect.CodeUnavailable, errors.New("down"))
		assert.Equal(t, connect.CodeUnavailable, connect.CodeOf(abort(r)))
		assert.Equal(t, placed[:1], *calls)
	})
}

func TestNewActionsRouterRejectsUnknownCluster(t *testing.T) {
	cfg := &config.MultiClusterConfig{
		HomeCluster:     "home",
		QueueClusterMap: map[string][]config.ClusterEntity{"x": {{ID: "missing"}}},
	}
	clients := map[string]actionsconnect.ActionsServiceClient{"home": &fakeActions{}}
	_, err := newActionsRouter(context.Background(), cfg, clients, nil)
	assert.Error(t, err)
}
