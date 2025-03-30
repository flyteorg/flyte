package concurrency

import (
	"context"
	"fmt"
	"strings"
	"sync"
	"time"

	"github.com/benbjohnson/clock"
	"github.com/prometheus/client_golang/prometheus"
	"golang.org/x/time/rate"
	"k8s.io/client-go/util/workqueue"

	concurrencyCore "github.com/flyteorg/flyte/flyteadmin/concurrency/core"
	"github.com/flyteorg/flyte/flyteadmin/concurrency/executor"
	"github.com/flyteorg/flyte/flyteadmin/concurrency/informer"
	"github.com/flyteorg/flyte/flyteadmin/concurrency/repositories/interfaces"
	"github.com/flyteorg/flyte/flyteadmin/pkg/repositories/models"
	"github.com/flyteorg/flyte/flyteidl/gen/pb-go/flyteidl/admin"
	idlCore "github.com/flyteorg/flyte/flyteidl/gen/pb-go/flyteidl/core"
	"github.com/flyteorg/flyte/flytestdlib/logger"
	"github.com/flyteorg/flyte/flytestdlib/promutils"
)

// ConcurrencyController implements core.ConcurrencyController
type ConcurrencyController struct {
	repo               interfaces.ConcurrencyRepoInterface
	policyEvaluator    concurrencyCore.PolicyEvaluator
	workflowExecutor   *executor.WorkflowExecutor
	launchPlanInformer *informer.LaunchPlanInformer

	// In-memory tracking of running executions per launch plan.
	// Note: map[NamedEntityIdentifier][]
	runningExecutions      map[string]map[string]bool
	runningExecutionsMutex sync.RWMutex

	// Workqueue for pending executions
	workqueue workqueue.RateLimitingInterface

	// Processing settings
	processingInterval time.Duration
	maxRetries         int
	workers            int

	// Metrics
	metrics *concurrencyMetrics

	// For controlled shutdown
	stopCh chan struct{}
	clock  clock.Clock
}

type concurrencyMetrics struct {
	pendingExecutions       prometheus.Gauge
	processingLatency       promutils.StopWatch
	processingSuccessCount  prometheus.Counter
	processingFailureCount  prometheus.Counter
	executionsQueuedCount   prometheus.Counter
	executionsAbortedCount  prometheus.Counter
	executionsAllowedCount  prometheus.Counter
	executionsRejectedCount prometheus.Counter
}

// NewConcurrencyController creates a new concurrency controller
func NewConcurrencyController(
	repo interfaces.ConcurrencyRepoInterface,
	workflowExecutor *executor.WorkflowExecutor,
	launchPlanInformer *informer.LaunchPlanInformer,
	processingInterval time.Duration,
	workers int,
	maxRetries int,
	scope promutils.Scope,
) *ConcurrencyController {
	controllerScope := scope.NewSubScope("controller")

	rateLimiter := workqueue.NewMaxOfRateLimiter(
		workqueue.NewItemExponentialFailureRateLimiter(100*time.Millisecond, 1000*time.Second),
		&workqueue.BucketRateLimiter{Limiter: rate.NewLimiter(rate.Limit(10), 100)},
	)

	return &ConcurrencyController{
		repo:               repo,
		policyEvaluator:    &concurrencyCore.DefaultPolicyEvaluator{},
		workflowExecutor:   workflowExecutor,
		launchPlanInformer: launchPlanInformer,
		runningExecutions:  make(map[string]map[string]bool),
		workqueue:          workqueue.NewRateLimitingQueue(rateLimiter),
		processingInterval: processingInterval,
		maxRetries:         maxRetries,
		workers:            workers,
		metrics: &concurrencyMetrics{
			pendingExecutions:       controllerScope.MustNewGauge("pending_executions", "Number of pending executions"),
			processingLatency:       controllerScope.MustNewStopWatch("processing_latency", "Time to process pending executions", time.Millisecond),
			processingSuccessCount:  controllerScope.MustNewCounter("processing_success_count", "Count of successful processing operations"),
			processingFailureCount:  controllerScope.MustNewCounter("processing_failure_count", "Count of failed processing operations"),
			executionsQueuedCount:   controllerScope.MustNewCounter("executions_queued_count", "Count of executions that were queued"),
			executionsAbortedCount:  controllerScope.MustNewCounter("executions_aborted_count", "Count of executions that were aborted"),
			executionsAllowedCount:  controllerScope.MustNewCounter("executions_allowed_count", "Count of executions that were allowed to run"),
			executionsRejectedCount: controllerScope.MustNewCounter("executions_rejected_count", "Count of executions that were rejected"),
		},
		stopCh: make(chan struct{}),
		clock:  clock.New(),
	}
}

