package subworkflow

import (
	"context"
	"reflect"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"

	"github.com/flyteorg/flyte/flyteidl/clients/go/coreutils"
	"github.com/flyteorg/flyte/flyteidl/gen/pb-go/flyteidl/core"
	iomocks "github.com/flyteorg/flyte/flyteplugins/go/tasks/pluginmachinery/io/mocks"
	"github.com/flyteorg/flyte/flytepropeller/pkg/apis/flyteworkflow/v1alpha1"
	"github.com/flyteorg/flyte/flytepropeller/pkg/apis/flyteworkflow/v1alpha1/mocks"
	"github.com/flyteorg/flyte/flytepropeller/pkg/compiler/validators"
	executorsmocks "github.com/flyteorg/flyte/flytepropeller/pkg/controller/executors/mocks"
	"github.com/flyteorg/flyte/flytepropeller/pkg/controller/nodes/interfaces"
	interfacesmocks "github.com/flyteorg/flyte/flytepropeller/pkg/controller/nodes/interfaces/mocks"
	recoverymocks "github.com/flyteorg/flyte/flytepropeller/pkg/controller/nodes/recovery/mocks"
	launchplanmocks "github.com/flyteorg/flyte/flytepropeller/pkg/controller/nodes/subworkflow/launchplan/mocks"
	"github.com/flyteorg/flyte/flytestdlib/promutils"
)

func compileVariableMap(literalMap *core.LiteralMap) *core.VariableMap {
	variableMap := make(map[string]*core.Variable)
	for name, literal := range literalMap.Literals {
		variableMap[name] = &core.Variable{
			Type: validators.LiteralTypeForLiteral(literal),
		}
	}

	return &core.VariableMap{
		Variables: variableMap,
	}
}

