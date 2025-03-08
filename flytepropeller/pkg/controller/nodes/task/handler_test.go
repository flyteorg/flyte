package task

import (
	"bytes"
	"context"
	"fmt"
	"testing"

	"github.com/golang/protobuf/proto"
	"github.com/pkg/errors"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	v1 "k8s.io/api/core/v1"
	v12 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	k8sfake "k8s.io/client-go/kubernetes/fake"

	"github.com/flyteorg/flyte/flyteidl/clients/go/coreutils"
	"github.com/flyteorg/flyte/flyteidl/gen/pb-go/flyteidl/admin"
	"github.com/flyteorg/flyte/flyteidl/gen/pb-go/flyteidl/core"
	"github.com/flyteorg/flyte/flyteidl/gen/pb-go/flyteidl/event"
	"github.com/flyteorg/flyte/flyteplugins/go/tasks/pluginmachinery"
	pluginCatalogMocks "github.com/flyteorg/flyte/flyteplugins/go/tasks/pluginmachinery/catalog/mocks"
	pluginCore "github.com/flyteorg/flyte/flyteplugins/go/tasks/pluginmachinery/core"
	pluginCoreMocks "github.com/flyteorg/flyte/flyteplugins/go/tasks/pluginmachinery/core/mocks"
	"github.com/flyteorg/flyte/flyteplugins/go/tasks/pluginmachinery/io"
	ioMocks "github.com/flyteorg/flyte/flyteplugins/go/tasks/pluginmachinery/io/mocks"
	"github.com/flyteorg/flyte/flyteplugins/go/tasks/pluginmachinery/ioutils"
	pluginK8s "github.com/flyteorg/flyte/flyteplugins/go/tasks/pluginmachinery/k8s"
	pluginK8sMocks "github.com/flyteorg/flyte/flyteplugins/go/tasks/pluginmachinery/k8s/mocks"
	eventsErr "github.com/flyteorg/flyte/flytepropeller/events/errors"
	"github.com/flyteorg/flyte/flytepropeller/pkg/apis/flyteworkflow/v1alpha1"
	flyteMocks "github.com/flyteorg/flyte/flytepropeller/pkg/apis/flyteworkflow/v1alpha1/mocks"
	controllerConfig "github.com/flyteorg/flyte/flytepropeller/pkg/controller/config"
	"github.com/flyteorg/flyte/flytepropeller/pkg/controller/executors/mocks"
	"github.com/flyteorg/flyte/flytepropeller/pkg/controller/nodes/handler"
	"github.com/flyteorg/flyte/flytepropeller/pkg/controller/nodes/interfaces"
	nodeMocks "github.com/flyteorg/flyte/flytepropeller/pkg/controller/nodes/interfaces/mocks"
	"github.com/flyteorg/flyte/flytepropeller/pkg/controller/nodes/task/codex"
	"github.com/flyteorg/flyte/flytepropeller/pkg/controller/nodes/task/config"
	"github.com/flyteorg/flyte/flytepropeller/pkg/controller/nodes/task/fakeplugins"
	"github.com/flyteorg/flyte/flytepropeller/pkg/controller/nodes/task/resourcemanager"
	rmConfig "github.com/flyteorg/flyte/flytepropeller/pkg/controller/nodes/task/resourcemanager/config"
	"github.com/flyteorg/flyte/flytestdlib/contextutils"
	"github.com/flyteorg/flyte/flytestdlib/promutils"
	"github.com/flyteorg/flyte/flytestdlib/promutils/labeled"
	"github.com/flyteorg/flyte/flytestdlib/storage"
)

var eventConfig = &controllerConfig.EventConfig{
	RawOutputPolicy: controllerConfig.RawOutputPolicyInline,
}

const testClusterID = "C1"

func Test_task_setDefault(t *testing.T) {
	type fields struct {
		defaultPlugin pluginCore.Plugin
	}
	type args struct {
		p pluginCore.Plugin
	}

	other := &pluginCoreMocks.Plugin{}
	other.On("GetID").Return("other")

	def := &pluginCoreMocks.Plugin{}
	def.On("GetID").Return("default")

	tests := []struct {
		name           string
		fields         fields
		args           args
		wantErr        bool
		defaultChanged bool
	}{
		{"no-default", fields{nil}, args{p: other}, false, true},
		{"default-exists", fields{defaultPlugin: def}, args{p: other}, false, false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tk := &Handler{
				defaultPlugin:    tt.fields.defaultPlugin,
				connectorService: &pluginCore.ConnectorService{},
			}
			if err := tk.setDefault(context.TODO(), tt.args.p); (err != nil) != tt.wantErr {
				t.Errorf("Handler.setDefault() error = %v, wantErr %v", err, tt.wantErr)
			}
			if tt.defaultChanged {
				assert.Equal(t, tk.defaultPlugin, tt.args.p)
			} else {
				assert.NotEqual(t, tk.defaultPlugin, tt.args.p)
			}
		})
	}
}

type testPluginRegistry struct {
	core []pluginCore.PluginEntry
	k8s  []pluginK8s.PluginEntry
}

func (t testPluginRegistry) GetCorePlugins() []pluginCore.PluginEntry {
	return t.core
}

func (t testPluginRegistry) GetK8sPlugins() []pluginK8s.PluginEntry {
	return t.k8s
}