// Initialize initializes the controller
func (c *ConcurrencyController) Initialize(ctx context.Context) error {
	logger.Info(ctx, "Initializing concurrency controller")

	// Start the launch plan informer
	c.launchPlanInformer.Start(ctx)

	// Load initial state
	initialState, err := c.loadInitialState(ctx)
	if err != nil {
		return err
	}

	c.runningExecutionsMutex.Lock()
	c.runningExecutions = initialState
	c.runningExecutionsMutex.Unlock()

	// Start background workers
	for i := 0; i < c.workers; i++ {
		go c.runWorker(ctx)
	}

	// Start periodic processing of pending executions
	go c.periodicallyProcessPendingExecutions(ctx)

	return nil
}

// loadInitialState loads the initial state of running executions from the database
func (c *ConcurrencyController) loadInitialState(ctx context.Context) (map[string]map[string]bool, error) {
	// Get all launch plans with concurrency settings
	launchPlans, err := c.repo.GetAllLaunchPlansWithConcurrency(ctx)
	if err != nil {
		return nil, err
	}

	runningExecutions := make(map[string]map[string]bool)

	// For each launch plan, get all running executions
	for _, lp := range launchPlans {
		launchPlanID := idlCore.Identifier{
			Project: lp.Project,
			Domain:  lp.Domain,
			Name:    lp.Name,
		}

		executions, err := c.repo.GetRunningExecutionsForLaunchPlan(ctx, launchPlanID)
		if err != nil {
			return nil, err
		}

		lpKey := getLaunchPlanKey(lp.Project, lp.Domain, lp.Name)
		executionMap := make(map[string]bool)

		for _, exec := range executions {
			executionKey := getExecutionKey(exec.Project, exec.Domain, exec.Name)
			executionMap[executionKey] = true
		}

		if len(executionMap) > 0 {
			runningExecutions[lpKey] = executionMap
		}
	}

	return runningExecutions, nil
}

// periodicallyProcessPendingExecutions periodically processes pending executions
func (c *ConcurrencyController) periodicallyProcessPendingExecutions(ctx context.Context) {
	ticker := c.clock.Ticker(c.processingInterval)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			if err := c.ProcessPendingExecutions(ctx); err != nil {
				logger.Errorf(ctx, "Failed to process pending executions: %v", err)
			}
		case <-c.stopCh:
			return
		}
	}
}

// ProcessPendingExecutions processes all pending executions
func (c *ConcurrencyController) ProcessPendingExecutions(ctx context.Context) error {
	startTime := c.clock.Now()
	defer func() {
		c.metrics.processingLatency.Observe(startTime, c.clock.Now())
	}()

	pendingExecutions, err := c.repo.GetPendingExecutions(ctx)
	if err != nil {
		c.metrics.processingFailureCount.Inc()
		return err
	}

	c.metrics.pendingExecutions.Set(float64(len(pendingExecutions)))

	// Process each pending execution
	for _, execution := range pendingExecutions {
		// Enqueue the execution for processing
		c.workqueue.Add(getExecutionKey(execution.Project, execution.Domain, execution.Name))
	}

	c.metrics.processingSuccessCount.Inc()
	return nil
}

// runWorker is a long-running function that processes items from the workqueue
func (c *ConcurrencyController) runWorker(ctx context.Context) {
	for c.processNextWorkItem(ctx) {
	}
}

// processNextWorkItem processes the next item from the workqueue
func (c *ConcurrencyController) processNextWorkItem(ctx context.Context) bool {
	item, shutdown := c.workqueue.Get()
	if shutdown {
		return false
	}

	// We call Done here so the workqueue knows we have finished
	// processing this item. We also must remember to call Forget if we
	// do not want this work item being re-queued.
	defer c.workqueue.Done(item)

	executionKey, ok := item.(string)
	if !ok {
		// As the item in the workqueue is actually invalid, we call
		// Forget to ensure it's never queued again.
		c.workqueue.Forget(item)
		logger.Errorf(ctx, "Expected string in workqueue but got %#v", item)
		return true
	}

	// Process the execution
	err := c.processExecution(ctx, executionKey)
	if err == nil {
		// If no error occurred, Forget this item so it doesn't get queued again
		c.workqueue.Forget(item)
		return true
	}

	// If an error occurred, requeue unless the maximum number of retries has been reached
	if c.workqueue.NumRequeues(item) < c.maxRetries {
		logger.Warningf(ctx, "Error processing execution %s: %v", executionKey, err)
		c.workqueue.AddRateLimited(item)
		return true
	}

	// Too many retries
	logger.Errorf(ctx, "Dropping execution %s from the workqueue: %v", executionKey, err)
	c.workqueue.Forget(item)
	return true
}

