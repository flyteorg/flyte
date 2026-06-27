package dataproxy

import (
	"context"
	"testing"

	"connectrpc.com/connect"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	"google.golang.org/protobuf/proto"

	"github.com/flyteorg/flyte/v2/flytestdlib/storage"
	storageMocks "github.com/flyteorg/flyte/v2/flytestdlib/storage/mocks"
	"github.com/flyteorg/flyte/v2/gen/go/flyteidl2/common"
	"github.com/flyteorg/flyte/v2/gen/go/flyteidl2/core"
	"github.com/flyteorg/flyte/v2/gen/go/flyteidl2/task"
	"github.com/flyteorg/flyte/v2/gen/go/flyteidl2/trigger"
	"github.com/flyteorg/flyte/v2/gen/go/flyteidl2/trigger/triggerconnect"
	"github.com/flyteorg/flyte/v2/gen/go/flyteidl2/workflow"
	workflowMocks "github.com/flyteorg/flyte/v2/gen/go/flyteidl2/workflow/workflowconnect/mocks"
)

// fakeTriggerClient implements triggerconnect.TriggerServiceClient but only supports
// GetTriggerRevisionDetails; all other methods panic via the embedded nil interface.
type fakeTriggerClient struct {
	triggerconnect.TriggerServiceClient
	resp *trigger.GetTriggerRevisionDetailsResponse
	err  error
}

func (f *fakeTriggerClient) GetTriggerRevisionDetails(
	context.Context,
	*connect.Request[trigger.GetTriggerRevisionDetailsRequest],
) (*connect.Response[trigger.GetTriggerRevisionDetailsResponse], error) {
	if f.err != nil {
		return nil, f.err
	}
	return connect.NewResponse(f.resp), nil
}

func testTriggerID() *common.TriggerIdentifier {
	return &common.TriggerIdentifier{
		Name: &common.TriggerName{
			Org:     "org",
			Project: "proj",
			Domain:  "dev",
			Name:    "t0",
		},
		Revision: 1,
	}
}

func testActionID() *common.ActionIdentifier {
	return &common.ActionIdentifier{
		Run: &common.RunIdentifier{
			Org:     "org",
			Project: "proj",
			Domain:  "dev",
			Name:    "run1",
		},
		Name: "a0",
	}
}

func testNamedLiterals() []*task.NamedLiteral {
	return []*task.NamedLiteral{
		{
			Name: "test",
			Value: &core.Literal{
				Value: &core.Literal_Scalar{
					Scalar: &core.Scalar{
						Value: &core.Scalar_Primitive{
							Primitive: &core.Primitive{
								Value: &core.Primitive_StringValue{StringValue: "hello world"},
							},
						},
					},
				},
			},
		},
	}
}

func testVariableMap() *core.VariableMap {
	return &core.VariableMap{
		Variables: []*core.VariableEntry{
			{
				Key: "test",
				Value: &core.Variable{
					Type: &core.LiteralType{
						Type: &core.LiteralType_Simple{Simple: core.SimpleType_STRING},
					},
				},
			},
		},
	}
}

func assertHelloWorldSchema(t *testing.T, resp *connect.Response[workflow.LiteralsToLaunchFormJsonResponse]) {
	t.Helper()
	schema := resp.Msg.GetJson().AsMap()
	properties, ok := schema["properties"].(map[string]any)
	require.True(t, ok)
	testField, ok := properties["test"].(map[string]any)
	require.True(t, ok)
	assert.Equal(t, "hello world", testField["default"])
}

func TestLiteralsToLaunchFormJson_Inline(t *testing.T) {
	svc := NewTranslatorService(nil, nil, nil)

	resp, err := svc.LiteralsToLaunchFormJson(context.Background(), connect.NewRequest(&workflow.LiteralsToLaunchFormJsonRequest{
		Literals:  testNamedLiterals(),
		Variables: testVariableMap(),
	}))

	require.NoError(t, err)
	assertHelloWorldSchema(t, resp)
}

func TestLiteralsToLaunchFormJson_OffloadedURI(t *testing.T) {
	inputsURI := "s3://test-bucket/metadata/proj/dev/run1/a0/inputs.pb"
	storedInputs := &task.Inputs{Literals: testNamedLiterals()}

	runClient := workflowMocks.NewRunServiceClient(t)
	runClient.EXPECT().GetActionDataURIs(mock.Anything, mock.Anything).Return(
		connect.NewResponse(&workflow.GetActionDataURIsResponse{
			InputsUri:  inputsURI,
			OutputsUri: "s3://test-bucket/metadata/proj/dev/run1/a0/outputs.pb",
		}), nil)

	mockComposedStore := storageMocks.NewComposedProtobufStore(t)
	mockComposedStore.On("ReadProtobuf", mock.Anything, storage.DataReference(inputsURI), mock.Anything).
		Run(func(args mock.Arguments) {
			msg := args.Get(2).(proto.Message)
			proto.Reset(msg)
			proto.Merge(msg, storedInputs)
		}).Return(nil)

	svc := NewTranslatorService(&storage.DataStore{ComposedProtobufStore: mockComposedStore}, runClient, nil)

	resp, err := svc.LiteralsToLaunchFormJson(context.Background(), connect.NewRequest(&workflow.LiteralsToLaunchFormJsonRequest{
		Variables:   testVariableMap(),
		LiteralsUri: inputsURI,
		Owner:       &workflow.LiteralsToLaunchFormJsonRequest_ActionId{ActionId: testActionID()},
	}))

	require.NoError(t, err)
	assertHelloWorldSchema(t, resp)
}

