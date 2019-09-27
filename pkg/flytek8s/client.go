// Collection of shared utils for initializing kubernetes clients within flyteadmin.
package flytek8s

import (
	"context"

	"github.com/lyft/flyteadmin/pkg/errors"
	"google.golang.org/grpc/codes"

	runtimeInterfaces "github.com/lyft/flyteadmin/pkg/runtime/interfaces"
	"github.com/lyft/flytestdlib/logger"
	restclient "k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

// Reads secret values from paths specified in the config to initialize a Kubernetes rest client Config.
func RemoteClusterConfig(host string, auth runtimeInterfaces.Auth) (*restclient.Config, error) {
	tokenString, err := auth.GetToken()
	if err != nil {
		return nil, errors.NewFlyteAdminErrorf(codes.Internal, "Failed to get auth token: %+v", err)
	}

	caCert, err := auth.GetCA()
	if err != nil {
		return nil, errors.NewFlyteAdminErrorf(codes.Internal, "Failed to get auth CA: %+v", err)
	}

	tlsClientConfig := restclient.TLSClientConfig{}
	tlsClientConfig.CAData = caCert
	return &restclient.Config{
		Host:            host,
		TLSClientConfig: tlsClientConfig,
		BearerToken:     tokenString,
	}, nil
}

func GetRestClientConfigForCluster(cluster runtimeInterfaces.ClusterConfig) (*restclient.Config, error) {
	kubeConfiguration, err := RemoteClusterConfig(cluster.Endpoint, cluster.Auth)

	if err != nil {
		return nil, err
	}
	logger.Debugf(context.Background(), "successfully loaded kube configuration from %v", cluster)
	return kubeConfiguration, nil
}

// Initializes a config using a variety of configurable or default fallback options that can be passed to a Kubernetes client on
// initialization.
func GetRestClientConfig(kubeConfig, master string,
	k8sCluster *runtimeInterfaces.ClusterConfig) (*restclient.Config, error) {
	var kubeConfiguration *restclient.Config
	var err error

	if kubeConfig != "" {
		kubeConfiguration, err = clientcmd.BuildConfigFromFlags(master, kubeConfig)
		if err != nil {
			return nil, errors.NewFlyteAdminErrorf(codes.InvalidArgument, "Error building kubeconfig: %v", err)
		}
		logger.Debugf(context.Background(), "successfully loaded kube config from %s", kubeConfig)
	} else if k8sCluster != nil {
		return GetRestClientConfigForCluster(*k8sCluster)
	} else {
		kubeConfiguration, err = restclient.InClusterConfig()
		if err != nil {
			return nil, errors.NewFlyteAdminErrorf(codes.Internal, "Cannot get incluster kubeconfig : %v", err.Error())
		}
		logger.Debug(context.Background(), "successfully loaded kube configuration from in cluster config")
	}
	return kubeConfiguration, nil
}

// Initializes a kubernetes Client which performs CRD operations on Kubernetes objects
func NewKubeClient(kubeConfig, master string, k8sCluster *runtimeInterfaces.ClusterConfig) (client.Client, error) {
	kubeConfiguration, err := GetRestClientConfig(kubeConfig, master, k8sCluster)
	if err != nil {
		return nil, err
	}
	return client.New(kubeConfiguration, client.Options{})
}
