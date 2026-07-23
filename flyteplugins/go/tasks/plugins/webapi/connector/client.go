package connector

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
	"google.golang.org/grpc/keepalive"
	"google.golang.org/grpc/status"

	"github.com/flyteorg/flyte/v2/flytestdlib/config"
	"github.com/flyteorg/flyte/v2/flytestdlib/grpcutils"
	"github.com/flyteorg/flyte/v2/flytestdlib/logger"
	"github.com/flyteorg/flyte/v2/gen/go/flyteidl2/connector"
)

const defaultTaskTypeVersion = 0

// DefaultGRPCServiceConfig is the gRPC service config applied to a connector
// connection when its Deployment does not set DefaultServiceConfig. It enables
// round-robin load balancing across connector replicas and retries transient
// UNAVAILABLE failures so a single dropped connection does not fail the task.
// It is exported so callers that register connectors programmatically (e.g. the
// operator app controller) can set it explicitly.
const DefaultGRPCServiceConfig = `{
  "loadBalancingConfig": [{"round_robin":{}}],
  "methodConfig": [{
    "name": [{"service": "flyteidl2.connector.AsyncConnectorService"}],
    "retryPolicy": {
      "maxAttempts": 4,
      "initialBackoff": "0.2s",
      "maxBackoff": "3s",
      "backoffMultiplier": 2.0,
      "retryableStatusCodes": ["UNAVAILABLE"]
    }
  }],
  "retryThrottling": {"maxTokens": 10, "tokenRatio": 0.1}
}`

type Connector struct {
	// ConnectorDeployment is the connector deployment where this connector is running.
	ConnectorDeployment *Deployment
	// ConnectorID is the ID of the connector.
	ConnectorID string
	// IsConnectorApp indicates whether this connector is a connector app.
	IsConnectorApp bool
}

