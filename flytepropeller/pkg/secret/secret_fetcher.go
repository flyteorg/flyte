package secret

import (
	"context"
	"fmt"

	gcpsm "cloud.google.com/go/secretmanager/apiv1"
	awsConfig "github.com/aws/aws-sdk-go-v2/config"
	awssm "github.com/aws/aws-sdk-go-v2/service/secretsmanager"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"

	"github.com/flyteorg/flyte/flytepropeller/pkg/secret/config"
	"github.com/flyteorg/flyte/flytestdlib/logger"
)

type SecretFetcher interface {
	GetSecretValue(ctx context.Context, secretID string) (*SecretValue, error)
}

type SecretValue struct {
	StringValue string
	BinaryValue []byte
}

func NewSecretFetcher(ctx context.Context, cfg config.EmbeddedSecretManagerConfig) (SecretFetcher, error) {
	switch cfg.Type {
	case config.EmbeddedSecretManagerTypeAWS:
		awsCfg, err := awsConfig.LoadDefaultConfig(ctx, awsConfig.WithRegion(cfg.AWSConfig.Region))
		if err != nil {
			logger.Errorf(ctx, "failed to start secret manager service due to %v", err)
			return nil, fmt.Errorf("failed to start secret manager service due to %v", err)
		}
		return NewAWSSecretFetcher(cfg.AWSConfig, awssm.NewFromConfig(awsCfg)), nil
	case config.EmbeddedSecretManagerTypeGCP:
		gcpSmClient, err := gcpsm.NewClient(ctx)
		if err != nil {
			logger.Errorf(ctx, "failed to start secret manager service due to %v", err)
			return nil, fmt.Errorf("failed to start secret manager service due to %v", err)
		}
		return NewGCPSecretFetcher(cfg.GCPConfig, gcpSmClient), nil
	case config.EmbeddedSecretManagerTypeAzure:
		client, err := getAzureKVSecretsClient(cfg.AzureConfig.VaultURI)
		if err != nil {
			logger.Errorf(ctx, "failed to start secret manager service due to %v", err)
			return nil, fmt.Errorf("failed to start secret manager service due to %v", err)
		}
		logger.Info(ctx, "Using Union Azure Secret Name Codec")
		return NewAzureSecretFetcher(client), nil
	case config.EmbeddedSecretManagerTypeK8s:
		logger.Infof(ctx, "use k8s secret manager service")
		kubeConfig, err := rest.InClusterConfig()
		if err != nil {
			logger.Errorf(ctx, "Failed to get kubernetes config: %v", err)
			return nil, fmt.Errorf("failed to start secret manager service due to %v", err)
		}
		kubeClientset, err := kubernetes.NewForConfig(kubeConfig)
		if err != nil {
			logger.Errorf(ctx, "Failed to create kubernetes clientset: %v", err)
			return nil, fmt.Errorf("failed to start secret manager service due to %v", err)
		}
		secretClient := kubeClientset.CoreV1().Secrets(cfg.K8sConfig.Namespace)
		return NewK8sSecretFetcher(secretClient), nil
	}

	return nil, fmt.Errorf("failed to start secret fetcher service due to unsupported type %v. Only supported for aws and gcp right now", cfg.Type)
}
