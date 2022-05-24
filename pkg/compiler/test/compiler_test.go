package test

import (
	"encoding/json"
	"flag"
	"io/ioutil"
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"testing"
	"time"

	"github.com/flyteorg/flytepropeller/pkg/visualize"

	"github.com/flyteorg/flyteidl/gen/pb-go/flyteidl/admin"

	"k8s.io/apimachinery/pkg/util/sets"

	"github.com/go-test/deep"

	"github.com/ghodss/yaml"

	"github.com/flyteorg/flyteidl/clients/go/coreutils"
	"github.com/flyteorg/flyteidl/gen/pb-go/flyteidl/core"
	"github.com/flyteorg/flytepropeller/pkg/compiler"
	"github.com/flyteorg/flytepropeller/pkg/compiler/common"
	"github.com/flyteorg/flytepropeller/pkg/compiler/errors"
	"github.com/flyteorg/flytepropeller/pkg/compiler/transformers/k8s"
	"github.com/golang/protobuf/jsonpb"
	"github.com/golang/protobuf/proto"
	"github.com/stretchr/testify/assert"
)

var update = flag.Bool("update", false, "Update .golden files")

func makeDefaultInputs(iface *core.TypedInterface) *core.LiteralMap {
	if iface == nil || iface.GetInputs() == nil {
		return nil
	}

	res := make(map[string]*core.Literal, len(iface.GetInputs().Variables))
	for inputName, inputVar := range iface.GetInputs().Variables {
		// A workaround because the coreutils don't support the "StructuredDataSet" type
		if reflect.TypeOf(inputVar.Type.Type) == reflect.TypeOf(&core.LiteralType_StructuredDatasetType{}) {
			res[inputName] = &core.Literal{
				Value: &core.Literal_Scalar{
					Scalar: &core.Scalar{
						Value: &core.Scalar_StructuredDataset{
							StructuredDataset: &core.StructuredDataset{
								Metadata: &core.StructuredDatasetMetadata{
									StructuredDatasetType: inputVar.Type.Type.(*core.LiteralType_StructuredDatasetType).StructuredDatasetType,
								},
							},
						},
					},
				},
			}
		} else if reflect.TypeOf(inputVar.Type.Type) == reflect.TypeOf(&core.LiteralType_Simple{}) && inputVar.Type.GetSimple() == core.SimpleType_DATETIME {
			res[inputName] = coreutils.MustMakeLiteral(time.UnixMicro(10))
		} else {
			res[inputName] = coreutils.MustMakeDefaultLiteralForType(inputVar.Type)
		}
	}

	return &core.LiteralMap{
		Literals: res,
	}
}

func setDefaultFields(task *core.TaskTemplate) {
	if container := task.GetContainer(); container != nil {
		if container.Config == nil {
			container.Config = []*core.KeyValuePair{}
		}

		container.Config = append(container.Config, &core.KeyValuePair{
			Key:   "testKey1",
			Value: "testValue1",
		})
		container.Config = append(container.Config, &core.KeyValuePair{
			Key:   "testKey2",
			Value: "testValue2",
		})
		container.Config = append(container.Config, &core.KeyValuePair{
			Key:   "testKey3",
			Value: "testValue3",
		})
	}
}

func mustCompileTasks(t *testing.T, tasks []*core.TaskTemplate) []*core.CompiledTask {
	compiledTasks := make([]*core.CompiledTask, 0, len(tasks))
	for _, inputTask := range tasks {
		setDefaultFields(inputTask)
		task, err := compiler.CompileTask(inputTask)
		compiledTasks = append(compiledTasks, task)
		assert.NoError(t, err)
		if err != nil {
			assert.FailNow(t, err.Error())
		}
	}

	return compiledTasks
}