func TestGetCatalogKey(t *testing.T) {
	identifier := &core.Identifier{
		Project: "foo",
		Domain:  "bar",
		Name:    "baz",
		Version: "bat",
	}

	inputs := &core.LiteralMap{
		Literals: map[string]*core.Literal{
			"i0": coreutils.MustMakeLiteral("foo"),
		},
	}

	fixedInputs := &core.LiteralMap{
		Literals: map[string]*core.Literal{
			"i1": coreutils.MustMakeLiteral("bar"),
		},
	}

	mergedLiterals := make(map[string]*core.Literal)
	for k, v := range inputs.Literals {
		mergedLiterals[k] = v
	}
	for k, v := range fixedInputs.Literals {
		mergedLiterals[k] = v
	}
	mergedInputs := &core.LiteralMap{
		Literals: mergedLiterals,
	}

	typedInterface := &core.TypedInterface{
		Inputs: compileVariableMap(inputs),
		Outputs: &core.VariableMap{
			Variables: map[string]*core.Variable{
				"o0": {
					Type: &core.LiteralType{
						Type: &core.LiteralType_Simple{
							Simple: core.SimpleType_INTEGER,
						},
					},
				},
			},
		},
	}

	mergedInterface := &core.TypedInterface{
		Inputs:  compileVariableMap(mergedInputs),
		Outputs: typedInterface.Outputs,
	}

	tests := []struct {
		name              string
		resourceType      core.ResourceType
		fixedInputs       *core.LiteralMap
		expectedInterface *core.TypedInterface
		expectedInputs    *core.LiteralMap
	}{
		{
			name:              "Subworkflow",
			resourceType:      core.ResourceType_WORKFLOW,
			expectedInterface: typedInterface,
			expectedInputs:    inputs,
		},
		{
			name:              "Launchplan",
			resourceType:      core.ResourceType_LAUNCH_PLAN,
			expectedInterface: typedInterface,
			expectedInputs:    inputs,
		},
		{
			name:              "LaunchplanFixedInputs",
			resourceType:      core.ResourceType_LAUNCH_PLAN,
			fixedInputs:       fixedInputs,
			expectedInterface: mergedInterface,
			expectedInputs:    mergedInputs,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			nodeID := *identifier
			nodeID.ResourceType = tt.resourceType

			// mock everything
			mockWorkflowNode := &mocks.ExecutableWorkflowNode{}
			mockExecutionContext := &executorsmocks.ExecutionContext{}
			nCtx := &interfacesmocks.NodeExecutionContext{}

			switch tt.resourceType {
			case core.ResourceType_WORKFLOW:
				subworkflowID := "foo"
				mockWorkflowNode.OnGetSubWorkflowRef().Return(&subworkflowID)
				mockWorkflowNode.OnGetLaunchPlanRefID().Return(nil)

				mockSubWorkflow := &mocks.ExecutableSubWorkflow{}
				mockSubWorkflow.OnGetIdentifier().Return(identifier)
				mockSubWorkflow.OnGetInterface().Return(typedInterface)

				mockExecutionContext.OnFindSubWorkflow(mock.Anything).Return(mockSubWorkflow)
			case core.ResourceType_LAUNCH_PLAN:
				mockWorkflowNode.OnGetSubWorkflowRef().Return(nil)
				launchPlanID := &v1alpha1.Identifier{
					Identifier: identifier,
				}
				mockWorkflowNode.OnGetLaunchPlanRefID().Return(launchPlanID)

				mockLaunchPlan := &mocks.ExecutableLaunchPlan{}
				mockLaunchPlan.OnGetId().Return(identifier)
				mockLaunchPlan.OnGetInterface().Return(typedInterface)
				mockLaunchPlan.OnGetFixedInputs().Return(tt.fixedInputs)

				mockExecutionContext.OnFindLaunchPlanMatch(mock.Anything).Return(mockLaunchPlan)
			}

			mockInputReader := &iomocks.InputReader{}
			mockInputReader.OnGetMatch(mock.Anything).Return(inputs, nil)
			nCtx.OnInputReader().Return(mockInputReader)

			mockNode := &mocks.ExecutableNode{}
			mockNode.OnGetWorkflowNode().Return(mockWorkflowNode)

			nCtx.OnNode().Return(mockNode)
			nCtx.OnExecutionContext().Return(mockExecutionContext)

			mockLaunchPlanExecutor := &launchplanmocks.Executor{}
			mockRecoveryClient := &recoverymocks.Client{}

			handler := New(nil, mockLaunchPlanExecutor, mockRecoveryClient, eventConfig, promutils.NewTestScope())
			cacheableHandler, ok := handler.(interfaces.CacheableNodeHandler)
			assert.True(t, ok)

			// execute GetCatalogKey
			catalogKeyResult, err := cacheableHandler.GetCatalogKey(context.Background(), nCtx)
			assert.NoError(t, err)

			// validate the result
			assert.True(t, reflect.DeepEqual(catalogKeyResult.Identifier, *identifier))
			assert.True(t, reflect.DeepEqual(catalogKeyResult.TypedInterface, *tt.expectedInterface))

			inputs, err := catalogKeyResult.InputReader.Get(context.Background())
			assert.NoError(t, err)
			assert.True(t, reflect.DeepEqual(*inputs, *tt.expectedInputs))
		})
	}
}

func TestIsCacheable(t *testing.T) {
	nCtx := &interfacesmocks.NodeExecutionContext{}

	mockLaunchPlanExecutor := &launchplanmocks.Executor{}
	mockRecoveryClient := &recoverymocks.Client{}

	handler := New(nil, mockLaunchPlanExecutor, mockRecoveryClient, eventConfig, promutils.NewTestScope())
	cacheableHandler, ok := handler.(interfaces.CacheableNodeHandler)
	assert.True(t, ok)

	// execute IsCacheable
	cacheable, cacheSerializable, err := cacheableHandler.IsCacheable(context.Background(), nCtx)
	assert.NoError(t, err)

	// validate results
	assert.False(t, cacheable)
	assert.False(t, cacheSerializable)
}
