package presto

import (
	"context"
	"net/url"
	"testing"
	"time"

	"github.com/flyteorg/flyteplugins/go/tasks/pluginmachinery/core"
	"github.com/flyteorg/flyteplugins/go/tasks/plugins/presto/client"
	prestoMocks "github.com/flyteorg/flyteplugins/go/tasks/plugins/presto/client/mocks"
	"github.com/flyteorg/flyteplugins/go/tasks/plugins/presto/config"
	mocks2 "github.com/flyteorg/flytestdlib/cache/mocks"
	stdConfig "github.com/flyteorg/flytestdlib/config"
	"github.com/flyteorg/flytestdlib/contextutils"
	"github.com/flyteorg/flytestdlib/promutils"
	"github.com/flyteorg/flytestdlib/promutils/labeled"

	"github.com/flyteorg/flyteidl/gen/pb-go/flyteidl/plugins"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"

	"github.com/flyteorg/flyteplugins/go/tasks/pluginmachinery/core/mocks"
)

func init() {
	labeled.SetMetricKeys(contextutils.NamespaceKey)
}

func TestInTerminalState(t *testing.T) {
	var stateTests = []struct {
		phase      ExecutionPhase
		isTerminal bool
	}{
		{phase: PhaseNotStarted, isTerminal: false},
		{phase: PhaseQueued, isTerminal: false},
		{phase: PhaseSubmitted, isTerminal: false},
		{phase: PhaseQuerySucceeded, isTerminal: true},
		{phase: PhaseQueryFailed, isTerminal: true},
	}

	for _, tt := range stateTests {
		t.Run(tt.phase.String(), func(t *testing.T) {
			e := ExecutionState{CurrentPhase: tt.phase}
			res := InTerminalState(e)
			assert.Equal(t, tt.isTerminal, res)
		})
	}
}

func TestIsNotYetSubmitted(t *testing.T) {
	var stateTests = []struct {
		phase             ExecutionPhase
		isNotYetSubmitted bool
	}{
		{phase: PhaseNotStarted, isNotYetSubmitted: true},
		{phase: PhaseQueued, isNotYetSubmitted: true},
		{phase: PhaseSubmitted, isNotYetSubmitted: false},
		{phase: PhaseQuerySucceeded, isNotYetSubmitted: false},
		{phase: PhaseQueryFailed, isNotYetSubmitted: false},
	}

	for _, tt := range stateTests {
		t.Run(tt.phase.String(), func(t *testing.T) {
			e := ExecutionState{CurrentPhase: tt.phase}
			res := IsNotYetSubmitted(e)
			assert.Equal(t, tt.isNotYetSubmitted, res)
		})
	}
}

func TestValidatePrestoStatement(t *testing.T) {
	prestoQuery := plugins.PrestoQuery{
		RoutingGroup: "adhoc",
		Catalog:      "hive",
		Schema:       "city",
		Statement:    "",
	}
	err := validatePrestoStatement(prestoQuery)
	assert.Error(t, err)
}

func TestConstructTaskLog(t *testing.T) {
	expected := "https://prestoproxy-internal.flyteorg.net:443"
	u, err := url.Parse(expected)
	assert.NoError(t, err)
	taskLog := ConstructTaskLog(ExecutionState{CommandID: "123", URI: u.String()})
	assert.Equal(t, expected, taskLog.Uri)
}

func TestConstructTaskInfo(t *testing.T) {
	empty := ConstructTaskInfo(ExecutionState{})
	assert.Nil(t, empty)

	expected := "https://prestoproxy-internal.flyteorg.net:443"
	u, err := url.Parse(expected)
	assert.NoError(t, err)

	e := ExecutionState{
		CurrentPhase:     PhaseQuerySucceeded,
		CommandID:        "123",
		SyncFailureCount: 0,
		URI:              u.String(),
	}

	taskInfo := ConstructTaskInfo(e)
	assert.Equal(t, "https://prestoproxy-internal.flyteorg.net:443", taskInfo.Logs[0].Uri)
	assert.Len(t, taskInfo.ExternalResources, 1)
	assert.Equal(t, taskInfo.ExternalResources[0].ExternalID, "123")
}

func TestMapExecutionStateToPhaseInfo(t *testing.T) {
	t.Run("NotStarted", func(t *testing.T) {
		e := ExecutionState{
			CurrentPhase: PhaseNotStarted,
		}
		phaseInfo := MapExecutionStateToPhaseInfo(e)
		assert.Equal(t, core.PhaseNotReady, phaseInfo.Phase())
	})

	t.Run("Queued", func(t *testing.T) {
		e := ExecutionState{
			CurrentPhase:         PhaseQueued,
			CreationFailureCount: 0,
		}
		phaseInfo := MapExecutionStateToPhaseInfo(e)
		assert.Equal(t, core.PhaseRunning, phaseInfo.Phase())

		e = ExecutionState{
			CurrentPhase:         PhaseQueued,
			CreationFailureCount: 100,
		}
		phaseInfo = MapExecutionStateToPhaseInfo(e)
		assert.Equal(t, core.PhaseRetryableFailure, phaseInfo.Phase())

	})

	t.Run("Submitted", func(t *testing.T) {
		e := ExecutionState{
			CurrentPhase: PhaseSubmitted,
		}
		phaseInfo := MapExecutionStateToPhaseInfo(e)
		assert.Equal(t, core.PhaseRunning, phaseInfo.Phase())
	})
}

