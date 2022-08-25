package controller

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/flyteorg/flyteidl/gen/pb-go/flyteidl/admin"
	"github.com/flyteorg/flyteidl/gen/pb-go/flyteidl/core"

	eventErrors "github.com/flyteorg/flytepropeller/events/errors"
	"github.com/flyteorg/flytepropeller/pkg/apis/flyteworkflow/v1alpha1"
	"github.com/flyteorg/flytepropeller/pkg/controller/config"
	workflowErrors "github.com/flyteorg/flytepropeller/pkg/controller/workflow/errors"
	"github.com/flyteorg/flytepropeller/pkg/controller/workflowstore"
	"github.com/flyteorg/flytepropeller/pkg/controller/workflowstore/mocks"

	"github.com/flyteorg/flytestdlib/promutils"
	"github.com/flyteorg/flytestdlib/storage"
	storagemocks "github.com/flyteorg/flytestdlib/storage/mocks"

	"github.com/pkg/errors"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"

	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

type mockExecutor struct {
	HandleCb        func(ctx context.Context, w *v1alpha1.FlyteWorkflow) error
	HandleAbortedCb func(ctx context.Context, w *v1alpha1.FlyteWorkflow, maxRetries uint32) error
}

func (m *mockExecutor) Initialize(ctx context.Context) error {
	return nil
}

func (m *mockExecutor) HandleAbortedWorkflow(ctx context.Context, w *v1alpha1.FlyteWorkflow, maxRetries uint32) error {
	return m.HandleAbortedCb(ctx, w, maxRetries)
}

func (m *mockExecutor) HandleFlyteWorkflow(ctx context.Context, w *v1alpha1.FlyteWorkflow) error {
	return m.HandleCb(ctx, w)
}

