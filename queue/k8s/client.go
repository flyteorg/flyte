package k8s

import (
	"context"
	"fmt"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes/scheme"
	"sigs.k8s.io/controller-runtime/pkg/client"

	executorv1 "github.com/flyteorg/flyte/v2/executor/api/v1"
	"github.com/flyteorg/flyte/v2/flytestdlib/logger"
	"github.com/flyteorg/flyte/v2/gen/go/flyteidl2/common"
	"github.com/flyteorg/flyte/v2/gen/go/flyteidl2/workflow"
)

// QueueClient implements queue operations using Kubernetes TaskAction CRs
type QueueClient struct {
	k8sClient client.Client
	namespace string
}

// NewQueueClient creates a new Kubernetes-based queue client
func NewQueueClient(k8sClient client.Client, namespace string) *QueueClient {
	return &QueueClient{
		k8sClient: k8sClient,
		namespace: namespace,
	}
}

// EnqueueAction creates a TaskAction CR in Kubernetes
func (q *QueueClient) EnqueueAction(ctx context.Context, req *workflow.EnqueueActionRequest) error {
	logger.Infof(ctx, "Enqueuing action: %s/%s/%s/%s/%s",
		req.ActionId.Run.Org, req.ActionId.Run.Project, req.ActionId.Run.Domain,
		req.ActionId.Run.Name, req.ActionId.Name)

	// Build ActionSpec from EnqueueActionRequest
	actionSpec := buildActionSpec(req)

	// Determine if this is the root action
	isRootAction := req.ParentActionName == nil || *req.ParentActionName == ""

	// Create TaskAction CR
	taskAction := &executorv1.TaskAction{
		ObjectMeta: metav1.ObjectMeta{
			Name:      buildTaskActionName(req.ActionId),
			Namespace: q.namespace,
			Labels: map[string]string{
				"flyte.org/org":         req.ActionId.Run.Org,
				"flyte.org/project":     req.ActionId.Run.Project,
				"flyte.org/domain":      req.ActionId.Run.Domain,
				"flyte.org/run":         req.ActionId.Run.Name,
				"flyte.org/action":      req.ActionId.Name,
				"flyte.org/action-type": "task", // TODO: detect from spec
				"flyte.org/is-root":     fmt.Sprintf("%t", isRootAction),
			},
		},
		Spec: executorv1.TaskActionSpec{},
	}

	// If this is not the root action, set the owner reference to the parent action
	if !isRootAction {
		parentActionID := &common.ActionIdentifier{
			Run:  req.ActionId.Run,
			Name: *req.ParentActionName,
		}
		parentTaskActionName := buildTaskActionName(parentActionID)

		// Get the parent TaskAction to set as owner
		parentTaskAction := &executorv1.TaskAction{}
		if err := q.k8sClient.Get(ctx, client.ObjectKey{
			Name:      parentTaskActionName,
			Namespace: q.namespace,
		}, parentTaskAction); err != nil {
			return fmt.Errorf("failed to get parent TaskAction %s: %w", parentTaskActionName, err)
		}

		// Set owner reference with cascading deletion enabled
		blockOwnerDeletion := true
		taskAction.OwnerReferences = []metav1.OwnerReference{
			{
				APIVersion:         "flyte.org/v1",
				Kind:               "TaskAction",
				Name:               parentTaskAction.Name,
				UID:                parentTaskAction.UID,
				BlockOwnerDeletion: &blockOwnerDeletion,
			},
		}

		logger.Infof(ctx, "Setting parent TaskAction %s as owner for %s",
			parentTaskActionName, taskAction.Name)
	} else {
		logger.Infof(ctx, "Creating root TaskAction: %s", taskAction.Name)
	}

	// Set the ActionSpec bytes
	if err := taskAction.Spec.SetActionSpec(actionSpec); err != nil {
		return fmt.Errorf("failed to set action spec: %w", err)
	}

	// Create in Kubernetes
	if err := q.k8sClient.Create(ctx, taskAction); err != nil {
		return fmt.Errorf("failed to create TaskAction CR: %w", err)
	}

	logger.Infof(ctx, "Successfully created TaskAction CR: %s", taskAction.Name)
	return nil
}

