package hive

import (
	"context"
	"net/url"
	"testing"

	"github.com/lyft/flyteidl/gen/pb-go/flyteidl/plugins"

	mocks2 "github.com/lyft/flytestdlib/cache/mocks"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"

	"github.com/lyft/flyteplugins/go/tasks/pluginmachinery/core"
	"github.com/lyft/flyteplugins/go/tasks/pluginmachinery/core/mocks"
	pluginsCoreMocks "github.com/lyft/flyteplugins/go/tasks/pluginmachinery/core/mocks"
	"github.com/lyft/flyteplugins/go/tasks/plugins/hive/client"
	quboleMocks "github.com/lyft/flyteplugins/go/tasks/plugins/hive/client/mocks"
	"github.com/lyft/flyteplugins/go/tasks/plugins/hive/config"
)

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
			e := ExecutionState{Phase: tt.phase}
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
			e := ExecutionState{Phase: tt.phase}
			res := IsNotYetSubmitted(e)
			assert.Equal(t, tt.isNotYetSubmitted, res)
		})
	}
}

func TestGetQueryInfo(t *testing.T) {
	ctx := context.Background()

	taskTemplate := GetSingleHiveQueryTaskTemplate()
	mockTaskReader := &mocks.TaskReader{}
	mockTaskReader.On("Read", mock.Anything).Return(&taskTemplate, nil)

	mockTaskExecutionContext := mocks.TaskExecutionContext{}
	mockTaskExecutionContext.On("TaskReader").Return(mockTaskReader)

	taskMetadata := &pluginsCoreMocks.TaskExecutionMetadata{}
	taskMetadata.On("GetNamespace").Return("myproject-staging")
	taskMetadata.On("GetLabels").Return(map[string]string{"sample": "label"})
	mockTaskExecutionContext.On("TaskExecutionMetadata").Return(taskMetadata)

	query, cluster, tags, timeout, err := GetQueryInfo(ctx, &mockTaskExecutionContext)
	assert.NoError(t, err)
	assert.Equal(t, "select 'one'", query)
	assert.Equal(t, "default", cluster)
	assert.Equal(t, []string{"flyte_plugin_test", "ns:myproject-staging", "sample:label"}, tags)
	assert.Equal(t, 500, int(timeout))
}

func TestValidateQuboleHiveJob(t *testing.T) {
	hiveJob := plugins.QuboleHiveJob{
		ClusterLabel: "default",
		Tags:         []string{"flyte_plugin_test", "sample:label"},
		Query:        nil,
	}
	err := validateQuboleHiveJob(hiveJob)
	assert.Error(t, err)
}

func TestConstructTaskLog(t *testing.T) {
	expected := "https://wellness.qubole.com/v2/analyze?command_id=123"
	u, err := url.Parse(expected)
	assert.NoError(t, err)
	taskLog := ConstructTaskLog(ExecutionState{CommandId: "123", URI: u.String()})
	assert.Equal(t, expected, taskLog.Uri)
}

func TestConstructTaskInfo(t *testing.T) {
	empty := ConstructTaskInfo(ExecutionState{})
	assert.Nil(t, empty)

	expected := "https://wellness.qubole.com/v2/analyze?command_id=123"
	u, err := url.Parse(expected)
	assert.NoError(t, err)

	e := ExecutionState{
		Phase:            PhaseQuerySucceeded,
		CommandId:        "123",
		SyncFailureCount: 0,
		URI:              u.String(),
	}

	taskInfo := ConstructTaskInfo(e)
	assert.Equal(t, "https://wellness.qubole.com/v2/analyze?command_id=123", taskInfo.Logs[0].Uri)
}

func TestMapExecutionStateToPhaseInfo(t *testing.T) {
	c := client.NewQuboleClient(config.GetQuboleConfig())
	t.Run("NotStarted", func(t *testing.T) {
		e := ExecutionState{
			Phase: PhaseNotStarted,
		}
		phaseInfo := MapExecutionStateToPhaseInfo(e, c)
		assert.Equal(t, core.PhaseNotReady, phaseInfo.Phase())
	})

	t.Run("Queued", func(t *testing.T) {
		e := ExecutionState{
			Phase:                PhaseQueued,
			CreationFailureCount: 0,
		}
		phaseInfo := MapExecutionStateToPhaseInfo(e, c)
		assert.Equal(t, core.PhaseQueued, phaseInfo.Phase())

		e = ExecutionState{
			Phase:                PhaseQueued,
			CreationFailureCount: 100,
		}
		phaseInfo = MapExecutionStateToPhaseInfo(e, c)
		assert.Equal(t, core.PhaseRetryableFailure, phaseInfo.Phase())

	})

	t.Run("Submitted", func(t *testing.T) {
		e := ExecutionState{
			Phase: PhaseSubmitted,
		}
		phaseInfo := MapExecutionStateToPhaseInfo(e, c)
		assert.Equal(t, core.PhaseRunning, phaseInfo.Phase())
	})
}

