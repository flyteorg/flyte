// Collection of shared utils for initializing kubernetes clients within flyteadmin.
package flytek8s

import (
	"context"
	"os"

	"github.com/flyteorg/flyteadmin/pkg/config"
	"github.com/flyteorg/flyteadmin/pkg/errors"
	"google.golang.org/grpc/codes"

	runtimeInterfaces "github.com/flyteorg/flyteadmin/pkg/runtime/interfaces"
	"github.com/flyteorg/flytestdlib/logger"
	_ "k8s.io/client-go/plugin/pkg/client/auth/gcp" // to overcome gke auth provider issue
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

// Initializes a config using a variety of configurable or default fallback options that can be passed to a Kubernetes client on
// initialization.
func GetRestClientConfig(kubeConfigPathString, master string,
	k8sCluster *runtimeInterfaces.ClusterConfig) (*restclient.Config, error) {
	var kubeConfiguration *restclient.Config
	var err error

	kubeClientConfig := &config.GetConfig().KubeClientConfig
	if kubeConfigPathString != "" {
		// ExpandEnv allows using $HOME in the path and it will automatically map to the right OS's user home
		kubeConfigPath := os.ExpandEnv(kubeConfigPathString)
		kubeConfiguration, err = clientcmd.BuildConfigFromFlags(master, kubeConfigPath)
		if err != nil {
			return nil, errors.NewFlyteAdminErrorf(codes.InvalidArgument, "Error building kubeconfig: %v", err)
		}
		logger.Debugf(context.Background(), "successfully loaded kube config from %s", kubeConfigPathString)
	} else if k8sCluster != nil {
		kubeConfiguration, err = RemoteClusterConfig(k8sCluster.Endpoint, k8sCluster.Auth)
		if err != nil {
			return nil, err
		}
		logger.Debugf(context.Background(), "successfully loaded kube configuration from %v", k8sCluster)

		if k8sCluster.KubeClientConfig != nil {
			logger.Debugf(context.Background(), "using rest config from remote cluster override for k8s cluster %s", k8sCluster.Name)
			kubeClientConfig = k8sCluster.KubeClientConfig
		}
	} else {
		kubeConfiguration, err = restclient.InClusterConfig()
		if err != nil {
			return nil, errors.NewFlyteAdminErrorf(codes.Internal, "Cannot get incluster kubeconfig : %v", err.Error())
		}
		logger.Debug(context.Background(), "successfully loaded kube configuration from in cluster config")
	}

	if kubeClientConfig != nil {
		kubeConfiguration.QPS = float32(kubeClientConfig.QPS)
		kubeConfiguration.Burst = kubeClientConfig.Burst
		kubeConfiguration.Timeout = kubeClientConfig.Timeout.Duration
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
