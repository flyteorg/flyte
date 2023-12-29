package util

import (
	"context"

	"github.com/golang/protobuf/proto"

	"github.com/flyteorg/flyte/flyteadmin/pkg/common"
	dataInterfaces "github.com/flyteorg/flyte/flyteadmin/pkg/data/interfaces"
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

func shouldFetchData(config *runtimeInterfaces.RemoteDataConfig, urlBlob admin.UrlBlob) bool {
	return config.Scheme == common.Local || config.Scheme == common.None || config.MaxSizeInBytes == 0 ||
		urlBlob.Bytes < config.MaxSizeInBytes
}

func shouldFetchOutputData(config *runtimeInterfaces.RemoteDataConfig, urlBlob admin.UrlBlob, outputURI string) bool {
	return len(outputURI) > 0 && shouldFetchData(config, urlBlob)
}

// GetInputs returns an inputs URL blob and if config settings permit, inline inputs data for an execution.
func GetInputs(ctx context.Context, urlData dataInterfaces.RemoteURLInterface,
	remoteDataConfig *runtimeInterfaces.RemoteDataConfig, storageClient *storage.DataStore, inputURI string) (
	*core.InputData, *admin.UrlBlob, error) {
	var inputsURLBlob admin.UrlBlob
	var fullInputs core.InputData

	if len(inputURI) == 0 {
		return &fullInputs, &inputsURLBlob, nil
	}

	var err error
	if remoteDataConfig.SignedURL.Enabled {
		inputsURLBlob, err = urlData.Get(ctx, inputURI)
		if err != nil {
			return nil, nil, err
		}
	}

	if shouldFetchData(remoteDataConfig, inputsURLBlob) {
		oldInputs := &core.LiteralMap{}
		msgIndex, err := storageClient.ReadProtobufAny(ctx, storage.DataReference(inputURI), &fullInputs, oldInputs)
		if err != nil {
			// If we fail to read the protobuf from the remote store, we shouldn't fail the request altogether.
			// Instead we return the signed URL blob so that the client can use that to fetch the input data.
			logger.Warningf(ctx, "Failed to read inputs from URI [%s] with err: %v", inputURI, err)
		} else if msgIndex == 1 {
			fullInputs = core.InputData{
				Inputs: oldInputs,
			}
		}
	}

	return &fullInputs, &inputsURLBlob, nil
}

// ExecutionClosure defines common methods in NodeExecutionClosure and TaskExecutionClosure used to return output data.
type ExecutionClosure interface {
	GetOutputUri() string //nolint
	GetOutputData() *core.LiteralMap
	GetFullOutputs() *core.OutputData
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

func (c workflowExecutionClosure) GetFullOutputs() *core.OutputData {
	if c.ExecutionClosure.GetOutputs() != nil && c.ExecutionClosure.GetOutputs().GetValues() != nil {
		return &core.OutputData{
			Outputs: c.ExecutionClosure.GetOutputs().GetValues(),
		}
	} else if oldOutputs := c.ExecutionClosure.GetOutputData(); oldOutputs != nil {
		return &core.OutputData{
			Outputs: oldOutputs,
		}
	}

	return c.ExecutionClosure.GetFullOutputs()
}

// ToExecutionClosureInterface converts a workflow execution closure to an implementation of the ExecutionClosure
// interface for use in producing execution output data.
func ToExecutionClosureInterface(closure *admin.ExecutionClosure) ExecutionClosure {
	return &workflowExecutionClosure{
		ExecutionClosure: closure,
	}
}

// GetOutputs returns an outputs URL blob and if config settings permit, inline outputs data for an execution.
func GetOutputs(ctx context.Context, urlData dataInterfaces.RemoteURLInterface,
	remoteDataConfig *runtimeInterfaces.RemoteDataConfig, storageClient *storage.DataStore, closure ExecutionClosure) (
	*core.OutputData, *admin.UrlBlob, error) {
	var outputsURLBlob admin.UrlBlob
	var fullOutputs = &core.OutputData{}
	if closure == nil {
		return fullOutputs, &outputsURLBlob, nil
	}

	if len(closure.GetOutputUri()) > 0 && remoteDataConfig.SignedURL.Enabled {
		var err error
		outputsURLBlob, err = urlData.Get(ctx, closure.GetOutputUri())
		if err != nil {
			return nil, nil, err
		}
	}

	if closure.GetFullOutputs() != nil {
		if int64(proto.Size(closure.GetFullOutputs())) < remoteDataConfig.MaxSizeInBytes {
			fullOutputs = closure.GetFullOutputs()
		} else {
			logger.Debugf(ctx, "execution closure contains output data that exceeds max data size for responses")
		}
	} else if closure.GetOutputData() != nil {
		if int64(proto.Size(closure.GetOutputData())) < remoteDataConfig.MaxSizeInBytes {
			fullOutputs = &core.OutputData{
				Outputs: closure.GetOutputData(),
			}
		} else {
			logger.Debugf(ctx, "execution closure contains output data that exceeds max data size for responses")
		}
	} else if shouldFetchOutputData(remoteDataConfig, outputsURLBlob, closure.GetOutputUri()) {
		oldOutputs := &core.LiteralMap{}
		msgIndex, err := storageClient.ReadProtobufAny(ctx, storage.DataReference(closure.GetOutputUri()), fullOutputs, oldOutputs)
		if err != nil {
			// If we fail to read the protobuf from the remote store, we shouldn't fail the request altogether.
			// Instead we return the signed URL blob so that the client can use that to fetch the output data.
			logger.Warningf(ctx, "Failed to read outputs from URI [%s] with err: %v", closure.GetOutputUri(), err)
		} else if msgIndex == 1 {
			fullOutputs = &core.OutputData{
				Outputs: oldOutputs,
			}
		}
	}

	return fullOutputs, &outputsURLBlob, nil
}
