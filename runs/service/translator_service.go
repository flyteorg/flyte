package service

import (
	"context"

	"connectrpc.com/connect"

	"github.com/flyteorg/flyte/v2/gen/go/flyteidl2/workflow"
	"github.com/flyteorg/flyte/v2/gen/go/flyteidl2/workflow/workflowconnect"
	"github.com/flyteorg/flyte/v2/runs/service/converter"
)

// TranslatorService implements the TranslatorServiceHandler interface for translating
// between Flyte literals and JSON representations.
type TranslatorService struct {
	workflowconnect.UnimplementedTranslatorServiceHandler
}

// NewTranslatorService creates a new TranslatorService instance.
func NewTranslatorService() *TranslatorService {
	return &TranslatorService{}
}

// Ensure we implement the interface
var _ workflowconnect.TranslatorServiceHandler = (*TranslatorService)(nil)

// LiteralsToLaunchFormJson converts a list of NamedLiterals to an RSJF-compliant JSON schema.
func (s *TranslatorService) LiteralsToLaunchFormJson(
	ctx context.Context,
	req *connect.Request[workflow.LiteralsToLaunchFormJsonRequest],
) (*connect.Response[workflow.LiteralsToLaunchFormJsonResponse], error) {
	schema, err := converter.LiteralsToLaunchFormJson(ctx, req.Msg.GetLiterals(), req.Msg.GetVariables())
	if err != nil {
		return nil, err
	}

	return connect.NewResponse(&workflow.LiteralsToLaunchFormJsonResponse{Json: schema}), nil
}

// LaunchFormJsonToLiterals converts an RSJF JSON schema to a list of NamedLiterals.
func (s *TranslatorService) LaunchFormJsonToLiterals(
	ctx context.Context,
	req *connect.Request[workflow.LaunchFormJsonToLiteralsRequest],
) (*connect.Response[workflow.LaunchFormJsonToLiteralsResponse], error) {
	literals, err := converter.LaunchFormJsonToLiterals(ctx, req.Msg.GetJson())
	if err != nil {
		return nil, err
	}

	return connect.NewResponse(&workflow.LaunchFormJsonToLiteralsResponse{
		Literals: literals,
	}), nil
}

// TaskSpecToLaunchFormJson converts a TaskSpec to an RSJF-compliant JSON schema.
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

// JsonValuesToLiterals converts raw JSON values and type definitions (VariableMap)
// into NamedLiterals.
func (s *TranslatorService) JsonValuesToLiterals(
	ctx context.Context,
	req *connect.Request[workflow.JsonValuesToLiteralsRequest],
) (*connect.Response[workflow.JsonValuesToLiteralsResponse], error) {
	literals, err := converter.JSONValuesToLiterals(ctx, req.Msg.GetVariables(), req.Msg.GetValues())
	if err != nil {
		return nil, err
	}

	return connect.NewResponse(&workflow.JsonValuesToLiteralsResponse{
		Literals: literals,
	}), nil
}
