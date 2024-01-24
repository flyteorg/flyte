package agent

import (
	"context"
	"crypto/x509"
	"fmt"

	"golang.org/x/exp/maps"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/credentials"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/grpclog"
	"google.golang.org/grpc/status"

	"github.com/flyteorg/flyte/flyteidl/gen/pb-go/flyteidl/admin"
	"github.com/flyteorg/flyte/flyteidl/gen/pb-go/flyteidl/service"
	"github.com/flyteorg/flyte/flytestdlib/config"
	"github.com/flyteorg/flyte/flytestdlib/logger"
)

// ClientSet contains the clients exposed to communicate with various agent services.
type ClientSet struct {
	agentClients         map[string]service.AsyncAgentServiceClient    // map[endpoint] => client
	agentMetadataClients map[string]service.AgentMetadataServiceClient // map[endpoint] => client
}

func getGrpcConnection(ctx context.Context, agent *Agent) (*grpc.ClientConn, error) {
	var opts []grpc.DialOption

	if agent.Insecure {
		opts = append(opts, grpc.WithTransportCredentials(insecure.NewCredentials()))
	} else {
		pool, err := x509.SystemCertPool()
		if err != nil {
			return nil, err
		}

		creds := credentials.NewClientTLSFromCert(pool, "")
		opts = append(opts, grpc.WithTransportCredentials(creds))
	}

	if len(agent.DefaultServiceConfig) != 0 {
		opts = append(opts, grpc.WithDefaultServiceConfig(agent.DefaultServiceConfig))
	}

	var err error
	conn, err := grpc.Dial(agent.Endpoint, opts...)
	if err != nil {
		return nil, err
	}
	defer func() {
		if err != nil {
			if cerr := conn.Close(); cerr != nil {
				grpclog.Infof("Failed to close conn to %s: %v", agent, cerr)
			}
			return
		}
		go func() {
			<-ctx.Done()
			if cerr := conn.Close(); cerr != nil {
				grpclog.Infof("Failed to close conn to %s: %v", agent, cerr)
			}
		}()
	}()

	return conn, nil
}

func getFinalTimeout(operation string, agent *Agent) config.Duration {
	if t, exists := agent.Timeouts[operation]; exists {
		return t
	}

	return agent.DefaultTimeout
}

func getFinalContext(ctx context.Context, operation string, agent *Agent) (context.Context, context.CancelFunc) {
	timeout := getFinalTimeout(operation, agent).Duration
	if timeout == 0 {
		return ctx, func() {}
	}

	return context.WithTimeout(ctx, timeout)
}

func initializeAgentRegistry(cs *ClientSet) (map[string]*Agent, error) {
	agentRegistry := make(map[string]*Agent)
	cfg := GetConfig()
	var agentDeployments []*Agent

	// Ensure that the old configuration is backward compatible
	for taskType, agentID := range cfg.AgentForTaskTypes {
		agentRegistry[taskType] = cfg.Agents[agentID]
	}

	if len(cfg.DefaultAgent.Endpoint) != 0 {
		agentDeployments = append(agentDeployments, &cfg.DefaultAgent)
	}
	agentDeployments = append(agentDeployments, maps.Values(cfg.Agents)...)
	for _, agentDeployment := range agentDeployments {
		client := cs.agentMetadataClients[agentDeployment.Endpoint]

		finalCtx, cancel := getFinalContext(context.Background(), "ListAgents", agentDeployment)
		defer cancel()

		res, err := client.ListAgents(finalCtx, &admin.ListAgentsRequest{})
		if err != nil {
			grpcStatus, ok := status.FromError(err)
			if grpcStatus.Code() == codes.Unimplemented {
				// we should not panic here, as we want to continue to support old agent settings
				logger.Infof(context.Background(), "list agent method not implemented for agent: [%v]", agentDeployment)
				continue
			}

			if !ok {
				return nil, fmt.Errorf("failed to list agent: [%v] with a non-gRPC error: [%v]", agentDeployment, err)
			}

			return nil, fmt.Errorf("failed to list agent: [%v] with error: [%v]", agentDeployment, err)
		}

		agents := res.GetAgents()
		for _, agent := range agents {
			supportedTaskTypes := agent.SupportedTaskTypes
			for _, supportedTaskType := range supportedTaskTypes {
				agentRegistry[supportedTaskType] = agentDeployment
			}
		}
	}

	return agentRegistry, nil
}

func initializeClients(ctx context.Context) (*ClientSet, error) {
	agentClients := make(map[string]service.AsyncAgentServiceClient)
	agentMetadataClients := make(map[string]service.AgentMetadataServiceClient)

	var agentDeployments []*Agent
	cfg := GetConfig()

	if len(cfg.DefaultAgent.Endpoint) != 0 {
		agentDeployments = append(agentDeployments, &cfg.DefaultAgent)
	}
	agentDeployments = append(agentDeployments, maps.Values(cfg.Agents)...)
	for _, agentDeployment := range agentDeployments {
		conn, err := getGrpcConnection(ctx, agentDeployment)
		if err != nil {
			return nil, err
		}
		agentClients[agentDeployment.Endpoint] = service.NewAsyncAgentServiceClient(conn)
		agentMetadataClients[agentDeployment.Endpoint] = service.NewAgentMetadataServiceClient(conn)
	}

	return &ClientSet{
		agentClients:         agentClients,
		agentMetadataClients: agentMetadataClients,
	}, nil
}
