package connector

import (
	"context"
	"crypto/x509"
	"github.com/flyteorg/flyte/flyteidl/gen/pb-go/flyteidl/admin"
	"golang.org/x/exp/maps"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/credentials"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/grpclog"
	"google.golang.org/grpc/status"
	"strings"

	"github.com/flyteorg/flyte/flyteidl/gen/pb-go/flyteidl/service"
	"github.com/flyteorg/flyte/flytestdlib/config"
	"github.com/flyteorg/flyte/flytestdlib/logger"
)

const defaultTaskTypeVersion = 0

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

func getConnectorRegistry(ctx context.Context, cs *ClientSet) Registry {
	newConnectorRegistry := make(Registry)
	cfg := GetConfig()
	var connectorDeployments []*Deployment

	if len(cfg.DefaultConnector.Endpoint) != 0 {
		connectorDeployments = append(connectorDeployments, &cfg.DefaultConnector)
	}
	connectorDeployments = append(connectorDeployments, maps.Values(cfg.ConnectorDeployments)...)
	for _, connectorDeployment := range connectorDeployments {
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

		connectorSupportedTaskCategories := make(map[string]struct{})
		for _, connector := range res.GetAgents() {
			deprecatedSupportedTaskTypes := connector.GetSupportedTaskTypes()
			for _, supportedTaskType := range deprecatedSupportedTaskTypes {
				connector := &Connector{ConnectorDeployment: connectorDeployment, IsSync: connector.GetIsSync()}
				newConnectorRegistry[supportedTaskType] = map[int32]*Connector{defaultTaskTypeVersion: connector}
				connectorSupportedTaskCategories[supportedTaskType] = struct{}{}
			}

			supportedTaskCategories := connector.GetSupportedTaskCategories()
			for _, supportedCategory := range supportedTaskCategories {
				connector := &Connector{ConnectorDeployment: connectorDeployment, IsSync: connector.GetIsSync()}
				supportedCategoryName := supportedCategory.GetName()
				newConnectorRegistry[supportedCategoryName] = map[int32]*Connector{supportedCategory.GetVersion(): connector}
				connectorSupportedTaskCategories[supportedCategoryName] = struct{}{}
			}
		}
		logger.Infof(ctx, "ConnectorDeployment [%v] supports the following task types: [%v]", connectorDeployment.Endpoint,
			strings.Join(maps.Keys(connectorSupportedTaskCategories), ", "))
	}

	// If the connector doesn't implement the metadata service, we construct the registry based on the configuration
	for taskType, connectorDeploymentID := range cfg.ConnectorForTaskTypes {
		if connectorDeployment, ok := cfg.ConnectorDeployments[connectorDeploymentID]; ok {
			if _, ok := newConnectorRegistry[taskType]; !ok {
				connector := &Connector{ConnectorDeployment: connectorDeployment, IsSync: false}
				newConnectorRegistry[taskType] = map[int32]*Connector{defaultTaskTypeVersion: connector}
			}
		}
	}

	// Ensure that the old configuration is backward compatible
	for _, taskType := range cfg.SupportedTaskTypes {
		if _, ok := newConnectorRegistry[taskType]; !ok {
			connector := &Connector{ConnectorDeployment: &cfg.DefaultConnector, IsSync: false}
			newConnectorRegistry[taskType] = map[int32]*Connector{defaultTaskTypeVersion: connector}
		}
	}

	logger.Infof(ctx, "ConnectorDeployments support the following task types: [%v]", strings.Join(maps.Keys(newConnectorRegistry), ", "))
	return newConnectorRegistry
}

func getConnectorClientSets(ctx context.Context) *ClientSet {
	clientSet := &ClientSet{
		asyncConnectorClients:    make(map[string]service.AsyncAgentServiceClient),
		syncConnectorClients:     make(map[string]service.SyncAgentServiceClient),
		connectorMetadataClients: make(map[string]service.AgentMetadataServiceClient),
	}

	var connectorDeployments []*Deployment
	cfg := GetConfig()

	if len(cfg.DefaultConnector.Endpoint) != 0 {
		connectorDeployments = append(connectorDeployments, &cfg.DefaultConnector)
	}
	connectorDeployments = append(connectorDeployments, maps.Values(cfg.ConnectorDeployments)...)
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
