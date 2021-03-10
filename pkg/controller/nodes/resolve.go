package nodes

import (
	"context"

	"github.com/flyteorg/flyteidl/gen/pb-go/flyteidl/core"
	"github.com/flyteorg/flytepropeller/pkg/apis/flyteworkflow/v1alpha1"
	"github.com/flyteorg/flytepropeller/pkg/controller/executors"
	"github.com/flyteorg/flytepropeller/pkg/controller/nodes/errors"
	"github.com/flyteorg/flytestdlib/logger"
)

func ResolveBindingData(ctx context.Context, outputResolver OutputResolver, nl executors.NodeLookup, bindingData *core.BindingData) (*core.Literal, error) {
	logger.Debugf(ctx, "Resolving binding data")

	literal := &core.Literal{}
	if bindingData == nil {
		logger.Debugf(ctx, "bindingData is nil")
		return nil, nil
	}
	switch bindingData.GetValue().(type) {
	case *core.BindingData_Collection:

		logger.Debugf(ctx, "bindingData.GetValue() [%v] is of type Collection", bindingData.GetValue())
		literalCollection := make([]*core.Literal, 0, len(bindingData.GetCollection().GetBindings()))
		for _, b := range bindingData.GetCollection().GetBindings() {
			l, err := ResolveBindingData(ctx, outputResolver, nl, b)
			if err != nil {
				logger.Debugf(ctx, "Failed to resolve binding data. Error: [%v]", err)
				return nil, err
			}

			literalCollection = append(literalCollection, l)
		}

		literal.Value = &core.Literal_Collection{
			Collection: &core.LiteralCollection{
				Literals: literalCollection,
			},
		}
	case *core.BindingData_Map:
		logger.Debugf(ctx, "bindingData.GetValue() [%v] is of type Map", bindingData.GetValue())
		literalMap := make(map[string]*core.Literal, len(bindingData.GetMap().GetBindings()))
		for k, v := range bindingData.GetMap().GetBindings() {
			l, err := ResolveBindingData(ctx, outputResolver, nl, v)
			if err != nil {
				logger.Debugf(ctx, "Failed to resolve binding data. Error: [%v]", err)
				return nil, err
			}

			literalMap[k] = l
		}

		literal.Value = &core.Literal_Map{
			Map: &core.LiteralMap{
				Literals: literalMap,
			},
		}
	case *core.BindingData_Promise:
		logger.Debugf(ctx, "bindingData.GetValue() [%v] is of type Promise", bindingData.GetValue())

		upstreamNodeID := bindingData.GetPromise().GetNodeId()
		bindToVar := bindingData.GetPromise().GetVar()

		if nl == nil {
			return nil, errors.Errorf(errors.IllegalStateError, upstreamNodeID,
				"Trying to resolve output from previous node, without providing the workflow for variable [%s]",
				bindToVar)
		}

		if upstreamNodeID == "" {
			return nil, errors.Errorf(errors.BadSpecificationError, "missing",
				"No nodeId (missing) specified for binding in Workflow.")
		}

		n, ok := nl.GetNode(upstreamNodeID)
		if !ok {
			return nil, errors.Errorf(errors.IllegalStateError, "id", upstreamNodeID,
				"Undefined node in Workflow")
		}

		return outputResolver.ExtractOutput(ctx, nl, n, bindToVar)
	case *core.BindingData_Scalar:
		logger.Debugf(ctx, "bindingData.GetValue() [%v] is of type Scalar", bindingData.GetValue())
		literal.Value = &core.Literal_Scalar{Scalar: bindingData.GetScalar()}
	}
	return literal, nil
}

func Resolve(ctx context.Context, outputResolver OutputResolver, nl executors.NodeLookup, nodeID v1alpha1.NodeID, bindings []*v1alpha1.Binding) (*core.LiteralMap, error) {
	logger.Debugf(ctx, "bindings: [%v]", bindings)
	literalMap := make(map[string]*core.Literal, len(bindings))
	for _, binding := range bindings {
		logger.Debugf(ctx, "Resolving binding: [%v]", binding)
		varName := binding.GetVar()
		l, err := ResolveBindingData(ctx, outputResolver, nl, binding.GetBinding())
		if err != nil {
			return nil, errors.Wrapf(errors.BindingResolutionError, nodeID, err, "Error binding Var [%v].[%v]", "wf", binding.GetVar())
		}

		literalMap[varName] = l
	}
	return &core.LiteralMap{
		Literals: literalMap,
	}, nil
}
