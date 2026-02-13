package k8s

import (
	"context"
	"fmt"
	"strings"
	"sync"
	"time"

	"google.golang.org/protobuf/proto"
	"k8s.io/apimachinery/pkg/watch"
	"k8s.io/client-go/kubernetes/scheme"
	"sigs.k8s.io/controller-runtime/pkg/client"

	executorv1 "github.com/flyteorg/flyte/v2/executor/api/v1"
	"github.com/flyteorg/flyte/v2/flytestdlib/logger"
	"github.com/flyteorg/flyte/v2/gen/go/flyteidl2/common"
	"github.com/flyteorg/flyte/v2/gen/go/flyteidl2/core"
)

// ActionUpdate represents an update to a TaskAction
type ActionUpdate struct {
	ActionID         *common.ActionIdentifier
	ParentActionName string
	StateJSON        string
	Phase            string
	OutputUri        string
	IsDeleted        bool
	TaskType         string
	ShortName        string
}

// StateClient implements state operations using Kubernetes TaskAction CRs
type StateClient struct {
	k8sClient  client.WithWatch
	namespace  string
	bufferSize int

	// Watch management
	mu sync.RWMutex
	// Map parent action name to subscriber channels.
	// Multiple callers may watch the same parent action concurrently.
	// TODO: add a prometheus counter for dropped updates when metrics are wired up for the state service
	subscribers map[string]map[chan *ActionUpdate]struct{}
	stopCh      chan struct{}
	watching    bool
}

// NewStateClient creates a new Kubernetes-based state client
func NewStateClient(k8sClient client.WithWatch, namespace string, bufferSize int) *StateClient {
	return &StateClient{
		k8sClient:   k8sClient,
		namespace:   namespace,
		bufferSize:  bufferSize,
		subscribers: make(map[string]map[chan *ActionUpdate]struct{}),
	}
}

// GetState retrieves the state JSON for a TaskAction
func (c *StateClient) GetState(ctx context.Context, actionID *common.ActionIdentifier) (string, error) {
	taskActionName := buildTaskActionName(actionID)

	taskAction := &executorv1.TaskAction{}
	if err := c.k8sClient.Get(ctx, client.ObjectKey{
		Name:      taskActionName,
		Namespace: c.namespace,
	}, taskAction); err != nil {
		return "", fmt.Errorf("failed to get TaskAction %s: %w", taskActionName, err)
	}

	return taskAction.Status.StateJSON, nil
}

// PutState updates the state JSON for a TaskAction
func (c *StateClient) PutState(ctx context.Context, actionID *common.ActionIdentifier, stateJSON string) error {
	taskActionName := buildTaskActionName(actionID)

	// Get current TaskAction
	taskAction := &executorv1.TaskAction{}
	if err := c.k8sClient.Get(ctx, client.ObjectKey{
		Name:      taskActionName,
		Namespace: c.namespace,
	}, taskAction); err != nil {
		return fmt.Errorf("failed to get TaskAction %s: %w", taskActionName, err)
	}

	// Skip update if the stateJSON does not change
	if taskAction.Status.StateJSON == stateJSON {
		return nil
	}

	// Update state JSON
	taskAction.Status.StateJSON = stateJSON

	// Update status subresource
	if err := c.k8sClient.Status().Update(ctx, taskAction); err != nil {
		return fmt.Errorf("failed to update TaskAction status %s: %w", taskActionName, err)
	}

	logger.Infof(ctx, "Updated state for TaskAction: %s", taskActionName)
	return nil
}

// ListRunActions lists all TaskActions belonging to a run.
func (c *StateClient) ListRunActions(ctx context.Context, runID *common.RunIdentifier) ([]*executorv1.TaskAction, error) {
	taskActionList := &executorv1.TaskActionList{}
	listOpts := []client.ListOption{
		client.InNamespace(c.namespace),
		client.MatchingLabels{
			"flyte.org/org":     runID.Org,
			"flyte.org/project": runID.Project,
			"flyte.org/domain":  runID.Domain,
			"flyte.org/run":     runID.Name,
		},
	}

	if err := c.k8sClient.List(ctx, taskActionList, listOpts...); err != nil {
		return nil, fmt.Errorf("failed to list TaskActions for run: %w", err)
	}

	result := make([]*executorv1.TaskAction, len(taskActionList.Items))
	for i := range taskActionList.Items {
		result[i] = &taskActionList.Items[i]
	}
	return result, nil
}

