package nodes

import (
	"context"

	"github.com/lyft/flyteidl/gen/pb-go/flyteidl/core"

	"github.com/lyft/flytepropeller/pkg/apis/flyteworkflow/v1alpha1"
	"github.com/lyft/flytepropeller/pkg/controller/nodes/errors"
)

func ResolveBindingData(ctx context.Context, outputResolver OutputResolver, w v1alpha1.BaseWorkflowWithStatus, bindingData *core.BindingData) (*core.Literal, error) {
	literal := &core.Literal{}
	if bindingData == nil {
		return nil, nil
	}
	switch bindingData.GetValue().(type) {
	case *core.BindingData_Collection:
		literalCollection := make([]*core.Literal, 0, len(bindingData.GetCollection().GetBindings()))
		for _, b := range bindingData.GetCollection().GetBindings() {
			l, err := ResolveBindingData(ctx, outputResolver, w, b)
			if err != nil {
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
		literalMap := make(map[string]*core.Literal, len(bindingData.GetMap().GetBindings()))
		for k, v := range bindingData.GetMap().GetBindings() {
			l, err := ResolveBindingData(ctx, outputResolver, w, v)
			if err != nil {
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
		upstreamNodeID := bindingData.GetPromise().GetNodeId()
		bindToVar := bindingData.GetPromise().GetVar()
		if w == nil {
			return nil, errors.Errorf(errors.IllegalStateError, upstreamNodeID,
				"Trying to resolve output from previous node, without providing the workflow for variable [%s]",
				bindToVar)
		}

		if upstreamNodeID == "" {
			return nil, errors.Errorf(errors.BadSpecificationError, "missing",
				"No nodeId (missing) specified for binding in Workflow.")
		}

		n, ok := w.GetNode(upstreamNodeID)
		if !ok {
			return nil, errors.Errorf(errors.IllegalStateError, w.GetID(), upstreamNodeID,
				"Undefined node in Workflow")
		}

		return outputResolver.ExtractOutput(ctx, w, n, bindToVar)
	case *core.BindingData_Scalar:
		literal.Value = &core.Literal_Scalar{Scalar: bindingData.GetScalar()}
	}

	return literal, nil
}

func Resolve(ctx context.Context, outputResolver OutputResolver, w v1alpha1.BaseWorkflowWithStatus, nodeID v1alpha1.NodeID, bindings []*v1alpha1.Binding) (*core.LiteralMap, error) {
	literalMap := make(map[string]*core.Literal, len(bindings))
	for _, binding := range bindings {
		varName := binding.GetVar()
		l, err := ResolveBindingData(ctx, outputResolver, w, binding.GetBinding())
		if err != nil {
			return nil, errors.Wrapf(errors.BindingResolutionError, nodeID, err, "Error binding Var [%v].[%v]", w.GetID(), binding.GetVar())
		}

		literalMap[varName] = l
	}
	return &core.LiteralMap{
		Literals: literalMap,
	}, nil
}