func TestDynamic(t *testing.T) {
	errors.SetConfig(errors.Config{IncludeSource: true})
	assert.NoError(t, filepath.Walk("testdata/dynamic", func(path string, info os.FileInfo, err error) error {
		if info.IsDir() {
			return nil
		}

		t.Run(path, func(t *testing.T) {
			// If you want to debug a single use-case. Uncomment this line.
			//if !strings.HasSuffix(path, "success_1.json") {
			//	t.SkipNow()
			//}

			raw, err := ioutil.ReadFile(path)
			assert.NoError(t, err)
			wf := &core.DynamicJobSpec{}
			err = jsonpb.UnmarshalString(string(raw), wf)
			if !assert.NoError(t, err) {
				t.FailNow()
			}

			t.Log("Compiling Workflow")
			compiledTasks := mustCompileTasks(t, wf.Tasks)
			wfTemplate := &core.WorkflowTemplate{
				Id: &core.Identifier{
					Domain:  "domain",
					Name:    "name",
					Version: "version",
				},
				Interface: &core.TypedInterface{
					Inputs: &core.VariableMap{Variables: map[string]*core.Variable{}},
					Outputs: &core.VariableMap{Variables: map[string]*core.Variable{
						"o0": {
							Type: &core.LiteralType{
								Type: &core.LiteralType_CollectionType{
									CollectionType: &core.LiteralType{
										Type: &core.LiteralType_Simple{
											Simple: core.SimpleType_INTEGER,
										},
									},
								},
							},
						},
					}},
				},
				Nodes:   wf.Nodes,
				Outputs: wf.Outputs,
			}
			compiledWfc, err := compiler.CompileWorkflow(wfTemplate, wf.Subworkflows, compiledTasks,
				[]common.InterfaceProvider{})
			if !assert.NoError(t, err) {
				t.FailNow()
			}

			inputs := makeDefaultInputs(compiledWfc.Primary.Template.Interface)

			flyteWf, err := k8s.BuildFlyteWorkflow(compiledWfc,
				inputs,
				&core.WorkflowExecutionIdentifier{
					Project: "hello",
					Domain:  "domain",
					Name:    "name",
				},
				"namespace")
			if assert.NoError(t, err) {
				raw, err := json.Marshal(flyteWf)
				if assert.NoError(t, err) {
					assert.NotEmpty(t, raw)
				}
			}
		})

		return nil
	}))
}

func getAllSubNodeIDs(n *core.Node) sets.String {
	res := sets.NewString()
	if branchNode := n.GetBranchNode(); branchNode != nil {
		thenNode := branchNode.IfElse.Case.ThenNode
		if hasPromiseInputs(thenNode.GetInputs()) {
			res.Insert(thenNode.GetId())
		}

		res = res.Union(getAllSubNodeIDs(thenNode))

		for _, other := range branchNode.IfElse.Other {
			if hasPromiseInputs(other.ThenNode.GetInputs()) {
				res.Insert(other.ThenNode.GetId())
			}

			res = res.Union(getAllSubNodeIDs(other.ThenNode))
		}

		if elseNode := branchNode.IfElse.GetElseNode(); elseNode != nil {
			if hasPromiseInputs(elseNode.GetInputs()) {
				res.Insert(elseNode.GetId())
			}

			res = res.Union(getAllSubNodeIDs(elseNode))
		}
	}

	// TODO: Support Sub workflow

	return res
}

type nodePredicate func(n *core.Node) bool

var hasPromiseNodePredicate = func(n *core.Node) bool {
	return hasPromiseInputs(n.GetInputs())
}

var allNodesPredicate = func(n *core.Node) bool {
	return true
}

func getAllMatchingNodes(wf *core.CompiledWorkflow, predicate nodePredicate) sets.String {
	s := sets.NewString()
	for _, n := range wf.Template.Nodes {
		if predicate(n) {
			s.Insert(n.GetId())
		}

		s = s.Union(getAllSubNodeIDs(n))
	}

	return s
}

func bindingHasPromiseInputs(binding *core.BindingData) bool {
	switch v := binding.GetValue().(type) {
	case *core.BindingData_Collection:
		for _, d := range v.Collection.Bindings {
			if bindingHasPromiseInputs(d) {
				return true
			}
		}
	case *core.BindingData_Map:
		for _, d := range v.Map.Bindings {
			if bindingHasPromiseInputs(d) {
				return true
			}
		}
	case *core.BindingData_Promise:
		return true
	}

	return false
}

func hasPromiseInputs(bindings []*core.Binding) bool {
	for _, b := range bindings {
		if bindingHasPromiseInputs(b.Binding) {
			return true
		}
	}

	return false
}