func TestPropeller_Handle(t *testing.T) {
	scope := promutils.NewTestScope()
	ctx := context.TODO()
	s := workflowstore.NewInMemoryWorkflowStore()
	exec := &mockExecutor{}
	cfg := &config.Config{
		MaxWorkflowRetries: 0,
	}

	p := NewPropellerHandler(ctx, cfg, nil, s, exec, scope)

	const namespace = "test"
	const name = "123"
	t.Run("notPresent", func(t *testing.T) {
		assert.NoError(t, p.Handle(ctx, namespace, name))
	})

	t.Run("stale", func(t *testing.T) {
		scope := promutils.NewTestScope()
		s := &mocks.FlyteWorkflow{}
		exec := &mockExecutor{}
		p := NewPropellerHandler(ctx, cfg, nil, s, exec, scope)
		s.OnGetMatch(mock.Anything, mock.Anything, mock.Anything).Return(nil, errors.Wrap(workflowstore.ErrStaleWorkflowError, "stale")).Once()
		assert.NoError(t, p.Handle(ctx, namespace, name))
	})

	t.Run("terminated-and-finalized", func(t *testing.T) {
		w := &v1alpha1.FlyteWorkflow{
			ObjectMeta: v1.ObjectMeta{
				Name:      name,
				Namespace: namespace,
			},
			WorkflowSpec: &v1alpha1.WorkflowSpec{
				ID: "w1",
			},
			Status: v1alpha1.WorkflowStatus{
				Phase: v1alpha1.WorkflowPhaseFailed,
			},
		}
		SetCompletedLabel(w, time.Now())
		assert.NoError(t, s.Create(ctx, w))
		assert.NoError(t, p.Handle(ctx, namespace, name))

		r, err := s.Get(ctx, namespace, name)
		assert.NoError(t, err)
		assert.Equal(t, v1alpha1.WorkflowPhaseFailed, r.GetExecutionStatus().GetPhase())
		assert.Equal(t, 0, len(r.Finalizers))
		assert.True(t, HasCompletedLabel(r))
		assert.Equal(t, uint32(0), r.Status.FailedAttempts)
	})

	t.Run("terminated", func(t *testing.T) {
		assert.NoError(t, s.Create(ctx, &v1alpha1.FlyteWorkflow{
			ObjectMeta: v1.ObjectMeta{
				Name:      name,
				Namespace: namespace,
			},
			WorkflowSpec: &v1alpha1.WorkflowSpec{
				ID: "w1",
			},
			Status: v1alpha1.WorkflowStatus{
				Phase: v1alpha1.WorkflowPhaseFailed,
			},
		}))
		assert.NoError(t, p.Handle(ctx, namespace, name))

		r, err := s.Get(ctx, namespace, name)
		assert.NoError(t, err)
		assert.Equal(t, v1alpha1.WorkflowPhaseFailed, r.GetExecutionStatus().GetPhase())
		assert.Equal(t, 0, len(r.Finalizers))
		assert.True(t, HasCompletedLabel(r))
		assert.Equal(t, uint32(0), r.Status.FailedAttempts)
	})

	t.Run("happy", func(t *testing.T) {
		assert.NoError(t, s.Create(ctx, &v1alpha1.FlyteWorkflow{
			ObjectMeta: v1.ObjectMeta{
				Name:      name,
				Namespace: namespace,
			},
			WorkflowSpec: &v1alpha1.WorkflowSpec{
				ID: "w1",
			},
		}))
		exec.HandleCb = func(ctx context.Context, w *v1alpha1.FlyteWorkflow) error {
			w.GetExecutionStatus().UpdatePhase(v1alpha1.WorkflowPhaseSucceeding, "done", nil)
			return nil
		}
		assert.NoError(t, p.Handle(ctx, namespace, name))

		r, err := s.Get(ctx, namespace, name)
		assert.NoError(t, err)
		assert.Equal(t, v1alpha1.WorkflowPhaseSucceeding, r.GetExecutionStatus().GetPhase())
		assert.Equal(t, 1, len(r.Finalizers))
		assert.False(t, HasCompletedLabel(r))
		assert.Equal(t, uint32(0), r.Status.FailedAttempts)
	})

	t.Run("error", func(t *testing.T) {
		assert.NoError(t, s.Create(ctx, &v1alpha1.FlyteWorkflow{
			ObjectMeta: v1.ObjectMeta{
				Name:      name,
				Namespace: namespace,
			},
			WorkflowSpec: &v1alpha1.WorkflowSpec{
				ID: "w1",
			},
		}))
		exec.HandleCb = func(ctx context.Context, w *v1alpha1.FlyteWorkflow) error {
			return fmt.Errorf("failed")
		}
		assert.Error(t, p.Handle(ctx, namespace, name))

		r, err := s.Get(ctx, namespace, name)
		assert.NoError(t, err)
		assert.Equal(t, v1alpha1.WorkflowPhaseReady, r.GetExecutionStatus().GetPhase())
		assert.Equal(t, 0, len(r.Finalizers))
		assert.Equal(t, uint32(1), r.Status.FailedAttempts)
		assert.False(t, HasCompletedLabel(r))
	})

	t.Run("abort", func(t *testing.T) {
		assert.NoError(t, s.Create(ctx, &v1alpha1.FlyteWorkflow{
			ObjectMeta: v1.ObjectMeta{
				Name:      name,
				Namespace: namespace,
			},
			WorkflowSpec: &v1alpha1.WorkflowSpec{
				ID: "w1",
			},
			Status: v1alpha1.WorkflowStatus{
				FailedAttempts: 1,
			},
		}))
		exec.HandleAbortedCb = func(ctx context.Context, w *v1alpha1.FlyteWorkflow, maxRetries uint32) error {
			w.GetExecutionStatus().UpdatePhase(v1alpha1.WorkflowPhaseFailed, "done", nil)
			return nil
		}
		assert.NoError(t, p.Handle(ctx, namespace, name))

		r, err := s.Get(ctx, namespace, name)

		assert.NoError(t, err)
		assert.Equal(t, v1alpha1.WorkflowPhaseFailed, r.GetExecutionStatus().GetPhase())
		assert.Equal(t, 0, len(r.Finalizers))
		assert.True(t, HasCompletedLabel(r))
		assert.Equal(t, uint32(1), r.Status.FailedAttempts)
	})

	t.Run("abort-error", func(t *testing.T) {
		assert.NoError(t, s.Create(ctx, &v1alpha1.FlyteWorkflow{
			ObjectMeta: v1.ObjectMeta{
				Name:      name,
				Namespace: namespace,
			},
			WorkflowSpec: &v1alpha1.WorkflowSpec{
				ID: "w1",
			},
			Status: v1alpha1.WorkflowStatus{
				FailedAttempts: 1,
				Phase:          v1alpha1.WorkflowPhaseRunning,
			},
		}))
		exec.HandleAbortedCb = func(ctx context.Context, w *v1alpha1.FlyteWorkflow, maxRetries uint32) error {
			return fmt.Errorf("abort error")
		}
		assert.Error(t, p.Handle(ctx, namespace, name))
		r, err := s.Get(ctx, namespace, name)
		assert.NoError(t, err)
		assert.Equal(t, v1alpha1.WorkflowPhaseRunning, r.GetExecutionStatus().GetPhase())
		assert.Equal(t, 0, len(r.Finalizers))
		assert.False(t, HasCompletedLabel(r))
		assert.Equal(t, uint32(2), r.Status.FailedAttempts)
	})

	t.Run("abort_panics", func(t *testing.T) {
		assert.NoError(t, s.Create(ctx, &v1alpha1.FlyteWorkflow{
			ObjectMeta: v1.ObjectMeta{
				Name:       name,
				Namespace:  namespace,
				Finalizers: []string{"x"},
			},
			WorkflowSpec: &v1alpha1.WorkflowSpec{
				ID: "w1",
			},
			Status: v1alpha1.WorkflowStatus{
				FailedAttempts: 1,
				Phase:          v1alpha1.WorkflowPhaseRunning,
			},
		}))
		exec.HandleAbortedCb = func(ctx context.Context, w *v1alpha1.FlyteWorkflow, maxRetries uint32) error {
			panic("error")
		}
		assert.Error(t, p.Handle(ctx, namespace, name))

		r, err := s.Get(ctx, namespace, name)
		assert.NoError(t, err)
		assert.Equal(t, v1alpha1.WorkflowPhaseRunning, r.GetExecutionStatus().GetPhase())
		assert.Equal(t, 1, len(r.Finalizers))
		assert.False(t, HasCompletedLabel(r))
		assert.Equal(t, uint32(2), r.Status.FailedAttempts)
	})

	t.Run("noUpdate", func(t *testing.T) {
		assert.NoError(t, s.Create(ctx, &v1alpha1.FlyteWorkflow{
			ObjectMeta: v1.ObjectMeta{
				Name:       name,
				Namespace:  namespace,
				Finalizers: []string{"f1"},
			},
			WorkflowSpec: &v1alpha1.WorkflowSpec{
				ID: "w1",
			},
			Status: v1alpha1.WorkflowStatus{
				Phase: v1alpha1.WorkflowPhaseSucceeding,
			},
		}))
		exec.HandleCb = func(ctx context.Context, w *v1alpha1.FlyteWorkflow) error {
			w.GetExecutionStatus().UpdatePhase(v1alpha1.WorkflowPhaseSucceeding, "", nil)
			return nil
		}
		assert.NoError(t, p.Handle(ctx, namespace, name))

		r, err := s.Get(ctx, namespace, name)
		assert.NoError(t, err)
		assert.Equal(t, v1alpha1.WorkflowPhaseSucceeding, r.GetExecutionStatus().GetPhase())
		assert.False(t, HasCompletedLabel(r))
		assert.Equal(t, 1, len(r.Finalizers))
	})

	t.Run("handlingPanics", func(t *testing.T) {
		assert.NoError(t, s.Create(ctx, &v1alpha1.FlyteWorkflow{
			ObjectMeta: v1.ObjectMeta{
				Name:       name,
				Namespace:  namespace,
				Finalizers: []string{"f1"},
			},
			WorkflowSpec: &v1alpha1.WorkflowSpec{
				ID: "w1",
			},
			Status: v1alpha1.WorkflowStatus{
				Phase: v1alpha1.WorkflowPhaseSucceeding,
			},
		}))
		exec.HandleCb = func(ctx context.Context, w *v1alpha1.FlyteWorkflow) error {
			panic("error")
		}
		assert.Error(t, p.Handle(ctx, namespace, name))

		r, err := s.Get(ctx, namespace, name)
		assert.NoError(t, err)
		assert.Equal(t, v1alpha1.WorkflowPhaseSucceeding, r.GetExecutionStatus().GetPhase())
		assert.False(t, HasCompletedLabel(r))
		assert.Equal(t, 1, len(r.Finalizers))
		assert.Equal(t, uint32(1), r.Status.FailedAttempts)
	})

	t.Run("noUpdate", func(t *testing.T) {
		assert.NoError(t, s.Create(ctx, &v1alpha1.FlyteWorkflow{
			ObjectMeta: v1.ObjectMeta{
				Name:       name,
				Namespace:  namespace,
				Finalizers: []string{"f1"},
			},
			WorkflowSpec: &v1alpha1.WorkflowSpec{
				ID: "w1",
			},
			Status: v1alpha1.WorkflowStatus{
				Phase: v1alpha1.WorkflowPhaseSucceeding,
			},
		}))
		exec.HandleCb = func(ctx context.Context, w *v1alpha1.FlyteWorkflow) error {
			w.GetExecutionStatus().UpdatePhase(v1alpha1.WorkflowPhaseSucceeding, "", nil)
			return nil
		}
		assert.NoError(t, p.Handle(ctx, namespace, name))

		r, err := s.Get(ctx, namespace, name)
		assert.NoError(t, err)
		assert.Equal(t, v1alpha1.WorkflowPhaseSucceeding, r.GetExecutionStatus().GetPhase())
		assert.False(t, HasCompletedLabel(r))
		assert.Equal(t, 1, len(r.Finalizers))
	})

	t.Run("retriesExhaustedFinalize", func(t *testing.T) {
		assert.NoError(t, s.Create(ctx, &v1alpha1.FlyteWorkflow{
			ObjectMeta: v1.ObjectMeta{
				Name:       name,
				Namespace:  namespace,
				Finalizers: []string{"f1"},
			},
			WorkflowSpec: &v1alpha1.WorkflowSpec{
				ID: "w1",
			},
			Status: v1alpha1.WorkflowStatus{
				Phase:          v1alpha1.WorkflowPhaseRunning,
				FailedAttempts: 1,
			},
		}))
		abortCalled := false
		exec.HandleAbortedCb = func(ctx context.Context, w *v1alpha1.FlyteWorkflow, maxRetries uint32) error {
			w.Status.UpdatePhase(v1alpha1.WorkflowPhaseFailed, "Aborted", nil)
			abortCalled = true
			return nil
		}
		assert.NoError(t, p.Handle(ctx, namespace, name))

		r, err := s.Get(ctx, namespace, name)
		assert.NoError(t, err)
		assert.Equal(t, v1alpha1.WorkflowPhaseFailed, r.GetExecutionStatus().GetPhase())
		assert.Equal(t, 0, len(r.Finalizers))
		assert.True(t, HasCompletedLabel(r))
		assert.True(t, abortCalled)
	})

	t.Run("deletedShouldBeFinalized", func(t *testing.T) {
		n := v1.Now()
		assert.NoError(t, s.Create(ctx, &v1alpha1.FlyteWorkflow{
			ObjectMeta: v1.ObjectMeta{
				Name:              name,
				Namespace:         namespace,
				Finalizers:        []string{"f1"},
				DeletionTimestamp: &n,
			},
			WorkflowSpec: &v1alpha1.WorkflowSpec{
				ID: "w1",
			},
			Status: v1alpha1.WorkflowStatus{
				Phase: v1alpha1.WorkflowPhaseSucceeding,
			},
		}))
		exec.HandleAbortedCb = func(ctx context.Context, w *v1alpha1.FlyteWorkflow, maxRetries uint32) error {
			w.Status.UpdatePhase(v1alpha1.WorkflowPhaseAborted, "Aborted", nil)
			return nil
		}
		assert.NoError(t, p.Handle(ctx, namespace, name))

		r, err := s.Get(ctx, namespace, name)
		assert.NoError(t, err)
		assert.Equal(t, v1alpha1.WorkflowPhaseAborted, r.GetExecutionStatus().GetPhase())
		assert.Equal(t, 0, len(r.Finalizers))
		assert.True(t, HasCompletedLabel(r))
	})

	t.Run("deletedButAbortFailed", func(t *testing.T) {
		n := v1.Now()
		assert.NoError(t, s.Create(ctx, &v1alpha1.FlyteWorkflow{
			ObjectMeta: v1.ObjectMeta{
				Name:              name,
				Namespace:         namespace,
				Finalizers:        []string{"f1"},
				DeletionTimestamp: &n,
			},
			WorkflowSpec: &v1alpha1.WorkflowSpec{
				ID: "w1",
			},
			Status: v1alpha1.WorkflowStatus{
				Phase: v1alpha1.WorkflowPhaseSucceeding,
			},
		}))

		exec.HandleAbortedCb = func(ctx context.Context, w *v1alpha1.FlyteWorkflow, maxRetries uint32) error {
			return fmt.Errorf("failed")
		}

		assert.Error(t, p.Handle(ctx, namespace, name))

		r, err := s.Get(ctx, namespace, name)
		assert.NoError(t, err)
		assert.Equal(t, v1alpha1.WorkflowPhaseSucceeding, r.GetExecutionStatus().GetPhase())
		assert.Equal(t, []string{"f1"}, r.Finalizers)
		assert.False(t, HasCompletedLabel(r))
	})

	t.Run("removefinalizerOnTerminateSuccess", func(t *testing.T) {
		assert.NoError(t, s.Create(ctx, &v1alpha1.FlyteWorkflow{
			ObjectMeta: v1.ObjectMeta{
				Name:       name,
				Namespace:  namespace,
				Finalizers: []string{"f1"},
			},
			WorkflowSpec: &v1alpha1.WorkflowSpec{
				ID: "w1",
			},
		}))
		exec.HandleCb = func(ctx context.Context, w *v1alpha1.FlyteWorkflow) error {
			w.GetExecutionStatus().UpdatePhase(v1alpha1.WorkflowPhaseSuccess, "done", nil)
			return nil
		}
		assert.NoError(t, p.Handle(ctx, namespace, name))

		r, err := s.Get(ctx, namespace, name)
		assert.NoError(t, err)
		assert.Equal(t, v1alpha1.WorkflowPhaseSuccess, r.GetExecutionStatus().GetPhase())
		assert.Equal(t, 0, len(r.Finalizers))
		assert.True(t, HasCompletedLabel(r))
	})

	t.Run("removefinalizerOnTerminateFailure", func(t *testing.T) {
		assert.NoError(t, s.Create(ctx, &v1alpha1.FlyteWorkflow{
			ObjectMeta: v1.ObjectMeta{
				Name:       name,
				Namespace:  namespace,
				Finalizers: []string{"f1"},
			},
			WorkflowSpec: &v1alpha1.WorkflowSpec{
				ID: "w1",
			},
		}))
		exec.HandleCb = func(ctx context.Context, w *v1alpha1.FlyteWorkflow) error {
			w.GetExecutionStatus().UpdatePhase(v1alpha1.WorkflowPhaseFailed, "done", &core.ExecutionError{Kind: core.ExecutionError_USER, Code: "code", Message: "message"})
			return nil
		}
		assert.NoError(t, p.Handle(ctx, namespace, name))

		r, err := s.Get(ctx, namespace, name)
		assert.NoError(t, err)
		assert.Equal(t, v1alpha1.WorkflowPhaseFailed, r.GetExecutionStatus().GetPhase())
		assert.Equal(t, 0, len(r.Finalizers))
		assert.True(t, HasCompletedLabel(r))
	})
	t.Run("failOnExecutionNotFoundError", func(t *testing.T) {
		assert.NoError(t, s.Create(ctx, &v1alpha1.FlyteWorkflow{
			ObjectMeta: v1.ObjectMeta{
				Name:      name,
				Namespace: namespace,
			},
			WorkflowSpec: &v1alpha1.WorkflowSpec{
				ID: "w1",
			},
			Status: v1alpha1.WorkflowStatus{
				Phase:          v1alpha1.WorkflowPhaseRunning,
				FailedAttempts: 0,
			},
		}))
		exec.HandleCb = func(ctx context.Context, w *v1alpha1.FlyteWorkflow) error {
			return workflowErrors.Wrapf(workflowErrors.EventRecordingError, "",
				&eventErrors.EventError{
					Code:    eventErrors.ExecutionNotFound,
					Cause:   nil,
					Message: "The execution that the event belongs to does not exist",
				}, "failed to transition phase")
		}
		assert.NoError(t, p.Handle(ctx, namespace, name))

		r, err := s.Get(ctx, namespace, name)
		assert.NoError(t, err)
		assert.Equal(t, v1alpha1.WorkflowPhaseFailing, r.GetExecutionStatus().GetPhase())
		assert.Equal(t, 0, len(r.Finalizers))
		assert.False(t, HasCompletedLabel(r))
	})
	t.Run("failOnIncompatibleClusterError", func(t *testing.T) {
		assert.NoError(t, s.Create(ctx, &v1alpha1.FlyteWorkflow{
			ObjectMeta: v1.ObjectMeta{
				Name:      name,
				Namespace: namespace,
			},
			WorkflowSpec: &v1alpha1.WorkflowSpec{
				ID: "w1",
			},
			Status: v1alpha1.WorkflowStatus{
				Phase:          v1alpha1.WorkflowPhaseRunning,
				FailedAttempts: 0,
			},
		}))
		exec.HandleCb = func(ctx context.Context, w *v1alpha1.FlyteWorkflow) error {
			return workflowErrors.Wrapf(workflowErrors.EventRecordingError, "",
				&eventErrors.EventError{
					Code:    eventErrors.EventIncompatibleCusterError,
					Cause:   nil,
					Message: "The execution is recorded as running on a different cluster",
				}, "failed to transition phase")
		}
		assert.NoError(t, p.Handle(ctx, namespace, name))

		r, err := s.Get(ctx, namespace, name)
		assert.NoError(t, err)
		assert.Equal(t, v1alpha1.WorkflowPhaseFailing, r.GetExecutionStatus().GetPhase())
		assert.Equal(t, 0, len(r.Finalizers))
		assert.False(t, HasCompletedLabel(r))
	})
}

