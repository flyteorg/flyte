package google

import (
	"context"

	pluginmachinery "github.com/flyteorg/flyteplugins/go/tasks/pluginmachinery/k8s"
	"github.com/pkg/errors"
	"golang.org/x/oauth2"
	"google.golang.org/api/impersonate"
	"google.golang.org/grpc/credentials/oauth"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
)

const (
	gcpServiceAccountAnnotationKey = "iam.gke.io/gcp-service-account"
	workflowIdentityDocURL         = "https://cloud.google.com/kubernetes-engine/docs/how-to/workload-identity"
)

var impersonationScopes = []string{"https://www.googleapis.com/auth/bigquery"}

type GkeTaskWorkloadIdentityTokenSourceFactoryConfig struct {
	RemoteClusterConfig pluginmachinery.ClusterConfig `json:"remoteClusterConfig" pflag:"Configuration of remote GKE cluster"`
}

type gkeTaskWorkloadIdentityTokenSourceFactory struct {
	kubeClient kubernetes.Interface
}

func (m *gkeTaskWorkloadIdentityTokenSourceFactory) getGcpServiceAccount(
	ctx context.Context,
	identity Identity,
) (string, error) {
	if identity.K8sServiceAccount == "" {
		identity.K8sServiceAccount = "default"
	}
	serviceAccount, err := m.kubeClient.CoreV1().ServiceAccounts(identity.K8sNamespace).Get(
		ctx,
		identity.K8sServiceAccount,
		metav1.GetOptions{},
	)
	if err != nil {
		return "", errors.Wrapf(err, "failed to retrieve task k8s service account")
	}

	for key, value := range serviceAccount.Annotations {
		if key == gcpServiceAccountAnnotationKey {
			return value, nil
		}
	}

	return "", errors.Errorf(
		"[%v] annotation doesn't exist on k8s service account [%v/%v], read more at %v",
		gcpServiceAccountAnnotationKey,
		identity.K8sNamespace,
		identity.K8sServiceAccount,
		workflowIdentityDocURL)
}

func (m *gkeTaskWorkloadIdentityTokenSourceFactory) GetTokenSource(
	ctx context.Context,
	identity Identity,
) (oauth2.TokenSource, error) {
	gcpServiceAccount, err := m.getGcpServiceAccount(ctx, identity)
	if err != nil {
		return oauth.TokenSource{}, err
	}

	return impersonate.CredentialsTokenSource(ctx, impersonate.CredentialsConfig{
		TargetPrincipal: gcpServiceAccount,
		Scopes:          impersonationScopes,
	})
}

func getKubeClient(
	config *GkeTaskWorkloadIdentityTokenSourceFactoryConfig,
) (*kubernetes.Clientset, error) {
	var kubeCfg *rest.Config
	var err error
	if config.RemoteClusterConfig.Enabled {
		kubeCfg, err = pluginmachinery.KubeClientConfig(
			config.RemoteClusterConfig.Endpoint,
			config.RemoteClusterConfig.Auth,
		)
		if err != nil {
			return nil, errors.Wrapf(err, "Error building kubeconfig")
		}
	} else {
		kubeCfg, err = rest.InClusterConfig()
		if err != nil {
			return nil, errors.Wrapf(err, "Cannot get InCluster kubeconfig")
		}
	}

	kubeClient, err := kubernetes.NewForConfig(kubeCfg)
	if err != nil {
		return nil, errors.Wrapf(err, "Error building kubernetes clientset")
	}
	return kubeClient, err
}

func NewGkeTaskWorkloadIdentityTokenSourceFactory(
	config *GkeTaskWorkloadIdentityTokenSourceFactoryConfig,
) (TokenSourceFactory, error) {
	kubeClient, err := getKubeClient(config)
	if err != nil {
		return nil, err
	}
	return &gkeTaskWorkloadIdentityTokenSourceFactory{kubeClient: kubeClient}, nil
}
