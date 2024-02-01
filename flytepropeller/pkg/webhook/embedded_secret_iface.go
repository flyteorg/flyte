package webhook

import (
	"context"

	"cloud.google.com/go/secretmanager/apiv1/secretmanagerpb"
	"github.com/aws/aws-sdk-go-v2/service/secretsmanager"
	"github.com/googleapis/gax-go/v2"
)

//go:generate mockery --output=./mocks --case=underscore -name=AWSSecretsIface

// AWSSecretsIface AWS Secret Manager API interface used in the webhook for looking up the secret to mount on the user pod.
type AWSSecretsIface interface {
	GetSecretValue(context.Context, *secretsmanager.GetSecretValueInput, ...func(*secretsmanager.Options)) (*secretsmanager.GetSecretValueOutput, error)
}

// GCPSecretsIface GCP Secret Manager API interface used in the webhook  for looking up the secret to mount on the user pod.
//
//go:generate mockery --output=./mocks --case=underscore -name=GCPSecretsIface
type GCPSecretsIface interface {
	AccessSecretVersion(ctx context.Context, req *secretmanagerpb.AccessSecretVersionRequest, opts ...gax.CallOption) (*secretmanagerpb.AccessSecretVersionResponse, error)
}