func Test_task_Setup(t *testing.T) {
	corePluginType := "core"
	corePlugin := &pluginCoreMocks.Plugin{}
	corePlugin.On("GetID").Return(corePluginType)
	corePlugin.EXPECT().GetProperties().Return(pluginCore.PluginProperties{})

	corePluginDefaultType := "coredefault"
	corePluginDefault := &pluginCoreMocks.Plugin{}
	corePluginDefault.On("GetID").Return(corePluginDefaultType)
	corePluginDefault.EXPECT().GetProperties().Return(pluginCore.PluginProperties{})

	k8sPluginType := "k8s"
	k8sPlugin := &pluginK8sMocks.Plugin{}
	k8sPlugin.EXPECT().GetProperties().Return(pluginK8s.PluginProperties{})

	k8sPluginDefaultType := "k8sdefault"
	k8sPluginDefault := &pluginK8sMocks.Plugin{}
	k8sPluginDefault.EXPECT().GetProperties().Return(pluginK8s.PluginProperties{})

	loadErrorPluginType := "loadError"

	corePluginEntry := pluginCore.PluginEntry{
		ID:                  corePluginType,
		RegisteredTaskTypes: []pluginCore.TaskType{corePluginType},
		LoadPlugin: func(ctx context.Context, iCtx pluginCore.SetupContext) (pluginCore.Plugin, error) {
			return corePlugin, nil
		},
	}
	corePluginEntryDefault := pluginCore.PluginEntry{
		IsDefault:           true,
		ID:                  corePluginDefaultType,
		RegisteredTaskTypes: []pluginCore.TaskType{corePluginDefaultType},
		LoadPlugin: func(ctx context.Context, iCtx pluginCore.SetupContext) (pluginCore.Plugin, error) {
			return corePluginDefault, nil
		},
	}
	k8sPluginEntry := pluginK8s.PluginEntry{
		ID:                  k8sPluginType,
		Plugin:              k8sPlugin,
		RegisteredTaskTypes: []pluginCore.TaskType{k8sPluginType},
		ResourceToWatch:     &v1.Pod{},
	}
	k8sPluginEntryDefault := pluginK8s.PluginEntry{
		IsDefault:           true,
		ID:                  k8sPluginDefaultType,
		Plugin:              k8sPluginDefault,
		RegisteredTaskTypes: []pluginCore.TaskType{k8sPluginDefaultType},
		ResourceToWatch:     &v1.Pod{},
	}
	loadErrorPluginEntry := pluginCore.PluginEntry{
		ID:                  loadErrorPluginType,
		RegisteredTaskTypes: []pluginCore.TaskType{loadErrorPluginType},
		LoadPlugin: func(ctx context.Context, iCtx pluginCore.SetupContext) (pluginCore.Plugin, error) {
			return nil, fmt.Errorf("test")
		},
	}

	type wantFields struct {
		pluginIDs       map[pluginCore.TaskType]string
		defaultPluginID string
	}
	tests := []struct {
		name                string
		registry            PluginRegistryIface
		enabledPlugins      []string
		defaultForTaskTypes map[string]string
		fields              wantFields
		wantErr             bool
	}{
		{"no-plugins", testPluginRegistry{}, []string{}, map[string]string{}, wantFields{}, false},
		{"no-default-only-core", testPluginRegistry{
			core: []pluginCore.PluginEntry{corePluginEntry}, k8s: []pluginK8s.PluginEntry{},
		}, []string{corePluginType}, map[string]string{
			corePluginType: corePluginType},
			wantFields{
				pluginIDs: map[pluginCore.TaskType]string{corePluginType: corePluginType},
			}, false},
		{"no-default-only-k8s", testPluginRegistry{
			core: []pluginCore.PluginEntry{}, k8s: []pluginK8s.PluginEntry{k8sPluginEntry},
		}, []string{k8sPluginType}, map[string]string{
			k8sPluginType: k8sPluginType},
			wantFields{
				pluginIDs: map[pluginCore.TaskType]string{k8sPluginType: k8sPluginType},
			}, false},
		{"no-default", testPluginRegistry{}, []string{corePluginType, k8sPluginType}, map[string]string{
			corePluginType: corePluginType,
			k8sPluginType:  k8sPluginType,
		}, wantFields{
			pluginIDs: map[pluginCore.TaskType]string{},
		}, false},
		{"only-default-core", testPluginRegistry{
			core: []pluginCore.PluginEntry{corePluginEntry, corePluginEntryDefault}, k8s: []pluginK8s.PluginEntry{k8sPluginEntry},
		}, []string{corePluginType, corePluginDefaultType, k8sPluginType}, map[string]string{
			corePluginType:        corePluginType,
			corePluginDefaultType: corePluginDefaultType,
			k8sPluginType:         k8sPluginType,
		}, wantFields{
			pluginIDs:       map[pluginCore.TaskType]string{corePluginType: corePluginType, corePluginDefaultType: corePluginDefaultType, k8sPluginType: k8sPluginType},
			defaultPluginID: corePluginDefaultType,
		}, false},
		{"only-default-k8s", testPluginRegistry{
			core: []pluginCore.PluginEntry{corePluginEntry}, k8s: []pluginK8s.PluginEntry{k8sPluginEntryDefault},
		}, []string{corePluginType, k8sPluginDefaultType}, map[string]string{
			corePluginType:       corePluginType,
			k8sPluginDefaultType: k8sPluginDefaultType,
		}, wantFields{
			pluginIDs:       map[pluginCore.TaskType]string{corePluginType: corePluginType, k8sPluginDefaultType: k8sPluginDefaultType},
			defaultPluginID: k8sPluginDefaultType,
		}, false},
		{"default-both", testPluginRegistry{
			core: []pluginCore.PluginEntry{corePluginEntry, corePluginEntryDefault}, k8s: []pluginK8s.PluginEntry{k8sPluginEntry, k8sPluginEntryDefault},
		}, []string{corePluginType, corePluginDefaultType, k8sPluginType, k8sPluginDefaultType}, map[string]string{
			corePluginType:        corePluginType,
			corePluginDefaultType: corePluginDefaultType,
			k8sPluginType:         k8sPluginType,
			k8sPluginDefaultType:  k8sPluginDefaultType,
		}, wantFields{
			pluginIDs:       map[pluginCore.TaskType]string{corePluginType: corePluginType, corePluginDefaultType: corePluginDefaultType, k8sPluginType: k8sPluginType, k8sPluginDefaultType: k8sPluginDefaultType},
			defaultPluginID: corePluginDefaultType,
		}, false},
		{"partial-default-task-types",
			testPluginRegistry{
				core: []pluginCore.PluginEntry{corePluginEntry},
				k8s:  []pluginK8s.PluginEntry{k8sPluginEntry},
			},
			[]string{corePluginType, k8sPluginType},
			map[string]string{corePluginType: corePluginType},
			wantFields{
				pluginIDs: map[pluginCore.TaskType]string{
					corePluginType: corePluginType,
					k8sPluginType:  k8sPluginType,
				},
			},
			false},
		{"load-error",
			testPluginRegistry{
				core: []pluginCore.PluginEntry{loadErrorPluginEntry},
				k8s:  []pluginK8s.PluginEntry{},
			},
			[]string{loadErrorPluginType},
			map[string]string{corePluginType: loadErrorPluginType},
			wantFields{},
			true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			sCtx := &nodeMocks.SetupContext{}
			fakeKubeClient := mocks.NewFakeKubeClient()
			mockClientset := k8sfake.NewSimpleClientset()
			sCtx.On("KubeClient").Return(fakeKubeClient)
			sCtx.On("OwnerKind").Return("test")
			sCtx.On("EnqueueOwner").Return(pluginCore.EnqueueOwner(func(name types.NamespacedName) error { return nil }))
			sCtx.On("MetricsScope").Return(promutils.NewTestScope())

			tk, err := New(context.TODO(), fakeKubeClient, mockClientset, &pluginCatalogMocks.Client{}, eventConfig, testClusterID, promutils.NewTestScope())
			tk.cfg.TaskPlugins.EnabledPlugins = tt.enabledPlugins
			tk.cfg.TaskPlugins.DefaultForTaskTypes = tt.defaultForTaskTypes
			assert.NoError(t, err)
			tk.pluginRegistry = tt.registry
			if err := tk.Setup(context.TODO(), sCtx); err != nil {
				if !tt.wantErr {
					t.Errorf("Handler.Setup() error not expected. got = %v", err)
				}
			} else {
				if tt.wantErr {
					t.Errorf("Handler.Setup() error expected, got none!")
				}
				for k, v := range tt.fields.pluginIDs {
					p, ok := tk.defaultPlugins[k]
					if assert.True(t, ok, "plugin %s not found", k) {
						assert.Equal(t, v, p.GetID())
					}
				}
				if tt.fields.defaultPluginID != "" {
					if assert.NotNil(t, tk.defaultPlugin, "default plugin is nil") {
						assert.Equal(t, tk.defaultPlugin.GetID(), tt.fields.defaultPluginID)
					}
				}
			}
		})
	}
}

