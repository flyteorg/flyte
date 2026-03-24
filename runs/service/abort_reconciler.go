package service

import (
	"context"
	"fmt"
	"strings"
	"sync"
	"time"

	"connectrpc.com/connect"

	"github.com/flyteorg/flyte/v2/flytestdlib/logger"
	"github.com/flyteorg/flyte/v2/gen/go/flyteidl2/actions"
	"github.com/flyteorg/flyte/v2/gen/go/flyteidl2/actions/actionsconnect"
	"github.com/flyteorg/flyte/v2/gen/go/flyteidl2/common"
	"github.com/flyteorg/flyte/v2/runs/repository/interfaces"
)

// AbortReconcilerConfig holds tunables for the reconciler.
type AbortReconcilerConfig struct {
	// Workers is the number of concurrent pod-termination goroutines.
	Workers int
	// MaxAttempts is the maximum number of actionsClient.Abort calls per action before giving up.
	MaxAttempts int
	// QueueSize is the buffer size of the internal channel.
	QueueSize int
	// InitialDelay is the backoff duration before the first retry.
	InitialDelay time.Duration
	// MaxDelay caps the exponential backoff.
	MaxDelay time.Duration
}

func defaultConfig() AbortReconcilerConfig {
	return AbortReconcilerConfig{
		Workers:      5,
		MaxAttempts:  10,
		QueueSize:    1000,
		InitialDelay: time.Second,
		MaxDelay:     5 * time.Minute,
	}
}

// abortTask is a unit of work for a worker.
type abortTask struct {
	actionID *common.ActionIdentifier
	reason   string
	key      string // "org/project/domain/run/actionName"
}

// dedupeQueue is an in-memory, key-deduplicated work queue backed by a Go channel.
// Pushing a key that is already present (being processed or waiting for requeue) is a no-op.
type dedupeQueue struct {
	mu   sync.Mutex
	keys map[string]struct{}
	ch   chan abortTask
}

func newDedupeQueue(size int) *dedupeQueue {
	return &dedupeQueue{
		keys: make(map[string]struct{}),
		ch:   make(chan abortTask, size),
	}
}

// push adds task to the queue if its key is not already present.
// Returns false (no-op) when the key is a duplicate.
func (q *dedupeQueue) push(ctx context.Context, task abortTask) bool {
	q.mu.Lock()
	if _, exists := q.keys[task.key]; exists {
		q.mu.Unlock()
		return false
	}
	q.keys[task.key] = struct{}{}
	q.mu.Unlock()

	select {
	case q.ch <- task:
		return true
	case <-ctx.Done():
		q.mu.Lock()
		delete(q.keys, task.key)
		q.mu.Unlock()
		return false
	}
}

// scheduleRequeue re-enqueues the task after delay.
// The key remains in the set during the wait window so that any NOTIFY arriving during
// the backoff is correctly deduped (no duplicate processing).
func (q *dedupeQueue) scheduleRequeue(ctx context.Context, task abortTask, delay time.Duration) {
	time.AfterFunc(delay, func() {
		// Remove then re-push so push's dedup check passes.
		q.mu.Lock()
		delete(q.keys, task.key)
		q.mu.Unlock()
		q.push(ctx, task)
	})
}

// remove removes the key from the set (called on successful termination).
func (q *dedupeQueue) remove(key string) {
	q.mu.Lock()
	delete(q.keys, key)
	q.mu.Unlock()
}

// AbortReconciler watches for abort requests and drives pod termination to completion
// with exponential backoff retries.
type AbortReconciler struct {
	repo          interfaces.Repository
	actionsClient actionsconnect.ActionsServiceClient
	queue         *dedupeQueue
	cfg           AbortReconcilerConfig
}

// NewAbortReconciler creates a new AbortReconciler. Zero-value cfg fields are filled with defaults.
func NewAbortReconciler(repo interfaces.Repository, actionsClient actionsconnect.ActionsServiceClient, cfg AbortReconcilerConfig) *AbortReconciler {
	def := defaultConfig()
	if cfg.Workers <= 0 {
		cfg.Workers = def.Workers
	}
	if cfg.MaxAttempts <= 0 {
		cfg.MaxAttempts = def.MaxAttempts
	}
	if cfg.QueueSize <= 0 {
		cfg.QueueSize = def.QueueSize
	}
	if cfg.InitialDelay <= 0 {
		cfg.InitialDelay = def.InitialDelay
	}
	if cfg.MaxDelay <= 0 {
		cfg.MaxDelay = def.MaxDelay
	}
	return &AbortReconciler{
		repo:          repo,
		actionsClient: actionsClient,
		queue:         newDedupeQueue(cfg.QueueSize),
		cfg:           cfg,
	}
}

