package service

import (
	"context"
	"errors"
	"sync/atomic"
	"testing"
	"time"

	"connectrpc.com/connect"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"

	"github.com/flyteorg/flyte/v2/gen/go/flyteidl2/actions"
	actionsconnectmocks "github.com/flyteorg/flyte/v2/gen/go/flyteidl2/actions/actionsconnect/mocks"
	"github.com/flyteorg/flyte/v2/gen/go/flyteidl2/common"
	repoMocks "github.com/flyteorg/flyte/v2/runs/repository/mocks"
	"github.com/flyteorg/flyte/v2/runs/repository/models"
)

// newTestReconciler builds a reconciler wired to mocks with fast timing for tests.
func newTestReconciler(t *testing.T) (*repoMocks.ActionRepo, *actionsconnectmocks.ActionsServiceClient, *AbortReconciler) {
	t.Helper()
	actionRepo := repoMocks.NewActionRepo(t)
	repo := repoMocks.NewRepository(t)
	repo.On("ActionRepo").Return(actionRepo).Maybe()

	actionsClient := actionsconnectmocks.NewActionsServiceClient(t)

	reconciler := NewAbortReconciler(repo, actionsClient, AbortReconcilerConfig{
		Workers:      2,
		MaxAttempts:  3,
		QueueSize:    100,
		InitialDelay: 10 * time.Millisecond,
		MaxDelay:     50 * time.Millisecond,
	})
	return actionRepo, actionsClient, reconciler
}

func abortTestActionID() *common.ActionIdentifier {
	return &common.ActionIdentifier{
		Run: &common.RunIdentifier{
			Org:     "org",
			Project: "proj",
			Domain:  "dev",
			Name:    "rtest1",
		},
		Name: "rtest1",
	}
}

func abortTestAction(actionID *common.ActionIdentifier) *models.Action {
	now := time.Now()
	reason := "User requested abort"
	return &models.Action{
		Project:          actionID.Run.Project,
		Domain:           actionID.Run.Domain,
		RunName:          actionID.Run.Name,
		Name:             actionID.Name,
		AbortRequestedAt: &now,
		AbortReason:      &reason,
	}
}

func TestAbortReconciler_SuccessOnFirstAttempt(t *testing.T) {
	actionRepo, actionsClient, reconciler := newTestReconciler(t)
	actionID := abortTestActionID()

	var cleared atomic.Bool
	actionRepo.On("ListPendingAborts", mock.Anything).Return([]*models.Action{abortTestAction(actionID)}, nil).Once()
	actionRepo.On("GetAction", mock.Anything, mock.Anything).Return(abortTestAction(actionID), nil).Maybe()
	actionRepo.On("MarkAbortAttempt", mock.Anything, mock.Anything).Return(1, nil).Once()
	actionsClient.On("Abort", mock.Anything, mock.Anything).
		Return(connect.NewResponse(&actions.AbortResponse{}), nil).Once()
	actionRepo.On("ClearAbortRequest", mock.Anything, mock.Anything).
		Run(func(_ mock.Arguments) { cleared.Store(true) }).
		Return(nil).Once()

	ctx, cancel := context.WithTimeout(context.Background(), 500*time.Millisecond)
	defer cancel()
	go func() { _ = reconciler.Run(ctx) }()

	assert.Eventually(t, cleared.Load, 400*time.Millisecond, 5*time.Millisecond)
	actionsClient.AssertExpectations(t)
}

