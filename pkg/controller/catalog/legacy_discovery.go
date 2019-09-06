package catalog

import (
	"context"
	"encoding/base64"
	"fmt"
	"time"

	"github.com/golang/protobuf/proto"
	grpc_retry "github.com/grpc-ecosystem/go-grpc-middleware/retry"
	"github.com/lyft/flyteidl/gen/pb-go/flyteidl/core"
	"github.com/lyft/flyteidl/gen/pb-go/flyteidl/datacatalog"
	"github.com/lyft/flytestdlib/logger"
	"github.com/lyft/flytestdlib/pbhash"
	"github.com/lyft/flytestdlib/storage"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
)

const maxGrpcMsgSizeBytes = 41943040

// LegacyDiscovery encapsulates interactions with the Discovery service using a protobuf provided gRPC client.
type LegacyDiscovery struct {
	client datacatalog.ArtifactsClient
	store  storage.ProtobufStore
}

// Hash each value in the map and return it as the parameter value to be used to generate the Provenance.
func TransformToInputParameters(ctx context.Context, m *core.LiteralMap) ([]*datacatalog.Parameter, error) {
	var params = []*datacatalog.Parameter{}

	//  Note: The Discovery service will ensure that the output parameters are sorted so the hash is consistent.
	// If the values of the literalmap are also a map, pbhash ensures that maps are deterministically hashed as well
	for k, typedValue := range m.GetLiterals() {
		inputHash, err := pbhash.ComputeHashString(ctx, typedValue)
		if err != nil {
			return nil, err
		}
		params = append(params, &datacatalog.Parameter{
			Name:  k,
			Value: inputHash,
		})
	}

	return params, nil
}

func TransformToOutputParameters(ctx context.Context, m *core.LiteralMap) ([]*datacatalog.Parameter, error) {
	var params = []*datacatalog.Parameter{}
	for k, typedValue := range m.GetLiterals() {
		bytes, err := proto.Marshal(typedValue)

		if err != nil {
			return nil, err
		}
		params = append(params, &datacatalog.Parameter{
			Name:  k,
			Value: base64.StdEncoding.EncodeToString(bytes),
		})
	}
	return params, nil
}

func TransformFromParameters(m []*datacatalog.Parameter) (*core.LiteralMap, error) {
	paramsMap := make(map[string]*core.Literal)

	for _, p := range m {
		bytes, err := base64.StdEncoding.DecodeString(p.GetValue())
		if err != nil {
			return nil, err
		}
		literal := &core.Literal{}
		if err = proto.Unmarshal(bytes, literal); err != nil {
			return nil, err
		}
		paramsMap[p.Name] = literal
	}
	return &core.LiteralMap{
		Literals: paramsMap,
	}, nil
}

func (d *LegacyDiscovery) Get(ctx context.Context, task *core.TaskTemplate, inputPath storage.DataReference) (*core.LiteralMap, error) {
	inputs := &core.LiteralMap{}

	taskInterface := task.Interface
	// only download if there are inputs to the task
	if taskInterface != nil && taskInterface.Inputs != nil && len(taskInterface.Inputs.Variables) > 0 {
		if err := d.store.ReadProtobuf(ctx, inputPath, inputs); err != nil {
			return nil, err
		}
	}

	inputParams, err := TransformToInputParameters(ctx, inputs)
	if err != nil {
		return nil, err
	}

	artifactID := &datacatalog.ArtifactId{
		Name:    fmt.Sprintf("%s:%s:%s", task.Id.Project, task.Id.Domain, task.Id.Name),
		Version: task.Metadata.DiscoveryVersion,
		Inputs:  inputParams,
	}
	options := []grpc.CallOption{
		grpc.MaxCallRecvMsgSize(maxGrpcMsgSizeBytes),
		grpc.MaxCallSendMsgSize(maxGrpcMsgSizeBytes),
	}

	request := &datacatalog.GetRequest{
		Id: &datacatalog.GetRequest_ArtifactId{
			ArtifactId: artifactID,
		},
	}
	resp, err := d.client.Get(ctx, request, options...)

	logger.Infof(ctx, "Discovery Get response for artifact |%v|, resp: |%v|, error: %v", artifactID, resp, err)
	if err != nil {
		return nil, err
	}
	return TransformFromParameters(resp.Artifact.Outputs)
}

func GetDefaultGrpcOptions() []grpc_retry.CallOption {
	return []grpc_retry.CallOption{
		grpc_retry.WithBackoff(grpc_retry.BackoffLinear(100 * time.Millisecond)),
		grpc_retry.WithCodes(codes.DeadlineExceeded, codes.Unavailable, codes.Canceled),
		grpc_retry.WithMax(5),
	}
}
func (d *LegacyDiscovery) Put(ctx context.Context, task *core.TaskTemplate, execID *core.TaskExecutionIdentifier, inputPath storage.DataReference, outputPath storage.DataReference) error {
	inputs := &core.LiteralMap{}
	outputs := &core.LiteralMap{}

	taskInterface := task.Interface
	// only download if there are inputs to the task
	if taskInterface != nil && taskInterface.Inputs != nil && len(taskInterface.Inputs.Variables) > 0 {
		if err := d.store.ReadProtobuf(ctx, inputPath, inputs); err != nil {
			return err
		}
	}

	// only download if there are outputs to the task
	if taskInterface != nil && taskInterface.Outputs != nil && len(taskInterface.Outputs.Variables) > 0 {
		if err := d.store.ReadProtobuf(ctx, outputPath, outputs); err != nil {
			return err
		}
	}

	outputParams, err := TransformToOutputParameters(ctx, outputs)
	if err != nil {
		return err
	}

	inputParams, err := TransformToInputParameters(ctx, inputs)
	if err != nil {
		return err
	}

	artifactID := &datacatalog.ArtifactId{
		Name:    fmt.Sprintf("%s:%s:%s", task.Id.Project, task.Id.Domain, task.Id.Name),
		Version: task.Metadata.DiscoveryVersion,
		Inputs:  inputParams,
	}
	executionID := fmt.Sprintf("%s:%s:%s", execID.GetNodeExecutionId().GetExecutionId().GetProject(),
		execID.GetNodeExecutionId().GetExecutionId().GetDomain(), execID.GetNodeExecutionId().GetExecutionId().GetName())
	request := &datacatalog.CreateRequest{
		Ref:         artifactID,
		ReferenceId: executionID,
		Revision:    time.Now().Unix(),
		Outputs:     outputParams,
	}
	options := []grpc.CallOption{
		grpc.MaxCallRecvMsgSize(maxGrpcMsgSizeBytes),
		grpc.MaxCallSendMsgSize(maxGrpcMsgSizeBytes),
	}

	resp, err := d.client.Create(ctx, request, options...)
	logger.Infof(ctx, "Discovery Put response for artifact |%v|, resp: |%v|, err: %v", artifactID, resp, err)
	return err
}

func NewLegacyDiscovery(discoveryEndpoint string, store storage.ProtobufStore) *LegacyDiscovery {

	// No discovery endpoint passed. Skip client creation.
	if discoveryEndpoint == "" {
		return nil
	}

	opts := GetDefaultGrpcOptions()
	retryInterceptor := grpc.WithUnaryInterceptor(grpc_retry.UnaryClientInterceptor(opts...))
	conn, err := grpc.Dial(discoveryEndpoint, grpc.WithInsecure(), retryInterceptor)

	if err != nil {
		return nil
	}
	client := datacatalog.NewArtifactsClient(conn)

	return &LegacyDiscovery{
		client: client,
		store:  store,
	}
}
