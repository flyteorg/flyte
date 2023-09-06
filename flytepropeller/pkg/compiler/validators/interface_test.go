package validators

import (
	"reflect"
	"testing"
	"time"

	"github.com/flyteorg/flyteidl/clients/go/coreutils"
	"github.com/flyteorg/flyteidl/gen/pb-go/flyteidl/core"

	c "github.com/flyteorg/flytepropeller/pkg/compiler/common"
	"github.com/flyteorg/flytepropeller/pkg/compiler/common/mocks"
	"github.com/flyteorg/flytepropeller/pkg/compiler/errors"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"

	"google.golang.org/protobuf/types/known/durationpb"
)

func TestValidateInterface(t *testing.T) {
	t.Run("Happy path", func(t *testing.T) {
		errs := errors.NewCompileErrors()
		iface, ok := ValidateInterface(
			c.NodeID("node1"),
			&core.TypedInterface{
				Inputs: &core.VariableMap{
					Variables: map[string]*core.Variable{},
				},
				Outputs: &core.VariableMap{
					Variables: map[string]*core.Variable{},
				},
			},
			errs.NewScope(),
		)

		assertNonEmptyInterface(t, iface, ok, errs)
	})

	t.Run("Empty Inputs/Outputs", func(t *testing.T) {
		errs := errors.NewCompileErrors()
		iface, ok := ValidateInterface(
			c.NodeID("node1"),
			&core.TypedInterface{},
			errs.NewScope(),
		)

		assertNonEmptyInterface(t, iface, ok, errs)
	})

	t.Run("Empty Interface", func(t *testing.T) {
		errs := errors.NewCompileErrors()
		iface, ok := ValidateInterface(
			c.NodeID("node1"),
			nil,
			errs.NewScope(),
		)

		assertNonEmptyInterface(t, iface, ok, errs)
	})
}

func assertNonEmptyInterface(t testing.TB, iface *core.TypedInterface, ifaceOk bool, errs errors.CompileErrors) {
	assert.True(t, ifaceOk)
	assert.NotNil(t, iface)
	assert.False(t, errs.HasErrors())
	if !ifaceOk {
		t.Fatal(errs)
	}

	assert.NotNil(t, iface.Inputs)
	assert.NotNil(t, iface.Inputs.Variables)
	assert.NotNil(t, iface.Outputs)
	assert.NotNil(t, iface.Outputs.Variables)
}

