package hive

import (
	"context"
	"fmt"
	"net/url"
	"testing"
	"time"

	"github.com/flyteorg/flyteplugins/go/tasks/pluginmachinery/io"
	ioMock "github.com/flyteorg/flyteplugins/go/tasks/pluginmachinery/io/mocks"

	"github.com/flyteorg/flytestdlib/contextutils"
	"github.com/flyteorg/flytestdlib/promutils/labeled"

	"github.com/flyteorg/flytestdlib/promutils"

	idlCore "github.com/flyteorg/flyteidl/gen/pb-go/flyteidl/core"
	"github.com/flyteorg/flyteidl/gen/pb-go/flyteidl/plugins"

	mocks2 "github.com/flyteorg/flytestdlib/cache/mocks"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"

	"github.com/flyteorg/flyteplugins/go/tasks/pluginmachinery/core"
	"github.com/flyteorg/flyteplugins/go/tasks/pluginmachinery/core/mocks"
	pluginsCoreMocks "github.com/flyteorg/flyteplugins/go/tasks/pluginmachinery/core/mocks"
	"github.com/flyteorg/flyteplugins/go/tasks/plugins/hive/client"
	quboleMocks "github.com/flyteorg/flyteplugins/go/tasks/plugins/hive/client/mocks"
	"github.com/flyteorg/flyteplugins/go/tasks/plugins/hive/config"
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
	tCtx := GetMockTaskExecutionContext()

	query, cluster, tags, timeout, taskName, err := GetQueryInfo(ctx, tCtx)
	assert.NoError(t, err)
	assert.Equal(t, "select 'one'", query)
	assert.Equal(t, "default", cluster)
	assert.Equal(t, []string{"flyte_plugin_test", "ns:test-namespace", "label-1:val1"}, tags)
	assert.Equal(t, 500, int(timeout))
	assert.Equal(t, "sample_hive_task_test_name", taskName)
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
	taskLog := ConstructTaskLog(ExecutionState{CommandID: "123", URI: u.String()})
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
		CommandID:        "123",
		SyncFailureCount: 0,
		URI:              u.String(),
	}

	taskInfo := ConstructTaskInfo(e)
	assert.Equal(t, "https://wellness.qubole.com/v2/analyze?command_id=123", taskInfo.Logs[0].Uri)
	assert.Len(t, taskInfo.ExternalResources, 1)
	assert.Equal(t, taskInfo.ExternalResources[0].ExternalID, "123")
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

	t.Run("Write outputs file", func(t *testing.T) {
		e := ExecutionState{
			Phase: PhaseWriteOutputFile,
		}
		phaseInfo := MapExecutionStateToPhaseInfo(e, c)
		assert.Equal(t, core.PhaseRunning, phaseInfo.Phase())
		assert.Equal(t, uint32(1), phaseInfo.Version())
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
		mockMetrics := getQuboleHiveExecutorMetrics(promutils.NewTestScope())
		state, err := GetAllocationToken(ctx, tCtx, mockCurrentState, mockMetrics)
		assert.NoError(t, err)
		assert.Equal(t, PhaseQueued, state.Phase)
	})

	t.Run("exhausted", func(t *testing.T) {
		tCtx := GetMockTaskExecutionContext()
		mockResourceManager := tCtx.ResourceManager()
		x := mockResourceManager.(*mocks.ResourceManager)
		x.On("AllocateResource", mock.Anything, mock.Anything, mock.Anything, mock.Anything).
			Return(core.AllocationStatusExhausted, nil)

		mockCurrentState := ExecutionState{AllocationTokenRequestStartTime: time.Now()}
		mockMetrics := getQuboleHiveExecutorMetrics(promutils.NewTestScope())
		state, err := GetAllocationToken(ctx, tCtx, mockCurrentState, mockMetrics)
		assert.NoError(t, err)
		assert.Equal(t, PhaseNotStarted, state.Phase)
	})

	t.Run("namespace exhausted", func(t *testing.T) {
		tCtx := GetMockTaskExecutionContext()
		mockResourceManager := tCtx.ResourceManager()
		x := mockResourceManager.(*mocks.ResourceManager)
		x.On("AllocateResource", mock.Anything, mock.Anything, mock.Anything, mock.Anything).
			Return(core.AllocationStatusNamespaceQuotaExceeded, nil)

		mockCurrentState := ExecutionState{AllocationTokenRequestStartTime: time.Now()}
		mockMetrics := getQuboleHiveExecutorMetrics(promutils.NewTestScope())
		state, err := GetAllocationToken(ctx, tCtx, mockCurrentState, mockMetrics)
		assert.NoError(t, err)
		assert.Equal(t, PhaseNotStarted, state.Phase)
	})

	t.Run("Request start time, if empty in current state, should be set", func(t *testing.T) {
		tCtx := GetMockTaskExecutionContext()
		mockResourceManager := tCtx.ResourceManager()
		x := mockResourceManager.(*mocks.ResourceManager)
		x.On("AllocateResource", mock.Anything, mock.Anything, mock.Anything, mock.Anything).
			Return(core.AllocationStatusNamespaceQuotaExceeded, nil)

		mockCurrentState := ExecutionState{}
		mockMetrics := getQuboleHiveExecutorMetrics(promutils.NewTestScope())
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
		mockMetrics := getQuboleHiveExecutorMetrics(promutils.NewTestScope())
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
		mockQubole := &quboleMocks.QuboleClient{}
		mockQubole.On("KillCommand", mock.Anything, mock.MatchedBy(func(commandId string) bool {
			return commandId == "123456"
		}), mock.Anything).Run(func(_ mock.Arguments) {
			x = true
		}).Return(nil)

		err := Abort(ctx, GetMockTaskExecutionContext(), ExecutionState{Phase: PhaseSubmitted, CommandID: "123456"}, mockQubole, "fake-key")
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
			CommandID: "123456",
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

	err := Finalize(ctx, tCtx, state, getQuboleHiveExecutorMetrics(promutils.NewTestScope()))
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
		Identifier:     "my_wf_exec_project:my_wf_exec_domain:my_wf_exec_name",
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
		mock.Anything, mock.Anything, mock.Anything).Run(func(_ mock.Arguments) {
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
	assert.Equal(t, "453298043", newState.CommandID)
	assert.True(t, getOrCreateCalled)
	assert.True(t, quboleCalled)
}

func TestWriteOutputs(t *testing.T) {
	ctx := context.Background()
	tCtx := GetMockTaskExecutionContext()
	tCtx.OutputWriter().(*ioMock.OutputWriter).On("Put", mock.Anything, mock.Anything).Return(nil).Run(func(arguments mock.Arguments) {
		reader := arguments.Get(1).(io.OutputReader)
		literals, err1, err2 := reader.Read(context.Background())
		assert.Nil(t, err1)
		assert.NoError(t, err2)
		assert.NotNil(t, literals.Literals["results"].GetScalar().GetSchema())
	})

	state := ExecutionState{}
	newState, _ := WriteOutputs(ctx, tCtx, state)
	fmt.Println(newState)
}

func createMockQuboleCfg() *config.Config {
	return &config.Config{
		DefaultClusterLabel: "default",
		ClusterConfigs: []config.ClusterConfig{
			{PrimaryLabel: "primary A", Labels: []string{"primary A", "A", "label A", "A-prod"}, Limit: 10},
			{PrimaryLabel: "primary B", Labels: []string{"B"}, Limit: 10},
			{PrimaryLabel: "primary C", Labels: []string{"C-prod"}, Limit: 1},
		},
		DestinationClusterConfigs: []config.DestinationClusterConfig{
			{Project: "project A", Domain: "domain X", ClusterLabel: "A-prod"},
			{Project: "project A", Domain: "domain Y", ClusterLabel: "A"},
			{Project: "project A", Domain: "domain Z", ClusterLabel: "B"},
			{Project: "project C", Domain: "domain X", ClusterLabel: "C-prod"},
		},
	}
}

func Test_mapLabelToPrimaryLabel(t *testing.T) {
	ctx := context.TODO()
	mockQuboleCfg := createMockQuboleCfg()

	type args struct {
		ctx       context.Context
		quboleCfg *config.Config
		label     string
	}
	tests := []struct {
		name      string
		args      args
		want      string
		wantFound bool
	}{
		{name: "Label has a mapping", args: args{ctx: ctx, quboleCfg: mockQuboleCfg, label: "A-prod"}, want: "primary A", wantFound: true},
		{name: "Label has a typo", args: args{ctx: ctx, quboleCfg: mockQuboleCfg, label: "a"}, want: DefaultClusterPrimaryLabel, wantFound: false},
		{name: "Label has a mapping 2", args: args{ctx: ctx, quboleCfg: mockQuboleCfg, label: "C-prod"}, want: "primary C", wantFound: true},
		{name: "Label has a typo 2", args: args{ctx: ctx, quboleCfg: mockQuboleCfg, label: "C_prod"}, want: DefaultClusterPrimaryLabel, wantFound: false},
		{name: "Label has a mapping 3", args: args{ctx: ctx, quboleCfg: mockQuboleCfg, label: "primary A"}, want: "primary A", wantFound: true},
		{name: "Label has no mapping", args: args{ctx: ctx, quboleCfg: mockQuboleCfg, label: "D"}, want: DefaultClusterPrimaryLabel, wantFound: false},
		{name: "Label is an empty string", args: args{ctx: ctx, quboleCfg: mockQuboleCfg, label: ""}, want: DefaultClusterPrimaryLabel, wantFound: false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got, found := mapLabelToPrimaryLabel(tt.args.ctx, tt.args.quboleCfg, tt.args.label); got != tt.want || found != tt.wantFound {
				t.Errorf("mapLabelToPrimaryLabel() = (%v, %v), want (%v, %v)", got, found, tt.want, tt.wantFound)
			}
		})
	}
}

func createMockTaskExecutionContextWithProjectDomain(project string, domain string) *mocks.TaskExecutionContext {
	mockTaskExecutionContext := mocks.TaskExecutionContext{}
	taskExecID := &pluginsCoreMocks.TaskExecutionID{}
	taskExecID.OnGetID().Return(idlCore.TaskExecutionIdentifier{
		NodeExecutionId: &idlCore.NodeExecutionIdentifier{ExecutionId: &idlCore.WorkflowExecutionIdentifier{
			Project: project,
			Domain:  domain,
			Name:    "random name",
		}},
	})

	taskMetadata := &pluginsCoreMocks.TaskExecutionMetadata{}
	taskMetadata.OnGetTaskExecutionID().Return(taskExecID)
	mockTaskExecutionContext.On("TaskExecutionMetadata").Return(taskMetadata)
	return &mockTaskExecutionContext
}

func Test_getClusterPrimaryLabel(t *testing.T) {
	ctx := context.TODO()
	err := config.SetQuboleConfig(createMockQuboleCfg())
	assert.Nil(t, err)

	type args struct {
		ctx                  context.Context
		tCtx                 core.TaskExecutionContext
		clusterLabelOverride string
	}
	tests := []struct {
		name string
		args args
		want string
	}{
		{name: "Override is not empty + override has NO existing mapping + project-domain has an existing mapping", args: args{ctx: ctx, tCtx: createMockTaskExecutionContextWithProjectDomain("project A", "domain Z"), clusterLabelOverride: "AAAA"}, want: "primary B"},
		{name: "Override is not empty + override has NO existing mapping + project-domain has NO existing mapping", args: args{ctx: ctx, tCtx: createMockTaskExecutionContextWithProjectDomain("project A", "domain blah"), clusterLabelOverride: "blh"}, want: DefaultClusterPrimaryLabel},
		{name: "Override is not empty + override has an existing mapping + project-domain has NO existing mapping", args: args{ctx: ctx, tCtx: createMockTaskExecutionContextWithProjectDomain("project blah", "domain blah"), clusterLabelOverride: "C-prod"}, want: "primary C"},
		{name: "Override is not empty + override has an existing mapping + project-domain has an existing mapping", args: args{ctx: ctx, tCtx: createMockTaskExecutionContextWithProjectDomain("project A", "domain A"), clusterLabelOverride: "C-prod"}, want: "primary C"},
		{name: "Override is empty + project-domain has an existing mapping", args: args{ctx: ctx, tCtx: createMockTaskExecutionContextWithProjectDomain("project A", "domain X"), clusterLabelOverride: ""}, want: "primary A"},
		{name: "Override is empty + project-domain has an existing mapping2", args: args{ctx: ctx, tCtx: createMockTaskExecutionContextWithProjectDomain("project A", "domain Z"), clusterLabelOverride: ""}, want: "primary B"},
		{name: "Override is empty + project-domain has NO existing mapping", args: args{ctx: ctx, tCtx: createMockTaskExecutionContextWithProjectDomain("project A", "domain blah"), clusterLabelOverride: ""}, want: DefaultClusterPrimaryLabel},
		{name: "Override is empty + project-domain has NO existing mapping2", args: args{ctx: ctx, tCtx: createMockTaskExecutionContextWithProjectDomain("project blah", "domain X"), clusterLabelOverride: ""}, want: DefaultClusterPrimaryLabel},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := getClusterPrimaryLabel(tt.args.ctx, tt.args.tCtx, tt.args.clusterLabelOverride); got != tt.want {
				t.Errorf("getClusterPrimaryLabel() = %v, want %v", got, tt.want)
			}
		})
	}
}