// processExecution processes a single execution based on concurrency policy
func (c *ConcurrencyController) processExecution(ctx context.Context, executionKey string) error {
	// Parse the execution key
	parts := splitExecutionKey(executionKey)
	if len(parts) != 3 {
		return fmt.Errorf("invalid execution key: %s", executionKey)
	}

	// Reconstruct the execution ID
	executionID := idlCore.WorkflowExecutionIdentifier{
		Project: parts[0],
		Domain:  parts[1],
		Name:    parts[2],
	}

	// Get the execution from the database
	execution, err := c.repo.GetExecutionByID(ctx, executionID)
	if err != nil {
		return err
	}

	// If execution is no longer pending, skip processing
	if execution.Phase != idlCore.WorkflowExecution_PENDING.String() {
		return nil
	}

	// // Get the launch plan ID
	launchPlan, err := c.repo.GetLaunchPlanByID(ctx, execution.LaunchPlanID)
	if err != nil {
		return err
	}
	launchPlanID := idlCore.Identifier{
		Project: launchPlan.Project,
		Domain:  launchPlan.Domain,
		Name:    launchPlan.Name,
	}

	// Get the concurrency policy for the launch plan
	policy, err := c.launchPlanInformer.GetPolicy(ctx, launchPlanID)
	if err != nil {
		return err
	}

	// If no policy, or policy does not have a max concurrency, allow execution
	if policy == nil || policy.Max == 0 {
		return c.allowExecution(ctx, execution)
	}

	// Get current running executions count
	runningCount, err := c.GetExecutionCounts(ctx, launchPlanID)
	if err != nil {
		return err
	}

	// Check if the execution can proceed based on concurrency policy
	canExecute, _, err := c.policyEvaluator.CanExecute(ctx, *execution, policy, runningCount)
	if err != nil {
		return err
	}

	if canExecute {
		// Handle REPLACE policy - abort oldest execution first if needed
		if policy.Policy == admin.ConcurrencyPolicy_REPLACE && runningCount >= int(policy.Max) {
			oldestExec, err := c.repo.GetOldestRunningExecution(ctx, launchPlanID)
			if err != nil {
				return err
			}

			if oldestExec != nil {
				oldestExecID := idlCore.WorkflowExecutionIdentifier{
					Project: oldestExec.Project,
					Domain:  oldestExec.Domain,
					Name:    oldestExec.Name,
				}

				err = c.workflowExecutor.AbortExecution(ctx, oldestExecID, "Aborted due to concurrency policy REPLACE")
				if err != nil {
					return err
				}

				c.metrics.executionsAbortedCount.Inc()

				// Untrack the aborted execution
				c.UntrackExecution(ctx, oldestExecID)
			}
		}

		// Allow the execution to proceed
		return c.allowExecution(ctx, execution)
	} else {
		// Handle based on policy
		switch policy.Policy {
		case admin.ConcurrencyPolicy_ABORT:
			// Update execution phase to aborted
			err = c.repo.UpdateExecutionPhase(ctx, executionID, idlCore.WorkflowExecution_ABORTED)
			if err != nil {
				return err
			}

			// TODO Create Execution Event on Abort

			c.metrics.executionsAbortedCount.Inc()
			return err

		case admin.ConcurrencyPolicy_WAIT:
			// Keep the execution in pending state, it will be retried later
			c.metrics.executionsQueuedCount.Inc()
			return nil

		default:
			return fmt.Errorf("unknown concurrency policy: %v", policy.Policy)
		}
	}
}

// allowExecution allows an execution to proceed
func (c *ConcurrencyController) allowExecution(ctx context.Context, execution *models.Execution) error {
	// Create the execution
	err := c.workflowExecutor.CreateExecution(ctx, *execution)
	if err != nil {
		return err
	}

	// Get the launch plan from the DB
	launchPlan, err := c.repo.GetLaunchPlanByID(ctx, execution.LaunchPlanID)
	if err != nil {
		return err
	}
	launchPlanID := idlCore.Identifier{
		Project: launchPlan.Project,
		Domain:  launchPlan.Domain,
		Name:    launchPlan.Name,
	}

	err = c.TrackExecution(ctx, *execution, launchPlanID)
	if err != nil {
		return err
	}

	c.metrics.executionsAllowedCount.Inc()
	return nil
}

