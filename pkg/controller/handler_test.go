package controller

import (
	"context"
	"fmt"
	"testing"

	"github.com/lyft/flyteidl/gen/pb-go/flyteidl/core"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/lyft/flytepropeller/pkg/controller/config"
	"github.com/lyft/flytepropeller/pkg/controller/workflowstore"

	"github.com/lyft/flytestdlib/promutils"
	"github.com/stretchr/testify/assert"

	"github.com/lyft/flytepropeller/pkg/apis/flyteworkflow/v1alpha1"
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

	p := NewPropellerHandler(ctx, cfg, s, exec, scope)

	const namespace = "test"
	const name = "123"
	t.Run("notPresent", func(t *testing.T) {
		assert.NoError(t, p.Handle(ctx, namespace, name))
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
		assert.NoError(t, p.Handle(ctx, namespace, name))

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
		assert.NoError(t, p.Handle(ctx, namespace, name))

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
}

func TestPropellerHandler_Initialize(t *testing.T) {
	scope := promutils.NewTestScope()
	ctx := context.TODO()
	s := workflowstore.NewInMemoryWorkflowStore()
	exec := &mockExecutor{}
	cfg := &config.Config{
		MaxWorkflowRetries: 0,
	}

	p := NewPropellerHandler(ctx, cfg, s, exec, scope)

	assert.NoError(t, p.Initialize(ctx))
}
