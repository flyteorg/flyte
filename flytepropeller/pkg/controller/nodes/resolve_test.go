package nodes

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"

	"github.com/flyteorg/flyte/flyteidl/clients/go/coreutils"
	"github.com/flyteorg/flyte/flyteidl/gen/pb-go/flyteidl/core"
	"github.com/flyteorg/flyte/flytepropeller/pkg/apis/flyteworkflow/v1alpha1"
	"github.com/flyteorg/flyte/flytepropeller/pkg/controller/executors"
	"github.com/flyteorg/flyte/flytepropeller/pkg/utils"
	flyteassert "github.com/flyteorg/flyte/flytepropeller/pkg/utils/assert"
	"github.com/flyteorg/flyte/flytestdlib/promutils"
	"github.com/flyteorg/flyte/flytestdlib/storage"
)

var testScope = promutils.NewScope("test")

type dummyBaseWorkflow struct {
	executors.ControlFlow
	DummyStartNode v1alpha1.ExecutableNode
	ID             v1alpha1.WorkflowID
	ToNodeCb       func(name v1alpha1.NodeID) ([]v1alpha1.NodeID, error)
	FromNodeCb     func(name v1alpha1.NodeID) ([]v1alpha1.NodeID, error)
	GetNodeCb      func(nodeId v1alpha1.NodeID) (v1alpha1.ExecutableNode, bool)
	Status         map[v1alpha1.NodeID]*v1alpha1.NodeStatus
	DataStore      *storage.DataStore
	Interruptible  bool
}

func (d *dummyBaseWorkflow) GetDefinitionVersion() v1alpha1.WorkflowDefinitionVersion {
	return v1alpha1.WorkflowDefinitionVersion1
}

func (d *dummyBaseWorkflow) GetParentInfo() executors.ImmutableParentInfo {
	return nil
}

func (d *dummyBaseWorkflow) GetEventVersion() v1alpha1.EventVersion {
	return v1alpha1.EventVersion0
}

func (d *dummyBaseWorkflow) GetOnFailurePolicy() v1alpha1.WorkflowOnFailurePolicy {
	return v1alpha1.WorkflowOnFailurePolicy(core.WorkflowMetadata_FAIL_IMMEDIATELY)
}

func (d *dummyBaseWorkflow) ToNode(name v1alpha1.NodeID) ([]v1alpha1.NodeID, error) {
	return d.ToNodeCb(name)
}

func (d *dummyBaseWorkflow) GetOutputBindings() []*v1alpha1.Binding {
	return []*v1alpha1.Binding{}
}

func (d *dummyBaseWorkflow) GetOnFailureNode() v1alpha1.ExecutableNode {
	return nil
}

func (d *dummyBaseWorkflow) GetNodes() []v1alpha1.NodeID {
	return []v1alpha1.NodeID{d.DummyStartNode.GetID()}
}

func (d *dummyBaseWorkflow) GetConnections() *v1alpha1.Connections {
	return &v1alpha1.Connections{}
}

func (d *dummyBaseWorkflow) GetOutputs() *v1alpha1.OutputVarMap {
	return &v1alpha1.OutputVarMap{}
}

func (d *dummyBaseWorkflow) GetExecutionID() v1alpha1.ExecutionID {
	return v1alpha1.ExecutionID{
		WorkflowExecutionIdentifier: &core.WorkflowExecutionIdentifier{
			Name: "test",
		},
	}
}

func (d *dummyBaseWorkflow) GetK8sWorkflowID() types.NamespacedName {
	return types.NamespacedName{
		Name: "WF_Name",
	}
}

func (d *dummyBaseWorkflow) GetOwnerReference() v1.OwnerReference {
	return v1.OwnerReference{}
}

func (d *dummyBaseWorkflow) GetNamespace() string {
	return d.GetK8sWorkflowID().Namespace
}

func (d *dummyBaseWorkflow) GetCreationTimestamp() v1.Time {
	return v1.Now()
}

func (d *dummyBaseWorkflow) GetAnnotations() map[string]string {
	return map[string]string{}
}

func (d *dummyBaseWorkflow) GetLabels() map[string]string {
	return map[string]string{}
}

func (d *dummyBaseWorkflow) IsInterruptible() bool {
	return d.Interruptible
}

func (d *dummyBaseWorkflow) GetRawOutputDataConfig() v1alpha1.RawOutputDataConfig {
	return v1alpha1.RawOutputDataConfig{}
}

func (d *dummyBaseWorkflow) GetName() string {
	return d.ID
}