func assertNodeIDsInConnections(t testing.TB, nodeIDsWithDeps, allNodeIDs sets.String, connections *core.ConnectionSet) bool {
	actualNodeIDs := sets.NewString()
	for id, lst := range connections.Downstream {
		actualNodeIDs.Insert(id)
		actualNodeIDs.Insert(lst.Ids...)
	}

	for id, lst := range connections.Upstream {
		actualNodeIDs.Insert(id)
		actualNodeIDs.Insert(lst.Ids...)
	}

	notFoundInConnections := nodeIDsWithDeps.Difference(actualNodeIDs)
	correct := assert.Empty(t, notFoundInConnections, "All nodes must appear in connections")

	notFoundInNodes := actualNodeIDs.Difference(allNodeIDs)
	correct = correct && assert.Empty(t, notFoundInNodes, "All connections must correspond to existing nodes")

	return correct
}

func TestUseCases(t *testing.T) {
	runCompileTest(t, "branch")
	runCompileTest(t, "snacks-core")
}

func protoMarshal(v any) ([]byte, error) {
	m := jsonpb.Marshaler{}
	str, err := m.MarshalToString(v.(proto.Message))
	return []byte(str), err
}

func storeOrDiff(t testing.TB, f func(obj any) ([]byte, error), obj any, path string) bool {
	raw, err := f(obj)
	if !assert.NoError(t, err) {
		return false
	}

	if *update {
		err = ioutil.WriteFile(path, raw, os.ModePerm)
		if !assert.NoError(t, err) {
			return false
		}

	} else {
		goldenRaw, err := ioutil.ReadFile(path)
		if !assert.NoError(t, err) {
			return false
		}

		if diff := deep.Equal(string(raw), string(goldenRaw)); diff != nil {
			t.Errorf("Compiled() Diff = %v\r\n got = %v\r\n want = %v", diff, string(raw), string(goldenRaw))
		}
	}

	return true
}

