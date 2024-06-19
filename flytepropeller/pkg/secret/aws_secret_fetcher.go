package secret

import (
	"context"
	"errors"
	"fmt"

	"github.com/aws/aws-sdk-go-v2/aws"
	awssm "github.com/aws/aws-sdk-go-v2/service/secretsmanager"
	"github.com/aws/aws-sdk-go-v2/service/secretsmanager/types"

	"github.com/flyteorg/flyte/flytepropeller/pkg/secret/config"
	stdlibErrors "github.com/flyteorg/flyte/flytestdlib/errors"
	"github.com/flyteorg/flyte/flytestdlib/logger"
)

type AWSSecretFetcher struct {
	client AWSSecretsIface
	cfg    config.AWSConfig
}

func (a AWSSecretFetcher) Get(ctx context.Context, key string) (string, error) {
	return a.GetSecretValue(ctx, key)
}

func (a AWSSecretFetcher) GetSecretValue(ctx context.Context, secretID string) (string, error) {
	logger.Infof(ctx, "Got fetch secret Request for %v!", secretID)
	resp, err := a.client.GetSecretValue(ctx, &awssm.GetSecretValueInput{
		SecretId:     aws.String(secretID),
		VersionStage: aws.String(AWSSecretLatesVersion),
	})
	if err != nil {
		var notFound *types.ResourceNotFoundException
		if errors.As(err, &notFound) {
			wrappedErr := stdlibErrors.Wrapf(ErrCodeSecretNotFound, err, fmt.Sprintf(SecretNotFoundErrorFormat, secretID))
			logger.Warn(ctx, wrappedErr)
			return "", wrappedErr
		}
		wrappedErr := stdlibErrors.Wrapf(ErrCodeSecretReadFailure, err, fmt.Sprintf(SecretReadFailureErrorFormat, secretID))
		logger.Error(ctx, wrappedErr)
		return "", wrappedErr
	}
	if resp.SecretString == nil || *resp.SecretString == "" {
		wrappedErr := stdlibErrors.Wrapf(ErrCodeSecretNil, err, fmt.Sprintf(SecretNilErrorFormat, secretID))
		logger.Error(ctx, wrappedErr)
		return "", wrappedErr
	}
	return *resp.SecretString, nil
}

// NewAWSSecretFetcher creates a secret value fetcher for AWS
func NewAWSSecretFetcher(cfg config.AWSConfig, client AWSSecretsIface) SecretFetcher {
	return AWSSecretFetcher{cfg: cfg, client: client}
}