func Test_task_ResolvePlugin(t *testing.T) {
	defaultID := "default"
	someID := "some"
	defaultPlugin := &pluginCoreMocks.Plugin{}
	defaultPlugin.On("GetID").Return(defaultID)
	somePlugin := &pluginCoreMocks.Plugin{}
	somePlugin.On("GetID").Return(someID)
	type fields struct {
		plugins        map[pluginCore.TaskType]pluginCore.Plugin
		defaultPlugin  pluginCore.Plugin
		pluginsForType map[pluginCore.TaskType]map[pluginID]pluginCore.Plugin
	}
	type args struct {
		ttype           string
		executionConfig v1alpha1.ExecutionConfig
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		want    string
		wantErr bool
	}{
		{"no-plugins", fields{}, args{}, "", true},
		{"default",
			fields{
				defaultPlugin: defaultPlugin,
			}, args{ttype: someID}, defaultID, false},
		{"actual",
			fields{
				plugins: map[pluginCore.TaskType]pluginCore.Plugin{
					someID: somePlugin,
				},
				defaultPlugin: defaultPlugin,
			}, args{ttype: someID}, someID, false},
		{"override",
			fields{
				plugins:       make(map[pluginCore.TaskType]pluginCore.Plugin),
				defaultPlugin: defaultPlugin,
				pluginsForType: map[pluginCore.TaskType]map[pluginID]pluginCore.Plugin{
					someID: {
						someID: somePlugin,
					},
				},
			}, args{ttype: someID, executionConfig: v1alpha1.ExecutionConfig{
				TaskPluginImpls: map[string]v1alpha1.TaskPluginOverride{
					someID: {
						PluginIDs:             []string{someID},
						MissingPluginBehavior: admin.PluginOverride_FAIL,
					},
				},
			}}, someID, false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tk := Handler{
				defaultPlugins:   tt.fields.plugins,
				defaultPlugin:    tt.fields.defaultPlugin,
				pluginsForType:   tt.fields.pluginsForType,
				connectorService: &pluginCore.ConnectorService{},
			}
			got, err := tk.ResolvePlugin(context.TODO(), tt.args.ttype, tt.args.executionConfig)
			if (err != nil) != tt.wantErr {
				t.Errorf("Handler.ResolvePlugin() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if err == nil {
				assert.Equal(t, tt.want, got.GetID())
			}
		})
	}
}

type fakeBufferedEventRecorder struct {
	evs []*event.TaskExecutionEvent
}

func (f *fakeBufferedEventRecorder) RecordTaskEvent(ctx context.Context, ev *event.TaskExecutionEvent, eventConfig *controllerConfig.EventConfig) error {
	f.evs = append(f.evs, ev)
	return nil
}

func (f *fakeBufferedEventRecorder) RecordNodeEvent(ctx context.Context, ev *event.NodeExecutionEvent, eventConfig *controllerConfig.EventConfig) error {
	return nil
}

type taskNodeStateHolder struct {
	s handler.TaskNodeState
}

func (t *taskNodeStateHolder) ClearNodeStatus() {
}

func (t *taskNodeStateHolder) PutTaskNodeState(s handler.TaskNodeState) error {
	t.s = s
	return nil
}

func (t taskNodeStateHolder) PutBranchNode(s handler.BranchNodeState) error {
	panic("not implemented")
}

func (t taskNodeStateHolder) PutWorkflowNodeState(s handler.WorkflowNodeState) error {
	panic("not implemented")
}

func (t taskNodeStateHolder) PutDynamicNodeState(s handler.DynamicNodeState) error {
	panic("not implemented")
}

func (t taskNodeStateHolder) PutGateNodeState(s handler.GateNodeState) error {
	panic("not implemented")
}

func (t taskNodeStateHolder) PutArrayNodeState(s handler.ArrayNodeState) error {
	panic("not implemented")
}

func CreateNoopResourceManager(ctx context.Context, scope promutils.Scope) resourcemanager.BaseResourceManager {
	rmBuilder, _ := resourcemanager.GetResourceManagerBuilderByType(ctx, rmConfig.TypeNoop, scope)
	rm, _ := rmBuilder.BuildResourceManager(ctx)
	return rm
}