func TestLiteralsToLaunchFormJson_Trigger(t *testing.T) {
	inputsURI := "s3://test-bucket/metadata/proj/dev/triggers/t0/inputs.pb"
	storedInputs := &task.Inputs{Literals: testNamedLiterals()}

	triggerClient := &fakeTriggerClient{
		resp: &trigger.GetTriggerRevisionDetailsResponse{
			Trigger: &trigger.TriggerDetails{
				Id: testTriggerID(),
				Spec: &trigger.TriggerSpec{
					InputWrapper: &trigger.TriggerSpec_OffloadedInputData{
						OffloadedInputData: &common.OffloadedInputData{
							Uri:        inputsURI,
							InputsHash: "hash",
						},
					},
				},
			},
		},
	}

	mockComposedStore := storageMocks.NewComposedProtobufStore(t)
	mockComposedStore.On("ReadProtobuf", mock.Anything, storage.DataReference(inputsURI), mock.Anything).
		Run(func(args mock.Arguments) {
			msg := args.Get(2).(proto.Message)
			proto.Reset(msg)
			proto.Merge(msg, storedInputs)
		}).Return(nil)

	svc := NewTranslatorService(&storage.DataStore{ComposedProtobufStore: mockComposedStore}, nil, triggerClient)

	resp, err := svc.LiteralsToLaunchFormJson(context.Background(), connect.NewRequest(&workflow.LiteralsToLaunchFormJsonRequest{
		Variables: testVariableMap(),
		Owner:     &workflow.LiteralsToLaunchFormJsonRequest_TriggerId{TriggerId: testTriggerID()},
	}))

	require.NoError(t, err)
	assertHelloWorldSchema(t, resp)
}

func TestLiteralsToLaunchFormJson_Trigger_NoOffloadedData(t *testing.T) {
	triggerClient := &fakeTriggerClient{
		resp: &trigger.GetTriggerRevisionDetailsResponse{
			Trigger: &trigger.TriggerDetails{
				Id:   testTriggerID(),
				Spec: &trigger.TriggerSpec{},
			},
		},
	}

	svc := NewTranslatorService(nil, nil, triggerClient)

	_, err := svc.LiteralsToLaunchFormJson(context.Background(), connect.NewRequest(&workflow.LiteralsToLaunchFormJsonRequest{
		Variables: testVariableMap(),
		Owner:     &workflow.LiteralsToLaunchFormJsonRequest_TriggerId{TriggerId: testTriggerID()},
	}))

	require.Error(t, err)
	assert.Equal(t, connect.CodeInvalidArgument, connect.CodeOf(err))
	assert.Contains(t, err.Error(), "no offloaded input data")
}

func TestLiteralsToLaunchFormJson_OffloadedURI_MissingActionId(t *testing.T) {
	svc := NewTranslatorService(nil, nil, nil)

	_, err := svc.LiteralsToLaunchFormJson(context.Background(), connect.NewRequest(&workflow.LiteralsToLaunchFormJsonRequest{
		Variables:   testVariableMap(),
		LiteralsUri: "s3://test-bucket/metadata/proj/dev/run1/a0/inputs.pb",
	}))

	require.Error(t, err)
	assert.Equal(t, connect.CodeInvalidArgument, connect.CodeOf(err))
	assert.Contains(t, err.Error(), "action_id is required")
}

func TestLiteralsToLaunchFormJson_OffloadedURI_Mismatch(t *testing.T) {
	runClient := workflowMocks.NewRunServiceClient(t)
	runClient.EXPECT().GetActionDataURIs(mock.Anything, mock.Anything).Return(
		connect.NewResponse(&workflow.GetActionDataURIsResponse{
			InputsUri:  "s3://test-bucket/metadata/proj/dev/run1/a0/inputs.pb",
			OutputsUri: "s3://test-bucket/metadata/proj/dev/run1/a0/outputs.pb",
		}), nil)

	svc := NewTranslatorService(nil, runClient, nil)

	_, err := svc.LiteralsToLaunchFormJson(context.Background(), connect.NewRequest(&workflow.LiteralsToLaunchFormJsonRequest{
		Variables:   testVariableMap(),
		LiteralsUri: "s3://test-bucket/some/other/object.pb",
		Owner:       &workflow.LiteralsToLaunchFormJsonRequest_ActionId{ActionId: testActionID()},
	}))

	require.Error(t, err)
	assert.Equal(t, connect.CodeInvalidArgument, connect.CodeOf(err))
	assert.Contains(t, err.Error(), "does not match any data URI")
}
