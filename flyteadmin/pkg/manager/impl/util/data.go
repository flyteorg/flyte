package util

import (
	"context"

	"github.com/golang/protobuf/proto"

	runtimeInterfaces "github.com/flyteorg/flyte/flyteadmin/pkg/runtime/interfaces"
	"github.com/flyteorg/flyte/flyteidl/gen/pb-go/flyteidl/admin"
	"github.com/flyteorg/flyte/flyteidl/gen/pb-go/flyteidl/core"
	"github.com/flyteorg/flyte/flytestdlib/logger"
	"github.com/flyteorg/flyte/flytestdlib/storage"
)

const (
	OutputsFile = "outputs.pb"
	DeckFile    = "deck.html"
)

// GetInputs returns an inputs URL blob and if config settings permit, inline inputs data for an execution.
func GetInputs(ctx context.Context,
	storageClient *storage.DataStore,
	inputURI string,
) *core.LiteralMap {
	var fullInputs core.LiteralMap

	if len(inputURI) == 0 {
		return &fullInputs
	}

	err := storageClient.ReadProtobuf(ctx, storage.DataReference(inputURI), &fullInputs)
	if err != nil {
		// If we fail to read the protobuf from the remote store, we shouldn't fail the request altogether.
		// Instead we return the signed URL blob so that the client can use that to fetch the input data.
		logger.Warningf(ctx, "Failed to read inputs from URI [%s] with err: %v", inputURI, err)
	}
	return &fullInputs
}

// ExecutionClosure defines common methods in NodeExecutionClosure and TaskExecutionClosure used to return output data.
type ExecutionClosure interface {
	GetOutputUri() string //nolint
	GetOutputData() *core.LiteralMap
}

// Wrapper around an admin.ExecutionClosure object which conforms to the output interface
// used by admin.NodeExecutionClosure and admin.TaskExecutionClosure
// Due to historical reasons, the workflow execution closure message is slightly different.
type workflowExecutionClosure struct {
	*admin.ExecutionClosure
}

func (c workflowExecutionClosure) GetOutputUri() string { //nolint
	var outputURI string
	if c.ExecutionClosure != nil && c.ExecutionClosure.GetOutputs() != nil {
		outputURI = c.ExecutionClosure.GetOutputs().GetUri()
	}
	return outputURI
}

func (c workflowExecutionClosure) GetOutputData() *core.LiteralMap {
	if c.ExecutionClosure.GetOutputs() != nil && c.ExecutionClosure.GetOutputs().GetValues() != nil {
		return c.ExecutionClosure.GetOutputs().GetValues()
	}
	return c.ExecutionClosure.GetOutputData()
}

// ToExecutionClosureInterface converts a workflow execution closure to an implementation of the ExecutionClosure
// interface for use in producing execution output data.
func ToExecutionClosureInterface(closure *admin.ExecutionClosure) ExecutionClosure {
	return &workflowExecutionClosure{
		ExecutionClosure: closure,
	}
}

// GetOutputs returns an outputs URL blob and if config settings permit, inline outputs data for an execution.
func GetOutputs(ctx context.Context,
	remoteDataConfig *runtimeInterfaces.RemoteDataConfig,
	storageClient *storage.DataStore,
	closure ExecutionClosure,
) *core.LiteralMap {
	fullOutputs := &core.LiteralMap{}
	if closure == nil {
		return fullOutputs
	}

	if closure.GetOutputData() != nil {
		if int64(proto.Size(closure.GetOutputData())) < remoteDataConfig.MaxSizeInBytes {
			fullOutputs = closure.GetOutputData()
		} else {
			logger.Debugf(ctx, "execution closure contains output data that exceeds max data size for responses")
		}
	} else if len(closure.GetOutputUri()) > 0 {
		err := storageClient.ReadProtobuf(ctx, storage.DataReference(closure.GetOutputUri()), fullOutputs)
		if err != nil {
			// If we fail to read the protobuf from the remote store, we shouldn't fail the request altogether.
			// Instead we return the signed URL blob so that the client can use that to fetch the output data.
			logger.Warningf(ctx, "Failed to read outputs from URI [%s] with err: %v", closure.GetOutputUri(), err)
		}
	}

	return fullOutputs
}
