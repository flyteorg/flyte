package connector

import (
	"context"
	"crypto/x509"
	"fmt"
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
	"github.com/flyteorg/flyte/flyteplugins/go/tasks/pluginmachinery"
	"github.com/flyteorg/flyte/flytestdlib/config"
	"github.com/flyteorg/flyte/flytestdlib/logger"
)

const defaultTaskTypeVersion = 0
const defaultDeploymentID = "default"

type Connector struct {
	// IsSync indicates whether this connector is a sync connector. Sync connectors are expected to return their
	// results synchronously when called by propeller. Given that sync connectors can affect the performance
	// of the system, it's important to enforce strict timeout policies.
	// An Async connector, on the other hand, is required to be able to identify jobs by an
	// identifier and query for job statuses as jobs progress.
	IsSync bool
	// ConnectorDeployment is the connector deployment where this connector is running.
	ConnectorDeployment *Deployment
}

// ClientSet contains the clients exposed to communicate with various connector services.
type ClientSet struct {
	asyncConnectorClients    map[string]service.AsyncAgentServiceClient    // map[endpoint] => AsyncConnectorServiceClient
	syncConnectorClients     map[string]service.SyncAgentServiceClient     // map[endpoint] => SyncConnectorServiceClient
	connectorMetadataClients map[string]service.AgentMetadataServiceClient // map[endpoint] => ConnectorMetadataServiceClient
}

func getGrpcConnection(ctx context.Context, connector *Deployment) (*grpc.ClientConn, error) {
	var opts []grpc.DialOption

	if connector.Insecure {
		opts = append(opts, grpc.WithTransportCredentials(insecure.NewCredentials()))
	} else {
		pool, err := x509.SystemCertPool()
		if err != nil {
			return nil, err
		}

		creds := credentials.NewClientTLSFromCert(pool, "")
		opts = append(opts, grpc.WithTransportCredentials(creds))
	}

	if len(connector.DefaultServiceConfig) != 0 {
		opts = append(opts, grpc.WithDefaultServiceConfig(connector.DefaultServiceConfig))
	}

	var err error
	conn, err := grpc.Dial(connector.Endpoint, opts...)
	if err != nil {
		return nil, err
	}

	go func() {
		<-ctx.Done()
		if cerr := conn.Close(); cerr != nil {
			grpclog.Infof("Failed to close conn to %s: %v", connector, cerr)
		}
	}()

	return conn, nil
}

func getFinalTimeout(operation string, connector *Deployment) config.Duration {
	if t, exists := connector.Timeouts[operation]; exists {
		return t
	}

	return connector.DefaultTimeout
}

func getFinalContext(ctx context.Context, operation string, connector *Deployment) (context.Context, context.CancelFunc) {
	timeout := getFinalTimeout(operation, connector).Duration
	if timeout == 0 {
		return ctx, func() {}
	}

	return context.WithTimeout(ctx, timeout)
}