// ClientSet contains the clients exposed to communicate with various connector services.
type ClientSet struct {
	asyncConnectorClients    map[string]connector.AsyncConnectorServiceClient    // map[endpoint] => AsyncConnectorServiceClient
	connectorMetadataClients map[string]connector.ConnectorMetadataServiceClient // map[endpoint] => ConnectorMetadataServiceClient
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

	serviceConfig := connector.DefaultServiceConfig
	if len(serviceConfig) == 0 {
		serviceConfig = DefaultGRPCServiceConfig
	}
	opts = append(opts, grpc.WithDefaultServiceConfig(serviceConfig))

	// Keepalive is opt-in per deployment. It is only safe for connectors fronted
	// by an L7 gateway (e.g. Knative/Kourier Envoy) that reaps idle connections;
	// a directly-reached grpc-go or grpcio (C-core) server rejects frequent idle
	// pings with GOAWAY "too_many_pings".
	if connector.Keepalive != nil {
		opts = append(opts, grpc.WithKeepaliveParams(keepalive.ClientParameters{
			Time:                connector.Keepalive.Time.Duration,
			Timeout:             connector.Keepalive.Timeout.Duration,
			PermitWithoutStream: connector.Keepalive.PermitWithoutStream,
		}))
	}

	clientMetrics := grpcutils.GrpcClientMetrics()
	opts = append(opts,
		grpc.WithChainUnaryInterceptor(clientMetrics.UnaryClientInterceptor()),
		grpc.WithChainStreamInterceptor(clientMetrics.StreamClientInterceptor()),
	)

	var err error
	conn, err := grpc.Dial(connector.Endpoint, opts...) //nolint: staticcheck
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

func updateRegistry(
	ctx context.Context,
	cs *ClientSet,
	newConnectorRegistry Registry,
	connectorDeployments map[string]*Deployment,
	isConnectorApp bool,
) {
	for connectorID, connectorDeployment := range connectorDeployments {
		client, ok := cs.connectorMetadataClients[connectorDeployment.Endpoint]
		if !ok {
			logger.Warningf(ctx, "Connector client not found in the clientSet for the endpoint: %v", connectorDeployment.Endpoint)
			continue
		}

		finalCtx, cancel := getFinalContext(ctx, "ListConnectors", connectorDeployment)
		defer cancel()

		res, err := client.ListConnectors(finalCtx, &connector.ListConnectorsRequest{})
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
		for _, agent := range res.GetConnectors() {
			supportedTaskCategories := agent.GetSupportedTaskCategories()
			for _, supportedCategory := range supportedTaskCategories {
				connector := &Connector{
					ConnectorDeployment: connectorDeployment,
					ConnectorID:         connectorID,
					IsConnectorApp:      isConnectorApp,
				}
				supportedCategoryName := supportedCategory.GetName()
				registryKey := RegistryKey{project: connectorDeployment.Project, domain: connectorDeployment.Domain, taskTypeName: supportedCategoryName, taskTypeVersion: supportedCategory.GetVersion()}
				newConnectorRegistry[registryKey] = connector
				connectorSupportedTaskCategories[supportedCategoryName] = struct{}{}
			}
		}
		logger.Infof(ctx, "ConnectorDeployment [%v] supports the following task types: [%v]", connectorDeployment.Endpoint,
			strings.Join(maps.Keys(connectorSupportedTaskCategories), ", "))
	}
}

func getConnectorRegistry(ctx context.Context, cs *ClientSet) Registry {
	newConnectorRegistry := make(Registry)
	cfg := GetConfig()

	connectorDeployments := make(map[string]*Deployment)

	if len(cfg.DefaultConnector.Endpoint) != 0 {
		connectorDeployments["defaultConnector"] = &cfg.DefaultConnector
	}

	for connectorID, deployment := range cfg.ConnectorDeployments {
		connectorDeployments[connectorID] = deployment
	}

	// Update registry with regular connectors
	updateRegistry(ctx, cs, newConnectorRegistry, connectorDeployments, false)

	// If the connector doesn't implement the metadata service, we construct the registry based on the configuration
	for taskType, connectorDeploymentID := range cfg.ConnectorForTaskTypes {
		if connectorDeployment, ok := cfg.ConnectorDeployments[connectorDeploymentID]; ok {
			registryKey := RegistryKey{project: connectorDeployment.Project, domain: connectorDeployment.Domain, taskTypeName: taskType, taskTypeVersion: defaultTaskTypeVersion}
			connector := &Connector{
				ConnectorDeployment: connectorDeployment,
				ConnectorID:         connectorDeploymentID,
				IsConnectorApp:      false,
			}
			newConnectorRegistry[registryKey] = connector
		}
	}

	// Ensure that the old configuration is backward compatible
	for _, taskType := range cfg.SupportedTaskTypes {
		registryKey := RegistryKey{project: cfg.DefaultConnector.Project, domain: cfg.DefaultConnector.Domain, taskTypeName: taskType, taskTypeVersion: defaultTaskTypeVersion}
		if _, ok := newConnectorRegistry[registryKey]; !ok {
			connector := &Connector{
				ConnectorDeployment: &cfg.DefaultConnector,
				ConnectorID:         "defaultConnector",
				IsConnectorApp:      false,
			}
			newConnectorRegistry[registryKey] = connector
		}
	}

	// Update registry with connector apps
	updateRegistry(ctx, cs, newConnectorRegistry, cfg.ConnectorApps, true)

	logger.Infof(ctx, "ConnectorDeployments support the following task types: [%v]", strings.Join(newConnectorRegistry.getSupportedTaskTypes(), ", "))
	return newConnectorRegistry
}

func getConnectorClientSets(ctx context.Context) *ClientSet {
	clientSet := &ClientSet{
		asyncConnectorClients:    make(map[string]connector.AsyncConnectorServiceClient),
		connectorMetadataClients: make(map[string]connector.ConnectorMetadataServiceClient),
	}

	var connectorDeployments []*Deployment
	cfg := GetConfig()

	if len(cfg.DefaultConnector.Endpoint) != 0 {
		connectorDeployments = append(connectorDeployments, &cfg.DefaultConnector)
	}

	for _, deployment := range cfg.ConnectorDeployments {
		connectorDeployments = append(connectorDeployments, deployment)
	}

	for _, deployment := range cfg.ConnectorApps {
		connectorDeployments = append(connectorDeployments, deployment)
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
		clientSet.asyncConnectorClients[connectorDeployment.Endpoint] = connector.NewAsyncConnectorServiceClient(conn)
		clientSet.connectorMetadataClients[connectorDeployment.Endpoint] = connector.NewConnectorMetadataServiceClient(conn)
	}
	return clientSet
}
