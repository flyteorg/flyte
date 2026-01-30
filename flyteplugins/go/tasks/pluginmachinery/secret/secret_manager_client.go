package secret

import (
	"context"

	"cloud.google.com/go/secretmanager/apiv1/secretmanagerpb"
	"github.com/Azure/azure-sdk-for-go/sdk/security/keyvault/azsecrets"
	"github.com/aws/aws-sdk-go-v2/service/secretsmanager"
	"github.com/googleapis/gax-go/v2"
)

//go:generate mockery --output=./mocks --case=underscore -name=AWSSecretManagerClient

// AWSSecretManagerClient AWS Secret Manager API interface used in the webhook for looking up the secret to mount on the user pod.
type AWSSecretManagerClient interface {
	GetSecretValue(context.Context, *secretsmanager.GetSecretValueInput, ...func(*secretsmanager.Options)) (*secretsmanager.GetSecretValueOutput, error)
}

// GCPSecretManagerClient GCP Secret Manager API interface used in the webhook for looking up the secret to mount on the user pod.
//
//go:generate mockery --output=./mocks --case=underscore -name=GCPSecretManagerClient
type GCPSecretManagerClient interface {
	AccessSecretVersion(ctx context.Context, req *secretmanagerpb.AccessSecretVersionRequest, opts ...gax.CallOption) (*secretmanagerpb.AccessSecretVersionResponse, error)
}

// AzureKeyVaultClient Azure Key Vault API interface used in the webhook for looking up the secret to mount on the user pod.
//
//go:generate mockery --output=./mocks --case=underscore -name=AzureKeyVaultClient
type AzureKeyVaultClient interface {
	GetSecret(ctx context.Context, name string, version string, options *azsecrets.GetSecretOptions) (azsecrets.GetSecretResponse, error)
}
