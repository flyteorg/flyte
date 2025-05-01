package webapi

import (
	"context"
	"testing"
	"time"

	"github.com/go-test/deep"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	testing2 "k8s.io/utils/clock/testing"

	"github.com/flyteorg/flyte/flyteplugins/go/tasks/pluginmachinery/core"
	mocks2 "github.com/flyteorg/flyte/flyteplugins/go/tasks/pluginmachinery/core/mocks"
	"github.com/flyteorg/flyte/flyteplugins/go/tasks/pluginmachinery/webapi"
	"github.com/flyteorg/flyte/flyteplugins/go/tasks/pluginmachinery/webapi/mocks"
	"github.com/flyteorg/flyte/flytestdlib/contextutils"
	"github.com/flyteorg/flyte/flytestdlib/promutils"
	"github.com/flyteorg/flyte/flytestdlib/promutils/labeled"
)

func init() {
	labeled.SetMetricKeys(contextutils.NamespaceKey)
}

func newPluginWithProperties(properties webapi.PluginConfig) *mocks.AsyncPlugin {
	m := &mocks.AsyncPlugin{}
	m.EXPECT().GetConfig().Return(properties)
	return m
}

func Test_allocateToken(t *testing.T) {
	ctx := context.Background()
	metrics := newMetrics(promutils.NewTestScope())

	tNow := time.Now()
	clck := testing2.NewFakeClock(tNow)

	tID := &mocks2.TaskExecutionID{}
	tID.EXPECT().GetGeneratedName().Return("abc")

	tMeta := &mocks2.TaskExecutionMetadata{}
	tMeta.EXPECT().GetTaskExecutionID().Return(tID)

	rm := &mocks2.ResourceManager{}
	rm.EXPECT().AllocateResource(ctx, core.ResourceNamespace("ns"), "abc", mock.Anything).Return(core.AllocationStatusGranted, nil)
	rm.EXPECT().AllocateResource(ctx, core.ResourceNamespace("ns"), "abc2", mock.Anything).Return(core.AllocationStatusExhausted, nil)

	tCtx := &mocks2.TaskExecutionContext{}
	tCtx.EXPECT().TaskExecutionMetadata().Return(tMeta)
	tCtx.EXPECT().ResourceManager().Return(rm)

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
		p.EXPECT().ResourceRequirements(ctx, tCtx).Return("ns", core.ResourceConstraintsSpec{}, nil)
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
		tID.EXPECT().GetGeneratedName().Return("abc2")

		tMeta := &mocks2.TaskExecutionMetadata{}
		tMeta.EXPECT().GetTaskExecutionID().Return(tID)

		rm := &mocks2.ResourceManager{}
		rm.EXPECT().AllocateResource(ctx, core.ResourceNamespace("ns"), "abc", mock.Anything).Return(core.AllocationStatusGranted, nil)
		rm.EXPECT().AllocateResource(ctx, core.ResourceNamespace("ns"), "abc2", mock.Anything).Return(core.AllocationStatusExhausted, nil)

		tCtx := &mocks2.TaskExecutionContext{}
		tCtx.EXPECT().TaskExecutionMetadata().Return(tMeta)
		tCtx.EXPECT().ResourceManager().Return(rm)

		p.EXPECT().ResourceRequirements(ctx, tCtx).Return("ns", core.ResourceConstraintsSpec{}, nil)
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
	tID.EXPECT().GetGeneratedName().Return("abc")

	tMeta := &mocks2.TaskExecutionMetadata{}
	tMeta.EXPECT().GetTaskExecutionID().Return(tID)

	rm := &mocks2.ResourceManager{}
	rm.EXPECT().AllocateResource(ctx, core.ResourceNamespace("ns"), "abc", mock.Anything).Return(core.AllocationStatusGranted, nil)
	rm.EXPECT().AllocateResource(ctx, core.ResourceNamespace("ns"), "abc2", mock.Anything).Return(core.AllocationStatusExhausted, nil)
	rm.EXPECT().ReleaseResource(ctx, core.ResourceNamespace("ns"), "abc").Return(nil)

	tCtx := &mocks2.TaskExecutionContext{}
	tCtx.EXPECT().TaskExecutionMetadata().Return(tMeta)
	tCtx.EXPECT().ResourceManager().Return(rm)

	p := newPluginWithProperties(webapi.PluginConfig{
		ResourceQuotas: map[core.ResourceNamespace]int{
			"ns": 1,
		},
	})
	p.EXPECT().ResourceRequirements(ctx, tCtx).Return("ns", core.ResourceConstraintsSpec{}, nil)

	a := newTokenAllocator(clck)
	assert.NoError(t, a.releaseToken(ctx, p, tCtx, metrics))
}