func Test_task_Handle_NoCatalog(t *testing.T) {

	inputs := &core.LiteralMap{
		Literals: map[string]*core.Literal{
			"foo": coreutils.MustMakeLiteral("bar"),
		},
	}
	createNodeContext := func(pluginPhase pluginCore.Phase, pluginVer uint32, pluginResp fakeplugins.NextPhaseState, recorder interfaces.EventRecorder, ttype string, s *taskNodeStateHolder, allowIncrementParallelism bool) *nodeMocks.NodeExecutionContext {
		wfExecID := &core.WorkflowExecutionIdentifier{
			Project: "project",
			Domain:  "domain",
			Name:    "name",
		}

		nodeID := "n1"

		nm := &nodeMocks.NodeExecutionMetadata{}
		nm.EXPECT().GetAnnotations().Return(map[string]string{})
		nm.EXPECT().GetNodeExecutionID().Return(&core.NodeExecutionIdentifier{
			NodeId:      nodeID,
			ExecutionId: wfExecID,
		})
		nm.EXPECT().GetK8sServiceAccount().Return("service-account")
		nm.EXPECT().GetLabels().Return(map[string]string{})
		nm.EXPECT().GetNamespace().Return("namespace")
		nm.EXPECT().GetOwnerID().Return(types.NamespacedName{Namespace: "namespace", Name: "name"})
		nm.EXPECT().GetOwnerReference().Return(v12.OwnerReference{
			Kind: "sample",
			Name: "name",
		})
		nm.EXPECT().IsInterruptible().Return(false)

		tk := &core.TaskTemplate{
			Id:   &core.Identifier{ResourceType: core.ResourceType_TASK, Project: "proj", Domain: "dom", Version: "ver"},
			Type: "test",
			Metadata: &core.TaskMetadata{
				Discoverable: false,
			},
			Interface: &core.TypedInterface{
				Inputs: &core.VariableMap{
					Variables: map[string]*core.Variable{
						"foo": {
							Type: &core.LiteralType{
								Type: &core.LiteralType_Simple{
									Simple: core.SimpleType_STRING,
								},
							},
						},
					},
				},
				Outputs: &core.VariableMap{
					Variables: map[string]*core.Variable{
						"x": {
							Type: &core.LiteralType{
								Type: &core.LiteralType_Simple{
									Simple: core.SimpleType_BOOLEAN,
								},
							},
						},
					},
				},
			},
		}
		taskID := &core.Identifier{}
		tr := &nodeMocks.TaskReader{}
		tr.EXPECT().GetTaskID().Return(taskID)
		tr.EXPECT().GetTaskType().Return(ttype)
		tr.EXPECT().Read(mock.Anything).Return(tk, nil)

		ns := &flyteMocks.ExecutableNodeStatus{}
		ns.EXPECT().GetDataDir().Return("data-dir")
		ns.EXPECT().GetOutputDir().Return("data-dir")

		res := &v1.ResourceRequirements{}
		n := &flyteMocks.ExecutableNode{}
		ma := 5
		n.EXPECT().GetRetryStrategy().Return(&v1alpha1.RetryStrategy{MinAttempts: &ma})
		n.EXPECT().GetResources().Return(res)

		ir := &ioMocks.InputReader{}
		ir.EXPECT().GetInputPath().Return("input")
		ir.EXPECT().Get(mock.Anything).Return(inputs, nil)
		nCtx := &nodeMocks.NodeExecutionContext{}
		nCtx.EXPECT().NodeExecutionMetadata().Return(nm)
		nCtx.EXPECT().Node().Return(n)
		nCtx.EXPECT().InputReader().Return(ir)
		ds, err := storage.NewDataStore(
			&storage.Config{
				Type: storage.TypeMemory,
			},
			promutils.NewTestScope(),
		)
		assert.NoError(t, err)
		nCtx.EXPECT().DataStore().Return(ds)
		nCtx.EXPECT().CurrentAttempt().Return(uint32(1))
		nCtx.EXPECT().TaskReader().Return(tr)
		nCtx.EXPECT().NodeStatus().Return(ns)
		nCtx.EXPECT().NodeID().Return(nodeID)
		nCtx.EXPECT().EventsRecorder().Return(recorder)
		nCtx.EXPECT().EnqueueOwnerFunc().Return(nil)

		nCtx.EXPECT().RawOutputPrefix().Return("s3://sandbox/")
		nCtx.EXPECT().OutputShardSelector().Return(ioutils.NewConstantShardSelector([]string{"x"}))

		executionContext := &mocks.ExecutionContext{}
		executionContext.EXPECT().GetExecutionConfig().Return(v1alpha1.ExecutionConfig{})
		executionContext.EXPECT().GetEventVersion().Return(v1alpha1.EventVersion0)
		executionContext.EXPECT().GetParentInfo().Return(nil)
		if allowIncrementParallelism {
			executionContext.EXPECT().IncrementParallelism().Return(1)
		}
		nCtx.EXPECT().ExecutionContext().Return(executionContext)

		st := bytes.NewBuffer([]byte{})
		cod := codex.GobStateCodec{}
		assert.NoError(t, cod.Encode(pluginResp, st))
		nr := &nodeMocks.NodeStateReader{}
		nr.EXPECT().GetTaskNodeState().Return(handler.TaskNodeState{
			PluginState:        st.Bytes(),
			PluginPhase:        pluginPhase,
			PluginPhaseVersion: pluginVer,
		})
		nCtx.EXPECT().NodeStateReader().Return(nr)
		nCtx.EXPECT().NodeStateWriter().Return(s)
		return nCtx
	}

	noopRm := CreateNoopResourceManager(context.TODO(), promutils.NewTestScope())

	type args struct {
		startingPluginPhase        pluginCore.Phase
		startingPluginPhaseVersion int
		expectedState              fakeplugins.NextPhaseState
	}
	type want struct {
		handlerPhase handler.EPhase
		wantErr      bool
		event        bool
		eventPhase   core.TaskExecution_Phase
		incrParallel bool
		checkpoint   bool
	}
	tests := []struct {
		name string
		args args
		want want
	}{
		{
			"success",
			args{
				startingPluginPhase:        pluginCore.PhaseUndefined,
				startingPluginPhaseVersion: 0,
				expectedState: fakeplugins.NextPhaseState{
					Phase:        pluginCore.PhaseSuccess,
					PhaseVersion: 0,
					TaskInfo:     nil,
					TaskErr:      nil,
					OutputExists: true,
					OrError:      false,
				},
			},
			want{
				handlerPhase: handler.EPhaseSuccess,
				event:        true,
				eventPhase:   core.TaskExecution_SUCCEEDED,
				checkpoint:   true,
			},
		},
		{
			"success-output-missing",
			args{
				startingPluginPhase:        pluginCore.PhaseUndefined,
				startingPluginPhaseVersion: 0,
				expectedState: fakeplugins.NextPhaseState{
					Phase:        pluginCore.PhaseSuccess,
					PhaseVersion: 0,
					OutputExists: false,
				},
			},
			want{
				handlerPhase: handler.EPhaseRetryableFailure,
				event:        true,
				eventPhase:   core.TaskExecution_FAILED,
				checkpoint:   true,
			},
		},
		{
			"success-output-err-recoverable",
			args{
				startingPluginPhase:        pluginCore.PhaseUndefined,
				startingPluginPhaseVersion: 0,
				expectedState: fakeplugins.NextPhaseState{
					Phase:        pluginCore.PhaseSuccess,
					PhaseVersion: 0,
					TaskInfo:     nil,
					TaskErr: &io.ExecutionError{
						IsRecoverable: true,
					},
					OutputExists: false,
					OrError:      false,
				},
			},
			want{
				handlerPhase: handler.EPhaseRetryableFailure,
				event:        true,
				eventPhase:   core.TaskExecution_FAILED,
				checkpoint:   true,
			},
		},
		{
			"success-output-err-non-recoverable",
			args{
				startingPluginPhase:        pluginCore.PhaseUndefined,
				startingPluginPhaseVersion: 0,
				expectedState: fakeplugins.NextPhaseState{
					Phase:        pluginCore.PhaseSuccess,
					PhaseVersion: 0,
					TaskInfo:     nil,
					TaskErr: &io.ExecutionError{
						IsRecoverable: false,
					},
					OutputExists: false,
					OrError:      false,
				},
			},
			want{
				handlerPhase: handler.EPhaseFailed,
				event:        true,
				eventPhase:   core.TaskExecution_FAILED,
				checkpoint:   true,
			},
		},
		{
			"running",
			args{
				startingPluginPhase:        pluginCore.PhaseUndefined,
				startingPluginPhaseVersion: 0,
				expectedState: fakeplugins.NextPhaseState{
					Phase:        pluginCore.PhaseRunning,
					PhaseVersion: 0,
					TaskInfo: &pluginCore.TaskInfo{
						Logs: []*core.TaskLog{
							{Name: "x", Uri: "y"},
						},
					},
				},
			},
			want{
				handlerPhase: handler.EPhaseRunning,
				event:        true,
				eventPhase:   core.TaskExecution_RUNNING,
				incrParallel: true,
			},
		},
		{
			"running-no-event-phaseversion",
			args{
				startingPluginPhase:        pluginCore.PhaseRunning,
				startingPluginPhaseVersion: 1,
				expectedState: fakeplugins.NextPhaseState{
					Phase:        pluginCore.PhaseRunning,
					PhaseVersion: 1,
					TaskInfo: &pluginCore.TaskInfo{
						Logs: []*core.TaskLog{
							{Name: "x", Uri: "y"},
						},
					},
				},
			},
			want{
				handlerPhase: handler.EPhaseRunning,
				event:        false,
				incrParallel: true,
			},
		},
		{
			"running-error",
			args{
				startingPluginPhase:        pluginCore.PhaseRunning,
				startingPluginPhaseVersion: 1,
				expectedState: fakeplugins.NextPhaseState{
					OrError: true,
				},
			},
			want{
				handlerPhase: handler.EPhaseUndefined,
				event:        false,
				wantErr:      true,
				incrParallel: true,
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			state := &taskNodeStateHolder{}
			ev := &fakeBufferedEventRecorder{}
			// #nosec G115
			nCtx := createNodeContext(tt.args.startingPluginPhase, uint32(tt.args.startingPluginPhaseVersion), tt.args.expectedState, ev, "test", state, tt.want.incrParallel)
			c := &pluginCatalogMocks.Client{}
			tk := Handler{
				cfg: &config.Config{MaxErrorMessageLength: 100},
				defaultPlugins: map[pluginCore.TaskType]pluginCore.Plugin{
					"test": fakeplugins.NewPhaseBasedPlugin(),
				},
				pluginScope:      promutils.NewTestScope(),
				catalog:          c,
				resourceManager:  noopRm,
				taskMetricsMap:   make(map[MetricKey]*taskMetrics),
				eventConfig:      eventConfig,
				connectorService: &pluginCore.ConnectorService{},
			}
			got, err := tk.Handle(context.TODO(), nCtx)
			if (err != nil) != tt.want.wantErr {
				t.Errorf("Handler.Handle() error = %v, wantErr %v", err, tt.want.wantErr)
				return
			}
			if err == nil {
				assert.Equal(t, tt.want.handlerPhase.String(), got.Info().GetPhase().String())
				if tt.want.event {
					if assert.Equal(t, 1, len(ev.evs)) {
						e := ev.evs[0]
						assert.Equal(t, tt.want.eventPhase.String(), e.GetPhase().String())
						if tt.args.expectedState.TaskInfo != nil {
							assert.Equal(t, tt.args.expectedState.TaskInfo.Logs, e.GetLogs())
						}
						if e.GetPhase() == core.TaskExecution_RUNNING || e.GetPhase() == core.TaskExecution_SUCCEEDED {
							assert.True(t, proto.Equal(inputs, e.GetInputData()))
						}
					}
				} else {
					assert.Equal(t, 0, len(ev.evs))
				}
				expectedPhase := tt.args.expectedState.Phase
				if tt.args.expectedState.Phase.IsSuccess() && !tt.args.expectedState.OutputExists {
					expectedPhase = pluginCore.PhaseRetryableFailure
				}
				if tt.args.expectedState.TaskErr != nil {
					if tt.args.expectedState.TaskErr.IsRecoverable {
						expectedPhase = pluginCore.PhaseRetryableFailure
					} else {
						expectedPhase = pluginCore.PhasePermanentFailure
					}
				}
				assert.Equal(t, expectedPhase.String(), state.s.PluginPhase.String())
				assert.Equal(t, tt.args.expectedState.PhaseVersion, state.s.PluginPhaseVersion)
				if tt.want.checkpoint {
					assert.Equal(t, "s3://sandbox/x/name-n1-1/_flytecheckpoints",
						got.Info().GetInfo().TaskNodeInfo.TaskNodeMetadata.GetCheckpointUri())
				} else {
					assert.True(t, got.Info().GetInfo() == nil || got.Info().GetInfo().TaskNodeInfo == nil ||
						got.Info().GetInfo().TaskNodeInfo.TaskNodeMetadata == nil ||
						len(got.Info().GetInfo().TaskNodeInfo.TaskNodeMetadata.GetCheckpointUri()) == 0)
				}
			}
		})
	}
}