func TestAbortReconciler_RetriesOnFailure(t *testing.T) {
	actionRepo, actionsClient, reconciler := newTestReconciler(t)
	actionID := abortTestActionID()

	var cleared atomic.Bool
	actionRepo.On("ListPendingAborts", mock.Anything).Return([]*models.Action{abortTestAction(actionID)}, nil).Once()
	actionRepo.On("GetAction", mock.Anything, mock.Anything).Return(abortTestAction(actionID), nil).Maybe()

	actionRepo.On("MarkAbortAttempt", mock.Anything, mock.Anything).Return(1, nil).Once()
	actionsClient.On("Abort", mock.Anything, mock.Anything).Return(nil, errors.New("unavailable")).Once()

	actionRepo.On("MarkAbortAttempt", mock.Anything, mock.Anything).Return(2, nil).Once()
	actionsClient.On("Abort", mock.Anything, mock.Anything).Return(nil, errors.New("unavailable")).Once()

	actionRepo.On("MarkAbortAttempt", mock.Anything, mock.Anything).Return(3, nil).Once()
	actionsClient.On("Abort", mock.Anything, mock.Anything).
		Return(connect.NewResponse(&actions.AbortResponse{}), nil).Once()
	actionRepo.On("ClearAbortRequest", mock.Anything, mock.Anything).
		Run(func(_ mock.Arguments) { cleared.Store(true) }).
		Return(nil).Once()

	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()
	go func() { _ = reconciler.Run(ctx) }()

	assert.Eventually(t, cleared.Load, 1500*time.Millisecond, 10*time.Millisecond)
	actionsClient.AssertExpectations(t)
}

func TestAbortReconciler_GivesUpAtMaxAttempts(t *testing.T) {
	actionRepo, actionsClient, reconciler := newTestReconciler(t)
	actionID := abortTestActionID()

	var cleared atomic.Bool
	actionRepo.On("ListPendingAborts", mock.Anything).Return([]*models.Action{abortTestAction(actionID)}, nil).Once()
	actionRepo.On("GetAction", mock.Anything, mock.Anything).Return(abortTestAction(actionID), nil).Maybe()

	for i := 1; i <= 3; i++ {
		actionRepo.On("MarkAbortAttempt", mock.Anything, mock.Anything).Return(i, nil).Once()
		actionsClient.On("Abort", mock.Anything, mock.Anything).Return(nil, errors.New("always fails")).Once()
	}
	actionRepo.On("ClearAbortRequest", mock.Anything, mock.Anything).
		Run(func(_ mock.Arguments) { cleared.Store(true) }).
		Return(nil).Once()

	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()
	go func() { _ = reconciler.Run(ctx) }()

	assert.Eventually(t, cleared.Load, 1500*time.Millisecond, 10*time.Millisecond)
	// No more abort attempts should happen after giving up.
	actionsClient.AssertExpectations(t)
}

func TestAbortReconciler_DeduplicatesQueue(t *testing.T) {
	actionRepo, actionsClient, reconciler := newTestReconciler(t)
	actionID := abortTestActionID()

	var cleared atomic.Bool
	actionRepo.On("ListPendingAborts", mock.Anything).Return([]*models.Action{abortTestAction(actionID)}, nil).Once()
	actionRepo.On("GetAction", mock.Anything, mock.Anything).Return(abortTestAction(actionID), nil).Maybe()
	actionRepo.On("MarkAbortAttempt", mock.Anything, mock.Anything).Return(1, nil).Once()
	actionsClient.On("Abort", mock.Anything, mock.Anything).
		Run(func(_ mock.Arguments) { time.Sleep(20 * time.Millisecond) }).
		Return(connect.NewResponse(&actions.AbortResponse{}), nil).Once()
	actionRepo.On("ClearAbortRequest", mock.Anything, mock.Anything).
		Run(func(_ mock.Arguments) { cleared.Store(true) }).
		Return(nil).Once()

	ctx, cancel := context.WithTimeout(context.Background(), 500*time.Millisecond)
	defer cancel()
	go func() { _ = reconciler.Run(ctx) }()

	// Push a duplicate while the first is being processed.
	time.Sleep(5 * time.Millisecond)
	reconciler.Push(ctx, actionID, "dup")

	assert.Eventually(t, cleared.Load, 400*time.Millisecond, 5*time.Millisecond)
	// Abort should only have been called once.
	actionsClient.AssertNumberOfCalls(t, "Abort", 1)
}