// CheckConcurrencyConstraints checks if an execution can proceed
func (c *ConcurrencyController) CheckConcurrencyConstraints(ctx context.Context, launchPlanID idlCore.Identifier) (bool, error) {
	policy, err := c.GetConcurrencyPolicy(ctx, launchPlanID)
	if err != nil {
		return false, err
	}

	// If no policy or no max concurrency, allow execution
	if policy == nil || policy.Max == 0 {
		return true, nil
	}

	runningCount, err := c.GetExecutionCounts(ctx, launchPlanID)
	if err != nil {
		return false, err
	}

	// Allow if count is less than max
	return runningCount < int(policy.Max), nil
}

// GetConcurrencyPolicy gets the concurrency policy for a launch plan
func (c *ConcurrencyController) GetConcurrencyPolicy(ctx context.Context, launchPlanID idlCore.Identifier) (*admin.SchedulerPolicy, error) {
	return c.launchPlanInformer.GetPolicy(ctx, launchPlanID)
}

// TrackExecution adds a running execution to tracking
func (c *ConcurrencyController) TrackExecution(ctx context.Context, execution models.Execution, launchPlanID idlCore.Identifier) error {
	lpKey := getLaunchPlanKey(launchPlanID.Project, launchPlanID.Domain, launchPlanID.Name)
	execKey := getExecutionKey(execution.Project, execution.Domain, execution.Name)

	c.runningExecutionsMutex.Lock()
	defer c.runningExecutionsMutex.Unlock()

	execMap, exists := c.runningExecutions[lpKey]
	if !exists {
		execMap = make(map[string]bool)
		c.runningExecutions[lpKey] = execMap
	}

	execMap[execKey] = true
	return nil
}

// UntrackExecution removes a completed execution from tracking
func (c *ConcurrencyController) UntrackExecution(ctx context.Context, executionID idlCore.WorkflowExecutionIdentifier) error {
	// Get the execution to determine its launch plan
	execution, err := c.repo.GetExecutionByID(ctx, executionID)
	if err != nil {
		return err
	}

	if execution == nil {
		return nil // Execution already removed
	}

	// Convert to launch plan key using the correct launch plan information
	lpKey := getLaunchPlanKey(executionID.Project, executionID.Domain, executionID.Name)
	execKey := getExecutionKey(executionID.Project, executionID.Domain, executionID.Name)

	c.runningExecutionsMutex.Lock()
	defer c.runningExecutionsMutex.Unlock()

	execMap, exists := c.runningExecutions[lpKey]
	if !exists {
		return nil
	}

	delete(execMap, execKey)

	// If no more executions for this launch plan, remove the map
	if len(execMap) == 0 {
		delete(c.runningExecutions, lpKey)
	}

	return nil
}

// GetExecutionCounts returns the count of running executions for a launch plan
func (c *ConcurrencyController) GetExecutionCounts(ctx context.Context, launchPlanID idlCore.Identifier) (int, error) {
	lpKey := getLaunchPlanKey(launchPlanID.Project, launchPlanID.Domain, launchPlanID.Name)

	c.runningExecutionsMutex.RLock()
	defer c.runningExecutionsMutex.RUnlock()

	executionMap, exists := c.runningExecutions[lpKey]
	if !exists {
		return 0, nil
	}

	return len(executionMap), nil
}

// Close gracefully shuts down the controller
func (c *ConcurrencyController) Close() error {
	logger.Info(context.Background(), "Shutting down concurrency controller")
	close(c.stopCh)
	c.workqueue.ShutDown()
	return nil
}

func (c *ConcurrencyController) RecordPendingExecutions(count int64) {
	c.metrics.pendingExecutions.Set(float64(count))
}

func (c *ConcurrencyController) RecordProcessedExecution(policy string, allowed bool) {
	if allowed {
		c.metrics.executionsAllowedCount.Inc()
	} else {
		c.metrics.executionsRejectedCount.Inc()
	}
}

func (c *ConcurrencyController) RecordLatency(operation string, startTime time.Time) {
	switch operation {
	case "processing":
		c.metrics.processingLatency.Observe(startTime, c.clock.Now())
	}
}

func getLaunchPlanKey(project, domain, name string) string {
	return fmt.Sprintf("lp:%s:%s:%s", project, domain, name)
}

func getExecutionKey(project, domain, name string) string {
	return fmt.Sprintf("exec:%s:%s:%s", project, domain, name)
}

func splitExecutionKey(key string) []string {
	return strings.Split(key, ":")
}