func (d *dummyBaseWorkflow) GetServiceAccountName() string {
	return ""
}

func (d *dummyBaseWorkflow) GetSecurityContext() core.SecurityContext {
	return core.SecurityContext{}
}

func (d *dummyBaseWorkflow) GetTask(id v1alpha1.TaskID) (v1alpha1.ExecutableTask, error) {
	return nil, nil
}

func (d *dummyBaseWorkflow) FindSubWorkflow(subID v1alpha1.WorkflowID) v1alpha1.ExecutableSubWorkflow {
	return nil
}

func (d *dummyBaseWorkflow) GetExecutionStatus() v1alpha1.ExecutableWorkflowStatus {
	return nil
}

func (d *dummyBaseWorkflow) GetNodeExecutionStatus(_ context.Context, id v1alpha1.NodeID) v1alpha1.ExecutableNodeStatus {
	n, ok := d.Status[id]
	if ok {
		n.DataReferenceConstructor = d.DataStore
		return n
	}
	n = &v1alpha1.NodeStatus{
		MutableStruct: v1alpha1.MutableStruct{},
	}
	d.Status[id] = n
	n.DataReferenceConstructor = d.DataStore
	return n
}

func (d *dummyBaseWorkflow) StartNode() v1alpha1.ExecutableNode {
	return d.DummyStartNode
}

func (d *dummyBaseWorkflow) GetID() v1alpha1.WorkflowID {
	return d.ID
}

func (d *dummyBaseWorkflow) FromNode(name v1alpha1.NodeID) ([]v1alpha1.NodeID, error) {
	return d.FromNodeCb(name)
}

func (d *dummyBaseWorkflow) GetNode(nodeID v1alpha1.NodeID) (v1alpha1.ExecutableNode, bool) {
	return d.GetNodeCb(nodeID)
}

func (d *dummyBaseWorkflow) GetExecutionConfig() v1alpha1.ExecutionConfig {
	return v1alpha1.ExecutionConfig{}
}

func (d *dummyBaseWorkflow) GetConsoleURL() string { return "" }

func createDummyBaseWorkflow(dataStore *storage.DataStore) *dummyBaseWorkflow {
	return &dummyBaseWorkflow{
		ID: "w1",
		Status: map[v1alpha1.NodeID]*v1alpha1.NodeStatus{
			v1alpha1.StartNodeID: {},
		},
		DataStore: dataStore,
	}
}

func createInmemoryDataStore(t testing.TB, scope promutils.Scope) *storage.DataStore {
	cfg := storage.Config{
		Type: storage.TypeMemory,
	}
	d, err := storage.NewDataStore(&cfg, scope)
	assert.NoError(t, err)
	return d
}

func createFailingDatastore(_ testing.TB, scope promutils.Scope) *storage.DataStore {
	return storage.NewCompositeDataStore(storage.URLPathConstructor{}, storage.NewDefaultProtobufStore(utils.FailingRawStore{}, scope))
}

