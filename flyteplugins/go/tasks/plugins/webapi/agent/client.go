package agent

import (
	"context"

	"google.golang.org/grpc"

	"github.com/flyteorg/flyte/flyteidl/gen/pb-go/flyteidl/service"
)

type GetAgentClientFunc func(ctx context.Context, agent *Agent, connectionCache map[*Agent]*grpc.ClientConn) (service.AsyncAgentServiceClient, error)
type GetAgentMetadataClientFunc func(ctx context.Context, agent *Agent, connCache map[*Agent]*grpc.ClientConn) (service.AgentMetadataServiceClient, error)

// Clientset contains the clients exposed to communicate with various agent services.
type ClientFuncSet struct {
	getAgentClient         GetAgentClientFunc
	getAgentMetadataClient GetAgentMetadataClientFunc
}

func getAgentClientFunc(ctx context.Context, agent *Agent, connectionCache map[*Agent]*grpc.ClientConn) (service.AsyncAgentServiceClient, error) {
	conn, err := getGrpcConnection(ctx, agent, connectionCache)
	if err != nil {
		return nil, err
	}

	return service.NewAsyncAgentServiceClient(conn), nil
}

func getAgentMetadataClientFunc(ctx context.Context, agent *Agent, connectionCache map[*Agent]*grpc.ClientConn) (service.AgentMetadataServiceClient, error) {
	conn, err := getGrpcConnection(ctx, agent, connectionCache)
	if err != nil {
		return nil, err
	}

	return service.NewAgentMetadataServiceClient(conn), nil
}

func initializeClientFunc() *ClientFuncSet {
	return &ClientFuncSet{
		getAgentClient:         getAgentClientFunc,
		getAgentMetadataClient: getAgentMetadataClientFunc,
	}
}
