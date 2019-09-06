package common

import (
	"context"

	"github.com/lyft/flytepropeller/pkg/controller/nodes/errors"
	"github.com/lyft/flytestdlib/logger"
	"github.com/lyft/flytestdlib/storage"

	"github.com/lyft/flyteidl/gen/pb-go/flyteidl/core"
	"github.com/lyft/flytepropeller/pkg/apis/flyteworkflow/v1alpha1"
	"github.com/lyft/flytepropeller/pkg/controller/nodes/handler"
)

func CreateAliasMap(aliases []v1alpha1.Alias) map[string]string {
	aliasToVarMap := make(map[string]string, len(aliases))
	for _, alias := range aliases {
		aliasToVarMap[alias.GetAlias()] = alias.GetVar()
	}
	return aliasToVarMap
}

// A simple output resolver that expects an outputs.pb at the data directory of the node.
type SimpleOutputsResolver struct {
	store storage.ProtobufStore
}

func (r SimpleOutputsResolver) ExtractOutput(ctx context.Context, w v1alpha1.ExecutableWorkflow, n v1alpha1.ExecutableNode,
	bindToVar handler.VarName) (values *core.Literal, err error) {
	d := &handler.Data{}
	nodeStatus := w.GetNodeExecutionStatus(n.GetID())
	outputsFileRef := v1alpha1.GetOutputsFile(nodeStatus.GetDataDir())
	if err := r.store.ReadProtobuf(ctx, outputsFileRef, d); err != nil {
		return nil, errors.Wrapf(errors.CausedByError, n.GetID(), err, "Failed to GetPrevious data from dataDir [%v]", nodeStatus.GetDataDir())
	}

	if d.Literals == nil {
		return nil, errors.Errorf(errors.OutputsNotFoundError, n.GetID(),
			"Outputs not found at [%v]", outputsFileRef)
	}

	aliasMap := CreateAliasMap(n.GetOutputAlias())
	if variable, ok := aliasMap[bindToVar]; ok {
		logger.Debugf(ctx, "Mapping [%v].[%v] -> [%v].[%v]", n.GetID(), variable, n.GetID(), bindToVar)
		bindToVar = variable
	}

	l, ok := d.Literals[bindToVar]
	if !ok {
		return nil, errors.Errorf(errors.OutputsNotFoundError, n.GetID(),
			"Failed to find [%v].[%v]", n.GetID(), bindToVar)
	}

	return l, nil
}

// Creates a simple output resolver that expects an outputs.pb at the data directory of the node.
func NewSimpleOutputsResolver(store storage.ProtobufStore) SimpleOutputsResolver {
	return SimpleOutputsResolver{
		store: store,
	}
}