func watchConnectors(ctx context.Context, cs *ClientSet) {
	cfg := GetConfig()
	var connectorDeploymentIDs []string
	var connectorDeployments []*Deployment

	// Merge DefaultConnector (if endpoint is not empty)
	if len(cfg.DefaultConnector.Endpoint) != 0 {
		connectorDeploymentIDs = append(connectorDeploymentIDs, defaultDeploymentID)
		connectorDeployments = append(connectorDeployments, &cfg.DefaultConnector) 
	}
	connectorDeploymentIDs = append(connectorDeploymentIDs, maps.Keys(cfg.ConnectorDeployments)...)
	connectorDeployments = append(connectorDeployments, maps.Values(cfg.ConnectorDeployments)...)
	
	for idx, connectorDeployment := range connectorDeployments {
		deploymentID := connectorDeploymentIDs[idx]
		client, ok := cs.connectorMetadataClients[connectorDeployment.Endpoint]
		if !ok {
			logger.Warningf(ctx, "Connector client not found in the clientSet for the endpoint: %v", connectorDeployment.Endpoint)
			continue
		}

		finalCtx, cancel := getFinalContext(ctx, "ListConnectors", connectorDeployment)
		defer cancel()

		res, err := client.ListAgents(finalCtx, &admin.ListAgentsRequest{})
		if err != nil {
			grpcStatus, ok := status.FromError(err)
			if grpcStatus.Code() == codes.Unimplemented {
				// we should not panic here, as we want to continue to support old connector settings
				logger.Warningf(finalCtx, "list connector method not implemented for connector: [%v]", connectorDeployment.Endpoint)
				continue
			}

			if !ok {
				logger.Errorf(finalCtx, "failed to list connector: [%v] with a non-gRPC error: [%v]", connectorDeployment.Endpoint, err)
				continue
			}

			logger.Errorf(finalCtx, "failed to list connector: [%v] with error: [%v]", connectorDeployment.Endpoint, err)
			continue
		}

		// If a connector's support task type plugin was not registered yet, we should do registration
		connectorSupportedTaskCategories := make(map[string]struct{})
		for _, connector := range res.GetAgents() {
			deprecatedSupportedTaskTypes := connector.GetSupportedTaskTypes()
			supportedTaskCategories := connector.GetSupportedTaskCategories()
			// Process deprecated supported task types
			for _, supportedTaskType := range deprecatedSupportedTaskTypes {
				versionedTaskType := createOrUpdatePlugin(ctx, supportedTaskType, defaultTaskTypeVersion, deploymentID, connectorDeployment, cs)
				connectorSupportedTaskCategories[versionedTaskType] = struct{}{}
			}
			// Process supported task categories
			for _, supportedCategory := range supportedTaskCategories {
				if supportedCategory.Version != defaultTaskTypeVersion {
					versionedTaskType := createOrUpdatePlugin(ctx, supportedCategory.Name, supportedCategory.Version, deploymentID, connectorDeployment, cs)
					connectorSupportedTaskCategories[versionedTaskType] = struct{}{}
				}
			}
		}
		logger.Infof(ctx, "ConnectorDeployment [%v] supports the following task types: [%v]", connectorDeployment.Endpoint,
			strings.Join(maps.Keys(connectorSupportedTaskCategories), ", "))
	}
	// always overwrite with connectorForTaskTypes config
	for taskType, connectorDeploymentID := range cfg.ConnectorForTaskTypes {
		if deployment, ok := cfg.ConnectorDeployments[connectorDeploymentID]; ok {
			createOrUpdatePlugin(ctx, taskType, defaultTaskTypeVersion, connectorDeploymentID, deployment, cs)
		}
	}
	// Ensure that the old configuration is backward compatible
	for _, taskType := range cfg.SupportedTaskTypes {
		versionedTaskType := fmt.Sprintf("%s_%d", taskType, defaultTaskTypeVersion)
		if ok := pluginmachinery.PluginRegistry().IsConnectorCorePluginRegistered(versionedTaskType, defaultDeploymentID); !ok {
			createOrUpdatePlugin(ctx, taskType, defaultTaskTypeVersion, defaultDeploymentID, &cfg.DefaultConnector, cs)
		}
	}
}

func getConnectorClientSets(ctx context.Context) *ClientSet {
	clientSet := &ClientSet{
		asyncConnectorClients:    make(map[string]service.AsyncAgentServiceClient),
		syncConnectorClients:     make(map[string]service.SyncAgentServiceClient),
		connectorMetadataClients: make(map[string]service.AgentMetadataServiceClient),
	}

	connectorDeployments := make(map[string]*Deployment)
	cfg := GetConfig()

	// Merge ConnectorDeployments
	for key, deployment := range cfg.ConnectorDeployments {
		connectorDeployments[key] = deployment
	}
	
	// Merge DefaultConnector (if endpoint is not empty)
	if len(cfg.DefaultConnector.Endpoint) != 0 {
		connectorDeployments["default"] = &cfg.DefaultConnector
	}
	for _, connectorDeployment := range connectorDeployments {
		if _, ok := clientSet.connectorMetadataClients[connectorDeployment.Endpoint]; ok {
			logger.Infof(ctx, "Connector client already initialized for [%v]", connectorDeployment.Endpoint)
			continue
		}
		conn, err := getGrpcConnection(ctx, connectorDeployment)
		if err != nil {
			logger.Errorf(ctx, "failed to create connection to connector: [%v] with error: [%v]", connectorDeployment, err)
			continue
		}
		clientSet.syncConnectorClients[connectorDeployment.Endpoint] = service.NewSyncAgentServiceClient(conn)
		clientSet.asyncConnectorClients[connectorDeployment.Endpoint] = service.NewAsyncAgentServiceClient(conn)
		clientSet.connectorMetadataClients[connectorDeployment.Endpoint] = service.NewAgentMetadataServiceClient(conn)
	}
	return clientSet
}