// Run starts the reconciler. It blocks until ctx is cancelled.
func (r *AbortReconciler) Run(ctx context.Context) error {
	logger.Infof(ctx, "AbortReconciler starting (%d workers, max %d attempts)", r.cfg.Workers, r.cfg.MaxAttempts)

	// Start workers first so they can drain the queue as startupScan fills it.
	// If workers started after the scan, a pending-abort count exceeding QueueSize
	// would cause push() to block forever (no consumer, full channel).
	var wg sync.WaitGroup
	for i := 0; i < r.cfg.Workers; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			r.runWorker(ctx)
		}()
	}

	// Startup scan: enqueue any actions that were left pending from before this process started.
	if err := r.startupScan(ctx); err != nil {
		logger.Errorf(ctx, "AbortReconciler startup scan failed: %v", err)
		// Non-fatal — the NOTIFY watcher will still pick up new aborts.
	}

	// Watch for new abort requests via NOTIFY (or polling on SQLite).
	payloads := make(chan string, 50)
	errs := make(chan error, 10)
	go r.repo.ActionRepo().WatchAbortRequests(ctx, payloads, errs)

	for {
		select {
		case <-ctx.Done():
			wg.Wait()
			return nil
		case err := <-errs:
			logger.Warnf(ctx, "AbortReconciler watch error: %v", err)
		case payload := <-payloads:
			task, err := parseAbortPayload(payload)
			if err != nil {
				logger.Warnf(ctx, "AbortReconciler: invalid payload %q: %v", payload, err)
				continue
			}
			// Fetch current abort_reason from DB (NOTIFY payload does not carry reason).
			action, err := r.repo.ActionRepo().GetAction(ctx, task.actionID)
			if err != nil {
				logger.Warnf(ctx, "AbortReconciler: failed to fetch action %s: %v", task.key, err)
				continue
			}
			if action.AbortRequestedAt == nil {
				// Already cleared (race between scan and notify) — skip.
				continue
			}
			if action.AbortReason != nil {
				task.reason = *action.AbortReason
			}
			r.queue.push(ctx, task)
		}
	}
}

// startupScan enqueues all actions that have abort_requested_at set.
func (r *AbortReconciler) startupScan(ctx context.Context) error {
	pending, err := r.repo.ActionRepo().ListPendingAborts(ctx)
	if err != nil {
		return err
	}
	for _, a := range pending {
		task := abortTask{
			actionID: &common.ActionIdentifier{
				Run: &common.RunIdentifier{
					Org:     a.Org,
					Project: a.Project,
					Domain:  a.Domain,
					Name:    a.RunName,
				},
				Name: a.Name,
			},
			key: fmt.Sprintf("%s/%s/%s/%s/%s", a.Org, a.Project, a.Domain, a.RunName, a.Name),
		}
		if a.AbortReason != nil {
			task.reason = *a.AbortReason
		}
		r.queue.push(ctx, task)
	}
	logger.Infof(ctx, "AbortReconciler startup scan enqueued %d pending abort(s)", len(pending))
	return nil
}

// runWorker processes tasks from the queue until ctx is cancelled.
func (r *AbortReconciler) runWorker(ctx context.Context) {
	for {
		select {
		case <-ctx.Done():
			return
		case task := <-r.queue.ch:
			r.processTask(ctx, task)
		}
	}
}

// processTask increments the attempt counter then calls actionsClient.Abort.
// On success it clears the DB flag. On failure it schedules a retry or gives up.
func (r *AbortReconciler) processTask(ctx context.Context, task abortTask) {
	attemptCount, err := r.repo.ActionRepo().MarkAbortAttempt(ctx, task.actionID)
	if err != nil {
		logger.Errorf(ctx, "AbortReconciler: failed to mark attempt for %s: %v", task.key, err)
		// Re-enqueue without counting — the DB row is authoritative; try again later.
		r.queue.scheduleRequeue(ctx, task, r.cfg.InitialDelay)
		return
	}

	reason := task.reason
	_, abortErr := r.actionsClient.Abort(ctx, connect.NewRequest(&actions.AbortRequest{
		ActionId: task.actionID,
		Reason:   &reason,
	}))

	if abortErr == nil || isAlreadyTerminated(abortErr) {
		// Success (or already gone — treat as success).
		if clearErr := r.repo.ActionRepo().ClearAbortRequest(ctx, task.actionID); clearErr != nil {
			logger.Errorf(ctx, "AbortReconciler: failed to clear abort request for %s: %v", task.key, clearErr)
		}
		r.queue.remove(task.key)
		logger.Infof(ctx, "AbortReconciler: successfully aborted %s (attempt %d)", task.key, attemptCount)
		return
	}

	logger.Warnf(ctx, "AbortReconciler: abort failed for %s (attempt %d/%d): %v",
		task.key, attemptCount, r.cfg.MaxAttempts, abortErr)

	if attemptCount >= r.cfg.MaxAttempts {
		logger.Errorf(ctx, "AbortReconciler: giving up on %s after %d attempts — manual intervention may be required",
			task.key, attemptCount)
		if clearErr := r.repo.ActionRepo().ClearAbortRequest(ctx, task.actionID); clearErr != nil {
			logger.Errorf(ctx, "AbortReconciler: failed to clear abort request for %s: %v", task.key, clearErr)
		}
		r.queue.remove(task.key)
		return
	}

	// Exponential backoff: initialDelay * 2^(attempt-1), capped at maxDelay.
	delay := r.cfg.InitialDelay * (1 << (attemptCount - 1))
	if delay > r.cfg.MaxDelay {
		delay = r.cfg.MaxDelay
	}
	logger.Infof(ctx, "AbortReconciler: scheduling retry for %s in %s", task.key, delay)
	r.queue.scheduleRequeue(ctx, task, delay)
}

// isAlreadyTerminated returns true for errors that indicate the action is already gone.
func isAlreadyTerminated(err error) bool {
	if err == nil {
		return false
	}
	connectErr, ok := err.(*connect.Error)
	if !ok {
		return false
	}
	return connectErr.Code() == connect.CodeNotFound
}

// parseAbortPayload parses "org/project/domain/run/actionName" into an abortTask.
func parseAbortPayload(payload string) (abortTask, error) {
	parts := strings.SplitN(payload, "/", 5)
	if len(parts) != 5 {
		return abortTask{}, fmt.Errorf("expected 5 parts, got %d", len(parts))
	}
	return abortTask{
		key: payload,
		actionID: &common.ActionIdentifier{
			Run: &common.RunIdentifier{
				Org:     parts[0],
				Project: parts[1],
				Domain:  parts[2],
				Name:    parts[3],
			},
			Name: parts[4],
		},
	}, nil
}