func Test_task_Abort(t *testing.T) {
	createNodeCtx := func(ev interfaces.EventRecorder) *nodeMocks.NodeExecutionContext {
		wfExecID := &core.WorkflowExecutionIdentifier{
			Project: "project",
			Domain:  "domain",
			Name:    "name",
		}

		nodeID := "n1"

		nm := &nodeMocks.NodeExecutionMetadata{}
		nm.EXPECT().GetAnnotations().Return(map[string]string{})
		nm.EXPECT().GetNodeExecutionID().Return(&core.NodeExecutionIdentifier{
			NodeId:      nodeID,
			ExecutionId: wfExecID,
		})
		nm.EXPECT().GetK8sServiceAccount().Return("service-account")
		nm.EXPECT().GetLabels().Return(map[string]string{})
		nm.EXPECT().GetNamespace().Return("namespace")
		nm.EXPECT().GetOwnerID().Return(types.NamespacedName{Namespace: "namespace", Name: "name"})
		nm.EXPECT().GetOwnerReference().Return(v12.OwnerReference{
			Kind: "sample",
			Name: "name",
		})
		nm.EXPECT().IsInterruptible().Return(false)

		taskID := &core.Identifier{}
		tr := &nodeMocks.TaskReader{}
		tr.EXPECT().GetTaskID().Return(taskID)
		tr.EXPECT().GetTaskType().Return("x")

		ns := &flyteMocks.ExecutableNodeStatus{}
		ns.EXPECT().GetDataDir().Return(storage.DataReference("data-dir"))
		ns.EXPECT().GetOutputDir().Return(storage.DataReference("output-dir"))

		res := &v1.ResourceRequirements{}
		n := &flyteMocks.ExecutableNode{}
		ma := 5
		n.EXPECT().GetRetryStrategy().Return(&v1alpha1.RetryStrategy{MinAttempts: &ma})
		n.EXPECT().GetResources().Return(res)

		ir := &ioMocks.InputReader{}
		nCtx := &nodeMocks.NodeExecutionContext{}
		nCtx.EXPECT().NodeExecutionMetadata().Return(nm)
		nCtx.EXPECT().Node().Return(n)
		nCtx.EXPECT().InputReader().Return(ir)
		ds, err := storage.NewDataStore(
			&storage.Config{
				Type: storage.TypeMemory,
			},
			promutils.NewTestScope(),
		)
		assert.NoError(t, err)
		nCtx.EXPECT().DataStore().Return(ds)
		nCtx.EXPECT().CurrentAttempt().Return(uint32(1))
		nCtx.EXPECT().TaskReader().Return(tr)
		nCtx.EXPECT().NodeStatus().Return(ns)
		nCtx.EXPECT().NodeID().Return("n1")
		nCtx.EXPECT().EnqueueOwnerFunc().Return(nil)
		nCtx.EXPECT().EventsRecorder().Return(ev)

		executionContext := &mocks.ExecutionContext{}
		executionContext.EXPECT().GetExecutionConfig().Return(v1alpha1.ExecutionConfig{})
		executionContext.EXPECT().GetParentInfo().Return(nil)
		executionContext.EXPECT().GetEventVersion().Return(v1alpha1.EventVersion0)
		nCtx.EXPECT().ExecutionContext().Return(executionContext)

		nCtx.EXPECT().RawOutputPrefix().Return("s3://sandbox/")
		nCtx.EXPECT().OutputShardSelector().Return(ioutils.NewConstantShardSelector([]string{"x"}))

		st := bytes.NewBuffer([]byte{})
		a := 45
		type test struct {
			A int
		}
		cod := codex.GobStateCodec{}
		assert.NoError(t, cod.Encode(test{A: a}, st))
		nr := &nodeMocks.NodeStateReader{}
		nr.EXPECT().GetTaskNodeState().Return(handler.TaskNodeState{
			PluginState: st.Bytes(),
		})
		nCtx.EXPECT().NodeStateReader().Return(nr)
		return nCtx
	}

	noopRm := CreateNoopResourceManager(context.TODO(), promutils.NewTestScope())

	incompatibleClusterEventsRecorder := nodeMocks.EventRecorder{}
	incompatibleClusterEventsRecorder.EXPECT().RecordTaskEvent(mock.Anything, mock.Anything, mock.Anything).Return(
		&eventsErr.EventError{
			Code: eventsErr.EventIncompatibleCusterError,
		})

	type fields struct {
		defaultPluginCallback func() pluginCore.Plugin
	}
	type args struct {
		ev interfaces.EventRecorder
	}
	tests := []struct {
		name        string
		fields      fields
		args        args
		wantErr     bool
		abortCalled bool
	}{
		{"no-plugin", fields{defaultPluginCallback: func() pluginCore.Plugin {
			return nil
		}}, args{nil}, true, false},

		{"abort-fails", fields{defaultPluginCallback: func() pluginCore.Plugin {
			p := &pluginCoreMocks.Plugin{}
			p.On("GetID").Return("id")
			p.EXPECT().GetProperties().Return(pluginCore.PluginProperties{})
			p.On("Abort", mock.Anything, mock.Anything).Return(fmt.Errorf("error"))
			return p
		}}, args{nil}, true, true},
		{"abort-success", fields{defaultPluginCallback: func() pluginCore.Plugin {
			p := &pluginCoreMocks.Plugin{}
			p.On("GetID").Return("id")
			p.EXPECT().GetProperties().Return(pluginCore.PluginProperties{})
			p.On("Abort", mock.Anything, mock.Anything).Return(nil)
			return p
		}}, args{ev: &fakeBufferedEventRecorder{}}, false, true},
		{"abort-swallows-incompatible-cluster-err", fields{defaultPluginCallback: func() pluginCore.Plugin {
			p := &pluginCoreMocks.Plugin{}
			p.On("GetID").Return("id")
			p.EXPECT().GetProperties().Return(pluginCore.PluginProperties{})
			p.On("Abort", mock.Anything, mock.Anything).Return(nil)
			return p
		}}, args{ev: &incompatibleClusterEventsRecorder}, false, true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			m := tt.fields.defaultPluginCallback()
			tk := Handler{
				defaultPlugin:    m,
				resourceManager:  noopRm,
				connectorService: &pluginCore.ConnectorService{},
				eventConfig:      eventConfig,
			}
			nCtx := createNodeCtx(tt.args.ev)
			if err := tk.Abort(context.TODO(), nCtx, "reason"); (err != nil) != tt.wantErr {
				t.Errorf("Handler.Abort() error = %v, wantErr %v", err, tt.wantErr)
			}
			c := 0
			if tt.abortCalled {
				c = 1
				if !tt.wantErr {
					switch tt.args.ev.(type) {
					case *fakeBufferedEventRecorder:
						assert.Len(t, tt.args.ev.(*fakeBufferedEventRecorder).evs, 1)
					case *nodeMocks.EventRecorder:
						assert.Len(t, tt.args.ev.(*nodeMocks.EventRecorder).Calls, 1)
					}
				}
			}
			if m != nil {
				m.(*pluginCoreMocks.Plugin).AssertNumberOfCalls(t, "Abort", c)
			}
		})
	}
}

