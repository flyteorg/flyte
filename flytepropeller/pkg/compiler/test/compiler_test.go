package test

import (
	"encoding/json"
	"flag"
	"os"
	"path/filepath"
	"reflect"
	"regexp"
	"strings"
	"testing"
	"time"

	"github.com/go-test/deep"
	"github.com/stretchr/testify/assert"
	"google.golang.org/protobuf/jsonpb"
	"google.golang.org/protobuf/proto"
	"k8s.io/apimachinery/pkg/util/sets"

	"github.com/flyteorg/flyte/flyteidl/clients/go/coreutils"
	"github.com/flyteorg/flyte/flyteidl/gen/pb-go/flyteidl/admin"
	"github.com/flyteorg/flyte/flyteidl/gen/pb-go/flyteidl/core"
	"github.com/flyteorg/flyte/flytepropeller/pkg/compiler"
	"github.com/flyteorg/flyte/flytepropeller/pkg/compiler/common"
	"github.com/flyteorg/flyte/flytepropeller/pkg/compiler/errors"
	"github.com/flyteorg/flyte/flytepropeller/pkg/compiler/transformers/k8s"
	"github.com/flyteorg/flyte/flytepropeller/pkg/visualize"
	"github.com/flyteorg/flyte/flytestdlib/utils"
)

var update = flag.Bool("update", false, "Update .golden files")

