package k8s

import (
	"context"

	"connectrpc.com/connect"
	"github.com/prometheus/client_golang/prometheus"
	"google.golang.org/protobuf/proto"

	executorv1 "github.com/flyteorg/flyte/v2/executor/api/v1"
	"github.com/flyteorg/flyte/v2/flytestdlib/logger"
	"github.com/flyteorg/flyte/v2/flytestdlib/promutils"
	"github.com/flyteorg/flyte/v2/gen/go/flyteidl2/actions"
	"github.com/flyteorg/flyte/v2/gen/go/flyteidl2/common"
	"github.com/flyteorg/flyte/v2/gen/go/flyteidl2/workflow"
)

// Reasons an action of a recovery run runs fresh anyway.
const (
	rerunReasonForced   = "forced"
	rerunReasonMissing  = "missing"
	rerunReasonNonFinal = "non_final"
	rerunReasonNoOutput = "no_output"
	rerunReasonScope    = "scope"
)

type recoveryMetrics struct {
	recovered    prometheus.Counter
	rerun        *prometheus.CounterVec
	lookupErrors prometheus.Counter
}

// The counters no-op on a nil receiver: an ActionsClient built without NewActionsClient
// (tests) has no scope to register against.
func (m *recoveryMetrics) hit() {
	if m != nil {
		m.recovered.Inc()
	}
}

func (m *recoveryMetrics) fresh(reason string) {
	if m != nil {
		m.rerun.WithLabelValues(reason).Inc()
	}
}

func (m *recoveryMetrics) lookupFailed() {
	if m != nil {
		m.lookupErrors.Inc()
	}
}

func newRecoveryMetrics(scope promutils.Scope) *recoveryMetrics {
	s := scope.NewSubScope("recovery")
	return &recoveryMetrics{
		recovered:    s.MustNewCounter("recovered_total", "Actions reused from the source run instead of executed"),
		rerun:        s.MustNewCounterVec("rerun_total", "Actions of a recovery run that ran fresh anyway", "reason"),
		lookupErrors: s.MustNewCounter("lookup_errors_total", "Failed source-action lookups, each degraded into a fresh run"),
	}
}

// resolveRecoveredFrom runs the recovery gates for one enqueued action and returns what to
// stamp on its CR, or nil to create it exactly as an ordinary action.
//
// Fail-open throughout: recovery is an optimisation, so anything unexpected degrades into a
// fresh execution rather than failing the enqueue.
func (c *ActionsClient) resolveRecoveredFrom(
	ctx context.Context,
	taskAction *executorv1.TaskAction,
	action *actions.Action,
	isRoot bool,
) *executorv1.RecoveredFrom {
	// The root re-runs: it is the action that decides which children to enqueue at all.
	if isRoot {
		return nil
	}

	recoveryContext := taskAction.Spec.RecoveryContext
	if recoveryContext == nil || c.runClient == nil {
		return nil
	}

	relation := &common.Relation{}
	if err := proto.Unmarshal(recoveryContext.Relation, relation); err != nil {
		logger.Warnf(ctx, "recovery: failed to unmarshal relation for action %s: %v", action.ActionId.Name, err)
		c.recoveryMetrics.lookupFailed()
		return nil
	}
	source := relation.GetRelatedTo()
	if relation.GetRelationType() != common.RelationType_RELATION_TYPE_RECOVER || source == nil {
		return nil
	}

	actionName := action.ActionId.Name
	for _, forced := range recoveryContext.ForceRerunActions {
		if forced == actionName {
			c.recoveryMetrics.fresh(rerunReasonForced)
			return nil
		}
	}

	// CreateRun rejects a cross-scope relation, so reaching this is a bug rather than user
	// input — but the lookup is keyed by run identity alone, so it stays checked here too.
	target := action.ActionId.Run
	if source.GetProject() != target.GetProject() || source.GetDomain() != target.GetDomain() {
		logger.Warnf(ctx, "recovery: source run %s/%s is out of scope for %s/%s",
			source.GetProject(), source.GetDomain(), target.GetProject(), target.GetDomain())
		c.recoveryMetrics.fresh(rerunReasonScope)
		return nil
	}

	resp, err := c.runClient.LookupAction(ctx, connect.NewRequest(&workflow.LookupActionRequest{
		ActionId: &common.ActionIdentifier{Run: source, Name: actionName},
	}))
	if err != nil {
		if connect.CodeOf(err) == connect.CodeNotFound {
			c.recoveryMetrics.fresh(rerunReasonMissing)
			return nil
		}
		logger.Warnf(ctx, "recovery: lookup of %s in run %s failed, running fresh: %v",
			actionName, source.GetName(), err)
		c.recoveryMetrics.lookupFailed()
		return nil
	}

	switch {
	case !isRecoverablePhase(resp.Msg.GetPhase()):
		c.recoveryMetrics.fresh(rerunReasonNonFinal)
		return nil
	case resp.Msg.GetOutputUri() == "" && resp.Msg.GetOutput() == nil:
		// Nothing to hand downstream, so reusing the row would only hide a re-run. A
		// signalled condition has no outputs file but does carry an inline Literal, so
		// the two are checked together — gating on the URI alone re-ran every condition,
		// which then paused for a signal the source run had already been given.
		c.recoveryMetrics.fresh(rerunReasonNoOutput)
		return nil
	}

	// A condition's result rides in the CR rather than in object storage.
	var output []byte
	if literal := resp.Msg.GetOutput(); literal != nil {
		marshaled, err := proto.Marshal(literal)
		if err != nil {
			logger.Warnf(ctx, "recovery: failed to marshal output of %s in run %s, running fresh: %v",
				actionName, source.GetName(), err)
			c.recoveryMetrics.fresh(rerunReasonNoOutput)
			return nil
		}
		output = marshaled
	}

	c.recoveryMetrics.hit()
	return &executorv1.RecoveredFrom{
		SourceRunName: source.GetName(),
		OutputUri:     resp.Msg.GetOutputUri(),
		Output:        output,
		Attempts:      resp.Msg.GetAttempts(),
		CacheStatus:   int32(resp.Msg.GetCacheStatus()),
	}
}

// RECOVERED counts: recovering a recovery run is the ordinary way a run survives repeated
// intermittent failures, and its output URI is already fully resolved.
func isRecoverablePhase(phase common.ActionPhase) bool {
	return phase == common.ActionPhase_ACTION_PHASE_SUCCEEDED ||
		phase == common.ActionPhase_ACTION_PHASE_RECOVERED
}

// isConditionResultPhase reports the terminal phases in which a condition action carries a
// resolved signal value: SUCCEEDED when it was signalled in this run, RECOVERED when the
// value was replayed from the run being recovered.
func isConditionResultPhase(phase common.ActionPhase) bool {
	return phase == common.ActionPhase_ACTION_PHASE_SUCCEEDED ||
		phase == common.ActionPhase_ACTION_PHASE_RECOVERED
}