func Test_task_Abort_v1(t *testing.T) {
	createNodeCtx := func(ev interfaces.EventRecorder) *nodeMocks.NodeExecutionContext {
		wfExecID := &core.WorkflowExecutionIdentifier{
			Project: "project",
			Domain:  "domain",
			Name:    "name",
		}

		nodeID := "n1"

		nm := &nodeMocks.NodeExecutionMetadata{}
		nm.EXPECT().GetAnnotations().Return(map[string]string{})
		nm.EXPECT().GetNodeExecutionID().Return(&core.NodeExecutionIdentifier{
			NodeId:      nodeID,
			ExecutionId: wfExecID,
		})
		nm.EXPECT().GetK8sServiceAccount().Return("service-account")
		nm.EXPECT().GetLabels().Return(map[string]string{})
		nm.EXPECT().GetNamespace().Return("namespace")
		nm.EXPECT().GetOwnerID().Return(types.NamespacedName{Namespace: "namespace", Name: "name"})
		nm.EXPECT().GetOwnerReference().Return(v12.OwnerReference{
			Kind: "sample",
			Name: "name",
		})
		nm.EXPECT().IsInterruptible().Return(false)

		taskID := &core.Identifier{}
		tr := &nodeMocks.TaskReader{}
		tr.EXPECT().GetTaskID().Return(taskID)
		tr.EXPECT().GetTaskType().Return("x")

		ns := &flyteMocks.ExecutableNodeStatus{}
		ns.EXPECT().GetDataDir().Return(storage.DataReference("data-dir"))
		ns.EXPECT().GetOutputDir().Return(storage.DataReference("output-dir"))

		res := &v1.ResourceRequirements{}
		n := &flyteMocks.ExecutableNode{}
		ma := 5
		n.EXPECT().GetRetryStrategy().Return(&v1alpha1.RetryStrategy{MinAttempts: &ma})
		n.EXPECT().GetResources().Return(res)

		ir := &ioMocks.InputReader{}
		nCtx := &nodeMocks.NodeExecutionContext{}
		nCtx.EXPECT().NodeExecutionMetadata().Return(nm)
		nCtx.EXPECT().Node().Return(n)
		nCtx.EXPECT().InputReader().Return(ir)
		ds, err := storage.NewDataStore(
			&storage.Config{
				Type: storage.TypeMemory,
			},
			promutils.NewTestScope(),
		)
		assert.NoError(t, err)
		nCtx.EXPECT().DataStore().Return(ds)
		nCtx.EXPECT().CurrentAttempt().Return(uint32(1))
		nCtx.EXPECT().TaskReader().Return(tr)
		nCtx.EXPECT().NodeStatus().Return(ns)
		nCtx.EXPECT().NodeID().Return("n1")
		nCtx.EXPECT().EnqueueOwnerFunc().Return(nil)
		nCtx.EXPECT().EventsRecorder().Return(ev)

		executionContext := &mocks.ExecutionContext{}
		executionContext.EXPECT().GetExecutionConfig().Return(v1alpha1.ExecutionConfig{})
		executionContext.EXPECT().GetParentInfo().Return(nil)
		executionContext.EXPECT().GetEventVersion().Return(v1alpha1.EventVersion1)
		nCtx.EXPECT().ExecutionContext().Return(executionContext)

		nCtx.EXPECT().RawOutputPrefix().Return("s3://sandbox/")
		nCtx.EXPECT().OutputShardSelector().Return(ioutils.NewConstantShardSelector([]string{"x"}))

		st := bytes.NewBuffer([]byte{})
		a := 45
		type test struct {
			A int
		}
		cod := codex.GobStateCodec{}
		assert.NoError(t, cod.Encode(test{A: a}, st))
		nr := &nodeMocks.NodeStateReader{}
		nr.EXPECT().GetTaskNodeState().Return(handler.TaskNodeState{
			PluginState: st.Bytes(),
		})
		nCtx.EXPECT().NodeStateReader().Return(nr)
		return nCtx
	}

	noopRm := CreateNoopResourceManager(context.TODO(), promutils.NewTestScope())

	incompatibleClusterEventsRecorder := nodeMocks.EventRecorder{}
	incompatibleClusterEventsRecorder.EXPECT().RecordTaskEvent(mock.Anything, mock.Anything, mock.Anything).Return(
		&eventsErr.EventError{
			Code: eventsErr.EventIncompatibleCusterError,
		})

	type fields struct {
		defaultPluginCallback func() pluginCore.Plugin
	}
	type args struct {
		ev interfaces.EventRecorder
	}
	tests := []struct {
		name        string
		fields      fields
		args        args
		wantErr     bool
		abortCalled bool
	}{
		{"no-plugin", fields{defaultPluginCallback: func() pluginCore.Plugin {
			return nil
		}}, args{nil}, true, false},

		{"abort-fails", fields{defaultPluginCallback: func() pluginCore.Plugin {
			p := &pluginCoreMocks.Plugin{}
			p.On("GetID").Return("id")
			p.EXPECT().GetProperties().Return(pluginCore.PluginProperties{})
			p.On("Abort", mock.Anything, mock.Anything).Return(fmt.Errorf("error"))
			return p
		}}, args{nil}, true, true},
		{"abort-success", fields{defaultPluginCallback: func() pluginCore.Plugin {
			p := &pluginCoreMocks.Plugin{}
			p.On("GetID").Return("id")
			p.EXPECT().GetProperties().Return(pluginCore.PluginProperties{})
			p.On("Abort", mock.Anything, mock.Anything).Return(nil)
			return p
		}}, args{ev: &fakeBufferedEventRecorder{}}, false, true},
		{"abort-swallows-incompatible-cluster-err", fields{defaultPluginCallback: func() pluginCore.Plugin {
			p := &pluginCoreMocks.Plugin{}
			p.On("GetID").Return("id")
			p.EXPECT().GetProperties().Return(pluginCore.PluginProperties{})
			p.On("Abort", mock.Anything, mock.Anything).Return(nil)
			return p
		}}, args{ev: &incompatibleClusterEventsRecorder}, false, true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			m := tt.fields.defaultPluginCallback()
			tk := Handler{
				defaultPlugin:    m,
				resourceManager:  noopRm,
				connectorService: &pluginCore.ConnectorService{},
				eventConfig:      eventConfig,
			}
			nCtx := createNodeCtx(tt.args.ev)
			if err := tk.Abort(context.TODO(), nCtx, "reason"); (err != nil) != tt.wantErr {
				t.Errorf("Handler.Abort() error = %v, wantErr %v", err, tt.wantErr)
			}
			c := 0
			if tt.abortCalled {
				c = 1
				if !tt.wantErr {
					switch tt.args.ev.(type) {
					case *fakeBufferedEventRecorder:
						assert.Len(t, tt.args.ev.(*fakeBufferedEventRecorder).evs, 1)
					case *nodeMocks.EventRecorder:
						assert.Len(t, tt.args.ev.(*nodeMocks.EventRecorder).Calls, 1)
					}
				}
			}
			if m != nil {
				m.(*pluginCoreMocks.Plugin).AssertNumberOfCalls(t, "Abort", c)
			}
		})
	}
}

