package dataproxy

import (
	"context"
	"fmt"
	"strings"

	"connectrpc.com/connect/v2"

	"github.com/flyteorg/flyte/v2/dataproxy/converter"
	"github.com/flyteorg/flyte/v2/flyteplugins/go/tasks/pluginmachinery/ioutils"
	"github.com/flyteorg/flyte/v2/flytestdlib/storage"
	"github.com/flyteorg/flyte/v2/gen/go/flyteidl2/common"
	"github.com/flyteorg/flyte/v2/gen/go/flyteidl2/task"
	"github.com/flyteorg/flyte/v2/gen/go/flyteidl2/trigger"
	"github.com/flyteorg/flyte/v2/gen/go/flyteidl2/trigger/triggerconnect"
	"github.com/flyteorg/flyte/v2/gen/go/flyteidl2/workflow"
	"github.com/flyteorg/flyte/v2/gen/go/flyteidl2/workflow/workflowconnect"
)

// TranslatorService implements the TranslatorServiceHandler interface for translating
// between Flyte literals and JSON representations. It is served from the dataproxy
// binary so that translation requests do not transit the control plane.
type TranslatorService struct {
	workflowconnect.UnimplementedTranslatorServiceHandler

	dataStore     *storage.DataStore
	runClient     workflowconnect.RunServiceClient
	triggerClient triggerconnect.TriggerServiceClient
}

func NewTranslatorService(dataStore *storage.DataStore, runClient workflowconnect.RunServiceClient, triggerClient triggerconnect.TriggerServiceClient) *TranslatorService {
	return &TranslatorService{
		dataStore:     dataStore,
		runClient:     runClient,
		triggerClient: triggerClient,
	}
}

var _ workflowconnect.TranslatorServiceHandler = (*TranslatorService)(nil)

func (s *TranslatorService) LiteralsToLaunchFormJson(
	ctx context.Context,
	req *workflow.LiteralsToLaunchFormJsonRequest,
) (*workflow.LiteralsToLaunchFormJsonResponse, error) {
	literals := req.GetLiterals()
	switch {
	case req.GetTriggerId() != nil:
		var err error
		literals, err = s.readTriggerLiterals(ctx, req.GetTriggerId())
		if err != nil {
			return nil, err
		}
	case req.GetLiteralsUri() != "":
		var err error
		literals, err = s.readOffloadedLiterals(ctx, req)
		if err != nil {
			return nil, err
		}
	}
	schema, err := converter.LiteralsToLaunchFormJson(ctx, literals, req.GetVariables())
	if err != nil {
		return nil, err
	}
	return &workflow.LiteralsToLaunchFormJsonResponse{Json: schema}, nil
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
		return nil, connect.Errorf(connect.CodeInvalidArgument, "action_id is required when literals_uri is set")
	}

	urisResp, err := s.runClient.GetActionDataURIs(ctx, &workflow.GetActionDataURIsRequest{
		ActionId: actionID,
	})
	if err != nil {
		return nil, err
	}

	uri := req.GetLiteralsUri()
	if uri != urisResp.GetInputsUri() && uri != urisResp.GetOutputsUri() {
		return nil, connect.Errorf(connect.CodeInvalidArgument,
			"literals_uri does not match any data URI of action %s", actionID.GetName())
	}

	// Both inputs.pb and outputs.pb deserialize as task.Inputs (a NamedLiteral list).
	var inputsOrOutputs task.Inputs
	if err := s.dataStore.ReadProtobuf(ctx, storage.DataReference(uri), &inputsOrOutputs); err != nil {
		return nil, connect.NewError(connect.CodeInternal, fmt.Errorf("failed to read literals from %s: %w", uri, err).Error())
	}
	return inputsOrOutputs.GetLiterals(), nil
}

// readTriggerLiterals reads a trigger's offloaded inputs from storage. The offloaded
// input URI is taken from the trigger's own spec (looked up via TriggerService) rather
// than supplied by the client, so there is no client-provided URI to validate.
func (s *TranslatorService) readTriggerLiterals(
	ctx context.Context,
	triggerID *common.TriggerIdentifier,
) ([]*task.NamedLiteral, error) {
	resp, err := s.triggerClient.GetTriggerRevisionDetails(ctx, &trigger.GetTriggerRevisionDetailsRequest{
		Id: triggerID,
	})
	if err != nil {
		return nil, err
	}

	uri := resp.GetTrigger().GetSpec().GetOffloadedInputData().GetUri()
	if uri == "" {
		return nil, connect.Errorf(connect.CodeInvalidArgument,
			"trigger %s has no offloaded input data", triggerID.GetName().GetName())
	}

	// UploadInputs writes the literals to <dir>/inputs.pb but reports the *directory* as
	// OffloadedInputData.uri, so the suffix has to be appended before reading. Every other
	// reader does the same: the executor re-appends it via ioutils.NewInputFilePaths, and
	// CreateRun records inputs_uri as inputPrefix + "/inputs.pb". Tolerate a URI that already
	// names the file so a full path keeps working.
	trimmed := strings.TrimRight(uri, "/")
	inputRef := storage.DataReference(trimmed + "/" + ioutils.InputsSuffix)
	if strings.HasSuffix(trimmed, "/"+ioutils.InputsSuffix) {
		inputRef = storage.DataReference(trimmed)
	}

	var inputs task.Inputs
	if err := s.dataStore.ReadProtobuf(ctx, inputRef, &inputs); err != nil {
		return nil, connect.NewError(connect.CodeInternal, fmt.Errorf("failed to read trigger literals from %s: %w", inputRef, err).Error())
	}
	return inputs.GetLiterals(), nil
}

func (s *TranslatorService) LaunchFormJsonToLiterals(
	ctx context.Context,
	req *workflow.LaunchFormJsonToLiteralsRequest,
) (*workflow.LaunchFormJsonToLiteralsResponse, error) {
	literals, err := converter.LaunchFormJsonToLiterals(ctx, req.GetJson())
	if err != nil {
		return nil, err
	}
	return &workflow.LaunchFormJsonToLiteralsResponse{Literals: literals}, nil
}

func (s *TranslatorService) TaskSpecToLaunchFormJson(
	ctx context.Context,
	req *workflow.TaskSpecToLaunchFormJsonRequest,
) (*workflow.TaskSpecToLaunchFormJsonResponse, error) {
	schema, err := converter.TaskSpecToLaunchFormJson(ctx, req.GetTaskSpec())
	if err != nil {
		return nil, err
	}
	return &workflow.TaskSpecToLaunchFormJsonResponse{Json: schema}, nil
}

func (s *TranslatorService) JsonValuesToLiterals(
	ctx context.Context,
	req *workflow.JsonValuesToLiteralsRequest,
) (*workflow.JsonValuesToLiteralsResponse, error) {
	literals, err := converter.JSONValuesToLiterals(ctx, req.GetVariables(), req.GetValues())
	if err != nil {
		return nil, err
	}
	return &workflow.JsonValuesToLiteralsResponse{Literals: literals}, nil
}
