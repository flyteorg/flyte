package webapi

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/mock"

	mocks2 "github.com/flyteorg/flyteplugins/go/tasks/pluginmachinery/core/mocks"
	"github.com/flyteorg/flyteplugins/go/tasks/pluginmachinery/webapi/mocks"

	"github.com/stretchr/testify/assert"

	testing2 "k8s.io/utils/clock/testing"

	"github.com/flyteorg/flytestdlib/contextutils"
	"github.com/flyteorg/flytestdlib/promutils/labeled"

	"github.com/flyteorg/flytestdlib/promutils"
	"github.com/go-test/deep"

	"github.com/flyteorg/flyteplugins/go/tasks/pluginmachinery/core"
	"github.com/flyteorg/flyteplugins/go/tasks/pluginmachinery/webapi"
)

func init() {
	labeled.SetMetricKeys(contextutils.NamespaceKey)
}

func newPluginWithProperties(properties webapi.PluginConfig) *mocks.AsyncPlugin {
	m := &mocks.AsyncPlugin{}
	m.OnGetConfig().Return(properties)
	return m
}

func Test_allocateToken(t *testing.T) {
	ctx := context.Background()
	metrics := newMetrics(promutils.NewTestScope())

	tNow := time.Now()
	clck := testing2.NewFakeClock(tNow)

	tID := &mocks2.TaskExecutionID{}
	tID.OnGetGeneratedName().Return("abc")

	tMeta := &mocks2.TaskExecutionMetadata{}
	tMeta.OnGetTaskExecutionID().Return(tID)

	rm := &mocks2.ResourceManager{}
	rm.OnAllocateResourceMatch(ctx, core.ResourceNamespace("ns"), "abc", mock.Anything).Return(core.AllocationStatusGranted, nil)
	rm.OnAllocateResourceMatch(ctx, core.ResourceNamespace("ns"), "abc2", mock.Anything).Return(core.AllocationStatusExhausted, nil)

	tCtx := &mocks2.TaskExecutionContext{}
	tCtx.OnTaskExecutionMetadata().Return(tMeta)
	tCtx.OnResourceManager().Return(rm)

	state := &State{}

	p := newPluginWithProperties(webapi.PluginConfig{
		ResourceQuotas: map[core.ResourceNamespace]int{
			"ns": 1,
		},
	})

	t.Run("no quota", func(t *testing.T) {
		p := newPluginWithProperties(webapi.PluginConfig{ResourceQuotas: nil})
		a := newTokenAllocator(clck)
		gotNewState, _, err := a.allocateToken(ctx, p, nil, nil, metrics)
		assert.NoError(t, err)
		if diff := deep.Equal(gotNewState, &State{
			AllocationTokenRequestStartTime: tNow,
			Phase:                           PhaseAllocationTokenAcquired,
		}); len(diff) > 0 {
			t.Errorf("allocateToken() gotNewState = %v, Diff: %v", gotNewState, diff)
		}
	})

	t.Run("Allocation Successful", func(t *testing.T) {
		p.OnResourceRequirements(ctx, tCtx).Return("ns", core.ResourceConstraintsSpec{}, nil)
		a := newTokenAllocator(clck)
		gotNewState, _, err := a.allocateToken(ctx, p, tCtx, state, metrics)
		assert.NoError(t, err)
		if diff := deep.Equal(gotNewState, &State{
			AllocationTokenRequestStartTime: tNow,
			Phase:                           PhaseAllocationTokenAcquired,
		}); len(diff) > 0 {
			t.Errorf("allocateToken() gotNewState = %v, Diff: %v", gotNewState, diff)
		}
	})

	t.Run("Allocation Failed", func(t *testing.T) {
		tID := &mocks2.TaskExecutionID{}
		tID.OnGetGeneratedName().Return("abc2")

		tMeta := &mocks2.TaskExecutionMetadata{}
		tMeta.OnGetTaskExecutionID().Return(tID)

		rm := &mocks2.ResourceManager{}
		rm.OnAllocateResourceMatch(ctx, core.ResourceNamespace("ns"), "abc", mock.Anything).Return(core.AllocationStatusGranted, nil)
		rm.OnAllocateResourceMatch(ctx, core.ResourceNamespace("ns"), "abc2", mock.Anything).Return(core.AllocationStatusExhausted, nil)

		tCtx := &mocks2.TaskExecutionContext{}
		tCtx.OnTaskExecutionMetadata().Return(tMeta)
		tCtx.OnResourceManager().Return(rm)

		p.OnResourceRequirements(ctx, tCtx).Return("ns", core.ResourceConstraintsSpec{}, nil)
		a := newTokenAllocator(clck)
		gotNewState, _, err := a.allocateToken(ctx, p, tCtx, state, metrics)
		assert.NoError(t, err)
		if diff := deep.Equal(gotNewState, &State{
			AllocationTokenRequestStartTime: tNow,
			Phase:                           PhaseNotStarted,
		}); len(diff) > 0 {
			t.Errorf("allocateToken() gotNewState = %v, Diff: %v", gotNewState, diff)
		}
	})
}

func Test_releaseToken(t *testing.T) {
	ctx := context.Background()
	metrics := newMetrics(promutils.NewTestScope())

	tNow := time.Now()
	clck := testing2.NewFakeClock(tNow)

	tID := &mocks2.TaskExecutionID{}
	tID.OnGetGeneratedName().Return("abc")

	tMeta := &mocks2.TaskExecutionMetadata{}
	tMeta.OnGetTaskExecutionID().Return(tID)

	rm := &mocks2.ResourceManager{}
	rm.OnAllocateResourceMatch(ctx, core.ResourceNamespace("ns"), "abc", mock.Anything).Return(core.AllocationStatusGranted, nil)
	rm.OnAllocateResourceMatch(ctx, core.ResourceNamespace("ns"), "abc2", mock.Anything).Return(core.AllocationStatusExhausted, nil)
	rm.OnReleaseResource(ctx, core.ResourceNamespace("ns"), "abc").Return(nil)

	tCtx := &mocks2.TaskExecutionContext{}
	tCtx.OnTaskExecutionMetadata().Return(tMeta)
	tCtx.OnResourceManager().Return(rm)

	p := newPluginWithProperties(webapi.PluginConfig{
		ResourceQuotas: map[core.ResourceNamespace]int{
			"ns": 1,
		},
	})
	p.OnResourceRequirements(ctx, tCtx).Return("ns", core.ResourceConstraintsSpec{}, nil)

	a := newTokenAllocator(clck)
	assert.NoError(t, a.releaseToken(ctx, p, tCtx, metrics))
}
