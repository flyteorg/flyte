package test

import (
	"encoding/json"
	"flag"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"
	"testing"

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
	"github.com/flyteorg/flytepropeller/pkg/visualize"
	"github.com/golang/protobuf/jsonpb"
	"github.com/golang/protobuf/proto"
	"github.com/stretchr/testify/assert"
)

var update = flag.Bool("update", false, "Update .golden files")
var reverse = flag.Bool("reverse", false, "Reverse .golden files")

func makeDefaultInputs(iface *core.TypedInterface) *core.LiteralMap {
	if iface == nil || iface.GetInputs() == nil {
		return nil
	}

	res := make(map[string]*core.Literal, len(iface.GetInputs().Variables))
	for _, e := range iface.GetInputs().Variables {
		val := coreutils.MustMakeDefaultLiteralForType(e.GetVar().Type)
		res[e.GetName()] = val
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

func marshalProto(t *testing.T, filename string, p proto.Message) {
	marshaller := &jsonpb.Marshaler{}
	s, err := marshaller.MarshalToString(p)
	assert.NoError(t, err)

	if err != nil {
		return
	}

	originalRaw, err := proto.Marshal(p)
	assert.NoError(t, err)
	assert.NoError(t, ioutil.WriteFile(strings.Replace(filename, filepath.Ext(filename), ".pb", 1), originalRaw, os.ModePerm))

	m := map[string]interface{}{}
	err = json.Unmarshal([]byte(s), &m)
	assert.NoError(t, err)

	b, err := yaml.Marshal(m)
	assert.NoError(t, err)
	assert.NoError(t, ioutil.WriteFile(strings.Replace(filename, filepath.Ext(filename), ".yaml", 1), b, os.ModePerm))
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
					Inputs: &core.VariableMap{Variables: []*core.VariableMapEntry{}},
					Outputs: &core.VariableMap{Variables: []*core.VariableMapEntry{
						{
							Name: "o0",
							Var: &core.Variable{
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

			inputs := map[string]interface{}{}
			for _, e := range compiledWfc.Primary.Template.Interface.Inputs.Variables {
				inputs[e.GetName()] = coreutils.MustMakeDefaultLiteralForType(e.GetVar().Type)
			}

			flyteWf, err := k8s.BuildFlyteWorkflow(compiledWfc,
				coreutils.MustMakeLiteral(inputs).GetMap(),
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

func TestBranches(t *testing.T) {
	errors.SetConfig(errors.Config{IncludeSource: true})
	assert.NoError(t, filepath.Walk("testdata/branch", func(path string, info os.FileInfo, err error) error {
		if info.IsDir() {
			if filepath.Base(info.Name()) != "branch" {
				return filepath.SkipDir
			}

			return nil
		}

		t.Run(path, func(t *testing.T) {
			// If you want to debug a single use-case. Uncomment this line.
			//if !strings.HasSuffix(path, "mycereal_condition_has_no_deps.json") {
			//	t.SkipNow()
			//}

			raw, err := ioutil.ReadFile(path)
			assert.NoError(t, err)
			wf := &core.WorkflowClosure{}
			if filepath.Ext(path) == ".json" {
				err = jsonpb.UnmarshalString(string(raw), wf)
				if !assert.NoError(t, err) {
					t.FailNow()
				}
			} else if filepath.Ext(path) == ".pb" {
				m := &jsonpb.Marshaler{
					Indent: "  ",
				}

				err = proto.Unmarshal(raw, wf)
				if !assert.NoError(t, err) {
					tsk := &admin.TaskSpec{}
					if !assert.NoError(t, proto.Unmarshal(raw, tsk)) {
						t.FailNow()
					}

					raw, _ := m.MarshalToString(tsk)
					err = ioutil.WriteFile(strings.TrimSuffix(path, filepath.Ext(path))+"_task.json", []byte(raw), os.ModePerm)
					if !assert.NoError(t, err) {
						t.FailNow()
					}

					return
				}

				raw, err := m.MarshalToString(wf)
				if !assert.NoError(t, err) {
					t.FailNow()
				}

				err = ioutil.WriteFile(strings.TrimSuffix(path, filepath.Ext(path))+".json", []byte(raw), os.ModePerm)
				if !assert.NoError(t, err) {
					t.FailNow()
				}
			}

			t.Log("Compiling Workflow")
			compiledTasks := mustCompileTasks(t, wf.Tasks)
			compiledWfc, err := compiler.CompileWorkflow(wf.Workflow, []*core.WorkflowTemplate{}, compiledTasks,
				[]common.InterfaceProvider{})
			if !assert.NoError(t, err) {
				t.FailNow()
			}

			m := &jsonpb.Marshaler{
				Indent: "  ",
			}
			rawStr, err := m.MarshalToString(compiledWfc)
			if !assert.NoError(t, err) {
				t.Fail()
			}

			compiledFilePath := filepath.Join(filepath.Dir(path), "compiled", filepath.Base(path))
			if *update {
				err = ioutil.WriteFile(compiledFilePath, []byte(rawStr), os.ModePerm)
				if !assert.NoError(t, err) {
					t.Fail()
				}
			} else {
				goldenRaw, err := ioutil.ReadFile(compiledFilePath)
				if !assert.NoError(t, err) {
					t.Fail()
				}

				if diff := deep.Equal(rawStr, string(goldenRaw)); diff != nil {
					t.Errorf("Compiled() Diff = %v\r\n got = %v\r\n want = %v\nYou might need to run `make golden` to generate compiled json files.", diff, rawStr, string(goldenRaw))
				}
			}

			allNodeIDs := getAllMatchingNodes(compiledWfc.Primary, allNodesPredicate)
			nodeIDsWithDeps := getAllMatchingNodes(compiledWfc.Primary, hasPromiseNodePredicate)
			if !assertNodeIDsInConnections(t, nodeIDsWithDeps, allNodeIDs, compiledWfc.Primary.Connections) {
				t.FailNow()
			}

			inputs := map[string]interface{}{}
			for _, e := range compiledWfc.Primary.Template.Interface.Inputs.Variables {
				inputs[e.GetName()] = coreutils.MustMakeDefaultLiteralForType(e.GetVar().Type)
			}

			flyteWf, err := k8s.BuildFlyteWorkflow(compiledWfc,
				coreutils.MustMakeLiteral(inputs).GetMap(),
				&core.WorkflowExecutionIdentifier{
					Project: "hello",
					Domain:  "domain",
					Name:    "name",
				},
				"namespace")
			if assert.NoError(t, err) {
				raw, err := json.MarshalIndent(flyteWf, "", "  ")
				if assert.NoError(t, err) {
					assert.NotEmpty(t, raw)
				}

				k8sObjectFilepath := filepath.Join(filepath.Dir(path), "k8s", filepath.Base(path))
				if *update {
					err = ioutil.WriteFile(k8sObjectFilepath, raw, os.ModePerm)
					if !assert.NoError(t, err) {
						t.Fail()
					}
				} else {
					goldenRaw, err := ioutil.ReadFile(k8sObjectFilepath)
					if !assert.NoError(t, err) {
						t.Fail()
					}

					if diff := deep.Equal(string(raw), string(goldenRaw)); diff != nil {
						t.Errorf("K8sObject() Diff = %v\r\n got = %v\r\n want = %v\nYou might need to run `make golden` to generate k8s object json files.", diff, rawStr, string(goldenRaw))
					}

				}
			}
		})

		return nil
	}))
}

func TestReverseEngineerFromYaml(t *testing.T) {
	root := "testdata"
	errors.SetConfig(errors.Config{IncludeSource: true})
	assert.NoError(t, filepath.Walk(root, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		if info.IsDir() {
			return nil
		}

		if !strings.HasSuffix(path, ".yaml") {
			return nil
		}

		if strings.HasSuffix(path, "-inputs.yaml") {
			return nil
		}

		ext := ".yaml"

		testName := strings.TrimLeft(path, root)
		testName = strings.Trim(testName, string(os.PathSeparator))
		testName = strings.TrimSuffix(testName, ext)
		testName = strings.Replace(testName, string(os.PathSeparator), "_", -1)

		t.Run(testName, func(t *testing.T) {
			t.Log("Reading from file")
			raw, err := ioutil.ReadFile(path)
			assert.NoError(t, err)

			raw, err = yaml.YAMLToJSON(raw)
			assert.NoError(t, err)

			t.Log("Unmarshalling Workflow Closure")
			wf := &core.WorkflowClosure{}
			err = jsonpb.UnmarshalString(string(raw), wf)
			assert.NoError(t, err)
			assert.NotNil(t, wf)
			if err != nil {
				return
			}

			t.Log("Compiling Workflow")
			compiledWf, err := compiler.CompileWorkflow(wf.Workflow, []*core.WorkflowTemplate{}, mustCompileTasks(t, wf.Tasks), []common.InterfaceProvider{})
			assert.NoError(t, err)
			if err != nil {
				return
			}

			inputs := makeDefaultInputs(compiledWf.Primary.Template.GetInterface())
			if *reverse {
				marshalProto(t, strings.Replace(path, ext, fmt.Sprintf("-inputs%v", ext), -1), inputs)
			}

			t.Log("Building k8s resource")
			_, err = k8s.BuildFlyteWorkflow(compiledWf, inputs, nil, "")
			assert.NoError(t, err)
			if err != nil {
				return
			}

			dotFormat := visualize.ToGraphViz(compiledWf.Primary)
			t.Logf("GraphViz Dot: %v\n", dotFormat)

			if *reverse {
				marshalProto(t, path, wf)
			}
		})

		return nil
	}))
}

func TestCompileAndBuild(t *testing.T) {
	root := "testdata"
	errors.SetConfig(errors.Config{IncludeSource: true})
	assert.NoError(t, filepath.Walk(root, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		if info.IsDir() {
			return nil
		}

		if ext := filepath.Ext(path); ext != ".pb" {
			return nil
		}

		if strings.HasSuffix(path, "-inputs.pb") {
			return nil
		}

		testName := strings.TrimLeft(path, root)
		testName = strings.Trim(testName, string(os.PathSeparator))
		testName = strings.Trim(testName, filepath.Ext(testName))
		testName = strings.Replace(testName, string(os.PathSeparator), "_", -1)

		t.Run(testName, func(t *testing.T) {
			t.Log("Reading from file")
			raw, err := ioutil.ReadFile(path)
			assert.NoError(t, err)

			t.Log("Unmarshalling Workflow Closure")
			wf := &core.WorkflowClosure{}
			err = proto.Unmarshal(raw, wf)
			assert.NoError(t, err)
			assert.NotNil(t, wf)
			if err != nil {
				return
			}

			t.Log("Compiling Workflow")
			compiledWf, err := compiler.CompileWorkflow(wf.Workflow, []*core.WorkflowTemplate{}, mustCompileTasks(t, wf.Tasks), []common.InterfaceProvider{})
			assert.NoError(t, err)
			if err != nil {
				return
			}

			inputs := makeDefaultInputs(compiledWf.Primary.Template.GetInterface())
			if *update {
				marshalProto(t, strings.Replace(path, filepath.Ext(path), fmt.Sprintf("-inputs%v", filepath.Ext(path)), -1), inputs)
			}

			t.Log("Building k8s resource")
			_, err = k8s.BuildFlyteWorkflow(compiledWf, inputs, nil, "")
			assert.NoError(t, err)
			if err != nil {
				return
			}

			dotFormat := visualize.ToGraphViz(compiledWf.Primary)
			t.Logf("GraphViz Dot: %v\n", dotFormat)

			if *update {
				marshalProto(t, path, wf)
			}
		})

		return nil
	}))
}
