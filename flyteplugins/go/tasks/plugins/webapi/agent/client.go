package agent

import (
	"context"
	"crypto/x509"
	"strings"

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

	go func() {
		<-ctx.Done()
		if cerr := conn.Close(); cerr != nil {
			grpclog.Infof("Failed to close conn to %s: %v", agent, cerr)
		}
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

func getAgentRegistry(ctx context.Context, cs *ClientSet) Registry {
	newAgentRegistry := make(Registry)
	cfg := GetConfig()
	var agentDeployments []*Deployment

	if len(cfg.DefaultAgent.Endpoint) != 0 {
		agentDeployments = append(agentDeployments, &cfg.DefaultAgent)
	}
	agentDeployments = append(agentDeployments, maps.Values(cfg.AgentDeployments)...)
	for _, agentDeployment := range agentDeployments {
		client, ok := cs.agentMetadataClients[agentDeployment.Endpoint]
		if !ok {
			logger.Warningf(ctx, "Agent client not found in the clientSet for the endpoint: %v", agentDeployment.Endpoint)
			continue
		}

		finalCtx, cancel := getFinalContext(ctx, "ListAgents", agentDeployment)
		defer cancel()

		res, err := client.ListAgents(finalCtx, &admin.ListAgentsRequest{})
		if err != nil {
			grpcStatus, ok := status.FromError(err)
			if grpcStatus.Code() == codes.Unimplemented {
				// we should not panic here, as we want to continue to support old agent settings
				logger.Warningf(finalCtx, "list agent method not implemented for agent: [%v]", agentDeployment.Endpoint)
				continue
			}

			if !ok {
				logger.Errorf(finalCtx, "failed to list agent: [%v] with a non-gRPC error: [%v]", agentDeployment.Endpoint, err)
				continue
			}

			logger.Errorf(finalCtx, "failed to list agent: [%v] with error: [%v]", agentDeployment.Endpoint, err)
			continue
		}

		agentSupportedTaskCategories := make(map[string]struct{})
		for _, agent := range res.GetAgents() {
			deprecatedSupportedTaskTypes := agent.GetSupportedTaskTypes()
			for _, supportedTaskType := range deprecatedSupportedTaskTypes {
				agent := &Agent{AgentDeployment: agentDeployment, IsSync: agent.GetIsSync()}
				newAgentRegistry[supportedTaskType] = map[int32]*Agent{defaultTaskTypeVersion: agent}
				agentSupportedTaskCategories[supportedTaskType] = struct{}{}
			}

			supportedTaskCategories := agent.GetSupportedTaskCategories()
			for _, supportedCategory := range supportedTaskCategories {
				agent := &Agent{AgentDeployment: agentDeployment, IsSync: agent.GetIsSync()}
				supportedCategoryName := supportedCategory.GetName()
				newAgentRegistry[supportedCategoryName] = map[int32]*Agent{supportedCategory.GetVersion(): agent}
				agentSupportedTaskCategories[supportedCategoryName] = struct{}{}
			}

		}
		logger.Infof(ctx, "AgentDeployment [%v] supports the following task types: [%v]", agentDeployment.Endpoint,
			strings.Join(maps.Keys(agentSupportedTaskCategories), ", "))
	}

	// Always replace the registry with the settings defined in the configuration
	for taskType, agentDeploymentID := range cfg.AgentForTaskTypes {
		if agentDeployment, ok := cfg.AgentDeployments[agentDeploymentID]; ok {
			agent := &Agent{AgentDeployment: agentDeployment, IsSync: false}
			newAgentRegistry[taskType] = map[int32]*Agent{defaultTaskTypeVersion: agent}
		}
	}

	// Ensure that the old configuration is backward compatible
	for _, taskType := range cfg.SupportedTaskTypes {
		if _, ok := newAgentRegistry[taskType]; !ok {
			agent := &Agent{AgentDeployment: &cfg.DefaultAgent, IsSync: false}
			newAgentRegistry[taskType] = map[int32]*Agent{defaultTaskTypeVersion: agent}
		}
	}

	logger.Infof(ctx, "AgentDeployments support the following task types: [%v]", strings.Join(maps.Keys(newAgentRegistry), ", "))
	return newAgentRegistry
}

func getAgentClientSets(ctx context.Context) *ClientSet {
	clientSet := &ClientSet{
		asyncAgentClients:    make(map[string]service.AsyncAgentServiceClient),
		syncAgentClients:     make(map[string]service.SyncAgentServiceClient),
		agentMetadataClients: make(map[string]service.AgentMetadataServiceClient),
	}

	var agentDeployments []*Deployment
	cfg := GetConfig()

	if len(cfg.DefaultAgent.Endpoint) != 0 {
		agentDeployments = append(agentDeployments, &cfg.DefaultAgent)
	}
	agentDeployments = append(agentDeployments, maps.Values(cfg.AgentDeployments)...)
	for _, agentDeployment := range agentDeployments {
		if _, ok := clientSet.agentMetadataClients[agentDeployment.Endpoint]; ok {
			logger.Infof(ctx, "Agent client already initialized for [%v]", agentDeployment.Endpoint)
			continue
		}
		conn, err := getGrpcConnection(ctx, agentDeployment)
		if err != nil {
			logger.Errorf(ctx, "failed to create connection to agent: [%v] with error: [%v]", agentDeployment, err)
			continue
		}
		clientSet.syncAgentClients[agentDeployment.Endpoint] = service.NewSyncAgentServiceClient(conn)
		clientSet.asyncAgentClients[agentDeployment.Endpoint] = service.NewAsyncAgentServiceClient(conn)
		clientSet.agentMetadataClients[agentDeployment.Endpoint] = service.NewAgentMetadataServiceClient(conn)
	}
	return clientSet
}
