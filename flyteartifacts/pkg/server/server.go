package server

import (
	"github.com/flyteorg/flyteidl/gen/pb-go/flyteidl/artifact"
	"google.golang.org/grpc"
	"net"

	"context"

	_ "net/http/pprof" // Required to serve application.
)

type ArtifactService struct {
	artifact.UnimplementedArtifactRegistryServer
}

func NewArtifactService() *ArtifactService {
	return &ArtifactService{}
}

func Serve(ctx context.Context, opts ...grpc.ServerOption) error {
	var serverOpts []grpc.ServerOption

	serverOpts = append(serverOpts, grpc.MaxRecvMsgSize(3000000))
	serverOpts = append(serverOpts, opts...)
	grpcServer := grpc.NewServer(serverOpts...)
	server := NewArtifactService()

	artifact.RegisterArtifactRegistryServer(grpcServer, server)

	lis, err := net.Listen("tcp", "localhost:50051")
	if err != nil {
		return err
	}
	err = grpcServer.Serve(lis)

	return err
}

func (a *ArtifactService) CreateArtifact(ctx context.Context, request *artifact.CreateArtifactRequest) (*artifact.CreateArtifactResponse, error) {
	return &artifact.CreateArtifactResponse{}, nil
}

func (a *ArtifactService) GetArtifact(ctx context.Context, request *artifact.GetArtifactRequest) (*artifact.GetArtifactResponse, error) {
	return &artifact.GetArtifactResponse{}, nil
}

func (a *ArtifactService) CreateTrigger(ctx context.Context, request *artifact.CreateTriggerRequest) (*artifact.CreateTriggerResponse, error) {
	return &artifact.CreateTriggerResponse{}, nil
}

func (a *ArtifactService) DeleteTrigger(ctx context.Context, request *artifact.DeleteTriggerRequest) (*artifact.DeleteTriggerResponse, error) {
	return &artifact.DeleteTriggerResponse{}, nil
}

func (a *ArtifactService) AddTag(ctx context.Context, request *artifact.AddTagRequest) (*artifact.AddTagResponse, error) {
	return &artifact.AddTagResponse{}, nil
}

func (a *ArtifactService) RegisterProducer(ctx context.Context, request *artifact.RegisterProducerRequest) (*artifact.RegisterResponse, error) {
	return &artifact.RegisterResponse{}, nil
}

func (a *ArtifactService) RegisterConsumer(ctx context.Context, request *artifact.RegisterConsumerRequest) (*artifact.RegisterResponse, error) {
	return &artifact.RegisterResponse{}, nil
}