// ListChildActions lists all TaskActions that are children of the given parent action
func (c *StateClient) ListChildActions(ctx context.Context, parentActionID *common.ActionIdentifier) ([]*executorv1.TaskAction, error) {
	// List all TaskActions in the same run
	taskActionList := &executorv1.TaskActionList{}
	listOpts := []client.ListOption{
		client.InNamespace(c.namespace),
		client.MatchingLabels{
			"flyte.org/org":     parentActionID.Run.Org,
			"flyte.org/project": parentActionID.Run.Project,
			"flyte.org/domain":  parentActionID.Run.Domain,
			"flyte.org/run":     parentActionID.Run.Name,
		},
	}

	if err := c.k8sClient.List(ctx, taskActionList, listOpts...); err != nil {
		return nil, fmt.Errorf("failed to list TaskActions: %w", err)
	}

	// Filter for the parent and its children
	var result []*executorv1.TaskAction
	for i := range taskActionList.Items {
		action := &taskActionList.Items[i]
		// Include the parent action itself
		if action.Spec.ActionName == parentActionID.Name {
			result = append(result, action)
			continue
		}
		// Include direct children
		if action.Spec.ParentActionName != nil && *action.Spec.ParentActionName == parentActionID.Name {
			result = append(result, action)
		}
	}

	return result, nil
}

// GetTaskAction retrieves a specific TaskAction
func (c *StateClient) GetTaskAction(ctx context.Context, actionID *common.ActionIdentifier) (*executorv1.TaskAction, error) {
	taskActionName := buildTaskActionName(actionID)

	taskAction := &executorv1.TaskAction{}
	if err := c.k8sClient.Get(ctx, client.ObjectKey{
		Name:      taskActionName,
		Namespace: c.namespace,
	}, taskAction); err != nil {
		return nil, fmt.Errorf("failed to get TaskAction %s: %w", taskActionName, err)
	}

	return taskAction, nil
}

// Subscribe creates a new subscription channel for action updates for specified parent action name
func (c *StateClient) Subscribe(parentActionName string) chan *ActionUpdate {
	c.mu.Lock()
	defer c.mu.Unlock()

	ch := make(chan *ActionUpdate, c.bufferSize)
	if c.subscribers[parentActionName] == nil {
		c.subscribers[parentActionName] = make(map[chan *ActionUpdate]struct{})
	}
	c.subscribers[parentActionName][ch] = struct{}{}
	return ch
}

// Unsubscribe removes the given channel from the subscription list for the parent action name
func (c *StateClient) Unsubscribe(parentActionName string, ch chan *ActionUpdate) {
	c.mu.Lock()
	defer c.mu.Unlock()

	if channels, ok := c.subscribers[parentActionName]; ok {
		delete(channels, ch)
		close(ch)
		if len(channels) == 0 {
			delete(c.subscribers, parentActionName)
		}
	}
}

// StartWatching starts watching TaskAction resources and notifies all subscribers
func (c *StateClient) StartWatching(ctx context.Context) error {
	c.mu.Lock()
	if c.watching {
		c.mu.Unlock()
		return nil
	}
	c.watching = true
	c.stopCh = make(chan struct{})
	c.mu.Unlock()

	logger.Infof(ctx, "Starting TaskAction watcher for namespace: %s", c.namespace)

	go c.watchLoop(ctx)

	return nil
}

// watchLoop continuously watches TaskAction resources
func (c *StateClient) watchLoop(ctx context.Context) {
	for {
		select {
		case <-c.stopCh:
			logger.Infof(ctx, "TaskAction watcher stopped")
			return
		case <-ctx.Done():
			logger.Infof(ctx, "TaskAction watcher context cancelled")
			return
		default:
			if err := c.doWatch(ctx); err != nil {
				logger.Warnf(ctx, "Watch error, will retry: %v", err)
				time.Sleep(5 * time.Second)
			}
		}
	}
}

// doWatch performs a single watch operation
func (c *StateClient) doWatch(ctx context.Context) error {
	taskActionList := &executorv1.TaskActionList{}

	watcher, err := c.k8sClient.Watch(ctx, taskActionList, client.InNamespace(c.namespace))
	if err != nil {
		return fmt.Errorf("failed to start watch: %w", err)
	}
	defer watcher.Stop()

	for {
		select {
		case <-c.stopCh:
			return nil
		case <-ctx.Done():
			return ctx.Err()
		case event, ok := <-watcher.ResultChan():
			if !ok {
				return fmt.Errorf("watch channel closed")
			}
			c.handleWatchEvent(ctx, event)
		}
	}
}

