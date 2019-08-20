package types

import (
	"context"
	"fmt"

	"github.com/lyft/flyteidl/gen/pb-go/flyteidl/core"
	"github.com/lyft/flytestdlib/storage"
)

// Provides a default implementation for ResolveOutputs method by reading 'outputs.pb' from task directory into a LiteralMap.
type OutputsResolver struct {
	store storage.ComposedProtobufStore
}

func (r OutputsResolver) ResolveOutputs(ctx context.Context, taskCtx TaskContext, outputVariables ...VarName) (
	values map[VarName]*core.Literal, err error) {

	d := &core.LiteralMap{}
	outputsFileRef := taskCtx.GetOutputsFile()
	if err := r.store.ReadProtobuf(ctx, outputsFileRef, d); err != nil {
		return nil, fmt.Errorf("failed to read data from dataDir [%v]. Error: %v", taskCtx.GetOutputsFile(), err)
	}

	if d == nil || d.Literals == nil {
		return nil, fmt.Errorf("outputs from Task [%v] not found at [%v]", taskCtx.GetTaskExecutionID().GetGeneratedName(),
			outputsFileRef)
	}

	values = make(map[VarName]*core.Literal, len(outputVariables))
	for _, varName := range outputVariables {
		l, ok := d.Literals[varName]
		if !ok {
			return nil, fmt.Errorf("failed to find [%v].[%v]", taskCtx.GetTaskExecutionID().GetGeneratedName(), varName)
		}

		values[varName] = l
	}

	return values, nil
}

// Creates a default outputs resolver that expects a LiteralMap to exist in the task's outputFile location.
func NewOutputsResolver(store storage.ComposedProtobufStore) OutputsResolver {
	return OutputsResolver{store: store}
}
