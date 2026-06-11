package dataproxy

import (
	"context"
	"fmt"

	"connectrpc.com/connect"

	"github.com/flyteorg/flyte/v2/dataproxy/converter"
	"github.com/flyteorg/flyte/v2/flytestdlib/storage"
	"github.com/flyteorg/flyte/v2/gen/go/flyteidl2/task"
	"github.com/flyteorg/flyte/v2/gen/go/flyteidl2/workflow"
	"github.com/flyteorg/flyte/v2/gen/go/flyteidl2/workflow/workflowconnect"
)

// TranslatorService implements the TranslatorServiceHandler interface for translating
// between Flyte literals and JSON representations. It is served from the dataproxy
// binary so that translation requests do not transit the control plane.
type TranslatorService struct {
	workflowconnect.UnimplementedTranslatorServiceHandler

	dataStore *storage.DataStore
	runClient workflowconnect.RunServiceClient
}

func NewTranslatorService(dataStore *storage.DataStore, runClient workflowconnect.RunServiceClient) *TranslatorService {
	return &TranslatorService{
		dataStore: dataStore,
		runClient: runClient,
	}
}

var _ workflowconnect.TranslatorServiceHandler = (*TranslatorService)(nil)

func (s *TranslatorService) LiteralsToLaunchFormJson(
	ctx context.Context,
	req *connect.Request[workflow.LiteralsToLaunchFormJsonRequest],
) (*connect.Response[workflow.LiteralsToLaunchFormJsonResponse], error) {
	literals := req.Msg.GetLiterals()
	if req.Msg.GetLiteralsUri() != "" {
		var err error
		literals, err = s.readOffloadedLiterals(ctx, req.Msg)
		if err != nil {
			return nil, err
		}
	}
	schema, err := converter.LiteralsToLaunchFormJson(ctx, literals, req.Msg.GetVariables())
	if err != nil {
		return nil, err
	}
	return connect.NewResponse(&workflow.LiteralsToLaunchFormJsonResponse{Json: schema}), nil
}

// readOffloadedLiterals reads action literals from the object store location named by
// literals_uri. The URI must match one of the action's data URIs reported by RunService,
// which both authorizes the read (RunService checks access to the action) and prevents
// arbitrary storage paths from being supplied.
func (s *TranslatorService) readOffloadedLiterals(
	ctx context.Context,
	req *workflow.LiteralsToLaunchFormJsonRequest,
) ([]*task.NamedLiteral, error) {
	actionID := req.GetActionId()
	if actionID == nil {
		return nil, connect.NewError(connect.CodeInvalidArgument, fmt.Errorf("action_id is required when literals_uri is set"))
	}

	urisResp, err := s.runClient.GetActionDataURIs(ctx, connect.NewRequest(&workflow.GetActionDataURIsRequest{
		ActionId: actionID,
	}))
	if err != nil {
		return nil, err
	}

	uri := req.GetLiteralsUri()
	if uri != urisResp.Msg.GetInputsUri() && uri != urisResp.Msg.GetOutputsUri() {
		return nil, connect.NewError(connect.CodeInvalidArgument,
			fmt.Errorf("literals_uri does not match any data URI of action %s", actionID.GetName()))
	}

	// Both inputs.pb and outputs.pb deserialize as task.Inputs (a NamedLiteral list).
	var inputsOrOutputs task.Inputs
	if err := s.dataStore.ReadProtobuf(ctx, storage.DataReference(uri), &inputsOrOutputs); err != nil {
		return nil, connect.NewError(connect.CodeInternal, fmt.Errorf("failed to read literals from %s: %w", uri, err))
	}
	return inputsOrOutputs.GetLiterals(), nil
}

func (s *TranslatorService) LaunchFormJsonToLiterals(
	ctx context.Context,
	req *connect.Request[workflow.LaunchFormJsonToLiteralsRequest],
) (*connect.Response[workflow.LaunchFormJsonToLiteralsResponse], error) {
	literals, err := converter.LaunchFormJsonToLiterals(ctx, req.Msg.GetJson())
	if err != nil {
		return nil, err
	}
	return connect.NewResponse(&workflow.LaunchFormJsonToLiteralsResponse{Literals: literals}), nil
}

func (s *TranslatorService) TaskSpecToLaunchFormJson(
	ctx context.Context,
	req *connect.Request[workflow.TaskSpecToLaunchFormJsonRequest],
) (*connect.Response[workflow.TaskSpecToLaunchFormJsonResponse], error) {
	schema, err := converter.TaskSpecToLaunchFormJson(ctx, req.Msg.GetTaskSpec())
	if err != nil {
		return nil, err
	}
	return connect.NewResponse(&workflow.TaskSpecToLaunchFormJsonResponse{Json: schema}), nil
}

func (s *TranslatorService) JsonValuesToLiterals(
	ctx context.Context,
	req *connect.Request[workflow.JsonValuesToLiteralsRequest],
) (*connect.Response[workflow.JsonValuesToLiteralsResponse], error) {
	literals, err := converter.JSONValuesToLiterals(ctx, req.Msg.GetVariables(), req.Msg.GetValues())
	if err != nil {
		return nil, err
	}
	return connect.NewResponse(&workflow.JsonValuesToLiteralsResponse{Literals: literals}), nil
}