func TestPropeller_Handle_TurboMode(t *testing.T) {
	scope := promutils.NewTestScope()
	ctx := context.TODO()
	s := workflowstore.NewInMemoryWorkflowStore()
	exec := &mockExecutor{}
	cfg := &config.Config{
		MaxWorkflowRetries: 0,
		MaxStreakLength:    5,
	}

	const namespace = "test"
	const name = "123"

	p := NewPropellerHandler(ctx, cfg, nil, s, exec, scope)

	t.Run("error", func(t *testing.T) {
		assert.NoError(t, s.Create(ctx, &v1alpha1.FlyteWorkflow{
			ObjectMeta: v1.ObjectMeta{
				Name:      name,
				Namespace: namespace,
			},
			WorkflowSpec: &v1alpha1.WorkflowSpec{
				ID: "w1",
			},
		}))
		called := false
		exec.HandleCb = func(ctx context.Context, w *v1alpha1.FlyteWorkflow) error {
			if called {
				return fmt.Errorf("already called once")
			}
			called = true
			return fmt.Errorf("failed")
		}
		assert.Error(t, p.Handle(ctx, namespace, name))

		r, err := s.Get(ctx, namespace, name)
		assert.NoError(t, err)
		assert.Equal(t, v1alpha1.WorkflowPhaseReady, r.GetExecutionStatus().GetPhase())
		assert.Equal(t, 0, len(r.Finalizers))
		assert.Equal(t, uint32(1), r.Status.FailedAttempts)
		assert.False(t, HasCompletedLabel(r))
	})

	t.Run("abort", func(t *testing.T) {
		assert.NoError(t, s.Create(ctx, &v1alpha1.FlyteWorkflow{
			ObjectMeta: v1.ObjectMeta{
				Name:      name,
				Namespace: namespace,
			},
			WorkflowSpec: &v1alpha1.WorkflowSpec{
				ID: "w1",
			},
			Status: v1alpha1.WorkflowStatus{
				FailedAttempts: 1,
			},
		}))

		called := false
		exec.HandleAbortedCb = func(ctx context.Context, w *v1alpha1.FlyteWorkflow, maxRetries uint32) error {
			if called {
				return fmt.Errorf("already called once")
			}
			called = true
			w.GetExecutionStatus().UpdatePhase(v1alpha1.WorkflowPhaseFailed, "done", nil)
			return nil
		}
		assert.NoError(t, p.Handle(ctx, namespace, name))

		r, err := s.Get(ctx, namespace, name)

		assert.NoError(t, err)
		assert.Equal(t, v1alpha1.WorkflowPhaseFailed, r.GetExecutionStatus().GetPhase())
		assert.Equal(t, 0, len(r.Finalizers))
		assert.True(t, HasCompletedLabel(r))
		assert.Equal(t, uint32(1), r.Status.FailedAttempts)
	})

	t.Run("abort-error", func(t *testing.T) {
		assert.NoError(t, s.Create(ctx, &v1alpha1.FlyteWorkflow{
			ObjectMeta: v1.ObjectMeta{
				Name:      name,
				Namespace: namespace,
			},
			WorkflowSpec: &v1alpha1.WorkflowSpec{
				ID: "w1",
			},
			Status: v1alpha1.WorkflowStatus{
				FailedAttempts: 1,
				Phase:          v1alpha1.WorkflowPhaseRunning,
			},
		}))

		called := false
		exec.HandleAbortedCb = func(ctx context.Context, w *v1alpha1.FlyteWorkflow, maxRetries uint32) error {
			if called {
				return fmt.Errorf("already called once")
			}
			called = true
			return fmt.Errorf("abort error")
		}
		assert.Error(t, p.Handle(ctx, namespace, name))
		r, err := s.Get(ctx, namespace, name)
		assert.NoError(t, err)
		assert.Equal(t, v1alpha1.WorkflowPhaseRunning, r.GetExecutionStatus().GetPhase())
		assert.Equal(t, 0, len(r.Finalizers))
		assert.False(t, HasCompletedLabel(r))
		assert.Equal(t, uint32(2), r.Status.FailedAttempts)
	})

	t.Run("noUpdate", func(t *testing.T) {
		assert.NoError(t, s.Create(ctx, &v1alpha1.FlyteWorkflow{
			ObjectMeta: v1.ObjectMeta{
				Name:       name,
				Namespace:  namespace,
				Finalizers: []string{"f1"},
			},
			WorkflowSpec: &v1alpha1.WorkflowSpec{
				ID: "w1",
			},
			Status: v1alpha1.WorkflowStatus{
				Phase: v1alpha1.WorkflowPhaseSucceeding,
			},
		}))
		called := false
		exec.HandleCb = func(ctx context.Context, w *v1alpha1.FlyteWorkflow) error {
			if called {
				return fmt.Errorf("already called once")
			}
			called = true
			w.GetExecutionStatus().UpdatePhase(v1alpha1.WorkflowPhaseSucceeding, "", nil)
			return nil
		}
		assert.NoError(t, p.Handle(ctx, namespace, name))

		r, err := s.Get(ctx, namespace, name)
		assert.NoError(t, err)
		assert.Equal(t, v1alpha1.WorkflowPhaseSucceeding, r.GetExecutionStatus().GetPhase())
		assert.False(t, HasCompletedLabel(r))
		assert.Equal(t, 1, len(r.Finalizers))
	})

	t.Run("happy-nochange", func(t *testing.T) {
		assert.NoError(t, s.Create(ctx, &v1alpha1.FlyteWorkflow{
			ObjectMeta: v1.ObjectMeta{
				Name:      name,
				Namespace: namespace,
			},
			WorkflowSpec: &v1alpha1.WorkflowSpec{
				ID: "w1",
			},
		}))
		called := 0
		exec.HandleCb = func(ctx context.Context, w *v1alpha1.FlyteWorkflow) error {
			if called >= 2 {
				return fmt.Errorf("already called once")
			}
			called++
			w.GetExecutionStatus().UpdatePhase(v1alpha1.WorkflowPhaseSucceeding, "done", nil)
			return nil
		}
		assert.NoError(t, p.Handle(ctx, namespace, name))

		r, err := s.Get(ctx, namespace, name)
		assert.NoError(t, err)
		assert.Equal(t, v1alpha1.WorkflowPhaseSucceeding, r.GetExecutionStatus().GetPhase())
		assert.Equal(t, 1, len(r.Finalizers))
		assert.False(t, HasCompletedLabel(r))
		assert.Equal(t, uint32(0), r.Status.FailedAttempts)
	})

	t.Run("happy-success", func(t *testing.T) {
		assert.NoError(t, s.Create(ctx, &v1alpha1.FlyteWorkflow{
			ObjectMeta: v1.ObjectMeta{
				Name:      name,
				Namespace: namespace,
			},
			WorkflowSpec: &v1alpha1.WorkflowSpec{
				ID: "w1",
			},
		}))
		called := 0
		exec.HandleCb = func(ctx context.Context, w *v1alpha1.FlyteWorkflow) error {
			if called >= 2 {
				return fmt.Errorf("already called once")
			}
			if called == 0 {
				w.GetExecutionStatus().UpdatePhase(v1alpha1.WorkflowPhaseSucceeding, "done", nil)
			} else {
				w.GetExecutionStatus().UpdatePhase(v1alpha1.WorkflowPhaseSuccess, "done", nil)
			}
			called++
			return nil
		}
		assert.NoError(t, p.Handle(ctx, namespace, name))

		r, err := s.Get(ctx, namespace, name)
		assert.NoError(t, err)
		assert.Equal(t, v1alpha1.WorkflowPhaseSuccess.String(), r.GetExecutionStatus().GetPhase().String())
		assert.Equal(t, 0, len(r.Finalizers))
		assert.True(t, HasCompletedLabel(r))
		assert.Equal(t, uint32(0), r.Status.FailedAttempts)
		assert.Equal(t, 2, called)
	})

}