func makeDefaultInputs(iface *core.TypedInterface) *core.LiteralMap {
	if iface == nil || iface.GetInputs() == nil {
		return nil
	}

	res := make(map[string]*core.Literal, len(iface.GetInputs().GetVariables()))
	for inputName, inputVar := range iface.GetInputs().GetVariables() {
		// A workaround because the coreutils don't support the "StructuredDataSet" type
		if reflect.TypeOf(inputVar.GetType().GetType()) == reflect.TypeOf(&core.LiteralType_StructuredDatasetType{}) {
			res[inputName] = &core.Literal{
				Value: &core.Literal_Scalar{
					Scalar: &core.Scalar{
						Value: &core.Scalar_StructuredDataset{
							StructuredDataset: &core.StructuredDataset{
								Metadata: &core.StructuredDatasetMetadata{
									StructuredDatasetType: inputVar.GetType().GetType().(*core.LiteralType_StructuredDatasetType).StructuredDatasetType,
								},
							},
						},
					},
				},
			}
		} else if reflect.TypeOf(inputVar.GetType().GetType()) == reflect.TypeOf(&core.LiteralType_Simple{}) && inputVar.GetType().GetSimple() == core.SimpleType_DATETIME {
			res[inputName] = coreutils.MustMakeLiteral(time.UnixMicro(10))
		} else {
			res[inputName] = coreutils.MustMakeDefaultLiteralForType(inputVar.GetType())
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

			raw, err := os.ReadFile(path)
			assert.NoError(t, err)
			wf := &core.DynamicJobSpec{}
			err = utils.UnmarshalBytesToPb(raw, wf)
			if !assert.NoError(t, err) {
				t.FailNow()
			}

			t.Log("Compiling Workflow")
			compiledTasks := mustCompileTasks(t, wf.GetTasks())
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
				Nodes:   wf.GetNodes(),
				Outputs: wf.GetOutputs(),
			}
			compiledWfc, err := compiler.CompileWorkflow(wfTemplate, wf.GetSubworkflows(), compiledTasks,
				[]common.InterfaceProvider{})
			if !assert.NoError(t, err) {
				t.FailNow()
			}

			inputs := makeDefaultInputs(compiledWfc.GetPrimary().GetTemplate().GetInterface())

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
		thenNode := branchNode.GetIfElse().GetCase().GetThenNode()
		if hasPromiseInputs(thenNode.GetInputs()) {
			res.Insert(thenNode.GetId())
		}

		res = res.Union(getAllSubNodeIDs(thenNode))

		for _, other := range branchNode.GetIfElse().GetOther() {
			if hasPromiseInputs(other.GetThenNode().GetInputs()) {
				res.Insert(other.GetThenNode().GetId())
			}

			res = res.Union(getAllSubNodeIDs(other.GetThenNode()))
		}

		if elseNode := branchNode.GetIfElse().GetElseNode(); elseNode != nil {
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
	for _, n := range wf.GetTemplate().GetNodes() {
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
		for _, d := range v.Collection.GetBindings() {
			if bindingHasPromiseInputs(d) {
				return true
			}
		}
	case *core.BindingData_Map:
		for _, d := range v.Map.GetBindings() {
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
		if bindingHasPromiseInputs(b.GetBinding()) {
			return true
		}
	}

	return false
}

func assertNodeIDsInConnections(t testing.TB, nodeIDsWithDeps, allNodeIDs sets.String, connections *core.ConnectionSet) bool {
	actualNodeIDs := sets.NewString()
	for id, lst := range connections.GetDownstream() {
		actualNodeIDs.Insert(id)
		actualNodeIDs.Insert(lst.GetIds()...)
	}

	for id, lst := range connections.GetUpstream() {
		actualNodeIDs.Insert(id)
		actualNodeIDs.Insert(lst.GetIds()...)
	}

	notFoundInConnections := nodeIDsWithDeps.Difference(actualNodeIDs)
	correct := assert.Empty(t, notFoundInConnections, "All nodes must appear in connections")

	notFoundInNodes := actualNodeIDs.Difference(allNodeIDs)
	correct = correct && assert.Empty(t, notFoundInNodes, "All connections must correspond to existing nodes")

	return correct
}

func TestUseCases(t *testing.T) {
	// This first test doesn't seem to do anything, all the paths are nil
	//runCompileTest(t, "branch")
	runCompileTest(t, "snacks-core")
}

func protoMarshal(v any) ([]byte, error) {
	m := jsonpb.Marshaler{}
	str, err := m.MarshalToString(v.(proto.Message))
	return []byte(str), err
}

var multiSpaces = regexp.MustCompile(`\s+`)

func storeOrDiff(t testing.TB, f func(obj any) ([]byte, error), obj any, path string) bool {
	raw, err := f(obj)
	if !assert.NoError(t, err) {
		return false
	}

	if *update {
		err = os.WriteFile(path, raw, os.ModePerm) // #nosec G306
		if !assert.NoError(t, err) {
			return false
		}

	} else {
		goldenRaw, err := os.ReadFile(path)
		if !assert.NoError(t, err) {
			return false
		}

		trimmedRaw := multiSpaces.ReplaceAllString(string(raw), " ")
		trimmedGolden := multiSpaces.ReplaceAllString(string(goldenRaw), " ")

		if diff := deep.Equal(trimmedRaw, trimmedGolden); diff != nil {
			t.Errorf("Compiled() Diff = %v\r\n got = %v\r\n want = %v", diff, trimmedRaw, trimmedGolden)
		}
	}

	return true
}

func runCompileTest(t *testing.T, dirName string) {
	errors.SetConfig(errors.Config{IncludeSource: true})
	// Compile Tasks
	compiledTasks := make(map[string]*core.CompiledTask)

	t.Run("tasks-"+dirName, func(t *testing.T) {
		paths, err := filepath.Glob("testdata/" + dirName + "/*.pb")
		if !assert.NoError(t, err) {
			t.FailNow()
		}

		for _, p := range paths {
			raw, err := os.ReadFile(p)
			assert.NoError(t, err)
			tsk := &admin.TaskSpec{}
			err = proto.Unmarshal(raw, tsk)
			if err != nil {
				t.Logf("Failed to parse %s as a Task, skipping: %v", p, err)
				continue
			}

			t.Run(p, func(t *testing.T) {
				inputTask := tsk.GetTemplate()
				setDefaultFields(inputTask)
				task, err := compiler.CompileTask(inputTask)
				if !assert.NoError(t, err) {
					t.FailNow()
				}
				compiledTasks[tsk.GetTemplate().GetId().String()] = task

				// unmarshal from json file to compare rather than yaml
				taskFile := filepath.Join(filepath.Dir(p), "compiled", strings.TrimRight(filepath.Base(p), filepath.Ext(p))+"_task.json")
				taskBytes, err := os.ReadFile(taskFile)
				assert.NoError(t, err)
				compiledTaskFromFile := &core.CompiledTask{}
				err = utils.UnmarshalBytesToPb(taskBytes, compiledTaskFromFile)
				assert.NoError(t, err)
				assert.True(t, proto.Equal(task, compiledTaskFromFile))
			})
		}
	})

	// Compile Workflows
	t.Run("workflows-"+dirName, func(t *testing.T) {
		paths, err := filepath.Glob(filepath.Join("testdata", dirName, "*.pb"))
		if !assert.NoError(t, err) {
			t.FailNow()
		}

		for _, p := range paths {
			raw, err := os.ReadFile(p)
			assert.NoError(t, err)
			wf := &core.WorkflowClosure{}
			err = proto.Unmarshal(raw, wf)
			if err != nil {
				t.Logf("Failed to parse %s as a WorkflowClosure, skipping: %v", p, err)
				continue
			}

			t.Run(p, func(t *testing.T) {
				inputWf := wf.GetWorkflow()

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

				allNodeIDs := getAllMatchingNodes(compiledWfc.GetPrimary(), allNodesPredicate)
				nodeIDsWithDeps := getAllMatchingNodes(compiledWfc.GetPrimary(), hasPromiseNodePredicate)
				if !assertNodeIDsInConnections(t, nodeIDsWithDeps, allNodeIDs, compiledWfc.GetPrimary().GetConnections()) {
					t.FailNow()
				}

				if !storeOrDiff(t, protoMarshal, compiledWfc, filepath.Join(filepath.Dir(p), "compiled", strings.TrimRight(filepath.Base(p), filepath.Ext(p))+"_wf.json")) {
					t.FailNow()
				}
			})
		}
	})

	// Build K8s Workflows
	t.Run("k8s-"+dirName, func(t *testing.T) {
		paths, err := filepath.Glob(filepath.Join("testdata", dirName, "compiled", "*_wf.json"))
		if !assert.NoError(t, err) {
			t.FailNow()
		}

		for _, p := range paths {
			t.Run(p, func(t *testing.T) {
				raw, err := os.ReadFile(p)
				if !assert.NoError(t, err) {
					t.FailNow()
				}

				compiledWfc := &core.CompiledWorkflowClosure{}
				if !assert.NoError(t, utils.UnmarshalBytesToPb(raw, compiledWfc)) {
					t.FailNow()
				}

				inputs := makeDefaultInputs(compiledWfc.GetPrimary().GetTemplate().GetInterface())

				dotFormat := visualize.ToGraphViz(compiledWfc.GetPrimary())
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

				file := filepath.Join(filepath.Dir(filepath.Dir(p)), "k8s", strings.TrimRight(filepath.Base(p), filepath.Ext(p))+"_crd.json")
				if !storeOrDiff(t, json.Marshal, flyteWf, file) {
					t.FailNow()
				}
			})
		}
	})
}