// AbortQueuedRun deletes the root TaskAction CR for a run
// Kubernetes will automatically cascade delete all child TaskActions via OwnerReferences
func (q *QueueClient) AbortQueuedRun(ctx context.Context, runID *common.RunIdentifier, reason *string) error {
	logger.Infof(ctx, "Aborting run by deleting root TaskAction: %s/%s/%s/%s",
		runID.Org, runID.Project, runID.Domain, runID.Name)

	// Find the root TaskAction for this run
	taskActionList := &executorv1.TaskActionList{}
	listOpts := []client.ListOption{
		client.InNamespace(q.namespace),
		client.MatchingLabels{
			"flyte.org/org":     runID.Org,
			"flyte.org/project": runID.Project,
			"flyte.org/domain":  runID.Domain,
			"flyte.org/run":     runID.Name,
			"flyte.org/is-root": "true",
		},
	}

	if err := q.k8sClient.List(ctx, taskActionList, listOpts...); err != nil {
		return fmt.Errorf("failed to list TaskActions: %w", err)
	}

	if len(taskActionList.Items) == 0 {
		logger.Warnf(ctx, "No root TaskAction found for run %s", runID.Name)
		return nil
	}

	if len(taskActionList.Items) > 1 {
		logger.Warnf(ctx, "Found %d root TaskActions for run %s, expected 1",
			len(taskActionList.Items), runID.Name)
	}

	// Delete the root TaskAction - Kubernetes will cascade delete all children
	rootTaskAction := &taskActionList.Items[0]
	if err := q.k8sClient.Delete(ctx, rootTaskAction); err != nil {
		return fmt.Errorf("failed to delete root TaskAction %s: %w", rootTaskAction.Name, err)
	}

	logger.Infof(ctx, "Deleted root TaskAction: %s (children will be cascade deleted by Kubernetes)",
		rootTaskAction.Name)
	return nil
}

// AbortQueuedAction deletes a specific TaskAction CR
// If the action has children, Kubernetes will cascade delete them via OwnerReferences
func (q *QueueClient) AbortQueuedAction(ctx context.Context, actionID *common.ActionIdentifier, reason *string) error {
	logger.Infof(ctx, "Aborting action: %s/%s/%s/%s/%s",
		actionID.Run.Org, actionID.Run.Project, actionID.Run.Domain,
		actionID.Run.Name, actionID.Name)

	taskActionName := buildTaskActionName(actionID)

	// Get the TaskAction to check if it exists and if it's a root action
	taskAction := &executorv1.TaskAction{}
	if err := q.k8sClient.Get(ctx, client.ObjectKey{
		Name:      taskActionName,
		Namespace: q.namespace,
	}, taskAction); err != nil {
		return fmt.Errorf("failed to get TaskAction %s: %w", taskActionName, err)
	}

	isRoot := taskAction.Labels["flyte.org/is-root"] == "true"

	// Delete the TaskAction - children will be cascade deleted by Kubernetes
	if err := q.k8sClient.Delete(ctx, taskAction); err != nil {
		return fmt.Errorf("failed to delete TaskAction %s: %w", taskActionName, err)
	}

	if isRoot {
		logger.Infof(ctx, "Deleted root TaskAction %s (all child actions will be cascade deleted)",
			taskActionName)
	} else {
		logger.Infof(ctx, "Deleted TaskAction %s (any child actions will be cascade deleted)",
			taskActionName)
	}

	return nil
}

// Helper functions

// buildTaskActionName generates a Kubernetes-compliant name for the TaskAction
func buildTaskActionName(actionID *common.ActionIdentifier) string {
	// K8s names must be lowercase alphanumeric, dashes, or dots
	// Format: org-project-domain-run-action
	return fmt.Sprintf("%s-%s-%s-%s-%s",
		actionID.Run.Org,
		actionID.Run.Project,
		actionID.Run.Domain,
		actionID.Run.Name,
		actionID.Name,
	)
}

// buildActionSpec converts EnqueueActionRequest to ActionSpec
func buildActionSpec(req *workflow.EnqueueActionRequest) *workflow.ActionSpec {
	actionSpec := &workflow.ActionSpec{
		ActionId:         req.ActionId,
		ParentActionName: req.ParentActionName,
		RunSpec:          req.RunSpec,
		InputUri:         req.InputUri,
		RunOutputBase:    req.RunOutputBase,
		Group:            req.Group,
	}

	// Set the spec based on the request type
	switch spec := req.Spec.(type) {
	case *workflow.EnqueueActionRequest_Task:
		actionSpec.Spec = &workflow.ActionSpec_Task{
			Task: spec.Task,
		}
	case *workflow.EnqueueActionRequest_Trace:
		actionSpec.Spec = &workflow.ActionSpec_Trace{
			Trace: spec.Trace,
		}
	case *workflow.EnqueueActionRequest_Condition:
		actionSpec.Spec = &workflow.ActionSpec_Condition{
			Condition: spec.Condition,
		}
	}

	return actionSpec
}

// InitScheme adds the executor API types to the scheme
func InitScheme() error {
	return executorv1.AddToScheme(scheme.Scheme)
}