func TestPropellerHandler_Initialize(t *testing.T) {
	scope := promutils.NewTestScope()
	ctx := context.TODO()
	s := workflowstore.NewInMemoryWorkflowStore()
	exec := &mockExecutor{}
	cfg := &config.Config{
		MaxWorkflowRetries: 0,
	}

	p := NewPropellerHandler(ctx, cfg, nil, s, exec, scope)

	assert.NoError(t, p.Initialize(ctx))
}

func TestNewPropellerHandler_UpdateFailure(t *testing.T) {
	ctx := context.TODO()
	cfg := &config.Config{
		MaxWorkflowRetries: 0,
	}

	const namespace = "test"
	const name = "123"

	t.Run("unknown error", func(t *testing.T) {
		scope := promutils.NewTestScope()
		s := &mocks.FlyteWorkflow{}
		exec := &mockExecutor{}
		p := NewPropellerHandler(ctx, cfg, nil, s, exec, scope)
		wf := &v1alpha1.FlyteWorkflow{
			ObjectMeta: v1.ObjectMeta{
				Name:      name,
				Namespace: namespace,
			},
			WorkflowSpec: &v1alpha1.WorkflowSpec{
				ID: "w1",
			},
		}
		s.OnGetMatch(mock.Anything, mock.Anything, mock.Anything).Return(wf, nil)
		s.OnUpdateMatch(mock.Anything, mock.Anything, mock.Anything).Return(nil, fmt.Errorf("unknown error")).Once()

		err := p.Handle(ctx, namespace, name)
		assert.Error(t, err)
	})

	t.Run("too-large-fail-repeat", func(t *testing.T) {
		scope := promutils.NewTestScope()
		s := &mocks.FlyteWorkflow{}
		exec := &mockExecutor{}
		p := NewPropellerHandler(ctx, cfg, nil, s, exec, scope)
		wf := &v1alpha1.FlyteWorkflow{
			ObjectMeta: v1.ObjectMeta{
				Name:      name,
				Namespace: namespace,
			},
			WorkflowSpec: &v1alpha1.WorkflowSpec{
				ID: "w1",
			},
		}
		s.OnGetMatch(mock.Anything, mock.Anything, mock.Anything).Return(wf, nil)
		s.OnUpdateMatch(mock.Anything, mock.Anything, mock.Anything).Return(nil, errors.Wrap(workflowstore.ErrWorkflowToLarge, "too large")).Twice()

		err := p.Handle(ctx, namespace, name)
		assert.Error(t, err)
	})

	t.Run("too-large-success", func(t *testing.T) {
		scope := promutils.NewTestScope()
		s := &mocks.FlyteWorkflow{}
		exec := &mockExecutor{}
		p := NewPropellerHandler(ctx, cfg, nil, s, exec, scope)
		wf := &v1alpha1.FlyteWorkflow{
			ObjectMeta: v1.ObjectMeta{
				Name:      name,
				Namespace: namespace,
			},
			WorkflowSpec: &v1alpha1.WorkflowSpec{
				ID: "w1",
			},
		}
		exec.HandleCb = func(ctx context.Context, w *v1alpha1.FlyteWorkflow) error {
			w.GetExecutionStatus().UpdatePhase(v1alpha1.WorkflowPhaseRunning, "done", nil)
			return nil
		}
		s.OnGetMatch(mock.Anything, mock.Anything, mock.Anything).Return(wf, nil)
		s.On("Update", mock.Anything, mock.Anything, mock.Anything).Return(nil, errors.Wrap(workflowstore.ErrWorkflowToLarge, "too large")).Once()
		s.On("Update", mock.Anything, mock.Anything, mock.Anything).Return(nil, nil).Once()

		err := p.Handle(ctx, namespace, name)
		assert.NoError(t, err)
	})
}