func TestResolveBindingData(t *testing.T) {
	ctx := context.Background()
	outputRef := v1alpha1.DataReference("output-ref")
	n1 := &v1alpha1.NodeSpec{
		ID: "n1",
		OutputAliases: []v1alpha1.Alias{
			{Alias: core.Alias{
				Var:   "x",
				Alias: "m",
			}},
		},
	}

	n2 := &v1alpha1.NodeSpec{
		ID: "n2",
		OutputAliases: []v1alpha1.Alias{
			{Alias: core.Alias{
				Var:   "x",
				Alias: "m",
			}},
		},
	}

	outputPath := v1alpha1.GetOutputsFile(outputRef)

	w := &dummyBaseWorkflow{
		Status: map[v1alpha1.NodeID]*v1alpha1.NodeStatus{
			"n2": {
				DataDir:   outputRef,
				OutputDir: outputRef,
			},
		},
		GetNodeCb: func(nodeId v1alpha1.NodeID) (v1alpha1.ExecutableNode, bool) {
			switch nodeId {
			case "n1":
				return n1, true
			case "n2":
				return n2, true
			}
			return nil, false
		},
	}

	t.Run("StaticBinding", func(t *testing.T) {
		w := &dummyBaseWorkflow{}
		b := utils.MustMakePrimitiveBindingData(1)
		l, err := ResolveBindingData(ctx, nil, w, b)
		assert.NoError(t, err)
		flyteassert.EqualLiterals(t, coreutils.MustMakeLiteral(1), l)
	})

	t.Run("PromiseMissingNode", func(t *testing.T) {
		w := &dummyBaseWorkflow{
			GetNodeCb: func(nodeId v1alpha1.NodeID) (v1alpha1.ExecutableNode, bool) {
				return nil, false
			},
		}
		b := utils.MakeBindingDataPromise("n1", "x")
		_, err := ResolveBindingData(ctx, nil, w, b)
		assert.Error(t, err)
	})

	t.Run("PromiseMissing", func(t *testing.T) {
		store := createInmemoryDataStore(t, testScope.NewSubScope("1"))
		r := remoteFileOutputResolver{store: store}
		b := utils.MakeBindingDataPromise("n1", "x")
		_, err := ResolveBindingData(ctx, r, w, b)
		assert.Error(t, err)
	})

	t.Run("PromiseMissingWithData", func(t *testing.T) {
		store := createInmemoryDataStore(t, testScope.NewSubScope("2"))
		r := remoteFileOutputResolver{store: store}
		m, err := coreutils.MakeLiteralMap(map[string]interface{}{"z": 1})
		assert.NoError(t, err)
		assert.NoError(t, store.WriteProtobuf(ctx, outputPath, storage.Options{}, m))
		b := utils.MakeBindingDataPromise("n1", "x")
		_, err = ResolveBindingData(ctx, r, w, b)
		assert.Error(t, err)
	})

	t.Run("PromiseFound", func(t *testing.T) {
		store := createInmemoryDataStore(t, testScope.NewSubScope("3"))
		r := remoteFileOutputResolver{store: store}
		m, err := coreutils.MakeLiteralMap(map[string]interface{}{"x": 1})
		assert.NoError(t, err)
		assert.NoError(t, store.WriteProtobuf(ctx, outputPath, storage.Options{}, m))

		b := utils.MakeBindingDataPromise("n2", "x")
		l, err := ResolveBindingData(ctx, r, w, b)
		if assert.NoError(t, err) {
			flyteassert.EqualLiterals(t, coreutils.MustMakeLiteral(1), l)
		}
	})

	t.Run("NullBinding", func(t *testing.T) {
		l, err := ResolveBindingData(ctx, nil, w, nil)
		assert.NoError(t, err)
		assert.Nil(t, l)
	})

	t.Run("NullWorkflowPromise", func(t *testing.T) {
		store := createInmemoryDataStore(t, testScope.NewSubScope("4"))
		r := remoteFileOutputResolver{store: store}
		m, err := coreutils.MakeLiteralMap(map[string]interface{}{"x": 1})
		assert.NoError(t, err)
		assert.NoError(t, store.WriteProtobuf(ctx, outputPath, storage.Options{}, m))
		b := utils.MakeBindingDataPromise("n1", "x")
		_, err = ResolveBindingData(ctx, r, nil, b)
		assert.Error(t, err)
	})

	t.Run("PromiseFoundAlias", func(t *testing.T) {
		store := createInmemoryDataStore(t, testScope.NewSubScope("5"))
		r := remoteFileOutputResolver{store: store}
		m, err := coreutils.MakeLiteralMap(map[string]interface{}{"x": 1})
		assert.NoError(t, err)
		assert.NoError(t, store.WriteProtobuf(ctx, outputPath, storage.Options{}, m))
		b := utils.MakeBindingDataPromise("n2", "m")
		l, err := ResolveBindingData(ctx, r, w, b)
		if assert.NoError(t, err) {
			flyteassert.EqualLiterals(t, coreutils.MustMakeLiteral(1), l)
		}
	})

	t.Run("BindingDataMap", func(t *testing.T) {
		store := createInmemoryDataStore(t, testScope.NewSubScope("6"))
		r := remoteFileOutputResolver{store: store}
		// Store output of previous
		m, err := coreutils.MakeLiteralMap(map[string]interface{}{"x": 1})
		assert.NoError(t, err)
		assert.NoError(t, store.WriteProtobuf(ctx, outputPath, storage.Options{}, m))
		m2 := &core.LiteralMap{}
		assert.NoError(t, store.ReadProtobuf(ctx, outputPath, m2))
		// Output of current
		b := utils.MakeBindingDataMap(
			utils.NewPair("x", utils.MakeBindingDataPromise("n2", "x")),
			utils.NewPair("z", utils.MustMakePrimitiveBindingData(5)),
		)
		l, err := ResolveBindingData(ctx, r, w, b)
		if assert.NoError(t, err) {
			expected, err := coreutils.MakeLiteralMap(map[string]interface{}{"x": 1, "z": 5})
			assert.NoError(t, err)
			flyteassert.EqualLiteralMap(t, expected, l.GetMap())
		}

	})

	t.Run("BindingDataMapFailedPromise", func(t *testing.T) {
		store := createInmemoryDataStore(t, testScope.NewSubScope("7"))
		r := remoteFileOutputResolver{store: store}
		// do not store anything

		// Output of current
		b := utils.MakeBindingDataMap(
			utils.NewPair("x", utils.MakeBindingDataPromise("n1", "x")),
			utils.NewPair("z", utils.MustMakePrimitiveBindingData(5)),
		)
		_, err := ResolveBindingData(ctx, r, w, b)
		assert.Error(t, err)
	})

	t.Run("BindingDataCollection", func(t *testing.T) {
		store := createInmemoryDataStore(t, testScope.NewSubScope("8"))
		r := remoteFileOutputResolver{store: store}
		// Store random value
		m, err := coreutils.MakeLiteralMap(map[string]interface{}{"jj": 1})
		assert.NoError(t, err)
		assert.NoError(t, store.WriteProtobuf(ctx, outputPath, storage.Options{}, m))

		// binding of current npde
		b := utils.MakeBindingDataCollection(
			utils.MakeBindingDataPromise("n1", "x"),
			utils.MustMakePrimitiveBindingData(5),
		)
		_, err = ResolveBindingData(ctx, r, w, b)
		assert.Error(t, err)

	})
}

