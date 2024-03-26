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

type Agent struct {
	// IsSync indicates whether this agent is a sync agent. Sync agents are expected to return their
	// results synchronously when called by propeller. Given that sync agents can affect the performance
	// of the system, it's important to enforce strict timeout policies.
	// An Async agent, on the other hand, is required to be able to identify jobs by an
	// identifier and query for job statuses as jobs progress.
	IsSync bool
	// AgentDeployment is the agent deployment where this agent is running.
	AgentDeployment *Deployment
}

// ClientSet contains the clients exposed to communicate with various agent services.
type ClientSet struct {
	asyncAgentClients    map[string]service.AsyncAgentServiceClient    // map[endpoint] => AsyncAgentServiceClient
	syncAgentClients     map[string]service.SyncAgentServiceClient     // map[endpoint] => SyncAgentServiceClient
	agentMetadataClients map[string]service.AgentMetadataServiceClient // map[endpoint] => AgentMetadataServiceClient
}

func getGrpcConnection(ctx context.Context, agent *Deployment) (*grpc.ClientConn, error) {
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

func getFinalTimeout(operation string, agent *Deployment) config.Duration {
	if t, exists := agent.Timeouts[operation]; exists {
		return t
	}

	return agent.DefaultTimeout
}

func getFinalContext(ctx context.Context, operation string, agent *Deployment) (context.Context, context.CancelFunc) {
	timeout := getFinalTimeout(operation, agent).Duration
	if timeout == 0 {
		return ctx, func() {}
	}

	return context.WithTimeout(ctx, timeout)
}

func initializeAgentRegistry(cs *ClientSet) (Registry, error) {
	logger.Infof(context.Background(), "Initializing agent registry")
	agentRegistry := make(Registry)
	cfg := GetConfig()
	var agentDeployments []*Deployment

	// Ensure that the old configuration is backward compatible
	for taskType, agentDeploymentID := range cfg.AgentForTaskTypes {
		agent := Agent{AgentDeployment: cfg.AgentDeployments[agentDeploymentID], IsSync: false}
		agentRegistry[taskType] = map[int32]*Agent{defaultTaskTypeVersion: &agent}
	}

	if len(cfg.DefaultAgent.Endpoint) != 0 {
		agentDeployments = append(agentDeployments, &cfg.DefaultAgent)
	}
	agentDeployments = append(agentDeployments, maps.Values(cfg.AgentDeployments)...)
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

		for _, agent := range res.GetAgents() {
			deprecatedSupportedTaskTypes := agent.SupportedTaskTypes
			for _, supportedTaskType := range deprecatedSupportedTaskTypes {
				agent := &Agent{AgentDeployment: agentDeployment, IsSync: agent.IsSync}
				agentRegistry[supportedTaskType] = map[int32]*Agent{defaultTaskTypeVersion: agent}
			}

			supportedTaskCategories := agent.SupportedTaskCategories
			for _, supportedCategory := range supportedTaskCategories {
				agent := &Agent{AgentDeployment: agentDeployment, IsSync: agent.IsSync}
				agentRegistry[supportedCategory.GetName()] = map[int32]*Agent{supportedCategory.GetVersion(): agent}
			}
			logger.Infof(context.Background(), "[%v] is a sync agent: [%v]", agent.Name, agent.IsSync)
			logger.Infof(context.Background(), "[%v] supports task category: [%v]", agent.Name, supportedTaskCategories)
		}
	}

	return agentRegistry, nil
}

func initializeClients(ctx context.Context) (*ClientSet, error) {
	logger.Infof(ctx, "Initializing agent clients")

	asyncAgentClients := make(map[string]service.AsyncAgentServiceClient)
	syncAgentClients := make(map[string]service.SyncAgentServiceClient)
	agentMetadataClients := make(map[string]service.AgentMetadataServiceClient)

	var agentDeployments []*Deployment
	cfg := GetConfig()

	if len(cfg.DefaultAgent.Endpoint) != 0 {
		agentDeployments = append(agentDeployments, &cfg.DefaultAgent)
	}
	agentDeployments = append(agentDeployments, maps.Values(cfg.AgentDeployments)...)
	for _, agentService := range agentDeployments {
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
