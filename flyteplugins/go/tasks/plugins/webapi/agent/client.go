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

const defaultTaskTypeVersion = 0

// ClientSet contains the clients exposed to communicate with various agent services.
type ClientSet struct {
	asyncAgentClients    map[string]service.AsyncAgentServiceClient    // map[endpoint] => AsyncAgentServiceClient
	syncAgentClients     map[string]service.SyncAgentServiceClient     // map[endpoint] => SyncAgentServiceClient
	agentMetadataClients map[string]service.AgentMetadataServiceClient // map[endpoint] => AgentMetadataServiceClient
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

func initializeAgentRegistry(cs *ClientSet) (map[string]map[int32]*Agent, error) {
	agentRegistry := make(map[string]map[int32]*Agent)
	cfg := GetConfig()
	var agentServices []*Agent

	// Ensure that the old configuration is backward compatible
	for taskType, agentID := range cfg.AgentForTaskTypes {
		agentRegistry[taskType] = map[int32]*Agent{defaultTaskTypeVersion: cfg.Agents[agentID]}
	}

	if len(cfg.DefaultAgent.Endpoint) != 0 {
		agentServices = append(agentServices, &cfg.DefaultAgent)
	}
	agentServices = append(agentServices, maps.Values(cfg.Agents)...)
	for i := 0; i < len(agentServices); i++ {
		client := cs.agentMetadataClients[agentServices[i].Endpoint]

		finalCtx, cancel := getFinalContext(context.Background(), "ListAgents", agentServices[i])
		defer cancel()

		res, err := client.ListAgents(finalCtx, &admin.ListAgentsRequest{})
		if err != nil {
			grpcStatus, ok := status.FromError(err)
			if grpcStatus.Code() == codes.Unimplemented {
				// we should not panic here, as we want to continue to support old agent settings
				logger.Infof(context.Background(), "list agent method not implemented for agent: [%v]", agentServices[i])
				continue
			}

			if !ok {
				return nil, fmt.Errorf("failed to list agent: [%v] with a non-gRPC error: [%v]", agentServices[i], err)
			}

			return nil, fmt.Errorf("failed to list agent: [%v] with error: [%v]", agentServices[i], err)
		}

		agents := res.GetAgents()
		for j := 0; j < len(agents); j++ {
			supportedTaskTypes := agents[j].SupportedTaskTypes
			for _, supportedTaskType := range supportedTaskTypes {
				finalAgent := agentServices[i]
				finalAgent.IsSync = true
				agentRegistry[supportedTaskType.GetName()] = map[int32]*Agent{supportedTaskType.Version: finalAgent}
			}
			logger.Infof(context.Background(), "[%v] Agent is a sync agent: [%v]", agents[j].Name, agents[j].IsSync)
			logger.Infof(context.Background(), "[%v] Agent supports task types: [%v]", agents[j].Name, supportedTaskTypes)
		}
	}

	logger.Infof(context.Background(), "Agent registry initialized: [%v]", agentRegistry["mock_openai"][0])

	return agentRegistry, nil
}

func initializeClients(ctx context.Context) (*ClientSet, error) {
	asyncAgentClients := make(map[string]service.AsyncAgentServiceClient)
	syncAgentClients := make(map[string]service.SyncAgentServiceClient)
	agentMetadataClients := make(map[string]service.AgentMetadataServiceClient)

	var agentServices []*Agent
	cfg := GetConfig()

	if len(cfg.DefaultAgent.Endpoint) != 0 {
		agentServices = append(agentServices, &cfg.DefaultAgent)
	}
	agentServices = append(agentServices, maps.Values(cfg.Agents)...)
	for _, agentService := range agentServices {
		conn, err := getGrpcConnection(ctx, agentService)
		if err != nil {
			return nil, err
		}
		syncAgentClients[agentService.Endpoint] = service.NewSyncAgentServiceClient(conn)
		asyncAgentClients[agentService.Endpoint] = service.NewAsyncAgentServiceClient(conn)
		agentMetadataClients[agentService.Endpoint] = service.NewAgentMetadataServiceClient(conn)
	}

	return &ClientSet{
		syncAgentClients:     syncAgentClients,
		asyncAgentClients:    asyncAgentClients,
		agentMetadataClients: agentMetadataClients,
	}, nil
}