func Test_task_Finalize(t *testing.T) {

	createNodeContext := func() *nodeMocks.NodeExecutionContext {
		wfExecID := &core.WorkflowExecutionIdentifier{
			Project: "project",
			Domain:  "domain",
			Name:    "name",
		}

		nodeID := "n1"

		nm := &nodeMocks.NodeExecutionMetadata{}
		nm.EXPECT().GetAnnotations().Return(map[string]string{})
		nm.EXPECT().GetNodeExecutionID().Return(&core.NodeExecutionIdentifier{
			NodeId:      nodeID,
			ExecutionId: wfExecID,
		})
		nm.EXPECT().GetK8sServiceAccount().Return("service-account")
		nm.EXPECT().GetLabels().Return(map[string]string{})
		nm.EXPECT().GetNamespace().Return("namespace")
		nm.EXPECT().GetOwnerID().Return(types.NamespacedName{Namespace: "namespace", Name: "name"})
		nm.EXPECT().GetOwnerReference().Return(v12.OwnerReference{
			Kind: "sample",
			Name: "name",
		})

		taskID := &core.Identifier{}
		tk := &core.TaskTemplate{
			Id:   taskID,
			Type: "test",
			Interface: &core.TypedInterface{
				Outputs: &core.VariableMap{
					Variables: map[string]*core.Variable{
						"x": {
							Type: &core.LiteralType{
								Type: &core.LiteralType_Simple{
									Simple: core.SimpleType_BOOLEAN,
								},
							},
						},
					},
				},
			},
		}
		tr := &nodeMocks.TaskReader{}
		tr.EXPECT().GetTaskID().Return(taskID)
		tr.EXPECT().GetTaskType().Return("x")
		tr.EXPECT().Read(mock.Anything).Return(tk, nil)

		ns := &flyteMocks.ExecutableNodeStatus{}
		ns.EXPECT().GetDataDir().Return(storage.DataReference("data-dir"))
		ns.EXPECT().GetOutputDir().Return(storage.DataReference("output-dir"))

		res := &v1.ResourceRequirements{}
		n := &flyteMocks.ExecutableNode{}
		ma := 5
		n.EXPECT().GetRetryStrategy().Return(&v1alpha1.RetryStrategy{MinAttempts: &ma})
		n.EXPECT().GetResources().Return(res)

		ir := &ioMocks.InputReader{}
		ir.EXPECT().Get(mock.Anything).Return(&core.LiteralMap{}, nil)
		nCtx := &nodeMocks.NodeExecutionContext{}
		nCtx.EXPECT().NodeExecutionMetadata().Return(nm)
		nCtx.EXPECT().Node().Return(n)
		nCtx.EXPECT().InputReader().Return(ir)
		ds, err := storage.NewDataStore(
			&storage.Config{
				Type: storage.TypeMemory,
			},
			promutils.NewTestScope(),
		)
		assert.NoError(t, err)
		nCtx.EXPECT().DataStore().Return(ds)
		nCtx.EXPECT().CurrentAttempt().Return(uint32(1))
		nCtx.EXPECT().TaskReader().Return(tr)
		nCtx.EXPECT().NodeStatus().Return(ns)
		nCtx.EXPECT().NodeID().Return("n1")
		nCtx.EXPECT().EventsRecorder().Return(nil)
		nCtx.EXPECT().EnqueueOwnerFunc().Return(nil)

		executionContext := &mocks.ExecutionContext{}
		executionContext.EXPECT().GetExecutionConfig().Return(v1alpha1.ExecutionConfig{})
		executionContext.EXPECT().GetParentInfo().Return(nil)
		executionContext.EXPECT().GetEventVersion().Return(v1alpha1.EventVersion0)
		nCtx.EXPECT().ExecutionContext().Return(executionContext)

		nCtx.EXPECT().RawOutputPrefix().Return("s3://sandbox/")
		nCtx.EXPECT().OutputShardSelector().Return(ioutils.NewConstantShardSelector([]string{"x"}))

		st := bytes.NewBuffer([]byte{})
		a := 45
		type test struct {
			A int
		}
		cod := codex.GobStateCodec{}
		assert.NoError(t, cod.Encode(test{A: a}, st))
		nr := &nodeMocks.NodeStateReader{}
		nr.On("GetTaskNodeState").Return(handler.TaskNodeState{
			PluginState: st.Bytes(),
		})
		nCtx.On("NodeStateReader").Return(nr)
		return nCtx
	}

	noopRm := CreateNoopResourceManager(context.TODO(), promutils.NewTestScope())
	type fields struct {
		defaultPluginCallback func() pluginCore.Plugin
	}
	tests := []struct {
		name     string
		fields   fields
		wantErr  bool
		finalize bool
	}{
		{"no-plugin", fields{defaultPluginCallback: func() pluginCore.Plugin {
			return nil
		}}, true, false},
		{"finalize-fails", fields{defaultPluginCallback: func() pluginCore.Plugin {
			p := &pluginCoreMocks.Plugin{}
			p.On("GetID").Return("id")
			p.EXPECT().GetProperties().Return(pluginCore.PluginProperties{})
			p.On("Finalize", mock.Anything, mock.Anything).Return(fmt.Errorf("error"))
			return p
		}}, true, true},
		{"finalize-success", fields{defaultPluginCallback: func() pluginCore.Plugin {
			p := &pluginCoreMocks.Plugin{}
			p.On("GetID").Return("id")
			p.EXPECT().GetProperties().Return(pluginCore.PluginProperties{})
			p.On("Finalize", mock.Anything, mock.Anything).Return(nil)
			return p
		}}, false, true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			nCtx := createNodeContext()

			catalog := &pluginCatalogMocks.Client{}
			m := tt.fields.defaultPluginCallback()
			mockClientset := k8sfake.NewSimpleClientset()
			tk, err := New(context.TODO(), mocks.NewFakeKubeClient(), mockClientset, catalog, eventConfig, testClusterID, promutils.NewTestScope())
			assert.NoError(t, err)
			tk.defaultPlugin = m
			tk.resourceManager = noopRm
			if err := tk.Finalize(context.TODO(), nCtx); (err != nil) != tt.wantErr {
				t.Errorf("Handler.Finalize() error = %v, wantErr %v", err, tt.wantErr)
			}
			c := 0
			if tt.finalize {
				c = 1
			}
			if m != nil {
				m.(*pluginCoreMocks.Plugin).AssertNumberOfCalls(t, "Finalize", c)
			}
		})
	}
}

