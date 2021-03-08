package webapi

import (
	"context"
	"fmt"

	"k8s.io/utils/clock"

	"github.com/flyteorg/flytestdlib/logger"

	"github.com/flyteorg/flyteplugins/go/tasks/pluginmachinery/webapi"

	"github.com/flyteorg/flyteplugins/go/tasks/pluginmachinery/core"
)

type tokenAllocator struct {
	clock clock.Clock
}

func newTokenAllocator(c clock.Clock) tokenAllocator {
	return tokenAllocator{
		clock: c,
	}
}

func (a tokenAllocator) allocateToken(ctx context.Context, p webapi.AsyncPlugin, tCtx core.TaskExecutionContext, state *State, metrics Metrics) (
	newState *State, phaseInfo core.PhaseInfo, err error) {
	if len(p.GetConfig().ResourceQuotas) == 0 {
		// No quota, return success
		return &State{
			AllocationTokenRequestStartTime: a.clock.Now(),
			Phase:                           PhaseAllocationTokenAcquired,
		}, core.PhaseInfoQueued(a.clock.Now(), 0, "No allocation token required"), nil
	}

	ns, constraints, err := p.ResourceRequirements(ctx, tCtx)
	if err != nil {
		logger.Errorf(ctx, "Failed to calculate resource requirements for task. Error: %v", err)
		return nil, core.PhaseInfo{}, err
	}

	token := tCtx.TaskExecutionMetadata().GetTaskExecutionID().GetGeneratedName()
	allocationStatus, err := tCtx.ResourceManager().AllocateResource(ctx, ns, token, constraints)
	if err != nil {
		logger.Errorf(ctx, "Failed to allocate resources for task. Error: %v", err)
		return nil, core.PhaseInfo{}, err
	}

	switch allocationStatus {
	case core.AllocationStatusGranted:
		metrics.AllocationGranted.Inc(ctx)
		metrics.ResourceWaitTime.Observe(float64(a.clock.Since(state.AllocationTokenRequestStartTime).Milliseconds()))
		return &State{
			AllocationTokenRequestStartTime: a.clock.Now(),
			Phase:                           PhaseAllocationTokenAcquired,
		}, core.PhaseInfoQueued(a.clock.Now(), 0, "Allocation token required"), nil
	case core.AllocationStatusNamespaceQuotaExceeded:
	case core.AllocationStatusExhausted:
		metrics.AllocationNotGranted.Inc(ctx)
		logger.Infof(ctx, "Couldn't allocate token because allocation status is [%v].", allocationStatus.String())
		startTime := state.AllocationTokenRequestStartTime
		if startTime.IsZero() {
			startTime = a.clock.Now()
		}

		return &State{
				AllocationTokenRequestStartTime: startTime,
				Phase:                           PhaseNotStarted,
			}, core.PhaseInfoQueued(
				a.clock.Now(), 0, "Quota for task has exceeded. The request is enqueued."), nil
	}

	return nil, core.PhaseInfo{}, fmt.Errorf("allocation status undefined [%v]", allocationStatus)
}

func (a tokenAllocator) releaseToken(ctx context.Context, p webapi.AsyncPlugin, tCtx core.TaskExecutionContext, metrics Metrics) error {
	ns, _, err := p.ResourceRequirements(ctx, tCtx)
	if err != nil {
		logger.Errorf(ctx, "Failed to calculate resource requirements for task. Error: %v", err)
		return err
	}

	token := tCtx.TaskExecutionMetadata().GetTaskExecutionID().GetGeneratedName()
	err = tCtx.ResourceManager().ReleaseResource(ctx, ns, token)
	if err != nil {
		metrics.ResourceReleaseFailed.Inc(ctx)
		logger.Errorf(ctx, "Failed to release resources for task. Error: %v", err)
		return err
	}

	metrics.ResourceReleased.Inc(ctx)
	return nil
}
