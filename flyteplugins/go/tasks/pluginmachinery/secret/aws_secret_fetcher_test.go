package secret

import (
	"context"
	"fmt"
	"testing"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/secretsmanager"
	"github.com/aws/aws-sdk-go-v2/service/secretsmanager/types"
	"github.com/stretchr/testify/assert"

	"github.com/flyteorg/flyte/v2/flyteplugins/go/tasks/pluginmachinery/secret/config"
	"github.com/flyteorg/flyte/v2/flyteplugins/go/tasks/pluginmachinery/secret/mocks"
	stdlibErrors "github.com/flyteorg/flyte/v2/flytestdlib/errors"
	"github.com/flyteorg/flyte/v2/flytestdlib/promutils"
)

var (
	ctx       context.Context
	scope     promutils.Scope
	awsClient *mocks.AWSSecretManagerClient
)

const secretID = "secretID"

func SetupTest() {
	scope = promutils.NewTestScope()
	ctx = context.Background()
	awsClient = &mocks.AWSSecretManagerClient{}
}

func TestGetSecretValueAWS(t *testing.T) {
	t.Run("get secret successful", func(t *testing.T) {
		SetupTest()
		awsSecretsFetcher := NewAWSSecretFetcher(config.AWSConfig{}, awsClient)
		awsClient.OnGetSecretValueMatch(ctx, &secretsmanager.GetSecretValueInput{
			SecretId:     aws.String(secretID),
			VersionStage: aws.String(AWSSecretLatestVersion),
		}).Return(&secretsmanager.GetSecretValueOutput{
			SecretString: aws.String("secretValue"),
		}, nil)

		_, err := awsSecretsFetcher.GetSecretValue(ctx, "secretID")
		assert.NoError(t, err)
	})

	t.Run("get secret not found", func(t *testing.T) {
		SetupTest()
		awsSecretsFetcher := NewAWSSecretFetcher(config.AWSConfig{}, awsClient)
		cause := &types.ResourceNotFoundException{}
		awsClient.OnGetSecretValueMatch(ctx, &secretsmanager.GetSecretValueInput{
			SecretId:     aws.String(secretID),
			VersionStage: aws.String(AWSSecretLatestVersion),
		}).Return(nil, cause)

		_, err := awsSecretsFetcher.GetSecretValue(ctx, "secretID")
		assert.Equal(t, stdlibErrors.Wrapf(ErrCodeSecretNotFound, cause, fmt.Sprintf(SecretNotFoundErrorFormat, secretID)), err)
	})

	t.Run("get secret read failure", func(t *testing.T) {
		SetupTest()
		awsSecretsFetcher := NewAWSSecretFetcher(config.AWSConfig{}, awsClient)
		cause := fmt.Errorf("some error")
		awsClient.OnGetSecretValueMatch(ctx, &secretsmanager.GetSecretValueInput{
			SecretId:     aws.String(secretID),
			VersionStage: aws.String(AWSSecretLatestVersion),
		}).Return(nil, cause)

		_, err := awsSecretsFetcher.GetSecretValue(ctx, "secretID")
		assert.Equal(t, stdlibErrors.Wrapf(ErrCodeSecretReadFailure, cause, fmt.Sprintf(SecretReadFailureErrorFormat, secretID)), err)
	})
}