func TestGetAllocationToken(t *testing.T) {
	ctx := context.Background()

	t.Run("allocation granted", func(t *testing.T) {
		tCtx := GetMockTaskExecutionContext()
		mockResourceManager := tCtx.ResourceManager()
		x := mockResourceManager.(*mocks.ResourceManager)
		x.On("AllocateResource", mock.Anything, mock.Anything, mock.Anything).
			Return(core.AllocationStatusGranted, nil)

		state, err := GetAllocationToken(ctx, quboleResourceNamespace, tCtx)
		assert.NoError(t, err)
		assert.Equal(t, PhaseQueued, state.Phase)
	})

	t.Run("exhausted", func(t *testing.T) {
		tCtx := GetMockTaskExecutionContext()
		mockResourceManager := tCtx.ResourceManager()
		x := mockResourceManager.(*mocks.ResourceManager)
		x.On("AllocateResource", mock.Anything, mock.Anything, mock.Anything).
			Return(core.AllocationStatusExhausted, nil)

		state, err := GetAllocationToken(ctx, quboleResourceNamespace, tCtx)
		assert.NoError(t, err)
		assert.Equal(t, PhaseNotStarted, state.Phase)
	})

	t.Run("namespace exhausted", func(t *testing.T) {
		tCtx := GetMockTaskExecutionContext()
		mockResourceManager := tCtx.ResourceManager()
		x := mockResourceManager.(*mocks.ResourceManager)
		x.On("AllocateResource", mock.Anything, mock.Anything, mock.Anything).
			Return(core.AllocationStatusNamespaceQuotaExceeded, nil)

		state, err := GetAllocationToken(ctx, quboleResourceNamespace, tCtx)
		assert.NoError(t, err)
		assert.Equal(t, PhaseNotStarted, state.Phase)
	})
}

func TestAbort(t *testing.T) {
	ctx := context.Background()

	t.Run("Terminate called when not in terminal state", func(t *testing.T) {
		var x = false
		mockQubole := &quboleMocks.QuboleClient{}
		mockQubole.On("KillCommand", mock.Anything, mock.MatchedBy(func(commandId string) bool {
			return commandId == "123456"
		}), mock.Anything).Run(func(_ mock.Arguments) {
			x = true
		}).Return(nil)

		err := Abort(ctx, GetMockTaskExecutionContext(), ExecutionState{Phase: PhaseSubmitted, CommandId: "123456"}, mockQubole, "fake-key")
		assert.NoError(t, err)
		assert.True(t, x)
	})

	t.Run("Terminate not called when in terminal state", func(t *testing.T) {
		var x = false
		mockQubole := &quboleMocks.QuboleClient{}
		mockQubole.On("KillCommand", mock.Anything, mock.Anything, mock.Anything).Run(func(_ mock.Arguments) {
			x = true
		}).Return(nil)

		err := Abort(ctx, GetMockTaskExecutionContext(), ExecutionState{
			Phase:     PhaseQuerySucceeded,
			CommandId: "123456",
		}, mockQubole, "fake-key")
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

	err := Finalize(ctx, tCtx, quboleResourceNamespace, state)
	assert.NoError(t, err)
	assert.True(t, called)
}

func TestMonitorQuery(t *testing.T) {
	ctx := context.Background()
	tCtx := GetMockTaskExecutionContext()
	state := ExecutionState{
		Phase: PhaseSubmitted,
	}
	var getOrCreateCalled = false
	mockCache := &mocks2.AutoRefresh{}
	mockCache.OnGetOrCreateMatch("my_wf_exec_project:my_wf_exec_domain:my_wf_exec_name", mock.Anything).Return(ExecutionStateCacheItem{
		ExecutionState: ExecutionState{Phase: PhaseQuerySucceeded},
		Id:             "my_wf_exec_project:my_wf_exec_domain:my_wf_exec_name",
	}, nil).Run(func(_ mock.Arguments) {
		getOrCreateCalled = true
	})

	newState, err := MonitorQuery(ctx, tCtx, state, mockCache)
	assert.NoError(t, err)
	assert.True(t, getOrCreateCalled)
	assert.Equal(t, PhaseQuerySucceeded, newState.Phase)
}

func TestKickOffQuery(t *testing.T) {
	ctx := context.Background()
	tCtx := GetMockTaskExecutionContext()

	var quboleCalled = false
	quboleCommandDetails := &client.QuboleCommandDetails{
		ID:     int64(453298043),
		Status: client.QuboleStatusWaiting,
	}
	mockQubole := &quboleMocks.QuboleClient{}
	mockQubole.OnExecuteHiveCommandMatch(mock.Anything, mock.Anything, mock.Anything, mock.Anything,
		mock.Anything, mock.Anything).Run(func(_ mock.Arguments) {
		quboleCalled = true
	}).Return(quboleCommandDetails, nil)

	var getOrCreateCalled = false
	mockCache := &mocks2.AutoRefresh{}
	mockCache.OnGetOrCreate(mock.Anything, mock.Anything).Run(func(_ mock.Arguments) {
		getOrCreateCalled = true
	}).Return(ExecutionStateCacheItem{}, nil)

	state := ExecutionState{}
	newState, err := KickOffQuery(ctx, tCtx, state, mockQubole, mockCache, config.GetQuboleConfig())
	assert.NoError(t, err)
	assert.Equal(t, PhaseSubmitted, newState.Phase)
	assert.Equal(t, "453298043", newState.CommandId)
	assert.True(t, getOrCreateCalled)
	assert.True(t, quboleCalled)
}
