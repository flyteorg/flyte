package controller

import (
	"context"
	"fmt"
	"time"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/log"

	flyteorgv1 "github.com/flyteorg/flyte/v2/executor/api/v1"
	"github.com/flyteorg/flyte/v2/gen/go/flyteidl2/common"
	"github.com/flyteorg/flyte/v2/gen/go/flyteidl2/core"
	"github.com/flyteorg/flyte/v2/gen/go/flyteidl2/task"
	"github.com/flyteorg/flyte/v2/gen/go/flyteidl2/workflow"
	"google.golang.org/protobuf/types/known/timestamppb"
)

func actionIdentifierOf(taskAction *flyteorgv1.TaskAction) *common.ActionIdentifier {
	return &common.ActionIdentifier{
		Run: &common.RunIdentifier{
			Project: taskAction.Spec.Project,
			Domain:  taskAction.Spec.Domain,
			Name:    taskAction.Spec.RunName,
		},
		Name: taskAction.Spec.ActionName,
	}
}

// reconcileRecovered marks an action terminal from its spec, without dispatching a plugin.
func (r *TaskActionReconciler) reconcileRecovered(
	ctx context.Context,
	taskAction *flyteorgv1.TaskAction,
	original *flyteorgv1.TaskAction,
) (ctrl.Result, error) {
	logger := log.FromContext(ctx)
	recovered := taskAction.Spec.RecoveredFrom
	msg := fmt.Sprintf("recovered from run %q", recovered.SourceRunName)

	setCondition(taskAction, flyteorgv1.ConditionTypeProgressing, metav1.ConditionFalse,
		flyteorgv1.ConditionReasonRecovered, msg)
	setCondition(taskAction, flyteorgv1.ConditionTypeSucceeded, metav1.ConditionTrue,
		flyteorgv1.ConditionReasonRecovered, msg)
	appendPhaseHistory(taskAction, string(flyteorgv1.ConditionReasonRecovered), msg)
	taskAction.Status.Attempts = recoveredAttempts(recovered)
	taskAction.Status.CacheStatus = core.CatalogCacheStatus(recovered.CacheStatus)

	if !taskActionStatusChanged(original.Status, taskAction.Status) {
		return ctrl.Result{}, nil
	}

	if err := r.recordEvent(ctx, recoveredActionEvent(taskAction, r.cluster)); err != nil {
		logger.Error(err, "failed to persist recovered action event", "name", taskAction.Name)
		return ctrl.Result{}, err
	}

	if err := r.persistStatusWithRetry(ctx, taskAction, func(latest *flyteorgv1.TaskAction) {
		latest.Status = taskAction.Status
	}); err != nil {
		logger.Error(err, "failed to persist recovered status", "name", taskAction.Name)
		return ctrl.Result{}, err
	}

	return ctrl.Result{}, r.ensureTerminalLabels(ctx, taskAction)
}

// recoveredActionEvent reports the source run's outputs verbatim. outputRefs is bypassed
// deliberately: it derives a URI from this run's output base, which nothing ever wrote to.
func recoveredActionEvent(taskAction *flyteorgv1.TaskAction, cluster string) *workflow.ActionEvent {
	recovered := taskAction.Spec.RecoveredFrom
	now := timestamppb.New(time.Now())

	event := &workflow.ActionEvent{
		Id:           actionIdentifierOf(taskAction),
		Attempt:      recoveredAttempts(recovered),
		Phase:        common.ActionPhase_ACTION_PHASE_RECOVERED,
		UpdatedTime:  now,
		ReportedTime: now,
		Cluster:      cluster,
		CacheStatus:  core.CatalogCacheStatus(recovered.CacheStatus),
	}
	if recovered.OutputUri != "" {
		event.Outputs = &task.OutputReferences{OutputUri: recovered.OutputUri}
	}
	return event
}

// ActionEvent.attempt is validated as greater than zero.
func recoveredAttempts(recovered *flyteorgv1.RecoveredFrom) uint32 {
	if recovered.Attempts == 0 {
		return 1
	}
	return recovered.Attempts
}