func TestGetAllocationToken(t *testing.T) {
	ctx := context.Background()

	t.Run("allocation granted", func(t *testing.T) {
		tCtx := GetMockTaskExecutionContext()
		mockResourceManager := tCtx.ResourceManager()
		x := mockResourceManager.(*mocks.ResourceManager)
		x.On("AllocateResource", mock.Anything, mock.Anything, mock.Anything, mock.Anything).
			Return(core.AllocationStatusGranted, nil)

		mockCurrentState := ExecutionState{AllocationTokenRequestStartTime: time.Now()}
		mockMetrics := getPrestoExecutorMetrics(promutils.NewTestScope())
		state, err := GetAllocationToken(ctx, tCtx, mockCurrentState, mockMetrics)
		assert.NoError(t, err)
		assert.Equal(t, PhaseQueued, state.CurrentPhase)
	})

	t.Run("exhausted", func(t *testing.T) {
		tCtx := GetMockTaskExecutionContext()
		mockResourceManager := tCtx.ResourceManager()
		x := mockResourceManager.(*mocks.ResourceManager)
		x.On("AllocateResource", mock.Anything, mock.Anything, mock.Anything, mock.Anything).
			Return(core.AllocationStatusExhausted, nil)

		mockCurrentState := ExecutionState{AllocationTokenRequestStartTime: time.Now()}
		mockMetrics := getPrestoExecutorMetrics(promutils.NewTestScope())
		state, err := GetAllocationToken(ctx, tCtx, mockCurrentState, mockMetrics)
		assert.NoError(t, err)
		assert.Equal(t, PhaseNotStarted, state.CurrentPhase)
	})

	t.Run("namespace exhausted", func(t *testing.T) {
		tCtx := GetMockTaskExecutionContext()
		mockResourceManager := tCtx.ResourceManager()
		x := mockResourceManager.(*mocks.ResourceManager)
		x.On("AllocateResource", mock.Anything, mock.Anything, mock.Anything, mock.Anything).
			Return(core.AllocationStatusNamespaceQuotaExceeded, nil)

		mockCurrentState := ExecutionState{AllocationTokenRequestStartTime: time.Now()}
		mockMetrics := getPrestoExecutorMetrics(promutils.NewTestScope())
		state, err := GetAllocationToken(ctx, tCtx, mockCurrentState, mockMetrics)
		assert.NoError(t, err)
		assert.Equal(t, PhaseNotStarted, state.CurrentPhase)
	})

	t.Run("Request start time, if empty in current state, should be set", func(t *testing.T) {
		tCtx := GetMockTaskExecutionContext()
		mockResourceManager := tCtx.ResourceManager()
		x := mockResourceManager.(*mocks.ResourceManager)
		x.On("AllocateResource", mock.Anything, mock.Anything, mock.Anything, mock.Anything).
			Return(core.AllocationStatusNamespaceQuotaExceeded, nil)

		mockCurrentState := ExecutionState{}
		mockMetrics := getPrestoExecutorMetrics(promutils.NewTestScope())
		state, err := GetAllocationToken(ctx, tCtx, mockCurrentState, mockMetrics)
		assert.NoError(t, err)
		assert.Equal(t, state.AllocationTokenRequestStartTime.IsZero(), false)
	})

	t.Run("Request start time, if already set in current state, should be maintained", func(t *testing.T) {
		tCtx := GetMockTaskExecutionContext()
		mockResourceManager := tCtx.ResourceManager()
		x := mockResourceManager.(*mocks.ResourceManager)
		x.On("AllocateResource", mock.Anything, mock.Anything, mock.Anything, mock.Anything).
			Return(core.AllocationStatusGranted, nil)

		startTime := time.Now()
		mockCurrentState := ExecutionState{AllocationTokenRequestStartTime: startTime}
		mockMetrics := getPrestoExecutorMetrics(promutils.NewTestScope())
		state, err := GetAllocationToken(ctx, tCtx, mockCurrentState, mockMetrics)
		assert.NoError(t, err)
		assert.Equal(t, state.AllocationTokenRequestStartTime.IsZero(), false)
		assert.Equal(t, state.AllocationTokenRequestStartTime, startTime)
	})
}