// handleWatchEvent processes a watch event
func (c *StateClient) handleWatchEvent(ctx context.Context, event watch.Event) {
	taskAction, ok := event.Object.(*executorv1.TaskAction)
	if !ok {
		logger.Warnf(ctx, "Received non-TaskAction object in watch event")
		return
	}

	var parentName string
	if taskAction.Spec.ParentActionName != nil {
		parentName = *taskAction.Spec.ParentActionName
	}

	// Determine short name: use spec.ShortName if set, otherwise extract from template ID
	shortName := taskAction.Spec.ShortName
	if shortName == "" && len(taskAction.Spec.TaskTemplate) > 0 {
		shortName = extractShortNameFromTemplate(taskAction.Spec.TaskTemplate)
	}

	update := &ActionUpdate{
		ActionID: &common.ActionIdentifier{
			Run: &common.RunIdentifier{
				Org:     taskAction.Spec.Org,
				Project: taskAction.Spec.Project,
				Domain:  taskAction.Spec.Domain,
				Name:    taskAction.Spec.RunName,
			},
			Name: taskAction.Spec.ActionName,
		},
		ParentActionName: parentName,
		StateJSON:        taskAction.Status.StateJSON,
		Phase:            GetPhaseFromConditions(taskAction),
		OutputUri:        buildOutputUri(taskAction),
		IsDeleted:        event.Type == watch.Deleted,
		TaskType:         taskAction.Spec.TaskType,
		ShortName:        shortName,
	}

	c.notifySubscribers(ctx, update)
}

// notifySubscribers sends an update to all subscribers
func (c *StateClient) notifySubscribers(ctx context.Context, update *ActionUpdate) {
	c.mu.RLock()
	defer c.mu.RUnlock()

	for ch := range c.subscribers[update.ParentActionName] {
		select {
		case ch <- update:
		default:
			logger.Warnf(ctx, "subscriber channel full, dropping update for parent action: %s", update.ParentActionName)
		}
	}
}

// StopWatching stops the TaskAction watcher
func (c *StateClient) StopWatching() {
	c.mu.Lock()
	defer c.mu.Unlock()

	if c.watching && c.stopCh != nil {
		close(c.stopCh)
		c.watching = false
	}
}

// GetPhaseFromConditions extracts the phase from TaskAction conditions.
func GetPhaseFromConditions(taskAction *executorv1.TaskAction) string {
	for _, cond := range taskAction.Status.Conditions {
		switch cond.Type {
		case string(executorv1.ConditionTypeSucceeded):
			if cond.Status == "True" {
				return "PHASE_SUCCEEDED"
			}
		case string(executorv1.ConditionTypeFailed):
			if cond.Status == "True" {
				return "PHASE_FAILED"
			}
		case string(executorv1.ConditionTypeProgressing):
			if cond.Status == "True" {
				switch cond.Reason {
				case string(executorv1.ConditionReasonQueued):
					return "PHASE_QUEUED"
				case string(executorv1.ConditionReasonInitializing):
					return "PHASE_INITIALIZING"
				case string(executorv1.ConditionReasonExecuting):
					return "PHASE_RUNNING"
				}
			}
		}
	}
	return "PHASE_UNSPECIFIED"
}

// buildTaskActionName generates a Kubernetes-compliant name for the TaskAction
func buildTaskActionName(actionID *common.ActionIdentifier) string {
	return fmt.Sprintf("%s-%s-%s-%s-%s",
		actionID.Run.Org,
		actionID.Run.Project,
		actionID.Run.Domain,
		actionID.Run.Name,
		actionID.Name,
	)
}

// buildOutputUri computes the action-specific output URI from the TaskAction spec.
func buildOutputUri(ta *executorv1.TaskAction) string {
	if ta.Spec.RunOutputBase == "" {
		return ""
	}
	return strings.TrimRight(ta.Spec.RunOutputBase, "/") + "/" + ta.Spec.ActionName
}

// InitScheme adds the executor API types to the scheme
func InitScheme() error {
	return executorv1.AddToScheme(scheme.Scheme)
}

// extractShortNameFromTemplate extracts a human-readable function name from a serialized TaskTemplate.
// It splits on '.' and returns the last part.
func extractShortNameFromTemplate(templateBytes []byte) string {
	tmpl := &core.TaskTemplate{}
	if err := proto.Unmarshal(templateBytes, tmpl); err != nil {
		return ""
	}
	if tmpl.GetId() == nil {
		return ""
	}
	name := tmpl.GetId().GetName()
	if name == "" {
		return ""
	}
	// Split on '.' and take the last part
	parts := strings.Split(name, ".")
	if len(parts) > 0 {
		return parts[len(parts)-1]
	}
	return name
}