func TestResolve(t *testing.T) {
	ctx := context.Background()
	outputRef := v1alpha1.DataReference("output-ref")
	n1 := &v1alpha1.NodeSpec{
		ID: "n1",
		OutputAliases: []v1alpha1.Alias{
			{Alias: core.Alias{
				Var:   "x",
				Alias: "m",
			}},
		},
	}

	outputPath := v1alpha1.GetOutputsFile(outputRef)

	w := &dummyBaseWorkflow{
		Status: map[v1alpha1.NodeID]*v1alpha1.NodeStatus{
			"n1": {
				DataDir:   outputRef,
				OutputDir: outputRef,
			},
		},
		GetNodeCb: func(nodeId v1alpha1.NodeID) (v1alpha1.ExecutableNode, bool) {
			if nodeId == "n1" {
				return n1, true
			}
			return nil, false
		},
	}

	t.Run("SimpleResolve", func(t *testing.T) {
		store := createInmemoryDataStore(t, testScope.NewSubScope("9"))
		r := remoteFileOutputResolver{store: store}
		// Store output of previous
		m, err := coreutils.MakeLiteralMap(map[string]interface{}{"x": 1})
		assert.NoError(t, err)
		assert.NoError(t, store.WriteProtobuf(ctx, outputPath, storage.Options{}, m))

		// bindings
		b := []*v1alpha1.Binding{
			{
				Binding: utils.MakeBinding("map", utils.MakeBindingDataMap(
					utils.NewPair("x", utils.MakeBindingDataPromise("n1", "x")),
					utils.NewPair("z", utils.MustMakePrimitiveBindingData(5)),
				)),
			},
			{
				Binding: utils.MakeBinding("simple", utils.MustMakePrimitiveBindingData(1)),
			},
		}

		expected, err := coreutils.MakeLiteralMap(map[string]interface{}{
			"map":    map[string]interface{}{"x": 1, "z": 5},
			"simple": coreutils.MustMakePrimitiveLiteral(1),
		})
		assert.NoError(t, err)

		l, err := Resolve(ctx, r, w, "n2", b)
		if assert.NoError(t, err) {
			assert.NotNil(t, l)
			if assert.NoError(t, err) {
				flyteassert.EqualLiteralMap(t, expected, l)
			}
		}
	})

	t.Run("SimpleResolveFail", func(t *testing.T) {
		store := createInmemoryDataStore(t, testScope.NewSubScope("10"))
		r := remoteFileOutputResolver{store: store}
		// Store has no previous output

		// bindings
		b := []*v1alpha1.Binding{
			{
				Binding: utils.MakeBinding("map", utils.MakeBindingDataMap(
					utils.NewPair("x", utils.MakeBindingDataPromise("n1", "x")),
					utils.NewPair("z", utils.MustMakePrimitiveBindingData(5)),
				)),
			},
			{
				Binding: utils.MakeBinding("simple", utils.MustMakePrimitiveBindingData(1)),
			},
		}

		_, err := Resolve(ctx, r, w, "n2", b)
		assert.Error(t, err)
	})

}
