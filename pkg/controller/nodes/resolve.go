package nodes

import (
	"context"

	"github.com/lyft/flytepropeller/pkg/controller/nodes/common"

	"github.com/lyft/flyteidl/gen/pb-go/flyteidl/core"
	"github.com/lyft/flytepropeller/pkg/apis/flyteworkflow/v1alpha1"
	"github.com/lyft/flytepropeller/pkg/controller/nodes/errors"
	"github.com/lyft/flytepropeller/pkg/controller/nodes/handler"
	"github.com/lyft/flytestdlib/storage"
)

func ResolveBindingData(ctx context.Context, h HandlerFactory, w v1alpha1.ExecutableWorkflow, bindingData *core.BindingData, store storage.ProtobufStore) (*core.Literal, error) {
	literal := &core.Literal{}
	if bindingData == nil {
		return nil, nil
	}
	switch bindingData.GetValue().(type) {
	case *core.BindingData_Collection:
		literalCollection := make([]*core.Literal, 0, len(bindingData.GetCollection().GetBindings()))
		for _, b := range bindingData.GetCollection().GetBindings() {
			l, err := ResolveBindingData(ctx, h, w, b, store)
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
			l, err := ResolveBindingData(ctx, h, w, v, store)
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

		nodeHandler, err := h.GetHandler(n.GetKind())
		if err != nil {
			return nil, errors.Wrapf(errors.CausedByError, n.GetID(), err, "Failed to find handler for node kind [%v]", n.GetKind())
		}

		resolver, casted := nodeHandler.(handler.OutputResolver)
		if !casted {
			// If the handler doesn't implement output resolver, use simple resolver which expects an outputs.pb at the
			// output location of the task.
			if store == nil {
				return nil, errors.Errorf(errors.IllegalStateError, w.GetID(), n.GetID(), "System error. Promise lookup without store.")
			}

			resolver = common.NewSimpleOutputsResolver(store)
		}

		return resolver.ExtractOutput(ctx, w, n, bindToVar)
	case *core.BindingData_Scalar:
		literal.Value = &core.Literal_Scalar{Scalar: bindingData.GetScalar()}
	}
	return literal, nil
}

func Resolve(ctx context.Context, h HandlerFactory, w v1alpha1.ExecutableWorkflow, nodeID v1alpha1.NodeID, bindings []*v1alpha1.Binding, store storage.ProtobufStore) (*handler.Data, error) {
	literalMap := make(map[string]*core.Literal, len(bindings))
	for _, binding := range bindings {
		l, err := ResolveBindingData(ctx, h, w, binding.GetBinding(), store)
		if err != nil {
			return nil, errors.Wrapf(errors.BindingResolutionError, nodeID, err, "Error binding Var [%v].[%v]", w.GetID(), binding.GetVar())
		}
		literalMap[binding.GetVar()] = l
	}
	return &core.LiteralMap{
		Literals: literalMap,
	}, nil
}
