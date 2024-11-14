package util

import (
	"context"
	"fmt"
	"net/url"
	"strings"

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
	DeckFile = "deck.html"
)

type GetObjectRequest struct {
	Cluster  string
	Org      string
	Project  string
	Domain   string
	Protocol string
	Bucket   string
	Prefix   string
}

type GetObjectResponse struct {
	Contents []byte
}

type ObjectStore interface {
	GetObject(context.Context, GetObjectRequest) (GetObjectResponse, error)
}

func shouldFetchData(config *runtimeInterfaces.RemoteDataConfig, urlBlob admin.UrlBlob) bool {
	return config.Scheme == common.Local || config.Scheme == common.None || config.MaxSizeInBytes == 0 ||
		urlBlob.Bytes < config.MaxSizeInBytes
}

func shouldFetchOutputData(config *runtimeInterfaces.RemoteDataConfig, urlBlob admin.UrlBlob, outputURI string) bool {
	return len(outputURI) > 0 && shouldFetchData(config, urlBlob)
}

// GetInputs returns an inputs URL blob and if config settings permit, inline inputs data for an execution.
func GetInputs(ctx context.Context,
	urlData dataInterfaces.RemoteURLInterface,
	remoteDataConfig *runtimeInterfaces.RemoteDataConfig,
	storageClient *storage.DataStore,
	cluster, org, project, domain, inputURI string,
	objectStore ObjectStore,
) (*core.LiteralMap, *admin.UrlBlob, error) {
	var inputsURLBlob admin.UrlBlob
	var fullInputs core.LiteralMap

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
		if IsLocalURI(ctx, storageClient, inputURI) {
			err = storageClient.ReadProtobuf(ctx, storage.DataReference(inputURI), &fullInputs)
		} else {
			err = readFromDataPlane(ctx, objectStore, cluster, org, project, domain, inputURI, &fullInputs)
		}
	}
	return &fullInputs, &inputsURLBlob, err
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
	urlData dataInterfaces.RemoteURLInterface,
	remoteDataConfig *runtimeInterfaces.RemoteDataConfig,
	storageClient *storage.DataStore,
	closure ExecutionClosure,
	cluster, org, project, domain string,
	objectStore ObjectStore,
) (*core.LiteralMap, *admin.UrlBlob, error) {
	var outputsURLBlob admin.UrlBlob
	var fullOutputs = &core.LiteralMap{}
	if closure == nil {
		return fullOutputs, &outputsURLBlob, nil
	}

	var err error
	if len(closure.GetOutputUri()) > 0 && remoteDataConfig.SignedURL.Enabled {
		outputsURLBlob, err = urlData.Get(ctx, closure.GetOutputUri())
		if err != nil {
			return nil, nil, err
		}
	}

	if closure.GetOutputData() != nil {
		if int64(proto.Size(closure.GetOutputData())) < remoteDataConfig.MaxSizeInBytes {
			fullOutputs = closure.GetOutputData()
		} else {
			logger.Debugf(ctx, "execution closure contains output data that exceeds max data size for responses")
		}
	} else if shouldFetchOutputData(remoteDataConfig, outputsURLBlob, closure.GetOutputUri()) {
		if IsLocalURI(ctx, storageClient, closure.GetOutputUri()) {
			err = storageClient.ReadProtobuf(ctx, storage.DataReference(closure.GetOutputUri()), fullOutputs)
		} else {
			err = readFromDataPlane(ctx, objectStore, cluster, org, project, domain, closure.GetOutputUri(), fullOutputs)
		}
	}
	return fullOutputs, &outputsURLBlob, err
}

func IsLocalURI(ctx context.Context, store *storage.DataStore, uri string) bool {
	base := store.GetBaseContainerFQN(ctx)
	return strings.HasPrefix(uri, string(base))
}

func readFromDataPlane(ctx context.Context,
	objectStore ObjectStore,
	cluster, org, project, domain, reference string,
	msg proto.Message,
) error {
	if objectStore == nil {
		return fmt.Errorf("no object store implementation found")
	}

	refURL, err := url.Parse(reference)
	if err != nil {
		return fmt.Errorf("failed to parse data reference: %w", err)
	}

	out, err := objectStore.GetObject(ctx, GetObjectRequest{
		Cluster:  cluster,
		Org:      org,
		Protocol: refURL.Scheme,
		Bucket:   refURL.Host,
		Prefix:   refURL.Path,
		Project:  project,
		Domain:   domain,
	})
	if err != nil {
		return fmt.Errorf("failed to fetch object: %w", err)
	}

	err = proto.Unmarshal(out.Contents, msg)
	if err != nil {
		return fmt.Errorf("failed to unmarshal object contents: %w", err)
	}
	return nil
}