func TestValidateUnderlyingInterface(t *testing.T) {
	t.Run("Invalid empty node", func(t *testing.T) {
		wfBuilder := mocks.WorkflowBuilder{}
		nodeBuilder := mocks.NodeBuilder{}
		nodeBuilder.OnGetCoreNode().Return(&core.Node{})
		nodeBuilder.OnGetId().Return("node_1")
		nodeBuilder.OnGetInterface().Return(nil)
		errs := errors.NewCompileErrors()
		iface, ifaceOk := ValidateUnderlyingInterface(&wfBuilder, &nodeBuilder, errs.NewScope())
		assert.False(t, ifaceOk)
		assert.Nil(t, iface)
		assert.True(t, errs.HasErrors())
	})

	t.Run("Task Node", func(t *testing.T) {
		task := mocks.Task{}
		task.On("GetInterface").Return(nil)

		wfBuilder := mocks.WorkflowBuilder{}
		wfBuilder.On("GetTask", mock.MatchedBy(func(id core.Identifier) bool {
			return id.String() == (&core.Identifier{
				Name: "Task_1",
			}).String()
		})).Return(&task, true)

		taskNode := &core.TaskNode{
			Reference: &core.TaskNode_ReferenceId{
				ReferenceId: &core.Identifier{
					Name: "Task_1",
				},
			},
		}

		nodeBuilder := mocks.NodeBuilder{}
		nodeBuilder.On("GetCoreNode").Return(&core.Node{
			Target: &core.Node_TaskNode{
				TaskNode: taskNode,
			},
		})
		nodeBuilder.OnGetInterface().Return(nil)

		nodeBuilder.On("GetTaskNode").Return(taskNode)
		nodeBuilder.On("GetId").Return("node_1")
		nodeBuilder.On("SetInterface", mock.Anything).Return()

		errs := errors.NewCompileErrors()
		iface, ifaceOk := ValidateUnderlyingInterface(&wfBuilder, &nodeBuilder, errs.NewScope())
		assertNonEmptyInterface(t, iface, ifaceOk, errs)
	})

	t.Run("Workflow Node", func(t *testing.T) {
		wfBuilder := mocks.WorkflowBuilder{}
		wfBuilder.On("GetCoreWorkflow").Return(&core.CompiledWorkflow{
			Template: &core.WorkflowTemplate{
				Id: &core.Identifier{
					Name: "Ref_1",
				},
			},
		})
		workflowNode := &core.WorkflowNode{
			Reference: &core.WorkflowNode_LaunchplanRef{
				LaunchplanRef: &core.Identifier{
					Name: "Ref_1",
				},
			},
		}

		nodeBuilder := mocks.NodeBuilder{}
		nodeBuilder.On("GetCoreNode").Return(&core.Node{
			Target: &core.Node_WorkflowNode{
				WorkflowNode: workflowNode,
			},
		})

		nodeBuilder.On("GetWorkflowNode").Return(workflowNode)
		nodeBuilder.On("GetId").Return("node_1")
		nodeBuilder.On("SetInterface", mock.Anything).Return()
		nodeBuilder.On("GetInputs").Return([]*core.Binding{})
		nodeBuilder.OnGetInterface().Return(nil)

		t.Run("Self", func(t *testing.T) {
			errs := errors.NewCompileErrors()
			_, ifaceOk := ValidateUnderlyingInterface(&wfBuilder, &nodeBuilder, errs.NewScope())
			assert.False(t, ifaceOk)

			wfBuilder := mocks.WorkflowBuilder{}
			wfBuilder.On("GetCoreWorkflow").Return(&core.CompiledWorkflow{
				Template: &core.WorkflowTemplate{
					Id: &core.Identifier{
						Name: "Ref_1",
					},
					Interface: &core.TypedInterface{
						Inputs: &core.VariableMap{
							Variables: map[string]*core.Variable{},
						},
						Outputs: &core.VariableMap{
							Variables: map[string]*core.Variable{},
						},
					},
				},
			})

			errs = errors.NewCompileErrors()
			iface, ifaceOk := ValidateUnderlyingInterface(&wfBuilder, &nodeBuilder, errs.NewScope())
			assertNonEmptyInterface(t, iface, ifaceOk, errs)
		})

		t.Run("LP_Ref", func(t *testing.T) {
			lp := mocks.InterfaceProvider{}
			lp.On("GetID").Return(&core.Identifier{Name: "Ref_1"})
			lp.On("GetExpectedInputs").Return(&core.ParameterMap{
				Parameters: map[string]*core.Parameter{
					"required": {
						Var: &core.Variable{
							Type: &core.LiteralType{Type: &core.LiteralType_Simple{Simple: core.SimpleType_INTEGER}},
						},
						Behavior: &core.Parameter_Required{
							Required: true,
						},
					},
					"default_value": {
						Var: &core.Variable{
							Type: &core.LiteralType{Type: &core.LiteralType_Simple{Simple: core.SimpleType_INTEGER}},
						},
						Behavior: &core.Parameter_Default{
							Default: coreutils.MustMakeLiteral(5),
						},
					},
				},
			})
			lp.On("GetExpectedOutputs").Return(&core.VariableMap{})

			wfBuilder := mocks.WorkflowBuilder{}
			wfBuilder.On("GetCoreWorkflow").Return(&core.CompiledWorkflow{
				Template: &core.WorkflowTemplate{
					Id: &core.Identifier{
						Name: "Ref_2",
					},
				},
			})

			wfBuilder.On("GetLaunchPlan", mock.Anything).Return(nil, false)

			errs := errors.NewCompileErrors()
			_, ifaceOk := ValidateUnderlyingInterface(&wfBuilder, &nodeBuilder, errs.NewScope())
			assert.False(t, ifaceOk)

			wfBuilder = mocks.WorkflowBuilder{}
			wfBuilder.On("GetCoreWorkflow").Return(&core.CompiledWorkflow{
				Template: &core.WorkflowTemplate{
					Id: &core.Identifier{
						Name: "Ref_2",
					},
				},
			})

			wfBuilder.On("GetLaunchPlan", matchIdentifier(core.Identifier{Name: "Ref_1"})).Return(&lp, true)

			errs = errors.NewCompileErrors()
			iface, ifaceOk := ValidateUnderlyingInterface(&wfBuilder, &nodeBuilder, errs.NewScope())
			assertNonEmptyInterface(t, iface, ifaceOk, errs)
		})

		t.Run("Subwf", func(t *testing.T) {
			subWf := core.CompiledWorkflow{
				Template: &core.WorkflowTemplate{
					Interface: &core.TypedInterface{
						Inputs:  &core.VariableMap{},
						Outputs: &core.VariableMap{},
					},
				},
			}

			wfBuilder := mocks.WorkflowBuilder{}
			wfBuilder.On("GetCoreWorkflow").Return(&core.CompiledWorkflow{
				Template: &core.WorkflowTemplate{
					Id: &core.Identifier{
						Name: "Ref_2",
					},
				},
			})

			wfBuilder.On("GetLaunchPlan", mock.Anything).Return(nil, false)

			errs := errors.NewCompileErrors()
			_, ifaceOk := ValidateUnderlyingInterface(&wfBuilder, &nodeBuilder, errs.NewScope())
			assert.False(t, ifaceOk)

			wfBuilder = mocks.WorkflowBuilder{}
			wfBuilder.On("GetCoreWorkflow").Return(&core.CompiledWorkflow{
				Template: &core.WorkflowTemplate{
					Id: &core.Identifier{
						Name: "Ref_2",
					},
				},
			})

			wfBuilder.On("GetSubWorkflow", matchIdentifier(core.Identifier{Name: "Ref_1"})).Return(&subWf, true)

			workflowNode.Reference = &core.WorkflowNode_SubWorkflowRef{
				SubWorkflowRef: &core.Identifier{Name: "Ref_1"},
			}

			errs = errors.NewCompileErrors()
			iface, ifaceOk := ValidateUnderlyingInterface(&wfBuilder, &nodeBuilder, errs.NewScope())
			assertNonEmptyInterface(t, iface, ifaceOk, errs)
		})
	})

	t.Run("GateNode", func(t *testing.T) {
		t.Run("Approve", func(t *testing.T) {
			wfBuilder := mocks.WorkflowBuilder{}

			gateNode := &core.GateNode{
				Condition: &core.GateNode_Approve{
					Approve: &core.ApproveCondition{
						SignalId: "foo",
					},
				},
			}

			nodeBuilder := mocks.NodeBuilder{}
			nodeBuilder.On("GetCoreNode").Return(&core.Node{
				Target: &core.Node_GateNode{
					GateNode: gateNode,
				},
			})
			nodeBuilder.OnGetInterface().Return(nil)
			nodeBuilder.OnGetInputs().Return(nil)

			nodeBuilder.On("GetGateNode").Return(gateNode)
			nodeBuilder.On("GetId").Return("node_1")
			nodeBuilder.On("SetInterface", mock.Anything).Return()

			errs := errors.NewCompileErrors()
			iface, ifaceOk := ValidateUnderlyingInterface(&wfBuilder, &nodeBuilder, errs.NewScope())
			assertNonEmptyInterface(t, iface, ifaceOk, errs)
		})

		t.Run("Signal", func(t *testing.T) {
			wfBuilder := mocks.WorkflowBuilder{}

			gateNode := &core.GateNode{
				Condition: &core.GateNode_Signal{
					Signal: &core.SignalCondition{
						SignalId: "foo",
						Type: &core.LiteralType{
							Type: &core.LiteralType_Simple{
								Simple: core.SimpleType_BOOLEAN,
							},
						},
						OutputVariableName: "foo",
					},
				},
			}

			nodeBuilder := mocks.NodeBuilder{}
			nodeBuilder.On("GetCoreNode").Return(&core.Node{
				Target: &core.Node_GateNode{
					GateNode: gateNode,
				},
			})
			nodeBuilder.OnGetInterface().Return(nil)

			nodeBuilder.On("GetGateNode").Return(gateNode)
			nodeBuilder.On("GetId").Return("node_1")
			nodeBuilder.On("SetInterface", mock.Anything).Return()

			errs := errors.NewCompileErrors()
			iface, ifaceOk := ValidateUnderlyingInterface(&wfBuilder, &nodeBuilder, errs.NewScope())
			assertNonEmptyInterface(t, iface, ifaceOk, errs)
		})

		t.Run("Sleep", func(t *testing.T) {
			wfBuilder := mocks.WorkflowBuilder{}

			gateNode := &core.GateNode{
				Condition: &core.GateNode_Sleep{
					Sleep: &core.SleepCondition{
						Duration: durationpb.New(time.Minute),
					},
				},
			}

			nodeBuilder := mocks.NodeBuilder{}
			nodeBuilder.On("GetCoreNode").Return(&core.Node{
				Target: &core.Node_GateNode{
					GateNode: gateNode,
				},
			})
			nodeBuilder.OnGetInterface().Return(nil)

			nodeBuilder.On("GetGateNode").Return(gateNode)
			nodeBuilder.On("GetId").Return("node_1")
			nodeBuilder.On("SetInterface", mock.Anything).Return()

			errs := errors.NewCompileErrors()
			iface, ifaceOk := ValidateUnderlyingInterface(&wfBuilder, &nodeBuilder, errs.NewScope())
			assertNonEmptyInterface(t, iface, ifaceOk, errs)
		})
	})

	t.Run("ArrayNode", func(t *testing.T) {
		// mock underlying task node
		iface := &core.TypedInterface{
			Inputs: &core.VariableMap{
				Variables: map[string]*core.Variable{
					"foo": {
						Type: &core.LiteralType{
							Type: &core.LiteralType_Simple{
								Simple: core.SimpleType_INTEGER,
							},
						},
					},
				},
			},
			Outputs: &core.VariableMap{
				Variables: map[string]*core.Variable{
					"bar": {
						Type: &core.LiteralType{
							Type: &core.LiteralType_Simple{
								Simple: core.SimpleType_FLOAT,
							},
						},
					},
				},
			},
		}

		taskNode := &core.Node{
			Id: "node_1",
			Target: &core.Node_TaskNode{
				TaskNode: &core.TaskNode{
					Reference: &core.TaskNode_ReferenceId{
						ReferenceId: &core.Identifier{
							Name: "Task_1",
						},
					},
				},
			},
		}

		task := mocks.Task{}
		task.On("GetInterface").Return(iface)

		taskNodeBuilder := &mocks.NodeBuilder{}
		taskNodeBuilder.On("GetCoreNode").Return(taskNode)
		taskNodeBuilder.On("GetId").Return(taskNode.Id)
		taskNodeBuilder.On("GetTaskNode").Return(taskNode.Target.(*core.Node_TaskNode).TaskNode)
		taskNodeBuilder.On("GetInterface").Return(nil)
		taskNodeBuilder.On("SetInterface", mock.AnythingOfType("*core.TypedInterface")).Return(nil)

		wfBuilder := mocks.WorkflowBuilder{}
		wfBuilder.On("GetTask", mock.MatchedBy(func(id core.Identifier) bool {
			return id.String() == (&core.Identifier{
				Name: "Task_1",
			}).String()
		})).Return(&task, true)
		wfBuilder.On("GetOrCreateNodeBuilder", mock.MatchedBy(func(node *core.Node) bool {
			return node.Id == "node_1"
		})).Return(taskNodeBuilder)

		// mock array node
		arrayNode := &core.Node{
			Id: "node_2",
			Target: &core.Node_ArrayNode{
				ArrayNode: &core.ArrayNode{
					Node: taskNode,
				},
			},
		}

		nodeBuilder := mocks.NodeBuilder{}
		nodeBuilder.On("GetArrayNode").Return(arrayNode.Target.(*core.Node_ArrayNode).ArrayNode)
		nodeBuilder.On("GetCoreNode").Return(arrayNode)
		nodeBuilder.On("GetId").Return(arrayNode.Id)
		nodeBuilder.On("GetInterface").Return(nil)
		nodeBuilder.On("SetInterface", mock.Anything).Return()

		// compute arrayNode interface
		errs := errors.NewCompileErrors()
		arrayNodeIface, ifaceOk := ValidateUnderlyingInterface(&wfBuilder, &nodeBuilder, errs.NewScope())
		assertNonEmptyInterface(t, arrayNodeIface, ifaceOk, errs)
		assert.True(t, reflect.DeepEqual(arrayNodeIface, iface))
	})
}

func matchIdentifier(id core.Identifier) interface{} {
	return mock.MatchedBy(func(arg core.Identifier) bool {
		return arg.String() == id.String()
	})
}