func TestAbort(t *testing.T) {
	ctx := context.Background()

	t.Run("Terminate called when not in terminal state", func(t *testing.T) {
		var x = false

		mockPresto := &prestoMocks.PrestoClient{}
		mockPresto.On("KillCommand", mock.Anything, mock.MatchedBy(func(commandId string) bool {
			return commandId == "123456"
		}), mock.Anything).Run(func(_ mock.Arguments) {
			x = true
		}).Return(nil)

		err := Abort(ctx, ExecutionState{CurrentPhase: PhaseSubmitted, CommandID: "123456"}, mockPresto)
		assert.NoError(t, err)
		assert.True(t, x)
	})

	t.Run("Terminate not called when in terminal state", func(t *testing.T) {
		var x = false

		mockPresto := &prestoMocks.PrestoClient{}
		mockPresto.On("KillCommand", mock.Anything, mock.Anything).Run(func(_ mock.Arguments) {
			x = true
		}).Return(nil)

		err := Abort(ctx, ExecutionState{CurrentPhase: PhaseQuerySucceeded, CommandID: "123456"}, mockPresto)
		assert.NoError(t, err)
		assert.False(t, x)
	})
}

func TestFinalize(t *testing.T) {
	// Test that Finalize releases resources
	ctx := context.Background()
	tCtx := GetMockTaskExecutionContext()
	state := ExecutionState{}
	var called = false
	mockResourceManager := tCtx.ResourceManager()
	x := mockResourceManager.(*mocks.ResourceManager)
	x.On("ReleaseResource", mock.Anything, mock.Anything, mock.Anything).Run(func(_ mock.Arguments) {
		called = true
	}).Return(nil)

	err := Finalize(ctx, tCtx, state, getPrestoExecutorMetrics(promutils.NewTestScope()))
	assert.NoError(t, err)
	assert.True(t, called)
}

func TestMonitorQuery(t *testing.T) {
	ctx := context.Background()
	tCtx := GetMockTaskExecutionContext()
	state := ExecutionState{
		CurrentPhase: PhaseSubmitted,
	}
	var getOrCreateCalled = false
	mockCache := &mocks2.AutoRefresh{}
	mockCache.OnGetOrCreateMatch(mock.AnythingOfType("string"), mock.Anything).Return(ExecutionStateCacheItem{
		ExecutionState: ExecutionState{CurrentPhase: PhaseQuerySucceeded},
		Identifier:     "my_wf_exec_project:my_wf_exec_domain:my_wf_exec_name",
	}, nil).Run(func(_ mock.Arguments) {
		getOrCreateCalled = true
	})

	newState, err := MonitorQuery(ctx, tCtx, state, mockCache)
	assert.NoError(t, err)
	assert.True(t, getOrCreateCalled)
	assert.Equal(t, PhaseQuerySucceeded, newState.CurrentPhase)
}

func TestKickOffQuery(t *testing.T) {
	ctx := context.Background()
	tCtx := GetMockTaskExecutionContext()

	var prestoCalled = false

	prestoExecuteResponse := client.PrestoExecuteResponse{
		ID: "1234567",
	}
	mockPresto := &prestoMocks.PrestoClient{}
	mockPresto.OnExecuteCommandMatch(mock.Anything, mock.Anything, mock.Anything, mock.Anything,
		mock.Anything, mock.Anything).Run(func(_ mock.Arguments) {
		prestoCalled = true
	}).Return(prestoExecuteResponse, nil)
	var getOrCreateCalled = false
	mockCache := &mocks2.AutoRefresh{}
	mockCache.OnGetOrCreate(mock.Anything, mock.Anything).Run(func(_ mock.Arguments) {
		getOrCreateCalled = true
	}).Return(ExecutionStateCacheItem{}, nil)

	state := ExecutionState{}
	newState, err := KickOffQuery(ctx, tCtx, state, mockPresto, mockCache)
	assert.NoError(t, err)
	assert.Equal(t, PhaseSubmitted, newState.CurrentPhase)
	assert.Equal(t, "1234567", newState.CommandID)
	assert.True(t, getOrCreateCalled)
	assert.True(t, prestoCalled)
}

func createMockPrestoCfg() *config.Config {
	return &config.Config{
		Environment:         config.URLMustParse(""),
		DefaultRoutingGroup: "adhoc",
		RoutingGroupConfigs: []config.RoutingGroupConfig{{Name: "adhoc", Limit: 250}, {Name: "etl", Limit: 100}},
		RefreshCacheConfig: config.RefreshCacheConfig{
			Name:         "presto",
			SyncPeriod:   stdConfig.Duration{Duration: 3 * time.Second},
			Workers:      15,
			LruCacheSize: 2000,
		},
	}
}

func Test_mapLabelToPrimaryLabel(t *testing.T) {
	ctx := context.TODO()
	mockPrestoCfg := createMockPrestoCfg()

	type args struct {
		ctx          context.Context
		routingGroup string
		prestoCfg    *config.Config
	}
	tests := []struct {
		name string
		args args
		want string
	}{
		{name: "Routing group is found in configs", args: args{ctx: ctx, routingGroup: "etl", prestoCfg: mockPrestoCfg}, want: "etl"},
		{name: "Use routing group default when not found in configs", args: args{ctx: ctx, routingGroup: "test", prestoCfg: mockPrestoCfg}, want: mockPrestoCfg.DefaultRoutingGroup},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.want, resolveRoutingGroup(tt.args.ctx, tt.args.routingGroup, tt.args.prestoCfg))
		})
	}
}