func runCompileTest(t *testing.T, dirName string) {
	errors.SetConfig(errors.Config{IncludeSource: true})
	// Compile Tasks
	t.Run("tasks-"+dirName, func(t *testing.T) {
		//t.Parallel()

		paths, err := filepath.Glob("testdata/" + dirName + "/*.pb")
		if !assert.NoError(t, err) {
			t.FailNow()
		}

		for _, p := range paths {
			raw, err := ioutil.ReadFile(p)
			assert.NoError(t, err)
			tsk := &admin.TaskSpec{}
			err = proto.Unmarshal(raw, tsk)
			if err != nil {
				t.Logf("Failed to parse %s as a Task, skipping: %v", p, err)
				continue
			}

			t.Run(p, func(t *testing.T) {
				//t.Parallel()
				if !storeOrDiff(t, yaml.Marshal, tsk, strings.TrimSuffix(p, filepath.Ext(p))+"_task.yaml") {
					t.FailNow()
				}

				inputTask := tsk.Template
				setDefaultFields(inputTask)
				task, err := compiler.CompileTask(inputTask)
				if !assert.NoError(t, err) {
					t.FailNow()
				}

				if !storeOrDiff(t, yaml.Marshal, task, filepath.Join(filepath.Dir(p), "compiled", strings.TrimRight(filepath.Base(p), filepath.Ext(p))+"_task.yaml")) {
					t.FailNow()
				}

				if !storeOrDiff(t, protoMarshal, tsk, filepath.Join(filepath.Dir(p), "compiled", strings.TrimRight(filepath.Base(p), filepath.Ext(p))+"_task.json")) {
					t.FailNow()
				}
			})
		}
	})

	// Load Compiled Tasks
	paths, err := filepath.Glob(filepath.Join("testdata", dirName, "compiled", "*_task.json"))
	if !assert.NoError(t, err) {
		t.FailNow()
	}

	compiledTasks := make(map[string]*core.CompiledTask, len(paths))
	for _, f := range paths {
		raw, err := ioutil.ReadFile(f)
		if !assert.NoError(t, err) {
			t.FailNow()
		}

		tsk := &core.CompiledTask{}
		err = jsonpb.UnmarshalString(string(raw), tsk)
		if !assert.NoError(t, err) {
			t.FailNow()
		}

		compiledTasks[tsk.Template.Id.String()] = tsk
	}

	// Compile Workflows
	t.Run("workflows-"+dirName, func(t *testing.T) {
		//t.Parallel()

		paths, err = filepath.Glob(filepath.Join("testdata", dirName, "*.pb"))
		if !assert.NoError(t, err) {
			t.FailNow()
		}

		for _, p := range paths {
			raw, err := ioutil.ReadFile(p)
			assert.NoError(t, err)
			wf := &core.WorkflowClosure{}
			err = proto.Unmarshal(raw, wf)
			if err != nil {
				t.Logf("Failed to parse %s as a WorkflowClosure, skipping: %v", p, err)
				continue
			}

			t.Run(p, func(t *testing.T) {
				//t.Parallel()
				if !storeOrDiff(t, yaml.Marshal, wf, strings.TrimSuffix(p, filepath.Ext(p))+"_wf.yaml") {
					t.FailNow()
				}

				inputWf := wf.Workflow

				reqs, err := compiler.GetRequirements(inputWf, nil)
				if !assert.NoError(t, err) {
					t.FailNow()
				}

				tasks := make([]*core.CompiledTask, 0, len(reqs.GetRequiredTaskIds()))

				for _, taskID := range reqs.GetRequiredTaskIds() {
					compiledTask, found := compiledTasks[taskID.String()]
					if !assert.True(t, found, "Could not find compiled task %s", taskID) {
						t.FailNow()
					}

					tasks = append(tasks, compiledTask)
				}

				compiledWfc, err := compiler.CompileWorkflow(inputWf, []*core.WorkflowTemplate{}, tasks,
					[]common.InterfaceProvider{})
				if !assert.NoError(t, err) {
					t.FailNow()
				}

				if !storeOrDiff(t, yaml.Marshal, compiledWfc, filepath.Join(filepath.Dir(p), "compiled", strings.TrimRight(filepath.Base(p), filepath.Ext(p))+"_wf.yaml")) {
					t.FailNow()
				}

				if !storeOrDiff(t, protoMarshal, compiledWfc, filepath.Join(filepath.Dir(p), "compiled", strings.TrimRight(filepath.Base(p), filepath.Ext(p))+"_wf.json")) {
					t.FailNow()
				}

				allNodeIDs := getAllMatchingNodes(compiledWfc.Primary, allNodesPredicate)
				nodeIDsWithDeps := getAllMatchingNodes(compiledWfc.Primary, hasPromiseNodePredicate)
				if !assertNodeIDsInConnections(t, nodeIDsWithDeps, allNodeIDs, compiledWfc.Primary.Connections) {
					t.FailNow()
				}
			})
		}
	})

	// Build K8s Workflows
	t.Run("k8s-"+dirName, func(t *testing.T) {
		//t.Parallel()

		paths, err = filepath.Glob(filepath.Join("testdata", dirName, "compiled", "*_wf.json"))
		if !assert.NoError(t, err) {
			t.FailNow()
		}

		for _, p := range paths {
			t.Run(p, func(t *testing.T) {
				//t.Parallel()
				raw, err := ioutil.ReadFile(p)
				if !assert.NoError(t, err) {
					t.FailNow()
				}

				compiledWfc := &core.CompiledWorkflowClosure{}
				if !assert.NoError(t, jsonpb.UnmarshalString(string(raw), compiledWfc)) {
					t.FailNow()
				}

				inputs := makeDefaultInputs(compiledWfc.Primary.Template.Interface)

				dotFormat := visualize.ToGraphViz(compiledWfc.Primary)
				t.Logf("GraphViz Dot: %v\n", dotFormat)

				flyteWf, err := k8s.BuildFlyteWorkflow(compiledWfc,
					inputs,
					&core.WorkflowExecutionIdentifier{
						Project: "hello",
						Domain:  "domain",
						Name:    "name",
					},
					"namespace")
				if !assert.NoError(t, err) {
					t.FailNow()
				}

				if !storeOrDiff(t, yaml.Marshal, flyteWf, filepath.Join(filepath.Dir(filepath.Dir(p)), "k8s", strings.TrimRight(filepath.Base(p), filepath.Ext(p))+".yaml")) {
					t.FailNow()
				}
			})
		}
	})
}