func TestAbortReconciler_StartupScanPicksUpPending(t *testing.T) {
	actionRepo, actionsClient, reconciler := newTestReconciler(t)

	action1ID := &common.ActionIdentifier{
		Run: &common.RunIdentifier{Org: "o", Project: "p", Domain: "d", Name: "run1"},
		Name: "run1",
	}
	action2ID := &common.ActionIdentifier{
		Run: &common.RunIdentifier{Org: "o", Project: "p", Domain: "d", Name: "run2"},
		Name: "run2",
	}

	now := time.Now()
	reason := "abort"
	pending := []*models.Action{
		{Project: "p", Domain: "d", RunName: "run1", Name: "run1", AbortRequestedAt: &now, AbortReason: &reason},
		{Project: "p", Domain: "d", RunName: "run2", Name: "run2", AbortRequestedAt: &now, AbortReason: &reason},
	}
	actionRepo.On("ListPendingAborts", mock.Anything).Return(pending, nil).Once()
	actionRepo.On("GetAction", mock.Anything, mock.Anything).Return(abortTestAction(action1ID), nil).Maybe()

	var count atomic.Int32
	for _, id := range []*common.ActionIdentifier{action1ID, action2ID} {
		actionRepo.On("MarkAbortAttempt", mock.Anything, mock.MatchedBy(func(a *common.ActionIdentifier) bool {
			return a.Name == id.Name
		})).Return(1, nil).Once()
		actionsClient.On("Abort", mock.Anything, mock.MatchedBy(func(req *connect.Request[actions.AbortRequest]) bool {
			return req.Msg.ActionId.Name == id.Name
		})).Return(connect.NewResponse(&actions.AbortResponse{}), nil).Once()
		actionRepo.On("ClearAbortRequest", mock.Anything, mock.MatchedBy(func(a *common.ActionIdentifier) bool {
			return a.Name == id.Name
		})).Run(func(_ mock.Arguments) { count.Add(1) }).Return(nil).Once()
	}

	ctx, cancel := context.WithTimeout(context.Background(), 500*time.Millisecond)
	defer cancel()
	go func() { _ = reconciler.Run(ctx) }()

	assert.Eventually(t, func() bool { return count.Load() == 2 }, 400*time.Millisecond, 5*time.Millisecond)
	actionsClient.AssertExpectations(t)
}

func TestAbortReconciler_NotFoundTreatedAsSuccess(t *testing.T) {
	actionRepo, actionsClient, reconciler := newTestReconciler(t)
	actionID := abortTestActionID()

	var cleared atomic.Bool
	actionRepo.On("ListPendingAborts", mock.Anything).Return([]*models.Action{abortTestAction(actionID)}, nil).Once()
	actionRepo.On("GetAction", mock.Anything, mock.Anything).Return(abortTestAction(actionID), nil).Maybe()
	actionRepo.On("MarkAbortAttempt", mock.Anything, mock.Anything).Return(1, nil).Once()
	notFoundErr := connect.NewError(connect.CodeNotFound, errors.New("action not found"))
	actionsClient.On("Abort", mock.Anything, mock.Anything).Return(nil, notFoundErr).Once()
	actionRepo.On("ClearAbortRequest", mock.Anything, mock.Anything).
		Run(func(_ mock.Arguments) { cleared.Store(true) }).
		Return(nil).Once()

	ctx, cancel := context.WithTimeout(context.Background(), 500*time.Millisecond)
	defer cancel()
	go func() { _ = reconciler.Run(ctx) }()

	assert.Eventually(t, cleared.Load, 400*time.Millisecond, 5*time.Millisecond,
		"ClearAbortRequest should be called even when actionsClient returns NotFound")
	// No retry should happen — only 1 abort call expected.
	actionsClient.AssertNumberOfCalls(t, "Abort", 1)
}