func TestNew(t *testing.T) {
	mockClientset := k8sfake.NewSimpleClientset()
	got, err := New(context.TODO(), mocks.NewFakeKubeClient(), mockClientset, &pluginCatalogMocks.Client{}, eventConfig, testClusterID, promutils.NewTestScope())
	assert.NoError(t, err)
	assert.NotNil(t, got)
	assert.NotNil(t, got.defaultPlugins)
	assert.NotNil(t, got.metrics)
	assert.Equal(t, got.pluginRegistry, pluginmachinery.PluginRegistry())
}

func init() {
	labeled.SetMetricKeys(contextutils.ProjectKey, contextutils.DomainKey, contextutils.WorkflowIDKey,
		contextutils.TaskIDKey)
}

func Test_task_Handle_ValidateOutputErr(t *testing.T) {
	ctx := context.TODO()
	nodeID := "n1"
	execConfig := v1alpha1.ExecutionConfig{}

	tk := &core.TaskTemplate{
		Id:   &core.Identifier{ResourceType: core.ResourceType_TASK, Project: "proj", Domain: "dom", Version: "ver"},
		Type: "test",
		Interface: &core.TypedInterface{
			Outputs: &core.VariableMap{
				Variables: map[string]*core.Variable{
					"x": {
						Type: &core.LiteralType{
							Type: &core.LiteralType_Simple{
								Simple: core.SimpleType_BOOLEAN,
							},
						},
					},
				},
			},
		},
	}
	taskID := &core.Identifier{}
	tr := &nodeMocks.TaskReader{}
	tr.EXPECT().GetTaskID().Return(taskID)
	tr.EXPECT().Read(mock.Anything).Return(tk, nil)

	expectedErr := errors.Wrapf(ioutils.ErrRemoteFileExceedsMaxSize, "test file size exceeded")
	r := &ioMocks.OutputReader{}
	r.EXPECT().IsError(ctx).Return(false, nil)
	r.EXPECT().Exists(ctx).Return(true, expectedErr)

	h := Handler{}
	result, err := h.ValidateOutput(ctx, nodeID, nil, r, nil, execConfig, tr)
	assert.NoError(t, err)
	assert.False(t, result.IsRecoverable)
}