func TestPropellerHandler_OffloadedWorkflowClosure(t *testing.T) {
	ctx := context.TODO()

	const name = "123"
	const namespace = "test"

	s := workflowstore.NewInMemoryWorkflowStore()
	assert.NoError(t, s.Create(ctx, &v1alpha1.FlyteWorkflow{
		ObjectMeta: v1.ObjectMeta{
			Name:      name,
			Namespace: namespace,
		},
		WorkflowClosureReference: "some-file-location",
	}))

	exec := &mockExecutor{}
	exec.HandleCb = func(ctx context.Context, w *v1alpha1.FlyteWorkflow) error {
		w.GetExecutionStatus().UpdatePhase(v1alpha1.WorkflowPhaseSucceeding, "done", nil)
		return nil
	}

	cfg := &config.Config{
		MaxWorkflowRetries: 0,
	}

	t.Run("Happy", func(t *testing.T) {
		scope := promutils.NewTestScope()

		protoStore := &storagemocks.ComposedProtobufStore{}
		protoStore.OnReadProtobufMatch(mock.Anything, mock.Anything, mock.Anything).Run(func(args mock.Arguments) {
			// populate mock CompiledWorkflowClosure that satisfies just enough to compile
			wfClosure := args.Get(2)
			assert.NotNil(t, wfClosure)
			casted := wfClosure.(*admin.WorkflowClosure)
			casted.CompiledWorkflow = &core.CompiledWorkflowClosure{
				Primary: &core.CompiledWorkflow{
					Template: &core.WorkflowTemplate{
						Id: &core.Identifier{},
					},
				},
			}
		}).Return(nil)
		dataStore := storage.NewCompositeDataStore(storage.URLPathConstructor{}, protoStore)
		p := NewPropellerHandler(ctx, cfg, dataStore, s, exec, scope)

		assert.NoError(t, p.Handle(ctx, namespace, name))

		r, err := s.Get(ctx, namespace, name)
		assert.NoError(t, err)
		assert.Equal(t, v1alpha1.WorkflowPhaseSucceeding, r.GetExecutionStatus().GetPhase())
		assert.Equal(t, 1, len(r.Finalizers))
		assert.False(t, HasCompletedLabel(r))
		assert.Equal(t, uint32(0), r.Status.FailedAttempts)
		assert.Nil(t, r.WorkflowSpec)
		assert.Nil(t, r.SubWorkflows)
		assert.Nil(t, r.Tasks)
	})

	t.Run("Error", func(t *testing.T) {
		scope := promutils.NewTestScope()

		protoStore := &storagemocks.ComposedProtobufStore{}
		protoStore.OnReadProtobufMatch(mock.Anything, mock.Anything, mock.Anything).Return(fmt.Errorf("foo"))
		dataStore := storage.NewCompositeDataStore(storage.URLPathConstructor{}, protoStore)
		p := NewPropellerHandler(ctx, cfg, dataStore, s, exec, scope)

		err := p.Handle(ctx, namespace, name)
		assert.Error(t, err)
	})

	t.Run("TryMutate failure is handled", func(t *testing.T) {
		scope := promutils.NewTestScope()

		protoStore := &storagemocks.ComposedProtobufStore{}
		protoStore.OnReadProtobufMatch(mock.Anything, mock.Anything, mock.Anything).Return(fmt.Errorf("foo"))
		exec.HandleCb = func(ctx context.Context, w *v1alpha1.FlyteWorkflow) error {
			return fmt.Errorf("foo")
		}
		dataStore := storage.NewCompositeDataStore(storage.URLPathConstructor{}, protoStore)
		p := NewPropellerHandler(ctx, cfg, dataStore, s, exec, scope)

		err := p.Handle(ctx, namespace, name)
		assert.Error(t, err, "foo")

		r, err := s.Get(ctx, namespace, name)
		assert.NoError(t, err)
		assert.Nil(t, r.WorkflowSpec)
		assert.Nil(t, r.SubWorkflows)
		assert.Nil(t, r.Tasks)
	})
}
