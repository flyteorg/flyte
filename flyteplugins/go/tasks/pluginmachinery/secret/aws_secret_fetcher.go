package secret

import (
	"context"
	"errors"
	"fmt"

	"github.com/aws/aws-sdk-go-v2/aws"
	awssm "github.com/aws/aws-sdk-go-v2/service/secretsmanager"
	"github.com/aws/aws-sdk-go-v2/service/secretsmanager/types"

	"github.com/flyteorg/flyte/v2/flyteplugins/go/tasks/pluginmachinery/secret/config"
	stdlibErrors "github.com/flyteorg/flyte/v2/flytestdlib/errors"
	"github.com/flyteorg/flyte/v2/flytestdlib/logger"
)

const (
	AWSSecretLatestVersion = "AWSCURRENT"
)

type AWSSecretFetcher struct {
	client AWSSecretManagerClient
	cfg    config.AWSConfig
}

func (a AWSSecretFetcher) GetSecretValue(ctx context.Context, secretID string) (*SecretValue, error) {
	logger.Infof(ctx, "Got fetch secret Request for %v!", secretID)
	resp, err := a.client.GetSecretValue(ctx, &awssm.GetSecretValueInput{
		SecretId:     aws.String(secretID),
		VersionStage: aws.String(AWSSecretLatestVersion),
	})

	if err != nil {
		var notFound *types.ResourceNotFoundException
		if errors.As(err, &notFound) {
			wrappedErr := stdlibErrors.Wrapf(ErrCodeSecretNotFound, err, fmt.Sprintf(SecretNotFoundErrorFormat, secretID))
			logger.Warn(ctx, wrappedErr)
			return nil, wrappedErr
		}
		wrappedErr := stdlibErrors.Wrapf(ErrCodeSecretReadFailure, err, fmt.Sprintf(SecretReadFailureErrorFormat, secretID))
		logger.Error(ctx, wrappedErr)
		return nil, wrappedErr
	}

	if (resp.SecretString == nil || *resp.SecretString == "") && resp.SecretBinary == nil {
		wrappedErr := stdlibErrors.Wrapf(ErrCodeSecretNil, err, fmt.Sprintf(SecretNilErrorFormat, secretID))
		logger.Error(ctx, wrappedErr)
		return nil, wrappedErr
	}

	secretValue := &SecretValue{}
	if resp.SecretString != nil {
		secretValue.StringValue = *resp.SecretString
	} else {
		secretValue.BinaryValue = resp.SecretBinary
	}

	return secretValue, nil
}

// NewAWSSecretFetcher creates a secret value fetcher for AWS
func NewAWSSecretFetcher(cfg config.AWSConfig, client AWSSecretManagerClient) SecretFetcher {
	return AWSSecretFetcher{cfg: cfg, client: client}
}
